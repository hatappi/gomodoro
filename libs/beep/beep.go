package beep

import (
	"io"
	"io/ioutil"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	_ "github.com/hatappi/gomodoro/libs/assets/statik"
	"github.com/rakyll/statik/fs"
)

func Beep() error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	b, err := statikFS.Open("/sounds/bell.mp3")
	if err != nil {
		return err
	}

	c := ioutil.NopCloser(b)
	d, err := mp3.NewDecoder(c)
	if err != nil {
		return err
	}
	defer d.Close()

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := io.Copy(p, d); err != nil {
		return err
	}

	return nil
}
