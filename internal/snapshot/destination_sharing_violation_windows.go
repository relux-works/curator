//go:build windows

package snapshot

import (
	"errors"

	"golang.org/x/sys/windows"
)

func isDestinationSharingViolation(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}
