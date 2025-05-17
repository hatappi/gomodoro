// Package graphql provides a GraphQL client implementation for interacting with the Gomodoro GraphQL API
package graphql

//go:generate go run github.com/Khan/genqlient ./genqlient.yaml

import (
	"context"
	"fmt"
	"sync"
	"time"

	gqllib "github.com/Khan/genqlient/graphql"

	"github.com/hatappi/gomodoro/internal/core/event"
)

const (
	// defaultChannelBufferSize is the default buffer size for event and error channels.
	defaultChannelBufferSize = 10
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
	input EventReceivedInput,
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
	wsRespChan, id, err := OnEventReceived(ctx, c.subscriptionClient, input)
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
			case *OnEventReceivedEventReceivedEventPayloadEventPomodoroPayload:
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
			case *OnEventReceivedEventReceivedEventPayloadEventTaskPayload:
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

func convertEventTypeToEvent(eventType EventType) (event.EventType, error) {
	switch eventType {
	case EventTypePomodoroStarted:
		return event.PomodoroStarted, nil
	case EventTypePomodoroPaused:
		return event.PomodoroPaused, nil
	case EventTypePomodoroResumed:
		return event.PomodoroResumed, nil
	case EventTypePomodoroCompleted:
		return event.PomodoroCompleted, nil
	case EventTypePomodoroStopped:
		return event.PomodoroStopped, nil
	case EventTypePomodoroTick:
		return event.PomodoroTick, nil
	case EventTypeTaskCreated:
		return event.TaskCreated, nil
	case EventTypeTaskUpdated:
		return event.TaskUpdated, nil
	case EventTypeTaskDeleted:
		return event.TaskDeleted, nil
	case EventTypeTaskCompleted:
		return event.TaskCompleted, nil
	default:
		return event.EventType(""), fmt.Errorf("unknown event type: %s", eventType)
	}
}

func convertPomodoroStateToEvent(state PomodoroState) (event.PomodoroState, error) {
	switch state {
	case PomodoroStateActive:
		return event.PomodoroStateActive, nil
	case PomodoroStatePaused:
		return event.PomodoroStatePaused, nil
	case PomodoroStateFinished:
		return event.PomodoroStateFinished, nil
	default:
		return event.PomodoroState(""), fmt.Errorf("unknown pomodoro state: %s", state)
	}
}

func convertPomodoroPhaseToEvent(phase PomodoroPhase) (event.PomodoroPhase, error) {
	switch phase {
	case PomodoroPhaseWork:
		return event.PomodoroPhaseWork, nil
	case PomodoroPhaseShortBreak:
		return event.PomodoroPhaseShortBreak, nil
	case PomodoroPhaseLongBreak:
		return event.PomodoroPhaseLongBreak, nil
	default:
		return event.PomodoroPhase(""), fmt.Errorf("unknown pomodoro phase: %s", phase)
	}
}
