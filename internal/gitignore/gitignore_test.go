package gitignore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func gitProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestEnsureFailsWithoutIgnoreThenFixes(t *testing.T) {
	project := gitProject(t)
	entries := []string{".agents/", ".claude/skills/"}

	err := Ensure(project, entries, false)
	if err == nil || !strings.Contains(err.Error(), ".agents/") {
		t.Fatalf("err = %v, want missing entries", err)
	}

	if err := Ensure(project, entries, true); err != nil {
		t.Fatalf("fix failed: %v", err)
	}
	payload, _ := os.ReadFile(filepath.Join(project, ".gitignore"))
	text := string(payload)
	if !strings.Contains(text, BlockComment) || !strings.Contains(text, ".agents/") {
		t.Fatalf("gitignore content:\n%s", text)
	}

	// idempotent: no duplicate lines on a second fix
	if err := Ensure(project, entries, true); err != nil {
		t.Fatal(err)
	}
	payload, _ = os.ReadFile(filepath.Join(project, ".gitignore"))
	if strings.Count(string(payload), ".agents/") != 1 {
		t.Fatalf("duplicated entries:\n%s", payload)
	}
}

func TestEnsurePassesWhenAlreadyIgnored(t *testing.T) {
	project := gitProject(t)
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(project, []string{".agents/"}, false); err != nil {
		t.Fatalf("already-ignored entry reported missing: %v", err)
	}
}

// TestMissingExitOneIsNotIgnored pins the policy outcome: check-ignore exit 1
// still means "not ignored", reported as a NotIgnoredError, never as a tool
// failure.
func TestMissingExitOneIsNotIgnored(t *testing.T) {
	project := gitProject(t)
	missing, err := Missing(project, []string{".agents/"})
	if err != nil {
		t.Fatalf("Missing on an unignored entry: %v", err)
	}
	if len(missing) != 1 || missing[0] != ".agents/" {
		t.Fatalf("missing = %q, want [\".agents/\"]", missing)
	}
	err = Ensure(project, []string{".agents/"}, false)
	if err == nil {
		t.Fatal("Ensure on an unignored entry must fail")
	}
	if !IsNotIgnored(err) {
		t.Fatalf("err = %v (%T), want a NotIgnoredError", err, err)
	}
	if !strings.Contains(err.Error(), "generated paths are not ignored by git") {
		t.Fatalf("err = %q, want the stable policy message", err)
	}
}

// TestMissingSpawnFailureIsToolError pins the conflation fix (BUG-260921-1fpaij):
// a git spawn failure is a tool error carrying the kernel diagnostic, never
// the "not ignored" policy outcome. The injected git is the shape the hosted
// gate observed: a PATH-resolvable 0755 script whose shebang interpreter is
// 0644, so the kernel refuses the exec with EACCES.
func TestMissingSpawnFailureIsToolError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	project := gitProject(t)
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

	if _, err := Missing(project, []string{".agents/"}); err == nil {
		t.Fatal("Missing with a non-executable git must fail")
	} else {
		if IsNotIgnored(err) {
			t.Fatalf("err = %q, want a tool error, not the policy outcome", err)
		}
		if !strings.Contains(err.Error(), "git check-ignore failed") {
			t.Fatalf("err = %q, want the wrapped operation context", err)
		}
		if !strings.Contains(err.Error(), ".agents/") {
			t.Fatalf("err = %q, want the entry name for context", err)
		}
		if strings.Contains(err.Error(), ".curator-probe") {
			t.Fatalf("err = %q, must not leak the probe path", err)
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Fatalf("err = %q, want the kernel refusal preserved", err)
		}
		if !errors.Is(err, syscall.EACCES) {
			t.Fatalf("err = %q, want the EACCES chain preserved for callers", err)
		}
	}
	if err := Ensure(project, []string{".agents/"}, false); err == nil {
		t.Fatal("Ensure with a non-executable git must fail")
	} else {
		if IsNotIgnored(err) {
			t.Fatalf("err = %q, want a tool error, not the policy outcome", err)
		}
		if strings.Contains(err.Error(), "generated paths are not ignored") {
			t.Fatalf("err = %q, must never carry the policy message", err)
		}
	}
}

func TestMissingRetriesOnceOnSpawnEACCES(t *testing.T) {
	previousRunner := checkIgnoreCommandRunner
	t.Cleanup(func() { checkIgnoreCommandRunner = previousRunner })

	calls := 0
	checkIgnoreCommandRunner = func(*exec.Cmd) error {
		calls++
		if calls == 1 {
			return checkIgnoreSpawnError(syscall.EACCES)
		}
		return nil
	}

	missing, err := Missing(t.TempDir(), []string{".agents/"})
	if err != nil {
		t.Fatalf("Missing after transient EACCES: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %q, want no missing entries after successful retry", missing)
	}
	if calls != 2 {
		t.Fatalf("runner calls = %d, want one retry after the first EACCES", calls)
	}
}

