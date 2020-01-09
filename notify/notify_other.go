// +build !darwin

package notify

import (
	"os/exec"
)

func Notify(title, message string) error {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}

	cmd := exec.Command(osa, "-e", `display notification "`+message+`" with title "`+title+`" sound name "Glass"`)
	return cmd.Run()
}
