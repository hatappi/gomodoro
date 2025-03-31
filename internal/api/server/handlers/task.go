package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/core"
)

// TaskHandler handles task-related API endpoints
type TaskHandler struct {
	taskService *core.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *core.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
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

// GetTasks handles GET /api/tasks
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		logger.Error(err, "Failed to get tasks")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve tasks")
		return
	}

	taskResponses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponses = append(taskResponses, convertTaskToResponse(task))
	}

	RespondWithJSON(w, http.StatusOK, taskResponses)
}

// GetTask handles GET /api/tasks/:id
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)
	taskID := chi.URLParam(r, "id")

	task, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		logger.Error(err, "Failed to get task", "taskID", taskID)
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve task")
		return
	}

	if task == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, convertTaskToResponse(task))
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(err, "Failed to parse request body")
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Title is required")
		return
	}

	task, err := h.taskService.CreateTask(ctx, req.Title)
	if err != nil {
		logger.Error(err, "Failed to create task")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to create task")
		return
	}

	RespondWithJSON(w, http.StatusCreated, convertTaskToResponse(task))
}

// UpdateTask handles PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)
	taskID := chi.URLParam(r, "id")

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(err, "Failed to parse request body")
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Title is required")
		return
	}

	existingTask, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		logger.Error(err, "Failed to get task", "taskID", taskID)
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve task")
		return
	}

	if existingTask == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	updatedTask, err := h.taskService.UpdateTask(ctx, taskID, req.Title, req.Completed)
	if err != nil {
		logger.Error(err, "Failed to update task", "taskID", taskID)
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to update task")
		return
	}

	RespondWithJSON(w, http.StatusOK, convertTaskToResponse(updatedTask))
}

// DeleteTask handles DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)
	taskID := chi.URLParam(r, "id")

	existingTask, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		logger.Error(err, "Failed to get task", "taskID", taskID)
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve task")
		return
	}

	if existingTask == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	err = h.taskService.DeleteTask(ctx, taskID)
	if err != nil {
		logger.Error(err, "Failed to delete task", "taskID", taskID)
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to delete task")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Helper function to convert core.Task to TaskResponse
func convertTaskToResponse(t *core.Task) TaskResponse {
	return TaskResponse{
		ID:        t.ID,
		Title:     t.Title,
		CreatedAt: t.CreatedAt,
		Completed: t.Completed,
	}
}
