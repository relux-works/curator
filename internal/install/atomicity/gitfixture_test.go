// Shared fixture-git spawn path for the atomicity test binary.
//
// This is the same helper as internal/install's gitfixture_test.go, kept as a
// per-package copy because Go test files cannot be imported across packages.
// Keep the retry policy and diagnostic keys in sync between the two.
//
// Hosted macos-latest gate runs 35340496757, 35481906193 and 35510984798 each
// failed one test with `fork/exec /opt/homebrew/bin/git: permission denied`
// from a git spawn. Go reports every child-startup failure (exec, in-child
// chdir into cmd.Dir, dup) as `fork/exec <binary>: <errno>`, so the bare
// error cannot tell a transient spawn refusal from a broken image binary.
// Every fixture git invocation in this package therefore goes through
// gitFixture, which adds two things and nothing else:
//
//  1. An in-process diagnostic captured at the moment a spawn fails: the
//     PATH resolution of git, lstat/stat of the resolved binary and of
//     cmd.Dir, the working directory, and the NOFILE/NPROC rlimits. The
//     next occurrence proves its cause from the go-test stream.
//  2. One bounded retry, only when the spawn itself fails with EACCES or
//     EAGAIN, after a short backoff, recorded in the test log. A non-zero
//     git exit is never retried: that is a fixture or behaviour signal,
//     not a spawn flake.
//
// Success-path behaviour is unchanged: same argv, same environment, same
// working directory, same combined output.
package atomicity

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// gitFixtureBackoff is the single wait before the one bounded retry of a
// fixture git spawn that failed with EACCES/EAGAIN. It only ever delays a
// test that would otherwise have failed outright.
const gitFixtureBackoff = 200 * time.Millisecond

// gitRunner runs one prepared git command. The nil runner is the real
// exec runner; tests inject fakes to drive the retry policy.
type gitRunner func(cmd *exec.Cmd) ([]byte, error)

// gitFixtureConfig carries the test seams of runGitFixture. The zero value is
// the real path: PATH-resolved "git", the real exec runner, a real backoff
// sleep, and no logging. Tests override binary with an absolute path to
// drive a real kernel spawn failure, and runner/sleep/logf with fakes.
type gitFixtureConfig struct {
	binary string
	runner gitRunner
	sleep  func(time.Duration)
	logf   func(string, ...any)
}

func realGitRunner(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

// gitFixtureOutcome is the assertable result of one fixture git invocation.
type gitFixtureOutcome struct {
	output   []byte
	attempts int
	// fatal is the message body after "git <args>: ". Empty on success.
	fatal string
}

// gitFixture runs one fixture git command and returns its combined output.
// displayArgs names the invocation in failure messages (the caller's own
// args, without the signing-policy prefix); gitArgs is the full argv after
// "git". Any failure is fatal to the calling test. A bounded retry is logged
// through t.Logf, so the go-test stream records it.
func gitFixture(t testing.TB, dir string, displayArgs, gitArgs, extraEnv []string) []byte {
	t.Helper()
	return gitFixtureWithBinary(t, "", dir, displayArgs, gitArgs, extraEnv)
}

// gitFixtureWithBinary is gitFixture with an explicit binary. An empty binary
// is the default "git" PATH lookup; callers that pre-resolve git (the draft
// transport wrapper fixtures) pass their resolved absolute path through
// unchanged.
func gitFixtureWithBinary(t testing.TB, binary, dir string, displayArgs, gitArgs, extraEnv []string) []byte {
	t.Helper()
	outcome := runGitFixture(dir, gitArgs, extraEnv, gitFixtureConfig{binary: binary, logf: t.Logf})
	if outcome.fatal == "" {
		return outcome.output
	}
	t.Fatalf("git %v: %s", displayArgs, outcome.fatal)
	return nil
}

// runGitFixture executes one fixture git invocation with the bounded retry
// policy. It never fails the test itself, so unit rows can assert attempts,
// fatal text and logging directly.
func runGitFixture(dir string, gitArgs, extraEnv []string, cfg gitFixtureConfig) gitFixtureOutcome {
	binary := cfg.binary
	if binary == "" {
		binary = "git"
	}
	runner := cfg.runner
	if runner == nil {
		runner = realGitRunner
	}
	sleep := cfg.sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	logf := cfg.logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	output, resolved, err := gitAttempt(binary, dir, gitArgs, extraEnv, runner)
	if err == nil {
		return gitFixtureOutcome{output: output, attempts: 1}
	}
	if isGitExitError(err) {
		// A git process ran and reported failure. That is a fixture or
		// behaviour signal; retrying would weaken the test.
		return gitFixtureOutcome{output: output, attempts: 1, fatal: fmt.Sprintf("%v\n%s", err, output)}
	}
	if !isRetryableSpawnError(err) {
		return gitFixtureOutcome{output: output, attempts: 1, fatal: fmt.Sprintf("%v\n%s\n%s", err, output, describeGitSpawnFailure(binary, resolved, dir))}
	}
	firstErr := err
	firstDiagnostic := describeGitSpawnFailure(binary, resolved, dir)
	logf("gitfixture: git %v spawn failed (%v); retrying once after %s\n%s", gitArgs, firstErr, gitFixtureBackoff, firstDiagnostic)
	sleep(gitFixtureBackoff)

	output, resolved, err = gitAttempt(binary, dir, gitArgs, extraEnv, runner)
	if err == nil {
		logf("gitfixture: git %v succeeded on retry", gitArgs)
		return gitFixtureOutcome{output: output, attempts: 2}
	}
	if isGitExitError(err) {
		return gitFixtureOutcome{output: output, attempts: 2, fatal: fmt.Sprintf("%v (first spawn attempt failed: %v, retried once)\n%s", err, firstErr, output)}
	}
	return gitFixtureOutcome{output: output, attempts: 2, fatal: fmt.Sprintf("%v (first attempt: %v)\n%s\n%s", err, firstErr, output, describeGitSpawnFailure(binary, resolved, dir))}
}

// gitAttempt builds and runs one fixture git invocation.
func gitAttempt(binary, dir string, gitArgs, extraEnv []string, runner gitRunner) (output []byte, resolved string, err error) {
	cmd := exec.Command(binary, gitArgs...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), extraEnv...)
	resolved = resolveGitFixture(binary)
	output, err = runner(cmd)
	return output, resolved, err
}

