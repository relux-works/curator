//go:build windows

package envprofile

import "golang.org/x/sys/windows"

// MSYS path conversion first opens the path for attributes/security, closes
// that handle and then opens it for data. This reproduces that access shape
// without an installed POSIX runtime; real SSH controls remain required.
func signaturePipeMetadataProbe(path string) error {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(name,
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES|windows.FILE_READ_EA,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	return windows.CloseHandle(handle)
}
