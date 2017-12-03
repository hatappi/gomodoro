package main

import (
	"flag"
	"fmt"
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
	"github.com/mitchellh/go-homedir"
	"github.com/nsf/termbox-go"
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

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	confPath := flag.String("c", fmt.Sprintf("%s/.gomodoro/config.toml", homeDir), "config path")
	appDir := flag.String("a", fmt.Sprintf("%s/.gomodoro", homeDir), "application directory")
	longBreakSec := flag.Int("l", 15*60, "long break sec")
	shortBreakSec := flag.Int("s", 5*60, "short break sec")
	workSec := flag.Int("w", 25*60, "work sec")
	flag.Parse()

	conf = config.LoadConfig(*confPath)

	if conf.AppDir == "" {
		conf.AppDir = *appDir
	}
	if conf.LongBreakSec == 0 {
		conf.LongBreakSec = *longBreakSec
	}
	if conf.ShortBreakSec == 0 {
		conf.ShortBreakSec = *shortBreakSec
	}
	if conf.WorkSec == 0 {
		conf.WorkSec = *workSec
	}
}

func main() {
	taskList, err := task.GetNameList(conf.AppDir)
	if err != nil {
		panic(err)
	}

	selectTask, err := TaskSelectHandler.Get(taskList)
	if err != nil {
		panic(err)
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
	go watiKey()
	wg.Add(1)
	wg.Wait()
	timerClient.Close()
	os.Exit(1)
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
