/*
Copyright 2021 zhyass.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/presslabs/controller-util/syncer"
	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/internal"
	"github.com/zhyass/mysql-operator/utils"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const maxStatusesQuantity = 10
const checkNodeStatusRetry = 3

type StatusUpdater struct {
	log logr.Logger

	ctx context.Context

	*cluster.Cluster

	cli client.Client
}

func NewStatusUpdater(log logr.Logger, ctx context.Context, cli client.Client, c *cluster.Cluster) *StatusUpdater {
	return &StatusUpdater{
		log:     log,
		ctx:     ctx,
		Cluster: c,
		cli:     cli,
	}
}

// Object returns the object for which sync applies.
func (s *StatusUpdater) Object() interface{} { return nil }

// GetObject returns the object for which sync applies
// Deprecated: use github.com/presslabs/controller-util/syncer.Object() instead.
func (s *StatusUpdater) GetObject() interface{} { return nil }

// Owner returns the object owner or nil if object does not have one.
func (s *StatusUpdater) ObjectOwner() runtime.Object { return s.Cluster }

// GetOwner returns the object owner or nil if object does not have one.
// Deprecated: use github.com/presslabs/controller-util/syncer.ObjectOwner() instead.
func (s *StatusUpdater) GetOwner() runtime.Object { return s.Cluster }

func (s *StatusUpdater) Sync(ctx context.Context) (syncer.SyncResult, error) {
	clusterCondition := mysqlv1.ClusterCondition{
		Type:               mysqlv1.ClusterInit,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	s.Status.State = mysqlv1.ClusterInit

	list := core.PodList{}
	err := s.cli.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace:     s.Namespace,
			LabelSelector: s.GetLabels().AsSelector(),
		},
	)
	if err != nil {
		return syncer.SyncResult{}, err
	}

	// get ready nodes.
	var readyNodes []core.Pod
	for _, pod := range list.Items {
		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case core.ContainersReady:
				if cond.Status == core.ConditionTrue {
					readyNodes = append(readyNodes, pod)
				}
			case core.PodScheduled:
				if cond.Reason == core.PodReasonUnschedulable {
					clusterCondition = mysqlv1.ClusterCondition{
						Type:               mysqlv1.ClusterError,
						Status:             core.ConditionTrue,
						LastTransitionTime: metav1.NewTime(time.Now()),
						Reason:             core.PodReasonUnschedulable,
						Message:            cond.Message,
					}
					s.Status.State = mysqlv1.ClusterError
				}
			}
		}
	}

	s.Status.ReadyNodes = len(readyNodes)
	if s.Status.ReadyNodes == int(*s.Spec.Replicas) {
		s.Status.State = mysqlv1.ClusterReady
		clusterCondition.Type = mysqlv1.ClusterReady
	}

	if len(s.Status.Conditions) == 0 {
		s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
	} else {
		lastCond := s.Status.Conditions[len(s.Status.Conditions)-1]
		if lastCond.Type != clusterCondition.Type {
			s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
		}
	}
	if len(s.Status.Conditions) > maxStatusesQuantity {
		s.Status.Conditions = s.Status.Conditions[len(s.Status.Conditions)-maxStatusesQuantity:]
	}

	// update ready nodes' status.
	return syncer.SyncResult{}, s.updateNodeStatus(s.cli, readyNodes)
}

func (s *StatusUpdater) updateNodeStatus(cli client.Client, pods []core.Pod) error {
	sctName := s.GetNameForResource(utils.Secret)
	svcName := s.GetNameForResource(utils.HeadlessSVC)
	port := utils.MysqlPort
	nameSpace := s.Namespace

	secret := &core.Secret{}
	if err := cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: nameSpace,
			Name:      sctName,
		},
		secret,
	); err != nil {
		s.log.V(1).Info("secret '%s' not found", sctName)
		return nil
	}
	user, ok := secret.Data["metrics-user"]
	if !ok {
		return fmt.Errorf("failed to get the user: %s", user)
	}
	password, ok := secret.Data["metrics-password"]
	if !ok {
		return fmt.Errorf("failed to get the password: %s", password)
	}

	for _, pod := range pods {
		host := fmt.Sprintf("%s.%s.%s", pod.Name, svcName, nameSpace)
		index := s.getNodeStatusIndex(host)
		node := &s.Status.Nodes[index]
		node.Message = ""
		node.Healthy = false

		isLeader, err := checkRole(&pod)
		if err != nil {
			s.log.Error(err, "failed to check the node role", "node", node.Name)
			node.Message = err.Error()
		}
		// update mysqlv1.NodeConditionLeader.
		s.updateNodeCondition(node, 1, isLeader)

		isLagged, isReplicating, isReadOnly := core.ConditionUnknown, core.ConditionUnknown, core.ConditionUnknown
		runner, err := internal.NewSQLRunner(utils.BytesToString(user), utils.BytesToString(password), host, port)
		if err != nil {
			s.log.Error(err, "failed to connect the mysql", "node", node.Name)
			node.Message = err.Error()
		} else {
			isLagged, isReplicating, err = runner.CheckSlaveStatusWithRetry(checkNodeStatusRetry)
			if err != nil {
				s.log.Error(err, "failed to check slave status", "node", node.Name)
				node.Message = err.Error()
			}

			isReadOnly, err = runner.CheckReadOnly()
			if err != nil {
				s.log.Error(err, "failed to check read only", "node", node.Name)
				node.Message = err.Error()
			}
		}
		if runner != nil {
			runner.Close()
		}

		if isLeader == core.ConditionTrue && isReadOnly != core.ConditionFalse {
			s.log.V(1).Info("try to correct the leader writeable", "node", node.Name)
			correctLeaderReadOnly(&pod)
		}

		// update mysqlv1.NodeConditionLagged.
		s.updateNodeCondition(node, 0, isLagged)
		// update mysqlv1.NodeConditionReplicating.
		s.updateNodeCondition(node, 3, isReplicating)
		// update mysqlv1.NodeConditionReadOnly.
		s.updateNodeCondition(node, 2, isReadOnly)

		setNodeStatusHealthy(node)
	}

	return nil
}

func (s *StatusUpdater) getNodeStatusIndex(name string) int {
	len := len(s.Status.Nodes)
	for i := 0; i < len; i++ {
		if s.Status.Nodes[i].Name == name {
			return i
		}
	}

	lastTransitionTime := metav1.NewTime(time.Now())
	status := mysqlv1.NodeStatus{
		Name:    name,
		Healthy: false,
		Conditions: []mysqlv1.NodeCondition{
			{
				Type:               mysqlv1.NodeConditionLagged,
				Status:             core.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               mysqlv1.NodeConditionLeader,
				Status:             core.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               mysqlv1.NodeConditionReadOnly,
				Status:             core.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               mysqlv1.NodeConditionReplicating,
				Status:             core.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
		},
	}
	s.Status.Nodes = append(s.Status.Nodes, status)
	return len
}

func (s *StatusUpdater) updateNodeCondition(node *mysqlv1.NodeStatus, idx int, status core.ConditionStatus) {
	if node.Conditions[idx].Status != status {
		t := time.Now()
		s.log.V(3).Info(fmt.Sprintf("Found status change for node %q condition %q: %q -> %q; setting lastTransitionTime to %v",
			node.Name, node.Conditions[idx].Type, node.Conditions[idx].Status, status, t))
		node.Conditions[idx].Status = status
		node.Conditions[idx].LastTransitionTime = metav1.NewTime(t)
	}
}

func checkRole(pod *core.Pod) (core.ConditionStatus, error) {
	command := []string{"xenoncli", "raft", "status"}
	status := core.ConditionUnknown
	executor, err := internal.NewPodExecutor()
	if err != nil {
		return status, err
	}

	stdout, stderr, err := executor.Exec(pod, "xenon", command...)
	if err != nil {
		return status, err
	}

	if len(stderr) != 0 {
		return status, fmt.Errorf("run command %s in xenon failed: %s", command, stderr)
	}

	var out map[string]interface{}
	if err = json.Unmarshal(stdout, &out); err != nil {
		return status, err
	}

	if out["state"] == "LEADER" {
		return core.ConditionTrue, nil
	}

	if out["state"] == "FOLLOWER" {
		return core.ConditionFalse, nil
	}

	return status, nil
}

func correctLeaderReadOnly(pod *core.Pod) error {
	executor, err := internal.NewPodExecutor()
	if err != nil {
		return err
	}

	err = executor.SetGlobalSysVar(pod, "SET GLOBAL read_only=off")
	if err != nil {
		return err
	}

	return executor.SetGlobalSysVar(pod, "SET GLOBAL super_read_only=off")
}

func setNodeStatusHealthy(node *mysqlv1.NodeStatus) {
	if node.Conditions[0].Status == core.ConditionFalse {
		if node.Conditions[1].Status == core.ConditionFalse &&
			node.Conditions[2].Status == core.ConditionTrue &&
			node.Conditions[3].Status == core.ConditionTrue {
			node.Healthy = true
		} else if node.Conditions[1].Status == core.ConditionTrue &&
			node.Conditions[2].Status == core.ConditionFalse &&
			node.Conditions[3].Status == core.ConditionFalse {
			node.Healthy = true
		}
	}
}
