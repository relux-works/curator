//go:build windows

package pathboundary

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// Lstat may defer Windows file-ID loading until os.SameFile, which could then
// observe a replacement. File.Stat obtains the ID from this no-follow handle
// immediately, before the later link-count proof and identity comparison.
func snapshotFileInfo(path string, _ os.FileInfo) (os.FileInfo, error) {
	f, err := openSnapshotMetadata(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}

func openSnapshotMetadata(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// Read attributes only: proving structure must not require payload access.
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(handle), path)
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		_ = f.Close()
		return nil, err
	}
	kind, err := windows.GetFileType(handle)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if kind != windows.FILE_TYPE_DISK ||
		info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		_ = f.Close()
		return nil, fmt.Errorf("snapshot payload must be a regular disk file without reparse points")
	}
	return f, nil
}

func singleLink(path string, observed os.FileInfo) error {
	f, err := openSnapshotMetadata(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		return err
	}
	if info.NumberOfLinks != 1 {
		return fmt.Errorf("regular file must have exactly one link (observed %d)", info.NumberOfLinks)
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
