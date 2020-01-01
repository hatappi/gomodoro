package screen

import (
	"github.com/gdamore/tcell"
)

type Client struct {
	screen tcell.Screen

	Quit chan struct{}
}

func NewClient() (*Client, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).Background(tcell.ColorWhite))

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	return &Client{
		screen: s,
		Quit:   make(chan struct{}),
	}, nil
}

func (c *Client) Start() {
	go func() {
		for {
			ev := c.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(c.Quit)
					return
				}
			case *tcell.EventResize:
				c.screen.Sync()
			}
		}
	}()
}

func (c *Client) ScreenSize() (int, int) {
	return c.screen.Size()
}

func (c *Client) Clear() {
	c.screen.Clear()
}

func (c *Client) Finish() {
	c.screen.Fini()
}
