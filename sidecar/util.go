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
	"io"
	"os"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/zhyass/mysql-operator/utils"
)

var (
	log                 = logf.Log.WithName("sidecar")
	mysqlServerIDOffset = 100
	configPath          = utils.ConfVolumeMountPath
	configMapPath       = utils.ConfMapVolumeMountPath
	dataPath            = utils.DataVolumeMountPath
	extraConfPath       = utils.ConfVolumeMountPath + "/conf.d"
	scriptsPath         = utils.ScriptsVolumeMountPath
	sysPath             = utils.SysVolumeMountPath
	xenonPath           = utils.XenonVolumeMountPath
	initFilePath        = utils.InitFileVolumeMountPath
)

// copyFile the src file to dst.
// nolint: gosec
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := in.Close(); err1 != nil {
			log.Error(err1, "failed to close source file", "src_file", src)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := out.Close(); err1 != nil {
			log.Error(err1, "failed to close destination file", "dest_file", dst)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func getEnvValue(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Info("environment is not set", "key", key)
	}

	return value
}

// Generate mysql server-id from pod ordinal index.
func generateServerID(name string) (int, error) {
	idx := strings.LastIndexAny(name, "-")
	if idx == -1 {
		return -1, fmt.Errorf("failed to extract ordinal from hostname: %s", name)
	}

	ordinal, err := strconv.Atoi(name[idx+1:])
	if err != nil {
		log.Error(err, "failed to extract ordinal form hostname", "hostname", name)
		return -1, fmt.Errorf("failed to extract ordinal from hostname: %s", name)
	}
	return mysqlServerIDOffset + ordinal, nil
}
