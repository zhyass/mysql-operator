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

type mysql struct {
	*cluster.Cluster

	name string
}

func (c *mysql) getName() string {
	return c.name
}

func (c *mysql) getImage() string {
	img := utils.MysqlImageVersions[c.GetMySQLVersion()]
	return img
}

func (c *mysql) getCommand() []string {
	return nil
}

func (c *mysql) getEnvVars() []core.EnvVar {
	sctName := c.GetNameForResource(utils.Secret)

	envs := []core.EnvVar{
		getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false),
		getEnvVarFromSecret(sctName, "MYSQL_REPL_USER", "replication-user", true),
		getEnvVarFromSecret(sctName, "MYSQL_REPL_PASSWORD", "replication-password", true),
		getEnvVarFromSecret(sctName, "MYSQL_USER", "mysql-user", true),
		getEnvVarFromSecret(sctName, "MYSQL_PASSWORD", "mysql-password", true),
		getEnvVarFromSecret(sctName, "MYSQL_DATABASE", "mysql-database", true),
		getEnvVarFromSecret(sctName, "METRICS_USER", "metrics-user", true),
		getEnvVarFromSecret(sctName, "METRICS_PASSWORD", "metrics-password", true),
	}

	if c.Spec.MysqlOpts.InitTokuDB {
		envs = append(envs, core.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envs
}

func (c *mysql) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *mysql) getResources() core.ResourceRequirements {
	return c.Spec.MysqlOpts.Resources
}

func (c *mysql) getPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          utils.MysqlPortName,
			ContainerPort: utils.MysqlPort,
		},
	}
}

func (c *mysql) getLivenessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"sh", "-c", "mysqladmin ping -uroot -p${MYSQL_ROOT_PASSWORD}"},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *mysql) getReadinessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"sh", "-c", `mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SELECT 1"`},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *mysql) getVolumeMounts() []core.VolumeMount {
	return []core.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: "/etc/mysql/conf.d",
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: "/var/lib/mysql",
		},
		{
			Name:      utils.LogsVolumeName,
			MountPath: "/var/log/mysql",
		},
	}
}
