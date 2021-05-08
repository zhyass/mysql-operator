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

package utils

const (
	MysqlPortName = "mysql"
	MysqlPort     = 3306

	MetricsPortName = "metrics"
	MetricsPort     = 9104

	XenonPortName = "xenon"
	XenonPort     = 8801

	ReplicationUser = "qc_repl"
)

var (
	// MySQLDefaultVersion is the version for mysql that should be used
	MySQLDefaultVersion = "5.7.33"

	// MySQLTagsToSemVer maps simple version to semver versions
	MySQLTagsToSemVer = map[string]string{
		"5.7": "5.7.33",
	}

	// MysqlImageVersions is a map of supported mysql version and their image
	MysqlImageVersions = map[string]string{
		"5.7.33": "xenondb/percona:5.7.33",
	}
)

// containers names
const (
	// init containers
	ContainerInitMysqlName = "init-mysql"

	// containers
	ContainerMysqlName   = "mysql"
	ContainerXenonName   = "xenon"
	ContainerMetricsName = "metrics"
	ContainerSlowLogName = "slowlog"
)

// volumes names
const (
	ConfVolumeName    = "conf"
	ConfMapVolumeName = "config-map"
	LogsVolumeName    = "logs"
	DataVolumeName    = "data"
	SysVolumeName     = "host-sys"
	ScriptsVolumeName = "scripts"
	XenonVolumeName   = "xenon"
)

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

const (
	// HeadlessSVC is the alias of the headless service resource
	HeadlessSVC ResourceName = "headless"
	// StatefulSet is the alias of the statefulset resource
	StatefulSet ResourceName = "mysql"
	// ConfigMap is the alias for mysql configs, the config map resource
	ConfigMap ResourceName = "config-files"
	// LeaderService is the name of the service that points to leader node
	LeaderService ResourceName = "leader-service"
	// FollowerService is the name of a service that points healthy followers (excludes leader)
	FollowerService ResourceName = "follower-service"
	// Secret is the name of the "private" secret that contains operator related credentials
	Secret ResourceName = "secret"

	Role ResourceName = "role"

	RoleBinding ResourceName = "rolebinding"
)
