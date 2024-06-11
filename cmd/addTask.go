package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/editor"
	"github.com/hatappi/gomodoro/internal/task"
)

const initialText = `# Please write one task per line`

func newAddTaskCmd() *cobra.Command {
	addTaskCmd := &cobra.Command{
		Use:   "add-task TASK_NAME",
		Short: "add task",
		Long: `This command adds a task.
Please specify the task name in the argument.
if you doesn't specify task name, editor starts up.
And add a task using the editor.
`,
		RunE: func(_ *cobra.Command, args []string) error {
			config, err := config.GetConfig()
			if err != nil {
				return err
			}

			var newTasks []string

			if len(args) > 0 {
				newTasks = append(newTasks, strings.Join(args, " "))
			}

			if len(newTasks) == 0 {
				lines, err := editor.ContentsByLine(initialText)
				if err != nil {
					return err
				}

				for _, l := range lines {
					// ignore empty string and comment
					if l == "" || strings.HasPrefix(l, "#") {
						continue
					}

					newTasks = append(newTasks, l)
				}
			}

			for _, newTask := range newTasks {
				if err := task.AddTask(config.TaskFile, newTask); err != nil {
					return err
				}

				fmt.Printf("added '%s'\n", newTask)
			}

			return nil
		},
	}

	return addTaskCmd
}
