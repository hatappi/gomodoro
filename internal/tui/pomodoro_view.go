package tui

import (
	"context"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/domain/model"
	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/screen/draw"
)

type PomodoroView struct {
	config       *config.Config
	screenClient screen.Client
}

func NewPomodoroView(cfg *config.Config, sc screen.Client) *PomodoroView {
	return &PomodoroView{
		config:       cfg,
		screenClient: sc,
	}
}

func (v *PomodoroView) SelectNextTask(ctx context.Context, currentTask *model.Task) (PomodoroAction, error) {
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
			return PomodoroActionContinue, nil
		case screen.EventCancel:
			return PomodoroActionCancel, errors.ErrCancel
		case screen.EventRune:
			if string(e) == "c" {
				v.screenClient.Clear()
				return PomodoroActionChange, nil
			}
		}
	}
}
