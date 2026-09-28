//go:build windows

package envprofile

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var setFileInformationByHandle = windows.NewLazySystemDLL("kernel32.dll").NewProc("SetFileInformationByHandle")

func openReadNoFollow(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		_ = windows.CloseHandle(handle)
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("%s: %s", DiagWriteWouldFollowLink, path)
	}
	return os.NewFile(uintptr(handle), path), nil
}

// renameManagedEntry replaces the directory entry through a handle to the
// source entry. MoveFileEx cannot replace an existing directory symlink on
// Windows; FileRenameInfo can replace the reparse-point entry without
// opening its target.
func renameManagedEntry(source, target string) error {
	name, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(name, windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	nameUTF16, err := windows.UTF16FromString(target)
	if err != nil {
		return err
	}
	const filenameOffset = 20 // FILE_RENAME_INFO.FileName offset on Windows.
	buffer := make([]byte, filenameOffset+len(nameUTF16)*2)
	buffer[0] = 1                                  // ReplaceIfExists = TRUE.
	fileNameLength := uint32(len(nameUTF16)-1) * 2 // exclude UTF-16 NUL.
	*(*uint32)(unsafe.Pointer(&buffer[16])) = fileNameLength
	for i, unit := range nameUTF16[:len(nameUTF16)-1] {
		*(*uint16)(unsafe.Pointer(&buffer[filenameOffset+i*2])) = unit
	}
	result, _, callErr := setFileInformationByHandle.Call(
		uintptr(handle), 3, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	runtime.KeepAlive(buffer)
	if result == 0 {
		if callErr != windows.ERROR_SUCCESS {
			return callErr
		}
		return windows.ERROR_INVALID_FUNCTION
	}
	return nil
}
