// Package view provides UI components for the TUI
package view

import (
	"context"
	"fmt"
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

	selectCursor int
	selectOffset int
}

// NewTaskView creates a new task view instance.
func NewTaskView(cfg *config.Config, sc screen.Client) *TaskView {
	return &TaskView{
		config:       cfg,
		screenClient: sc,
	}
}

// SelectTask displays the task selection UI and returns the selected task.
func (v *TaskView) SelectTask(
	_ context.Context,
	tasks []*core.Task,
	resetCursorPosition bool,
) (*core.Task, constants.TaskAction, error) {
	v.screenClient.Clear()

	if resetCursorPosition || len(tasks) == 0 {
		v.selectOffset = 0
		v.selectCursor = 0
	}

	for {
		renderableHeight := v.getSelectRenderableHeight()

		renderedTasks := v.renderTasks(tasks, v.selectOffset, renderableHeight, v.selectCursor)

		e := <-v.screenClient.GetEventChan()
		switch e := e.(type) {
		case screen.EventCancel:
			return nil, constants.TaskActionCancel, gomodoro_error.ErrCancel
		case screen.EventEnter:
			return renderedTasks[v.selectCursor], constants.TaskActionNone, nil
		case screen.EventKeyDown:
			v.selectCursor++

			renderedTaskNum := len(renderedTasks)

			if v.selectCursor >= renderedTaskNum {
				if (v.selectOffset + renderedTaskNum) >= len(tasks) {
					v.selectCursor = renderedTaskNum - 1
					continue
				}

				v.selectOffset += renderedTaskNum
				v.selectCursor = 0
				v.screenClient.Clear()
			}
		case screen.EventKeyUp:
			v.selectCursor--

			if v.selectCursor < 0 {
				if v.selectOffset == 0 {
					v.selectCursor = 0
					continue
				}

				v.selectOffset = max(v.selectOffset-renderableHeight, 0)
				v.selectCursor = renderableHeight - 1
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
				return renderedTasks[v.selectCursor], constants.TaskActionDelete, nil
			}
		case screen.EventScreenResize:
			renderableHeight := v.getSelectRenderableHeight()

			v.selectCursor = min(v.selectCursor, renderableHeight-1)

			v.screenClient.Clear()
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

// renderTasks displays the task list.
func (v *TaskView) renderTasks(tasks []*core.Task, offset, renderbleNum, cursorPosition int) []*core.Task {
	w, h := v.screenClient.ScreenSize()

	limit := min(v.selectOffset+renderbleNum, len(tasks))

	renderedTasks := make([]*core.Task, 0, limit-offset)
	for y, t := range tasks[offset:limit] {
		name := fmt.Sprintf("%3d: %s", offset+y+1, t.Title)
		opts := []draw.Option{}
		if y == cursorPosition {
			opts = []draw.Option{
				draw.WithBackgroundColor(v.config.Color.SelectedLine),
				draw.WithForegroundColor(v.config.Color.Font),
			}
		}
		tw := runewidth.StringWidth(name)
		if d := w - tw; d > 0 {
			name += strings.Repeat(" ", d)
		}
		_ = draw.Sentence(v.screenClient.GetScreen(), 0, y, w, name, true, opts...)

		renderedTasks = append(renderedTasks, t)
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

	return renderedTasks
}

func (v *TaskView) getSelectRenderableHeight() int {
	_, h := v.screenClient.ScreenSize()
	if h <= 1 {
		return 0
	}

	return h - 1 // Subtract 1 for the footer
}
