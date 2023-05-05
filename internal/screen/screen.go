// Package screen represents screen
package screen

import (
	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/internal/config"
)

// NewScreen initilize screen.
func NewScreen(config *config.Config) (tcell.Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(
		tcell.StyleDefault.Foreground(
			config.Color.Font,
		).Background(
			config.Color.Background,
		),
	)
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	return s, nil
}
