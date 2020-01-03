package screen

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

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
			opts := []DrawOption{}
			if y == i {
				opts = []DrawOption{
					WithBackgroundColor(tcell.ColorBlue),
				}
			}
			tw := runewidth.StringWidth(t)
			if d := w - tw; d > 0 {
				t += strings.Repeat(" ", d)
			}
			_ = c.DrawSentence(0, y, w, t, opts...)
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
		x := c.DrawSentence(0, 0, w, msg)

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

func (c *clientImpl) DrawSentence(x, y, width int, str string, opts ...DrawOption) int {
	style := tcell.StyleDefault
	for _, opt := range opts {
		style = opt(style)
	}

	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				c.screen.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				c.screen.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		c.screen.SetContent(x+i, y, deferred[0], deferred[1:], style)
	}

	c.screen.Show()

	return x + i + dwidth
}
