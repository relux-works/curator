// Package testcli runs the compiled curator binary and host tools for
// conformance tests that must cross the process boundary.
//
// The crossconformance package forbids process spawning in its own
// files (guard_test.go), so every child process a conformance test
// starts — building the CLI once per test run, invoking it with an
// isolated home, driving fixture git commands, cross-compiling the
// manager for another GOOS — lives behind this seam. Production code
// never imports it.
package testcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var binaryOnce struct {
	sync.Mutex
	path  string
	err   error
	built bool
}

// ModuleRoot returns the repository root derived from this package's
// own location, independent of the caller's working directory.
func ModuleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("testcli: runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// Binary builds ./cmd/curator once per test process and returns its
// absolute path.
func Binary(t *testing.T) string {
	t.Helper()
	root := ModuleRoot(t)
	binaryOnce.Lock()
	defer binaryOnce.Unlock()
	if !binaryOnce.built {
		binaryOnce.built = true
		dir, err := os.MkdirTemp("", "curator-conformance-bin")
		if err != nil {
			binaryOnce.err = err
		} else {
			name := "curator"
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			bin := filepath.Join(dir, name)
			cmd := exec.Command("go", "build", "-o", bin, "./cmd/curator") // #nosec G204 -- fixed test-harness build of the manager binary.
			cmd.Dir = root
			if out, err := cmd.CombinedOutput(); err != nil {
				binaryOnce.err = fmt.Errorf("build curator: %v\n%s", err, out)
			} else {
				binaryOnce.path = bin
			}
		}
	}
	if binaryOnce.err != nil {
		t.Fatalf("testcli binary: %v", binaryOnce.err)
	}
	return binaryOnce.path
}

// Run starts one child process, appends env to the ambient
// environment, feeds stdin when non-empty, and returns the exit code
// with captured stdout and stderr. dir == "" inherits the working
// directory. A start failure (as opposed to a non-zero exit) is
// fatal.
func Run(t *testing.T, dir string, env []string, stdin, name string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(name, args...) // #nosec G204 -- test-only seam; callers pass the built manager binary or fixed fixture tools.
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %s %v: %v", name, args, err)
		}
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

// Output runs one child process and returns its trimmed stdout. Any
// failure is fatal.
func Output(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...) // #nosec G204 -- test-only seam; callers pass fixed fixture tools.
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			t.Fatalf("%s %v: %v\n%s", name, args, err, exit.Stderr)
		}
		t.Fatalf("%s %v: %v", name, args, err)
	}
	return strings.TrimSpace(string(out))
}

// LookPath resolves a host tool without failing.
func LookPath(file string) (string, bool) {
	path, err := exec.LookPath(file)
	if err != nil {
		return "", false
	}
	return path, true
}

// RequireGit resolves git or skips the test.
func RequireGit(t *testing.T) string {
	t.Helper()
	path, ok := LookPath("git")
	if !ok {
		t.Skip("git is not available")
	}
	return path
}

// gitEnv pins the fixture commit identity for every git invocation.
func gitEnv() []string {
	return []string{
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	}
}

// Git runs one fixture git command with signatures disabled. A
// missing git skips; any failure is fatal.
func Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	RequireGit(t)
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	if code, _, stderr := Run(t, dir, gitEnv(), "", "git", gitArgs...); code != 0 {
		t.Fatalf("git %v: exit %d\n%s", args, code, stderr)
	}
}

// GitOutput runs one fixture git command and returns its trimmed
// stdout. A missing git skips; any failure is fatal.
func GitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	RequireGit(t)
	code, stdout, stderr := Run(t, dir, gitEnv(), "", "git", args...)
	if code != 0 {
		t.Fatalf("git %v: exit %d\n%s", args, code, stderr)
	}
	return strings.TrimSpace(stdout)
}

// GoBuild cross-compiles ./cmd/curator to output with extraEnv
// appended to the environment. Any failure is fatal.
func GoBuild(t *testing.T, dir, output string, extraEnv []string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/curator") // #nosec G204 -- fixed test-harness cross-compile of the manager binary.
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
}
