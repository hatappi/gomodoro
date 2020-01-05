// Package timer countdown duration
package timer

import (
	"math"
	"time"

	"github.com/gdamore/tcell"
	"golang.org/x/xerrors"

	"github.com/hatappi/gomodoro/errors"
	"github.com/hatappi/gomodoro/screen"
	"github.com/hatappi/gomodoro/screen/draw"
)

// Timer interface
type Timer interface {
	Run(int) error
	Stop()
	SetTitle(string)

	ChangeFontColor(tcell.Color)
}

type timerImpl struct {
	title        string
	ticker       *time.Ticker
	screenClient screen.Client
	stopped      bool

	fontColor      tcell.Color
	pauseFontColor tcell.Color
}

// NewTimer initilize Timer
func NewTimer(c screen.Client, title string) Timer {
	return &timerImpl{
		title:          title,
		ticker:         nil,
		screenClient:   c,
		fontColor:      tcell.ColorGreen,
		pauseFontColor: tcell.ColorDarkOrange,
	}
}

func (t *timerImpl) SetTitle(title string) {
	t.title = title
}

func (t *timerImpl) ChangeFontColor(c tcell.Color) {
	t.fontColor = c
}

// Run timer
func (t *timerImpl) Run(duration int) error {
	s := t.screenClient.GetScreen()

	drawFn := func(duration int, title string, opts ...draw.Option) error {
		w, h := t.screenClient.ScreenSize()
		min := duration / 60
		sec := duration % 60

		x := float64(w) / 16
		y := float64(h) / 16

		printLine := 2.0
		cw := float64(w) * 14 / 16
		ch := float64(h) * 14 / 16
		ch -= printLine

		mag, err := getMagnification(cw, ch)
		if err != nil {
			return err
		}

		x = math.Trunc(x + ((cw - (draw.TimerWidth * mag)) / 2))
		y = math.Trunc(y + ((ch - (draw.TimerHeight * mag)) / 2))

		t.screenClient.Clear()
		draw.Sentence(s, int(x), int(y), int(draw.TimerWidth*mag), title, true)
		draw.Timer(s, int(x), int(y)+2, int(mag), min, sec, opts...)
		draw.Sentence(
			s,
			0,
			h-1,
			w,
			"(e): end timer / (Enter): stop start timer",
			true,
			draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
		)

		return nil
	}

	t.Start()
	defer t.Stop()

	opts := []draw.Option{
		draw.WithBackgroundColor(t.fontColor),
	}

	for {
		err := drawFn(duration, t.title, opts...)
		if err != nil {
			if xerrors.Is(err, errors.ErrScreenSmall) {
				t.screenClient.Clear()
				w, h := t.screenClient.ScreenSize()
				draw.Sentence(t.screenClient.GetScreen(), 0, h/2-1, w, "Please large screen", true)

				select {
				case <-t.ticker.C:
					continue
				case e := <-t.screenClient.GetEventChan():
					switch e.(type) {
					case screen.EventCancel:
						return errors.ErrCancel
					case screen.EventScreenResize:
						continue
					}
				}
			}
			return err
		}

		if duration == 0 {
			return nil
		}

		select {
		case e := <-t.screenClient.GetEventChan():
			switch e := e.(type) {
			case screen.EventCancel:
				return errors.ErrCancel
			case screen.EventRune:
				if rune(e) == rune(101) { // e
					duration = 0
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
			duration--
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

func getMagnification(w, h float64) (float64, error) {
	x := math.Round(w / draw.TimerWidth)
	y := math.Round(h / draw.TimerHeight)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, errors.ErrScreenSmall
		}

		if w >= draw.TimerWidth*mag && h >= draw.TimerHeight*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}
