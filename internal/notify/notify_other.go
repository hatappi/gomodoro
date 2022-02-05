//go:build !darwin
// +build !darwin

// Package notify notification
package notify

import (
	"golang.org/x/xerrors"
)

// Notify return unsupported error
func Notify(title, message string) error {
	return xerrors.New("unsupported notification")
}
