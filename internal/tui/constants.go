package tui

// TimerAction represents timer-specific actions
type TimerAction string

const (
	TimerActionNone   TimerAction = ""
	TimerActionCancel TimerAction = "timer:cancel"
	TimerActionToggle TimerAction = "timer:toggle"
	TimerActionStop   TimerAction = "timer:stop"
)

// TaskAction represents task-specific actions
type TaskAction string

const (
	TaskActionNone   TaskAction = ""
	TaskActionCancel TaskAction = "task:cancel"
	TaskActionNew    TaskAction = "task:new"
	TaskActionDelete TaskAction = "task:delete"
)

// PomodoroAction represents pomodoro-specific actions
type PomodoroAction string

const (
	PomodoroActionNone     PomodoroAction = ""
	PomodoroActionCancel   PomodoroAction = "pomodoro:cancel"
	PomodoroActionContinue PomodoroAction = "pomodoro:continue"
	PomodoroActionChange   PomodoroAction = "pomodoro:change"
)
