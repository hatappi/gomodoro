// +build !darwin

package notify

import (
	"golang.org/x/xerrors"
)

func Notify(title, message string) error {
	return xerrors.New("unsupported notification")
}
