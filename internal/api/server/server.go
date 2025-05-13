// Package server provides the HTTP API server implementation
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"

	"github.com/hatappi/gomodoro/graph"
	"github.com/hatappi/gomodoro/graph/resolver"
	"github.com/hatappi/gomodoro/internal/api/server/handlers"
	servermiddleware "github.com/hatappi/gomodoro/internal/api/server/middleware"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// Server represents the API server.
type Server struct {
	config           *config.APIConfig
	router           *chi.Mux
	httpServer       *http.Server
	logger           logr.Logger
	pomodoroService  *core.PomodoroService
	taskService      *core.TaskService
	eventBus         event.EventBus
	webSocketHandler *EventWebSocketHandler
}

// NewServer creates a new API server instance.
func NewServer(
	config *config.APIConfig,
	logger logr.Logger,
	pomodoroService *core.PomodoroService,
	taskService *core.TaskService,
	eventBus event.EventBus,
) *Server {
	router := chi.NewRouter()

	server := &Server{
		config:           config,
		router:           router,
		logger:           logger,
		pomodoroService:  pomodoroService,
		taskService:      taskService,
		eventBus:         eventBus,
		webSocketHandler: NewEventWebSocketHandler(logger, eventBus),
	}

	server.webSocketHandler.SetupEventSubscription()
	server.setupMiddleware()
	server.setupRoutes()
	server.setupGraphQL(eventBus)

	return server
}

// setupMiddleware configures the middleware for the server.
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Use(servermiddleware.ErrorHandler())
}

// setupRoutes configures the routes for the server.
func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			s.logger.Error(err, "Failed to write health check response")
		}
	})

	s.router.Route("/api", func(r chi.Router) {
		r.Use(servermiddleware.JSONContentType)

		pomodoroHandler := handlers.NewPomodoroHandler(s.pomodoroService)
		r.Route("/pomodoro", func(r chi.Router) {
			r.Get("/", pomodoroHandler.GetCurrentPomodoro)
			r.Post("/start", pomodoroHandler.StartPomodoro)
			r.Post("/pause", pomodoroHandler.PausePomodoro)
			r.Post("/resume", pomodoroHandler.ResumePomodoro)
			r.Delete("/", pomodoroHandler.StopPomodoro)
		})

		taskHandler := handlers.NewTaskHandler(s.taskService)
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.GetTasks)
			r.Post("/", taskHandler.CreateTask)
			r.Get("/{id}", taskHandler.GetTask)
			r.Put("/{id}", taskHandler.UpdateTask)
			r.Delete("/{id}", taskHandler.DeleteTask)
		})

		r.HandleFunc("/events/ws", s.webSocketHandler.ServeHTTP)
	})
}

// setupGraphQL initializes the GraphQL handler and routes.
func (s *Server) setupGraphQL(eventBus event.EventBus) {
	resolver := &resolver.Resolver{
		EventBus: eventBus,
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }},
	})

	srv.Use(extension.Introspection{})

	s.router.Route("/graphql", func(r chi.Router) {
		r.Handle("/query", srv)
		r.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql/query"))
	})
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
	}

	return nil
}

// Listen starts listening on the configured address and returns a net.Listener.
func (s *Server) Listen() (net.Listener, error) {
	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("Listening API server", "addr", s.config.Addr)

	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	return ln, nil
}

// Serve starts serving HTTP requests using the provided listener.
func (s *Server) Serve(ln net.Listener) error {
	s.logger.Info("Serving API server")
	if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
