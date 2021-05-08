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

package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zhyass/mysql-operator/sidecar"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sidecarName  = "sidecar"
	sidecarShort = "A simple helper for mysql operator."
)

var (
	log = logf.Log.WithName("sidecar")
	cmd = &cobra.Command{
		Use:   sidecarName,
		Short: sidecarShort,
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("run the sidecar, see help section")
			os.Exit(1)
		},
	}
)

func main() {
	cfg := sidecar.NewConfig()

	initCmd := sidecar.NewInitCommand(cfg)
	cmd.AddCommand(initCmd)

	if err := cmd.Execute(); err != nil {
		log.Error(err, "failed to execute command", "cmd", cmd)
		os.Exit(1)
	}
}
