// Package cmd has sampleCmd defined
package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

		tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
		s, err := tcell.NewScreen()
		if err != nil {
			return err
		}

		if err = s.Init(); err != nil {
			return err
		}
		defer s.Fini()

		s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite))
		s.Clear()

		quit := make(chan struct{})
		go func() {
			for {
				ev := s.PollEvent()
				switch ev := ev.(type) {
				case *tcell.EventKey:
					switch ev.Key() {
					case tcell.KeyEscape, tcell.KeyEnter:
						close(quit)
						return
					}
				case *tcell.EventResize:
					s.Sync()
				}
			}
		}()

		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		for {
			w, h := s.Size()

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

			x = math.Round(x + ((cw - (TIMER_WIDTH * mag)) / 2))
			y = math.Round(y + ((ch - (TIMER_HEIGHT * mag)) / 2))

			s.Clear()
			printSentence(s, int(x), int(y), int(TIMER_WIDTH*mag), "今年は令和2年です")
			DrawTimer(s, int(x), int(y)+2, int(mag), min, sec)

			select {
			case <-quit:
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
	x := math.Round(w / TIMER_WIDTH)
	y := math.Round(h / TIMER_HEIGHT)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, errors.New("screen is small")
		}

		if w >= TIMER_WIDTH*mag && h >= TIMER_HEIGHT*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}

var separator = `
-
#
-
#
-
`

var numbers = []string{
	`
####
#--#
#--#
#--#
####
	`,
	`
---#
---#
---#
---#
---#
`,
	`
####
---#
####
#---
####
`,
	`
####
---#
####
---#
####
`,
	`
#-#-
#-#-
####
--#-
--#-
`,
	`
####
#---
####
---#
####
`,
	`
#---
#---
####
#--#
####
`,
	`
####
---#
---#
---#
---#
`,
	`
####
#--#
####
#--#
####
`,
	`
####
#--#
####
---#
####
`,
}

const (
	TIMER_WIDTH  = 21
	TIMER_HEIGHT = 5

	number_width  = 4
	number_height = 5

	separater_width  = 1
	separater_height = 5

	whitespace_width = 1
)

func DrawTimer(s tcell.Screen, x, y, mag, min, sec int) {
	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(s, x, y, mag, string(minStr[0]))

	x += (number_width + whitespace_width) * mag
	drawNumber(s, x, y, mag, string(minStr[1]))

	x += (number_width + whitespace_width) * mag
	drawSeparater(s, x, y, mag)

	x += (separater_width + whitespace_width) * mag
	drawNumber(s, x, y, mag, string(secStr[0]))

	x += (number_width + whitespace_width) * mag
	drawNumber(s, x, y, mag, string(secStr[1]))
}

func drawNumber(s tcell.Screen, x, y, mag int, nStr string) {
	n, _ := strconv.Atoi(nStr)
	t := strings.Split(strings.Replace(numbers[n], "\n", "", -1), "")
	draw(s, t, number_width, number_height, x, y, mag)
}

func drawSeparater(s tcell.Screen, x, y, mag int) {
	t := strings.Split(strings.Replace(separator, "\n", "", -1), "")
	draw(s, t, separater_width, separater_height, x, y, mag)
}

func draw(s tcell.Screen, t []string, w, h, x, y, mag int) {
	st := tcell.StyleDefault
	st = st.Background(tcell.ColorGreen)
	gl := ' '

	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			if t[(row*w)+(col)] == "#" {
				for pRow := 0; pRow < mag; pRow++ {
					for pCol := 0; pCol < mag; pCol++ {
						s.SetCell(x+(col*mag)+pCol, y+(row*mag)+pRow, st, gl)
					}
				}
			}
		}
	}

	s.Show()
}

func printSentence(s tcell.Screen, x, y, maxWidth int, str string) {
	remain := (maxWidth - runewidth.StringWidth(str)) / 2
	if remain >= 0 {
		for remain > 0 {
			str = " " + str
			remain--
		}
	} else {
		str = str[:maxWidth-3]
		str += "..."
	}

	style := tcell.StyleDefault
	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
}
