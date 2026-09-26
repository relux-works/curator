package stateread

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestReadsDistinguishAbsentFromBlockedParent proves that a missing leaf in
// a traversable directory is absence, while read/stat failures beneath a
// regular-file ancestor remain unreadable on every platform.
func TestReadsDistinguishAbsentFromBlockedParent(t *testing.T) {
	root := t.TempDir()
	absent := filepath.Join(root, "absent")
	if got, err := ReadFile(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadFile(absent) = (%+v, %v), want absent", got, err)
	}
	if got, err := ReadDir(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadDir(absent) = (%+v, %v), want absent", got, err)
	}
	if got, err := Lstat(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("Lstat(absent) = (%+v, %v), want absent", got, err)
	}

	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(blocker, "child")
	// A Windows ERROR_PATH_NOT_FOUND for a child beneath a regular file is
	// equivalent to os.ErrNotExist at this classifier boundary. Keep that
	// shape explicit on every runner so narrowing mutants are portable.
	if missingPath(blocked, os.ErrNotExist) {
		t.Fatalf("missingPath(%q, os.ErrNotExist) = true below an existing file", blocked)
	}
	if !missingPath(absent, os.ErrNotExist) {
		t.Fatal("missingPath for a missing leaf under a traversable directory = false, want true")
	}
	checks := []struct {
		name string
		read func() (Kind, error)
	}{
		{"read-file", func() (Kind, error) { result, err := ReadFile(blocked); return result.Kind, err }},
		{"read-dir", func() (Kind, error) { result, err := ReadDir(blocked); return result.Kind, err }},
		{"stat", func() (Kind, error) { result, err := Stat(blocked); return result.Kind, err }},
		{"lstat", func() (Kind, error) { result, err := Lstat(blocked); return result.Kind, err }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			kind, err := check.read()
			var stateErr *Error
			if kind != KindUnreadable || !errors.As(err, &stateErr) || stateErr.Kind != KindUnreadable || stateErr.Path != blocked {
				t.Fatalf("read below regular-file parent = (%q, %v), want typed unreadable for %s", kind, err, blocked)
			}
		})
	}
}
