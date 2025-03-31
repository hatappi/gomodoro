package event

import (
	"encoding/json"
)

// WebSocketEvent is the event structure sent/received over WebSocket
type WebSocketEvent struct {
	EventType string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type WebSocketEventBus struct {
	localBus *DefaultEventBus
	wsClient WebSocketClient
	isServer bool
}

type WebSocketClientConnection interface {
	Send(event WebSocketEvent) error
}

func newWebSocketEventBus(wsClient WebSocketClient, isServer bool) EventBus {
	bus := &WebSocketEventBus{
		localBus: NewEventBus(),
		wsClient: wsClient,
		isServer: isServer,
	}

	if wsClient != nil {
		wsClient.OnMessage(func(event WebSocketEvent) {
			bus.handleRemoteEvent(event)
		})
	}

	return bus
}

func NewServerWebSocketEventBus() EventBus {
	return newWebSocketEventBus(nil, true)
}

func NewClientWebSocketEventBus(wsClient WebSocketClient) EventBus {
	return newWebSocketEventBus(wsClient, false)
}

type EventInfo interface {
	GetEventType() EventType
}

func (b *WebSocketEventBus) Publish(event EventInfo) {
	b.localBus.Publish(event)

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	wsEvent := WebSocketEvent{
		EventType: string(event.GetEventType()),
		Payload:   json.RawMessage(payload),
	}

	if !b.isServer && b.wsClient != nil {
		b.wsClient.Send(wsEvent)
	}
}

func (b *WebSocketEventBus) Subscribe(eventType string, handler Handler) string {
	return b.localBus.Subscribe(eventType, handler)
}

func (b *WebSocketEventBus) SubscribeMulti(eventTypes []string, handler Handler) []string {
	subscriptionIDs := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		id := b.localBus.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}
	return subscriptionIDs
}

func (b *WebSocketEventBus) SubscribeChannel(eventTypes []string) (<-chan interface{}, func()) {
	ch := make(chan interface{}, 10)
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

func (b *WebSocketEventBus) Unsubscribe(subscriptionID string) {
	b.localBus.Unsubscribe(subscriptionID)
}

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
