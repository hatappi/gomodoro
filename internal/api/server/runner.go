package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/storage/file"
)

// ServerRunner manages API server lifecycle
type ServerRunner struct {
	config    *config.Config
	server    *Server
	isRunning bool
	mu        sync.Mutex
}

// NewServerRunner creates a new server runner
func NewServerRunner(config *config.Config) *ServerRunner {
	return &ServerRunner{
		config: config,
	}
}

// Start initializes and starts the API server
func (r *ServerRunner) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRunning {
		return nil
	}

	logger := log.FromContext(ctx)

	fileStorage, err := file.NewFileStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize file storage: %w", err)
	}

	eventBus := event.NewServerWebSocketEventBus()

	taskService := core.NewTaskService(fileStorage, eventBus)
	pomodoroService := core.NewPomodoroService(fileStorage, eventBus)

	r.server = NewServer(
		&r.config.API,
		logger,
		pomodoroService,
		taskService,
		eventBus,
	)

	ln, err := r.server.Listen()
	if err != nil {
		return err
	}

	latest, err := pomodoroService.GetLatestPomodoro()
	if err == nil && latest != nil {
		if err := pomodoroService.DeletePomodoro(ctx, latest.ID); err != nil {
			logger.Error(err, "Failed to delete latest pomodoro")
		}
	}

	go func() {
		if err := r.server.Serve(ln); err != nil {
			logger.Error(err, "Error serving API")
		}
	}()

	r.isRunning = true
	return nil
}

// Stop gracefully stops the API server
func (r *ServerRunner) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.isRunning || r.server == nil {
		r.mu.Unlock()
		return nil
	}
	server := r.server
	r.mu.Unlock()

	logger := log.FromContext(ctx)
	logger.Info("Stopping API server...")

	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := server.Stop(stopCtx)

	r.mu.Lock()
	r.isRunning = false
	r.server = nil
	r.mu.Unlock()

	return err
}

// IsRunning returns true if the server is running
func (r *ServerRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isRunning
}

// EnsureRunning checks if the API server is running and starts it if not.
// It uses the client to perform health checks.
func (r *ServerRunner) EnsureRunning(ctx context.Context) error {
	logger := log.FromContext(ctx)
	clientFactory := client.NewFactory(r.config.API)
	defer clientFactory.Close()

	_, err := clientFactory.Pomodoro().GetCurrent(ctx)
	if err == nil {
		logger.Info("API server is already running (checked via client)")
		return nil
	}
	logger.Info("API server health check failed, attempting to start...", "error", err.Error())

	if startErr := r.Start(ctx); startErr != nil {
		logger.Error(startErr, "Failed to start API server via runner")
		return fmt.Errorf("failed to start API server via runner: %w", startErr)
	}
	logger.Info("API server started successfully by this runner")

	r.isRunning = true

	return nil
}
