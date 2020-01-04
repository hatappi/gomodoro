// Package pomodoro technique
package pomodoro

import (
	"github.com/gdamore/tcell"
	"github.com/hatappi/gomodoro/task"
	"github.com/hatappi/gomodoro/timer"
	"github.com/hatappi/gomodoro/timer/screen"
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

	timerScreenClient screen.Client
	timer             timer.Timer
}

// NewPomodoro initilize Pomodoro
func NewPomodoro(options ...Option) (Pomodoro, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).Background(tcell.ColorWhite))
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	taskName := task.GetTask(s)

	c, err := screen.NewClient(s)
	if err != nil {
		return nil, err
	}

	p := &pomodoroImpl{
		workSec:           DefaultWorkSec,
		shortBreakSec:     DefaultShortBreakSec,
		longBreakSec:      DefaultLongBreakSec,
		timerScreenClient: c,
		timer:             timer.NewTimer(c, taskName),
		screen:            s,
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
		p.timerScreenClient.StartPollEvent()
		err := p.timer.Run(p.getDuration(loopCnt))
		if err != nil {
			return err
		}
		p.timerScreenClient.StopPollEvent()
		if p.timer.IsQuit() {
			return nil
		}

		w, h := p.screen.Size()
		// TODO: ここからtimerScreenClientを呼んでるのは微妙なのでなおす
		p.timerScreenClient.DrawSentence(
			0,
			h-1,
			w,
			"Please press Enter button for continue",
			screen.WithBackgroundColor(tcell.ColorRed),
		)

	L:
		for {
			ev := p.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEnter:
					break L
				case tcell.KeyEscape, tcell.KeyCtrlC:
					return nil
				}
			}
		}
		loopCnt++
	}
}

func (p *pomodoroImpl) Stop() {
}

func (p *pomodoroImpl) Finish() {
	p.timerScreenClient.Finish()
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
