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

	servermiddleware "github.com/hatappi/gomodoro/internal/api/server/middleware"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/graph"
	"github.com/hatappi/gomodoro/internal/graph/resolver"
)

// Server represents the API server.
type Server struct {
	config          *config.APIConfig
	router          *chi.Mux
	httpServer      *http.Server
	logger          logr.Logger
	pomodoroService *core.PomodoroService
	taskService     *core.TaskService
	eventBus        event.EventBus

	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime time.Duration) error
}

// NewServer creates a new API server instance.
func NewServer(
	config *config.APIConfig,
	logger logr.Logger,
	pomodoroService *core.PomodoroService,
	taskService *core.TaskService,
	eventBus event.EventBus,
	opts ...Option,
) *Server {
	router := chi.NewRouter()

	server := &Server{
		config:          config,
		router:          router,
		logger:          logger,
		pomodoroService: pomodoroService,
		taskService:     taskService,
		eventBus:        eventBus,
	}

	for _, opt := range opts {
		opt(server)
	}

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
	})
}

// setupGraphQL initializes the GraphQL handler and routes.
func (s *Server) setupGraphQL(eventBus event.EventBus) {
	resolver := &resolver.Resolver{
		EventBus:        eventBus,
		TaskService:     s.taskService,
		PomodoroService: s.pomodoroService,
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

// Start the HTTP server and blocks until it is stopped.
func (s *Server) Start(ctx context.Context, ln net.Listener) error {
	go s.handlePomodoroCompletionEvents(ctx)

	s.logger.Info("Serving API server")
	if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) handlePomodoroCompletionEvents(ctx context.Context) {
	busCh, unsubscribe := s.eventBus.SubscribeChannel([]event.EventType{event.PomodoroStopped, event.PomodoroCompleted})
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-busCh:
			if !ok {
				return
			}

			pomodoroEvent, ok := e.(event.PomodoroEvent)
			if !ok {
				continue
			}

			task, err := s.taskService.GetTaskByID(pomodoroEvent.TaskID)
			if err != nil {
				s.logger.Error(err, "Failed to get task by ID", "taskID", pomodoroEvent.TaskID)
				continue
			}

			isWorkTime := pomodoroEvent.Phase == event.PomodoroPhaseWork

			for _, completeFunc := range s.completeFuncs {
				if err := completeFunc(ctx, task.Title, isWorkTime, pomodoroEvent.ElapsedTime); err != nil {
					s.logger.Error(err, "Failed to execute complete function")
				}
			}
		}
	}
}
