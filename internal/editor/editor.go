// Package editor manage EDITOR
package editor

import (
	"os"
	"os/exec"
	"strings"
)

// ContentsByLine gets contents edited with editor by line.
func ContentsByLine(initialText string) ([]string, error) {
	tmpfile, err := os.CreateTemp("", "gomodoro")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()
	defer func() {
		_ = tmpfile.Close()
	}()

	if _, err = tmpfile.WriteString(initialText); err != nil {
		return nil, err
	}

	if err = edit(tmpfile.Name()); err != nil {
		return nil, err
	}

	b, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return strings.Split(string(b), "\n"), nil
}

func edit(filepath string) error {
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
