package errors

import (
	"golang.org/x/xerrors"
)

var (
	ErrCancel      = xerrors.New("cancel")
	ErrScreenSmall = xerrors.New("screen is small")
)
