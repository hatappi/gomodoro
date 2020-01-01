package screen

import (
	"github.com/gdamore/tcell"
)

// Client include related screen
type Client interface {
	Start()
	ScreenSize() (int, int)
	Clear()
	Finish()

	DrawSentence(x, y, maxWidth int, str string)
	DrawTimer(x, y, mag, min, sec int)

	GetQuitChan() chan struct{}
}

type clientImpl struct {
	screen tcell.Screen

	quit chan struct{}
}

// NewClient initilize Client
func NewClient() (Client, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).Background(tcell.ColorWhite))

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	return &clientImpl{
		screen: s,
		quit:   make(chan struct{}),
	}, nil
}

func (c *clientImpl) GetQuitChan() chan struct{} {
	return c.quit
}

// Start screen
func (c *clientImpl) Start() {
	go func() {
		for {
			ev := c.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(c.GetQuitChan())
					return
				}
			case *tcell.EventResize:
				c.screen.Sync()
			}
		}
	}()
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
