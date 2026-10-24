//go:build windows

package managerlock

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

func prepareManagerHome(path string) error {
	resolved, missing, err := resolveExistingPrefix(path, "manager home")
	if err != nil {
		return err
	}
	path = resolved
	for index := len(missing) - 1; index >= 0; index-- {
		path = filepath.Join(path, missing[index])
	}
	for _, rel := range []string{"", "state", "state/locks", "state/locks/v1"} {
		if err := privatedir.Prepare(filepath.Join(path, filepath.FromSlash(rel))); err != nil {
			return err
		}
	}
	return nil
}

func makeLockDirectories(path string) error { return privatedir.MakeAll(path) }

func openManagerLock(path string, create bool) (*os.File, error) {
	if create {
		file, err := privatedir.Create(path, 0o600)
		if err == nil || !errors.Is(err, os.ErrExist) {
			return file, err
		}
	}
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// Replacement must remain observable to the exact handle/path identity
	// witness. A sharing denial must not hide a stale acquisition from tests
	// or substitute for AssertCurrent's ownership check.
	handle, err := windows.CreateFile(path16, windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		_ = windows.CloseHandle(handle)
		return nil, os.ErrInvalid
	}
	return file, nil
}
