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
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

type metrics struct {
	*cluster.Cluster

	name string
}

func (c *metrics) getName() string {
	return c.name
}

func (c *metrics) getImage() string {
	return c.Spec.MetricsOpts.Image
}

func (c *metrics) getCommand() []string {
	return nil
}

func (c *metrics) getEnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		getEnvVarFromSecret(c.GetNameForResource(utils.Secret), "DATA_SOURCE_NAME", "data-source", true),
	}
}

func (c *metrics) getLifecycle() *corev1.Lifecycle {
	return nil
}

func (c *metrics) getResources() corev1.ResourceRequirements {
	return c.Spec.MetricsOpts.Resources
}

func (c *metrics) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.MetricsPortName,
			ContainerPort: utils.MetricsPort,
		},
	}
}

func (c *metrics) getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(utils.MetricsPort),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *metrics) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(utils.MetricsPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *metrics) getVolumeMounts() []corev1.VolumeMount {
	return nil
}
