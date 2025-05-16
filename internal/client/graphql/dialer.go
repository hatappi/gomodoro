// Package graphql provides a GraphQL client implementation for interacting with the Gomodoro GraphQL API
package graphql

import (
	"context"
	"net/http"

	gqllib "github.com/Khan/genqlient/graphql"
	"github.com/gorilla/websocket"
)

// GorillaWebSocketDialer implements gqllib.WebSocketDialer using gorilla/websocket.
type GorillaWebSocketDialer struct {
	dialer *websocket.Dialer
}

// NewGorillaWebSocketDialer creates a new GorillaWebSocketDialer.
func NewGorillaWebSocketDialer(dialer *websocket.Dialer) *GorillaWebSocketDialer {
	return &GorillaWebSocketDialer{dialer: dialer}
}

// DialContext satisfies the gqllib.WebSocketDialer interface.
func (d *GorillaWebSocketDialer) DialContext(
	ctx context.Context,
	urlStr string,
	requestHeader http.Header,
) (gqllib.WSConn, error) {
	//nolint:bodyclose
	conn, _, err := d.dialer.DialContext(ctx, urlStr, requestHeader)

	return conn, err
}
