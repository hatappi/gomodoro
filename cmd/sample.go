// Package cmd has sampleCmd defined
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	"github.com/hatappi/gomodoro/errors"
	"github.com/hatappi/gomodoro/logger"
	"github.com/hatappi/gomodoro/pomodoro"
	"github.com/hatappi/gomodoro/screen"
)

// sampleCmd represents the sample command
var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "show sample",
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := cmd.Flags().GetInt("duration")
		if err != nil {
			return err
		}

		if duration > 3600 {
			return fmt.Errorf("duration max value is 3600")
		}

		logfile, err := os.OpenFile("/tmp/gomodoro.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		logger.SetOutput(logfile)
		logger.SetLogLevel(logger.DebugLevel)

		logger.Infof("sample start")
		s, err := screen.NewScreen()
		if err != nil {
			return err
		}
		defer s.Fini()

		p := pomodoro.NewPomodoro(
			s,
			pomodoro.WithWorkSec(duration),
		)
		defer p.Finish()

		err = p.Start()
		if err != nil {
			if xerrors.Is(err, errors.ErrCancel) {
				return nil
			}
			return err
		}

		return nil
	},
}

func init() {
	sampleCmd.Flags().IntP("duration", "d", 300, "duration of timer")
	rootCmd.AddCommand(sampleCmd)
}
