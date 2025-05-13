package graphql

import (
	"context"
	"fmt"
	"sync"
	"time"

	gqllib "github.com/Khan/genqlient/graphql"
	"github.com/hatappi/gomodoro/internal/core/event"
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
func NewClientWrapper(queryClient gqllib.Client, subscriptionClient gqllib.WebSocketClient) *ClientWrapper {
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
	input EventReceivedInput,
) (eventStream <-chan interface{}, subscriptionID string, err error) {
	c.subscriptionClientMu.Lock()
	isStarted := c.isSubscriptionClientStarted
	c.subscriptionClientMu.Unlock()

	if !isStarted {
		return nil, "", fmt.Errorf("subscription client not started. Call ConnectSubscription first")
	}

	if c.subscriptionClient == nil {
		return nil, "", fmt.Errorf("subscription client is not initialized")
	}

	// Get the raw WebSocket response channel
	wsRespChan, id, err := OnEventReceived(ctx, c.subscriptionClient, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to subscribe to events: %w", err)
	}

	eventChan := make(chan interface{}, 10)

	go func() {
		defer close(eventChan)

		for wsResp := range wsRespChan {
			if wsResp.Data == nil || wsResp.Errors != nil {
				continue
			}

			event := wsResp.Data.EventReceived
			payload := event.Payload

			switch payload.GetTypename() {
			case "EventPomodoroPayload":
				if pomodoroPayload, ok := payload.(*OnEventReceivedEventReceivedEventPayloadEventPomodoroPayload); ok {
					domainEventType := convertToDomainEventType(event.EventType)
					eventChan <- event.PomodoroState(domainEventType, pomodoroPayload.State, pomodoroPayload.Id, pomodoroPayload.TaskId, pomodoroPayload.RemainingTime, pomodoroPayload.ElapsedTime, pomodoroPayload.Phase, pomodoroPayload.PhaseCount)
				}
			case "EventTaskPayload":
				if taskPayload, ok := payload.(*OnEventReceivedEventReceivedEventPayloadEventTaskPayload); ok {
					domainEventType := convertToDomainEventType(event.EventType)
					eventChan <- event.TaskState(domainEventType, taskPayload.Id, taskPayload.Title, taskPayload.Completed)
				}
			}
		}
	}()

	return eventChan, id, nil
}

// Convert GraphQL EventType to domain EventType
func convertToDomainEventType(eventType EventType) event.EventType {
	switch eventType {
	case EventTypePomodoroStarted:
		return event.PomodoroStarted
	case EventTypePomodoroPaused:
		return event.PomodoroPaused
	case EventTypePomodoroResumed:
		return event.PomodoroResumed
	case EventTypePomodoroCompleted:
		return event.PomodoroCompleted
	case EventTypePomodoroStopped:
		return event.PomodoroStopped
	case EventTypePomodoroTick:
		return event.PomodoroTick
	case EventTypeTaskCreated:
		return event.TaskCreated
	case EventTypeTaskUpdated:
		return event.TaskUpdated
	case EventTypeTaskDeleted:
		return event.TaskDeleted
	case EventTypeTaskCompleted:
		return event.TaskCompleted
	default:
		return event.EventType("")
	}
}

// Helper to convert GraphQL pomodoro event to domain PomodoroEvent
func (OnEventReceivedEventReceivedEvent) PomodoroState(
	eventType event.EventType,
	state PomodoroState,
	id string,
	taskID string,
	remainingTime int,
	elapsedTime int,
	phase PomodoroPhase,
	phaseCount int,
) event.PomodoroEvent {
	// Map GraphQL enum values to domain enum values
	var domainState event.PomodoroState
	switch state {
	case PomodoroStateActive:
		domainState = event.PomodoroStateActive
	case PomodoroStatePaused:
		domainState = event.PomodoroStatePaused
	case PomodoroStateFinished:
		domainState = event.PomodoroStateFinished
	}

	var domainPhase event.PomodoroPhase
	switch phase {
	case PomodoroPhaseWork:
		domainPhase = event.PomodoroPhaseWork
	case PomodoroPhaseShortBreak:
		domainPhase = event.PomodoroPhaseShortBreak
	case PomodoroPhaseLongBreak:
		domainPhase = event.PomodoroPhaseLongBreak
	}

	return event.PomodoroEvent{
		BaseEvent: event.BaseEvent{
			Type:      eventType,
			Timestamp: time.Now(),
		},
		ID:            id,
		State:         domainState,
		RemainingTime: time.Duration(remainingTime) * time.Second,
		ElapsedTime:   time.Duration(elapsedTime) * time.Second,
		TaskID:        taskID,
		Phase:         domainPhase,
		PhaseCount:    phaseCount,
	}
}

func (e OnEventReceivedEventReceivedEvent) TaskState(
	eventType event.EventType,
	id string,
	title string,
	completed bool,
) event.TaskEvent {
	return event.TaskEvent{
		BaseEvent: event.BaseEvent{
			Type:      eventType,
			Timestamp: time.Now(),
		},
		ID:        id,
		Title:     title,
		Completed: completed,
	}
}

func mapEventType(state event.PomodoroState) event.EventType {
	switch state {
	case event.PomodoroStateActive:
		return event.PomodoroStarted
	case event.PomodoroStatePaused:
		return event.PomodoroPaused
	case event.PomodoroStateFinished:
		return event.PomodoroCompleted
	default:
		return event.PomodoroTick
	}
}

func mapTaskEventType(eventType EventType) event.EventType {
	switch eventType {
	case EventTypeTaskCreated:
		return event.TaskCreated
	case EventTypeTaskUpdated:
		return event.TaskUpdated
	case EventTypeTaskDeleted:
		return event.TaskDeleted
	case EventTypeTaskCompleted:
		return event.TaskCompleted
	default:
		return event.TaskUpdated
	}
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
