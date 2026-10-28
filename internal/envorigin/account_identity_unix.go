//go:build unix

package envorigin

import "os"

// Process identity is independent of the injectable account-database lookup.
// These scalar reads never discover an account home or inspect credentials.
func processAccountIdentity() (uint64, uint64, error) {
	real, effective := os.Getuid(), os.Geteuid()
	if real < 0 || effective < 0 {
		return 0, 0, invalid("process account identity unavailable")
	}
	return uint64(real), uint64(effective), nil
}
