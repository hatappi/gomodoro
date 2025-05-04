// Package constants defines constants used across the TUI
package constants

// TimerAction represents timer-specific actions
type TimerAction string

const (
	// TimerActionNone indicates no action
	TimerActionNone TimerAction = ""
	// TimerActionCancel indicates the timer action was canceled
	TimerActionCancel TimerAction = "timer:cancel"
	// TimerActionToggle indicates the timer should toggle between running/paused
	TimerActionToggle TimerAction = "timer:toggle"
	// TimerActionStop indicates the timer should stop
	TimerActionStop TimerAction = "timer:stop"
)

// TaskAction represents task-specific actions
type TaskAction string

const (
	// TaskActionNone indicates no action
	TaskActionNone TaskAction = ""
	// TaskActionCancel indicates the task action was canceled
	TaskActionCancel TaskAction = "task:cancel"
	// TaskActionNew indicates a new task should be created
	TaskActionNew TaskAction = "task:new"
	// TaskActionDelete indicates a task should be deleted
	TaskActionDelete TaskAction = "task:delete"
)

// PomodoroAction represents pomodoro-specific actions
type PomodoroAction string

const (
	// PomodoroActionNone indicates no action
	PomodoroActionNone PomodoroAction = ""
	// PomodoroActionCancel indicates the pomodoro action was canceled
	PomodoroActionCancel PomodoroAction = "pomodoro:cancel"
	// PomodoroActionContinue indicates the pomodoro should continue with the same task
	PomodoroActionContinue PomodoroAction = "pomodoro:continue"
	// PomodoroActionChange indicates the pomodoro should change to a new task
	PomodoroActionChange PomodoroAction = "pomodoro:change"
)
