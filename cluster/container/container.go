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

type container interface {
	getName() string
	getImage() string
	getCommand() []string
	getEnvVars() []corev1.EnvVar
	getLifecycle() *corev1.Lifecycle
	getResources() corev1.ResourceRequirements
	getPorts() []corev1.ContainerPort
	getLivenessProbe() *corev1.Probe
	getReadinessProbe() *corev1.Probe
	getVolumeMounts() []corev1.VolumeMount
}

func EnsureContainer(name string, c *cluster.Cluster) corev1.Container {
	var ctr container
	switch name {
	case utils.ContainerInitSidecarName:
		ctr = &initSidecar{c, name}
	case utils.ContainerInitMysqlName:
		ctr = &initMysql{c, name}
	case utils.ContainerMysqlName:
		ctr = &mysql{c, name}
	case utils.ContainerXenonName:
		ctr = &xenon{c, name}
	case utils.ContainerMetricsName:
		ctr = &metrics{c, name}
	case utils.ContainerSlowLogName:
		ctr = &slowLog{c, name}
	case utils.ContainerAuditLogName:
		ctr = &auditLog{c, name}
	}

	return corev1.Container{
		Name:            ctr.getName(),
		Image:           ctr.getImage(),
		ImagePullPolicy: c.Spec.PodSpec.ImagePullPolicy,
		Command:         ctr.getCommand(),
		Env:             ctr.getEnvVars(),
		Lifecycle:       ctr.getLifecycle(),
		Resources:       ctr.getResources(),
		Ports:           ctr.getPorts(),
		LivenessProbe:   ctr.getLivenessProbe(),
		ReadinessProbe:  ctr.getReadinessProbe(),
		VolumeMounts:    ctr.getVolumeMounts(),
	}
}
