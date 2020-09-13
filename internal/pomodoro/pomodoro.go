// Package pomodoro technique
package pomodoro

import (
	"context"

	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
	"github.com/hatappi/gomodoro/internal/task"
	"github.com/hatappi/gomodoro/internal/timer"
)

// Pomodoro interface
type Pomodoro interface {
	Start(context.Context) error
	Stop()

	GetTimer() timer.Timer

	Finish()
}

type pomodoroImpl struct {
	workSec       int
	shortBreakSec int
	longBreakSec  int

	screenClient screen.Client
	taskClient   task.Client
	timer        timer.Timer

	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime int)
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(c screen.Client, taskFile string, options ...Option) Pomodoro {
	p := &pomodoroImpl{
		workSec:       config.DefaultWorkSec,
		shortBreakSec: config.DefaultShortBreakSec,
		longBreakSec:  config.DefaultLongBreakSec,
		screenClient:  c,
		taskClient:    task.NewClient(c, taskFile),
		timer:         timer.NewTimer(c),
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
			p.timer.ChangeFontColor(tcell.ColorGreen)
		} else {
			p.timer.ChangeFontColor(tcell.ColorBlue)
		}

		p.timer.SetDuration(p.getDuration(loopCnt))
		elapsedTime, err := p.timer.Run(ctx)
		if err != nil {
			return err
		}

		for _, cf := range p.completeFuncs {
			go cf(ctx, p.timer.GetTitle(), isWorkTime, elapsedTime)
		}

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
	L:
		for {
			e := <-p.screenClient.GetEventChan()
			switch e := e.(type) {
			case screen.EventEnter:
				break L
			case screen.EventCancel:
				return nil
			case screen.EventRune:
				if rune(e) == rune(99) { // c
					p.screenClient.Clear()
					t, err := p.taskClient.GetTask()
					if err != nil {
						return err
					}
					p.timer.SetTitle(t.Name)
					break L
				}
			}
		}
		loopCnt++
	}
}

func (p *pomodoroImpl) GetTimer() timer.Timer {
	return p.timer
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
