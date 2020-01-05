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
	timer        timer.Timer
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(s tcell.Screen, options ...Option) (Pomodoro, error) {
	c := screen.NewClient(s)
	c.StartPollEvent()

	taskName, err := task.GetTask(c)
	if err != nil {
		return nil, err
	}

	p := &pomodoroImpl{
		workSec:       DefaultWorkSec,
		shortBreakSec: DefaultShortBreakSec,
		longBreakSec:  DefaultLongBreakSec,
		screenClient:  c,
		timer:         timer.NewTimer(c, taskName),
	}

	for _, opt := range options {
		opt(p)
	}

	return p, nil
}

func (p *pomodoroImpl) Start() error {
	loopCnt := 1
	defer p.screenClient.StopPollEvent()
	for {
		w, h := p.screenClient.ScreenSize()

		if loopCnt%2 == 0 {
			p.timer.ChangeFontColor(tcell.ColorBlue)
		} else {
			p.timer.ChangeFontColor(tcell.ColorGreen)
		}
		err := p.timer.Run(p.getDuration(loopCnt))
		if err != nil {
			return err
		}

		draw.Sentence(
			p.screenClient.GetScreen(),
			0,
			h-1,
			w,
			"(Enter): continue / (c): change task",
			true,
			draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
		)
	L:
		for {
			select {
			case <-p.screenClient.GetEnterChan():
				break L
			case <-p.screenClient.GetCancelChan():
				return nil
			case r := <-p.screenClient.GetRuneChan():
				if r == rune(99) { // c
					p.screenClient.Clear()
					t, err := task.GetTask(p.screenClient)
					if err != nil {
						return err
					}
					p.timer.SetTitle(t)
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
