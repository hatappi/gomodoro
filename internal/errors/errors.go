// Package errors represents gomodoro errors
package errors

import (
	"errors"
)

var (
	// ErrCancel cancel error.
	ErrCancel = errors.New("cancel")
	// ErrScreenSmall screen is small error.
	ErrScreenSmall = errors.New("screen is small")
)
