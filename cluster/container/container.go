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

type container interface {
	getName() string
	getImage() string
	getArgs() []string
	getEnvVars() []core.EnvVar
	getLifecycle() *core.Lifecycle
	getResources() core.ResourceRequirements
	getPorts() []core.ContainerPort
	getLivenessProbe() *core.Probe
	getReadinessProbe() *core.Probe
	getVolumeMounts() []core.VolumeMount
}

func EnsureContainer(name string, c *cluster.Cluster) core.Container {
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
	}

	return core.Container{
		Name:            ctr.getName(),
		Image:           ctr.getImage(),
		ImagePullPolicy: c.Spec.PodSpec.ImagePullPolicy,
		Args:            ctr.getArgs(),
		Env:             ctr.getEnvVars(),
		Lifecycle:       ctr.getLifecycle(),
		Resources:       ctr.getResources(),
		Ports:           ctr.getPorts(),
		LivenessProbe:   ctr.getLivenessProbe(),
		ReadinessProbe:  ctr.getReadinessProbe(),
		VolumeMounts:    ctr.getVolumeMounts(),
	}
}
