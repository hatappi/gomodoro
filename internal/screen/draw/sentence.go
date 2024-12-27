package draw

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"
)

// Sentence is draw the sentence.
func Sentence(s tcell.Screen, x, y, width int, str string, centered bool, opts ...Option) int {
	if centered {
		str = adjustMessage(width, str)
	}

	style := tcell.StyleDefault
	for _, opt := range opts {
		style = opt(style)
	}

	i := 0
	deferred := []rune{}
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
		case 2: //nolint:mnd
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
	}

	s.Show()
	return x + i + dwidth
}

func adjustMessage(width int, str string) string {
	margin := width - runewidth.StringWidth(str)
	if margin >= 0 {
		lm := margin / 2 //nolint:mnd
		rm := width - runewidth.StringWidth(str) - lm
		return fmt.Sprintf("%s%s%s", strings.Repeat(" ", lm), str, strings.Repeat(" ", rm))
	}

	return fmt.Sprintf("%s...", str[:width-3])
}
