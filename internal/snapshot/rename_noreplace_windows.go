//go:build windows

package snapshot

import "golang.org/x/sys/windows"

func renameNoReplace(from, to string) error {
	fromPtr, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	toPtr, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFile(fromPtr, toPtr)
}
