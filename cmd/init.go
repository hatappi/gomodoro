// Package cmd has initCmd defined
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/gomodoro/internal/config"
)

const confTpl = `pomodoro:
  work_sec: {{ .Pomodoro.WorkSec }}
  short_break_sec: {{ .Pomodoro.ShortBreakSec }}
  long_break_sec: {{ .Pomodoro.LongBreakSec }}
# toggl:
#   # https://toggl.com/app/xxxx/projects/{project_id}/team
#   project_id:
#   # Toggl API token ref: https://toggl.com/app/profile
#   api_token:
# log_file: {{ .LogFile }}
# task_file: {{ .TaskFile }}
# unix_domain_socket_path: {{ .UnixDomainScoketPath }}
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init gomodoro config file",
	Long:  "init gomodoro config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := &config.Config{
			Pomodoro: config.PomodoroConfig{
				WorkSec:       config.DefaultWorkSec,
				ShortBreakSec: config.DefaultShortBreakSec,
				LongBreakSec:  config.DefaultLongBreakSec,
			},
			LogFile:              config.DefaultLogFile,
			TaskFile:             config.DefaultTaskFile,
			UnixDomainScoketPath: config.DefaultUnixDomainScoketPath,
		}

		t := template.Must(template.New("config").Parse(confTpl))

		isStdout, err := cmd.Flags().GetBool("stdout")
		if err != nil {
			return err
		}
		if isStdout {
			err = t.Execute(os.Stdout, conf)
			if err != nil {
				return err
			}

			return nil
		}

		confPath := viper.ConfigFileUsed()

		_, err = os.Stat(confPath)
		if err == nil {
			for {
				fmt.Printf("%s exist! overwrite? (y/n) ", confPath)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()

				switch scanner.Text() {
				case "y":
					goto L
				case "n":
					return nil
				}
			}
		L:
		}

		if err = os.MkdirAll(filepath.Dir(confPath), 0750); err != nil {
			return err
		}

		confFile, err := os.OpenFile(confPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer func() {
			_ = confFile.Close()
		}()

		err = t.Execute(confFile, conf)
		if err != nil {
			return err
		}

		fmt.Printf("success to create config file. (%s)\n", confPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("stdout", false, "output config to Stdout")
}
