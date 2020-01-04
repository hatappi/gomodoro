// Package screen provide screen management for task
package screen

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/hatappi/gomodoro/screen/draw"
	runewidth "github.com/mattn/go-runewidth"
)

// Client represents management Task function
type Client interface {
	CreateTask() string
	SelectTask([]string) string
}

type clientImpl struct {
	screen tcell.Screen
}

// NewClient initilize Client
func NewClient(s tcell.Screen) Client {
	return &clientImpl{
		screen: s,
	}
}

// SelectTask gets task
func (c *clientImpl) SelectTask(tasks []string) string {
	var tasksWithIndex []string
	for i, t := range tasks {
		tasksWithIndex = append(tasksWithIndex, fmt.Sprintf("%3d. %s", i+1, t))
	}

	offset := 0
	i := 0
	for {
		w, h := c.screen.Size()
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
			_ = draw.Sentence(c.screen, 0, y, w, t, opts...)
		}

		ev := c.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return ""
			case tcell.KeyEnter:
				return tasks[offset+i]
			case tcell.KeyDown:
				if offset+i == len(tasks)-1 {
					continue
				}

				if i < h-1 {
					i++
				} else {
					c.screen.Clear()
					offset += h
					i = 0
				}
			case tcell.KeyUp:
				if offset+i <= 0 {
					continue
				}

				if i > 0 {
					i--
				} else {
					c.screen.Clear()
					offset -= h
					i = h - 1
				}
			case tcell.KeyRune:
				switch ev.Rune() {
				case rune(106): // j
					c.screen.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone))
				case rune(107): // k
					c.screen.PostEventWait(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone))
				case rune(110): // n
					c.screen.Clear()
					if t := c.CreateTask(); t != "" {
						return t
					}
				}
			}
		case *tcell.EventResize:
			// reset
			i = 0
			offset = 0

			c.screen.Sync()
		}
	}
}

func (c *clientImpl) CreateTask() string {
	newTaskName := []rune{}
	for {
		msg := fmt.Sprintf("Please Input New Task Name >%s", string(newTaskName))
		w, _ := c.screen.Size()
		c.screen.Clear()
		x := draw.Sentence(c.screen, 0, 0, w, msg)

		gl := ' '
		st := tcell.StyleDefault
		st = st.Background(tcell.ColorGreen)
		c.screen.SetCell(x, 0, st, gl)
		c.screen.Show()

		ev := c.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return ""
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if l := len(newTaskName); l > 0 {
					newTaskName = newTaskName[:l-1]
				}
			case tcell.KeyEnter:
				return string(newTaskName)
			case tcell.KeyRune:
				newTaskName = append(newTaskName, ev.Rune())
			}
		case *tcell.EventResize:
			c.screen.Sync()
		}
	}
}
