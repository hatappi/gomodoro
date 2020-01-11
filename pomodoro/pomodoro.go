// Package pomodoro technique
package pomodoro

import (
	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/screen"
	"github.com/hatappi/gomodoro/screen/draw"
	"github.com/hatappi/gomodoro/task"
	"github.com/hatappi/gomodoro/timer"
)

const (
	// DefaultWorkSec default working second
	DefaultWorkSec = 1500
	// DefaultShortBreakSec default short break second
	DefaultShortBreakSec = 300
	// DefaultLongBreakSec default long break second
	DefaultLongBreakSec = 900
)

// Pomodoro interface
type Pomodoro interface {
	Start() error
	Stop()
	Finish()
}

type pomodoroImpl struct {
	workSec       int
	shortBreakSec int
	longBreakSec  int

	screenClient screen.Client
	taskClient   task.Client
	timer        timer.Timer

	completeFuncs []func(taskName string, isWorkTime bool, elapsedTime int)
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(s tcell.Screen, taskFile string, options ...Option) Pomodoro {
	c := screen.NewClient(s)
	c.StartPollEvent()

	p := &pomodoroImpl{
		workSec:       DefaultWorkSec,
		shortBreakSec: DefaultShortBreakSec,
		longBreakSec:  DefaultLongBreakSec,
		screenClient:  c,
		taskClient:    task.NewClient(c, taskFile),
		timer:         timer.NewTimer(c),
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *pomodoroImpl) Start() error {
	task, err := p.taskClient.GetTask()
	if err != nil {
		return err
	}
	p.timer.SetTitle(task.Name)

	loopCnt := 1
	for {
		w, h := p.screenClient.ScreenSize()

		isWorkTime := loopCnt%2 != 0

		if isWorkTime {
			p.timer.ChangeFontColor(tcell.ColorGreen)
		} else {
			p.timer.ChangeFontColor(tcell.ColorBlue)
		}

		p.timer.SetDuration(p.getDuration(loopCnt))
		elapsedTime, err := p.timer.Run()
		if err != nil {
			return err
		}

		for _, cf := range p.completeFuncs {
			go cf(p.timer.GetTitle(), isWorkTime, elapsedTime)
		}

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
