package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// gitFixtureRlimitsPattern pins the unix rlimits rendering: both ceilings as
// plain numbers. A sysctl-bytes blob or an "n/a" dodge fails the row.
var gitFixtureRlimitsPattern = regexp.MustCompile(`NOFILE cur=\d+ max=\d+; NPROC cur=\d+ max=\d+`)

// gitFixtureScript replays canned spawn results: each call returns the next
// entry, repeating the last one once the script is exhausted.
type gitFixtureScript struct {
	results []struct {
		output []byte
		err    error
	}
	calls int
}

func (s *gitFixtureScript) runner(*exec.Cmd) ([]byte, error) {
	s.calls++
	i := s.calls - 1
	if i >= len(s.results) {
		i = len(s.results) - 1
	}
	return s.results[i].output, s.results[i].err
}

func spawnErr(errno syscall.Errno) error {
	return &os.PathError{Op: "fork/exec", Path: "/opt/homebrew/bin/git", Err: errno}
}

func TestGitFixtureRetriesOnceOnSpawnEACCES(t *testing.T) {
	t.Parallel()
	script := &gitFixtureScript{results: []struct {
		output []byte
		err    error
	}{{nil, spawnErr(syscall.EACCES)}, {[]byte("ok"), nil}}}
	var logs []string
	var sleeps []time.Duration
	outcome := runGitFixture(t.TempDir(), []string{"add", "."}, nil, gitFixtureConfig{
		runner: script.runner,
		sleep:  func(d time.Duration) { sleeps = append(sleeps, d) },
		logf:   func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) },
	})
	if outcome.fatal != "" {
		t.Fatalf("fatal = %q, want success on retry", outcome.fatal)
	}
	if outcome.attempts != 2 || script.calls != 2 {
		t.Fatalf("attempts = %d, runner calls = %d, want 2 and 2", outcome.attempts, script.calls)
	}
	if string(outcome.output) != "ok" {
		t.Fatalf("output = %q, want the retry output", outcome.output)
	}
	// Pinned against the literal, not the constant, so a backoff change
	// fails this row.
	if len(sleeps) != 1 || sleeps[0] != 200*time.Millisecond {
		t.Fatalf("sleeps = %v, want one %s backoff", sleeps, 200*time.Millisecond)
	}
	if len(logs) != 2 || !strings.Contains(logs[0], "retrying once") || !strings.Contains(logs[1], "succeeded on retry") {
		t.Fatalf("logs = %q, want the retry line and the success line", logs)
	}
}

func TestGitFixtureRetriesOnceOnSpawnEAGAIN(t *testing.T) {
	t.Parallel()
	script := &gitFixtureScript{results: []struct {
		output []byte
		err    error
	}{{nil, spawnErr(syscall.EAGAIN)}, {[]byte("ok"), nil}}}
	outcome := runGitFixture(t.TempDir(), []string{"add", "."}, nil, gitFixtureConfig{runner: script.runner})
	if outcome.fatal != "" || outcome.attempts != 2 {
		t.Fatalf("outcome = %+v, want success on the second attempt", outcome)
	}
}

func TestGitFixtureNeverRetriesNonZeroExit(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	var logs []string
	sleeps := 0
	// A real git process that runs and exits non-zero: an unknown global
	// flag needs no repository, no network, and fails deterministically.
	outcome := runGitFixture(t.TempDir(), []string{"--not-a-git-flag-xyz"}, nil, gitFixtureConfig{
		sleep: func(time.Duration) { sleeps++ },
		logf:  func(format string, _ ...any) { logs = append(logs, format) },
	})
	if outcome.attempts != 1 {
		t.Fatalf("attempts = %d, want 1: a non-zero git exit must never be retried", outcome.attempts)
	}
	if outcome.fatal == "" || !strings.Contains(outcome.fatal, "exit") {
		t.Fatalf("fatal = %q, want the exit failure", outcome.fatal)
	}
	if strings.Contains(outcome.fatal, "gitfixture diagnostic") {
		t.Fatalf("fatal = %q, want no spawn diagnostic for a git exit", outcome.fatal)
	}
	if sleeps != 0 || len(logs) != 0 {
		t.Fatalf("sleeps = %d, logs = %q, want no backoff and no retry log", sleeps, logs)
	}
}

