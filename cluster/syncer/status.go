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
	"time"

	"github.com/presslabs/controller-util/syncer"
	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/cluster"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const maxStatusesQuantity = 10
const checkNodeStatusRetry = 3

type StatusUpdater struct {
	*cluster.Cluster

	cli client.Client
}

func NewStatusUpdater(cli client.Client, c *cluster.Cluster) *StatusUpdater {
	return &StatusUpdater{
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
	return syncer.SyncResult{}, s.UpdateNodeStatus(s.cli, readyNodes)
}
