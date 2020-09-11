// Package pomodoro Option
package pomodoro

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/notify"
	"github.com/hatappi/gomodoro/internal/toggl"
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

// WithNotify notify and sound when time is finished
func WithNotify(ctx context.Context) Option {
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
				log.FromContext(ctx).Warn("failed to notify", zap.Error(err))
			}
		})
	}
}

// WithRecordToggl record duration when work time is finished
func WithRecordToggl(ctx context.Context, togglClient *toggl.Client) Option {
	return func(p *pomodoroImpl) {
		p.completeFuncs = append(p.completeFuncs, func(taskName string, isWorkTime bool, elapsedTime int) {
			if !isWorkTime {
				return
			}

			s := time.Now().Add(-time.Duration(elapsedTime) * time.Second)

			if err := togglClient.PostTimeEntry(taskName, s, elapsedTime); err != nil {
				log.FromContext(ctx).Warn("failed to record time to toggle", zap.Error(err))
			}
		})
	}
}
