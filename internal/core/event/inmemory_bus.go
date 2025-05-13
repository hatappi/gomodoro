package event

import (
	"fmt"
	"sync"
)

const (
	defaultChannelBufferSize = 10
)

type InMemoryBus struct {
	subscribers map[string]map[string]Handler
	mu          sync.RWMutex
	idCounter   int
}

func NewInMemoryBus() EventBus {
	return &InMemoryBus{
		subscribers: make(map[string]map[string]Handler),
	}
}

// Publish sends an event to all subscribers of that event type.
func (b *InMemoryBus) Publish(event EventInfo) {
	eventType := string(event.GetEventType())

	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, exists := b.subscribers[eventType]; exists {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

func (b *InMemoryBus) SubscribeMulti(eventTypes []string, handler Handler) []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscriptionIDs := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		id := b.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}

	return subscriptionIDs
}

func (b *InMemoryBus) SubscribeChannel(eventTypes []string) (<-chan interface{}, func()) {
	ch := make(chan interface{}, defaultChannelBufferSize)
	subscriptionIDs := make([]string, 0, len(eventTypes))
	handler := func(e interface{}) {
		ch <- e
	}
	for _, eventType := range eventTypes {
		id := b.Subscribe(eventType, handler)
		subscriptionIDs = append(subscriptionIDs, id)
	}
	unsubscribe := func() {
		for _, id := range subscriptionIDs {
			b.Unsubscribe(id)
		}
		close(ch)
	}
	return ch, unsubscribe
}

// Subscribe registers a handler for a specific event type and returns a subscription ID.
func (b *InMemoryBus) Subscribe(eventType string, handler Handler) string {
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
