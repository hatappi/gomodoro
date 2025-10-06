// Package graphql provides a GraphQL client implementation for interacting with the Gomodoro GraphQL API
package graphql

//go:generate go run github.com/Khan/genqlient ./genqlient.yaml

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	gqllib "github.com/Khan/genqlient/graphql"
	"github.com/gorilla/websocket"

	"github.com/hatappi/gomodoro/internal/client/graphql/conv"
	gqlgen "github.com/hatappi/gomodoro/internal/client/graphql/generated"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
)

const (
	// defaultChannelBufferSize is the default buffer size for event and error channels.
	defaultChannelBufferSize = 10

	// defaultHandshakeTimeout is the default timeout for WebSocket handshaking.
	defaultHandshakeTimeout = 45 * time.Second // Added constant
)

// ClientWrapper wraps the genqlient clients for query/mutation and subscriptions.
// It provides a unified interface for GraphQL operations.
type ClientWrapper struct {
	queryClient        gqllib.Client
	subscriptionClient gqllib.WebSocketClient

	// Mutex and state for managing the subscription client's lifecycle.
	subscriptionClientMu        sync.Mutex
	isSubscriptionClientStarted bool
}

// NewClientWrapper creates a new GraphQL ClientWrapper.
func NewClientWrapper(apiConfig config.APIConfig) *ClientWrapper {
	queryClient := gqllib.NewClient(fmt.Sprintf("http://%s/graphql/query", apiConfig.Addr), http.DefaultClient)

	subscriptionClient := gqllib.NewClientUsingWebSocket(
		fmt.Sprintf("ws://%s/graphql/query", apiConfig.Addr),
		NewGorillaWebSocketDialer(&websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: defaultHandshakeTimeout,
		}),
	)

	return &ClientWrapper{
		queryClient:        queryClient,
		subscriptionClient: subscriptionClient,
	}
}

// ConnectSubscription starts the WebSocket connection for subscriptions.
func (c *ClientWrapper) ConnectSubscription(ctx context.Context) (<-chan error, error) {
	c.subscriptionClientMu.Lock()
	defer c.subscriptionClientMu.Unlock()

	if c.isSubscriptionClientStarted {
		//nolint:nilnil
		return nil, nil
	}

	errChan, err := c.subscriptionClient.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start subscription client: %w", err)
	}

	c.isSubscriptionClientStarted = true

	return errChan, nil
}

// DisconnectSubscription closes the WebSocket connection.
func (c *ClientWrapper) DisconnectSubscription() error {
	c.subscriptionClientMu.Lock()
	defer c.subscriptionClientMu.Unlock()

	if !c.isSubscriptionClientStarted {
		return nil
	}

	err := c.subscriptionClient.Close()

	// This error indicates that the client was already closed.
	// It's safe to ignore this error.
	if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		return fmt.Errorf("failed to close subscription client: %w", err)
	}

	c.isSubscriptionClientStarted = false

	return nil
}

// SubscribeToEvents subscribes to real-time events.
func (c *ClientWrapper) SubscribeToEvents(
	ctx context.Context,
	input gqlgen.EventReceivedInput,
) (<-chan event.EventInfo, <-chan error, string, error) {
	c.subscriptionClientMu.Lock()
	isStarted := c.isSubscriptionClientStarted
	c.subscriptionClientMu.Unlock()

	if !isStarted {
		return nil, nil, "", fmt.Errorf("subscription client not started. Call ConnectSubscription first")
	}

	gqlEventChan, id, err := gqlgen.OnEventReceived(ctx, c.subscriptionClient, input)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to subscribe to events: %w", err)
	}

	eventChan := make(chan event.EventInfo, defaultChannelBufferSize)
	errChan := make(chan error, defaultChannelBufferSize)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		for gqlEvent := range gqlEventChan {
			if gqlEvent.Errors != nil {
				errChan <- gqlEvent.Errors
				return
			}

			evt, err := conv.ToEventInfo(gqlEvent.Data.EventReceived.EventDetails)
			if err != nil {
				errChan <- fmt.Errorf("failed to convert event: %w", err)
				return
			}

			eventChan <- evt
		}
	}()

	return eventChan, errChan, id, nil
}

// Unsubscribe cancels a specific event subscription by its ID.
func (c *ClientWrapper) Unsubscribe(subscriptionID string) error {
	c.subscriptionClientMu.Lock()
	isStarted := c.isSubscriptionClientStarted
	c.subscriptionClientMu.Unlock()

	if !isStarted {
		return fmt.Errorf("subscription client not started, cannot unsubscribe")
	}

	err := c.subscriptionClient.Unsubscribe(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe (ID: %s): %w", subscriptionID, err)
	}

	return nil
}

// GetAllTasks retrieves all tasks from the server.
func (c *ClientWrapper) GetAllTasks(ctx context.Context) ([]*core.Task, error) {
	res, err := gqlgen.GetAllTasks(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	tasks := res.Tasks.Edges

	result := make([]*core.Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, conv.ToCoreTask(task.Node.TaskDetails))
	}

	return result, nil
}

// GetTask retrieves a task by ID from the server.
func (c *ClientWrapper) GetTask(ctx context.Context, id string) (*core.Task, error) {
	res, err := gqlgen.GetTask(ctx, c.queryClient, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return conv.ToCoreTask(res.Task.TaskDetails), nil
}

// CreateTask creates a new task on the server.
func (c *ClientWrapper) CreateTask(ctx context.Context, title string) (*core.Task, error) {
	res, err := gqlgen.CreateTask(ctx, c.queryClient, title)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return conv.ToCoreTask(res.CreateTask.TaskDetails), nil
}

// DeleteTask deletes a task on the server.
func (c *ClientWrapper) DeleteTask(ctx context.Context, id string) error {
	res, err := gqlgen.DeleteTask(ctx, c.queryClient, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if res.DeleteTask {
		return nil
	}

	return fmt.Errorf("failed to delete task")
}

// GetCurrentPomodoro retrieves the current active pomodoro session from the server.
func (c *ClientWrapper) GetCurrentPomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.GetCurrentPomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.GetCurrentPomodoro().PomodoroDetails)
}

// StartPomodoro starts a new pomodoro session on the server.
func (c *ClientWrapper) StartPomodoro(ctx context.Context, input gqlgen.StartPomodoroInput) (*core.Pomodoro, error) {
	res, err := gqlgen.StartPomodoro(ctx, c.queryClient, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.StartPomodoro.PomodoroDetails)
}

// PausePomodoro pauses the current active pomodoro session on the server.
func (c *ClientWrapper) PausePomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.PausePomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to pause pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.PausePomodoro.PomodoroDetails)
}

// ResumePomodoro resumes a paused pomodoro session on the server.
func (c *ClientWrapper) ResumePomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.ResumePomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resume pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.ResumePomodoro.PomodoroDetails)
}

// StopPomodoro stops the current active pomodoro session on the server.
func (c *ClientWrapper) StopPomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.StopPomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to stop pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.StopPomodoro.PomodoroDetails)
}

// ResetPomodoro resets the current pomodoro session on the server.
func (c *ClientWrapper) ResetPomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.ResetPomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to reset pomodoro: %w", err)
	}

	return conv.ToCorePomodoro(res.ResetPomodoro.PomodoroDetails)
}
