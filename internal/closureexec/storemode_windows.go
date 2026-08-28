//go:build windows

package closureexec

import "io/fs"

// blobModeMatchesClass reports whether a protected blob carries the read-only
// shape on Windows. Windows synthesizes permission bits from the read-only
// attribute and has no execute bit, so the executable class is not expressible
// through the mode; the immutability half of the contract — a regular file the
// write attribute cannot touch — is what the platform can prove, and the
// class itself remains bound by the observation record's content digest.
func blobModeMatchesClass(info fs.FileInfo, _ bool) bool {
	return info.Mode().IsRegular() && info.Mode().Perm()&0o222 == 0
}
