//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package envsnapshot

import (
	"errors"
	"os"
)

// No ACL/no-follow equivalence is claimed on unqualified platforms. Fail
// closed rather than derive executable semantics from a filename or Go's
// synthesized Windows permission bits. A future platform backend owns this.
func openDirectory(string) (*os.File, error) {
	return nil, errors.New("snapshot filesystem verification unavailable")
}
func openMember(*os.File, string, bool) (*os.File, error) {
	return nil, errors.New("snapshot filesystem verification unavailable")
}
func regularIndependent(os.FileInfo) bool  { return false }
func modeMatches(os.FileInfo, string) bool { return false }
func directoryMode(os.FileInfo) bool       { return false }
