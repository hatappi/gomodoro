// Package cmd has startCmd defined
package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// secondsPerMinute represents the number of seconds in a minute.
const secondsPerMinute = 60

func newRemainCmd() *cobra.Command {
	remainCmd := &cobra.Command{
		Use:   "remain",
		Short: "get remain time",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ignoreError, err := cmd.Flags().GetBool("ignore-error")
			if err != nil {
				return err
			}

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			factory := client.NewFactory(cfg.API)
			defer func() {
				// Since remain.go command is often used in status bars,
				// we don't want to print errors to stdout, so we just ignore the error here
				_ = factory.Close()
			}()

			pomodoroClient := factory.Pomodoro()

			ctx := cmd.Context()
			pomodoro, err := pomodoroClient.GetCurrent(ctx)
			if err != nil {
				if !ignoreError {
					return err
				}

				fmt.Printf("--:--")
				return nil
			}

			if pomodoro == nil {
				fmt.Printf("--:--")
				return nil
			}

			var remainingStr string
			if slices.Contains([]event.PomodoroState{event.PomodoroStateActive, event.PomodoroStatePaused}, pomodoro.State) {
				minutes := pomodoro.RemainingTime / secondsPerMinute
				seconds := pomodoro.RemainingTime % secondsPerMinute

				remainingStr = fmt.Sprintf("%02d:%02d", minutes, seconds)
			} else {
				remainingStr = "--:--"
			}

			fmt.Print(remainingStr)
			return nil
		},
	}

	remainCmd.Flags().BoolP("ignore-error", "i", false, "ignore error")

	return remainCmd
}
