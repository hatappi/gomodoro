// Package screen is managed screen
package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// DrawTimer is draw the timer
func (c *Client) DrawTimer(x, y, mag, min, sec int) {
	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(c.screen, x, y, mag, string(minStr[0]))

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(minStr[1]))

	x += (numberWidth + whitespaceWidth) * mag
	drawSeparater(c.screen, x, y, mag)

	x += (separaterWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[0]))

	x += (numberWidth + whitespaceWidth) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[1]))
}

func drawNumber(s tcell.Screen, x, y, mag int, nStr string) {
	n, _ := strconv.Atoi(nStr)
	t := strings.Split(strings.Replace(numbers[n], "\n", "", -1), "")
	draw(s, t, numberWidth, numberHeight, x, y, mag)
}

func drawSeparater(s tcell.Screen, x, y, mag int) {
	t := strings.Split(strings.Replace(separator, "\n", "", -1), "")
	draw(s, t, separaterWidth, separaterHeight, x, y, mag)
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
