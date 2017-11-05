package timer

import (
	"fmt"
	"time"

	"github.com/hatappi/gomodoro/libs/drawer"
)

type Timer struct {
	TaskName     string
	RemainSec    int
	Duration     int
	DrawerClient *drawer.Drawer
}

var (
	runnable = true
	ch       = make(chan bool)
)

func NewTimer(taskName string, remainSec int) *Timer {
	return &Timer{
		taskName,
		remainSec,
		remainSec,
		drawer.NewDrawer(taskName),
	}
}

func (t *Timer) Close() {
	t.DrawerClient.Close()
}

func (t *Timer) Start() {
	for {
		if !runnable {
			runnable = <-ch
		}
		t.DrawerClient.Draw(t.RemainSec/60, t.RemainSec%60)
		time.Sleep(time.Second)
		t.RemainSec -= 1
		if t.RemainSec < 0 {
			return
		}
	}
}

func (t *Timer) WaitForNext() {
	t.DrawerClient.DrawMessage("Please press any key to proceed")
	fmt.Scanln()
	t.DrawerClient.ClearMessage()
}

func (t *Timer) Toggle() {
	if runnable {
		runnable = false
	} else {
		ch <- true
	}
}

func (t *Timer) SetRemainSec(sec int) {
	t.RemainSec = sec
	t.Duration = sec
}
