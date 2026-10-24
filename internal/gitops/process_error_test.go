package gitops

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// Successful Git output remains available, while failed processes retain both
// their execution cause and diagnostics. An absent executable has no stderr and
// must not become the same empty diagnostic as an unavailable commit object.
// Nonempty Git output keeps its existing framing for transport classification.
func TestRunWithEnvPreservesProcessError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		out, err := runWithEnv(t.TempDir(), gitEnv(), "--version")
		if err != nil || !strings.HasPrefix(out, "git version ") {
			t.Fatalf("Git version control = %q, %v", out, err)
		}
	})
	t.Run("exit-with-stderr", func(t *testing.T) {
		out, err := runWithEnv(t.TempDir(), append(gitEnv(), "LC_ALL=C"), "--curator-invalid-option")
		var exit *exec.ExitError
		if out != "" || !errors.As(err, &exit) || exit.ExitCode() == 0 || !strings.Contains(err.Error(), "failed: unknown option:") {
			t.Fatalf("Git failure lost process cause or stderr: output=%q, err=%v", out, err)
		}
	})
	t.Run("missing-executable", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		out, err := runWithEnv(t.TempDir(), gitEnv(), "--version")
		if out != "" || !errors.Is(err, exec.ErrNotFound) || !strings.Contains(err.Error(), "executable file not found") {
			t.Fatalf("Git failure lost empty-stderr process cause: output=%q, err=%v", out, err)
		}
	})
}
