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

package container

import (
	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
	core "k8s.io/api/core/v1"
)

type slowLog struct {
	*cluster.Cluster

	name string
}

func (c *slowLog) getName() string {
	return c.name
}

func (c *slowLog) getImage() string {
	return c.Spec.PodSpec.BusyboxImage
}

func (c *slowLog) getCommand() []string {
	return []string{"tail", "-f", "/logs/mysql-slow.log"}
}

func (c *slowLog) getEnvVars() []core.EnvVar {
	return nil
}

func (c *slowLog) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *slowLog) getResources() core.ResourceRequirements {
	return c.Spec.PodSpec.Resources
}

func (c *slowLog) getPorts() []core.ContainerPort {
	return nil
}

func (c *slowLog) getLivenessProbe() *core.Probe {
	return nil
}

func (c *slowLog) getReadinessProbe() *core.Probe {
	return nil
}

func (c *slowLog) getVolumeMounts() []core.VolumeMount {
	return []core.VolumeMount{
		{
			Name:      utils.LogsVolumeName,
			MountPath: "/logs",
		},
	}
}
