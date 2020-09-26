// Package pomodoro technique
package pomodoro

import (
	"context"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
	"github.com/hatappi/gomodoro/internal/task"
	"github.com/hatappi/gomodoro/internal/timer"
)

// Pomodoro interface
type Pomodoro interface {
	Start(context.Context) error
	Stop()

	Finish()
}

type pomodoroImpl struct {
	config        *config.Config
	workSec       int
	shortBreakSec int
	longBreakSec  int

	screenClient screen.Client
	taskClient   task.Client
	timer        timer.Timer

	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime int)
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(conf *config.Config, screenClient screen.Client, timer timer.Timer, taskClient task.Client, options ...Option) Pomodoro {
	p := &pomodoroImpl{
		config:        conf,
		workSec:       config.DefaultWorkSec,
		shortBreakSec: config.DefaultShortBreakSec,
		longBreakSec:  config.DefaultLongBreakSec,
		screenClient:  screenClient,
		taskClient:    taskClient,
		timer:         timer,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *pomodoroImpl) Start(ctx context.Context) error {
	task, err := p.taskClient.GetTask()
	if err != nil {
		return err
	}
	p.timer.SetTitle(task.Name)

	loopCnt := 1
	for {
		isWorkTime := loopCnt%2 != 0

		if isWorkTime {
			p.timer.ChangeFontColor(p.config.Color.TimerWorkFont)
		} else {
			p.timer.ChangeFontColor(p.config.Color.TimerBreakFont)
		}

		p.timer.SetDuration(p.getDuration(loopCnt))
		elapsedTime, err := p.timer.Run(ctx)
		if err != nil {
			return err
		}

		for _, cf := range p.completeFuncs {
			go cf(ctx, p.timer.GetTitle(), isWorkTime, elapsedTime)
		}

		if err := p.selectNextAction(); err != nil {
			return err
		}

		loopCnt++
	}
}

func (p *pomodoroImpl) Stop() {
}

func (p *pomodoroImpl) Finish() {
	p.screenClient.Finish()
}

func (p *pomodoroImpl) getDuration(cnt int) int {
	setNum := cnt / 2

	switch {
	case setNum != 0 && cnt%2 == 0 && setNum%3 == 0:
		return p.longBreakSec
	case cnt%2 == 0:
		return p.shortBreakSec
	default:
		return p.workSec
	}
}

// selectNextAction selects next action.
// e.g create new task, use same task
func (p *pomodoroImpl) selectNextAction() error {
	w, h := p.screenClient.ScreenSize()
	draw.Sentence(
		p.screenClient.GetScreen(),
		0,
		h-1,
		w,
		"(Enter): continue / (c): change task / (d): delete task",
		true,
		draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
	)

	for {
		e := <-p.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventEnter:
			// use Same Task
			return nil
		case screen.EventCancel:
			return errors.ErrCancel
		case screen.EventRune:
			if rune(e) == rune(99) { // c
				p.screenClient.Clear()
				t, err := p.taskClient.GetTask()
				if err != nil {
					return err
				}
				p.timer.SetTitle(t.Name)
				return nil
			}
		}
	}
}
