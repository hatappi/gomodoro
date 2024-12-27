package screen

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/hatappi/go-kit/log"
)

// Event screen event.
type Event interface{}

// EventKeyUp press keyup event.
type EventKeyUp struct{}

// EventKeyDown press keydown event.
type EventKeyDown struct{}

// EventCancel cancel event.
type EventCancel struct{}

// EventEnter press Enter event.
type EventEnter struct{}

// EventDelete delete event.
type EventDelete struct{}

// EventRune press rune.
type EventRune rune

// EventScreenResize resize screen Event.
type EventScreenResize struct{}

// Client include related screen.
type Client interface {
	GetScreen() tcell.Screen
	ScreenSize() (int, int)

	Clear()
	Finish()

	StartPollEvent(ctx context.Context)
	StopPollEvent() error

	GetEventChan() chan Event
}

// IClient implements Client interface.
type IClient struct {
	screen tcell.Screen

	eventChan chan Event
}

// NewClient initilize Client.
func NewClient(s tcell.Screen) *IClient {
	return &IClient{
		screen:    s,
		eventChan: make(chan Event),
	}
}

// GetScreen gets screen.
//
//nolint:ireturn
func (c *IClient) GetScreen() tcell.Screen {
	return c.screen
}

// ScreenSize gets screen size.
func (c *IClient) ScreenSize() (int, int) {
	return c.screen.Size()
}

// Clear clears screen.
func (c *IClient) Clear() {
	c.screen.Clear()
}

// Finish finishes screen.
func (c *IClient) Finish() {
	c.screen.Fini()
}

// StartPollEvent starts polling event on goroutine.
func (c *IClient) StartPollEvent(ctx context.Context) {
	go func() {
		for {
			ev := c.screen.PollEvent()
			log.FromContext(ctx).V(1).Info("receive event", "event", ev)
			switch ev := ev.(type) {
			case *tcell.EventKey:
				//nolint:exhaustive
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

// GetEventChan gets an event connection.
func (c *IClient) GetEventChan() chan Event {
	return c.eventChan
}

type finishPollEvent struct{}

func (fpe *finishPollEvent) When() time.Time {
	return time.Now()
}

// StopPollEvent stops polling event of screen.
func (c *IClient) StopPollEvent() error {
	return c.screen.PostEvent(&finishPollEvent{})
}
