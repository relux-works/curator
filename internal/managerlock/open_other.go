//go:build !windows

package managerlock

import "os"

func prepareManagerHome(path string) error  { return os.MkdirAll(path, 0o700) }
func makeLockDirectories(path string) error { return os.MkdirAll(path, 0o700) }

func openManagerLock(path string, create bool) (*os.File, error) {
	flags := os.O_RDWR
	if create {
		flags |= os.O_CREATE
	}
	return os.OpenFile(path, flags, 0o600) // #nosec G304 -- derived canonical manager lock path.
}
