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
			selectTask.IsSet = false
			return selectTask, nil
		case termbox.KeySpace:
			newTaskName = append(newTaskName, ' ')
		case termbox.KeyEnter:
			if len(newTaskName) == 0 {
				continue
			}
			selectTask.Name = string(newTaskName)
			err := task.Save(append(tasks, selectTask.Name))
			if err != nil {
				return selectTask, err
			}
			return selectTask, nil
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(newTaskName) > 0 {
				newTaskName = newTaskName[:len(newTaskName)-1]
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			}
		default:
			newTaskName = append(newTaskName, ev.Ch)
		}
	}
	return selectTask, err
}