// resolveGitFixture resolves the fixture binary the way Start will, for
// diagnostics. An absolute binary needs no lookup.
func resolveGitFixture(binary string) string {
	if filepath.IsAbs(binary) {
		return binary
	}
	resolved, err := exec.LookPath(binary)
	if err != nil {
		return "unresolved: " + err.Error()
	}
	return resolved
}

// isGitExitError reports whether a git process ran and exited non-zero, as
// opposed to the spawn itself failing.
func isGitExitError(err error) bool {
	var exit *exec.ExitError
	return errors.As(err, &exit)
}

// isRetryableSpawnError reports a child-startup EACCES/EAGAIN: the process
// never ran. Every other error, including any non-zero git exit, is not
// retried.
func isRetryableSpawnError(err error) bool {
	if err == nil || isGitExitError(err) {
		return false
	}
	return errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EAGAIN)
}

// describeGitSpawnFailure gathers the in-process facts that discriminate a
// transient spawn refusal (EACCES/EAGAIN at exec) from a revoked cmd.Dir,
// a rewritten PATH, and a genuinely broken git binary.
func describeGitSpawnFailure(binary, resolved, dir string) string {
	var b strings.Builder
	b.WriteString("gitfixture diagnostic (spawn failed, no git process ran):")
	fmt.Fprintf(&b, "\n  argv0 %q resolved: %s", binary, resolved)
	if !strings.HasPrefix(resolved, "unresolved: ") {
		fmt.Fprintf(&b, "\n  binary: %s", statGitFixture(resolved))
	}
	if dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			fmt.Fprintf(&b, "\n  dir: (inherited) %s: %s", cwd, statGitFixture(cwd))
		} else {
			fmt.Fprintf(&b, "\n  dir: (inherited) getwd: %v", err)
		}
	} else {
		fmt.Fprintf(&b, "\n  dir %q: %s", dir, statGitFixture(dir))
	}
	if cwd, err := os.Getwd(); err == nil {
		fmt.Fprintf(&b, "\n  cwd: %s", cwd)
	} else {
		fmt.Fprintf(&b, "\n  cwd: getwd: %v", err)
	}
	fmt.Fprintf(&b, "\n  PATH: %s", os.Getenv("PATH"))
	fmt.Fprintf(&b, "\n  rlimits: %s", gitFixtureRlimits())
	return b.String()
}

// statGitFixture renders the lstat/stat facts for one diagnostic path.
func statGitFixture(path string) string {
	linfo, lerr := os.Lstat(path)
	if lerr != nil {
		return "lstat: " + lerr.Error()
	}
	sinfo, serr := os.Stat(path)
	if serr != nil {
		return fmt.Sprintf("lstat: mode=%s; stat: %v", linfo.Mode(), serr)
	}
	target := ""
	if linfo.Mode()&os.ModeSymlink != 0 {
		if link, err := os.Readlink(path); err == nil {
			target = " -> " + link
		} else {
			target = " -> (readlink: " + err.Error() + ")"
		}
		if eval, err := filepath.EvalSymlinks(path); err == nil {
			target += " (eval: " + eval + ")"
		}
	}
	return fmt.Sprintf("lstat: mode=%s%s; stat: mode=%s size=%d", linfo.Mode(), target, sinfo.Mode(), sinfo.Size())
}
