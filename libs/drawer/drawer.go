package drawer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var (
	strMin string
	strSec string
	firstX int
	firstY int
)

const (
	interval = 6
)

type Drawer struct {
	TaskName string
}

func NewDrawer(taskName string) *Drawer {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	return &Drawer{
		taskName,
	}
}

func (d *Drawer) Close() {
	termbox.HideCursor()
	termbox.Close()
}

func (d *Drawer) Draw(min, sec int, color termbox.Attribute) error {
	strMin = fmt.Sprintf("%02d", min)
	strSec = fmt.Sprintf("%02d", sec)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	width, height := termbox.Size()
	firstX = width/2 - 15
	firstY = height/2 - 3

	y := firstY - 3
	x := firstX

	// draw taskName
	for _, r := range []rune(d.TaskName) {
		termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorDefault)
		x += runewidth.RuneWidth(r)
	}

	// draw hour
	for _, i := range strings.Split(strMin, "") {
		num, err := strconv.Atoi(i)
		if err != nil {
			return err
		}
		arr := Num2StrArray(num)
		setCell(arr, color)
	}

	// draw separator
	setCell(SeparatorStrArray(), color)

	// draw min
	for _, i := range strings.Split(strSec, "") {
		num, err := strconv.Atoi(i)
		if err != nil {
			return err
		}
		arr := Num2StrArray(num)
		setCell(arr, color)
	}

	// draw timer
	termbox.Flush()

	return nil
}

func (d *Drawer) DrawMessage(msg string) {
	d.ClearMessage()
	msgWidth := len(msg)
	width, _ := termbox.Size()

	y := 2
	x := (width / 2) - (msgWidth / 2)

	for _, m := range []rune(msg) {
		termbox.SetCell(x, y, m, termbox.ColorWhite, termbox.ColorDefault)
		x += 1
	}
	termbox.Flush()
}

func (d *Drawer) ClearMessage() {
	width, _ := termbox.Size()
	for i := 0; i < width; i++ {
		termbox.SetCell(i, 2, ' ', termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
}

func setCell(arr []string, color termbox.Attribute) {
	x, y := firstX, firstY
	for _, txt := range arr {
		if txt == "\n" {
			y += 1
			x = firstX
			continue
		}
		if txt == "#" {
			termbox.SetCell(x, y, ' ', color, color)
		}
		x += 1
	}

	firstX = firstX + interval
}
