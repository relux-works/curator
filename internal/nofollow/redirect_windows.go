//go:build windows

package nofollow

import (
	"io/fs"
	"syscall"
)

func redirect(info fs.FileInfo) bool {
	if info.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 {
		return true
	}
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && data != nil && data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
