// Package view provides UI components for the TUI
package view

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/core"
	gomodoro_error "github.com/hatappi/gomodoro/internal/errors"
	"github.com/hatappi/gomodoro/internal/tui/constants"
	"github.com/hatappi/gomodoro/internal/tui/screen"
	"github.com/hatappi/gomodoro/internal/tui/screen/draw"
)

// TaskView handles task listing and selection UI.
type TaskView struct {
	config       *config.Config
	screenClient screen.Client
}

// NewTaskView creates a new task view instance.
func NewTaskView(cfg *config.Config, sc screen.Client) *TaskView {
	return &TaskView{
		config:       cfg,
		screenClient: sc,
	}
}

// RenderTasks displays the task list.
func (v *TaskView) RenderTasks(tasks []*core.Task, offset, limit, cursorPosition int) {
	w, h := v.screenClient.ScreenSize()

	for y, t := range tasks[offset:limit] {
		name := fmt.Sprintf("%3d: %s", offset+y+1, t.Title)
		opts := []draw.Option{}
		if y == cursorPosition {
			opts = []draw.Option{
				draw.WithBackgroundColor(v.config.Color.SelectedLine),
			}
		}
		tw := runewidth.StringWidth(name)
		if d := w - tw; d > 0 {
			name += strings.Repeat(" ", d)
		}
		_ = draw.Sentence(v.screenClient.GetScreen(), 0, y, w, name, true, opts...)
	}

	draw.Sentence(
		v.screenClient.GetScreen(),
		0,
		h-1,
		w,
		"(n): add new task / (d): delete task",
		true,
		draw.WithBackgroundColor(v.config.Color.StatusBarBackground),
	)
}

// SelectTask displays the task selection UI and returns the selected task.
func (v *TaskView) SelectTask(_ context.Context, tasks []*core.Task) (*core.Task, constants.TaskAction, error) {
	offset := 0
	cursorPosition := 0
	for {
		_, h := v.screenClient.ScreenSize()
		selectedHeight := h - 1
		limit := int(math.Min(float64(offset+selectedHeight), float64(len(tasks))))

		v.RenderTasks(tasks, offset, limit, cursorPosition)

		e := <-v.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return nil, constants.TaskActionCancel, gomodoro_error.ErrCancel
		case screen.EventEnter:
			return tasks[offset+cursorPosition], constants.TaskActionNone, nil
		case screen.EventKeyDown:
			cursorPosition++

			if offset+cursorPosition >= len(tasks) {
				cursorPosition = len(tasks) - offset - 1
				continue
			}

			if cursorPosition >= selectedHeight {
				v.screenClient.Clear()
				offset += selectedHeight
				cursorPosition = 0
			}
		case screen.EventKeyUp:
			cursorPosition--

			if offset <= 0 && cursorPosition <= 0 {
				cursorPosition = 0
				continue
			} else if cursorPosition < 0 {
				v.screenClient.Clear()
				offset -= selectedHeight
				cursorPosition = selectedHeight - 1
			}
		case screen.EventRune:
			s := v.screenClient.GetScreen()
			switch string(e) {
			case "j":
				if err := s.PostEvent(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)); err != nil {
					return nil, constants.TaskActionNone, err
				}
			case "k":
				if err := s.PostEvent(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)); err != nil {
					return nil, constants.TaskActionNone, err
				}
			case "n":
				v.screenClient.Clear()
				return nil, constants.TaskActionNew, nil
			case "d":
				si := offset + cursorPosition
				return tasks[si], constants.TaskActionDelete, nil
			}
		case screen.EventScreenResize:
			cursorPosition = 0
			offset = 0
		}
	}
}

// CreateTaskName displays an input prompt for creating a new task.
func (v *TaskView) CreateTaskName(_ context.Context) (string, error) {
	newTaskName := []rune{}
	s := v.screenClient.GetScreen()
	for {
		msg := fmt.Sprintf("new task> %s", string(newTaskName))
		w, _ := v.screenClient.ScreenSize()
		v.screenClient.Clear()
		x := draw.Sentence(s, 0, 0, w, msg, false)

		gl := ' '
		st := tcell.StyleDefault
		st = st.Background(v.config.Color.Cursor)
		s.SetCell(x, 0, st, gl)
		s.Show()

		e := <-v.screenClient.GetEventChan()
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
