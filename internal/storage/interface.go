// Package storage provides interfaces and implementations for persisting gomodoro data
package storage

import (
	"time"
)

// PomodoroState represents the current state of a pomodoro timer.
type PomodoroState string

const (
	// PomodoroStateActive indicates a running pomodoro.
	PomodoroStateActive PomodoroState = "active"
	// PomodoroStatePaused indicates a paused pomodoro.
	PomodoroStatePaused PomodoroState = "paused"
	// PomodoroStateFinished indicates a completed pomodoro.
	PomodoroStateFinished PomodoroState = "finished"
)

// PomodoroPhase represents whether the current period is work or break.
type PomodoroPhase string

const (
	// PomodoroPhaseWork indicates a work period.
	PomodoroPhaseWork PomodoroPhase = "work"
	// PomodoroPhaseShortBreak indicates a short break period.
	PomodoroPhaseShortBreak PomodoroPhase = "short_break"
	// PomodoroPhaseLongBreak indicates a long break period.
	PomodoroPhaseLongBreak PomodoroPhase = "long_break"
)

// Pomodoro represents a pomodoro session that can be persisted.
type Pomodoro struct {
	ID                string        `json:"id"`
	State             PomodoroState `json:"state"`
	StartTime         time.Time     `json:"start_time"`
	WorkDuration      time.Duration `json:"work_duration"`
	BreakDuration     time.Duration `json:"break_duration"`
	LongBreakDuration time.Duration `json:"long_break_duration"`
	RemainingTime     time.Duration `json:"remaining_time"`
	ElapsedTime       time.Duration `json:"elapsed_time"`
	Phase             PomodoroPhase `json:"phase"`
	PhaseCount        int           `json:"phase_count"`
	TaskID            string        `json:"task_id,omitempty"`
}

// Task represents a task that can be persisted.
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed"`
}

// PomodoroStorage defines the interface for pomodoro persistence operations.
type PomodoroStorage interface {
	// SavePomodoro stores a pomodoro session
	SavePomodoro(pomodoro *Pomodoro) error

	// GetLatestPomodoro retrieves the most recent pomodoro session
	GetLatestPomodoro() (*Pomodoro, error)

	// GetActivePomodoro retrieves the current active pomodoro session if any
	GetActivePomodoro() (*Pomodoro, error)

	// UpdatePomodoroState updates the state and remaining time of a pomodoro
	UpdatePomodoroState(id string, state PomodoroState, remainSec int, elapsedSec int) (*Pomodoro, error)

	// DeletePomodoro deletes a pomodoro session by ID
	DeletePomodoro(id string) error
}

// TaskStorage defines the interface for task persistence operations.
type TaskStorage interface {
	// SaveTask stores a task
	SaveTask(task *Task) error

	// GetTasks retrieves all tasks
	GetTasks() ([]*Task, error)

	// GetTaskByID retrieves a task by its ID
	GetTaskByID(id string) (*Task, error)

	// UpdateTask updates an existing task
	UpdateTask(task *Task) error

	// DeleteTask removes a task by its ID
	DeleteTask(id string) error
}

// Storage is the combined interface for all storage operations.
type Storage interface {
	PomodoroStorage
	TaskStorage
}
