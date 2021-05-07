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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

// NewFollowerSVCSyncer returns a service syncer.
func NewFollowerSVCSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	service := &core.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.FollowerService),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
	return syncer.NewObjectSyncer("FollowerSVC", c.Unwrap(), service, cli, func() error {
		service.Spec.Type = "ClusterIP"
		service.Spec.Selector = c.GetSelectorLabels()
		service.Spec.Selector["role"] = "follower"

		if len(service.Spec.Ports) != 1 {
			service.Spec.Ports = make([]core.ServicePort, 1)
		}

		service.Spec.Ports[0].Name = utils.MysqlPortName
		service.Spec.Ports[0].Port = utils.MysqlPort
		service.Spec.Ports[0].TargetPort = intstr.FromInt(utils.MysqlPort)
		return nil
	})
}
