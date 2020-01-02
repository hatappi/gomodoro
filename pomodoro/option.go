package pomodoro

type PomodoroOption func(*pomodoroImpl)

func WithWorkSec(s int) PomodoroOption {
	return func(p *pomodoroImpl) {
		p.workSec = s
	}
}
func WithShortBreakSec(s int) PomodoroOption {
	return func(p *pomodoroImpl) {
		p.shortBreakSec = s
	}
}
func WithLongBreakSec(s int) PomodoroOption {
	return func(p *pomodoroImpl) {
		p.longBreakSec = s
	}
}
