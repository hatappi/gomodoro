package task

import (
	"os"

	taskModel "github.com/hatappi/gomodoro/libs/models/task"
	"github.com/hatappi/gomodoro/libs/selector"
	"github.com/hatappi/gomodoro/libs/task"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func Get(tasks []string) (*taskModel.Task, error) {
	var (
		selectTask *taskModel.Task
		err        error
	)

	if len(tasks) > 0 {
		selectTask, err = selector.Task(tasks)
		if err != nil || !selectTask.IsSet {
			return selectTask, err
		}
	}

	if selectTask.Name != "" {
		return selectTask, nil
	}

	newTaskName := createNewTask()
	if newTaskName != "" {
		err := task.Save(append(tasks, newTaskName))
		if err != nil {
			return selectTask, err
		}
		selectTask.SetName(newTaskName)
	}

	if selectTask.Name == "" {
		selectTask.IsSet = false
	}
	return selectTask, nil
}

func createNewTask() string {
	termbox.Init()
	defer os.Stdout.Write([]byte("\x1b[?25h\r\x1b[0J"))
	newTaskName := []rune{}
	for {
		x := 0
		msg := append([]rune("Please Input New Task > "), newTaskName...)
		for _, r := range msg {
			termbox.SetCell(x, 0, r, termbox.ColorWhite, termbox.ColorDefault)
			x += runewidth.RuneWidth(r)
		}
		termbox.Flush()

		ev := termbox.PollEvent()
		switch ev.Key {
		case termbox.KeyEsc:
			return ""
		case termbox.KeySpace:
			newTaskName = append(newTaskName, ' ')
		case termbox.KeyEnter:
			if len(newTaskName) == 0 {
				continue
			}
			return string(newTaskName)
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(newTaskName) > 0 {
				newTaskName = newTaskName[:len(newTaskName)-1]
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			}
		default:
			newTaskName = append(newTaskName, ev.Ch)
		}
	}
	return string(newTaskName)
}
