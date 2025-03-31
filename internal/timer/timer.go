// Package timer countdown duration
package timer

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
)

// Timer interface.
type Timer interface {
	Run(ctx context.Context, taskName string) (int, error)
}

// ITimer implements Timer interface.
type ITimer struct {
	config         *config.Config
	screenClient   screen.Client
	pomodoroClient *client.PomodoroClient
	eventBus       event.EventBus
}

// NewTimer initilize Timer.
func NewTimer(config *config.Config, c screen.Client, pomodoroClient *client.PomodoroClient, eventBus event.EventBus) *ITimer {
	return &ITimer{
		config:         config,
		screenClient:   c,
		pomodoroClient: pomodoroClient,
		eventBus:       eventBus,
	}
}

// handleScreenEvent processes screen events
func (t *ITimer) handleScreenEvent(ctx context.Context, e interface{}) error {
	switch ev := e.(type) {
	case screen.EventCancel:
		return gomodoro_error.ErrCancel
	case screen.EventRune:
		if string(ev) == "e" {
			_, stopErr := t.pomodoroClient.Stop(ctx)
			if stopErr != nil {
				log.FromContext(ctx).Error(stopErr, "failed to stop pomodoro")
				return stopErr
			}
			return nil
		}
	case screen.EventEnter:
		t.Toggle(ctx)
	}
	return nil
}

// handlePomodoroEvent processes pomodoro events
func (t *ITimer) handlePomodoroEvent(ctx context.Context, e interface{}, title string) (bool, error) {
	ev, ok := e.(event.PomodoroEvent)
	if !ok {
		return false, nil
	}
	log.FromContext(ctx).Info("event", "event", ev, "remainSec", ev.RemainingTime.Seconds())

	remainSec := int(ev.RemainingTime.Seconds())

	var bc tcell.Color
	// Use event.PomodoroPhase constants for comparison
	if ev.Phase == event.PomodoroPhaseWork {
		bc = t.config.Color.TimerWorkFont
	} else if ev.Phase == event.PomodoroPhaseShortBreak || ev.Phase == event.PomodoroPhaseLongBreak {
		bc = t.config.Color.TimerBreakFont
	}

	if ev.Type == event.PomodoroPaused {
		bc = t.config.Color.TimerPauseFont
	}

	opts := []draw.Option{
		draw.WithBackgroundColor(bc),
	}
	drawErr := t.drawTimer(ctx, remainSec, title, opts...)
	if drawErr != nil {
		if errors.Is(drawErr, gomodoro_error.ErrScreenSmall) {
			t.screenClient.Clear()
			w, h := t.screenClient.ScreenSize()
			//nolint:mnd
			draw.Sentence(t.screenClient.GetScreen(), 0, h/2, w, "Please expand the screen size", true)

			select {
			case e := <-t.screenClient.GetEventChan():
				switch e.(type) {
				case screen.EventCancel:
					return false, gomodoro_error.ErrCancel
				case screen.EventScreenResize:
					return false, nil
				}
			}
		}
		return false, drawErr
	}

	if ev.Type == event.PomodoroCompleted || ev.Type == event.PomodoroStopped {
		return true, nil
	}
	return false, nil
}

// getCurrentPomodoro retrieves the current pomodoro and handles errors.
func (t *ITimer) getCurrentPomodoro(ctx context.Context) (int, error) {
	current, err := t.pomodoroClient.GetCurrent(ctx)
	if err != nil {
		return 0, err
	}
	return current.ElapsedTime, nil
}

