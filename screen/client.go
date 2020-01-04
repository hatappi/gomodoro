package screen

import (
	"time"

	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/logger"
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
	GetDelChan() chan interface{}
	GetRuneChan() chan rune
	GetKeyDownChan() chan interface{}
	GetKeyUpChan() chan interface{}
	GetResizeEventChan() chan interface{}
}

type clientImpl struct {
	screen tcell.Screen

	cancelChan      chan interface{}
	enterChan       chan interface{}
	delChan         chan interface{}
	runeChan        chan rune
	keyDownChan     chan interface{}
	keyUpChan       chan interface{}
	resizeEventChan chan interface{}
}

// NewClient initilize Client
func NewClient(s tcell.Screen) Client {
	return &clientImpl{
		screen:          s,
		cancelChan:      make(chan interface{}),
		enterChan:       make(chan interface{}),
		delChan:         make(chan interface{}),
		runeChan:        make(chan rune),
		keyDownChan:     make(chan interface{}),
		keyUpChan:       make(chan interface{}),
		resizeEventChan: make(chan interface{}),
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
			logger.Infof("%+v", ev)
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					c.cancelChan <- struct{}{}
				case tcell.KeyEnter:
					c.enterChan <- struct{}{}
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					c.delChan <- struct{}{}
				case tcell.KeyDown:
					c.keyDownChan <- struct{}{}
				case tcell.KeyUp:
					c.keyUpChan <- struct{}{}
				case tcell.KeyRune:
					c.runeChan <- ev.Rune()
				}
			case *tcell.EventResize:
				c.screen.Sync()
				c.resizeEventChan <- struct{}{}
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

func (c *clientImpl) GetDelChan() chan interface{} {
	return c.delChan
}

func (c *clientImpl) GetRuneChan() chan rune {
	return c.runeChan
}

func (c *clientImpl) GetKeyDownChan() chan interface{} {
	return c.keyDownChan
}

func (c *clientImpl) GetKeyUpChan() chan interface{} {
	return c.keyUpChan
}

func (c *clientImpl) GetResizeEventChan() chan interface{} {
	return c.resizeEventChan
}
