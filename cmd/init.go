// Package cmd has initCmd defined
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/gomodoro/internal/config"
)

const confTpl = `# pomodoro:
#   work_sec: {{ .Pomodoro.WorkSec }}
#   short_break_sec: {{ .Pomodoro.ShortBreakSec }}
#   long_break_sec: {{ .Pomodoro.LongBreakSec }}
# toggl:
#   enable: false
#   # https://track.toggl.com/{organization_id}/projects/{project_id}/team
#   project_id:
#   # https://track.toggl.com/organization/{organization_id}/workspaces/{workspace_id}
#   workspace_id:
#   # Toggl API token ref: https://toggl.com/app/profile
#   api_token:
# log_file: {{ .LogFile }}

## You can change the colors used within gomodoro.
## You need to specify W3C Color name (e.g. red) or HEX (.e.g. #ffffff)
# color:
#   font: "#2F4F4F"
#   background: "white"
#   selected_line: "blue"
#   status_bar_background: "black"
#   timer_pause_font: "#FF8C00"
#   timer_work_font: "green"
#   timer_break_font: "blue"
#   cursor: "green"
`

func newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "init gomodoro config file",
		Long:  "init gomodoro config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf := config.DefaultConfig()

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
				return fmt.Errorf("%s already exist", confPath)
			}

			//nolint:mnd
			if err = os.MkdirAll(filepath.Dir(confPath), 0o750); err != nil {
				return err
			}

			confFile, err := os.OpenFile(filepath.Clean(confPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:mnd
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

	initCmd.Flags().Bool("stdout", false, "output config to Stdout")

	return initCmd
}
