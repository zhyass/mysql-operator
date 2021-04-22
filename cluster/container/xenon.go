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

package container

import (
	"fmt"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
	core "k8s.io/api/core/v1"
)

type xenon struct {
	*cluster.Cluster

	name string
}

func (c *xenon) getName() string {
	return c.name
}

func (c *xenon) getImage() string {
	return c.Spec.XenonOpts.Image
}

func (c *xenon) getCommand() []string {
	return nil
}

func (c *xenon) getEnvVars() []core.EnvVar {
	sctName := c.GetNameForResource(cluster.Secret)

	rootPwd := getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false)
	replUser := getEnvVarFromSecret(sctName, "MYSQL_REPL_USER", "replication-user", true)
	replPwd := getEnvVarFromSecret(sctName, "MYSQL_REPL_PASSWORD", "replication-password", true)
	podHostName := core.EnvVar{
		Name: "POD_HOSTNAME",
		ValueFrom: &core.EnvVarSource{
			FieldRef: &core.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}
	host := core.EnvVar{
		Name:  "HOST",
		Value: fmt.Sprintf("$(POD_HOSTNAME).%s.%s", c.GetNameForResource(cluster.HeadlessSVC), c.Namespace),
	}

	env := []core.EnvVar{rootPwd, replUser, replPwd, podHostName, host}

	if c.Spec.MysqlOpts.InitTokudb {
		env = append(env, core.EnvVar{
			Name:  "Master_SysVars",
			Value: "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default",
		})
		env = append(env, core.EnvVar{
			Name:  "Slave_SysVars",
			Value: "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1",
		})
	} else {
		env = append(env, core.EnvVar{
			Name:  "Master_SysVars",
			Value: "sync_binlog=default;innodb_flush_log_at_trx_commit=default",
		})
		env = append(env, core.EnvVar{
			Name:  "Slave_SysVars",
			Value: "sync_binlog=1000;innodb_flush_log_at_trx_commit=1",
		})
	}
	return env
}

func (c *xenon) getLifecycle() *core.Lifecycle {
	return &core.Lifecycle{
		PostStart: &core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"sh", "-c", `until (xenoncli xenon ping && xenoncli cluster add "$(/scripts/create-peers.sh)") > /dev/null 2>&1; do sleep 2; done`},
			},
		},
	}
}

func (c *xenon) getResources() core.ResourceRequirements {
	return c.Spec.XenonOpts.Resources
}

func (c *xenon) getPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          utils.XenonPortName,
			ContainerPort: utils.XenonPort,
		},
	}
}

func (c *xenon) getLivenessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"pgrep", "xenon"},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *xenon) getReadinessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"sh", "-c", "xenoncli xenon ping"},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *xenon) getVolumeMounts() []core.VolumeMount {
	return []core.VolumeMount{
		{
			Name:      scriptsVolumeName,
			MountPath: "/scripts",
		},
	}
}
