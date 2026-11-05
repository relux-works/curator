//go:build darwin || linux

package godriver

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// The exact worker protection function must not give an exec'd native shell an
// inherited SIG_IGN. This models exec disposition, not an entire compiler run.
func TestWorkerPipeHandlerDoesNotLeakToExec(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	c := exec.Command(exe, "-test.run=^TestWorkerPipeExecHelper$")
	c.Env = append(os.Environ(), "CURATOR_TEST_PIPE_EXEC=parent")
	if b, err := c.CombinedOutput(); err != nil {
		t.Fatalf("pipe exec disposition: %v\n%s", err, b)
	}
}

// Parent and child re-executions isolate signal registration from other tests.
func TestWorkerPipeExecHelper(t *testing.T) {
	switch os.Getenv("CURATOR_TEST_PIPE_EXEC") {
	case "parent":
		stop := protectWorkerPipeWrites()
		defer stop()
		c := exec.Command("/bin/sh", "-c", "kill -PIPE $$; exit 33")
		err := c.Run()
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 33 {
			t.Fatal("exec child inherited ignored SIGPIPE")
		}
		if !errors.As(err, &exit) || !exit.Sys().(syscall.WaitStatus).Signaled() || exit.Sys().(syscall.WaitStatus).Signal() != syscall.SIGPIPE {
			t.Fatalf("unexpected exec disposition: %v", err)
		}

	}
}
