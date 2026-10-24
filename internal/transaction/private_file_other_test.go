//go:build !windows

package transaction

import (
	"os"
	"testing"
)

func broadenPrivateFileForTest(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
}
