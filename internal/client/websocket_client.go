package client

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"

	ev "github.com/hatappi/gomodoro/internal/core/event"
)

// WebSocketClientImpl is the actual WebSocket client implementation.
type WebSocketClientImpl struct {
	url             string
	conn            *websocket.Conn
	messageHandlers []func(ev.WebSocketEvent)
	mu              sync.RWMutex
	done            chan struct{}
}

// NewWebSocketClient creates a new WebSocket client.
func NewWebSocketClient(url string) *WebSocketClientImpl {
	return &WebSocketClientImpl{
		url:             url,
		messageHandlers: []func(ev.WebSocketEvent){},
		done:            make(chan struct{}),
	}
}

// Connect connects to the WebSocket server.
func (c *WebSocketClientImpl) Connect() error {
	conn, resp, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// Start message receive loop
	go c.readPump()

	return nil
}

// readPump continuously reads messages from the WebSocket connection.
func (c *WebSocketClientImpl) readPump() {
	defer func() {
		_ = c.conn.Close()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			var wsEvent ev.WebSocketEvent
			err := c.conn.ReadJSON(&wsEvent)
			if err != nil {
				// Exit when connection is closed
				return
			}

			// Invoke message handlers
			c.mu.RLock()
			for _, handler := range c.messageHandlers {
				handler(wsEvent)
			}
			c.mu.RUnlock()
		}
	}
}

// Send sends an event to the WebSocket server.
func (c *WebSocketClientImpl) Send(event ev.WebSocketEvent) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return fmt.Errorf("not connected to WebSocket server")
	}

	return c.conn.WriteJSON(event)
}

// OnMessage registers a handler for received messages.
func (c *WebSocketClientImpl) OnMessage(handler func(event ev.WebSocketEvent)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messageHandlers = append(c.messageHandlers, handler)
}

// Close closes the WebSocket connection.
func (c *WebSocketClientImpl) Close() error {
	close(c.done)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
