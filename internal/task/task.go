// Package task manage task
package task

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/hatappi/gomodoro/internal/client"
	"github.com/hatappi/gomodoro/internal/config"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/screen"
	"github.com/hatappi/gomodoro/internal/screen/draw"
)

// Task represents task.
type Task struct {
	ID        string `json:"id" yaml:"id"`
	Name      string `json:"title" yaml:"name"`
	Completed bool   `json:"completed" yaml:"completed"`
}

// Tasks is array of Task.
type Tasks []*Task

// Client task client.
type Client interface {
	GetTask(ctx context.Context) (*Task, error)
	LoadTasks(ctx context.Context) (Tasks, error)
	SaveTask(ctx context.Context, task *Task) error
}

// NewClient initializes Client.
func NewClient(config *config.Config, c screen.Client) *IClient {
	// Create API client
	apiClientFactory := client.NewFactory(config.API)

	taskClient := apiClientFactory.Task()

	return &IClient{
		config:       config,
		screenClient: c,
		apiClient:    taskClient,
	}
}

// IClient meets Client interface.
type IClient struct {
	config       *config.Config
	screenClient screen.Client
	apiClient    *client.TaskClient
}

// GetTask get selected Task.
func (c *IClient) GetTask(ctx context.Context) (*Task, error) {
	tasks, err := c.LoadTasks(ctx)
	if err != nil {
		return nil, err
	}

	var t *Task

	if len(tasks) > 0 {
		t, err = c.selectTaskName(ctx, tasks)
		if err != nil {
			return nil, err
		}
	}

	if t == nil {
		t, err = c.createTaskName(ctx, c.config, c.screenClient)
		if errors.Is(err, gomodoro_error.ErrCancel) {
			return nil, err
		}

		// Save new task
		err = c.SaveTask(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
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

//nolint:gocognit,cyclop
func (c *IClient) selectTaskName(ctx context.Context, tasks Tasks) (*Task, error) {
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
			return nil, gomodoro_error.ErrCancel
		case screen.EventEnter:
			return tasks[offset+cursorPosition], nil
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
				if err := s.PostEvent(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)); err != nil {
					return nil, err
				}
			case "k": // k
				if err := s.PostEvent(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)); err != nil {
					return nil, err
				}
			case "n": // n
				c.screenClient.Clear()
				return nil, nil
			case "d": // d
				si := offset + cursorPosition

				if err := c.apiClient.Delete(ctx, tasks[si].ID); err != nil {
					return nil, err
				}

				tasks = append(tasks[:si], tasks[si+1:]...)
				if len(tasks) == 0 {
					return nil, nil
				}

				c.screenClient.Clear()

				// when bottom task is deleted, key is up
				if len(tasks) == cursorPosition {
					if err := s.PostEvent(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)); err != nil {
						return nil, err
					}
				}
			}
		case screen.EventScreenResize:
			// reset
			cursorPosition = 0
			offset = 0
		}
	}
}

func (c *IClient) createTaskName(ctx context.Context, config *config.Config, sc screen.Client) (*Task, error) {
	newTaskName := []rune{}
	s := sc.GetScreen()
	for {
		msg := fmt.Sprintf("new task> %s", string(newTaskName))
		w, _ := sc.ScreenSize()
		sc.Clear()
		x := draw.Sentence(s, 0, 0, w, msg, false)

		gl := ' '
		st := tcell.StyleDefault
		st = st.Background(config.Color.Cursor)
		s.SetCell(x, 0, st, gl)
		s.Show()

		e := <-sc.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return nil, gomodoro_error.ErrCancel
		case screen.EventEnter:
			if len(newTaskName) == 0 {
				continue
			}

			resp, err := c.apiClient.Create(ctx, string(newTaskName))
			if err != nil {
				return nil, err
			}

			return &Task{
				ID:        resp.ID,
				Name:      resp.Title,
				Completed: resp.Completed,
			}, nil
		case screen.EventDelete:
			if l := len(newTaskName); l > 0 {
				newTaskName = newTaskName[:l-1]
			}
		case screen.EventRune:
			newTaskName = append(newTaskName, rune(e))
		}
	}
}

// LoadTasks loads tasks from the API
func (c *IClient) LoadTasks(ctx context.Context) (Tasks, error) {
	responses, err := c.apiClient.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make(Tasks, len(responses))
	for i, resp := range responses {
		tasks[i] = &Task{
			ID:        resp.ID,
			Name:      resp.Title,
			Completed: resp.Completed,
		}
	}
	return tasks, nil
}

// SaveTask saves a task using API client
func (c *IClient) SaveTask(ctx context.Context, task *Task) error {
	var resp *client.TaskResponse
	var err error

	if task.ID == "" {
		// Create new task
		resp, err = c.apiClient.Create(ctx, task.Name)
	} else {
		// Update existing task
		resp, err = c.apiClient.Update(ctx, task.ID, task.Name, task.Completed)
	}

	if err != nil {
		return err
	}

	// On success, set returned ID on task
	task.ID = resp.ID

	return nil
}
