//go:build darwin || linux

package godriver

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each scoped plant must fail at its intended assertion, after a passing real
// subprocess control. A build error or unrelated failure is not detection.
func TestWorkerExitNarrowingControls(t *testing.T) {
	for _, p := range []struct{ name, file, before, after, witness, reason string }{
		{"fatal-SIGPIPE", "orphan_unix.go", "signal.Notify(c, syscall.SIGPIPE)", "signal.Stop(c); _ = syscall.SIGPIPE", "TestWorkerClosedManagerPipes/header", "domain closure: status=signal broken pipe"},
		{"no-worker-deadline", "workerserver.go", "default:\n\t\t\t}\n\t\t\tstopOrphanedWorkerDomain()", "default:\n\t\t\t}\n\t\t\t// omitted watcher kill", "TestWorkerIndependentDeadline", "worker deadline watcher did not terminate domain"},
		{"no-success-teardown", "orphan_unix.go", "func finishWorkerDomain() { stopOrphanedWorkerDomain() }", "func finishWorkerDomain() {}", "TestWorkerSuccessfulShutdownKillsDescendants", "domain closure: status=exit 0"},
		{"collapsed-status", "workerclient.go", "client.command.ProcessState.ExitCode() == exitWorkerTeardownFailed", "client.command.ProcessState.ExitCode() == -999", "TestWorkerTeardownStatusDiagnostic", "teardown status diagnostic"},
		{"ignored-kill-error", "orphan_unix.go", "if err := orphanKill(-os.Getpid(), unix.SIGKILL); err != nil {", "if err := orphanKill(-os.Getpid(), unix.SIGKILL); err != nil && false {", "TestWorkerCheckedKill/kill-error", "teardown error status"},
		{"no-signal-wait", "orphan_unix.go", "orphanSignalWait()", "// omitted signal wait", "TestWorkerCheckedKill/delayed-signal", "successful kill fell through"},
		{"no-failure-teardown", "workerserver.go", "session.fail(err)\n\t\tstopOrphanedWorkerDomain()", "session.fail(err)", "TestWorkerFailureKillsDescendants/session-failure", "failure descendant survival marker"},
		{"EOF-only-reader", "workerserver.go", "if err != nil {\n\t\t\t\tselect {", "if errors.Is(err, errProtocolClosed) || errors.Is(err, io.ErrUnexpectedEOF) {\n\t\t\t\tselect {", "TestWorkerFailureKillsDescendants/read-error", "read error did not terminate live domain"},
	} {
		t.Run(p.name, func(t *testing.T) {
			run := func(overlay string) []byte {
				args := []string{"test", "-count=1", "-timeout=60s", "-v", "-run", "^" + strings.ReplaceAll(p.witness, "/", "$/^") + "$", "."}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				c := exec.CommandContext(t.Context(), "go", args...)
				b, err := c.CombinedOutput()
				if overlay == "" {
					if err != nil || !strings.Contains(string(b), "--- PASS:") {
						t.Fatalf("positive control: %v\n%s", err, b)
					}
				} else {
					if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(b), p.reason) || !strings.Contains(string(b), "--- FAIL:") {
						t.Fatalf("plant failed outside intended assertion: %v\n%s", err, b)
					}
				}
				return b
			}
			run("")
			path, _ := filepath.Abs(p.file)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(raw), p.before) != 1 {
				t.Fatal("plant anchor not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, p.file)
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(raw), p.before, p.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			overlay := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlay, b, 0600); err != nil {
				t.Fatal(err)
			}
			t.Logf("intended assertion detected %s:\n%s", p.name, run(overlay))
		})
	}
}
