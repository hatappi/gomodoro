package graphql // client.go や generated.go と同じパッケージ

import (
	"context"
	"net/http"
	"time"

	gqllib "github.com/Khan/genqlient/graphql"
	"github.com/gorilla/websocket"
)

// GorillaWebSocketDialer implements gqllib.WebSocketDialer using gorilla/websocket.
type GorillaWebSocketDialer struct {
	dialer *websocket.Dialer
}

// NewGorillaWebSocketDialer creates a new GorillaWebSocketDialer.
func NewGorillaWebSocketDialer(dialer *websocket.Dialer) *GorillaWebSocketDialer {
	if dialer == nil {
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		}
	}
	return &GorillaWebSocketDialer{dialer: dialer}
}

// DialContext satisfies the gqllib.WebSocketDialer interface.
func (d *GorillaWebSocketDialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (gqllib.WSConn, error) {
	conn, _, err := d.dialer.DialContext(ctx, urlStr, requestHeader)

	return conn, err
}
