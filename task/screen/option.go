// Package screen provide screen management for task
package screen

import (
	"github.com/gdamore/tcell"
)

// DrawOption for optoin of drawing
type DrawOption func(tcell.Style) tcell.Style

// WithBackgroundColor set BackgroundColor
func WithBackgroundColor(color tcell.Color) DrawOption {
	return func(s tcell.Style) tcell.Style {
		return s.Background(color)
	}
}
