// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"fmt"
	"net/http"
	"time"

	gqllib "github.com/Khan/genqlient/graphql"
	"github.com/gorilla/websocket"
	"github.com/hatappi/gomodoro/internal/client/graphql"
	"github.com/hatappi/gomodoro/internal/config"
)

// Factory provides a convenient way to create and manage API clients.
type Factory struct {
	pomodoro         *PomodoroClient
	task             *TaskClient
	wsClient         *WebSocketClientImpl
	gqlClientWrapper *graphql.ClientWrapper
}

// NewFactory creates a new client factory with the given API configuration and options.
func NewFactory(apiConfig config.APIConfig) *Factory {
	queryClient := gqllib.NewClient(fmt.Sprintf("http://%s/graphql/query", apiConfig.Addr), http.DefaultClient)

	underlyingGorillaDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}
	wsDialerAdapter := graphql.NewGorillaWebSocketDialer(underlyingGorillaDialer)

	// Create the subscription client using genqlient's library function
	subscriptionClientImpl := gqllib.NewClientUsingWebSocket(
		fmt.Sprintf("ws://%s/graphql/query", apiConfig.Addr),
		wsDialerAdapter,
	)

	// Pass subscriptionClientImpl directly
	gqlClientWrapper := graphql.NewClientWrapper(queryClient, subscriptionClientImpl)

	factory := &Factory{
		wsClient:         NewWebSocketClient(fmt.Sprintf("ws://%s/api/events/ws", apiConfig.Addr)),
		pomodoro:         NewPomodoroClient(fmt.Sprintf("http://%s", apiConfig.Addr)),
		task:             NewTaskClient(fmt.Sprintf("http://%s", apiConfig.Addr)),
		gqlClientWrapper: gqlClientWrapper,
	}

	return factory
}

// Pomodoro returns a client for interacting with Pomodoro API endpoints.
func (f *Factory) Pomodoro() *PomodoroClient {
	return f.pomodoro
}

// Task returns a client for interacting with Task API endpoints.
func (f *Factory) Task() *TaskClient {
	return f.task
}

// WebSocket returns a client for interacting with WebSocket endpoint.
func (f *Factory) WebSocket() (*WebSocketClientImpl, error) {
	if err := f.wsClient.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket: %w", err)
	}
	return f.wsClient, nil
}

// GraphQLClientWrapper provides access to the GraphQL client wrapper.
func (f *Factory) GraphQLClient() *graphql.ClientWrapper {
	return f.gqlClientWrapper
}

// Close closes all open connections.
func (f *Factory) Close() error {
	if f.wsClient != nil {
		if err := f.wsClient.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket client: %w", err)
		}
	}

	if f.gqlClientWrapper != nil {
		if err := f.gqlClientWrapper.DisconnectSubscription(); err != nil {
			return fmt.Errorf("failed to disconnect GraphQL subscription: %w", err)
		}
	}
	return nil
}
