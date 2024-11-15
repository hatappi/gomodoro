// Package cmd has startCmd defined
package cmd

import (
	"errors"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/gomodoro/internal/config"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/net/unix"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/pomodoro"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/task"
	"github.com/hatappi/gomodoro/internal/timer"
	"github.com/hatappi/gomodoro/internal/toggl"
)

// startCmd represents the start command.
func newStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "start pomodoro",
		Long: `start pomodoro.
if you want to change work time, break time,
please specify argument or config yaml.
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			config, err := config.GetConfig()
			if err != nil {
				return err
			}

			opts := []pomodoro.Option{
				pomodoro.WithWorkSec(config.Pomodoro.WorkSec),
				pomodoro.WithShortBreakSec(config.Pomodoro.ShortBreakSec),
				pomodoro.WithLongBreakSec(config.Pomodoro.LongBreakSec),
				pomodoro.WithNotify(),
			}

			if config.Toggl.Enable {
				togglClient := toggl.NewClient(config.Toggl.ProjectID, config.Toggl.WorkspaceID, config.Toggl.APIToken)
				opts = append(opts, pomodoro.WithRecordToggl(togglClient))
			}

			if config.Pixela.Enable {
				client := pixela.NewClient(config.Pixela.Token)
				opts = append(opts, pomodoro.WithRecordPixela(client, config.Pixela.UserName, config.Pixela.GraphID))
			}

			terminalScreen, err := screen.NewScreen(config)
			if err != nil {
				return err
			}

			screenClient := screen.NewClient(terminalScreen)
			screenClient.StartPollEvent(ctx)

			timer := timer.NewTimer(config, screenClient)
			taskClient := task.NewClient(config, screenClient, config.TaskFile)

			p := pomodoro.NewPomodoro(config, screenClient, timer, taskClient, opts...)
			defer p.Finish()

			server, err := unix.NewServer(config.UnixDomainScoketPath, timer)
			if err != nil {
				return err
			}
			defer server.Close()
			go server.Serve(ctx)

			err = p.Start(ctx)
			if err == nil || errors.Is(err, gomodoro_error.ErrCancel) {
				return nil
			}

			return err
		},
	}

	startCmd.Flags().IntP("work-sec", "w", config.DefaultWorkSec, "work seconds")
	_ = viper.BindPFlag("pomodoro.work_sec", startCmd.Flags().Lookup("work-sec"))

	startCmd.Flags().IntP("short-break-sec", "s", config.DefaultShortBreakSec, "short break seconds")
	_ = viper.BindPFlag("pomodoro.short_break_sec", startCmd.Flags().Lookup("short-break-sec"))

	startCmd.Flags().IntP("long-break-sec", "l", config.DefaultLongBreakSec, "long break seconds")
	_ = viper.BindPFlag("pomodoro.long_break_sec", startCmd.Flags().Lookup("long-break-sec"))

	home, _ := homedir.Expand(config.DefaultTaskFile)
	startCmd.Flags().StringP("task-file", "t", home, "task file path")
	_ = viper.BindPFlag("task_file", startCmd.Flags().Lookup("task-file"))

	return startCmd
}
