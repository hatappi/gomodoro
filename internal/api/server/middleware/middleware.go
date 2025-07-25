// Package middleware provides HTTP middleware components for the API server
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hatappi/go-kit/log"
)

// ErrorHandler is middleware that catches panics and returns appropriate error responses.
func ErrorHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.FromContext(r.Context())

			defer func() {
				if err := recover(); err != nil {
					logger.Error(fmt.Errorf("panic recovered"), "panic recovered", "error", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					}); err != nil {
						logger.Error(err, "failed to encode error response")
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
