// Package tui provides the terminal user interface for gomodoro
package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/domain/model"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/notify"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/toggl"
	"github.com/hatappi/gomodoro/internal/tui/constants"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/view"
)

// App is the main TUI application controller
type App struct {
	// Configuration and clients
	config         *config.Config
	screenClient   screen.Client
	pomodoroClient *client.PomodoroClient
	taskAPIClient  *client.TaskClient
	wsClient       event.WebSocketClient
	eventBus       event.EventBus

	// View components
	timerView    *view.TimerView
	taskView     *view.TaskView
	pomodoroView *view.PomodoroView
	errorView    *view.ErrorView

	// Pomodoro settings
	workSec       int
	shortBreakSec int
	longBreakSec  int

	// Completion handlers
	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime int)
}

// Option is a function that configures the App
type Option func(*App)

// WithWorkSec sets the work duration in seconds
func WithWorkSec(s int) Option {
	return func(a *App) {
		a.workSec = s
	}
}

// WithShortBreakSec sets the short break duration in seconds
func WithShortBreakSec(s int) Option {
	return func(a *App) {
		a.shortBreakSec = s
	}
}

// WithLongBreakSec sets the long break duration in seconds
func WithLongBreakSec(s int) Option {
	return func(a *App) {
		a.longBreakSec = s
	}
}

