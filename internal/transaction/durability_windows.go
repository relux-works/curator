//go:build windows

package transaction

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func syncRegular(path string) error {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return err
	}
	return errors.Join(windows.FlushFileBuffers(handle), windows.CloseHandle(handle))
}

func journalMode() os.FileMode {
	// Windows persists Go's writable permission requests as 0666. Chmod(0600)
	// therefore clears FILE_ATTRIBUTE_READONLY but cannot persist owner-only
	// permission bits; os.Lstat reports the resulting regular file as 0666.
	return 0o666
}

func syncDirectory(path string) error {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(path16, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	err = windows.FlushFileBuffers(handle)
	if errors.Is(err, windows.ERROR_INVALID_HANDLE) || errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		// Directory FlushFileBuffers is filesystem-dependent. Every namespace
		// mutation below also uses MOVEFILE_WRITE_THROUGH, which is the Windows
		// durability primitive for the directory entry itself.
		return nil
	}
	return err
}

func durableRename(from, to string) error {
	from16, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	to16, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	if err := windows.MoveFileEx(from16, to16, windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(to))
}
