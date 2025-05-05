package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// PomodoroHandler handles pomodoro-related API endpoints
type PomodoroHandler struct {
	pomodoroService *core.PomodoroService
}

// NewPomodoroHandler creates a new pomodoro handler
func NewPomodoroHandler(pomodoroService *core.PomodoroService) *PomodoroHandler {
	return &PomodoroHandler{
		pomodoroService: pomodoroService,
	}
}

// PomodoroResponse represents the response structure for pomodoro endpoints
type PomodoroResponse struct {
	ID            string              `json:"id"`
	State         event.PomodoroState `json:"state"`
	TaskID        string              `json:"task_id,omitempty"`
	StartTime     time.Time           `json:"start_time"`
	Phase         event.PomodoroPhase `json:"phase"`
	PhaseCount    int                 `json:"phase_count"`
	RemainingTime int                 `json:"remaining_time_sec"`
	ElapsedTime   int                 `json:"elapsed_time_sec"`
}

type StartPomodoroRequest struct {
	WorkDuration      int    `json:"work_duration_sec"`
	BreakDuration     int    `json:"break_duration_sec"`
	LongBreakDuration int    `json:"long_break_duration_sec"`
	TaskID            string `json:"task_id,omitempty"`
}

func (h *PomodoroHandler) GetCurrentPomodoro(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	pomodoro, err := h.pomodoroService.GetLatestPomodoro()
	if err != nil {
		logger.Error(err, "Failed to get latest pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve latest pomodoro")
		return
	}

	if pomodoro == nil {
		RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	response := convertPomodoroToResponse(pomodoro)
	RespondWithJSON(w, http.StatusOK, response)
}

func (h *PomodoroHandler) StartPomodoro(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	var req StartPomodoroRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(err, "Failed to parse request body")
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	workDurationTime := time.Duration(req.WorkDuration) * time.Second
	breakDurationTime := time.Duration(req.BreakDuration) * time.Second
	longBreakDurationTime := time.Duration(req.LongBreakDuration) * time.Second

	pomodoro, err := h.pomodoroService.StartPomodoro(ctx, workDurationTime, breakDurationTime, longBreakDurationTime, req.TaskID)
	if err != nil {
		logger.Error(err, "Failed to start pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to start pomodoro")
		return
	}

	response := convertPomodoroToResponse(pomodoro)
	RespondWithJSON(w, http.StatusCreated, response)
}

func (h *PomodoroHandler) PausePomodoro(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	pomodoro, err := h.pomodoroService.GetActivePomodoro()
	if err != nil {
		logger.Error(err, "Failed to get active pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve active pomodoro")
		return
	}

	pomodoro, err = h.pomodoroService.PausePomodoro(ctx, pomodoro.ID)
	if err != nil {
		logger.Error(err, "Failed to pause pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to pause pomodoro")
		return
	}

	if pomodoro == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "No active pomodoro to pause")
		return
	}

	response := convertPomodoroToResponse(pomodoro)
	RespondWithJSON(w, http.StatusOK, response)
}

func (h *PomodoroHandler) ResumePomodoro(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	pomodoro, err := h.pomodoroService.GetActivePomodoro()
	if err != nil {
		logger.Error(err, "Failed to get active pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve active pomodoro")
		return
	}

	pomodoro, err = h.pomodoroService.ResumePomodoro(ctx, pomodoro.ID)
	if err != nil {
		logger.Error(err, "Failed to resume pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to resume pomodoro")
		return
	}

	if pomodoro == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "No paused pomodoro to resume")
		return
	}

	response := convertPomodoroToResponse(pomodoro)
	RespondWithJSON(w, http.StatusOK, response)
}

func (h *PomodoroHandler) StopPomodoro(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	pomodoro, err := h.pomodoroService.GetActivePomodoro()
	if err != nil {
		logger.Error(err, "Failed to get active pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve active pomodoro")
		return
	}

	if pomodoro == nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "No active pomodoro to stop")
		return
	}

	if err := h.pomodoroService.StopPomodoro(ctx, pomodoro.ID); err != nil {
		logger.Error(err, "Failed to stop pomodoro")
		RespondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to stop pomodoro")
		return
	}

	response := convertPomodoroToResponse(pomodoro)
	RespondWithJSON(w, http.StatusOK, response)
}

func convertPomodoroToResponse(p *core.Pomodoro) PomodoroResponse {
	remainingSeconds := int(p.RemainingTime / time.Second)

	return PomodoroResponse{
		ID:            p.ID,
		State:         p.State,
		TaskID:        p.TaskID,
		StartTime:     p.StartTime,
		Phase:         p.Phase,
		PhaseCount:    p.PhaseCount,
		RemainingTime: remainingSeconds,
		ElapsedTime:   int(p.ElapsedTime / time.Second),
	}
}
