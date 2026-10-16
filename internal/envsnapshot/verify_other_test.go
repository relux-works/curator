//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package envsnapshot

import "testing"

// Unqualified filesystem platforms refuse instead of claiming tree integrity.
func TestUnqualifiedPlatformRefuses(t *testing.T) {
	_, err := Verify(t.TempDir(), "")
	requireReason(t, err, "tree-unreadable")
}
