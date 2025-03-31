package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/storage"
)

// Task represents a task with its current state
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed"`
}

// TaskService provides operations for managing tasks
type TaskService struct {
	storage  storage.TaskStorage
	eventBus event.EventBus
}

// NewTaskService creates a new task service instance
func NewTaskService(storage storage.TaskStorage, eventBus event.EventBus) *TaskService {
	return &TaskService{
		storage:  storage,
		eventBus: eventBus,
	}
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(ctx context.Context, title string) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	task := &storage.Task{
		ID:        uuid.New().String(),
		Title:     title,
		CreatedAt: time.Now(),
		Completed: false,
	}

	if err := s.storage.SaveTask(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	s.publishTaskEvent(event.TaskCreated, task)

	return s.storageTaskToCore(task), nil
}

func (s *TaskService) GetAllTasks() ([]*Task, error) {
	tasks, err := s.storage.GetTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	result := make([]*Task, len(tasks))
	for i, task := range tasks {
		result[i] = s.storageTaskToCore(task)
	}

	return result, nil
}

func (s *TaskService) GetTaskByID(id string) (*Task, error) {
	task, err := s.storage.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return s.storageTaskToCore(task), nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, title string, completed bool) (*Task, error) {
	task, err := s.storage.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if title != "" {
		task.Title = title
	}

	if task.Completed != completed {
		task.Completed = completed

		if completed {
			s.publishTaskEvent(event.TaskCompleted, task)
		}
	}

	if err := s.storage.UpdateTask(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	s.publishTaskEvent(event.TaskUpdated, task)

	return s.storageTaskToCore(task), nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	task, err := s.storage.GetTaskByID(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if err := s.storage.DeleteTask(id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	s.publishTaskEvent(event.TaskDeleted, task)

	return nil
}

func (s *TaskService) publishTaskEvent(eventType event.EventType, t *storage.Task) {
	e := event.TaskEvent{
		BaseEvent: event.BaseEvent{
			Type:      eventType,
			Timestamp: time.Now(),
		},
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
	}

	s.eventBus.Publish(e)
}

func (s *TaskService) storageTaskToCore(t *storage.Task) *Task {
	if t == nil {
		return nil
	}

	return &Task{
		ID:        t.ID,
		Title:     t.Title,
		CreatedAt: t.CreatedAt,
		Completed: t.Completed,
	}
}
