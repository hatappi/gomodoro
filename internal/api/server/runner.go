package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client/graphql"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/storage/file"
	"github.com/hatappi/gomodoro/internal/toggl"
)

const (
	// serverShutdownTimeout is the maximum time to wait for the server to shut down gracefully.
	serverShutdownTimeout = 5 * time.Second
)

// Runner manages API server lifecycle.
type Runner struct {
	config    *config.Config
	server    *Server
	isRunning bool
	mu        sync.Mutex
}

// NewRunner creates a new server runner.
func NewRunner(config *config.Config) *Runner {
	return &Runner{
		config: config,
	}
}

// Start initializes and starts the API server.
func (r *Runner) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRunning {
		return nil
	}

	fileStorage, err := file.NewFileStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize file storage: %w", err)
	}

	eventBus := event.NewInMemoryBus()

	taskService := core.NewTaskService(fileStorage, eventBus)
	pomodoroService := core.NewPomodoroService(fileStorage, eventBus)

	opts := []Option{
		WithCompletionLogging(),
	}

	if r.config.Toggl.Enable {
		togglClient := toggl.NewClient(r.config.Toggl.ProjectID, r.config.Toggl.WorkspaceID, r.config.Toggl.APIToken)
		opts = append(opts, WithRecordToggl(togglClient))
	}

	if r.config.Pixela.Enable {
		pixelaClient := pixela.NewClient(r.config.Pixela.Token)
		opts = append(opts, WithRecordPixela(pixelaClient, r.config.Pixela.UserName, r.config.Pixela.GraphID))
	}

	r.server = NewServer(r.config.API, pomodoroService, taskService, eventBus, opts...)

	ln, err := r.server.Listen()
	if err != nil {
		return err
	}

	latest, err := pomodoroService.GetLatestPomodoro()
	if err != nil {
		return fmt.Errorf("failed to get latest pomodoro: %w", err)
	}

	// If there's an active pomodoro, delete it to clean up the state
	if latest != nil {
		if err := pomodoroService.DeletePomodoro(ctx, latest.ID); err != nil {
			log.FromContext(ctx).Error(err, "Failed to delete latest pomodoro")
		}
	}

	go func() {
		if err := r.server.Start(ctx, ln); err != nil {
			log.FromContext(ctx).Error(err, "Error serving API")
		}
	}()

	r.isRunning = true

	return nil
}

// Stop gracefully stops the API server.
func (r *Runner) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.isRunning || r.server == nil {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	stopCtx, cancel := context.WithTimeout(ctx, serverShutdownTimeout)
	defer cancel()

	err := r.server.Stop(stopCtx)

	r.mu.Lock()
	r.isRunning = false
	r.server = nil
	r.mu.Unlock()

	return err
}

// EnsureRunning checks if the API server is running and starts it if not.
// It uses the client to perform health checks.
func (r *Runner) EnsureRunning(ctx context.Context) error {
	gqlClient := graphql.NewClientWrapper(r.config.API)

	_, err := gqlClient.GetCurrentPomodoro(ctx)
	if err == nil {
		return nil
	}

	if startErr := r.Start(ctx); startErr != nil {
		return fmt.Errorf("failed to start API server via runner: %w", startErr)
	}

	return nil
}
