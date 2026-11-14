//go:build windows

package envprofile

import (
	"golang.org/x/sys/windows"
	"os"
)

// File.Stat binds volume/file index now. Lstat defers the ID lookup until
// SameFile and would compare a rebuilt entry with itself at its reused path.
func storeVerificationFileInfo(path string) (os.FileInfo, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(handle), path)
	defer func() { _ = f.Close() }()
	return f.Stat()
}
