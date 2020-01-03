package pomodoro

// Option pomodoro option
type Option func(*pomodoroImpl)

// WithWorkSec set WorkSec
func WithWorkSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.workSec = s
	}
}

// WithShortBreakSec set ShortBreakSec
func WithShortBreakSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.shortBreakSec = s
	}
}

// WithLongBreakSec set LongBreakSec
func WithLongBreakSec(s int) Option {
	return func(p *pomodoroImpl) {
		p.longBreakSec = s
	}
}
