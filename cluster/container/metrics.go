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
	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	return []string{"sh", "-c", `DATA_SOURCE_NAME="$METRICS_USER:$METRICS_PASSWORD@(localhost:3306)/" /bin/mysqld_exporter`}
}

func (c *metrics) getEnvVars() []core.EnvVar {
	sctName := c.GetNameForResource(utils.Secret)
	metricsUsr := getEnvVarFromSecret(sctName, "METRICS_USER", "metrics-user", true)
	metricsPwd := getEnvVarFromSecret(sctName, "METRICS_PASSWORD", "metrics-password", true)
	return []core.EnvVar{metricsUsr, metricsPwd}
}

func (c *metrics) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *metrics) getResources() core.ResourceRequirements {
	return c.Spec.MetricsOpts.Resources
}

func (c *metrics) getPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          utils.MetricsPortName,
			ContainerPort: utils.MetricsPort,
		},
	}
}

func (c *metrics) getLivenessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			HTTPGet: &core.HTTPGetAction{
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

func (c *metrics) getReadinessProbe() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			HTTPGet: &core.HTTPGetAction{
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

func (c *metrics) getVolumeMounts() []core.VolumeMount {
	return nil
}
