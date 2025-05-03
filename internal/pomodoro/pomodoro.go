// Package pomodoro technique
package pomodoro

import (
	"context"
	"fmt"
	"time"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/task"
	"github.com/hatappi/gomodoro/internal/timer"
	"github.com/hatappi/gomodoro/internal/tui"
	"github.com/hatappi/gomodoro/internal/tui/screen"
)

// Pomodoro interface.
type Pomodoro interface {
	Start(ctx context.Context) error
	Finish()
}

// IPomodoro implements Pomodoro interface.
type IPomodoro struct {
	config        *config.Config
	workSec       int
	shortBreakSec int
	longBreakSec  int

	screenClient screen.Client
	taskClient   task.Client
	timer        timer.Timer

	// API client related
	pomodoroClient *client.PomodoroClient
	taskAPIClient  *client.TaskClient
	wsClient       event.WebSocketClient
	eventBus       event.EventBus

	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime int)
}

// NewPomodoro initilize Pomodoro.
func NewPomodoro(
	cfg *config.Config,
	sc screen.Client,
	timer timer.Timer,
	tc task.Client,
	opts ...Option,
) *IPomodoro {
	p := &IPomodoro{
		config:        cfg,
		workSec:       config.DefaultWorkSec,
		shortBreakSec: config.DefaultShortBreakSec,
		longBreakSec:  config.DefaultLongBreakSec,
		screenClient:  sc,
		taskClient:    tc,
		timer:         timer,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.wsClient != nil {
		p.eventBus = event.NewClientWebSocketEventBus(p.wsClient)
	}

	return p
}

// Start starts pomodoro.
func (p *IPomodoro) Start(ctx context.Context) error {
	task, err := p.taskClient.GetTask(ctx)
	if err != nil {
		return err
	}

	pomodoroView := tui.NewPomodoroView(p.config, p.screenClient)

	workDuration := time.Duration(p.workSec) * time.Second
	breakDuration := time.Duration(p.shortBreakSec) * time.Second
	longBreakDuration := time.Duration(p.longBreakSec) * time.Second

	for {
		type timerResult struct {
			elapsedTime int
			err         error
		}
		resultCh := make(chan timerResult, 1)
		go func() {
			elapsedTime, err := p.timer.Run(ctx, task.Name)
			resultCh <- timerResult{elapsedTime: elapsedTime, err: err}
		}()

		pomodoro, err := p.pomodoroClient.Start(ctx, workDuration, breakDuration, longBreakDuration, task.ID)
		if err != nil {
			return fmt.Errorf("failed to start pomodoro via API: %w", err)
		}

		res := <-resultCh
		if res.err != nil {
			return res.err
		}
		log.FromContext(ctx).Info("Pomodoro finished", "elapsedTime", res.elapsedTime, "err", nil)

		for _, cf := range p.completeFuncs {
			go cf(ctx, task.Name, pomodoro.Phase == event.PomodoroPhaseWork, res.elapsedTime)
		}

		action, err := pomodoroView.SelectNextTask(ctx, task)
		if err != nil {
			return err
		}

		switch action {
		case tui.PomodoroActionCancel:
			return errors.ErrCancel
		case tui.PomodoroActionContinue:
			// Continue with the same task
		case tui.PomodoroActionChange:
			// Change task
			newTask, err := p.taskClient.GetTask(ctx)
			if err != nil {
				return err
			}
			task = newTask
		}
	}
}

// Finish finishes Pomodoro.
func (p *IPomodoro) Finish() {
	// Stop pomodoro
	if p.pomodoroClient != nil {
		_, _ = p.pomodoroClient.Stop(context.Background())
	}

	p.screenClient.Finish()
}
