//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package buildsource

import (
	"errors"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestRejectsSpecialFile(t *testing.T) {
	root := t.TempDir()
	if err := unix.Mkfifo(filepath.Join(root, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(root); !errors.Is(err, ErrInvalidSnapshot) {
		t.Fatalf("Validate error = %v", err)
	}
}
