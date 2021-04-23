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

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mysqlv1 "github.com/zhyass/mysql-operator/api/v1"
	"github.com/zhyass/mysql-operator/utils"
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
		"mysql.radondb.io/cluster":     c.Name,
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/instance":   instance,
		"app.kubernetes.io/version":    c.GetMySQLVersion(),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/managed-by": "mysql.radondb.io",
	}

	if part, ok := c.Annotations["app.kubernetes.io/part-of"]; ok {
		labels["app.kubernetes.io/part-of"] = part
	}

	return labels
}

// GetSelectorLabels returns the labels that will be used as selector
func (c *Cluster) GetSelectorLabels() labels.Set {
	return labels.Set{
		"mysql.radondb.io/cluster":     c.Name,
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/managed-by": "mysql.radondb.io",
	}
}

// GetMySQLVersion returns the MySQL server version.
func (c *Cluster) GetMySQLVersion() string {
	version := c.Spec.MysqlVersion
	// lookup for an alias, usually this will solve 5.7 to 5.7.x
	if v, ok := utils.MySQLTagsToSemVer[version]; ok {
		version = v
	}

	if _, ok := utils.MysqlImageVersions[version]; !ok {
		version = utils.MySQLDefaultVersion
	}

	return version
}

func (c *Cluster) GetOwnHostName() string {
	return fmt.Sprintf("%s.%s.%s", c.ObjectMeta.Name,
		c.GetNameForResource(HeadlessSVC),
		c.Namespace)
}

func (c *Cluster) CreatePeers() string {
	str := ""
	for i := 0; i < int(*c.Spec.Replicas); i++ {
		if i > 0 {
			str += ","
		}
		str += c.GetPodHostName(i)
	}
	return str
}

func (c *Cluster) GetPodHostName(p int) string {
	return fmt.Sprintf("%s-%d.%s.%s", c.GetNameForResource(StatefulSet), p,
		c.GetNameForResource(HeadlessSVC),
		c.Namespace)
}

func (c *Cluster) EnsureVolumes() []core.Volume {
	var volumes []core.Volume
	if !c.Spec.Persistence.Enabled {
		volumes = append(volumes, core.Volume{
			Name: utils.DataVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		})
	}

	volumes = append(volumes,
		core.Volume{
			Name: utils.ConfVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		core.Volume{
			Name: utils.LogsVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		core.Volume{
			Name: utils.ConfMapVolumeName,
			VolumeSource: core.VolumeSource{
				ConfigMap: &core.ConfigMapVolumeSource{
					LocalObjectReference: core.LocalObjectReference{
						Name: c.GetNameForResource(ConfigMap),
					},
				},
			},
		},
		core.Volume{
			Name: utils.SysVolumeName,
			VolumeSource: core.VolumeSource{
				HostPath: &core.HostPathVolumeSource{
					Path: "/sys",
				},
			},
		},
	)
	return volumes
}

func (c *Cluster) EnsureVolumeClaimTemplates() []core.PersistentVolumeClaim {
	if !c.Spec.Persistence.Enabled {
		return nil
	}

	storageClassName := ""
	if c.Spec.Persistence.StorageClass != "" {
		storageClassName = c.Spec.Persistence.StorageClass
	}
	data := core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.DataVolumeName,
		},
		Spec: core.PersistentVolumeClaimSpec{
			AccessModes: c.Spec.Persistence.AccessModes,
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: resource.MustParse(c.Spec.Persistence.Size),
				},
			},
			StorageClassName: &storageClassName,
		},
	}
	return []core.PersistentVolumeClaim{data}
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
	// SlaveService is the name of a service that points healthy slaves (excludes master)
	SlaveService ResourceName = "slave-service"
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
	case StatefulSet, ConfigMap, HealthyNodesService, HeadlessSVC:
		return fmt.Sprintf("%s-mysql", clusterName)
	case MasterService:
		return fmt.Sprintf("%s-master", clusterName)
	case SlaveService:
		return fmt.Sprintf("%s-slave", clusterName)
	case Secret:
		return fmt.Sprintf("%s-secret", clusterName)
	default:
		return fmt.Sprintf("%s-mysql", clusterName)
	}
}
