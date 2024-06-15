//go:build !darwin
// +build !darwin

// Package notify notification
package notify

import (
	"errors"
)

// Notify return unsupported error.
func Notify(_, _ string) error {
	return errors.New("unsupported notification")
}
