//go:build windows

package scopes

import (
	"io/fs"
	"os"
	"syscall"
)

// isRedirect reports whether a scope member is a reparse point of any kind.
//
// Go maps some reparse points onto ModeSymlink and others onto ModeIrregular,
// and the exact mapping has moved between releases, so the attribute itself is
// checked as well: a junction that read as a plain directory would let a
// traversal be redirected away from the markers that hold live build keys.
func isRedirect(info fs.FileInfo) bool {
	if info == nil {
		return true
	}
	if info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
		return true
	}
	attributes, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return false
	}
	return attributes.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
