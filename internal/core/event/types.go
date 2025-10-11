// Package event provides event types and event handling mechanisms
package event

import (
	"time"
)

// EventType represents the type of an event.
//
//nolint:revive
type EventType string

const (
	// PomodoroStarted event is emitted when a pomodoro session starts.
	PomodoroStarted EventType = "pomodoro.started"
	// PomodoroPaused event is emitted when a pomodoro session is paused.
	PomodoroPaused EventType = "pomodoro.paused"
	// PomodoroResumed event is emitted when a paused pomodoro session is resumed.
	PomodoroResumed EventType = "pomodoro.resumed"
	// PomodoroCompleted event is emitted when a pomodoro session is completed.
	PomodoroCompleted EventType = "pomodoro.completed"
	// PomodoroStopped event is emitted when a pomodoro session is manually stopped.
	PomodoroStopped EventType = "pomodoro.stopped"
	// PomodoroReset event is emitted when a pomodoro session is reset.
	PomodoroReset EventType = "pomodoro.reset"
	// PomodoroTick event is emitted on each second during an active pomodoro.
	PomodoroTick EventType = "pomodoro.tick"

	// TaskCreated event is emitted when a task is created.
	TaskCreated EventType = "task.created"
	// TaskUpdated event is emitted when a task is updated.
	TaskUpdated EventType = "task.updated"
	// TaskDeleted event is emitted when a task is deleted.
	TaskDeleted EventType = "task.deleted"
)

// AllEventTypes contains a list of all available event types in the system.
var AllEventTypes = []EventType{
	PomodoroStarted, PomodoroPaused, PomodoroResumed,
	PomodoroCompleted, PomodoroStopped, PomodoroReset, PomodoroTick,
	TaskCreated, TaskUpdated, TaskDeleted,
}

// BaseEvent contains common fields for all events.
type BaseEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// PomodoroState represents the state of a pomodoro session.
type PomodoroState string

const (
	// PomodoroStateActive indicates a running pomodoro.
	PomodoroStateActive PomodoroState = "active"
	// PomodoroStatePaused indicates a paused pomodoro.
	PomodoroStatePaused PomodoroState = "paused"
	// PomodoroStateFinished indicates a completed pomodoro.
	PomodoroStateFinished PomodoroState = "finished"
)

// PomodoroPhase represents the phase of a pomodoro session.
type PomodoroPhase string

const (
	// PomodoroPhaseWork represents the work/focus phase of a pomodoro session.
	PomodoroPhaseWork PomodoroPhase = "work"
	// PomodoroPhaseShortBreak represents the short break phase between pomodoro sessions.
	PomodoroPhaseShortBreak PomodoroPhase = "short_break"
	// PomodoroPhaseLongBreak represents the long break phase after completing multiple pomodoro sessions.
	PomodoroPhaseLongBreak PomodoroPhase = "long_break"
)

// PomodoroEvent represents events related to pomodoro sessions.
type PomodoroEvent struct {
	BaseEvent
	ID            string        `json:"id"`
	State         PomodoroState `json:"state"`
	RemainingTime time.Duration `json:"remaining_time"`
	ElapsedTime   time.Duration `json:"elapsed_time"`
	TaskID        string        `json:"task_id,omitempty"`
	Phase         PomodoroPhase `json:"phase"`
	PhaseCount    int           `json:"phase_count"`
	PhaseDuration time.Duration `json:"phase_duration"`
}

// GetEventType returns the event type.
func (e PomodoroEvent) GetEventType() EventType {
	return e.BaseEvent.Type
}

// TaskEvent represents events related to tasks.
type TaskEvent struct {
	BaseEvent
	ID    string `json:"id"`
	Title string `json:"title"`
}

// GetEventType returns the event type.
func (e TaskEvent) GetEventType() EventType {
	return e.BaseEvent.Type
}
