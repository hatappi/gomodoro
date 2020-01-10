// Package errors represents gomodoro errors
package errors

import (
	"golang.org/x/xerrors"
)

var (
	// ErrCancel cancel error
	ErrCancel = xerrors.New("cancel")
	// ErrScreenSmall screen is small error
	ErrScreenSmall = xerrors.New("screen is small")
)
