package contextstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production entry points under test: EnsureState, EnsureGit, Exists,
// EntryDir, ContentHash.

// TestStateEntryIsContentKeyed checks the happy path: a state entry copies
// the tree, keys it by the content hash, and reuses it on the second call.
func TestStateEntryIsContentKeyed(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "agent-context.json"),
		[]byte(`{"schema_version": 1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	first, key, err := EnsureState(home, "context", "acme", source)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 64 || first != EntryDir(home, "context", "acme", key) {
		t.Fatalf("entry %q key %q", first, key)
	}
	if !Exists(home, "context", "acme", key) {
		t.Fatal("entry must exist after install")
	}
	second, key2, err := EnsureState(home, "context", "acme", source)
	if err != nil {
		t.Fatal(err)
	}
	if second != first || key2 != key {
		t.Fatal("identical source must reuse the entry")
	}
	if _, err := ContentHash(first); err != nil {
		t.Fatal(err)
	}
}

// TestStateEntryRejectsLinks narrows the archive-discipline gate: a symlink
// in a state tree must fail with profile_source_invalid. A mutant that
// copies through links must fail this test.
func TestStateEntryRejectsLinks(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "f.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("f.md", filepath.Join(source, "link.md")); err != nil {
		t.Skipf("this host cannot create symlinks: %v", err)
	}
	_, _, err := EnsureState(home, "context", "acme", source)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("symlink must fail as %s, got %v", DiagSourceInvalid, err)
	}
}

// TestStateEntryRejectsNestedGit narrows the same gate on nested .git
// entries: only a root-level .git is excluded.
func TestStateEntryRejectsNestedGit(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "sub", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err := EnsureState(home, "context", "acme", source)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("nested .git must fail as %s, got %v", DiagSourceInvalid, err)
	}
}

// TestStateEntryExcludesRootGit checks a root-level .git directory is
// skipped rather than copied.
func TestStateEntryExcludesRootGit(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".git", "HEAD"), []byte("ref\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, _, err := EnsureState(home, "context", "acme", source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(entry, ".git")); !os.IsNotExist(err) {
		t.Fatal("root .git must not be copied")
	}
}

// TestBadNameIsRejected checks the identifier gate on entry names.
func TestBadNameIsRejected(t *testing.T) {
	home := t.TempDir()
	if _, _, err := EnsureState(home, "context", "not a name", t.TempDir()); err == nil {
		t.Fatal("invalid name must fail")
	}
}
