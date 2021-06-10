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

	corev1 "k8s.io/api/core/v1"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
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

func (c *xenon) getEnvVars() []corev1.EnvVar {
	return nil
}

func (c *xenon) getLifecycle() *corev1.Lifecycle {
	arg := fmt.Sprintf("until (xenoncli xenon ping && xenoncli cluster add %s) > /dev/null 2>&1; do sleep 2; done", c.CreatePeers())
	return &corev1.Lifecycle{
		PostStart: &corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", arg},
			},
		},
	}
}

func (c *xenon) getResources() corev1.ResourceRequirements {
	return c.Spec.XenonOpts.Resources
}

func (c *xenon) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.XenonPortName,
			ContainerPort: utils.XenonPort,
		},
	}
}

func (c *xenon) getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
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

func (c *xenon) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
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

func (c *xenon) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
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
