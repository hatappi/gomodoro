// Package view provides UI components for the TUI
package view

import (
	"context"
	"math"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core/event"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui/constants"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/screen/draw"
)

// TimerView handles rendering of the timer UI
type TimerView struct {
	config       *config.Config
	screenClient screen.Client
}

// NewTimerView creates a new timer view instance
func NewTimerView(cfg *config.Config, sc screen.Client) *TimerView {
	return &TimerView{
		config:       cfg,
		screenClient: sc,
	}
}

const (
	marginTileRate = 16
)

// DrawTimer renders the timer UI with the current time and state
func (v *TimerView) DrawTimer(ctx context.Context, duration int, title string, phase event.PomodoroPhase, isPaused bool) error {
	screen := v.screenClient.GetScreen()

	screenWidth, screenHeight := v.screenClient.ScreenSize()

	leftMargin := float64(screenWidth) / marginTileRate
	rightMargin := float64(screenWidth) / marginTileRate
	topMargin := float64(screenHeight) / marginTileRate
	bottomMargin := float64(screenHeight) / marginTileRate

	renderWidth := float64(screenWidth) - leftMargin - rightMargin
	renderHeight := float64(screenHeight) - topMargin - bottomMargin

	textHeight := 2

	timerRenderWidth := renderWidth
	timerRenderHeight := renderHeight - float64(textHeight)

	mag, err := v.timerMagnification(timerRenderWidth, timerRenderHeight)
	if err != nil {
		return err
	}
	timerWidth := draw.TimerBaseWidth * mag
	timerHeight := draw.TimerBaseHeight * mag

	timerPaddingWidth := (timerRenderWidth - timerWidth) / 2
	timerPaddingHeight := (timerRenderHeight - timerHeight) / 2

	x := int(math.Round(leftMargin + timerPaddingWidth))
	y := int(math.Round(topMargin + timerPaddingHeight))
	log.FromContext(ctx).V(1).Info("screen information",
		"x", x,
		"y", y,
		"timerWidth", timerWidth,
		"timerHeight", timerHeight,
	)

	v.screenClient.Clear()

	draw.Sentence(screen, x, y, int(timerWidth), title, true)

	var bc tcell.Color
	if phase == event.PomodoroPhaseWork {
		bc = v.config.Color.TimerWorkFont
	} else if phase == event.PomodoroPhaseShortBreak || phase == event.PomodoroPhaseLongBreak {
		bc = v.config.Color.TimerBreakFont
	}

	if isPaused {
		bc = v.config.Color.TimerPauseFont
	}

	opts := []draw.Option{
		draw.WithBackgroundColor(bc),
	}

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
		draw.WithBackgroundColor(v.config.Color.StatusBarBackground),
	)

	return nil
}

// HandleScreenEvent processes user input events
func (v *TimerView) HandleScreenEvent(ctx context.Context, e interface{}) (constants.TimerAction, error) {
	switch ev := e.(type) {
	case screen.EventCancel:
		return constants.TimerActionCancel, gomodoro_error.ErrCancel
	case screen.EventRune:
		if string(ev) == "e" {
			return constants.TimerActionStop, nil
		}
	case screen.EventEnter:
		return constants.TimerActionToggle, nil
	}
	return constants.TimerActionNone, nil
}

func (v *TimerView) timerMagnification(w, h float64) (float64, error) {
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
