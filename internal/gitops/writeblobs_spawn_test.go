package gitops

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

// TestWriteBlobsSpawnFailureIsWrapped pins the writeBlobs spawn error to the
// operation-context shape the rest of the package uses: a failed
// `git cat-file --batch` start reports the operation and repository like the
// Wait path does, instead of surfacing the bare fork/exec text. The injected
// git is the shape the hosted gate observed (BUG-260920-3vfwch): a
// PATH-resolvable 0755 script whose shebang interpreter is 0644, so the
// kernel refuses the exec with EACCES.
func TestWriteBlobsSpawnFailureIsWrapped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	shimDir := t.TempDir()
	interp := filepath.Join(shimDir, "interp")
	if err := os.WriteFile(interp, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(interp, 0o644); err != nil {
		t.Fatal(err)
	}
	git := filepath.Join(shimDir, "git")
	if err := os.WriteFile(git, []byte("#!"+interp+"\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Sequential by construction: it rewrites PATH for the process, which
	// t.Setenv forbids under a parallel ancestor.
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	entries := []treeEntry{{
		mode: "100644", kind: "blob",
		oid:    strings.Repeat("a", 40),
		size:   2,
		path:   "SKILL.md",
		target: filepath.Join(t.TempDir(), "SKILL.md"),
	}}
	err := writeBlobs(filepath.Join(t.TempDir(), "repo"), entries)
	if err == nil {
		t.Fatal("writeBlobs with a non-executable git must fail")
	}
	if !strings.Contains(err.Error(), "git cat-file --batch failed") {
		t.Fatalf("err = %q, want the wrapped operation context", err)
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("err = %q, want the kernel refusal preserved", err)
	}
	if !errors.Is(err, syscall.EACCES) {
		t.Fatalf("err = %q, want the EACCES chain preserved for callers", err)
	}
}
