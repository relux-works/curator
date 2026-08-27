//go:build !windows

package scopes

import (
	"io/fs"
	"os"
)

// isRedirect reports whether a scope member redirects the traversal somewhere
// maintenance did not prove, which on these platforms means a symbolic link or
// any other non-directory, non-regular object.
func isRedirect(info fs.FileInfo) bool {
	if info == nil {
		return true
	}
	return info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0
}
