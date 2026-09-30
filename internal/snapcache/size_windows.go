//go:build windows

package snapcache

import (
	"io/fs"
	"syscall"
)

// allocatedSize reports the logical length: the file information Go exposes
// on Windows carries no allocation size, and section 10.1 reports the logical
// figure where the platform reports no allocation.
func allocatedSize(info fs.FileInfo) int64 { return info.Size() }

// fileIdentity is not derived on Windows: FileInfo from a directory walk
// carries no volume serial and file index, so hard links inside one entry
// are counted once per name there.
func fileIdentity(fs.FileInfo) (string, bool) { return "", false }

// isReparsePoint reports a directory junction or other reparse point, which
// retention never traverses or removes through.
func isReparsePoint(info fs.FileInfo) bool {
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && data != nil && data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
