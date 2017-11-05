package notification

import (
	"fmt"
	"os/exec"
	"runtime"
)

func NotifyDesktop(title, text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		code := fmt.Sprintf(`display notification "%s" with title "%s"`, text, title)
		cmd = exec.Command("osascript", "-e", code)
	case "linux":
		cmd = exec.Command("notify-send", "-i", title, text)
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
