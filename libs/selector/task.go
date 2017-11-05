package selector

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/nsf/termbox-go"
)

func Task(lines []string) (string, error) {
	clearend := "\x1b[0K"
	fillstart := ""
	fillend := "\x1b[0K\x1b[0m"

	fg := "30" // black
	bg := "44" // blue

	var qlines = []string{}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	out := os.Stdout

	// 終了シグナルが送られた場合にはこれを行う
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		<-sc
		out.Write([]byte("\x1b[?25h\x1b[0J"))
	}()

	out.Write([]byte("\x1b[?25l"))

	// ここで選択したものを表示している
	defer func() {
		e := recover()
		out.Write([]byte("\x1b[?25h\r\x1b[0J"))
		if e != nil {
			panic(e)
		}
	}()

	off := 0
	row := 0

	for {
		n := 0

		// render task list
		qlines = lines[off:]
		for i, line := range qlines {
			out.Write([]byte(fillstart))
			if off+i == row {
				out.Write([]byte("\x1b[" + fg + ";" + bg + "m" + line + fillend + "\r"))
			} else {
				out.Write([]byte(line + clearend + "\r"))
			}

			n++
			out.Write([]byte("\n"))
		}
		out.Write([]byte(fmt.Sprintf("\x1b[%dA", n)))

		// wait key input
		ev := termbox.PollEvent()
		switch ev.Key {
		case termbox.KeyCtrlC:
			return "", nil
		case termbox.KeyEsc:
			return "", nil
		case termbox.KeyEnter:
			return qlines[row], nil
		default:
			switch ev.Ch {
			case 106: // j
				if row < len(qlines)-1 {
					row++
				}
			case 107: // k
				if row > 0 {
					row--
				}
			case 110: // n
				return "", nil
			}
		}
	}
}
