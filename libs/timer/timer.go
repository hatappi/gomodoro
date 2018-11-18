package timer

import (
	"time"

	"github.com/hatappi/gomodoro/libs/drawer"
	"github.com/nsf/termbox-go"
)

type Timer struct {
	TaskName     string
	RemainSec    int
	ElapsedSec   int
	DrawerClient *drawer.Drawer
	IsBreak      bool
}

var (
	runnable = true
	ch       = make(chan bool)
)

func NewTimer(taskName string, remainSec int) *Timer {
	return &Timer{
		TaskName:     taskName,
		RemainSec:    remainSec,
		ElapsedSec:   0,
		DrawerClient: drawer.NewDrawer(taskName),
		IsBreak:      false,
	}
}

func (t *Timer) SetNewDrawer(taskName string) {
	t.DrawerClient = drawer.NewDrawer(taskName)
}

func (t *Timer) Close() {
	t.DrawerClient.Close()
}

func (t *Timer) Start() {
	for {
		if !runnable {
			runnable = <-ch
		}
		min, sec := t.GetRemainMinSec()

		if t.IsBreak {
			t.DrawerClient.Draw(min, sec, termbox.ColorGreen)
		} else {
			t.DrawerClient.Draw(min, sec, termbox.ColorBlue)
		}

		time.Sleep(time.Second)
		if t.RemainSec <= 0 {
			return
		}
		t.ElapsedSec += 1
		t.RemainSec -= 1
	}
}

func (t *Timer) End() {
	t.RemainSec = 0
	if t.IsBreak {
		t.DrawerClient.Draw(0, 0, termbox.ColorGreen)
	} else {
		t.DrawerClient.Draw(0, 0, termbox.ColorBlue)
	}
}

func (t *Timer) GetRemainMinSec() (int, int) {
	return t.RemainSec / 60, t.RemainSec % 60
}

func (t *Timer) WaitForNext() string {
	t.DrawerClient.DrawMessage("Please press space key to proceed")
	for {
		ev := termbox.PollEvent()
		switch ev.Key {
		case termbox.KeySpace:
			t.DrawerClient.ClearMessage()
			return t.TaskName
		default:
			if ev.Ch == 116 { // n
				t.DrawerClient.ClearMessage()
				return ""
			}
		}
	}
}

func (t *Timer) Toggle() {
	if runnable {
		t.DrawerClient.DrawMessage("Stop!!")
		runnable = false
	} else {
		ch <- true
	}
}

func (t *Timer) IsRunnable() bool {
	return runnable
}

func (t *Timer) SetRemainSec(sec int) {
	t.RemainSec = sec
	t.ElapsedSec = 0
}
