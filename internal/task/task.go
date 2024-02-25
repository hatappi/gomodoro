// Package task manage task
package task

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
	"gopkg.in/yaml.v2"

	"github.com/hatappi/gomodoro/internal/config"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
)

// Task represents task.
type Task struct {
	Name string `yaml:"name"`
}

// Tasks is array of Task.
type Tasks []*Task

// Client task client.
type Client interface {
	GetTask() (*Task, error)

	selectTaskName(tasks Tasks) (string, error)

	loadTasks() (Tasks, error)
	saveTasks(Tasks) error
}

// NewClient initilize Client.
func NewClient(config *config.Config, c screen.Client, taskFile string) *IClient {
	return &IClient{
		config:       config,
		taskFile:     taskFile,
		screenClient: c,
	}
}

// IClient meets Client interface.
type IClient struct {
	config       *config.Config
	taskFile     string
	screenClient screen.Client
}

// GetTask get selected Task.
func (c *IClient) GetTask() (*Task, error) {
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
		t.Name, err = createTaskName(c.config, c.screenClient)
		if errors.Is(err, gomodoro_error.ErrCancel) {
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

func (c *IClient) loadTasks() (Tasks, error) {
	t := Tasks{}

	// create empty file if not exist
	_, err := os.Stat(c.taskFile)
	if err != nil {
		os.WriteFile(c.taskFile, []byte{}, 0o600) //nolint:gomnd
	}

	b, err := os.ReadFile(c.taskFile)
	if err != nil {
		if errors.Is(err, &os.PathError{}) {
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

func (c *IClient) saveTasks(tasks Tasks) error {
	d, err := yaml.Marshal(tasks)
	if err != nil {
		return err
	}

	//nolint:gomnd
	err = os.WriteFile(c.taskFile, d, 0o600)
	if err != nil {
		return err
	}

	return nil
}

func (c *IClient) renderTasks(tasks Tasks, offset, limit, cursorPosition int) {
	w, h := c.screenClient.ScreenSize()

	for y, t := range tasks[offset:limit] {
		name := fmt.Sprintf("%3d: %s", offset+y+1, t.Name)
		opts := []draw.Option{}
		if y == cursorPosition {
			opts = []draw.Option{
				draw.WithBackgroundColor(c.config.Color.SelectedLine),
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
		draw.WithBackgroundColor(c.config.Color.StatusBarBackground),
	)
}

//nolint:gocognit
func (c *IClient) selectTaskName(tasks Tasks) (string, error) {
	offset := 0
	cursorPosition := 0
	for {
		_, h := c.screenClient.ScreenSize()
		selectedHeight := h - 1
		limit := int(math.Min(float64(offset+selectedHeight), float64(len(tasks))))

		c.renderTasks(tasks, offset, limit, cursorPosition)

		e := <-c.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return "", gomodoro_error.ErrCancel
		case screen.EventEnter:
			return tasks[offset+cursorPosition].Name, nil
		case screen.EventKeyDown:
			cursorPosition++

			if offset+cursorPosition >= len(tasks) {
				cursorPosition = len(tasks) - offset - 1
				continue
			}

			if cursorPosition >= selectedHeight {
				c.screenClient.Clear()
				offset += selectedHeight
				cursorPosition = 0
			}
		case screen.EventKeyUp:
			cursorPosition--

			if offset <= 0 && cursorPosition <= 0 {
				cursorPosition = 0
				continue
			} else if cursorPosition < 0 {
				c.screenClient.Clear()
				offset -= selectedHeight
				cursorPosition = selectedHeight - 1
			}
		case screen.EventRune:
			s := c.screenClient.GetScreen()
			switch string(e) {
			case "j": // j
				s.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone))
			case "k": // k
				s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
			case "n": // n
				c.screenClient.Clear()
				return "", nil
			case "d": // d
				si := offset + cursorPosition
				tasks = append(tasks[:si], tasks[si+1:]...)
				if err := c.saveTasks(tasks); err != nil {
					return "", err
				}

				if len(tasks) == 0 {
					return "", nil
				}
				c.screenClient.Clear()

				// when bottom task is deleted, key is up
				if len(tasks) == cursorPosition {
					s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
				}
			}
		case screen.EventScreenResize:
			// reset
			cursorPosition = 0
			offset = 0
		}
	}
}

func createTaskName(config *config.Config, c screen.Client) (string, error) {
	newTaskName := []rune{}
	s := c.GetScreen()
	for {
		msg := fmt.Sprintf("new task> %s", string(newTaskName))
		w, _ := c.ScreenSize()
		c.Clear()
		x := draw.Sentence(s, 0, 0, w, msg, false)

		gl := ' '
		st := tcell.StyleDefault
		st = st.Background(config.Color.Cursor)
		s.SetCell(x, 0, st, gl)
		s.Show()

		e := <-c.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return "", gomodoro_error.ErrCancel
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

// AddTask save task to file.
func AddTask(taskFile, name string) error {
	if name == "" {
		return errors.New("task name is empty")
	}

	c := &IClient{
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
