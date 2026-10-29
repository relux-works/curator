//go:build !unix && !windows

package envorigin

import "os"

// Authority-file ownership/protection transport is not yet qualified here.
// Refuse instead of interpreting POSIX permission bits as a Windows ACL proof.
func singleLink(string, os.FileInfo) bool  { return false }
func privateMode(string, os.FileInfo) bool { return false }
