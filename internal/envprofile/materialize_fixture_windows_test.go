//go:build windows

package envprofile

import (
	"errors"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// MoveFileEx cannot plant replacement of an open target. FileRenameInformationEx
// with POSIX semantics retains existing handles while replacing the named file.
// This exercises identity drift without changing production file operations.
func replaceCapturedFixture(t *testing.T, source, target string) error {
	t.Helper()
	name, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(name, windows.DELETE|windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	nameUTF16, err := windows.UTF16FromString(`\??\` + target)
	if err != nil {
		return err
	}
	type renameInformation struct {
		Flags          uint32
		RootDirectory  windows.Handle
		FileNameLength uint32
		FileName       [1]uint16
	}
	var layout renameInformation
	nameLength := len(nameUTF16) - 1
	buffer := make([]byte, int(unsafe.Offsetof(layout.FileName))+nameLength*2)
	info := (*renameInformation)(unsafe.Pointer(&buffer[0]))
	info.Flags = windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS | windows.FILE_RENAME_IGNORE_READONLY_ATTRIBUTE
	info.FileNameLength = uint32(nameLength * 2)
	copy(unsafe.Slice(&info.FileName[0], nameLength), nameUTF16[:nameLength])
	var iosb windows.IO_STATUS_BLOCK
	const fileRenameInformationEx = 65
	err = windows.NtSetInformationFile(handle, &iosb, &buffer[0], uint32(len(buffer)), fileRenameInformationEx)
	var status windows.NTStatus
	if errors.As(err, &status) && status.Errno() == windows.ERROR_SHARING_VIOLATION {
		t.Skip("this host cannot create replacement of an open capture object: sharing violation")
	}
	return err
}
