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
	"os"

	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

func NewTailCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail <file>",
		Short: "tail -f file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTailCommand(cfg, args); err != nil {
				log.Error(err, "tail command failed")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func runTailCommand(cfg *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("args count error: should be 1")
	}

	file := args[0]
	if exists, _ := checkIfPathExists(file); !exists {
		return fmt.Errorf("cannot find the file: %s", file)
	}

	log.Info("prepare to tail the file", "file", file)
	t, err := tail.TailFile(file, tail.Config{Follow: true})
	if err != nil {
		return fmt.Errorf("failed to tail %s: %s", file, err)
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	return nil
}
