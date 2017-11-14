package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hatappi/gomodoro/config"
	"github.com/hatappi/gomodoro/libs/beep"
	"github.com/hatappi/gomodoro/libs/notification"
	"github.com/hatappi/gomodoro/libs/selector"
	"github.com/hatappi/gomodoro/libs/task"
	"github.com/hatappi/gomodoro/libs/timer"
	"github.com/hatappi/gomodoro/libs/toggl"
	"github.com/mitchellh/go-homedir"
	"github.com/nsf/termbox-go"
)

var (
	wg          sync.WaitGroup
	timerClient *timer.Timer
	configPath  string
	start       time.Time
)

const (
	longBreakSetInterval = 3
)

func init() {
	flag.StringVar(&configPath, "config", "", "config path")
	flag.Parse()

	if configPath == "" {
		homeDir, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		configPath = homeDir + "/.gomodoro/config.toml"
	}
}

func main() {
	conf := config.LoadConfig(configPath)

	tasks, err := task.GetNameList()
	if err != nil {
		panic(err)
	}

	selectTask := ""

	if len(tasks) > 0 {
		selectTask, err = selector.Task(tasks)
		if err != nil {
			panic(err)
		}
	}
	if selectTask == "" {
		fmt.Print("Please Input New Task > ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		selectTask = scanner.Text()
		err := task.Save(append(tasks, selectTask))
		if err != nil {
			panic(err)
		}
	}

	cnt := 1
	timerClient = timer.NewTimer(selectTask, getTimerSec(cnt))
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
		return 15 * 60
	} else if cnt%2 == 0 {
		// short break
		return 5 * 60
	} else {
		// work
		return 25 * 60
	}
}
