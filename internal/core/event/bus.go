// Package event provides event types and event handling mechanisms
package event

import (
	"fmt"
	"sync"
)

// Handler is a function that handles an event.
type Handler func(event interface{})

// EventBus is an interface for event publishing and subscription.
//
//nolint:revive
type EventBus interface {
	// Publish sends an event to all subscribers of that event type
	Publish(event EventInfo)

	// Subscribe registers a handler for a specific event type
	Subscribe(eventType string, handler Handler) string

	// Unsubscribe removes a handler for a specific event type using the subscription ID
	Unsubscribe(subscriptionID string)

	// SubscribeMulti registers a handler for multiple event types and returns a slice of subscription IDs
	SubscribeMulti(eventTypes []string, handler Handler) []string

	// SubscribeChannel registers handlers for multiple event types
	// Returns a channel to receive events and an unsubscribe function
	SubscribeChannel(eventTypes []string) (<-chan interface{}, func())
}

// WebSocketClient defines WebSocket client functionality needed by the event bus.
type WebSocketClient interface {
	Connect() error
	Send(event WebSocketEvent) error
	OnMessage(handler func(event WebSocketEvent))
	Close() error
}

// DefaultEventBus is the standard implementation of EventBus.
type DefaultEventBus struct {
	subscribers map[string]map[string]Handler
	mu          sync.RWMutex
	idCounter   int
}

// NewEventBus creates a new event bus instance.
func NewEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		subscribers: make(map[string]map[string]Handler),
	}
}

// Publish sends an event to all subscribers of that event type.
func (b *DefaultEventBus) Publish(event EventInfo) {
	eventType := string(event.GetEventType())

	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, exists := b.subscribers[eventType]; exists {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// Subscribe registers a handler for a specific event type and returns a subscription ID.
func (b *DefaultEventBus) Subscribe(eventType string, handler Handler) string {
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
func (b *DefaultEventBus) Unsubscribe(subscriptionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for eventType, handlers := range b.subscribers {
		if _, exists := handlers[subscriptionID]; exists {
			delete(b.subscribers[eventType], subscriptionID)
			return
		}
	}
}
