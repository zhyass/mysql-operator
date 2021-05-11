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
	"strconv"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
	core "k8s.io/api/core/v1"
)

type initSidecar struct {
	*cluster.Cluster

	name string
}

func (c *initSidecar) getName() string {
	return c.name
}

func (c *initSidecar) getImage() string {
	return c.Spec.PodSpec.SidecarImage
}

func (c *initSidecar) getArgs() []string {
	return []string{"init"}
}

func (c *initSidecar) getEnvVars() []core.EnvVar {
	sctName := c.GetNameForResource(utils.Secret)
	envs := []core.EnvVar{
		{
			Name: "POD_HOSTNAME",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  "NAMESPACE",
			Value: c.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: c.GetNameForResource(utils.HeadlessSVC),
		},
		{
			Name:  "ADMIT_DEFEAT_HEARBEAT_COUNT",
			Value: strconv.Itoa(int(*c.Spec.XenonOpts.AdmitDefeatHearbeatCount)),
		},
		{
			Name:  "ELECTION_TIMEOUT",
			Value: strconv.Itoa(int(*c.Spec.XenonOpts.ElectionTimeout)),
		},
		{
			Name:  "MY_MYSQL_VERSION",
			Value: c.GetMySQLVersion(),
		},
		getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false),
		getEnvVarFromSecret(sctName, "MYSQL_REPL_USER", "replication-user", true),
		getEnvVarFromSecret(sctName, "MYSQL_REPL_PASSWORD", "replication-password", true),
	}

	if c.Spec.MetricsOpts.Enabled {
		envs = append(
			envs,
			getEnvVarFromSecret(sctName, "METRICS_USER", "metrics-user", true),
			getEnvVarFromSecret(sctName, "METRICS_PASSWORD", "metrics-password", true),
		)
	}

	if c.Spec.MysqlOpts.InitTokuDB {
		envs = append(envs, core.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envs
}

func (c *initSidecar) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *initSidecar) getResources() core.ResourceRequirements {
	return c.Spec.PodSpec.Resources
}

func (c *initSidecar) getPorts() []core.ContainerPort {
	return nil
}

func (c *initSidecar) getLivenessProbe() *core.Probe {
	return nil
}

func (c *initSidecar) getReadinessProbe() *core.Probe {
	return nil
}

func (c *initSidecar) getVolumeMounts() []core.VolumeMount {
	volumeMounts := []core.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.ConfMapVolumeName,
			MountPath: utils.ConfMapVolumeMountPath,
		},
		{
			Name:      utils.ScriptsVolumeName,
			MountPath: utils.ScriptsVolumeMountPath,
		},
		{
			Name:      utils.XenonVolumeName,
			MountPath: utils.XenonVolumeMountPath,
		},
		{
			Name:      utils.InitFileVolumeName,
			MountPath: utils.InitFileVolumeMountPath,
		},
	}

	if c.Spec.MysqlOpts.InitTokuDB {
		volumeMounts = append(volumeMounts,
			core.VolumeMount{
				Name:      utils.SysVolumeName,
				MountPath: utils.SysVolumeMountPath,
			},
		)
	}

	if c.Spec.Persistence.Enabled {
		volumeMounts = append(volumeMounts,
			core.VolumeMount{
				Name:      utils.DataVolumeName,
				MountPath: utils.DataVolumeMountPath,
			},
		)
	}

	return volumeMounts
}
