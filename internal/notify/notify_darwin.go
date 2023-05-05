//go:build darwin
// +build darwin

// Package notify notification
package notify

import (
	"os/exec"
)

// Notify osx notification using osascript.
func Notify(title, message string) error {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}

	//nolint:gosec
	cmd := exec.Command(osa, "-e", `display notification "`+message+`" with title "`+title+`" sound name "Glass"`)
	return cmd.Run()
}
