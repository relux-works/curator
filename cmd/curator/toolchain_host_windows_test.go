//go:build windows

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// fileAttributeTagInfo is FILE_ATTRIBUTE_TAG_INFO, which x/sys/windows names
// the class constant for but declares no struct for.
type fileAttributeTagInfo struct {
	FileAttributes uint32
	ReparseTag     uint32
}

// platformPathFact names the Windows attributes and reparse tag of a path,
// which is what decides how os.Lstat classified it. Go reports
// IO_REPARSE_TAG_SYMLINK and IO_REPARSE_TAG_MOUNT_POINT as ModeSymlink and any
// other tag as ModeIrregular without ModeDir, so a directory carrying an
// unrecognized tag stops satisfying a mode-based directory check while still
// serving every open and traversal normally. When the go-v1 boundary refuses a
// host GOROOT, this is the fact that says which of those it hit.
func platformPathFact(path string) string {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Sprintf(" reparse=<encode error: %v>", err)
	}
	var data windows.Win32FileAttributeData
	if err := windows.GetFileAttributesEx(path16, windows.GetFileExInfoStandard,
		(*byte)(unsafe.Pointer(&data))); err != nil {
		return fmt.Sprintf(" attributes=<err %v>", err)
	}
	report := fmt.Sprintf(" attributes=%#08x directory=%t reparsePoint=%t",
		data.FileAttributes,
		data.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0,
		data.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0)

	handle, err := windows.CreateFile(
		path16,
		windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return report + fmt.Sprintf(" open=<err %v>", err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var tag fileAttributeTagInfo
	if err := windows.GetFileInformationByHandleEx(handle, windows.FileAttributeTagInfo,
		(*byte)(unsafe.Pointer(&tag)), uint32(unsafe.Sizeof(tag))); err != nil {
		return report + fmt.Sprintf(" tag=<err %v>", err)
	}
	return report + fmt.Sprintf(" handleAttributes=%#08x reparseTag=%#08x", tag.FileAttributes, tag.ReparseTag)
}
