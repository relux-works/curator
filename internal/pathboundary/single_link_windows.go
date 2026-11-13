//go:build windows

package pathboundary

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func singleLink(path string, observed os.FileInfo) error {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	// Read attributes only: proving structure must not require payload access.
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(handle), path)
	defer func() { _ = f.Close() }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return err
	}
	kind, err := windows.GetFileType(handle)
	if err != nil {
		return err
	}
	if kind != windows.FILE_TYPE_DISK || info.NumberOfLinks != 1 ||
		info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		return fmt.Errorf("snapshot payload must be a single-link regular file")
	}
	bound, err := f.Stat()
	if err != nil {
		return err
	}
	if !bound.Mode().IsRegular() || !os.SameFile(observed, bound) {
		return fmt.Errorf("snapshot file identity changed during link metadata inspection")
	}
	return nil
}
