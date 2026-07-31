//go:build !windows

package transaction

import "os"

// completeNamespaceIdentity has nothing left to do outside Windows: os.Stat and
// os.Lstat return the device and inode os.SameFile compares, so the identity a
// pass records is already complete — and already independent of the filesystem
// — the moment the stat returns.
func completeNamespaceIdentity(_ targetNamespacePath, info os.FileInfo) (os.FileInfo, error) {
	return info, nil
}
