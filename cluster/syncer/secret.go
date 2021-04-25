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
	"github.com/presslabs/controller-util/rand"
	"github.com/presslabs/controller-util/syncer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhyass/mysql-operator/cluster"
	"github.com/zhyass/mysql-operator/utils"
)

const (
	rStrLen = 12
)

func NewSecretSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	secret := &core.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.Secret),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", c.Unwrap(), secret, cli, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		if c.Spec.MetricsOpts.Enabled {
			secret.Data["metrics-user"] = []byte("qc_metrics")
			if err := addRandomPassword(secret.Data, "metrics-password"); err != nil {
				return err
			}
		}

		secret.Data["replication-user"] = []byte("qc_repl")
		if err := addRandomPassword(secret.Data, "replication-password"); err != nil {
			return err
		}

		secret.Data["root-password"] = []byte(c.Spec.MysqlOpts.RootPassword)

		secret.Data["mysql-user"] = []byte(c.Spec.MysqlOpts.User)
		secret.Data["mysql-password"] = []byte(c.Spec.MysqlOpts.Password)
		secret.Data["mysql-database"] = []byte(c.Spec.MysqlOpts.Database)
		return nil
	})
}

// addRandomPassword checks if a key exists and if not registers a random string for that key
func addRandomPassword(data map[string][]byte, key string) error {
	if len(data[key]) == 0 {
		// NOTE: use only alpha-numeric string, this strings are used unescaped in MySQL queries (issue #314)
		random, err := rand.AlphaNumericString(rStrLen)
		if err != nil {
			return err
		}
		data[key] = []byte(random)
	}
	return nil
}
