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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/go-ini/ini"
	"github.com/spf13/cobra"
	"github.com/zhyass/mysql-operator/utils"
)

func NewInitCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "do some initialization operations.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInitCommand(cfg); err != nil {
				log.Error(err, "init command failed")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func runInitCommand(cfg *Config) error {
	var err error

	// remove lost+found.
	if exists, _ := checkIfPathExists(dataPath); exists {
		if err := os.RemoveAll(dataPath + "/lost+found"); err != nil {
			return fmt.Errorf("removing lost+found: %s", err)
		}
	}

	// copy appropriate my.cnf from config-map to config mount.
	if err = copyFile(path.Join(configMapPath, "my.cnf"), path.Join(configPath, "my.cnf")); err != nil {
		return fmt.Errorf("failed to copy my.cnf: %s", err)
	}

	if err = os.Mkdir(extraConfPath, os.FileMode(0755)); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("error mkdir %s: %s", extraConfPath, err)
		}
	}

	// build init.sql.
	initSqlPath := path.Join(initFilePath, "init.sql")
	if err = ioutil.WriteFile(initSqlPath, buildInitSql(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write init.sql: %s", err)
	}

	// build extra.cnf.
	serverIDConfig, err := buildExtraConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to build extra.cnf: %s", err)
	}
	// save extra.cnf to conf.d.
	if err := serverIDConfig.SaveTo(path.Join(extraConfPath, "extra.cnf")); err != nil {
		return fmt.Errorf("failed to save extra.cnf: %s", err)
	}

	// copy leader-start.sh from config-map to scripts mount.
	leaderStartPath := path.Join(scriptsPath, "leader-start.sh")
	if err = copyFile(path.Join(configMapPath, "leader-start.sh"), leaderStartPath); err != nil {
		return fmt.Errorf("failed to copy scripts: %s", err)
	}
	if err = os.Chmod(leaderStartPath, os.FileMode(0755)); err != nil {
		return fmt.Errorf("failed to chmod scripts: %s", err)
	}

	// copy leader-stop.sh from config-map to scripts mount.
	leaderStopPath := path.Join(scriptsPath, "leader-stop.sh")
	if err = copyFile(path.Join(configMapPath, "leader-stop.sh"), leaderStopPath); err != nil {
		return fmt.Errorf("failed to copy scripts: %s", err)
	}
	if err = os.Chmod(leaderStopPath, os.FileMode(0755)); err != nil {
		return fmt.Errorf("failed to chmod scripts: %s", err)
	}

	// for install tokudb.
	if cfg.InitTokuDB {
		arg := fmt.Sprintf("echo never > %s/enabled", sysPath)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable the transparent_hugepage: %s", err)
		}
	}

	// build xenon.json.
	xenonFilePath := path.Join(xenonPath, "xenon.json")
	if err = ioutil.WriteFile(xenonFilePath, buildXenonConf(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write xenon.json: %s", err)
	}

	// chown
	if err = os.Chown(dataPath, 1001, 1001); err != nil {
		return fmt.Errorf("failed to chown %s: %s", dataPath, err)
	}

	log.Info("init command success")
	return nil
}

func checkIfPathExists(path string) (bool, error) {
	_, err := os.Open(path)

	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		log.Error(err, "failed to open file", "file", path)
		return false, err
	}

	return true, nil
}

func buildExtraConfig(cfg *Config) (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("mysqld")
	id := cfg.generateServerID()

	if _, err := sec.NewKey("server-id", strconv.Itoa(id)); err != nil {
		return nil, err
	}

	return conf, nil
}

func buildXenonConf(cfg *Config) []byte {
	pingTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	heartbeatTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	requestTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount

	version := "mysql80"
	if cfg.MySQLVersion.Major == 5 {
		if cfg.MySQLVersion.Minor == 6 {
			version = "mysql56"
		} else {
			version = "mysql57"
		}
	}

	var masterSysVars, slaveSysVars string
	if cfg.InitTokuDB {
		masterSysVars = "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	} else {
		masterSysVars = "sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	}

	str := fmt.Sprintf(`{
    "log": {
        "level": "INFO"
    },
    "server": {
        "endpoint": "%s:%d"
    },
    "replication": {
        "passwd": "%s",
        "user": "%s"
    },
    "rpc": {
        "request-timeout": %d
    },
    "mysql": {
        "admit-defeat-ping-count": 3,
        "admin": "root",
        "ping-timeout": %d,
        "passwd": "%s",
        "host": "localhost",
        "version": "%s",
        "master-sysvars": "%s",
        "slave-sysvars": "%s",
        "port": 3306,
        "monitor-disabled": true
    },
    "raft": {
        "election-timeout": %d,
        "admit-defeat-hearbeat-count": %d,
        "heartbeat-timeout": %d,
        "meta-datadir": "/var/lib/xenon/",
        "leader-start-command": "/scripts/leader-start.sh",
        "leader-stop-command": "/scripts/leader-stop.sh",
        "semi-sync-degrade": true,
        "purge-binlog-disabled": true,
        "super-idle": false
    }
}
`, cfg.getOwnHostName(), utils.XenonPort, cfg.ReplicationPassword, cfg.ReplicationUser, requestTimeout,
		pingTimeout, cfg.RootPassword, version, masterSysVars, slaveSysVars, cfg.ElectionTimeout,
		cfg.AdmitDefeatHearbeatCount, heartbeatTimeout)
	return utils.StringToBytes(str)
}

func buildInitSql(cfg *Config) []byte {
	sql := fmt.Sprintf(`RESET MASTER;
SET @@SESSION.SQL_LOG_BIN=0;
DELETE FROM mysql.user WHERE user='%s';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* to '%s'@'%%' IDENTIFIED BY '%s';
`, cfg.ReplicationUser, cfg.ReplicationUser, cfg.ReplicationPassword)

	if len(cfg.MetricsUser) > 0 {
		sql += fmt.Sprintf(`DELETE FROM mysql.user WHERE user='%s';
GRANT SELECT, PROCESS, REPLICATION CLIENT ON *.* to '%s'@'localhost' IDENTIFIED BY '%s';
`, cfg.MetricsUser, cfg.MetricsUser, cfg.MetricsPassword)
	}

	sql += `FLUSH PRIVILEGES;`

	return utils.StringToBytes(sql)
}
