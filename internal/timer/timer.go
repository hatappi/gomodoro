// Package timer countdown duration
package timer

import (
	"context"
	"math"
	"time"

	"github.com/gdamore/tcell"
	"golang.org/x/xerrors"

	"github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
)

// Timer interface
type Timer interface {
	Run(context.Context) (int, error)
	Stop()
	SetTitle(string)
	GetTitle() string
	SetDuration(sec int)
	GetRemainSec() int

	ChangeFontColor(tcell.Color)
}

type timerImpl struct {
	title        string
	ticker       *time.Ticker
	screenClient screen.Client
	stopped      bool

	fontColor      tcell.Color
	pauseFontColor tcell.Color

	remainSec int
}

// NewTimer initilize Timer
func NewTimer(c screen.Client) Timer {
	return &timerImpl{
		ticker:         nil,
		screenClient:   c,
		fontColor:      tcell.ColorGreen,
		pauseFontColor: tcell.ColorDarkOrange,
	}
}

func (t *timerImpl) SetTitle(title string) {
	t.title = title
}

func (t *timerImpl) GetTitle() string {
	return t.title
}

func (t *timerImpl) SetDuration(d int) {
	t.remainSec = d
}

func (t *timerImpl) GetRemainSec() int {
	return t.remainSec
}

func (t *timerImpl) ChangeFontColor(c tcell.Color) {
	t.fontColor = c
}

// Run timer
func (t *timerImpl) Run(ctx context.Context) (int, error) {
	t.Start()
	defer t.Stop()

	opts := []draw.Option{
		draw.WithBackgroundColor(t.fontColor),
	}

	elapsedTime := 0

	for {
		err := t.drawTimer(ctx, t.remainSec, t.title, opts...)
		if err != nil {
			if xerrors.Is(err, errors.ErrScreenSmall) {
				t.screenClient.Clear()
				w, h := t.screenClient.ScreenSize()
				draw.Sentence(t.screenClient.GetScreen(), 0, h/2, w, "Please large screen", true)

				select {
				case <-t.ticker.C:
					continue
				case e := <-t.screenClient.GetEventChan():
					switch e.(type) {
					case screen.EventCancel:
						return elapsedTime, errors.ErrCancel
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
				return elapsedTime, errors.ErrCancel
			case screen.EventRune:
				if rune(e) == rune(101) { // e
					t.remainSec = 0
				}
			case screen.EventEnter:
				if t.stopped {
					opts = []draw.Option{
						draw.WithBackgroundColor(t.fontColor),
					}
					t.Start()
				} else {
					opts = []draw.Option{
						draw.WithBackgroundColor(t.pauseFontColor),
					}
					t.Stop()
				}
			}
		case <-t.ticker.C:
			t.remainSec--
			elapsedTime++
		}
	}
}

// Start timer
func (t *timerImpl) Start() {
	t.stopped = false
	t.ticker = time.NewTicker(1 * time.Second)
}

// Stop timer
func (t *timerImpl) Stop() {
	t.stopped = true
	t.ticker.Stop()
}

func (t *timerImpl) drawTimer(ctx context.Context, duration int, title string, opts ...draw.Option) error {
	s := t.screenClient.GetScreen()

	screenWidth, screenHeight := t.screenClient.ScreenSize()

	marginWidth := float64(screenWidth) / 16
	marginHeight := float64(screenHeight) / 16

	// renderWidth subtracts left and right margin from screen width
	renderWidth := float64(screenWidth) - (marginWidth * 2)
	// renderHeight subtracts top and bottom margin from screen height
	renderHeight := float64(screenHeight) - (marginHeight * 2)

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

	timerPaddingWidth := (timerRenderWidth - timerWidth) / 2
	timerPaddingHeight := (timerRenderHeight - timerHeight) / 2

	x := int(math.Round(marginWidth + timerPaddingWidth))
	y := int(math.Round(marginHeight + timerPaddingHeight))

	t.screenClient.Clear()

	draw.Sentence(s, x, y, int(draw.TimerBaseWidth*mag), title, true)

	min := duration / 60
	sec := duration % 60
	draw.Timer(s, x, y+textHeight, int(mag), min, sec, opts...)

	draw.Sentence(
		s,
		0,
		screenWidth-1,
		screenHeight,
		"(e): end timer / (Enter): stop start timer",
		true,
		draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
	)

	return nil
}

func timerMagnification(w, h float64) (float64, error) {
	x := math.Round(w / draw.TimerBaseWidth)
	y := math.Round(h / draw.TimerBaseHeight)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, errors.ErrScreenSmall
		}

		if w >= draw.TimerBaseWidth*mag && h >= draw.TimerBaseHeight*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}
