//go:build !windows

package buildsource

import (
	"os"
	"path/filepath"
	"testing"
)

// newFrozenRootCase replaces the validated root by renaming it aside while its
// handle is still open, which POSIX allows, and letting a same-content
// directory take the name it vacated.
func newFrozenRootCase(t *testing.T, parent string) frozenRootCase {
	t.Helper()
	root := filepath.Join(parent, "snapshot")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return frozenRootCase{root: root, replace: func(t *testing.T) {
		t.Helper()
		if err := os.Rename(root, filepath.Join(parent, "old")); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, root, "file", []byte("same"))
	}}
}

// writeInvalidProtocolPathEntry creates an entry whose name the protocol path
// rule refuses. A colon is an ordinary POSIX filename byte and an explicit
// PortableComponent rejection, so it names a file here rather than a stream.
func writeInvalidProtocolPathEntry(t *testing.T, root string) string {
	t.Helper()
	const name = "bad:name"
	writeTestFile(t, root, name, []byte("x"))
	return name
}
