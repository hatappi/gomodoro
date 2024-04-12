// Package draw is managed screen
package draw

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// Timer is draw the timer.
func Timer(s tcell.Screen, x, y, mag, min, sec int, opts ...Option) {
	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(s, x, y, mag, string(minStr[0]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(s, x, y, mag, string(minStr[1]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawSeparater(s, x, y, mag, opts...)

	x += (separaterWidth + whitespaceWidth) * mag
	drawNumber(s, x, y, mag, string(secStr[0]), opts...)

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(s, x, y, mag, string(secStr[1]), opts...)
}

func drawNumber(s tcell.Screen, x, y, mag int, nStr string, opts ...Option) {
	n, _ := strconv.Atoi(nStr)
	t := strings.Split(strings.ReplaceAll(numbers[n], "\n", ""), "")
	draw(s, t, numberWidth, numberHeight, x, y, mag, opts...)
}

func drawSeparater(s tcell.Screen, x, y, mag int, opts ...Option) {
	t := strings.Split(strings.ReplaceAll(separator, "\n", ""), "")
	draw(s, t, separaterWidth, separaterHeight, x, y, mag, opts...)
}

func draw(s tcell.Screen, t []string, w, h, x, y, mag int, opts ...Option) {
	st := tcell.StyleDefault
	for _, opt := range opts {
		st = opt(st)
	}
	gl := ' '

	for row := range h {
		for col := range w {
			if t[(row*w)+(col)] == "#" {
				for pRow := range mag {
					for pCol := range mag {
						s.SetCell(x+(col*mag)+pCol, y+(row*mag)+pRow, st, gl)
					}
				}
			}
		}
	}

	s.Show()
}
