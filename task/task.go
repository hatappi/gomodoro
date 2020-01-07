// Package task manage task
package task

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
	"gopkg.in/yaml.v2"

	"github.com/hatappi/gomodoro/errors"
	"github.com/hatappi/gomodoro/screen"
	"github.com/hatappi/gomodoro/screen/draw"
)

type Client interface {
	GetTask() (*Task, error)

	loadTasks() (Tasks, error)
	saveTasks(Tasks) error
}

type clientImpl struct {
	taskFile     string
	screenClient screen.Client
}

func NewClient(c screen.Client, taskFile string) Client {
	return &clientImpl{
		taskFile:     taskFile,
		screenClient: c,
	}
}

func (c *clientImpl) GetTask() (*Task, error) {
	tasks, err := c.loadTasks()
	if err != nil {
		return nil, err
	}

	var name string

	if len(tasks) > 0 {
		name, err = selectTaskName(c.screenClient, tasks.GetTaskNames())
		if err != nil {
			return nil, err
		}
	}

	t := &Task{
		Name: name,
	}

	if name == "" {
		t.Name = createTaskName(c.screenClient)
		tasks = append(tasks, t)
		err = c.saveTasks(tasks)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (c *clientImpl) loadTasks() (Tasks, error) {
	t := Tasks{}

	b, err := ioutil.ReadFile(c.taskFile)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return t, nil
		}

		return nil, err
	}

	err = yaml.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (c *clientImpl) saveTasks(tasks Tasks) error {
	d, err := yaml.Marshal(tasks)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.taskFile, d, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Task represents task
type Task struct {
	Name string `yaml:"name"`
}

// Tasks is array of Task
type Tasks []*Task

// GetTaskNames gets task names in array
func (ts Tasks) GetTaskNames() []string {
	var tn []string
	for _, t := range ts {
		tn = append(tn, t.Name)
	}

	return tn
}

func selectTaskName(c screen.Client, tasks []string) (string, error) {
	var tasksWithIndex []string
	for i, t := range tasks {
		tasksWithIndex = append(tasksWithIndex, fmt.Sprintf("%3d. %s", i+1, t))
	}

	offset := 0
	i := 0
	for {
		w, h := c.ScreenSize()
		limit := int(math.Min(float64(offset+h), float64(len(tasks))))

		for y, t := range tasksWithIndex[offset:limit] {
			opts := []draw.Option{}
			if y == i {
				opts = []draw.Option{
					draw.WithBackgroundColor(tcell.ColorBlue),
				}
			}
			tw := runewidth.StringWidth(t)
			if d := w - tw; d > 0 {
				t += strings.Repeat(" ", d)
			}
			_ = draw.Sentence(c.GetScreen(), 0, y, w, t, true, opts...)
		}

		e := <-c.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return "", errors.ErrCancel
		case screen.EventEnter:
			return tasks[offset+i], nil
		case screen.EventKeyDown:
			if offset+i == len(tasks)-1 {
				continue
			}

			if i < h-1 {
				i++
			} else {
				c.Clear()
				offset += h
				i = 0
			}
		case screen.EventKeyUp:
			if offset+i <= 0 {
				continue
			}

			if i > 0 {
				i--
			} else {
				c.Clear()
				offset -= h
				i = h - 1
			}
		case screen.EventRune:
			s := c.GetScreen()
			switch rune(e) {
			case rune(106): // j
				s.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone))
			case rune(107): // k
				s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
			case rune(110): // n
				c.Clear()
				return "", nil
			}
		case screen.EventScreenResize:
			// reset
			i = 0
			offset = 0
		}
	}
}

func createTaskName(c screen.Client) string {
	newTaskName := []rune{}
	s := c.GetScreen()
	for {
		msg := fmt.Sprintf("Please Input New Task Name >%s", string(newTaskName))
		w, _ := c.ScreenSize()
		c.Clear()
		x := draw.Sentence(s, 0, 0, w, msg, false)

		gl := ' '
		st := tcell.StyleDefault
		st = st.Background(tcell.ColorGreen)
		s.SetCell(x, 0, st, gl)
		s.Show()

		e := <-c.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return ""
		case screen.EventEnter:
			return string(newTaskName)
		case screen.EventDelete:
			if l := len(newTaskName); l > 0 {
				newTaskName = newTaskName[:l-1]
			}
		case screen.EventRune:
			newTaskName = append(newTaskName, rune(e))
		}
	}
}
