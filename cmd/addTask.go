package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/config"
	"github.com/hatappi/gomodoro/editor"
	"github.com/hatappi/gomodoro/task"
)

const initialText = `# Please write one task per line`

// addTaskCmd represents the addTask command
var addTaskCmd = &cobra.Command{
	Use:   "add-task TASK_NAME",
	Short: "add task",
	Long: `This command adds a task.
Please specify the task name in the argument.
if you doesn't specify task name, editor starts up.
And add a task using the editor.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()
		if err != nil {
			return err
		}
		taskPath, err := config.ExpandTaskFile()
		if err != nil {
			return err
		}

		name := strings.Join(args, " ")
		if name != "" {
			err = task.AddTask(taskPath, name)
			if err != nil {
				return err
			}
			fmt.Printf("add %s\n", name)
			return nil
		}

		ts, err := editor.GetSliceText(initialText)
		if err != nil {
			return err
		}

		for _, t := range ts {
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, "#") {
				continue
			}
			err = task.AddTask(taskPath, t)
			if err != nil {
				return err
			}
			fmt.Printf("add %s\n", t)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addTaskCmd)
}
