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

package controllers

import (
	"context"
	"fmt"
	"time"

	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/cluster"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const maxStatusesQuantity = 10

func (r *ClusterReconciler) updateStatus(ctx context.Context, c *cluster.Cluster, reconcileErr error) (err error) {
	clusterCondition := mysqlv1.ClusterCondition{
		Type:               mysqlv1.ClusterInit,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	c.Status.State = mysqlv1.ClusterInit

	if reconcileErr != nil {
		clusterCondition = mysqlv1.ClusterCondition{
			Type:               mysqlv1.ClusterError,
			Status:             core.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "ErrorReconcile",
			Message:            reconcileErr.Error(),
		}
		c.Status.State = mysqlv1.ClusterError
		return r.writeStatus(ctx, c, clusterCondition)
	}

	if c.Status.ReadyNodes == int(*c.Spec.Replicas) {
		clusterCondition.Type = mysqlv1.ClusterReady
		c.Status.State = mysqlv1.ClusterReady
		return r.writeStatus(ctx, c, clusterCondition)
	}

	list := core.PodList{}
	err = r.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace:     c.Namespace,
			LabelSelector: c.GetLabels().AsSelector(),
		},
	)
	if err != nil {
		return fmt.Errorf("get status: %v", err)
	}

	for _, pod := range list.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == core.PodScheduled && cond.Reason == core.PodReasonUnschedulable &&
				cond.LastTransitionTime.Time.Before(time.Now().Add(-1*time.Minute)) {
				clusterCondition = mysqlv1.ClusterCondition{
					Type:               mysqlv1.ClusterError,
					Status:             core.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             core.PodReasonUnschedulable,
					Message:            cond.Message,
				}
				c.Status.State = mysqlv1.ClusterError
				break
			}
		}
	}

	return r.writeStatus(ctx, c, clusterCondition)
}

func (r *ClusterReconciler) writeStatus(ctx context.Context, c *cluster.Cluster, status mysqlv1.ClusterCondition) error {
	if len(c.Status.Conditions) == 0 {
		c.Status.Conditions = append(c.Status.Conditions, status)
	} else {
		lastCond := c.Status.Conditions[len(c.Status.Conditions)-1]
		if lastCond.Type != status.Type {
			c.Status.Conditions = append(c.Status.Conditions, status)
		}
	}

	if len(c.Status.Conditions) > maxStatusesQuantity {
		c.Status.Conditions = c.Status.Conditions[len(c.Status.Conditions)-maxStatusesQuantity:]
	}

	return r.Status().Update(ctx, c.Unwrap())
}
