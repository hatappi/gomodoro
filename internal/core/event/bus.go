// Package event provides event types and event handling mechanisms
package event

// EventInfo defines the interface for event information that can be published.
//
//nolint:revive
type EventInfo interface {
	GetEventType() EventType
}

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
