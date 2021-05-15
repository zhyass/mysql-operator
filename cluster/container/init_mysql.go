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

type initMysql struct {
	*cluster.Cluster

	name string
}

func (c *initMysql) getName() string {
	return c.name
}

func (c *initMysql) getImage() string {
	img := utils.MysqlImageVersions[c.GetMySQLVersion()]
	return img
}

func (c *initMysql) getCommand() []string {
	return nil
}

func (c *initMysql) getEnvVars() []core.EnvVar {
	envs := []core.EnvVar{
		{
			Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
			Value: "yes",
		},
		{
			Name:  "MYSQL_ROOT_HOST",
			Value: "127.0.0.1",
		},
		{
			Name:  "MYSQL_INIT_ONLY",
			Value: "1",
		},
	}

	sctName := c.GetNameForResource(utils.Secret)
	envs = append(
		envs,
		getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false),
		getEnvVarFromSecret(sctName, "MYSQL_DATABASE", "mysql-database", true),
		getEnvVarFromSecret(sctName, "MYSQL_USER", "mysql-user", true),
		getEnvVarFromSecret(sctName, "MYSQL_PASSWORD", "mysql-password", true),
	)

	if c.Spec.MysqlOpts.InitTokuDB {
		envs = append(envs, core.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envs
}

func (c *initMysql) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *initMysql) getResources() core.ResourceRequirements {
	return c.Spec.MysqlOpts.Resources
}

func (c *initMysql) getPorts() []core.ContainerPort {
	return nil
}

func (c *initMysql) getLivenessProbe() *core.Probe {
	return nil
}

func (c *initMysql) getReadinessProbe() *core.Probe {
	return nil
}

func (c *initMysql) getVolumeMounts() []core.VolumeMount {
	return []core.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		},
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
		},
		{
			Name:      utils.InitFileVolumeName,
			MountPath: utils.InitFileVolumeMountPath,
		},
	}
}
