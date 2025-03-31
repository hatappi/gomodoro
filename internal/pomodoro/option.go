// Package pomodoro Option
package pomodoro

import (
	"context"
	"time"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/notify"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/toggl"
)

// Option pomodoro option.
type Option func(*IPomodoro)

// WithWorkSec set WorkSec.
func WithWorkSec(s int) Option {
	return func(p *IPomodoro) {
		p.workSec = s
	}
}

// WithShortBreakSec set ShortBreakSec.
func WithShortBreakSec(s int) Option {
	return func(p *IPomodoro) {
		p.shortBreakSec = s
	}
}

// WithLongBreakSec set LongBreakSec.
func WithLongBreakSec(s int) Option {
	return func(p *IPomodoro) {
		p.longBreakSec = s
	}
}

// WithNotify notify and sound when time is finished.
func WithNotify() Option {
	return func(p *IPomodoro) {
		p.completeFuncs = append(
			p.completeFuncs,
			func(ctx context.Context, taskName string, isWorkTime bool, _ int) {
				var message string
				if isWorkTime {
					message = "Finish work time"
				} else {
					message = "Finish break time"
				}

				if err := notify.Notify("gomodoro", taskName+":"+message); err != nil {
					log.FromContext(ctx).Error(err, "failed to notify")
				}
			},
		)
	}
}

// WithRecordToggl record duration when work time is finished.
func WithRecordToggl(togglClient *toggl.Client) Option {
	return func(p *IPomodoro) {
		p.completeFuncs = append(
			p.completeFuncs,
			func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime int) {
				if !isWorkTime {
					return
				}

				s := time.Now().Add(-time.Duration(elapsedTime) * time.Second)

				if err := togglClient.PostTimeEntry(ctx, taskName, s, elapsedTime); err != nil {
					log.FromContext(ctx).Error(err, "failed to record time to toggl")
				}
			},
		)
	}
}

// WithRecordPixela record pomodoro count when work time is finished.
func WithRecordPixela(client *pixela.Client, userName, graphID string) Option {
	return func(p *IPomodoro) {
		p.completeFuncs = append(
			p.completeFuncs,
			func(ctx context.Context, _ string, isWorkTime bool, _ int) {
				if !isWorkTime {
					return
				}

				if err := client.IncrementPixel(ctx, userName, graphID); err != nil {
					log.FromContext(ctx).Error(err, "failed to increment a pixel at Pixela")
				}
			},
		)
	}
}

// WithPomodoroClient sets the pomodoro API client
func WithPomodoroClient(pomodoroClient *client.PomodoroClient) Option {
	return func(p *IPomodoro) {
		p.pomodoroClient = pomodoroClient
	}
}

// WithPomodoroClient sets the pomodoro API client
func WithTaskClient(taskClient *client.TaskClient) Option {
	return func(p *IPomodoro) {
		p.taskAPIClient = taskClient
	}
}

// WithWebSocketClient sets the WebSocket client for real-time updates
func WithWebSocketClient(wsClient event.WebSocketClient) Option {
	return func(p *IPomodoro) {
		p.wsClient = wsClient
	}
}
