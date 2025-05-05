// Package client provides API clients for interacting with the Gomodoro API server
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hatappi/gomodoro/internal/core/event"
)

// PomodoroClient provides methods for interacting with pomodoro-related API endpoints
type PomodoroClient struct {
	*BaseClient
}

// NewPomodoroClient creates a new pomodoro client
func NewPomodoroClient(baseURL string, options ...Option) *PomodoroClient {
	return &PomodoroClient{
		BaseClient: NewBaseClient(baseURL, options...),
	}
}

// PomodoroResponse represents the response from pomodoro-related API endpoints
type PomodoroResponse struct {
	ID            string              `json:"id"`
	State         event.PomodoroState `json:"state"` // Use event.PomodoroState for type safety
	TaskID        string              `json:"task_id,omitempty"`
	StartTime     time.Time           `json:"start_time"`
	RemainingTime int                 `json:"remaining_time_sec"` // in seconds
	ElapsedTime   int                 `json:"elapsed_time_sec"`   // in seconds
	Phase         event.PomodoroPhase `json:"phase"`              // Use event.PomodoroPhase for type safety
	PhaseCount    int                 `json:"phase_count"`
}

// StartPomodoroRequest represents the request for starting a new pomodoro
type StartPomodoroRequest struct {
	WorkDuration      int    `json:"work_duration_sec"`       // in seconds
	BreakDuration     int    `json:"break_duration_sec"`      // in seconds
	LongBreakDuration int    `json:"long_break_duration_sec"` // in seconds
	TaskID            string `json:"task_id,omitempty"`
}

// GetCurrent retrieves the current active pomodoro session
func (c *PomodoroClient) GetCurrent(ctx context.Context) (*PomodoroResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/pomodoro", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Handle 404 or empty response case
	if resp.StatusCode == http.StatusNotFound || resp.ContentLength == 0 {
		return nil, nil
	}

	var result PomodoroResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Start begins a new pomodoro session
func (c *PomodoroClient) Start(ctx context.Context, workDuration, breakDuration time.Duration, longBreakDuration time.Duration, taskID string) (*PomodoroResponse, error) {
	req := StartPomodoroRequest{
		WorkDuration:      int(workDuration.Seconds()),
		BreakDuration:     int(breakDuration.Seconds()),
		LongBreakDuration: int(longBreakDuration.Seconds()),
		TaskID:            taskID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/pomodoro/start", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result PomodoroResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Pause pauses the current active pomodoro session
func (c *PomodoroClient) Pause(ctx context.Context) (*PomodoroResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/pomodoro/pause", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result PomodoroResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Resume resumes a paused pomodoro session
func (c *PomodoroClient) Resume(ctx context.Context) (*PomodoroResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/pomodoro/resume", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result PomodoroResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Stop ends the current pomodoro session
func (c *PomodoroClient) Stop(ctx context.Context) (*PomodoroResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/api/pomodoro", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var result PomodoroResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
