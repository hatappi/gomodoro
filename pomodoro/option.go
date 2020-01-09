// Package pomodoro Option
package pomodoro

import (
	"time"

	"github.com/hatappi/gomodoro/logger"
	"github.com/hatappi/gomodoro/notify"
	"github.com/hatappi/gomodoro/toggl"
)

// Option pomodoro option
type Option func(*pomodoroImpl)

// WithWorkSec set WorkSec
func WithWorkSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.workSec = s
	}
}

// WithShortBreakSec set ShortBreakSec
func WithShortBreakSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.shortBreakSec = s
	}
}

// WithLongBreakSec set LongBreakSec
func WithLongBreakSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.longBreakSec = s
	}
}

func WithNotify() Option {
	return func(p *pomodoroImpl) {
		p.completeFuncs = append(p.completeFuncs, func(taskName string, isWorkTime bool, elapsedTime int) {
			var message string
			if isWorkTime {
				message = "Finish work time"
			} else {
				message = "Finish brek time"
			}

			err := notify.Notify("gomodoro", taskName+":"+message)
			if err != nil {
				logger.Warnf("failed to notify: %s", err)
			}
		})
	}
}

func WithRecordToggl(togglClient *toggl.Client) Option {
	return func(p *pomodoroImpl) {
		p.completeFuncs = append(p.completeFuncs, func(taskName string, isWorkTime bool, elapsedTime int) {
			if !isWorkTime {
				return
			}

			s := time.Now().Add(-time.Duration(elapsedTime) * time.Second)
			err := togglClient.PostTimeEntry(taskName, s, elapsedTime)
			if err != nil {
				logger.Warnf("failed to notify: %s", err)
			}
		})
	}
}
