// Package cmd has startCmd defined
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/api/server"
	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/toggl"
	"github.com/hatappi/gomodoro/internal/tui"
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
			logger := log.FromContext(ctx)

			cfg, err := config.GetConfig()
			if err != nil {
				logger.Error(err, "Failed to get config")
				return err
			}

			serverRunner := server.NewRunner(cfg)

			err = serverRunner.EnsureRunning(ctx)
			if err != nil {
				logger.Error(err, "Failed to ensure API server is running")
				return fmt.Errorf("failed to ensure API server is running: %w", err)
			}

			defer func() {
				if err := serverRunner.Stop(ctx); err != nil {
					logger.Error(err, "Failed to stop API server")
				}
			}()

			return runTUIApp(ctx, cfg)
		},
	}

	startCmd.Flags().IntP("work-sec", "w", config.DefaultWorkSec, "work seconds")
	_ = viper.BindPFlag("pomodoro.work_sec", startCmd.Flags().Lookup("work-sec"))

	startCmd.Flags().IntP("short-break-sec", "s", config.DefaultShortBreakSec, "short break seconds")
	_ = viper.BindPFlag("pomodoro.short_break_sec", startCmd.Flags().Lookup("short-break-sec"))

	startCmd.Flags().IntP("long-break-sec", "l", config.DefaultLongBreakSec, "long break seconds")
	_ = viper.BindPFlag("pomodoro.long_break_sec", startCmd.Flags().Lookup("long-break-sec"))

	return startCmd
}

func runTUIApp(ctx context.Context, cfg *config.Config) error {
	logger := log.FromContext(ctx)

	// Create App options based on configuration
	var opts []tui.Option
	opts = append(opts, tui.WithWorkSec(cfg.Pomodoro.WorkSec))
	opts = append(opts, tui.WithShortBreakSec(cfg.Pomodoro.ShortBreakSec))
	opts = append(opts, tui.WithLongBreakSec(cfg.Pomodoro.LongBreakSec))
	opts = append(opts, tui.WithNotify())

	if cfg.Toggl.Enable {
		togglClient := toggl.NewClient(cfg.Toggl.ProjectID, cfg.Toggl.WorkspaceID, cfg.Toggl.APIToken)
		opts = append(opts, tui.WithRecordToggl(togglClient))
	}

	if cfg.Pixela.Enable {
		pixelaClient := pixela.NewClient(cfg.Pixela.Token)
		opts = append(opts, tui.WithRecordPixela(pixelaClient, cfg.Pixela.UserName, cfg.Pixela.GraphID))
	}

	clientFactory := client.NewFactory(cfg.API)
	defer clientFactory.Close()

	app, err := tui.NewApp(ctx, cfg, clientFactory, opts...)
	if err != nil {
		logger.Error(err, "Failed to create TUI App")
		return err
	}
	defer app.Finish()

	logger.Info("Starting Pomodoro session...")
	startErr := app.Run(ctx)
	if startErr != nil {
		if startErr == gomodoro_error.ErrCancel {
			logger.Info("Pomodoro session canceled by user.")
			return nil
		}

		return startErr
	}

	return nil
}
