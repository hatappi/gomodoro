// Package pomodoro technique
package pomodoro

import (
	"github.com/gdamore/tcell"
	"github.com/hatappi/gomodoro/timer"
	"github.com/hatappi/gomodoro/timer/screen"
)

const (
	//DefaultWorkSec       = 1500
	//DefaultShortBreakSec = 300
	//DefaultLongBreakSec  = 900
	DefaultWorkSec       = 10
	DefaultShortBreakSec = 3
	DefaultLongBreakSec  = 6
)

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

func NewPomodoro(options ...PomodoroOption) (Pomodoro, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).Background(tcell.ColorWhite))
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	c, err := screen.NewClient(s)
	if err != nil {
		return nil, err
	}

	p := &pomodoroImpl{
		workSec:           DefaultWorkSec,
		shortBreakSec:     DefaultShortBreakSec,
		longBreakSec:      DefaultLongBreakSec,
		timerScreenClient: c,
		timer:             timer.NewTimer(c),
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
		p.timerScreenClient.StartPollEvent()
		err := p.timer.Run(p.getDuration(loopCnt))
		if err != nil {
			return err
		}
		p.timerScreenClient.StopPollEvent()
		if p.timer.IsQuit() {
			return nil
		}

		w, _ := p.screen.Size()
		// TODO: ここからtimerScreenClientを呼んでるのは微妙なのでなおす
		p.timerScreenClient.DrawSentence(0, 0, w, "Pleaes press Enter Button")

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
	if setNum != 0 && cnt%2 == 0 && setNum%3 == 0 {
		return p.longBreakSec
	} else if cnt%2 == 0 {
		return p.shortBreakSec
	} else {
		return p.workSec
	}
}
