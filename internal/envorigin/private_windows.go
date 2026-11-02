//go:build windows

package envorigin

import (
	"os"

	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

// Inspect the exact handle without following a reparse point. FileInfo does
// not expose a Windows link count; GetFileInformationByHandle does. Before,
// opened and after identities are independently pinned by readOwnedRecord.
func singleLink(path string, expected os.FileInfo) bool {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	var metadata windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &metadata); err != nil || metadata.NumberOfLinks != 1 || metadata.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		return false
	}
	// Stat on a duplicate handle compares its file index with the caller's pin.
	var duplicate windows.Handle
	process := windows.CurrentProcess()
	if err = windows.DuplicateHandle(process, handle, process, &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return false
	}
	file := os.NewFile(uintptr(duplicate), path)
	if file == nil {
		windows.CloseHandle(duplicate)
		return false
	}
	defer file.Close()
	actual, err := file.Stat()
	return err == nil && os.SameFile(expected, actual)
}
func privateMode(path string, _ os.FileInfo) bool { return privatedir.ValidateFile(path) == nil }
