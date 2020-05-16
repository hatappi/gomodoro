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
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"

	"github.com/hatappi/gomodoro/errors"
	"github.com/hatappi/gomodoro/screen"
	"github.com/hatappi/gomodoro/screen/draw"
)

// Task represents task
type Task struct {
	Name string `yaml:"name"`
}

// Tasks is array of Task
type Tasks []*Task

// Client task client
type Client interface {
	GetTask() (*Task, error)

	selectTaskName(tasks Tasks) (string, error)

	loadTasks() (Tasks, error)
	saveTasks(Tasks) error
}

// NewClient initilize Client
func NewClient(c screen.Client, taskFile string) Client {
	return &clientImpl{
		taskFile:     taskFile,
		screenClient: c,
	}
}

type clientImpl struct {
	taskFile     string
	screenClient screen.Client
}

func (c *clientImpl) GetTask() (*Task, error) {
	tasks, err := c.loadTasks()
	if err != nil {
		return nil, err
	}

	var name string

	if len(tasks) > 0 {
		name, err = c.selectTaskName(tasks)
		if err != nil {
			return nil, err
		}
	}

	t := &Task{
		Name: name,
	}

	if t.Name == "" {
		t.Name, err = createTaskName(c.screenClient)
		if xerrors.Is(err, errors.ErrCancel) {
			return nil, err
		}
		tasks, err = c.loadTasks()
		if err != nil {
			return nil, err
		}

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

	err = ioutil.WriteFile(c.taskFile, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) selectTaskName(tasks Tasks) (string, error) {
	offset := 0
	i := 0
	for {
		w, h := c.screenClient.ScreenSize()
		selectedHeight := h - 1
		limit := int(math.Min(float64(offset+selectedHeight), float64(len(tasks))))

		for y, t := range tasks[offset:limit] {
			name := fmt.Sprintf("%3d. %s", y+1, t.Name)
			opts := []draw.Option{}
			if y == i {
				opts = []draw.Option{
					draw.WithBackgroundColor(tcell.ColorBlue),
				}
			}
			tw := runewidth.StringWidth(name)
			if d := w - tw; d > 0 {
				name += strings.Repeat(" ", d)
			}
			_ = draw.Sentence(c.screenClient.GetScreen(), 0, y, w, name, true, opts...)
		}

		draw.Sentence(
			c.screenClient.GetScreen(),
			0,
			h-1,
			w,
			"(n): add new task / (d): delete task",
			true,
			draw.WithBackgroundColor(draw.StatusBarBackgroundColor),
		)

		e := <-c.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return "", errors.ErrCancel
		case screen.EventEnter:
			return tasks[offset+i].Name, nil
		case screen.EventKeyDown:
			if offset+i >= len(tasks)-1 {
				i = len(tasks) - offset - 1
				continue
			}

			if i < selectedHeight-1 {
				i++
			} else {
				c.screenClient.Clear()
				offset += selectedHeight
				i = 0
			}
		case screen.EventKeyUp:
			if offset+i <= 0 {
				continue
			}

			if i > 0 {
				i--
			} else {
				c.screenClient.Clear()
				offset -= selectedHeight
				i = selectedHeight - 1
			}
		case screen.EventRune:
			s := c.screenClient.GetScreen()
			switch rune(e) {
			case rune(106): // j
				s.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone))
			case rune(107): // k
				s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
			case rune(110): // n
				c.screenClient.Clear()
				return "", nil
			case rune(100): // d
				si := offset + i
				tasks = append(tasks[:si], tasks[si+1:]...)
				err := c.saveTasks(tasks)
				if err != nil {
					return "", err
				}
				if len(tasks) == 0 {
					return "", nil
				}
				c.screenClient.Clear()

				// when bottom task is deleted, key is up
				if len(tasks) == i {
					s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
				}
			}
		case screen.EventScreenResize:
			// reset
			i = 0
			offset = 0
		}
	}
}

func createTaskName(c screen.Client) (string, error) {
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
			return "", errors.ErrCancel
		case screen.EventEnter:
			if len(newTaskName) == 0 {
				continue
			}
			return string(newTaskName), nil
		case screen.EventDelete:
			if l := len(newTaskName); l > 0 {
				newTaskName = newTaskName[:l-1]
			}
		case screen.EventRune:
			newTaskName = append(newTaskName, rune(e))
		}
	}
}

// AddTask save task to file
func AddTask(taskFile, name string) error {
	if name == "" {
		return xerrors.New("task name is empty")
	}

	c := &clientImpl{
		taskFile: taskFile,
	}

	tasks, err := c.loadTasks()
	if err != nil {
		return err
	}

	tasks = append(tasks, &Task{Name: name})

	err = c.saveTasks(tasks)
	if err != nil {
		return err
	}

	return nil
}
