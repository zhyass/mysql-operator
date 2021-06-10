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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/internal"
	"github.com/zhyass/mysql-operator/utils"
)

const maxStatusesQuantity = 10
const checkNodeStatusRetry = 3

type StatusUpdater struct {
	log logr.Logger

	*cluster.Cluster

	cli client.Client
}

func NewStatusUpdater(log logr.Logger, cli client.Client, c *cluster.Cluster) *StatusUpdater {
	return &StatusUpdater{
		log:     log,
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
	clusterCondition := apiv1.ClusterCondition{
		Type:               apiv1.ClusterInit,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	s.Status.State = apiv1.ClusterInit

	list := corev1.PodList{}
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
	var readyNodes []corev1.Pod
	for _, pod := range list.Items {
		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case corev1.ContainersReady:
				if cond.Status == corev1.ConditionTrue {
					readyNodes = append(readyNodes, pod)
				}
			case corev1.PodScheduled:
				if cond.Reason == corev1.PodReasonUnschedulable {
					clusterCondition = apiv1.ClusterCondition{
						Type:               apiv1.ClusterError,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(time.Now()),
						Reason:             corev1.PodReasonUnschedulable,
						Message:            cond.Message,
					}
					s.Status.State = apiv1.ClusterError
				}
			}
		}
	}

	s.Status.ReadyNodes = len(readyNodes)
	if s.Status.ReadyNodes == int(*s.Spec.Replicas) {
		s.Status.State = apiv1.ClusterReady
		clusterCondition.Type = apiv1.ClusterReady
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
	return syncer.SyncResult{}, s.updateNodeStatus(ctx, s.cli, readyNodes)
}

func (s *StatusUpdater) updateNodeStatus(ctx context.Context, cli client.Client, pods []corev1.Pod) error {
	sctName := s.GetNameForResource(utils.Secret)
	svcName := s.GetNameForResource(utils.HeadlessSVC)
	port := utils.MysqlPort
	nameSpace := s.Namespace

	secret := &corev1.Secret{}
	if err := cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: nameSpace,
			Name:      sctName,
		},
		secret,
	); err != nil {
		s.log.V(1).Info("secret not found", "name", sctName)
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
		podName := pod.Name
		host := fmt.Sprintf("%s.%s.%s", podName, svcName, nameSpace)
		index := s.getNodeStatusIndex(host)
		node := &s.Status.Nodes[index]
		node.Message = ""

		isLeader, err := checkRole(nameSpace, podName)
		if err != nil {
			s.log.Error(err, "failed to check the node role", "node", node.Name)
			node.Message = err.Error()
		}
		// update apiv1.NodeConditionLeader.
		s.updateNodeCondition(node, 1, isLeader)

		isLagged, isReplicating, isReadOnly := corev1.ConditionUnknown, corev1.ConditionUnknown, corev1.ConditionUnknown
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

		if isLeader == corev1.ConditionTrue && isReadOnly != corev1.ConditionFalse {
			s.log.V(1).Info("try to correct the leader writeable", "node", node.Name)
			correctLeaderReadOnly(nameSpace, podName)
		}

		// update apiv1.NodeConditionLagged.
		s.updateNodeCondition(node, 0, isLagged)
		// update apiv1.NodeConditionReplicating.
		s.updateNodeCondition(node, 3, isReplicating)
		// update apiv1.NodeConditionReadOnly.
		s.updateNodeCondition(node, 2, isReadOnly)

		if err = setPodHealthy(ctx, cli, &pod, node); err != nil {
			s.log.Error(err, "cannot update pod", "name", podName, "namespace", pod.Namespace)
		}
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
	status := apiv1.NodeStatus{
		Name: name,
		Conditions: []apiv1.NodeCondition{
			{
				Type:               apiv1.NodeConditionLagged,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1.NodeConditionLeader,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1.NodeConditionReadOnly,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1.NodeConditionReplicating,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
		},
	}
	s.Status.Nodes = append(s.Status.Nodes, status)
	return len
}

func (s *StatusUpdater) updateNodeCondition(node *apiv1.NodeStatus, idx int, status corev1.ConditionStatus) {
	if node.Conditions[idx].Status != status {
		t := time.Now()
		s.log.V(3).Info(fmt.Sprintf("Found status change for node %q condition %q: %q -> %q; setting lastTransitionTime to %v",
			node.Name, node.Conditions[idx].Type, node.Conditions[idx].Status, status, t))
		node.Conditions[idx].Status = status
		node.Conditions[idx].LastTransitionTime = metav1.NewTime(t)
	}
}

func checkRole(namespace, podName string) (corev1.ConditionStatus, error) {
	command := []string{"xenoncli", "raft", "status"}
	status := corev1.ConditionUnknown
	executor, err := internal.NewPodExecutor()
	if err != nil {
		return status, err
	}

	stdout, stderr, err := executor.Exec(namespace, podName, "xenon", command...)
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
		return corev1.ConditionTrue, nil
	}

	if out["state"] == "FOLLOWER" {
		return corev1.ConditionFalse, nil
	}

	return status, nil
}

func correctLeaderReadOnly(namespace, podName string) error {
	executor, err := internal.NewPodExecutor()
	if err != nil {
		return err
	}

	err = executor.SetGlobalSysVar(namespace, podName, "SET GLOBAL read_only=off")
	if err != nil {
		return err
	}

	return executor.SetGlobalSysVar(namespace, podName, "SET GLOBAL super_read_only=off")
}

func setPodHealthy(ctx context.Context, cli client.Client, pod *corev1.Pod, node *apiv1.NodeStatus) error {
	healthy := "no"
	if node.Conditions[0].Status == corev1.ConditionFalse {
		if node.Conditions[1].Status == corev1.ConditionFalse &&
			node.Conditions[2].Status == corev1.ConditionTrue &&
			node.Conditions[3].Status == corev1.ConditionTrue {
			healthy = "yes"
		} else if node.Conditions[1].Status == corev1.ConditionTrue &&
			node.Conditions[2].Status == corev1.ConditionFalse &&
			node.Conditions[3].Status == corev1.ConditionFalse {
			healthy = "yes"
		}
	}

	if pod.Labels["healthy"] != healthy {
		pod.Labels["healthy"] = healthy
		if err := cli.Update(ctx, pod); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}
