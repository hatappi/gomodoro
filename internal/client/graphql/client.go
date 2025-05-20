// Package graphql provides a GraphQL client implementation for interacting with the Gomodoro GraphQL API
package graphql

//go:generate go run github.com/Khan/genqlient ./genqlient.yaml

import (
	"context"
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

	underlyingGorillaDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: defaultHandshakeTimeout,
	}
	wsDialerAdapter := NewGorillaWebSocketDialer(underlyingGorillaDialer)

	subscriptionClient := gqllib.NewClientUsingWebSocket(
		fmt.Sprintf("ws://%s/graphql/query", apiConfig.Addr),
		wsDialerAdapter,
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

	if c.subscriptionClient == nil {
		return nil, fmt.Errorf("subscription client is not initialized")
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

	if c.subscriptionClient == nil {
		return fmt.Errorf("subscription client is not initialized, cannot disconnect")
	}

	err := c.subscriptionClient.Close()
	if err != nil {
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

	if c.subscriptionClient == nil {
		return nil, nil, "", fmt.Errorf("subscription client is not initialized")
	}

	// Get the raw WebSocket response channel
	wsRespChan, id, err := gqlgen.OnEventReceived(ctx, c.subscriptionClient, input)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to subscribe to events: %w", err)
	}

	eventChan := make(chan event.EventInfo, defaultChannelBufferSize)
	errChan := make(chan error, defaultChannelBufferSize)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		for wsResp := range wsRespChan {
			if wsResp.Errors != nil {
				errChan <- wsResp.Errors
				return
			}

			evt := wsResp.Data.EventReceived

			eventType, err := convertEventTypeToEvent(evt.EventType)
			if err != nil {
				errChan <- err
				return
			}

			switch payload := evt.Payload.(type) {
			case *gqlgen.OnEventReceivedEventReceivedEventPayloadEventPomodoroPayload:
				state, err := convertPomodoroStateToEvent(payload.State)
				if err != nil {
					errChan <- err
					return
				}

				phase, err := convertPomodoroPhaseToEvent(payload.Phase)
				if err != nil {
					errChan <- err
					return
				}

				eventChan <- event.PomodoroEvent{
					BaseEvent: event.BaseEvent{
						Type:      eventType,
						Timestamp: time.Now(),
					},
					ID:            payload.Id,
					State:         state,
					RemainingTime: time.Duration(payload.RemainingTime) * time.Second,
					ElapsedTime:   time.Duration(payload.ElapsedTime) * time.Second,
					TaskID:        payload.TaskId,
					Phase:         phase,
					PhaseCount:    payload.PhaseCount,
				}
			case *gqlgen.OnEventReceivedEventReceivedEventPayloadEventTaskPayload:
				eventChan <- event.TaskEvent{
					BaseEvent: event.BaseEvent{
						Type:      eventType,
						Timestamp: time.Now(),
					},
					ID:        payload.Id,
					Title:     payload.Title,
					Completed: payload.Completed,
				}
			}
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

	if c.subscriptionClient == nil {
		return fmt.Errorf("subscription client is not initialized")
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

	tasks := res.GetTasks()

	result := make([]*core.Task, 0, len(tasks.GetEdges()))
	for _, task := range tasks.GetEdges() {
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

	task := res.GetTask()

	return conv.ToCoreTask(task.TaskDetails), nil
}

// CreateTask creates a new task on the server.
func (c *ClientWrapper) CreateTask(ctx context.Context, title string) (*core.Task, error) {
	res, err := gqlgen.CreateTask(ctx, c.queryClient, title)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	task := res.CreateTask.TaskDetails

	return conv.ToCoreTask(task), nil
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

	pomodoro := res.GetCurrentPomodoro()

	return conv.ToCorePomodoro(pomodoro.PomodoroDetails)
}

// StartPomodoro starts a new pomodoro session on the server.
func (c *ClientWrapper) StartPomodoro(ctx context.Context, input gqlgen.StartPomodoroInput) (*core.Pomodoro, error) {
	res, err := gqlgen.StartPomodoro(ctx, c.queryClient, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start pomodoro: %w", err)
	}

	pomodoro := res.StartPomodoro.PomodoroDetails

	return conv.ToCorePomodoro(pomodoro)
}

// PausePomodoro pauses the current active pomodoro session on the server.
func (c *ClientWrapper) PausePomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.PausePomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to pause pomodoro: %w", err)
	}

	pomodoro := res.PausePomodoro.PomodoroDetails

	return conv.ToCorePomodoro(pomodoro)
}

// ResumePomodoro resumes a paused pomodoro session on the server.
func (c *ClientWrapper) ResumePomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.ResumePomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resume pomodoro: %w", err)
	}

	pomodoro := res.ResumePomodoro.PomodoroDetails

	return conv.ToCorePomodoro(pomodoro)
}

// StopPomodoro stops the current active pomodoro session on the server.
func (c *ClientWrapper) StopPomodoro(ctx context.Context) (*core.Pomodoro, error) {
	res, err := gqlgen.StopPomodoro(ctx, c.queryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to stop pomodoro: %w", err)
	}

	pomodoro := res.StopPomodoro.PomodoroDetails

	return conv.ToCorePomodoro(pomodoro)
}

func convertEventTypeToEvent(eventType gqlgen.EventType) (event.EventType, error) {
	switch eventType {
	case gqlgen.EventTypePomodoroStarted:
		return event.PomodoroStarted, nil
	case gqlgen.EventTypePomodoroPaused:
		return event.PomodoroPaused, nil
	case gqlgen.EventTypePomodoroResumed:
		return event.PomodoroResumed, nil
	case gqlgen.EventTypePomodoroCompleted:
		return event.PomodoroCompleted, nil
	case gqlgen.EventTypePomodoroStopped:
		return event.PomodoroStopped, nil
	case gqlgen.EventTypePomodoroTick:
		return event.PomodoroTick, nil
	case gqlgen.EventTypeTaskCreated:
		return event.TaskCreated, nil
	case gqlgen.EventTypeTaskUpdated:
		return event.TaskUpdated, nil
	case gqlgen.EventTypeTaskDeleted:
		return event.TaskDeleted, nil
	case gqlgen.EventTypeTaskCompleted:
		return event.TaskCompleted, nil
	default:
		return event.EventType(""), fmt.Errorf("unknown event type: %s", eventType)
	}
}

func convertPomodoroStateToEvent(state gqlgen.PomodoroState) (event.PomodoroState, error) {
	switch state {
	case gqlgen.PomodoroStateActive:
		return event.PomodoroStateActive, nil
	case gqlgen.PomodoroStatePaused:
		return event.PomodoroStatePaused, nil
	case gqlgen.PomodoroStateFinished:
		return event.PomodoroStateFinished, nil
	default:
		return event.PomodoroState(""), fmt.Errorf("unknown pomodoro state: %s", state)
	}
}

func convertPomodoroPhaseToEvent(phase gqlgen.PomodoroPhase) (event.PomodoroPhase, error) {
	switch phase {
	case gqlgen.PomodoroPhaseWork:
		return event.PomodoroPhaseWork, nil
	case gqlgen.PomodoroPhaseShortBreak:
		return event.PomodoroPhaseShortBreak, nil
	case gqlgen.PomodoroPhaseLongBreak:
		return event.PomodoroPhaseLongBreak, nil
	default:
		return event.PomodoroPhase(""), fmt.Errorf("unknown pomodoro phase: %s", phase)
	}
}
