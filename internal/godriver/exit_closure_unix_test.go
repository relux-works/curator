//go:build darwin || linux

package godriver

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Only test binaries admit this configuration; installed worker input and
// environment have no such switch. The seams execute in their own subprocess.
func init() {
	mode := os.Getenv("CURATOR_TEST_EXIT_SEAM")
	if mode == "session-failure" || mode == "worker-deadline" {
		workerRunFault = func(argv []string) error {
			if argv[0] != "build" {
				return nil
			}
			barrier := os.Getenv("CURATOR_TEST_EXIT_BARRIER")
			until := time.Now().Add(30 * time.Second)
			for time.Now().Before(until) {
				b, _ := os.ReadFile(filepath.Join(barrier, "pids"))
				if _, ready := os.Stat(filepath.Join(barrier, "fail")); len(strings.Fields(string(b))) == 2 && ready == nil {
					return diagnostic(CodeWorkerProtocolInvalid, "injected session failure with live children")
				}
				time.Sleep(10 * time.Millisecond)
			}
			return errors.New("injected session setup failed")
		}
	}
	if mode == "kill-error" {
		orphanKill = func(int, unix.Signal) error { return unix.EPERM }
		orphanSignalWait = func() { os.Exit(exitWorkerFailure) }
	}
	if mode == "delayed-signal" || mode == "delivery-bound" {
		orphanKill = func(int, unix.Signal) error { return nil }
		orphanSignalWait = func() {
			if mode == "delayed-signal" {
				_ = unix.Kill(-os.Getpid(), unix.SIGKILL)
				select {}
			}
		}
	}
	if mode == "read-error" {
		workerInputTransform = func(r io.Reader) io.Reader { return &exitReadError{Reader: r} }
	}
}

type exitReadError struct {
	io.Reader
	reads int
}

func (r *exitReadError) Read(p []byte) (int, error) {
	// Frame 1 (request), 2 (list) and 3 (permit) each have header + payload.
	// The fourth header is a parent-controlled barrier, then a non-EOF fault.
	r.reads++
	if r.reads == 7 {
		var b [1]byte
		_, _ = io.ReadFull(r.Reader, b[:])
		return 0, errors.New("injected parent read fault")
	}
	return r.Reader.Read(p)
}

// Error and delayed delivery are forced, never selected by a lucky schedule.
// The live success control requires SIGKILL; error/bound require exactly 4.
func TestWorkerCheckedKill(t *testing.T) {
	for _, mode := range []string{"native", "kill-error", "delayed-signal", "delivery-bound"} {
		t.Run(mode, func(t *testing.T) {
			exe, _ := os.Executable()
			c := exec.Command(exe, "-test.run=^TestWorkerKillHelper$")
			c.Env = append(os.Environ(), "CURATOR_TEST_EXIT_SEAM="+mode)
			c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			err := c.Run()
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("kill exit: %v", err)
			}
			status := exit.Sys().(syscall.WaitStatus)
			if mode == "kill-error" || mode == "delivery-bound" {
				if status.ExitStatus() != exitWorkerTeardownFailed {
					t.Fatalf("teardown error status=%v want 4", err)
				}
			} else if !status.Signaled() || status.Signal() != syscall.SIGKILL {
				t.Fatalf("successful kill fell through: %v", err)
			}
		})
	}
}
func TestWorkerKillHelper(_ *testing.T) {
	if os.Getenv("CURATOR_TEST_EXIT_SEAM") != "" {
		stopOrphanedWorkerDomain()
	}
}

// A forced session failure during a live build must deliver the diagnostic
// before killing the group. A non-EOF read fault must do the same even if the
// parent closes its pipe. Both cases check the acknowledged descendant PIDs
// and marker postconditions, independently of the worker's wait status.
func TestWorkerFailureKillsDescendants(t *testing.T) {
	for _, mode := range []string{"session-failure", "read-error"} {
		t.Run(mode, func(t *testing.T) {
			barrier := t.TempDir()
			rawWorkerEnvExtra = []string{"CURATOR_TEST_EXIT_SEAM=" + mode, "CURATOR_TEST_EXIT_BARRIER=" + barrier}
			defer func() { rawWorkerEnvExtra = nil }()
			s := newWorkerScenarioWithScript(t, func(f *workerFixture) stubScript {
				return stubScript{ListStdout: string(encodePackages(t, f.rootPackage())), Artifact: "artifact", BuildSleepMS: 60000, SpawnChildren: 2, SpawnPidFile: filepath.Join(barrier, "pids"), SpawnBarrier: barrier}
			})
			w := s.requireReady()
			w.send(workerMessage{Kind: kindList, Nonce: s.nonce})
			w.receive()
			w.send(workerMessage{Kind: kindPermit, Nonce: s.nonce, Permit: buildPermit(s.request.Secret, s.nonce, s.request.BuildArgv)})
			deadline := time.Now().Add(30 * time.Second)
			var ids []string
			for time.Now().Before(deadline) {
				b, _ := os.ReadFile(filepath.Join(barrier, "pids"))
				ids = strings.Fields(string(b))
				if len(ids) == 2 {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if len(ids) != 2 {
				t.Fatal("descendants not ready")
			}
			for _, id := range ids {
				pid, _ := strconv.Atoi(id)
				if !processAlive(pid) {
					t.Fatal("positive descendant not live")
				}
			}
			if mode == "session-failure" {
				if err := os.WriteFile(filepath.Join(barrier, "fail"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				_, _ = w.stdin.Write([]byte{1})
				_ = w.stdin.Close()
			}
			received := make(chan workerMessage, 1)
			go func() { m, _ := readMessage(w.stdout); received <- m }()
			select {
			case m := <-received:
				if m.Kind != kindFailure || m.Failure == nil || m.Failure.Code != CodeWorkerProtocolInvalid {
					t.Fatalf("failure frame=%+v", m)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("read error did not terminate live domain")
			}
			err := w.command.Wait()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || !exit.Sys().(syscall.WaitStatus).Signaled() {
				t.Errorf("failure did not kill worker: %v", err)
			}
			_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600)
			deadline = time.Now().Add(10 * time.Second)
			for _, id := range ids {
				pid, _ := strconv.Atoi(id)
				for processAlive(pid) && time.Now().Before(deadline) {
					time.Sleep(10 * time.Millisecond)
				}
				if processAlive(pid) {
					t.Errorf("failure descendant survived: %d", pid)
				}
				if _, err := os.Stat(filepath.Join(barrier, "survived-"+id)); !os.IsNotExist(err) {
					t.Errorf("failure descendant survival marker: %s", id)
				}
			}
		})
	}
}

// An absent response plus explicit teardown-failure status must remain typed.
func TestWorkerTeardownStatusDiagnostic(t *testing.T) {
	exe, _ := os.Executable()
	c := exec.Command(exe, "-test.run=^TestWorkerKillHelper$")
	c.Env = append(os.Environ(), "CURATOR_TEST_EXIT_SEAM=kill-error")
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	client := &workerClient{command: c, drained: true, stderr: &boundedBuffer{budget: &outputBudget{remaining: 1024}}}
	// Join here to force the exact process status before exercising classification;
	// teardown is already complete in this manager fixture.
	_ = c.Wait()
	client.finished = true
	if code := DiagnosticCode(client.protocolFailure(errProtocolClosed, "missing response")); code != CodeWorkerTeardownFailed {
		t.Fatalf("teardown status diagnostic=%s", code)
	}
}
