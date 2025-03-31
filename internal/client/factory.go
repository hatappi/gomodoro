// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"fmt"
	"time"

	"github.com/hatappi/gomodoro/internal/config"
)

// Factory provides a convenient way to create and manage API clients
type Factory struct {
	pomodoro *PomodoroClient
	task     *TaskClient
	wsClient *WebSocketClientImpl
}

// NewFactory creates a new client factory with the given API configuration and options
func NewFactory(apiConfig config.APIConfig) *Factory {
	httpClientOpts := []Option{
		WithBaseURL(fmt.Sprintf("http://%s", apiConfig.Addr)),
		WithTimeout(10 * time.Second),
	}

	factory := &Factory{
		wsClient: NewWebSocketClient(fmt.Sprintf("ws://%s/api/events/ws", apiConfig.Addr)),
		pomodoro: NewPomodoroClient(httpClientOpts...),
		task:     NewTaskClient(httpClientOpts...),
	}

	return factory
}

// Pomodoro returns a client for interacting with Pomodoro API endpoints
func (f *Factory) Pomodoro() *PomodoroClient {
	return f.pomodoro
}

// Task returns a client for interacting with Task API endpoints
func (f *Factory) Task() *TaskClient {
	return f.task
}

// WebSocket returns a client for interacting with WebSocket endpoint
func (f *Factory) WebSocket() (*WebSocketClientImpl, error) {
	if err := f.wsClient.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket: %w", err)
	}
	return f.wsClient, nil
}

// Close closes all open connections, including the WebSocket connection.
func (f *Factory) Close() error {
	if f.wsClient != nil {
		if err := f.wsClient.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket client: %w", err)
		}
	}
	return nil
}
