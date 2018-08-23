package timer

import (
	"fmt"
	"time"

	"github.com/hatappi/gomodoro/libs/drawer"
)

type Timer struct {
	TaskName     string
	RemainSec    int
	ElapsedSec   int
	DrawerClient *drawer.Drawer
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
		min, sec := t.GetRemainMinSec()
		t.DrawerClient.Draw(min, sec)
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
}

func (t *Timer) GetRemainMinSec() (int, int) {
	return t.RemainSec / 60, t.RemainSec % 60
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
}
