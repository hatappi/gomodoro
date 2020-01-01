package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

func (c *Client) DrawTimer(x, y, mag, min, sec int) {
	minStr := fmt.Sprintf("%02d", min)
	secStr := fmt.Sprintf("%02d", sec)

	drawNumber(c.screen, x, y, mag, string(minStr[0]))

	x += (number_width + whitespace_width) * mag
	drawNumber(c.screen, x, y, mag, string(minStr[1]))

	x += (number_width + whitespace_width) * mag
	drawSeparater(c.screen, x, y, mag)

	x += (separater_width + whitespace_width) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[0]))

	x += (number_width + whitespace_width) * mag
	drawNumber(c.screen, x, y, mag, string(secStr[1]))
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
