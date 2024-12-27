package draw

import (
	"github.com/gdamore/tcell/v2"
)

// Option for optoin of drawing.
type Option func(tcell.Style) tcell.Style

// WithBackgroundColor set BackgroundColor.
func WithBackgroundColor(color tcell.Color) Option {
	return func(s tcell.Style) tcell.Style {
		return s.Background(color)
	}
}

// WithForegroundColor set BackgroundColor.
func WithForegroundColor(color tcell.Color) Option {
	return func(s tcell.Style) tcell.Style {
		return s.Foreground(color)
	}
}
