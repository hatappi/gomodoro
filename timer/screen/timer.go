// Package screen is managed screen
package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// DrawSetting for setting of drawing
type DrawSetting struct {
	BackgroundColor tcell.Color
}

// DrawOption for optoin of drawing
type DrawOption func(tcell.Style) tcell.Style

// WithBackgroundColor set BackgroundColor
func WithBackgroundColor(color tcell.Color) DrawOption {
	return func(s tcell.Style) tcell.Style {
		return s.Background(color)
	}
}

// DrawTimer is draw the timer
func (c *clientImpl) DrawTimer(x, y, mag, min, sec int, opts ...DrawOption) {
	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(c.screen, x, y, mag, string(minStr[0]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(minStr[1]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawSeparater(c.screen, x, y, mag, opts...)

	x += (separaterWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[0]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[1]), opts...)
}

func drawNumber(s tcell.Screen, x, y, mag int, nStr string, opts ...DrawOption) {
	n, _ := strconv.Atoi(nStr)
	t := strings.Split(strings.Replace(numbers[n], "\n", "", -1), "")
	draw(s, t, numberWidth, numberHeight, x, y, mag, opts...)
}

func drawSeparater(s tcell.Screen, x, y, mag int, opts ...DrawOption) {
	t := strings.Split(strings.Replace(separator, "\n", "", -1), "")
	draw(s, t, separaterWidth, separaterHeight, x, y, mag, opts...)
}

func draw(s tcell.Screen, t []string, w, h, x, y, mag int, opts ...DrawOption) {
	st := tcell.StyleDefault
	st = st.Background(tcell.ColorGreen)
	for _, opt := range opts {
		st = opt(st)
	}
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
