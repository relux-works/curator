//go:build windows

package linkcap

import (
	"errors"
	"golang.org/x/sys/windows"
)

// Unavailable classifies platform errors that permit a declared projection-copy alternative.
func Unavailable(err error) bool {
	return errors.Is(err, windows.ERROR_NOT_SUPPORTED) || errors.Is(err, windows.ERROR_PRIVILEGE_NOT_HELD)
}
