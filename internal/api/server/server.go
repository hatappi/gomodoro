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
	"github.com/gorilla/websocket"

	"github.com/hatappi/go-kit/log"

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
	pomodoroService *core.PomodoroService
	taskService     *core.TaskService
	eventBus        event.EventBus

	completeFuncs []func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime time.Duration) error
}

// NewServer creates a new API server instance.
func NewServer(
	config *config.APIConfig,
	pomodoroService *core.PomodoroService,
	taskService *core.TaskService,
	eventBus event.EventBus,
	opts ...Option,
) *Server {
	router := chi.NewRouter()

	server := &Server{
		config:          config,
		router:          router,
		pomodoroService: pomodoroService,
		taskService:     taskService,
		eventBus:        eventBus,
	}

	for _, opt := range opts {
		opt(server)
	}

	server.setupMiddleware()
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

// Listen starts listening on the configured address and returns a net.Listener.
func (s *Server) Listen() (net.Listener, error) {
	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	return ln, nil
}

// Start the HTTP server and blocks until it is stopped.
func (s *Server) Start(ctx context.Context, ln net.Listener) error {
	go s.handlePomodoroCompletionEvents(ctx)

	log.FromContext(ctx).V(1).Info("Serving API server", "addr", s.config.Addr)
	if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	log.FromContext(ctx).V(1).Info("Stopping API server...")

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
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
				log.FromContext(ctx).Error(err, "Failed to get task by ID", "taskID", pomodoroEvent.TaskID)
				continue
			}

			isWorkTime := pomodoroEvent.Phase == event.PomodoroPhaseWork

			for _, completeFunc := range s.completeFuncs {
				if err := completeFunc(ctx, task.Title, isWorkTime, pomodoroEvent.ElapsedTime); err != nil {
					log.FromContext(ctx).Error(err, "Failed to execute complete function", "taskID", pomodoroEvent.TaskID)
				}
			}
		}
	}
}
