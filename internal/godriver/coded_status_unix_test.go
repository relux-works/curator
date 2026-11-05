//go:build darwin || linux

package godriver

// A coded cause plus status 4 must
// stay typed as teardown failure: proves the protocolFailure precedence only.

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestWorkerCodedCauseKeepsTeardownStatus(t *testing.T) {
	exe, _ := os.Executable()
	c := exec.Command(exe, "-test.run=^TestWorkerKillHelper$")
	c.Env = append(os.Environ(), "CURATOR_TEST_EXIT_SEAM=kill-error")
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	client := &workerClient{command: c, drained: true, stderr: &boundedBuffer{budget: &outputBudget{remaining: 1024}}}
	_ = c.Wait()
	client.finished = true
	cause := diagnostic(CodeWorkerProtocolInvalid, "coded cause")
	if code := DiagnosticCode(client.protocolFailure(cause, "missing response")); code != CodeWorkerTeardownFailed {
		t.Fatalf("coded cause diagnostic=%s", code)
	}
}
