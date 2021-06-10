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
	"github.com/presslabs/controller-util/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

// NewHeadlessSVCSyncer returns a service syncer.
func NewHeadlessSVCSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.HeadlessSVC),
			Namespace: c.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "mysql",
				"app.kubernetes.io/managed-by": "mysql.radondb.io",
			},
		},
	}

	return syncer.NewObjectSyncer("HeadlessSVC", c.Unwrap(), service, cli, func() error {
		service.Spec.Type = "ClusterIP"
		service.Spec.ClusterIP = "None"
		service.Spec.Selector = labels.Set{
			"app.kubernetes.io/name":       "mysql",
			"app.kubernetes.io/managed-by": "mysql.radondb.io",
		}

		// Use `publishNotReadyAddresses` to be able to access pods even if the pod is not ready.
		service.Spec.PublishNotReadyAddresses = true

		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       utils.MysqlPortName,
				Port:       utils.MysqlPort,
				TargetPort: intstr.FromInt(utils.MysqlPort),
			},
		}

		if c.Spec.MetricsOpts.Enabled {
			service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
				Name:       utils.MetricsPortName,
				Port:       utils.MetricsPort,
				TargetPort: intstr.FromInt(utils.MetricsPort),
			})
		}
		return nil
	})
}
