// Package server provides WebSocket handling
package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"

	"github.com/hatappi/gomodoro/internal/core/event"
)

// EventWebSocketHandler manages WebSocket connections and broadcasts events to clients.
type EventWebSocketHandler struct {
	upgrader websocket.Upgrader
	clients  map[string]*websocket.Conn
	mu       sync.RWMutex
	logger   logr.Logger
	eventBus event.EventBus
}

// NewEventWebSocketHandler creates a new EventWebSocketHandler with the given logger and event bus.
func NewEventWebSocketHandler(logger logr.Logger, eventBus event.EventBus) *EventWebSocketHandler {
	return &EventWebSocketHandler{
		upgrader: websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }},
		clients:  make(map[string]*websocket.Conn),
		logger:   logger,
		eventBus: eventBus,
	}
}

// ServeHTTP upgrades HTTP connections to WebSocket and manages client lifecycle.
func (h *EventWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error(err, "Failed to upgrade WebSocket connection")
		return
	}

	clientID := r.RemoteAddr
	h.mu.Lock()
	h.clients[clientID] = conn
	h.mu.Unlock()
	h.logger.Info("WebSocket client connected", "clientID", clientID)

	defer func() {
		h.mu.Lock()
		delete(h.clients, clientID)
		h.mu.Unlock()
		if err := conn.Close(); err != nil {
			h.logger.Error(err, "Error closing WebSocket connection", "clientID", clientID)
		}
		h.logger.Info("WebSocket client disconnected", "clientID", clientID)
	}()

	for {
		_, _, err := conn.NextReader()
		if err != nil {
			h.logger.Error(err, "Failed to read from WebSocket connection", "clientID", clientID)
			break
		}

		// NOTE: handle incoming client messages if needed
	}
}

// Broadcast sends the given WebSocketEvent to all connected clients, excluding any specified client IDs.
func (h *EventWebSocketHandler) Broadcast(event event.WebSocketEvent, excludeClientIDs ...string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, conn := range h.clients {
		skip := false
		for _, exclude := range excludeClientIDs {
			if id == exclude {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if err := conn.WriteJSON(event); err != nil {
			h.logger.Error(err, "Failed to write WebSocket message", "clientID", id)
		}
	}
}

// SetupEventSubscription subscribes to events on the event bus and broadcasts them to WebSocket clients.
func (h *EventWebSocketHandler) SetupEventSubscription() {
	eventTypes := make([]string, 0, len(event.AllEventTypes))
	for _, eventType := range event.AllEventTypes {
		eventTypes = append(eventTypes, string(eventType))
	}

	h.eventBus.SubscribeMulti(eventTypes, func(e interface{}) {
		evtInfo, ok := e.(event.EventInfo)
		if !ok {
			return
		}

		payload, err := json.Marshal(evtInfo)
		if err != nil {
			h.logger.Error(err, "Failed to marshal event for WebSocket")
			return
		}

		wsEvent := event.WebSocketEvent{
			EventType: string(evtInfo.GetEventType()),
			Payload:   payload,
		}
		h.Broadcast(wsEvent)
	})
}
