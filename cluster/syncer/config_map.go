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

	"github.com/blang/semver"
	"github.com/go-ini/ini"
	"github.com/presslabs/controller-util/syncer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

func NewConfigMapSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	cm := &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.ConfigMap),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", c.Unwrap(), cm, cli, func() error {
		data, err := buildMysqlConf(c)
		if err != nil {
			return fmt.Errorf("failed to create mysql configs: %s", err)
		}

		cm.Data = map[string]string{
			"node.cnf":   data,
			"xenon.json": buildXenonConf(c),
		}

		return nil
	})
}

func buildMysqlConf(c *cluster.Cluster) (string, error) {
	cfg := ini.Empty()
	sec := cfg.Section("mysqld")

	addKVConfigsToSection(sec, convertMapToKVConfig(mysqlCommonConfigs), c.Spec.MysqlOpts.MysqlConf)
	addKVConfigsToSection(sec, convertMapToKVConfig(mysqlStaticConfigs), c.Spec.MysqlOpts.MysqlConf)

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

func buildXenonConf(c *cluster.Cluster) string {
	admitDefeatHearbeatCount := *c.Spec.XenonOpts.AdmitDefeatHearbeatCount
	electionTimeout := *c.Spec.XenonOpts.ElectionTimeout
	pingTimeout := electionTimeout / admitDefeatHearbeatCount
	heartbeatTimeout := electionTimeout / admitDefeatHearbeatCount
	requestTimeout := electionTimeout / admitDefeatHearbeatCount

	host := c.GetOwnHostName()

	version := "mysql80"
	sv, err := semver.Make(c.GetMySQLVersion())
	if err != nil {
		log.Error(err, "failed to parse given MySQL version", "input", c.GetMySQLVersion())
	}
	if sv.Major == 5 {
		if sv.Minor == 6 {
			version = "mysql56"
		} else {
			version = "mysql57"
		}
	}

	return fmt.Sprintf(`{
	"log": {
		"level": "INFO"
	},
	"server": {
		"endpoint": "%s:%d"
	},
	"replication": {
		"passwd": "@@REPL_PASSWD@@",
		"user": "@@REPL_USER@@"
	},
	"rpc": {
		"request-timeout": %d
	},
	"mysql": {
		"admit-defeat-ping-count": 3,
		"admin": "root",
		"ping-timeout": %d,
		"passwd": "@@ROOT_PASSWD@@",
		"host": "localhost",
		"version": "%s",
		"master-sysvars": "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default",
		"slave-sysvars": "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1",
		"port": 3306,
		"monitor-disabled": true
	},
	"raft": {
		"election-timeout": %d,
		"admit-defeat-hearbeat-count": %d,
		"heartbeat-timeout": %d,
		"meta-datadir": "/var/lib/xenon/",
		"semi-sync-degrade": true,
		"purge-binlog-disabled": true,
		"super-idle": false
	}
}
`, host, utils.XenonPort, requestTimeout, pingTimeout, version, electionTimeout, admitDefeatHearbeatCount, heartbeatTimeout)
}
