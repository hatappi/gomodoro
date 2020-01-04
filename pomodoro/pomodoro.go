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
	screen tcell.Screen

	workSec       int
	shortBreakSec int
	longBreakSec  int

	screenClient screen.Client
	timer        timer.Timer
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(options ...Option) (Pomodoro, error) {
	s, err := screen.NewScreen()
	if err != nil {
		return nil, err
	}

	taskName := task.GetTask(s)

	c := screen.NewClient(s)

	p := &pomodoroImpl{
		workSec:       DefaultWorkSec,
		shortBreakSec: DefaultShortBreakSec,
		longBreakSec:  DefaultLongBreakSec,
		screenClient:  c,
		timer:         timer.NewTimer(c, taskName),
		screen:        s,
	}

	for _, opt := range options {
		opt(p)
	}

	return p, nil
}

func (p *pomodoroImpl) Start() error {
	loopCnt := 1
	for {
		if loopCnt%2 == 0 {
			p.timer.ChangeFontColor(tcell.ColorBlue)
		} else {
			p.timer.ChangeFontColor(tcell.ColorGreen)
		}
		p.screenClient.StartPollEvent()
		err := p.timer.Run(p.getDuration(loopCnt))
		if err != nil {
			return err
		}
		defer p.screenClient.StopPollEvent()
		if p.timer.IsQuit() {
			return nil
		}

		w, h := p.screen.Size()
		draw.Sentence(
			p.screen,
			0,
			h-1,
			w,
			"Please press Enter button for continue",
			draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
		)
	L:
		for {
			select {
			case <-p.screenClient.GetEnterChan():
				break L
			case <-p.screenClient.GetCancelChan():
				return nil
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
