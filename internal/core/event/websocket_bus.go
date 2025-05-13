package event

import (
	"encoding/json"
)

const (
	// defaultChannelBufferSize is the default buffer size for event channels.
	defaultChannelBufferSize = 10
)

// WebSocketEvent is the event structure sent/received over WebSocket.
type WebSocketEvent struct {
	EventType string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// WebSocketEventBus implements the EventBus interface using WebSockets for communication.
type WebSocketEventBus struct {
	localBus *DefaultEventBus
}

// WebSocketClientConnection defines the interface for sending events through a WebSocket connection.
type WebSocketClientConnection interface {
	Send(event WebSocketEvent) error
}

// NewServerWebSocketEventBus creates a new EventBus for the server side that communicates over WebSockets.
//
//nolint:ireturn
func NewServerWebSocketEventBus() EventBus {
	return &WebSocketEventBus{
		localBus: NewEventBus(),
	}
}

// EventInfo defines the interface for event information that can be published.
//
//nolint:revive
type EventInfo interface {
	GetEventType() EventType
}

// Publish sends an event both to local subscribers and over WebSocket to remote subscribers.
func (b *WebSocketEventBus) Publish(event EventInfo) {
	b.localBus.Publish(event)
}

// Subscribe registers a handler for a specific event type.
func (b *WebSocketEventBus) Subscribe(eventType string, handler Handler) string {
	return b.localBus.Subscribe(eventType, handler)
}

// SubscribeMulti registers a handler for multiple event types.
func (b *WebSocketEventBus) SubscribeMulti(eventTypes []string, handler Handler) []string {
	subscriptionIDs := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		id := b.localBus.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}
	return subscriptionIDs
}

// SubscribeChannel creates a channel to receive events for multiple event types.
func (b *WebSocketEventBus) SubscribeChannel(eventTypes []string) (<-chan interface{}, func()) {
	ch := make(chan interface{}, defaultChannelBufferSize)
	subscriptionIDs := make([]string, 0, len(eventTypes))
	handler := func(e interface{}) {
		ch <- e
	}
	for _, eventType := range eventTypes {
		id := b.localBus.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}
	unsubscribe := func() {
		for _, id := range subscriptionIDs {
			b.localBus.Unsubscribe(id)
		}
		close(ch)
	}
	return ch, unsubscribe
}

// Unsubscribe removes a subscription by its ID.
func (b *WebSocketEventBus) Unsubscribe(subscriptionID string) {
	b.localBus.Unsubscribe(subscriptionID)
}

// handleRemoteEvent processes an event received from a remote WebSocket connection.
func (b *WebSocketEventBus) handleRemoteEvent(wsEvent WebSocketEvent) {
	switch wsEvent.EventType {
	case string(PomodoroStarted), string(PomodoroPaused), string(PomodoroResumed),
		string(PomodoroCompleted), string(PomodoroStopped), string(PomodoroTick):
		var pomodoroEvent PomodoroEvent
		if err := json.Unmarshal(wsEvent.Payload, &pomodoroEvent); err == nil {
			b.localBus.Publish(pomodoroEvent)
		}
	case string(TaskCreated), string(TaskUpdated), string(TaskDeleted), string(TaskCompleted):
		var taskEvent TaskEvent
		if err := json.Unmarshal(wsEvent.Payload, &taskEvent); err == nil {
			b.localBus.Publish(taskEvent)
		}
	}
}
