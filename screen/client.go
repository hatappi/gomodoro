package screen

import (
	"time"

	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/logger"
)

// Event screen event
type Event interface{}

// EventKeyUp press keyup event
type EventKeyUp struct{}

// EventKeyDown press keydown event
type EventKeyDown struct{}

// EventCancel cancel event
type EventCancel struct{}

// EventEnter press Enter event
type EventEnter struct{}

// EventDelete delete event
type EventDelete struct{}

// EventRune press rune
type EventRune rune

// EventScreenResize resize screen Event
type EventScreenResize struct{}

// Client include related screen
type Client interface {
	GetScreen() tcell.Screen
	ScreenSize() (int, int)

	Clear()
	Finish()

	StartPollEvent()
	StopPollEvent()

	GetEventChan() chan Event
}

type clientImpl struct {
	screen tcell.Screen

	eventChan chan Event
}

// NewClient initilize Client
func NewClient(s tcell.Screen) Client {
	return &clientImpl{
		screen:    s,
		eventChan: make(chan Event),
	}
}

func (c *clientImpl) GetScreen() tcell.Screen {
	return c.screen
}

func (c *clientImpl) ScreenSize() (int, int) {
	return c.screen.Size()
}

func (c *clientImpl) Clear() {
	c.screen.Clear()
}

func (c *clientImpl) Finish() {
	c.screen.Fini()
}

func (c *clientImpl) StartPollEvent() {
	go func() {
		for {
			ev := c.screen.PollEvent()
			logger.Debugf("event is %+v", ev)
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					c.eventChan <- EventCancel{}
				case tcell.KeyEnter:
					c.eventChan <- EventEnter{}
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					c.eventChan <- EventDelete{}
				case tcell.KeyDown:
					c.eventChan <- EventKeyDown{}
				case tcell.KeyUp:
					c.eventChan <- EventKeyUp{}
				case tcell.KeyRune:
					c.eventChan <- EventRune(ev.Rune())
				}
			case *tcell.EventResize:
				c.screen.Sync()
				c.eventChan <- EventScreenResize{}
			case *finishPollEvent:
				return
			}
		}
	}()
}

func (c *clientImpl) GetEventChan() chan Event {
	return c.eventChan
}

type finishPollEvent struct{}

func (fpe *finishPollEvent) When() time.Time {
	return time.Now()
}

func (c *clientImpl) StopPollEvent() {
	c.screen.PostEventWait(&finishPollEvent{})
}
