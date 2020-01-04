// Package screen represents screen
package screen

import (
	"github.com/gdamore/tcell"
)

// NewScreen initilize screen
func NewScreen() (tcell.Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).Background(tcell.ColorWhite))
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	return s, nil
}
