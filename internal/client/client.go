// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

func NewBaseClient(baseURL string, options ...Option) *BaseClient {
	config := &Config{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
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
type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error (%d): %s - %s", e.StatusCode, e.Code, e.Message)
}

func (c *BaseClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	return resp, nil
}

func (c *BaseClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return &ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       "unknown_error",
				Message:    string(body),
			}
		}

		if apiResp.Error != nil {
			return &ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       apiResp.Error.Code,
				Message:    apiResp.Error.Message,
			}
		}

		return &ErrorResponse{
			StatusCode: resp.StatusCode,
			Code:       "unknown_error",
			Message:    "Unknown error occurred",
		}
	}

	if result == nil {
		return nil
	}

	var apiResp APIResponse
	apiResp.Data = result

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return &ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       apiResp.Error.Code,
				Message:    apiResp.Error.Message,
			}
		}
		return &ErrorResponse{
			StatusCode: resp.StatusCode,
			Code:       "unknown_error",
			Message:    "API reported failure but no error details provided",
		}
	}

	return nil
}