// Run timer.
func (t *ITimer) Run(ctx context.Context, taskName string) (int, error) {
	ch, unsubscribe := t.eventBus.SubscribeChannel([]string{
		string(event.PomodoroTick),
		string(event.PomodoroPaused),
		string(event.PomodoroStarted),
		string(event.PomodoroStopped),
		string(event.PomodoroCompleted),
	})
	defer unsubscribe()

	for {
		select {
		case e := <-t.screenClient.GetEventChan():
			err := t.handleScreenEvent(ctx, e)
			if err != nil {
				if errors.Is(err, gomodoro_error.ErrCancel) {
					elapsedTime, _ := t.getCurrentPomodoro(ctx)
					return elapsedTime, gomodoro_error.ErrCancel
				}
				return 0, err
			}
		case e := <-ch:
			isCompleted, err := t.handlePomodoroEvent(ctx, e, taskName)
			if err != nil {
				if errors.Is(err, gomodoro_error.ErrCancel) {
					elapsedTime, _ := t.getCurrentPomodoro(ctx)
					return elapsedTime, gomodoro_error.ErrCancel
				}
				return 0, err
			}

			if isCompleted {
				elapsedTime, _ := t.getCurrentPomodoro(ctx)
				return elapsedTime, nil
			}
		}
	}
}

// Toggle timer between stop and start.
func (t *ITimer) Toggle(ctx context.Context) {
	currPomodoro, _ := t.pomodoroClient.GetCurrent(ctx)
	if currPomodoro == nil {
		return
	}

	if currPomodoro.State == "paused" {
		_, _ = t.pomodoroClient.Resume(ctx)
	} else {
		_, _ = t.pomodoroClient.Pause(ctx)
	}
}

const (
	marginTileRate = 16
)

func (t *ITimer) drawTimer(ctx context.Context, duration int, title string, opts ...draw.Option) error {
	screen := t.screenClient.GetScreen()

	screenWidth, screenHeight := t.screenClient.ScreenSize()

	leftMargin := float64(screenWidth) / marginTileRate
	rightMargin := float64(screenWidth) / marginTileRate
	topMargin := float64(screenHeight) / marginTileRate
	bottomMargin := float64(screenHeight) / marginTileRate

	// renderWidth subtracts left and right margin from screen width
	renderWidth := float64(screenWidth) - leftMargin - rightMargin
	// renderHeight subtracts top and bottom margin from screen height
	renderHeight := float64(screenHeight) - topMargin - bottomMargin

	// text height is 1, but add 1 to include margin
	textHeight := 2

	timerRenderWidth := renderWidth
	timerRenderHeight := renderHeight - float64(textHeight)

	mag, err := timerMagnification(timerRenderWidth, timerRenderHeight)
	if err != nil {
		return err
	}
	timerWidth := draw.TimerBaseWidth * mag
	timerHeight := draw.TimerBaseHeight * mag

	timerPaddingWidth := (timerRenderWidth - timerWidth) / 2    //nolint:mnd
	timerPaddingHeight := (timerRenderHeight - timerHeight) / 2 //nolint:mnd

	x := int(math.Round(leftMargin + timerPaddingWidth))
	y := int(math.Round(topMargin + timerPaddingHeight))
	log.FromContext(ctx).V(1).Info("screen information",
		"x", x,
		"y", y,
		"timerWidth", timerWidth,
		"timerHeight", timerHeight,
	)

	t.screenClient.Clear()

	draw.Sentence(screen, x, y, int(timerWidth), title, true)

	m := duration / int(time.Minute.Seconds())
	s := duration % int(time.Minute.Seconds())
	draw.Timer(screen, x, y+textHeight, int(mag), m, s, opts...)

	draw.Sentence(
		screen,
		0,
		screenWidth-1,
		screenHeight,
		"(e): end timer / (Enter): stop start timer",
		true,
		draw.WithBackgroundColor(t.config.Color.StatusBarBackground),
	)

	return nil
}

func timerMagnification(w, h float64) (float64, error) {
	x := math.Round(w / draw.TimerBaseWidth)
	y := math.Round(h / draw.TimerBaseHeight)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, gomodoro_error.ErrScreenSmall
		}

		if w >= draw.TimerBaseWidth*mag && h >= draw.TimerBaseHeight*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}
