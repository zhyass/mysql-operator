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

package syncer

import (
	"fmt"

	"github.com/presslabs/controller-util/syncer"
	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewStatefulSetSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	obj := &apps.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(cluster.StatefulSet),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("StatefulSet", c.Unwrap(), obj, cli, func() error {
		c.Status.ReadyNodes = int(obj.Status.ReadyReplicas)

		obj.Spec.ServiceName = c.GetNameForResource(cluster.StatefulSet)
		obj.Spec.Replicas = c.Spec.Replicas
		obj.Spec.Selector = metav1.SetAsLabelSelector(c.GetSelectorLabels())

		obj.Spec.Template.ObjectMeta.Labels = getLabels(c)
		obj.Spec.Template.Annotations = c.Spec.PodSpec.Annotations
		if len(obj.Spec.Template.ObjectMeta.Annotations) == 0 {
			obj.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
		}
		obj.Spec.Template.ObjectMeta.Annotations["prometheus.io/scrape"] = "true"
		obj.Spec.Template.ObjectMeta.Annotations["prometheus.io/port"] = fmt.Sprintf("%d", utils.MetricsPort)

		return nil
	})
}

func getLabels(c *cluster.Cluster) map[string]string {
	labels := c.GetLabels()
	for k, v := range c.Spec.PodSpec.Labels {
		labels[k] = v
	}
	return labels
}

func ensurePodSpec(c *cluster.Cluster) core.PodSpec {
	return core.PodSpec{
		InitContainers:     ensureInitContainers(c),
		Containers:         nil,
		Volumes:            nil,
		SchedulerName:      c.Spec.PodSpec.SchedulerName,
		ServiceAccountName: c.Spec.PodSpec.ServiceAccountName,
		Affinity:           c.Spec.PodSpec.Affinity,
		PriorityClassName:  c.Spec.PodSpec.PriorityClassName,
		Tolerations:        c.Spec.PodSpec.Tolerations,
	}
}

func ensureInitContainers(c *cluster.Cluster) []core.Container {
	initMysql := core.Container{
		Name:         containerInitMysqlName,
		Image:        c.Spec.InitOpts.Image,
		Resources:    getResources(containerInitMysqlName, c),
		Command:      []string{"sh", "-c"},
		Args:         buildInitMysqlArgs(c),
		VolumeMounts: getVolumeMounts(containerInitMysqlName),
	}
	return []core.Container{initMysql}
}

func getResources(name string, c *cluster.Cluster) core.ResourceRequirements {
	switch name {
	case containerInitMysqlName:
		return c.Spec.InitOpts.Resources
	case containerMysqlName:
		return c.Spec.MysqlOpts.Resources
	case containerXenonName:
		return c.Spec.XenonOpts.Resources
	}
	return c.Spec.PodSpec.Resources
}

func getVolumeMounts(name string) []core.VolumeMount {
	var mounts []core.VolumeMount
	switch name {
	case containerInitMysqlName:
		mounts = []core.VolumeMount{
			{
				Name:      confVolumeName,
				MountPath: "/mnt/conf.d",
			},
			{
				Name:      scriptsVolumeName,
				MountPath: "/mnt/scripts",
			},
			{
				Name:      confMapVolumeName,
				MountPath: "/mnt/config-map",
			},
			{
				Name:      dataVolumeName,
				MountPath: "/mnt/data",
			},
			{
				Name:      sysVolumeName,
				MountPath: "/host-sys",
			},
		}
	}
	return mounts
}

func buildInitMysqlArgs(c *cluster.Cluster) []string {
	str := `# Generate mysql server-id from pod ordinal index.
ordinal=$(echo $(hostname) | tr -cd "[0-9]")
# Copy server-id.conf adding offset to avoid reserved server-id=0 value.
cat /mnt/config-map/server-id.cnf | sed s/@@SERVER_ID@@/$((100 + $ordinal))/g > /mnt/conf.d/server-id.cnf
# Copy appropriate conf.d files from config-map to config mount.
cp -f /mnt/config-map/node.cnf /mnt/conf.d/
cp -f /mnt/config-map/*.sh /mnt/scripts/
chmod +x /mnt/scripts/*
# remove lost+found.
rm -rf /mnt/data/lost+found
`
	if c.Spec.MysqlOpts.InitTokudb {
		str = fmt.Sprintf(`%s# For install tokudb.
printf '\nloose_tokudb_directio = ON\n' >> /mnt/conf.d/node.cnf
echo never > /host-sys/kernel/mm/transparent_hugepage/enabled
`, str)
	}
	return []string{str}
}