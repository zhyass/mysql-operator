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

type initMysql struct {
	*cluster.Cluster

	name string
}

func (c *initMysql) getName() string {
	return c.name
}

func (c *initMysql) getImage() string {
	return c.Spec.PodSpec.BusyboxImage
}

func (c *initMysql) getCommand() []string {
	str := `# Generate mysql server-id from pod ordinal index.
ordinal=$(echo $(hostname) | tr -cd "[0-9]")
# Copy server-id.conf adding offset to avoid reserved server-id=0 value.
cat /mnt/config-map/server-id.cnf | sed s/@@SERVER_ID@@/$((100 + $ordinal))/g > /mnt/conf.d/server-id.cnf
# Copy appropriate conf.d files from config-map to config mount.
cp -f /mnt/config-map/node.cnf /mnt/conf.d/
# remove lost+found.
rm -rf /mnt/data/lost+found
`
	if c.Spec.MysqlOpts.InitTokudb {
		str = fmt.Sprintf(`%s# For install tokudb.
printf '\nloose_tokudb_directio = ON\n' >> /mnt/conf.d/node.cnf
echo never > /host-sys/kernel/mm/transparent_hugepage/enabled
`, str)
	}
	return []string{"sh", "-c", str}
}

func (c *initMysql) getEnvVars() []core.EnvVar {
	return nil
}

func (c *initMysql) getLifecycle() *core.Lifecycle {
	return nil
}

func (c *initMysql) getResources() core.ResourceRequirements {
	return c.Spec.PodSpec.Resources
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
			MountPath: "/mnt/conf.d",
		},
		{
			Name:      utils.ConfMapVolumeName,
			MountPath: "/mnt/config-map",
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: "/mnt/data",
		},
		{
			Name:      utils.SysVolumeName,
			MountPath: "/host-sys",
		},
	}
}
