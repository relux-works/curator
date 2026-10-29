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
// git was invoked without running a transport.
func fakeGit() int {
	marker, err := os.OpenFile(os.Getenv("GIT_MARKER"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err == nil {
		_, _ = marker.WriteString("ran\n")
		_ = marker.Close()
	}
	return 1
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
