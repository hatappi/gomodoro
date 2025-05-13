// Package tui provides the terminal user interface for gomodoro
package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/client/graphql"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/notify"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/toggl"
	"github.com/hatappi/gomodoro/internal/tui/constants"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/view"
)

// Constants for timer control.
const (
	continueTimerSignal = -1 // Signal to continue timer processing
)

// App is the main TUI application controller.
type App struct {
	// Configuration and clients
	config         *config.Config
	screenClient   screen.Client
	pomodoroClient *client.PomodoroClient
	taskAPIClient  *client.TaskClient
	graphqlClient  *graphql.ClientWrapper

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

// Option is a function that configures the App.
type Option func(*App)

// WithWorkSec sets the work duration in seconds.
func WithWorkSec(s int) Option {
	return func(a *App) {
		a.workSec = s
	}
}

// WithShortBreakSec sets the short break duration in seconds.
func WithShortBreakSec(s int) Option {
	return func(a *App) {
		a.shortBreakSec = s
	}
}

// WithLongBreakSec sets the long break duration in seconds.
func WithLongBreakSec(s int) Option {
	return func(a *App) {
		a.longBreakSec = s
	}
}

// WithNotify adds desktop notification functionality.
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

// WithRecordToggl adds Toggl time tracking functionality.
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

// WithRecordPixela adds Pixela tracking functionality.
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

// NewApp creates a new TUI application instance.
func NewApp(cfg *config.Config, clientFactory *client.Factory, opts ...Option) (*App, error) {
	pomodoroClient := clientFactory.Pomodoro()
	taskClient := clientFactory.Task()

	gqlClient := clientFactory.GraphQLClient()

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
		graphqlClient:  gqlClient,
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

// Run starts the TUI application main loop.
func (a *App) Run(ctx context.Context) error {
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	a.screenClient.StartPollEvent(ctx)

	connectionErrChan, err := a.graphqlClient.ConnectSubscription(ctx)
	if err != nil {
		return err
	}

	go func() {
		if connectionErrChan == nil {
			return
		}

		for err := range connectionErrChan {
			log.FromContext(ctx).Info("Subscription connection error", "error", err)
			cancelCtx()
			return
		}
	}()

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
			elapsedTime, err := a.runTimer(ctx, task.Title)
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
			go cf(ctx, task.Title, pomodoro.Phase == event.PomodoroPhaseWork, res.elapsedTime)
		}

		action, err := a.pomodoroView.SelectNextTask(ctx)
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
		case constants.PomodoroActionNone:
			// no action
		}
	}
}

// Finish cleans up resources when the app is closed.
func (a *App) Finish(ctx context.Context) {
	_, _ = a.pomodoroClient.Stop(ctx)
	a.screenClient.Finish()
}

// selectTask handles task selection and creation.
func (a *App) selectTask(ctx context.Context) (*core.Task, error) {
	tasks, err := a.loadTasks(ctx)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return a.handleNewTask(ctx)
	}

	task, action, err := a.taskView.SelectTaskName(ctx, tasks)
	if err != nil {
		return nil, err
	}

	return a.processTaskAction(ctx, task, action)
}

// processTaskAction handles the action selected by the user for a task.
func (a *App) processTaskAction(ctx context.Context, task *core.Task, action constants.TaskAction) (*core.Task, error) {
	switch action {
	case constants.TaskActionCancel:
		return nil, gomodoro_error.ErrCancel

	case constants.TaskActionDelete:
		return a.handleDeleteTask(ctx, task)

	case constants.TaskActionNew:
		return a.handleNewTask(ctx)

	case constants.TaskActionNone:
		// No action, return selected task
		return task, nil
	}

	return task, nil
}

// handleDeleteTask deletes a task and returns a new selected task.
func (a *App) handleDeleteTask(ctx context.Context, task *core.Task) (*core.Task, error) {
	if task == nil {
		//nolint:nilnil
		return nil, nil
	}

	if err := a.deleteTask(ctx, task.ID); err != nil {
		return nil, err
	}

	tasks, err := a.loadTasks(ctx)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return a.handleNewTask(ctx)
	}

	task, _, err = a.taskView.SelectTaskName(ctx, tasks)
	return task, err
}

