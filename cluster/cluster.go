/*
Copyright 2021.

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

package cluster

import (
	"fmt"

	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
)

type Cluster struct {
	*mysqlv1.Cluster
}

func New(m *mysqlv1.Cluster) *Cluster {
	return &Cluster{
		Cluster: m,
	}
}

// Unwrap returns the api mysqlcluster object
func (c *Cluster) Unwrap() *mysqlv1.Cluster {
	return c.Cluster
}

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

const (
	// HeadlessSVC is the alias of the headless service resource
	HeadlessSVC ResourceName = "headless"
	// StatefulSet is the alias of the statefulset resource
	StatefulSet ResourceName = "mysql"
	// ConfigMap is the alias for mysql configs, the config map resource
	ConfigMap ResourceName = "config-files"
	// MasterService is the name of the service that points to master node
	MasterService ResourceName = "master-service"
	// HealthyReplicasService is the name of a service that points healthy replicas (excludes master)
	HealthyReplicasService ResourceName = "healthy-replicas-service"
	// HealthyNodesService is the name of a service that contains all healthy nodes
	HealthyNodesService ResourceName = "healthy-nodes-service"
	// Secret is the name of the "private" secret that contains operator related credentials
	Secret ResourceName = "secret"
)

// GetNameForResource returns the name of a resource from above
func (c *Cluster) GetNameForResource(name ResourceName) string {
	return GetNameForResource(name, c.Name)
}

// GetNameForResource returns the name of a resource for a cluster
func GetNameForResource(name ResourceName, clusterName string) string {
	switch name {
	case StatefulSet, ConfigMap, HealthyNodesService:
		return fmt.Sprintf("%s-mysql", clusterName)
	case MasterService:
		return fmt.Sprintf("%s-master", clusterName)
	case HealthyReplicasService:
		return fmt.Sprintf("%s-replicas", clusterName)
	case HeadlessSVC:
		return fmt.Sprintf("%s-headless", clusterName)
	case Secret:
		return fmt.Sprintf("%s-secret", clusterName)
	default:
		return fmt.Sprintf("%s-mysql", clusterName)
	}
}
