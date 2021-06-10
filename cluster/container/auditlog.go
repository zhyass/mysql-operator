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
	corev1 "k8s.io/api/core/v1"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

type auditLog struct {
	*cluster.Cluster

	name string
}

func (c *auditLog) getName() string {
	return c.name
}

func (c *auditLog) getImage() string {
	return c.Spec.PodSpec.BusyboxImage
}

func (c *auditLog) getCommand() []string {
	return []string{"tail", "-f", utils.LogsVolumeMountPath + "/mysql-audit.log"}
}

func (c *auditLog) getEnvVars() []corev1.EnvVar {
	return nil
}

func (c *auditLog) getLifecycle() *corev1.Lifecycle {
	return nil
}

func (c *auditLog) getResources() corev1.ResourceRequirements {
	return c.Spec.PodSpec.Resources
}

func (c *auditLog) getPorts() []corev1.ContainerPort {
	return nil
}

func (c *auditLog) getLivenessProbe() *corev1.Probe {
	return nil
}

func (c *auditLog) getReadinessProbe() *corev1.Probe {
	return nil
}

func (c *auditLog) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
		},
	}
}
