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
	"io"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	log                 = logf.Log.WithName("sidecar")
	mysqlServerIDOffset = 100
	configPath          = "/mnt/conf.d"
	configMapPath       = "/mnt/config-map"
	scriptsPath         = "/mnt/scripts"
	xenonPath           = "/mnt/xenon"
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
