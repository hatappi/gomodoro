// Package screen management timer screen
package screen

import (
	"time"

	"github.com/gdamore/tcell"
)

// Client include related screen
type Client interface {
	StartPollEvent()
	StopPollEvent()
	ScreenSize() (int, int)
	Clear()
	Finish()

	GetQuitChan() chan struct{}
	GetPauseChan() chan interface{}
	GetForceFinishChan() chan interface{}

	GetScreen() tcell.Screen
}

type clientImpl struct {
	screen tcell.Screen

	quit        chan struct{}
	pause       chan interface{}
	forceFinish chan interface{}

	pollEventStarted bool
}

// NewClient initilize Client
func NewClient(s tcell.Screen) (Client, error) {
	return &clientImpl{
		screen:      s,
		quit:        make(chan struct{}),
		pause:       make(chan interface{}),
		forceFinish: make(chan interface{}),
	}, nil
}

func (c *clientImpl) GetScreen() tcell.Screen {
	return c.screen
}

func (c *clientImpl) GetQuitChan() chan struct{} {
	return c.quit
}

func (c *clientImpl) GetPauseChan() chan interface{} {
	return c.pause
}

func (c *clientImpl) GetForceFinishChan() chan interface{} {
	return c.forceFinish
}

// Start screen
func (c *clientImpl) StartPollEvent() {
	if c.pollEventStarted {
		return
	}

	go func() {
		c.pollEventStarted = true
		for {
			ev := c.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					close(c.GetQuitChan())
					return
				case tcell.KeyEnter:
					c.pause <- struct{}{}
				case tcell.KeyRune:
					if ev.Rune() == rune(101) { // e
						c.forceFinish <- struct{}{}
					}
				}
			case *tcell.EventResize:
				c.screen.Sync()
			case *finishPollEvent:
				return
			}
		}
	}()
}

type finishPollEvent struct{}

func (fpe *finishPollEvent) When() time.Time {
	return time.Now()
}

func (c *clientImpl) StopPollEvent() {
	c.pollEventStarted = false
	c.screen.PostEventWait(&finishPollEvent{})
}

// ScreenSize get screen width and height
func (c *clientImpl) ScreenSize() (int, int) {
	return c.screen.Size()
}

// Clear screen
func (c *clientImpl) Clear() {
	c.screen.Clear()
}

// Finish screen
func (c *clientImpl) Finish() {
	c.screen.Fini()
}
