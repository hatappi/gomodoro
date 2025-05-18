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

const (
	// defaultHandshakeTimeout is the default timeout for WebSocket handshaking.
	defaultHandshakeTimeout = 45 * time.Second
)

// Factory provides a convenient way to create and manage API clients.
type Factory struct {
	gqlClientWrapper *graphql.ClientWrapper
}

// NewFactory creates a new client factory with the given API configuration and options.
func NewFactory(apiConfig config.APIConfig) *Factory {
	queryClient := gqllib.NewClient(fmt.Sprintf("http://%s/graphql/query", apiConfig.Addr), http.DefaultClient)

	underlyingGorillaDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: defaultHandshakeTimeout,
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
		gqlClientWrapper: gqlClientWrapper,
	}

	return factory
}

// GraphQLClient returns access to the GraphQL client wrapper.
func (f *Factory) GraphQLClient() *graphql.ClientWrapper {
	return f.gqlClientWrapper
}

// Close closes all open connections.
func (f *Factory) Close() error {
	if f.gqlClientWrapper != nil {
		if err := f.gqlClientWrapper.DisconnectSubscription(); err != nil {
			return fmt.Errorf("failed to disconnect GraphQL subscription: %w", err)
		}
	}
	return nil
}
