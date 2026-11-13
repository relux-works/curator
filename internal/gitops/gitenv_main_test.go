package gitops

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testgitenv"
)

// TestMain keeps every git this package's tests run off the operator's
// global and system git config; see internal/testgitenv. A copy of the test
// binary named git or ssh runs as a recording fake instead (see fakeTool).
func TestMain(m *testing.M) {
	switch strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe") {
	case "git":
		os.Exit(fakeGit())
	case "ssh":
		os.Exit(fakeSSH(os.Args[1:]))
	}
	restoreGitEnv := testgitenv.Isolate()
	code := m.Run()
	restoreGitEnv()
	os.Exit(code)
}

// fakeGit appends a line to $GIT_MARKER and fails, so a test can tell that
// git was invoked without running a transport. With $GIT_FAKE_STDERR set it
// instead records "LC_ALL=<value>", writes that text to standard error,
// creates a clone's destination and succeeds, so a test can drive the
// success-path stderr handling.
func fakeGit() int {
	stderr, succeed := os.LookupEnv("GIT_FAKE_STDERR")
	line := "ran\n"
	if succeed {
		line = "LC_ALL=" + os.Getenv("LC_ALL") + "\n"
	}
	marker, err := os.OpenFile(os.Getenv("GIT_MARKER"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err == nil {
		_, _ = marker.WriteString(line)
		_ = marker.Close()
	}
	if !succeed {
		return 1
	}
	for _, arg := range os.Args[1:] {
		if arg == "clone" {
			_ = os.MkdirAll(filepath.Join(os.Args[len(os.Args)-1], ".git"), 0o755)
		}
	}
	_, _ = os.Stderr.WriteString(stderr)
	return 0
}

// fakeSSH writes its arguments, one per line, to $SSH_LOG and fails, so a
// test can read the destination git hands ssh without contacting it.
func fakeSSH(args []string) int {
	_ = os.WriteFile(os.Getenv("SSH_LOG"), []byte(strings.Join(args, "\n")+"\n"), 0o600)
	return 1
}

// fakeTool copies this test binary into dir under name (name.exe on
// Windows), where TestMain runs it as that fake tool, and returns its path.
func fakeTool(t *testing.T, dir, name string) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, binary, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
