// Package task manage task
package task

import (
	"context"
	"errors"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/domain/model"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui"
	"github.com/hatappi/gomodoro/internal/tui/screen"
)

// For backward compatibility
type Task = model.Task
type Tasks = model.Tasks

// Client task client.
type Client interface {
	GetTask(ctx context.Context) (*Task, error)
	LoadTasks(ctx context.Context) (Tasks, error)
	SaveTask(ctx context.Context, task *Task) error
}

// NewClient initializes Client.
func NewClient(config *config.Config, c screen.Client) *IClient {
	// Create API client
	apiClientFactory := client.NewFactory(config.API)

	taskClient := apiClientFactory.Task()

	return &IClient{
		config:       config,
		screenClient: c,
		apiClient:    taskClient,
	}
}

// IClient meets Client interface.
type IClient struct {
	config       *config.Config
	screenClient screen.Client
	apiClient    *client.TaskClient
}

// GetTask get selected Task.
func (c *IClient) GetTask(ctx context.Context) (*Task, error) {
	tasks, err := c.LoadTasks(ctx)
	if err != nil {
		return nil, err
	}

	taskView := tui.NewTaskView(c.config, c.screenClient)

	var t *Task

	if len(tasks) > 0 {
		var action tui.TaskAction
		t, action, err = taskView.SelectTaskName(ctx, tasks)
		if err != nil {
			return nil, err
		}

		switch action {
		case tui.TaskActionCancel:
			return nil, gomodoro_error.ErrCancel
		case tui.TaskActionDelete:
			if t != nil {
				// Delete task
				if err := c.apiClient.Delete(ctx, t.ID); err != nil {
					return nil, err
				}

				// Reload tasks after deletion
				tasks, err = c.LoadTasks(ctx)
				if err != nil {
					return nil, err
				}

				// If no tasks left, create a new one
				if len(tasks) == 0 {
					t = nil
				} else {
					// Reselect after deletion
					t, _, err = taskView.SelectTaskName(ctx, tasks)
					if err != nil {
						return nil, err
					}
				}
			}
		case tui.TaskActionNew:
			t = nil
		}
	}

	if t == nil {
		// Create new task
		name, err := taskView.CreateTaskName(ctx)
		if errors.Is(err, gomodoro_error.ErrCancel) {
			return nil, err
		}

		resp, err := c.apiClient.Create(ctx, name)
		if err != nil {
			return nil, err
		}

		t = &Task{
			ID:        resp.ID,
			Name:      resp.Title,
			Completed: resp.Completed,
		}

		err = c.SaveTask(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// LoadTasks loads tasks from the API
func (c *IClient) LoadTasks(ctx context.Context) (Tasks, error) {
	responses, err := c.apiClient.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make(Tasks, len(responses))
	for i, resp := range responses {
		tasks[i] = &Task{
			ID:        resp.ID,
			Name:      resp.Title,
			Completed: resp.Completed,
		}
	}
	return tasks, nil
}

// SaveTask saves a task using API client
func (c *IClient) SaveTask(ctx context.Context, task *Task) error {
	var resp *client.TaskResponse
	var err error

	if task.ID == "" {
		// Create new task
		resp, err = c.apiClient.Create(ctx, task.Name)
	} else {
		// Update existing task
		resp, err = c.apiClient.Update(ctx, task.ID, task.Name, task.Completed)
	}

	if err != nil {
		return err
	}

	// On success, set returned ID on task
	task.ID = resp.ID

	return nil
}
