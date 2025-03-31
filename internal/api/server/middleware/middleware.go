// Package middleware provides HTTP middleware components for the API server
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hatappi/go-kit/log"
)

// RequestLogger is a middleware that logs request details
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.FromContext(r.Context())

			start := time.Now()

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)
			logger.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", lw.statusCode,
				"duration", duration,
				"ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// ErrorHandler is middleware that catches panics and returns appropriate error responses
func ErrorHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.FromContext(r.Context())

			defer func() {
				if err := recover(); err != nil {
					logger.Error(fmt.Errorf("Panic recovered"), "Panic recovered", "error", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ContentType is middleware for setting content type
func ContentType(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// JSONContentType sets the content type to application/json
func JSONContentType(next http.Handler) http.Handler {
	return ContentType("application/json")(next)
}

type contextKey string

const (
	// UserContextKey is the key used to store user information in request context
	UserContextKey contextKey = "user"
)

// loggingResponseWriter is a custom response writer that captures the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing the header
func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
