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

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errorConnectionStates = []string{
		"connecting to master",
		"reconnecting after a failed binlog dump request",
		"reconnecting after a failed master event read",
		"waiting to reconnect after a failed binlog dump request",
		"waiting to reconnect after a failed master event read",
	}
)

type SQLRunner struct {
	db *sql.DB
}

func NewSQLRunner(client client.Client, namespace, secretName, user, host string, port int) (*SQLRunner, error) {
	secret := &corev1.Secret{}
	if err := client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      secretName,
		},
		secret,
	); err != nil {
		return nil, err
	}

	password, ok := secret.Data[user]
	if !ok {
		return nil, fmt.Errorf("failed to get the user: %s", user)
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&interpolateParams=true",
		user, password, host, port,
	)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &SQLRunner{db}, nil
}

func (s *SQLRunner) CheckSlaveStatusWithRetry(retry uint32) (isLagged, isReplicating bool, err error) {
	for {
		if retry == 0 {
			break
		}

		if isLagged, isReplicating, err = s.checkSlaveStatus(); err == nil {
			return
		}

		time.Sleep(time.Second * 3)
		retry--
	}

	return
}

func (s *SQLRunner) checkSlaveStatus() (isLagged, isReplicating bool, err error) {
	var rows *sql.Rows
	rows, err = s.db.Query("show slave status;")
	if err != nil {
		return
	}

	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return
		}
		return
	}

	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		return
	}

	scanArgs := make([]interface{}, len(cols))
	for i := range scanArgs {
		scanArgs[i] = &sql.RawBytes{}
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return
	}

	slaveIOState := strings.ToLower(columnValue(scanArgs, cols, "Slave_IO_State"))
	slaveSQLRunning := columnValue(scanArgs, cols, "Slave_SQL_Running")
	lastSQLError := columnValue(scanArgs, cols, "Last_SQL_Error")
	secondsBehindMaster := columnValue(scanArgs, cols, "Seconds_Behind_Master")

	if stringInArray(slaveIOState, errorConnectionStates) {
		err = fmt.Errorf("Slave_IO_State: %s", slaveIOState)
		return
	}

	if slaveSQLRunning != "Yes" {
		err = fmt.Errorf("SQL thread error: %s", lastSQLError)
		return
	}

	isReplicating = true

	var longQueryTime float64
	if err = s.GetGlobalVariable("long_query_time", &longQueryTime); err != nil {
		return
	}

	sec, _ := strconv.ParseFloat(secondsBehindMaster, 64)
	if sec > longQueryTime*100 {
		isLagged = true
	}

	return
}

func (s *SQLRunner) CheckReadOnly() (bool, error) {
	var readOnly uint8
	if err := s.GetGlobalVariable("read_only", &readOnly); err != nil {
		return false, err
	}

	if readOnly == 0 {
		return false, nil
	}

	return true, nil
}

func (sr *SQLRunner) GetGlobalVariable(param string, val interface{}) error {
	query := fmt.Sprintf("select @@global.%s", param)
	return sr.db.QueryRow(query).Scan(val)
}

func (sr *SQLRunner) Close() error {
	return sr.db.Close()
}

func columnValue(scanArgs []interface{}, slaveCols []string, colName string) string {
	columnIndex := -1
	for idx := range slaveCols {
		if slaveCols[idx] == colName {
			columnIndex = idx
			break
		}
	}

	if columnIndex == -1 {
		return ""
	}

	return string(*scanArgs[columnIndex].(*sql.RawBytes))
}

func stringInArray(str string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, str)
	return index < len(strArray) && strArray[index] == str
}
