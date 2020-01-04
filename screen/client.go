package screen

import (
	"time"

	"github.com/gdamore/tcell"
)

// Client include related screen
type Client interface {
	GetScreen() tcell.Screen
	ScreenSize() (int, int)

	Clear()
	Finish()

	StartPollEvent()
	StopPollEvent()

	GetCancelChan() chan interface{}
	GetEnterChan() chan interface{}
	GetRuneChan() chan rune
}

type clientImpl struct {
	screen tcell.Screen

	cancelChan chan interface{}
	enterChan  chan interface{}
	runeChan   chan rune
}

// NewClient initilize Client
func NewClient(s tcell.Screen) Client {
	return &clientImpl{
		screen:     s,
		cancelChan: make(chan interface{}),
		enterChan:  make(chan interface{}),
		runeChan:   make(chan rune),
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
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					c.cancelChan <- struct{}{}
				case tcell.KeyEnter:
					c.enterChan <- struct{}{}
				case tcell.KeyRune:
					c.runeChan <- ev.Rune()
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
	c.screen.PostEventWait(&finishPollEvent{})
}

func (c *clientImpl) GetCancelChan() chan interface{} {
	return c.cancelChan
}

func (c *clientImpl) GetEnterChan() chan interface{} {
	return c.enterChan
}

func (c *clientImpl) GetRuneChan() chan rune {
	return c.runeChan
}
