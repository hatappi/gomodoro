// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TaskClient provides methods for interacting with task-related API endpoints
type TaskClient struct {
	*BaseClient
}

// NewTaskClient creates a new task client
func NewTaskClient(baseURL string, options ...Option) *TaskClient {
	return &TaskClient{
		BaseClient: NewBaseClient(baseURL, options...),
	}
}

// TaskResponse represents the response structure for task endpoints
type TaskResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed"`
}

// TaskRequest represents the request structure for creating/updating a task
type TaskRequest struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed,omitempty"`
}

// GetAll retrieves all tasks
func (c *TaskClient) GetAll(ctx context.Context) ([]TaskResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result []TaskResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Get retrieves a specific task by ID
func (c *TaskClient) Get(ctx context.Context, id string) (*TaskResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/tasks/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result TaskResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create creates a new task
func (c *TaskClient) Create(ctx context.Context, title string) (*TaskResponse, error) {
	req := TaskRequest{
		Title: title,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/tasks", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result TaskResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing task
func (c *TaskClient) Update(ctx context.Context, id, title string, completed bool) (*TaskResponse, error) {
	req := TaskRequest{
		Title:     title,
		Completed: completed,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/tasks/%s", id), req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result TaskResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete removes a task by ID
func (c *TaskClient) Delete(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/tasks/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
	}

	if err := c.parseResponse(resp, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to delete task")
	}

	return nil
}
