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

	"github.com/blang/semver"
	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/util"
	"k8s.io/apimachinery/pkg/labels"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var log = logf.Log.WithName("update-status")

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

// GetLabels returns cluster labels
func (c *Cluster) GetLabels() labels.Set {
	instance := c.Name
	if inst, ok := c.Annotations["app.kubernetes.io/instance"]; ok {
		instance = inst
	}

	component := "database"
	if comp, ok := c.Annotations["app.kubernetes.io/component"]; ok {
		component = comp
	}

	labels := labels.Set{
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/instance":   instance,
		"app.kubernetes.io/version":    c.GetMySQLSemVer().String(),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/managed-by": "mysql.radondb.io",
	}

	if part, ok := c.Annotations["app.kubernetes.io/part-of"]; ok {
		labels["app.kubernetes.io/part-of"] = part
	}

	return labels
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

// GetMySQLSemVer returns the MySQL server version in semver format, or the default one
func (c *Cluster) GetMySQLSemVer() semver.Version {
	version := c.Spec.MysqlVersion
	// lookup for an alias, usually this will solve 5.7 to 5.7.x
	if v, ok := util.MySQLTagsToSemVer[version]; ok {
		version = v
	}

	sv, err := semver.Make(version)
	if err != nil {
		log.Error(err, "failed to parse given MySQL version", "input", version)
	}

	// if there is an error will return 0.0.0
	return sv
}

func (c *Cluster) GetPodHostName(p int) string {
	return fmt.Sprintf("%s-%d.%s.%s", c.GetNameForResource(StatefulSet), p,
		c.GetNameForResource(HeadlessSVC),
		c.Namespace)
}

func (c *Cluster) GetOwnHostName() string {
	return fmt.Sprintf("%s.%s.%s", c.ObjectMeta.Name,
		c.GetNameForResource(HeadlessSVC),
		c.Namespace)
}
