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

func (c *xenon) getArgs() []string {
	return nil
}

func (c *xenon) getEnvVars() []core.EnvVar {
	return nil
}

func (c *xenon) getLifecycle() *core.Lifecycle {
	arg := fmt.Sprintf("until (xenoncli xenon ping && xenoncli cluster add %s) > /dev/null 2>&1; do sleep 2; done", c.CreatePeers())
	return &core.Lifecycle{
		PostStart: &core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"sh", "-c", arg},
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
			Name:      utils.ScriptsVolumeName,
			MountPath: utils.ScriptsVolumeMountPath,
		},
		{
			Name:      utils.XenonVolumeName,
			MountPath: utils.XenonVolumeMountPath,
		},
	}
}
