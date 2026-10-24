//go:build !windows

package gitops

import "testing"

func cloneTestRoot(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
