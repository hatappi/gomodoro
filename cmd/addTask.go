/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/editor"
)

const initialText = `# Please write one task per line`

// addTaskCmd represents the addTask command
var addTaskCmd = &cobra.Command{
	Use:   "add-task",
	Short: "add task",
	RunE: func(cmd *cobra.Command, args []string) error {
		ts, err := editor.GetSliceText(initialText)
		if err != nil {
			return err
		}

		taskNames := make([]string, 0)
		for _, t := range ts {
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, "#") {
				continue
			}
			taskNames = append(taskNames, t)
		}

		fmt.Printf("add %s\n", strings.Join(taskNames, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addTaskCmd)
}
