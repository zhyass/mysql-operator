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

package sidecar

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/zhyass/mysql-operator/utils"
)

type Config struct {
	HostName    string
	NameSpace   string
	ServiceName string

	// root password
	RootPassword string

	// replication user and password
	ReplicationUser     string
	ReplicationPassword string

	InitTokuDB bool

	MySQLVersion semver.Version

	AdmitDefeatHearbeatCount int32
	ElectionTimeout          int32
}

func NewConfig() *Config {
	mysqlVersion, err := semver.Parse(getEnvValue("MYSQL_VERSION"))
	if err != nil {
		log.Info("MYSQL_VERSION is not a semver version")
		mysqlVersion, _ = semver.Parse(utils.MySQLDefaultVersion)
	}

	initTokuDB := false
	if len(getEnvValue("INIT_TOKUDB")) > 0 {
		initTokuDB = true
	}

	admitDefeatHearbeatCount, err := strconv.ParseInt(getEnvValue("ADMIT_DEFEAT_HEARBEAT_COUNT"), 10, 32)
	if err != nil {
		admitDefeatHearbeatCount = 5
	}
	electionTimeout, err := strconv.ParseInt(getEnvValue("ELECTION_TIMEOUT"), 10, 32)
	if err != nil {
		electionTimeout = 10000
	}

	return &Config{
		HostName:    getEnvValue("POD_HOSTNAME"),
		NameSpace:   getEnvValue("NAMESPACE"),
		ServiceName: getEnvValue("SERVICE_NAME"),

		RootPassword: getEnvValue("MYSQL_ROOT_PASSWORD"),

		ReplicationUser:     getEnvValue("MYSQL_REPL_USER"),
		ReplicationPassword: getEnvValue("MYSQL_REPL_PASSWORD"),

		InitTokuDB: initTokuDB,

		MySQLVersion: mysqlVersion,

		AdmitDefeatHearbeatCount: int32(admitDefeatHearbeatCount),
		ElectionTimeout:          int32(electionTimeout),
	}
}

// Generate mysql server-id from pod ordinal index.
func (cfg *Config) generateServerID() int {
	l := strings.Split(cfg.HostName, "-")
	for i := len(l) - 1; i >= 0; i-- {
		if o, err := strconv.ParseInt(l[i], 10, 8); err == nil {
			return mysqlServerIDOffset + int(o)
		}
	}
	return mysqlServerIDOffset
}

func (cfg *Config) getOwnHostName() string {
	return fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)
}