// WithNotify adds desktop notification functionality
func WithNotify() Option {
	return func(a *App) {
		a.completeFuncs = append(
			a.completeFuncs,
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

// WithRecordToggl adds Toggl time tracking functionality
func WithRecordToggl(togglClient *toggl.Client) Option {
	return func(a *App) {
		a.completeFuncs = append(
			a.completeFuncs,
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

// WithRecordPixela adds Pixela tracking functionality
func WithRecordPixela(client *pixela.Client, userName, graphID string) Option {
	return func(a *App) {
		a.completeFuncs = append(
			a.completeFuncs,
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

// NewApp creates a new TUI application instance
func NewApp(cfg *config.Config, clientFactory *client.Factory, opts ...Option) (*App, error) {
	pomodoroClient := clientFactory.Pomodoro()
	taskClient := clientFactory.Task()
	wsClient, err := clientFactory.WebSocket()
	if err != nil {
		return nil, fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	terminalScreen, err := screen.NewScreen(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}
	screenClient := screen.NewClient(terminalScreen)

	app := &App{
		config:         cfg,
		screenClient:   screenClient,
		pomodoroClient: pomodoroClient,
		taskAPIClient:  taskClient,
		wsClient:       wsClient,
		eventBus:       event.NewClientWebSocketEventBus(wsClient),
		workSec:        config.DefaultWorkSec,
		shortBreakSec:  config.DefaultShortBreakSec,
		longBreakSec:   config.DefaultLongBreakSec,
	}

	// Apply all options
	for _, opt := range opts {
		opt(app)
	}

	// Initialize views
	app.timerView = view.NewTimerView(cfg, screenClient)
	app.taskView = view.NewTaskView(cfg, screenClient)
	app.pomodoroView = view.NewPomodoroView(cfg, screenClient)
	app.errorView = view.NewErrorView(cfg, screenClient)

	return app, nil
}

// Run starts the TUI application main loop
func (a *App) Run(ctx context.Context) error {
	a.screenClient.StartPollEvent(ctx)

	task, err := a.selectTask(ctx)
	if err != nil {
		return err
	}

	workDuration := time.Duration(a.workSec) * time.Second
	breakDuration := time.Duration(a.shortBreakSec) * time.Second
	longBreakDuration := time.Duration(a.longBreakSec) * time.Second

	for {
		type timerResult struct {
			elapsedTime int
			err         error
		}
		resultCh := make(chan timerResult, 1)
		go func() {
			elapsedTime, err := a.runTimer(ctx, task.Name)
			resultCh <- timerResult{elapsedTime: elapsedTime, err: err}
		}()

		pomodoro, err := a.pomodoroClient.Start(ctx, workDuration, breakDuration, longBreakDuration, task.ID)
		if err != nil {
			return err
		}

		res := <-resultCh
		if res.err != nil {
			return res.err
		}

		log.FromContext(ctx).Info("Pomodoro finished", "elapsedTime", res.elapsedTime, "err", nil)

		// Execute completion functions
		for _, cf := range a.completeFuncs {
			go cf(ctx, task.Name, pomodoro.Phase == event.PomodoroPhaseWork, res.elapsedTime)
		}

		action, err := a.pomodoroView.SelectNextTask(ctx, task)
		if err != nil {
			return err
		}

		switch action {
		case constants.PomodoroActionCancel:
			return gomodoro_error.ErrCancel
		case constants.PomodoroActionContinue:
			// Continue with the same task
		case constants.PomodoroActionChange:
			// Change task
			newTask, err := a.selectTask(ctx)
			if err != nil {
				return err
			}
			task = newTask
		}
	}
}

// Finish cleans up resources when the app is closed
func (a *App) Finish() {
	_, _ = a.pomodoroClient.Stop(context.Background())
	a.screenClient.Finish()
}

// selectTask handles task selection and creation
func (a *App) selectTask(ctx context.Context) (*model.Task, error) {
	tasks, err := a.loadTasks(ctx)
	if err != nil {
		return nil, err
	}

	var task *model.Task

	if len(tasks) > 0 {
		var action constants.TaskAction
		task, action, err = a.taskView.SelectTaskName(ctx, tasks)
		if err != nil {
			return nil, err
		}

		switch action {
		case constants.TaskActionCancel:
			return nil, gomodoro_error.ErrCancel
		case constants.TaskActionDelete:
			if task != nil {
				// Delete task
				if err := a.deleteTask(ctx, task.ID); err != nil {
					return nil, err
				}

				// Reload tasks after deletion
				tasks, err = a.loadTasks(ctx)
				if err != nil {
					return nil, err
				}

				// If no tasks left, create a new one
				if len(tasks) == 0 {
					task = nil
				} else {
					// Reselect after deletion
					task, _, err = a.taskView.SelectTaskName(ctx, tasks)
					if err != nil {
						return nil, err
					}
				}
			}
		case constants.TaskActionNew:
			task = nil
		}
	}

	if task == nil {
		// Create new task
		name, err := a.taskView.CreateTaskName(ctx)
		if err != nil {
			return nil, err
		}

		task, err = a.createTask(ctx, name)
		if err != nil {
			return nil, err
		}

		if err := a.saveTask(ctx, task); err != nil {
			return nil, err
		}
	}

	return task, nil
}

// loadTasks loads tasks from the API
func (a *App) loadTasks(ctx context.Context) (model.Tasks, error) {
	responses, err := a.taskAPIClient.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make(model.Tasks, len(responses))
	for i, resp := range responses {
		tasks[i] = &model.Task{
			ID:        resp.ID,
			Name:      resp.Title,
			Completed: resp.Completed,
		}
	}
	return tasks, nil
}

// createTask creates a new task
func (a *App) createTask(ctx context.Context, name string) (*model.Task, error) {
	resp, err := a.taskAPIClient.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	task := &model.Task{
		ID:        resp.ID,
		Name:      resp.Title,
		Completed: resp.Completed,
	}

	return task, nil
}

// deleteTask deletes a task by ID
func (a *App) deleteTask(ctx context.Context, taskID string) error {
	return a.taskAPIClient.Delete(ctx, taskID)
}

// saveTask saves changes to a task
func (a *App) saveTask(ctx context.Context, task *model.Task) error {
	var resp *client.TaskResponse
	var err error

	if task.ID == "" {
		// Create new task
		resp, err = a.taskAPIClient.Create(ctx, task.Name)
	} else {
		// Update existing task
		resp, err = a.taskAPIClient.Update(ctx, task.ID, task.Name, task.Completed)
	}

	if err != nil {
		return err
	}

	// On success, set returned ID on task
	task.ID = resp.ID

	return nil
}

// runTimer handles the timer display and events
func (a *App) runTimer(ctx context.Context, taskName string) (int, error) {
	ch, unsubscribe := a.eventBus.SubscribeChannel([]string{
		string(event.PomodoroTick),
		string(event.PomodoroPaused),
		string(event.PomodoroStarted),
		string(event.PomodoroStopped),
		string(event.PomodoroCompleted),
	})
	defer unsubscribe()

	for {
		select {
		case e := <-a.screenClient.GetEventChan():
			action, err := a.timerView.HandleScreenEvent(ctx, e)
			if err != nil {
				if err == gomodoro_error.ErrCancel {
					elapsedTime, _ := a.getCurrentPomodoro(ctx)
					return elapsedTime, gomodoro_error.ErrCancel
				}
				return 0, err
			}

			switch action {
			case constants.TimerActionCancel:
				elapsedTime, _ := a.getCurrentPomodoro(ctx)
				return elapsedTime, gomodoro_error.ErrCancel
			case constants.TimerActionStop:
				_, stopErr := a.pomodoroClient.Stop(ctx)
				if stopErr != nil {
					log.FromContext(ctx).Error(stopErr, "failed to stop pomodoro")
					return 0, stopErr
				}
			case constants.TimerActionToggle:
				a.toggleTimer(ctx)
			}

		case e := <-ch:
			ev, ok := e.(event.PomodoroEvent)
			if !ok {
				continue
			}
			log.FromContext(ctx).Info("event", "event", ev, "remainSec", ev.RemainingTime.Seconds())

			remainSec := int(ev.RemainingTime.Seconds())

			err := a.timerView.DrawTimer(ctx, remainSec, taskName, ev.Phase, ev.Type == event.PomodoroPaused)
			if err != nil {
				if err == gomodoro_error.ErrScreenSmall {
					a.screenClient.Clear()
					w, h := a.screenClient.ScreenSize()

					a.errorView.DrawSmallScreen(ctx, w, h)

					for {
						select {
						case e := <-a.screenClient.GetEventChan():
							switch e.(type) {
							case screen.EventCancel:
								elapsedTime, _ := a.getCurrentPomodoro(ctx)
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
				elapsedTime, _ := a.getCurrentPomodoro(ctx)
				return elapsedTime, nil
			}
		}
	}
}

// getCurrentPomodoro retrieves current pomodoro info from API
func (a *App) getCurrentPomodoro(ctx context.Context) (int, error) {
	current, err := a.pomodoroClient.GetCurrent(ctx)
	if err != nil {
		return 0, err
	}
	return current.ElapsedTime, nil
}

// toggleTimer toggles the timer between paused and running states
func (a *App) toggleTimer(ctx context.Context) {
	currPomodoro, _ := a.pomodoroClient.GetCurrent(ctx)
	if currPomodoro == nil {
		return
	}

	if currPomodoro.State == "paused" {
		_, _ = a.pomodoroClient.Resume(ctx)
	} else {
		_, _ = a.pomodoroClient.Pause(ctx)
	}
}
