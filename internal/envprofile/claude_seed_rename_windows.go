//go:build windows

package envprofile

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Open the source entry, encode FILE_RENAME_INFO and resolve the procedure
// before the final comparison. Publication itself is one native handle rename,
// using the same reparse-point replacement semantics as renameManagedEntry.
func prepareClaudeSeedRename(source, target string) (func() error, func(), error) {
	linkError := func(err error) error {
		return &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
	}
	name, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return nil, nil, linkError(err)
	}
	nameUTF16, err := windows.UTF16FromString(target)
	if err != nil {
		return nil, nil, linkError(err)
	}
	if err := setFileInformationByHandle.Find(); err != nil {
		return nil, nil, linkError(err)
	}
	address := setFileInformationByHandle.Addr()
	handle, err := windows.CreateFile(name, windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, nil, linkError(err)
	}
	// Match FILE_RENAME_INFO's native alignment on both 32- and 64-bit Windows.
	type renameInfo struct {
		replace uint8
		root    windows.Handle
		length  uint32
		name    [1]uint16
	}
	const filenameOffset = int(unsafe.Offsetof(renameInfo{}.name))
	const lengthOffset = int(unsafe.Offsetof(renameInfo{}.length))
	buffer := make([]byte, filenameOffset+len(nameUTF16)*2)
	buffer[0] = 1 // ReplaceIfExists = TRUE.
	*(*uint32)(unsafe.Pointer(&buffer[lengthOffset])) = uint32(len(nameUTF16)-1) * 2
	for i, unit := range nameUTF16[:len(nameUTF16)-1] {
		*(*uint16)(unsafe.Pointer(&buffer[filenameOffset+i*2])) = unit
	}
	return func() error {
		result, _, callErr := syscall.SyscallN(address,
			uintptr(handle), 3, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
		runtime.KeepAlive(buffer)
		if result == 0 {
			if callErr == windows.ERROR_SUCCESS {
				callErr = windows.ERROR_INVALID_FUNCTION
			}
			return linkError(callErr)
		}
		return nil
	}, func() { _ = windows.CloseHandle(handle) }, nil
}
