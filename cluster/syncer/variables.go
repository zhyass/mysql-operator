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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var log = logf.Log.WithName("cluster-syncer")

var mysqlCommonConfigs = map[string]string{
	"character_set_server":                            "utf8mb4",
	"interactive_timeout":                             "3600",
	"default-time-zone":                               "+08:00",
	"expire_logs_days":                                "7",
	"key_buffer_size":                                 "33554432",
	"log_bin_trust_function_creators":                 "1",
	"long_query_time":                                 "3",
	"binlog_cache_size":                               "32768",
	"binlog_stmt_cache_size":                          "32768",
	"max_connections":                                 "1024",
	"max_connect_errors":                              "655360",
	"query_cache_size":                                "0",
	"sync_master_info":                                "1000",
	"sync_relay_log":                                  "1000",
	"sync_relay_log_info":                             "1000",
	"table_open_cache":                                "2000",
	"thread_cache_size":                               "128",
	"wait_timeout":                                    "3600",
	"group_concat_max_len":                            "1024",
	"slave_rows_search_algorithms":                    "INDEX_SCAN,HASH_SCAN",
	"max_allowed_packet":                              "1073741824",
	"event_scheduler":                                 "OFF",
	"innodb_print_all_deadlocks":                      "0",
	"autocommit":                                      "1",
	"transaction-isolation":                           "READ-COMMITTED",
	"audit_log_policy":                                "NONE",
	"audit_log_rotate_on_size":                        "104857600",
	"audit_log_rotations":                             "6",
	"validate_password_policy":                        "MEDIUM",
	"validate_password_length":                        "8",
	"connection_control_failed_connections_threshold": "3",
	"connection_control_min_connection_delay":         "1000",
	"connection_control_max_connection_delay":         "2147483647",
	"explicit_defaults_for_timestamp":                 "0",
	"innodb_adaptive_hash_index":                      "0",
}

var mysqlStaticConfigs = map[string]string{
	"audit_log_format":            "OLD",
	"default-storage-engine":      "InnoDB",
	"back_log":                    "2048",
	"ft_min_word_len":             "4",
	"lower_case_table_names":      "0",
	"query_cache_type":            "OFF",
	"innodb_ft_max_token_size":    "84",
	"innodb_ft_min_token_size":    "3",
	"sql_mode":                    "STRICT_TRANS_TABLES,NO_ENGINE_SUBSTITUTION",
	"slave_parallel_workers":      "8",
	"slave_pending_jobs_size_max": "1073741824",
	"innodb_log_buffer_size":      "16777216",
	"innodb_log_file_size":        "1073741824",
	"innodb_log_files_in_group":   "2",
	"innodb_flush_method":         "O_DIRECT",
	"innodb_use_native_aio":       "1",
	"innodb_autoinc_lock_mode":    "2",
	"performance_schema":          "1",
}

// containers names
const (
	// init containers
	containerInitMysqlName = "init-mysql"

	// containers
	containerMysqlName   = "mysql"
	containerXenonName   = "xenon"
	containerMetricsName = "metrics"
	containeSlowLogName  = "slowlog"
)

// volumes names
const (
	confVolumeName    = "conf"
	confMapVolumeName = "config-map"
	scriptsVolumeName = "scripts"
	logsVolumeName    = "logs"
	dataVolumeName    = "data"
	sysVolumeName     = "host-sys"
)
