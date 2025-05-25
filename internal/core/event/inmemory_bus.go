// Package event provides event types and event handling mechanisms
package event

import (
	"fmt"
	"sync"
)

const (
	defaultChannelBufferSize = 10
)

// InMemoryBus implements EventBus interface with in-memory event distribution.
type InMemoryBus struct {
	subscribers map[EventType]map[string]Handler
	mu          sync.RWMutex
	idCounter   int
}

// NewInMemoryBus creates and initializes a new InMemoryBus instance.
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		subscribers: make(map[EventType]map[string]Handler),
	}
}

// Publish sends an event to all subscribers of that event type.
func (b *InMemoryBus) Publish(event EventInfo) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	eventType := event.GetEventType()

	if handlers, exists := b.subscribers[eventType]; exists {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// SubscribeMulti registers a handler for multiple event types and returns subscription IDs.
func (b *InMemoryBus) SubscribeMulti(eventTypes []EventType, handler Handler) []string {
	subscriptionIDs := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		id := b.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}

	return subscriptionIDs
}

// SubscribeChannel returns the channel and an unsubscribe function.
func (b *InMemoryBus) SubscribeChannel(eventTypes []EventType) (<-chan interface{}, func()) {
	ch := make(chan interface{}, defaultChannelBufferSize)
	handler := func(e interface{}) {
		ch <- e
	}

	subscriptionIDs := b.SubscribeMulti(eventTypes, handler)

	unsubscribe := func() {
		for _, id := range subscriptionIDs {
			b.Unsubscribe(id)
		}
		close(ch)
	}

	return ch, unsubscribe
}

// Subscribe registers a handler for a specific event type and returns a subscription ID.
func (b *InMemoryBus) Subscribe(eventType EventType, handler Handler) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[eventType]; !exists {
		b.subscribers[eventType] = make(map[string]Handler)
	}

	b.idCounter++
	id := fmt.Sprintf("%s-%d", eventType, b.idCounter)
	b.subscribers[eventType][id] = handler

	return id
}

// Unsubscribe removes a handler for a specific event type using the subscription ID.
func (b *InMemoryBus) Unsubscribe(subscriptionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for eventType, handlers := range b.subscribers {
		if _, exists := handlers[subscriptionID]; exists {
			delete(b.subscribers[eventType], subscriptionID)
			return
		}
	}
}