func TestGitFixtureNeverRetriesWrappedExitError(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	// Capture a genuine *exec.ExitError through the real runner, then prove
	// the policy recognises it through a wrapping layer.
	_, _, realErr := gitAttempt("git", t.TempDir(), []string{"--not-a-git-flag-xyz"}, nil, realGitRunner)
	if realErr == nil || !isGitExitError(realErr) {
		t.Fatalf("test control: no real exit error captured: %v", realErr)
	}
	script := &gitFixtureScript{results: []struct {
		output []byte
		err    error
	}{{nil, fmt.Errorf("launch: %w", realErr)}}}
	outcome := runGitFixture(t.TempDir(), []string{"add", "."}, nil, gitFixtureConfig{runner: script.runner})
	if outcome.attempts != 1 || outcome.fatal == "" {
		t.Fatalf("outcome = %+v, want a single unretried attempt", outcome)
	}
	if strings.Contains(outcome.fatal, "gitfixture diagnostic") {
		t.Fatalf("fatal = %q, want no spawn diagnostic for a git exit", outcome.fatal)
	}
}

func TestGitFixtureRetryPredicate(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	_, _, realErr := gitAttempt("git", t.TempDir(), []string{"--not-a-git-flag-xyz"}, nil, realGitRunner)
	if realErr == nil || !isGitExitError(realErr) {
		t.Fatalf("test control: no real exit error captured: %v", realErr)
	}
	rows := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"exit", realErr, false},
		{"wrapped exit", fmt.Errorf("launch: %w", realErr), false},
		{"EACCES", spawnErr(syscall.EACCES), true},
		{"EAGAIN", spawnErr(syscall.EAGAIN), true},
		{"ENOENT", spawnErr(syscall.ENOENT), false},
		{"EPERM", spawnErr(syscall.EPERM), false},
		{"wrapped EACCES", fmt.Errorf("spawn: %w", spawnErr(syscall.EACCES)), true},
		{"lookup failure", &exec.Error{Name: "git", Err: exec.ErrNotFound}, false},
	}
	for _, row := range rows {
		if got := isRetryableSpawnError(row.err); got != row.want {
			t.Errorf("isRetryableSpawnError(%s) = %v, want %v", row.name, got, row.want)
		}
	}
}

func TestGitFixturePersistentSpawnFailureCarriesDiagnostic(t *testing.T) {
	t.Parallel()
	script := &gitFixtureScript{results: []struct {
		output []byte
		err    error
	}{{nil, spawnErr(syscall.EACCES)}}}
	dir := t.TempDir()
	outcome := runGitFixture(dir, []string{"add", "."}, nil, gitFixtureConfig{runner: script.runner})
	if outcome.attempts != 2 {
		t.Fatalf("attempts = %d, want the bounded two", outcome.attempts)
	}
	for _, want := range []string{
		"gitfixture diagnostic",
		"resolved:",
		"binary:",
		"dir ",
		"cwd:",
		"PATH:",
		"rlimits:",
		"first attempt:",
	} {
		if !strings.Contains(outcome.fatal, want) {
			t.Fatalf("fatal = %q, want it to contain %q", outcome.fatal, want)
		}
	}
	// On unix both rlimits must render as numbers; other platforms report
	// unavailable.
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if !gitFixtureRlimitsPattern.MatchString(outcome.fatal) {
			t.Fatalf("fatal = %q, want numeric NOFILE/NPROC rlimits", outcome.fatal)
		}
	}
}

func TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic(t *testing.T) {
	t.Parallel()
	script := &gitFixtureScript{results: []struct {
		output []byte
		err    error
	}{{nil, spawnErr(syscall.ENOENT)}}}
	sleeps := 0
	outcome := runGitFixture(t.TempDir(), []string{"add", "."}, nil, gitFixtureConfig{
		runner: script.runner,
		sleep:  func(time.Duration) { sleeps++ },
	})
	if outcome.attempts != 1 || script.calls != 1 {
		t.Fatalf("attempts = %d, want a single attempt for a non-retryable errno", outcome.attempts)
	}
	if !strings.Contains(outcome.fatal, "gitfixture diagnostic") {
		t.Fatalf("fatal = %q, want the spawn diagnostic", outcome.fatal)
	}
	if sleeps != 0 {
		t.Fatalf("sleeps = %d, want no backoff without a retry", sleeps)
	}
}

