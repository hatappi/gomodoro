// Package cmd has sampleCmd defined
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"

	"github.com/hatappi/gomodoro/config"
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
		logfile, err := os.OpenFile("/tmp/gomodoro.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		logger.SetOutput(logfile)
		logger.SetLogLevel(logger.DebugLevel)

		logger.Infof("sample start")

		config, err := config.GetConfig()
		if err != nil {
			return err
		}
		logger.Infof("config is %+v", config)

		s, err := screen.NewScreen()
		if err != nil {
			return err
		}
		defer s.Fini()

		pc := config.Pomodoro
		p := pomodoro.NewPomodoro(
			s,
			pomodoro.WithWorkSec(pc.WorkSec),
			pomodoro.WithShortBreakSec(pc.ShortBreakSec),
			pomodoro.WithLongBreakSec(pc.LongBreakSec),
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
	sampleCmd.Flags().IntP("work-sec", "w", 1500, "work seconds")
	_ = viper.BindPFlag("pomodoro.work_sec", sampleCmd.Flags().Lookup("work-sec"))

	sampleCmd.Flags().IntP("short-break-sec", "s", 300, "short break seconds")
	_ = viper.BindPFlag("pomodoro.short_break_sec", sampleCmd.Flags().Lookup("short-break-sec"))

	sampleCmd.Flags().IntP("long-break-sec", "l", 900, "long break seconds")
	_ = viper.BindPFlag("pomodoro.long_break_sec", sampleCmd.Flags().Lookup("long-break-sec"))

	rootCmd.AddCommand(sampleCmd)
}
