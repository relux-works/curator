//go:build windows

package contextstore

import "os"

// hardLinked is not observable through os.FileInfo on Windows; the archive
// discipline for hard links is enforced by the extraction path instead.
func hardLinked(os.FileInfo) bool { return false }
