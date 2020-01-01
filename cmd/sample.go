// Package cmd has sampleCmd defined
package cmd

import (
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

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

		c, err := screen.NewClient()
		if err != nil {
			return err
		}
		defer c.Finish()

		c.Start()

		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		for {
			w, h := c.ScreenSize()

			min := duration / 60
			sec := duration % 60

			x := float64(w) / 16
			y := float64(h) / 16

			printLine := 2.0
			cw := float64(w) * 14 / 16
			ch := float64(h) * 14 / 16
			ch -= printLine

			mag, err := getMagnification(cw, ch)
			if err != nil {
				return err
			}

			x = math.Round(x + ((cw - (screen.TIMER_WIDTH * mag)) / 2))
			y = math.Round(y + ((ch - (screen.TIMER_HEIGHT * mag)) / 2))

			c.Clear()
			c.DrawSentence(int(x), int(y), int(screen.TIMER_WIDTH*mag), "今年は令和2年です")
			c.DrawTimer(int(x), int(y)+2, int(mag), min, sec)

			select {
			case <-c.Quit:
				return nil
			case <-t.C:
			}

			duration -= 1

			if duration == 0 {
				t.Stop()
			}
		}
	},
}

func init() {
	sampleCmd.Flags().IntP("duration", "d", 300, "duration of timer")
	rootCmd.AddCommand(sampleCmd)
}

func getMagnification(w, h float64) (float64, error) {
	x := math.Round(w / screen.TIMER_WIDTH)
	y := math.Round(h / screen.TIMER_HEIGHT)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, errors.New("screen is small")
		}

		if w >= screen.TIMER_WIDTH*mag && h >= screen.TIMER_HEIGHT*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}
