// Package cmd has sampleCmd defined
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/timer"
	"github.com/hatappi/gomodoro/timer/screen"
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

		c, err := screen.NewClient()
		if err != nil {
			return err
		}
		defer c.Finish()

		c.Start()

		t := timer.NewTimer(c)
		err = t.Run(duration)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	sampleCmd.Flags().IntP("duration", "d", 300, "duration of timer")
	rootCmd.AddCommand(sampleCmd)
}
