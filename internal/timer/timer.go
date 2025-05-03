// Package timer countdown duration
package timer

import (
	"context"
	"errors"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui"
	"github.com/hatappi/gomodoro/internal/tui/screen"
)

// Timer interface.
type Timer interface {
	Run(ctx context.Context, taskName string) (int, error)
}

// ITimer implements Timer interface.
type ITimer struct {
	config         *config.Config
	screenClient   screen.Client
	pomodoroClient *client.PomodoroClient
	eventBus       event.EventBus
}

// NewTimer initilize Timer.
func NewTimer(config *config.Config, c screen.Client, pomodoroClient *client.PomodoroClient, eventBus event.EventBus) *ITimer {
	return &ITimer{
		config:         config,
		screenClient:   c,
		pomodoroClient: pomodoroClient,
		eventBus:       eventBus,
	}
}

// getCurrentPomodoro retrieves the current pomodoro and handles errors.
func (t *ITimer) getCurrentPomodoro(ctx context.Context) (int, error) {
	current, err := t.pomodoroClient.GetCurrent(ctx)
	if err != nil {
		return 0, err
	}
	return current.ElapsedTime, nil
}

// Run timer.
func (t *ITimer) Run(ctx context.Context, taskName string) (int, error) {
	timerView := tui.NewTimerView(t.config, t.screenClient)
	errorView := tui.NewErrorView(t.config, t.screenClient)

	ch, unsubscribe := t.eventBus.SubscribeChannel([]string{
		string(event.PomodoroTick),
		string(event.PomodoroPaused),
		string(event.PomodoroStarted),
		string(event.PomodoroStopped),
		string(event.PomodoroCompleted),
	})
	defer unsubscribe()

	for {
		select {
		case e := <-t.screenClient.GetEventChan():
			action, err := timerView.HandleScreenEvent(ctx, e)
			if err != nil {
				if errors.Is(err, gomodoro_error.ErrCancel) {
					elapsedTime, _ := t.getCurrentPomodoro(ctx)
					return elapsedTime, gomodoro_error.ErrCancel
				}
				return 0, err
			}

			switch action {
			case tui.TimerActionCancel:
				elapsedTime, _ := t.getCurrentPomodoro(ctx)
				return elapsedTime, gomodoro_error.ErrCancel
			case tui.TimerActionStop:
				_, stopErr := t.pomodoroClient.Stop(ctx)
				if stopErr != nil {
					log.FromContext(ctx).Error(stopErr, "failed to stop pomodoro")
					return 0, stopErr
				}
			case tui.TimerActionToggle:
				t.Toggle(ctx)
			}

		case e := <-ch:
			ev, ok := e.(event.PomodoroEvent)
			if !ok {
				continue
			}
			log.FromContext(ctx).Info("event", "event", ev, "remainSec", ev.RemainingTime.Seconds())

			remainSec := int(ev.RemainingTime.Seconds())

			err := timerView.DrawTimer(ctx, remainSec, taskName, ev.Phase, ev.Type == event.PomodoroPaused)
			if err != nil {
				if errors.Is(err, gomodoro_error.ErrScreenSmall) {
					t.screenClient.Clear()
					w, h := t.screenClient.ScreenSize()

					errorView.DrawSmallScreen(ctx, w, h)

					for {
						select {
						case e := <-t.screenClient.GetEventChan():
							switch e.(type) {
							case screen.EventCancel:
								elapsedTime, _ := t.getCurrentPomodoro(ctx)
								return elapsedTime, gomodoro_error.ErrCancel
							case screen.EventScreenResize:
								break
							}
						}
						break
					}
				} else {
					return 0, err
				}
			}

			if ev.Type == event.PomodoroCompleted || ev.Type == event.PomodoroStopped {
				elapsedTime, _ := t.getCurrentPomodoro(ctx)
				return elapsedTime, nil
			}
		}
	}
}

// Toggle timer between stop and start.
func (t *ITimer) Toggle(ctx context.Context) {
	currPomodoro, _ := t.pomodoroClient.GetCurrent(ctx)
	if currPomodoro == nil {
		return
	}

	if currPomodoro.State == "paused" {
		_, _ = t.pomodoroClient.Resume(ctx)
	} else {
		_, _ = t.pomodoroClient.Pause(ctx)
	}
}
