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
	"bytes"
	"fmt"
	"sort"

	"github.com/go-ini/ini"
	"github.com/presslabs/controller-util/syncer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhyass/mysql-operator/cluster"
)

func NewConfigMapSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	cm := &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(cluster.ConfigMap),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", c.Unwrap(), cm, cli, func() error {
		cm.ObjectMeta.Labels = c.GetLabels()

		data, err := buildMysqlConf(c)
		if err != nil {
			return fmt.Errorf("failed to create mysql configs: %s", err)
		}

		cm.Data = map[string]string{
			"node.cnf": data,
		}

		// TODO: xenon.json
		return nil
	})
}

func buildMysqlConf(c *cluster.Cluster) (string, error) {
	cfg := ini.Empty()
	sec := cfg.Section("mysqld")

	addKVConfigsToSection(sec, convertMapToKVConfig(mysqlCommonConfigs), c.Spec.MysqlConf)
	addKVConfigsToSection(sec, convertMapToKVConfig(mysqlStaticConfigs), c.Spec.MysqlConf)

	data, err := writeConfigs(cfg)
	if err != nil {
		return "", err
	}

	return data, nil
}

// addKVConfigsToSection add a map[string]string to a ini.Section
func addKVConfigsToSection(s *ini.Section, extraMysqld ...map[string]intstr.IntOrString) {
	for _, extra := range extraMysqld {
		keys := []string{}
		for key := range extra {
			keys = append(keys, key)
		}

		// sort keys
		sort.Strings(keys)

		for _, k := range keys {
			value := extra[k]
			if _, err := s.NewKey(k, value.String()); err != nil {
				log.Error(err, "failed to add key to config section", "key", k, "value", extra[k], "section", s)
			}
		}
	}
}

func convertMapToKVConfig(m map[string]string) map[string]intstr.IntOrString {
	config := make(map[string]intstr.IntOrString)

	for key, value := range m {
		config[key] = intstr.Parse(value)
	}

	return config
}

// writeConfigs write to string ini.File
// nolint: interfacer
func writeConfigs(cfg *ini.File) (string, error) {
	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
