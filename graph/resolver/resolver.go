// Package resolver is the resolver package for the GraphQL schema.
package resolver

import (
	"github.com/hatappi/gomodoro/internal/core/event"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver serves as the root resolver for the GraphQL schema.
type Resolver struct {
	EventBus event.EventBus
}