// handleNewTask creates a new task.
func (a *App) handleNewTask(ctx context.Context) (*core.Task, error) {
	name, err := a.taskView.CreateTaskName(ctx)
	if err != nil {
		return nil, err
	}

	task, err := a.createTask(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := a.saveTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// loadTasks loads tasks from the API.
func (a *App) loadTasks(ctx context.Context) ([]*core.Task, error) {
	responses, err := a.taskAPIClient.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make([]*core.Task, len(responses))
	for i, resp := range responses {
		tasks[i] = &core.Task{
			ID:        resp.ID,
			Title:     resp.Title,
			CreatedAt: resp.CreatedAt,
			Completed: resp.Completed,
		}
	}
	return tasks, nil
}

// createTask creates a new task.
func (a *App) createTask(ctx context.Context, name string) (*core.Task, error) {
	resp, err := a.taskAPIClient.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	task := &core.Task{
		ID:        resp.ID,
		Title:     resp.Title,
		CreatedAt: resp.CreatedAt,
		Completed: resp.Completed,
	}

	return task, nil
}

// deleteTask deletes a task by ID.
func (a *App) deleteTask(ctx context.Context, taskID string) error {
	return a.taskAPIClient.Delete(ctx, taskID)
}

// saveTask saves changes to a task.
func (a *App) saveTask(ctx context.Context, task *core.Task) error {
	var resp *client.TaskResponse
	var err error

	if task.ID == "" {
		// Create new task
		resp, err = a.taskAPIClient.Create(ctx, task.Title)
	} else {
		// Update existing task
		resp, err = a.taskAPIClient.Update(ctx, task.ID, task.Title, task.Completed)
	}

	if err != nil {
		return err
	}

	// On success, set returned ID on task
	task.ID = resp.ID

	return nil
}

// runTimer handles the timer display and events.
func (a *App) runTimer(ctx context.Context, taskName string) (int, error) {
	eventChan, subID, err := a.graphqlClient.SubscribeToEvents(ctx, graphql.EventReceivedInput{
		EventCategory: []graphql.EventCategory{graphql.EventCategoryPomodoro},
	})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := a.graphqlClient.Unsubscribe(subID); err != nil {
			log.FromContext(ctx).Error(err, "failed to unsubscribe from events")
		}
	}()

	for {
		select {
		case e := <-a.screenClient.GetEventChan():
			elapsedTime, err := a.handleScreenEvent(ctx, e)
			if err != nil {
				return elapsedTime, err
			}
			if elapsedTime != continueTimerSignal {
				return elapsedTime, nil
			}

		case eventData, ok := <-eventChan:
			if !ok {
				continue
			}

			ev, ok := eventData.(event.PomodoroEvent)
			if !ok {
				continue
			}

			elapsedTime, err := a.handlePomodoroEvent(ctx, ev, taskName)
			if err != nil {
				return elapsedTime, err
			}
			if elapsedTime != continueTimerSignal {
				return elapsedTime, nil
			}
		}
	}
}

// handleScreenEvent processes screen events and returns elapsed time and error if action is completed.
func (a *App) handleScreenEvent(ctx context.Context, e interface{}) (int, error) {
	action, err := a.timerView.HandleScreenEvent(ctx, e)
	if err != nil {
		if errors.Is(err, gomodoro_error.ErrCancel) {
			elapsedTime, timeErr := a.getCurrentElapsedTime(ctx)
			if timeErr != nil {
				log.FromContext(ctx).Error(timeErr, "failed to get current elapsed time")
				return 0, gomodoro_error.ErrCancel
			}
			return elapsedTime, gomodoro_error.ErrCancel
		}
		return 0, err
	}

	switch action {
	case constants.TimerActionCancel:
		elapsedTime, err := a.getCurrentElapsedTime(ctx)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to get current elapsed time")
			return 0, gomodoro_error.ErrCancel
		}
		return elapsedTime, gomodoro_error.ErrCancel
	case constants.TimerActionStop:
		_, stopErr := a.pomodoroClient.Stop(ctx)
		if stopErr != nil {
			log.FromContext(ctx).Error(stopErr, "failed to stop pomodoro")
			return 0, stopErr
		}
	case constants.TimerActionToggle:
		a.toggleTimer(ctx)
	case constants.TimerActionNone:
		// no action
	}

	return continueTimerSignal, nil // Signal to continue processing
}

// handlePomodoroEvent processes pomodoro events and handles UI rendering.
func (a *App) handlePomodoroEvent(ctx context.Context, ev event.PomodoroEvent, taskName string) (int, error) {
	log.FromContext(ctx).Info("event", "event", ev, "remainSec", ev.RemainingTime.Seconds())

	remainSec := int(ev.RemainingTime.Seconds())

	err := a.timerView.DrawTimer(ctx, remainSec, taskName, ev.Phase, ev.Type == event.PomodoroPaused)
	if err != nil {
		if !errors.Is(err, gomodoro_error.ErrScreenSmall) {
			return 0, err
		}

		return a.handleSmallScreen(ctx, taskName)
	}

	if ev.Type == event.PomodoroCompleted || ev.Type == event.PomodoroStopped {
		elapsedTime, err := a.getCurrentElapsedTime(ctx)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to get current elapsed time")
			return 0, err
		}
		return elapsedTime, nil
	}

	return continueTimerSignal, nil // Signal to continue processing
}

// handleSmallScreen handles the case when the screen is too small.
func (a *App) handleSmallScreen(ctx context.Context, taskName string) (int, error) {
	a.screenClient.Clear()
	w, h := a.screenClient.ScreenSize()
	if err := a.errorView.DrawSmallScreen(ctx, w, h); err != nil {
		log.FromContext(ctx).Error(err, "failed to draw small screen")
	}

	for {
		e := <-a.screenClient.GetEventChan()
		switch e.(type) {
		case screen.EventCancel:
			elapsedTime, err := a.getCurrentElapsedTime(ctx)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed to get current elapsed time")
				return 0, gomodoro_error.ErrCancel
			}
			return elapsedTime, gomodoro_error.ErrCancel
		case screen.EventScreenResize:
			return a.runTimer(ctx, taskName)
		}
	}
}

// getCurrentElapsedTime retrieves elapsed time from current pomodoro session.
func (a *App) getCurrentElapsedTime(ctx context.Context) (int, error) {
	current, err := a.pomodoroClient.GetCurrent(ctx)
	if err != nil {
		return 0, err
	}
	return current.ElapsedTime, nil
}

// toggleTimer toggles the timer between paused and running states.
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
