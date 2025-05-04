// Package view provides UI components for the TUI
package view

import (
	"context"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/domain/model"
	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui/constants"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/screen/draw"
)

// PomodoroView handles pomodoro session UI components
type PomodoroView struct {
	config       *config.Config
	screenClient screen.Client
}

// NewPomodoroView creates a new pomodoro view instance
func NewPomodoroView(cfg *config.Config, sc screen.Client) *PomodoroView {
	return &PomodoroView{
		config:       cfg,
		screenClient: sc,
	}
}

// SelectNextTask displays options for continuing or changing tasks after a pomodoro cycle
func (v *PomodoroView) SelectNextTask(ctx context.Context, currentTask *model.Task) (constants.PomodoroAction, error) {
	w, h := v.screenClient.ScreenSize()
	draw.Sentence(
		v.screenClient.GetScreen(),
		0,
		h-1,
		w,
		"(Enter): continue / (c): change task",
		true,
		draw.WithBackgroundColor(v.config.Color.StatusBarBackground),
	)

	for {
		e := <-v.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventEnter:
			return constants.PomodoroActionContinue, nil
		case screen.EventCancel:
			return constants.PomodoroActionCancel, errors.ErrCancel
		case screen.EventRune:
			if string(e) == "c" {
				v.screenClient.Clear()
				return constants.PomodoroActionChange, nil
			}
		}
	}
}