func TestMissingDoesNotRetryGitExitStatus(t *testing.T) {
	previousRunner := checkIgnoreCommandRunner
	t.Cleanup(func() { checkIgnoreCommandRunner = previousRunner })

	calls := 0
	checkIgnoreCommandRunner = func(*exec.Cmd) error {
		calls++
		return &exec.ExitError{}
	}

	missing, err := Missing(t.TempDir(), []string{".agents/"})
	if err != nil {
		t.Fatalf("Missing after a Git exit status: %v", err)
	}
	if len(missing) != 1 || missing[0] != ".agents/" {
		t.Fatalf("missing = %q, want the existing not-ignored policy result", missing)
	}
	if calls != 1 {
		t.Fatalf("runner calls = %d, want no retry after Git starts and exits", calls)
	}
}

func TestEnsurePersistentSpawnEACCESFailsClosedAfterOneRetry(t *testing.T) {
	previousRunner := checkIgnoreCommandRunner
	t.Cleanup(func() { checkIgnoreCommandRunner = previousRunner })

	calls := 0
	checkIgnoreCommandRunner = func(*exec.Cmd) error {
		calls++
		return checkIgnoreSpawnError(syscall.EACCES)
	}

	project := t.TempDir()
	err := Ensure(project, []string{".agents/"}, true)
	if err == nil {
		t.Fatal("Ensure with persistent spawn EACCES must fail")
	}
	if IsNotIgnored(err) {
		t.Fatalf("err = %q, want a tool error rather than a policy outcome", err)
	}
	if !errors.Is(err, syscall.EACCES) {
		t.Fatalf("err = %q, want the persistent EACCES cause preserved", err)
	}
	if calls != 2 {
		t.Fatalf("runner calls = %d, want exactly two attempts", calls)
	}
	if _, statErr := os.Stat(filepath.Join(project, ".gitignore")); !os.IsNotExist(statErr) {
		t.Fatalf("persistent spawn error must not write .gitignore: %v", statErr)
	}
}

func TestMissingDoesNotRetryOtherSpawnErrors(t *testing.T) {
	for _, row := range []struct {
		name      string
		err       error
		wantErrno syscall.Errno
	}{
		{"EPERM spawn", checkIgnoreSpawnError(syscall.EPERM), syscall.EPERM},
		{"ENOENT spawn", checkIgnoreSpawnError(syscall.ENOENT), syscall.ENOENT},
		{"EACCES outside fork-exec", &os.PathError{Op: "open", Path: "/opt/homebrew/bin/git", Err: syscall.EACCES}, syscall.EACCES},
	} {
		t.Run(row.name, func(t *testing.T) {
			previousRunner := checkIgnoreCommandRunner
			t.Cleanup(func() { checkIgnoreCommandRunner = previousRunner })

			calls := 0
			checkIgnoreCommandRunner = func(*exec.Cmd) error {
				calls++
				return row.err
			}

			_, err := Missing(t.TempDir(), []string{".agents/"})
			if err == nil {
				t.Fatal("Missing with a non-retryable spawn error must fail")
			}
			if !errors.Is(err, row.wantErrno) {
				t.Fatalf("err = %q, want %v preserved", err, row.wantErrno)
			}
			if calls != 1 {
				t.Fatalf("runner calls = %d, want no retry for %v", calls, row.name)
			}
		})
	}
}

func checkIgnoreSpawnError(errno syscall.Errno) error {
	return &os.PathError{Op: "fork/exec", Path: "/opt/homebrew/bin/git", Err: errno}
}

// TestMissingNonRepositoryIsNotIgnored pins the exit-128 branch of
// BUG-260921-1fpaij (rework 1): a root that is not a git repository keeps the
// PRE-EXISTING outcome — git's own verdict, including "not a repository", is
// a policy outcome (not ignored), never a tool error. Only a git that cannot
// be executed is an error.
func TestMissingNonRepositoryIsNotIgnored(t *testing.T) {
	plain := t.TempDir()
	missing, err := Missing(plain, []string{".agents/"})
	if err != nil {
		t.Fatalf("Missing outside a repository: %v, want the pre-existing not-ignored outcome", err)
	}
	if len(missing) != 1 || missing[0] != ".agents/" {
		t.Fatalf("missing = %q, want [\".agents/\"]", missing)
	}
	err = Ensure(plain, []string{".agents/"}, false)
	if err == nil {
		t.Fatal("Ensure outside a repository must report not ignored")
	}
	if !IsNotIgnored(err) {
		t.Fatalf("err = %v (%T), want a NotIgnoredError", err, err)
	}
	if !strings.Contains(err.Error(), "generated paths are not ignored by git") {
		t.Fatalf("err = %q, want the stable policy message", err)
	}
}

// TestMissingGitAbsentIsToolError pins exec.ErrNotFound: git absent from PATH
// is a tool error, never "not ignored".
func TestMissingGitAbsentIsToolError(t *testing.T) {
	project := gitProject(t)
	empty := t.TempDir()
	t.Setenv("PATH", empty)
	if _, err := Missing(project, []string{".agents/"}); err == nil {
		t.Fatal("Missing without a git binary must fail")
	} else {
		if IsNotIgnored(err) {
			t.Fatalf("err = %q, want a tool error, not the policy outcome", err)
		}
		if !strings.Contains(err.Error(), "git check-ignore failed") {
			t.Fatalf("err = %q, want the wrapped operation context", err)
		}
		if !errors.Is(err, exec.ErrNotFound) {
			t.Fatalf("err = %q, want the ErrNotFound chain preserved", err)
		}
	}
}
