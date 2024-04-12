// Package timer countdown duration
package timer

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/gdamore/tcell"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/config"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
)

// Timer interface.
type Timer interface {
	Run(ctx context.Context) (int, error)
	Stop()
	SetTitle(title string)
	GetTitle() string
	GetRemainSec() int
	SetDuration(sec int)
	SetFontColor(c tcell.Color)
}

// ITimer implements Timer interface.
type ITimer struct {
	config       *config.Config
	title        string
	ticker       *time.Ticker
	screenClient screen.Client
	stopped      bool

	fontColor tcell.Color

	remainSec int
}

// NewTimer initilize Timer.
func NewTimer(config *config.Config, c screen.Client) *ITimer {
	return &ITimer{
		config:       config,
		ticker:       nil,
		screenClient: c,
		fontColor:    config.Color.TimerWorkFont,
	}
}

// SetTitle sets title.
func (t *ITimer) SetTitle(title string) {
	t.title = title
}

// GetTitle gets title.
func (t *ITimer) GetTitle() string {
	return t.title
}

// SetDuration sets remaining seconds.
func (t *ITimer) SetDuration(d int) {
	t.remainSec = d
}

// GetRemainSec get remaining seconds of timer.
func (t *ITimer) GetRemainSec() int {
	return t.remainSec
}

// SetFontColor sets cell color.
func (t *ITimer) SetFontColor(c tcell.Color) {
	t.fontColor = c
}

// Run timer.
func (t *ITimer) Run(ctx context.Context) (int, error) {
	t.Start()
	defer t.Stop()

	elapsedTime := 0

	for {
		opts := []draw.Option{}
		if t.stopped {
			opts = append(opts, draw.WithBackgroundColor(t.config.Color.TimerPauseFont))
		} else {
			opts = append(opts, draw.WithBackgroundColor(t.fontColor))
		}
		err := t.drawTimer(ctx, t.remainSec, t.title, opts...)
		if err != nil {
			if errors.Is(err, gomodoro_error.ErrScreenSmall) {
				t.screenClient.Clear()
				w, h := t.screenClient.ScreenSize()
				//nolint:gomnd
				draw.Sentence(t.screenClient.GetScreen(), 0, h/2, w, "Please expand the screen size", true)

				select {
				case <-t.ticker.C:
					continue
				case e := <-t.screenClient.GetEventChan():
					switch e.(type) {
					case screen.EventCancel:
						return elapsedTime, gomodoro_error.ErrCancel
					case screen.EventScreenResize:
						continue
					}
				}
			}
			return elapsedTime, err
		}

		if t.remainSec == 0 {
			return elapsedTime, nil
		}

		select {
		case e := <-t.screenClient.GetEventChan():
			switch e := e.(type) {
			case screen.EventCancel:
				return elapsedTime, gomodoro_error.ErrCancel
			case screen.EventRune:
				if string(e) == "e" { // e
					t.remainSec = 0
				}
			case screen.EventEnter:
				t.Toggle()
			}
		case <-t.ticker.C:
			t.remainSec--
			elapsedTime++
		}
	}
}

// Start timer.
func (t *ITimer) Start() {
	t.stopped = false
	t.ticker = time.NewTicker(1 * time.Second)
}

// Stop timer.
func (t *ITimer) Stop() {
	t.stopped = true
	t.ticker.Stop()
}

// Toggle timer between stop and start.
func (t *ITimer) Toggle() {
	if t.stopped {
		t.Start()
	} else {
		t.Stop()
	}
}

const (
	marginTileRate = 16
)

func (t *ITimer) drawTimer(ctx context.Context, duration int, title string, opts ...draw.Option) error {
	s := t.screenClient.GetScreen()

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

	timerPaddingWidth := (timerRenderWidth - timerWidth) / 2    //nolint:gomnd
	timerPaddingHeight := (timerRenderHeight - timerHeight) / 2 //nolint:gomnd

	x := int(math.Round(leftMargin + timerPaddingWidth))
	y := int(math.Round(topMargin + timerPaddingHeight))
	log.FromContext(ctx).V(1).Info("screen information",
		"x", x,
		"y", y,
		"timerWidth", timerWidth,
		"timerHeight", timerHeight,
	)

	t.screenClient.Clear()

	draw.Sentence(s, x, y, int(timerWidth), title, true)

	min := duration / int(time.Minute.Seconds())
	sec := duration % int(time.Minute.Seconds())
	draw.Timer(s, x, y+textHeight, int(mag), min, sec, opts...)

	draw.Sentence(
		s,
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
