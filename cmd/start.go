// Package cmd has startCmd defined
package cmd

import (
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/net/unix"
	"github.com/hatappi/gomodoro/internal/pomodoro"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/toggl"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start pomodoro",
	Long: `start pomodoro.
if you want to change work time, break time,
please specify argument or config yaml.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()
		if err != nil {
			return err
		}
		// pomodoro
		s, err := screen.NewScreen()
		if err != nil {
			return err
		}
		defer s.Fini()

		pc := config.Pomodoro
		opts := []pomodoro.Option{
			pomodoro.WithWorkSec(pc.WorkSec),
			pomodoro.WithShortBreakSec(pc.ShortBreakSec),
			pomodoro.WithLongBreakSec(pc.LongBreakSec),
			pomodoro.WithNotify(),
		}

		if config.Toggl.Enable() {
			togglClient := toggl.NewClient(config.Toggl.ProjectID, config.Toggl.APIToken)
			opts = append(opts, pomodoro.WithRecordToggl(togglClient))
		}

		tf, err := config.ExpandTaskFile()
		if err != nil {
			return err
		}

		p := pomodoro.NewPomodoro(s, tf, opts...)
		defer p.Finish()

		// unix domain socket server
		udsp, err := config.ExpandUnixDomainSocketPath()
		if err != nil {
			return err
		}

		server, err := unix.NewServer(udsp, p.GetTimer())
		if err != nil {
			return err
		}
		defer server.Close()
		go server.Serve()

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
	startCmd.Flags().IntP("work-sec", "w", config.DefaultWorkSec, "work seconds")
	_ = viper.BindPFlag("pomodoro.work_sec", startCmd.Flags().Lookup("work-sec"))

	startCmd.Flags().IntP("short-break-sec", "s", config.DefaultShortBreakSec, "short break seconds")
	_ = viper.BindPFlag("pomodoro.short_break_sec", startCmd.Flags().Lookup("short-break-sec"))

	startCmd.Flags().IntP("long-break-sec", "l", config.DefaultLongBreakSec, "long break seconds")
	_ = viper.BindPFlag("pomodoro.long_break_sec", startCmd.Flags().Lookup("long-break-sec"))

	home, _ := homedir.Expand(config.DefaultTaskFile)
	startCmd.Flags().StringP("task-file", "t", home, "task file path")
	_ = viper.BindPFlag("task_file", startCmd.Flags().Lookup("task-file"))

	rootCmd.AddCommand(startCmd)
}
