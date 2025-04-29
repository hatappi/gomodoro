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
	"github.com/hatappi/gomodoro/internal/core/event"
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

			return runWithAPIClient(ctx, cfg)
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

func runWithAPIClient(ctx context.Context, cfg *config.Config) error {
	logger := log.FromContext(ctx)

	clientFactory := client.NewFactory(cfg.API)
	defer clientFactory.Close()

	pomodoroClient := clientFactory.Pomodoro()
	taskClient := clientFactory.Task()

	_, err := pomodoroClient.GetCurrent(ctx)
	if err != nil {
		logger.Error(err, "Failed to connect to API server after ensuring it is running")
		return fmt.Errorf("failed to connect to API server: %w", err)
	}

	wsClient, err := clientFactory.WebSocket()
	if err != nil {
		logger.Error(err, "failed to get WebSocket client")
		return err
	}
	eventBus := event.NewClientWebSocketEventBus(wsClient)

	terminalScreen, err := screen.NewScreen(cfg)
	if err != nil {
		logger.Error(err, "Failed to create screen")
		return err
	}
	screenClient := screen.NewClient(terminalScreen)
	screenClient.StartPollEvent(ctx)

	timer := timer.NewTimer(cfg, screenClient, pomodoroClient, eventBus)
	localTaskClient := task.NewClient(cfg, screenClient)

	var opts []pomodoro.Option
	opts = append(opts, pomodoro.WithPomodoroClient(pomodoroClient))
	opts = append(opts, pomodoro.WithTaskClient(taskClient))
	opts = append(opts, pomodoro.WithWebSocketClient(wsClient))
	opts = append(opts, pomodoro.WithWorkSec(cfg.Pomodoro.WorkSec))
	opts = append(opts, pomodoro.WithShortBreakSec(cfg.Pomodoro.ShortBreakSec))
	opts = append(opts, pomodoro.WithLongBreakSec(cfg.Pomodoro.LongBreakSec))
	opts = append(opts, pomodoro.WithNotify())

	if cfg.Toggl.Enable {
		togglClient := toggl.NewClient(cfg.Toggl.ProjectID, cfg.Toggl.WorkspaceID, cfg.Toggl.APIToken)
		opts = append(opts, pomodoro.WithRecordToggl(togglClient))
	}

	if cfg.Pixela.Enable {
		pixelaClient := pixela.NewClient(cfg.Pixela.Token)
		opts = append(opts, pomodoro.WithRecordPixela(pixelaClient, cfg.Pixela.UserName, cfg.Pixela.GraphID))
	}

	p := pomodoro.NewPomodoro(cfg, screenClient, timer, localTaskClient, opts...)
	defer p.Finish()

	logger.Info("Starting Pomodoro session...")
	startErr := p.Start(ctx)
	if startErr != nil {
		logger.Error(startErr, "Pomodoro session failed to start or exited with error")
	} else {
		logger.Info("Pomodoro session finished.")
	}
	return startErr
}
