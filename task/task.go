// Package task manage task
package task

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/hatappi/gomodoro/screen"
	"github.com/hatappi/gomodoro/screen/draw"
)

// Task represents task
type Task struct {
	Name string
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

// GetTask get task name
func GetTask(c screen.Client) string {
	var tasks Tasks
	for i := 0; i < 50; i++ {
		t := &Task{
			Name: fmt.Sprintf("Task %d", i),
		}
		tasks = append(tasks, t)
	}

	return selectTask(c, tasks.GetTaskNames())
}

func selectTask(c screen.Client, tasks []string) string {
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

		select {
		case <-c.GetCancelChan():
			return ""
		case <-c.GetEnterChan():
			return tasks[offset+i]
		case <-c.GetKeyDownChan():
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
		case <-c.GetKeyUpChan():
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
		case r := <-c.GetRuneChan():
			s := c.GetScreen()
			switch r {
			case rune(106): // j
				s.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone))
			case rune(107): // k
				s.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
			case rune(110): // n
				c.Clear()
				if t := createTask(c); t != "" {
					return t
				}
			}
		case <-c.GetResizeEventChan():
			// reset
			i = 0
			offset = 0
		}
	}
}

func createTask(c screen.Client) string {
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

		select {
		case <-c.GetCancelChan():
			return ""
		case <-c.GetEnterChan():
			return string(newTaskName)
		case <-c.GetDelChan():
			if l := len(newTaskName); l > 0 {
				newTaskName = newTaskName[:l-1]
			}
		case r := <-c.GetRuneChan():
			newTaskName = append(newTaskName, r)
		}
	}
}
