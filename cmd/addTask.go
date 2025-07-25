package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/api/server"
	"github.com/hatappi/gomodoro/internal/client/graphql"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/editor"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			serverRunner := server.NewRunner(cfg)

			if err := serverRunner.EnsureRunning(ctx); err != nil {
				log.FromContext(ctx).Error(err, "Failed to ensure API server is running")
				return fmt.Errorf("failed to ensure API server is running: %w", err)
			}

			defer func() {
				if err := serverRunner.Stop(ctx); err != nil {
					log.FromContext(ctx).Error(err, "Failed to stop API server")
				}
			}()

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
					if l == "" || strings.HasPrefix(l, "#") {
						continue
					}

					newTasks = append(newTasks, l)
				}
			}

			gqlClient := graphql.NewClientWrapper(cfg.API)

			for _, newTaskTitle := range newTasks {
				task, err := gqlClient.CreateTask(ctx, newTaskTitle)
				if err != nil {
					return fmt.Errorf("failed to create task '%s': %w", newTaskTitle, err)
				}

				fmt.Printf("added task '%s' with ID '%s'\n", task.Title, task.ID)
			}

			return nil
		},
	}

	return addTaskCmd
}
