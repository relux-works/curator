//go:build unix

package closureexec

import "io/fs"

// blobModeMatchesClass reports whether a protected blob's permission bits are
// exactly the ones its output class publishes: owner read-and-execute for a
// native executable, owner read-only for everything else. On Unix the bits are
// the contract itself.
func blobModeMatchesClass(info fs.FileInfo, executable bool) bool {
	if executable {
		return info.Mode().Perm() == 0o500
	}
	return info.Mode().Perm() == 0o400
}
