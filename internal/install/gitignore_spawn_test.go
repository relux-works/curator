package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestProjectRefusesOnGitignoreSpawnFailure pins the production entry of
// BUG-260921-1fpaij: when git check-ignore cannot spawn, install.Project
// refuses with the spawn diagnostic (status failed), never with the
// "not ignored" policy message (status skipped). The injected git is the
// shape the hosted gate observed: a PATH-resolvable 0755 script whose
// shebang interpreter is 0644, so the kernel refuses the exec with EACCES.
func TestProjectRefusesOnGitignoreSpawnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

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

	result := e.install(Options{})
	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed: %+v", result.Status, result)
	}
	joined := strings.Join(append(append([]string{}, result.Errors...), result.Messages...), "\n")
	if !strings.Contains(joined, "git check-ignore failed") {
		t.Fatalf("refusal = %q, want the spawn operation context: %+v", joined, result)
	}
	if !strings.Contains(joined, "permission denied") {
		t.Fatalf("refusal = %q, want the kernel refusal preserved: %+v", joined, result)
	}
	if strings.Contains(joined, "generated paths are not ignored") {
		t.Fatalf("refusal = %q, must never carry the policy message: %+v", joined, result)
	}
}
