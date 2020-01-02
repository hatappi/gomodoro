// Package screen is managed screen
package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

type DrawSetting struct {
	BackgroundColor tcell.Color
}

type DrawOption func(*DrawSetting)

func WithBackgroundColor(color tcell.Color) DrawOption {
	return func(ds *DrawSetting) {
		ds.BackgroundColor = color
	}
}

// DrawTimer is draw the timer
func (c *clientImpl) DrawTimer(x, y, mag, min, sec int, opts ...DrawOption) {
	ds := &DrawSetting{
		BackgroundColor: tcell.ColorGreen,
	}
	for _, opt := range opts {
		opt(ds)
	}

	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(c.screen, x, y, mag, string(minStr[0]), ds)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(minStr[1]), ds)

	x += (numberWidth + whitespaceWidth) * mag
	drawSeparater(c.screen, x, y, mag, ds)

	x += (separaterWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[0]), ds)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[1]), ds)
}

func drawNumber(s tcell.Screen, x, y, mag int, nStr string, ds *DrawSetting) {
	n, _ := strconv.Atoi(nStr)
	t := strings.Split(strings.Replace(numbers[n], "\n", "", -1), "")
	draw(s, t, numberWidth, numberHeight, x, y, mag, ds)
}

func drawSeparater(s tcell.Screen, x, y, mag int, ds *DrawSetting) {
	t := strings.Split(strings.Replace(separator, "\n", "", -1), "")
	draw(s, t, separaterWidth, separaterHeight, x, y, mag, ds)
}

func draw(s tcell.Screen, t []string, w, h, x, y, mag int, ds *DrawSetting) {
	st := tcell.StyleDefault
	st = st.Background(ds.BackgroundColor)
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
