// Package pomodoro technique
package pomodoro

import (
	"context"

	"github.com/gdamore/tcell"

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
func NewPomodoro(conf *config.Config, sc screen.Client, timer timer.Timer, tc task.Client, opts ...Option) Pomodoro {
	p := &pomodoroImpl{
		config:        conf,
		workSec:       config.DefaultWorkSec,
		shortBreakSec: config.DefaultShortBreakSec,
		longBreakSec:  config.DefaultLongBreakSec,
		screenClient:  sc,
		taskClient:    tc,
		timer:         timer,
	}

	for _, opt := range opts {
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

	var loopCnt int
	for {
		isWorkTime := loopCnt%2 == 0
		intervalNum := loopCnt/2 + 1

		p.changeTimerConfig(isWorkTime, intervalNum)

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

func (p *pomodoroImpl) changeTimerConfig(isWorkTime bool, intervalNum int) {
	var (
		timerColor    tcell.Color
		timerDuration int
	)

	if isWorkTime {
		timerColor = p.config.Color.TimerWorkFont
		timerDuration = p.workSec
	} else {
		timerColor = p.config.Color.TimerBreakFont
		if intervalNum%3 == 0 {
			timerDuration = p.longBreakSec
		} else {
			timerDuration = p.shortBreakSec
		}
	}

	p.timer.SetFontColor(timerColor)
	p.timer.SetDuration(timerDuration)
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
		"(Enter): continue / (c): change task",
		true,
		draw.WithBackgroundColor(p.config.Color.StatusBarBackground),
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
