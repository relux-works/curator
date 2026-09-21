package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestProjectWriteBlobsSpawnFailureIsWrapped drives the production entry
// (install.Project over a v1 git source) with an injected git that presents
// the hosted-gate spawn refusal (BUG-260920-3vfwch) exactly at the writeBlobs
// `git cat-file --batch` spawn, and asserts the failure surfaces with the
// operation context instead of the bare fork/exec text.
//
// The shim stays 0755 and delegates every invocation to the real git, but
// after serving the `ls-tree` listing it replaces its own shebang
// interpreter with a 0644 file. The shim therefore keeps resolving on PATH
// (a 0644 shim itself would just be skipped by the PATH lookup) while the
// kernel refuses the next exec — the `cat-file --batch` child writeBlobs
// starts — with EACCES. A mutant restoring the bare `cmd.Start()` return
// fails this row: the error then carries no `git cat-file --batch`
// operation context.
func TestProjectWriteBlobsSpawnFailureIsWrapped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is unavailable")
	}
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	shimDir := t.TempDir()
	interp := filepath.Join(shimDir, "interp")
	if err := os.Symlink(shell, interp); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(shimDir, "git")
	script := "#!" + interp + "\n" +
		"REAL=" + realGit + "\n" +
		"INTERP=" + interp + "\n" +
		"TRIP=\"\"\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$a\" = \"ls-tree\" ]; then TRIP=1; break; fi\n" +
		"done\n" +
		"\"$REAL\" \"$@\"\n" +
		"code=$?\n" +
		"if [ -n \"$TRIP\" ]; then rm -f \"$INTERP\"; printf 'spawn blocked\\n' > \"$INTERP\"; chmod 0644 \"$INTERP\"; fi\n" +
		"exit $code\n"
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	// Sequential by construction: it rewrites PATH for the process, which
	// t.Setenv forbids under a parallel ancestor.
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result := e.install(Options{})
	if result.Status != "failed" {
		t.Fatalf("install = %+v, want a failed install", result)
	}
	joined := strings.Join(result.Errors, "\n")
	if !strings.Contains(joined, "git cat-file --batch failed") {
		t.Fatalf("errors = %q, want the wrapped cat-file operation context", joined)
	}
	if !strings.Contains(joined, "permission denied") {
		t.Fatalf("errors = %q, want the kernel refusal preserved", joined)
	}
	if info, err := os.Stat(interp); err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("interp mode = %v, %v; want the trip to have fired exactly once", info, err)
	}
}