func TestGitFixtureRealSpawnEACCESRetriesOnce(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	// A real kernel EACCES through the real runner: an absolute binary path
	// skips the PATH lookup, so the non-executable mode surfaces as the
	// fork/exec refusal the gate observed, not a lookup failure.
	blocker := filepath.Join(t.TempDir(), "non-executable-git")
	if err := os.WriteFile(blocker, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocker, 0o644); err != nil {
		t.Fatal(err)
	}
	var logs []string
	outcome := runGitFixture(t.TempDir(), []string{"--version"}, nil, gitFixtureConfig{
		binary: blocker,
		sleep:  func(time.Duration) {},
		logf:   func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) },
	})
	if outcome.attempts != 2 {
		t.Fatalf("attempts = %d, want the bounded two for a real EACCES", outcome.attempts)
	}
	if !strings.Contains(outcome.fatal, "permission denied") {
		t.Fatalf("fatal = %q, want the kernel refusal", outcome.fatal)
	}
	if len(logs) == 0 || !strings.Contains(logs[0], "retrying once") {
		t.Fatalf("logs = %q, want the retry line", logs)
	}
}

func TestGitFixtureUnresolvableGitFailsFastWithDiagnostic(t *testing.T) {
	// Sequential by construction: it rewrites PATH for the process, which
	// t.Setenv forbids under a parallel ancestor. An empty PATH directory
	// proves the lookup-failure forensics, the counterpart of the spawn
	// refusal above.
	t.Setenv("PATH", t.TempDir())
	outcome := runGitFixture(t.TempDir(), []string{"--version"}, nil, gitFixtureConfig{})
	if outcome.attempts != 1 {
		t.Fatalf("attempts = %d, want a single attempt for an unresolvable binary", outcome.attempts)
	}
	if !strings.Contains(outcome.fatal, "unresolved:") || !strings.Contains(outcome.fatal, "gitfixture diagnostic") {
		t.Fatalf("fatal = %q, want the unresolved-binary diagnostic", outcome.fatal)
	}
}

func TestGitFixtureOuterEntryRunsRealGit(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	// Drives the exact entry point the wired fixture helpers call.
	out := gitFixture(t, t.TempDir(), []string{"--version"}, []string{"--version"}, nil)
	if !strings.Contains(string(out), "git version") {
		t.Fatalf("output = %q, want the git version", out)
	}
}

// gitFixtureRecordingTB is a testing.TB that records Logf lines while
// delegating everything else (including Fatalf) to the real test. When flip
// is set, it runs exactly once, on the line that carries "retrying once".
type gitFixtureRecordingTB struct {
	testing.TB
	logs *[]string
	flip func()
}

func (tb gitFixtureRecordingTB) Logf(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	*tb.logs = append(*tb.logs, line)
	tb.TB.Logf(format, args...)
	if strings.Contains(line, "retrying once") && tb.flip != nil {
		tb.flip()
	}
}

func TestGitFixtureOuterEntryLogsRetryOnTransientEACCES(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("executable-bit spawn refusal is exercised on the unix runners")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	root := t.TempDir()
	// An absolute-path fixture git that starts non-executable: no PATH lookup
	// is involved, so the kernel refuses the exec with EACCES, the gate's
	// exact `fork/exec ...: permission denied` shape. The flip onto the real
	// git is event-driven: the recording TB swaps the symlink when it sees
	// the "retrying once" line, which runGitFixture emits synchronously after
	// attempt 1 failed and before the backoff sleep — so attempt 1 always
	// sees the blocker and the retry always sees the real git, with no
	// timing anywhere. (A PATH-resolved script with a 0644 shebang
	// interpreter is the same shape, but this host SIGKILLs freshly
	// materialised test executables, so the retry target is the real git
	// binary instead of a second fixture file.)
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("not executable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "git")
	if err := os.Symlink(blocker, link); err != nil {
		t.Fatal(err)
	}
	flip := func() {
		fresh := filepath.Join(root, "fresh")
		if err := os.Symlink(realGit, fresh); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(fresh, link); err != nil {
			t.Fatal(err)
		}
	}
	// Drives the outer entry through a real transient EACCES and pins the
	// wiring: without logf the retry succeeds silently and this row fails.
	var logs []string
	gitFixtureWithBinary(gitFixtureRecordingTB{TB: t, logs: &logs, flip: flip}, link, t.TempDir(), []string{"--version"}, []string{"--version"}, nil)
	if len(logs) != 2 || !strings.Contains(logs[0], "retrying once") || !strings.Contains(logs[1], "succeeded on retry") {
		t.Fatalf("logs = %q, want the retry line and the success line from the outer entry", logs)
	}
}
