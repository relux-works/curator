//go:build windows

package skillintegrity

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// os.FileInfo does not expose the Windows hardlink count. Inspect the handle
// without traversing reparse points. The caller must supply handle-backed
// identity if it needs to compare the earlier object; path-backed FileInfo
// loads identity lazily on Windows and is not an earlier identity receipt.
func singleLink(path string, info os.FileInfo) bool {
	f, err := openCaptureFile(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	observed, err := f.Stat()
	return err == nil && os.SameFile(info, observed)
}

func openCaptureFile(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	kind, kindErr := windows.GetFileType(handle)
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil || kindErr != nil || kind != windows.FILE_TYPE_DISK || info.NumberOfLinks != 1 || info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		_ = windows.CloseHandle(handle)
		return nil, errors.New("skill_integrity_boundary_invalid: payload must be a single-link regular file")
	}
	return os.NewFile(uintptr(handle), path), nil
}
