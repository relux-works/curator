//go:build windows

package snapcache

import (
	"fmt"
	"golang.org/x/sys/windows"
	"io/fs"
	"os"
	"syscall"
)

// allocatedSize reports the logical length: the file information Go exposes
// on Windows carries no allocation size, and section 10.1 reports the logical
// figure where the platform reports no allocation.
func allocatedSize(info fs.FileInfo) int64 { return info.Size() }

// fileIdentity uses the volume serial and file index from a nofollow handle.
// Allocation fallback does not waive the per-entry hard-link deduplication.
func fileIdentity(path string, info fs.FileInfo) (string, bool, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", false, err
	}
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return "", false, err
	}
	file := os.NewFile(uintptr(handle), path)
	defer func() { _ = file.Close() }()
	var data windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &data); err != nil {
		return "", false, err
	}
	if data.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		return "", false, fmt.Errorf("snapshot file %s is no longer regular", path)
	}
	opened, err := file.Stat()
	if err != nil {
		return "", false, err
	}
	if !os.SameFile(info, opened) {
		return "", false, fmt.Errorf("snapshot file %s changed while measuring", path)
	}
	return fmt.Sprintf("%d:%d:%d", data.VolumeSerialNumber, data.FileIndexHigh, data.FileIndexLow), true, nil
}

// isReparsePoint reports a directory junction or other reparse point, which
// retention never traverses or removes through.
func isReparsePoint(info fs.FileInfo) bool {
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && data != nil && data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
