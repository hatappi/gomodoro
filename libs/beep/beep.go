package beep

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"github.com/hatappi/gomodoro/libs/assets"
)

func Beep() error {
	// go-bindata
	b, err := assets.Asset("assets/sounds/bell.mp3")
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	c := ioutil.NopCloser(r)
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
