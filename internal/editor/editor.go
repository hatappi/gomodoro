// Package editor manage EDITOR
package editor

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
)

// GetSliceText get slice text edited with editor
func GetSliceText(initialText string) ([]string, error) {
	tmpfile, err := ioutil.TempFile("", "gomodoro")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()
	defer func() {
		_ = tmpfile.Close()
	}()

	if _, err = tmpfile.Write([]byte(initialText)); err != nil {
		return nil, err
	}

	if err = openEditor(tmpfile.Name()); err != nil {
		return nil, err
	}

	f, err := os.Open(tmpfile.Name())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	ts := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ts = append(ts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}

func openEditor(filepath string) error {
	cmdName := "vi"
	if e := os.Getenv("EDITOR"); e != "" {
		cmdName = e
	}

	c := exec.Command(cmdName, filepath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
