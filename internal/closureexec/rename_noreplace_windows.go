//go:build windows

package closureexec

import "golang.org/x/sys/windows"

func renameTreeNoReplace(source, target string) error {
	sourcePtr, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	targetPtr, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	return windows.MoveFile(sourcePtr, targetPtr)
}
