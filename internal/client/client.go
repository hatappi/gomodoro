// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"fmt"
	"net/http"
	"time"
)

const (
	// defaultClientTimeout is the default timeout for HTTP client requests.
	defaultClientTimeout = 10 * time.Second
)

// Config represents configuration for API clients.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Option is a functional option for configuring the client.
type Option func(*Config)

// WithTimeout sets the timeout for API requests.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = httpClient
	}
}

// BaseClient provides common functionality for API clients.
type BaseClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBaseClient creates a new BaseClient with the given base URL and options.
func NewBaseClient(baseURL string, options ...Option) *BaseClient {
	config := &Config{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: defaultClientTimeout,
		},
	}

	for _, option := range options {
		option(config)
	}

	return &BaseClient{
		baseURL:    config.BaseURL,
		httpClient: config.HTTPClient,
	}
}

// APIResponse represents the standard response structure from the API.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// ErrorData represents error details in an API response.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse represents an API error response.
//
//nolint:errname
type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error (%d): %s - %s", e.StatusCode, e.Code, e.Message)
}
