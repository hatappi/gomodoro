package commands

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hatappi/gomodoro/config"
	"github.com/hatappi/gomodoro/libs/beep"
	TaskSelectHandler "github.com/hatappi/gomodoro/libs/handler/selection/task"
	"github.com/hatappi/gomodoro/libs/models/task"
	"github.com/hatappi/gomodoro/libs/notification"
	"github.com/hatappi/gomodoro/libs/timer"
	"github.com/hatappi/gomodoro/libs/toggl"
	"github.com/nsf/termbox-go"
	"github.com/urfave/cli"
)

var (
	wg          sync.WaitGroup
	timerClient *timer.Timer
	conf        *config.Config
	start       time.Time
)

const (
	longBreakSetInterval = 3
)

func initConfig(c *cli.Context) {
	conf = config.LoadConfig(c.GlobalString("conf-path"))

	if conf.AppDir == "" {
		conf.AppDir = c.GlobalString("app-dir")
	}
	if conf.LongBreakSec == 0 {
		conf.LongBreakSec = c.Int("long-break-sec")
	}
	if conf.ShortBreakSec == 0 {
		conf.ShortBreakSec = c.Int("short-break-sec")
	}
	if conf.WorkSec == 0 {
		conf.WorkSec = c.Int("work-sec")
	}
}

func Start(c *cli.Context) error {
	initConfig(c)
	socketPath := c.GlobalString("socket-path")
	defer os.Remove(socketPath)

	taskList, err := task.GetNameList(conf.AppDir)
	if err != nil {
		return err
	}

	selectTask, err := TaskSelectHandler.Get(taskList)
	if err != nil {
		return err
	}
	if !selectTask.IsSet {
		os.Exit(0)
	}

	cnt := 1
	timerClient = timer.NewTimer(selectTask.Name, getTimerSec(cnt))

	go func() {
		// main loop
		for {
			start = time.Now()
			timerClient.Start()
			// if work time
			if cnt%2 == 1 {
				go toggl.PostTimeEntry(conf.Toggl, timerClient.TaskName, start, timerClient.Duration)
			}
			// notify
			go func() {
				err := beep.Beep()
				if err != nil {
					panic(err)
				}
			}()
			go func() {
				err := notification.NotifyDesktop("Gomodoro", "Finish!")
				if err != nil {
					panic(err)
				}
			}()
			timerClient.WaitForNext()
			cnt += 1
			timerClient.SetRemainSec(getTimerSec(cnt))
		}
	}()
	go openSocket(socketPath)
	go watiKey()
	wg.Add(1)
	wg.Wait()
	timerClient.Close()
	return nil
}

func openSocket(socketPath string) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			defer conn.Close()

			min, sec := timerClient.GetRemainMinSec()
			sendMsg := fmt.Sprintf("%02d:%02d", min, sec)
			conn.Write([]byte(sendMsg))
		}()
	}
}

func watiKey() {
	for {
		ev := termbox.PollEvent()
		switch ev.Key {
		case termbox.KeyCtrlC:
			wg.Done()
			return
		case termbox.KeyEsc:
			wg.Done()
			return
		case termbox.KeyEnter:
			timerClient.Toggle()
		}
	}
}

func getTimerSec(cnt int) int {
	setNum := cnt / 2
	if setNum != 0 && cnt%2 == 0 && setNum%longBreakSetInterval == 0 {
		// long break
		return conf.LongBreakSec
	} else if cnt%2 == 0 {
		// short break
		return conf.ShortBreakSec
	} else {
		// work
		return conf.WorkSec
	}
}
