// Package draw is managed screen
package draw

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// Timer is draw the timer
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
	t := strings.Split(strings.Replace(numbers[n], "\n", "", -1), "")
	draw(s, t, numberWidth, numberHeight, x, y, mag, opts...)
}

func drawSeparater(s tcell.Screen, x, y, mag int, opts ...Option) {
	t := strings.Split(strings.Replace(separator, "\n", "", -1), "")
	draw(s, t, separaterWidth, separaterHeight, x, y, mag, opts...)
}

func draw(s tcell.Screen, t []string, w, h, x, y, mag int, opts ...Option) {
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
