// Package view provides UI components for the TUI
package view

import (
	"context"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/screen/draw"
)

// ErrorView handles error message displays
type ErrorView struct {
	config       *config.Config
	screenClient screen.Client
}

// NewErrorView creates a new error view instance
func NewErrorView(cfg *config.Config, sc screen.Client) *ErrorView {
	return &ErrorView{
		config:       cfg,
		screenClient: sc,
	}
}

// DrawSmallScreen displays a message when the screen size is too small
func (v *ErrorView) DrawSmallScreen(ctx context.Context, w, h int) error {
	screen := v.screenClient.GetScreen()

	//nolint:mnd
	draw.Sentence(screen, 0, h/2, w, "Please expand the screen size", true)

	return nil
}
