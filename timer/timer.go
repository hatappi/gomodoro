package timer

import (
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/hatappi/gomodoro/timer/screen"
)

type Timer interface {
	Run(int) error
	Stop()
}

type TimerImpl struct {
	ticker       *time.Ticker
	screenClient screen.Client
	stopped      bool
}

func NewTimer(c screen.Client) Timer {
	return &TimerImpl{
		ticker:       nil,
		screenClient: c,
	}
}

func (t *TimerImpl) Run(duration int) error {
	t.Start()
	for {
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

		x = math.Round(x + ((cw - (screen.TimerWidth * mag)) / 2))
		y = math.Round(y + ((ch - (screen.TimerHeight * mag)) / 2))

		t.screenClient.Clear()
		t.screenClient.DrawSentence(int(x), int(y), int(screen.TimerWidth*mag), "今年は令和2年です")
		t.screenClient.DrawTimer(int(x), int(y)+2, int(mag), min, sec)

		select {
		case <-t.screenClient.GetQuitChan():
			return nil
		case <-t.screenClient.GetPauseChan():
			if t.stopped {
				t.Start()
			} else {
				t.Stop()
				continue
			}
		case <-t.ticker.C:
		}

		duration--

		if duration == 0 {
			t.Stop()
		}
	}
}

func (t *TimerImpl) Start() {
	t.stopped = false
	t.ticker = time.NewTicker(1 * time.Second)
}

func (t *TimerImpl) Stop() {
	t.stopped = true
	t.ticker.Stop()
}

func getMagnification(w, h float64) (float64, error) {
	x := math.Round(w / screen.TimerWidth)
	y := math.Round(h / screen.TimerHeight)
	mag := math.Max(x, y)

	for {
		if mag < 1.0 {
			return 0.0, errors.New("screen is small")
		}

		if w >= screen.TimerWidth*mag && h >= screen.TimerHeight*mag {
			break
		}

		mag -= 1.0
	}

	return mag, nil
}
