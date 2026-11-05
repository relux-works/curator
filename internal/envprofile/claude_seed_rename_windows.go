//go:build windows

package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var claudeSeedMoveFileEx = windows.NewLazySystemDLL("kernel32.dll").NewProc("MoveFileExW")

// MoveFileExW can replace this validated regular-file seed without preparing a
// source handle. Resolve paths, encode UTF-16 and load the API before the final
// comparison. No copy/delete, directory fallback or sharing-violation retry is
// allowed after that comparison.
func prepareClaudeSeedRename(source, target string) (func() error, error) {
	from, err := claudeSeedRenamePath(source)
	if err != nil {
		return nil, err
	}
	to, err := claudeSeedRenamePath(target)
	if err != nil {
		return nil, err
	}
	if err := claudeSeedMoveFileEx.Find(); err != nil {
		return nil, err
	}
	return func() error {
		result, _, callErr := claudeSeedMoveFileEx.Call(
			uintptr(unsafe.Pointer(from)), uintptr(unsafe.Pointer(to)), windows.MOVEFILE_REPLACE_EXISTING)
		runtime.KeepAlive(from)
		runtime.KeepAlive(to)
		if result == 0 {
			if callErr == windows.ERROR_SUCCESS {
				callErr = windows.ERROR_INVALID_FUNCTION
			}
			return &os.LinkError{Op: "rename", Old: source, New: target, Err: callErr}
		}
		return nil
	}, nil
}

func claudeSeedRenamePath(path string) (*uint16, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// MoveFileExW supports extended drive and UNC paths. Abs also cleans dot
	// components before the extended prefix disables Win32 path parsing.
	if !strings.HasPrefix(path, `\\?\`) {
		if strings.HasPrefix(path, `\\`) {
			path = `\\?\UNC\` + path[2:]
		} else {
			path = `\\?\` + path
		}
	}
	return windows.UTF16PtrFromString(path)
}
