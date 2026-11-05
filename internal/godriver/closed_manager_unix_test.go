//go:build darwin || linux

package godriver

// Closed-manager tests use real framed pipes and acknowledged tool processes.

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func exitLive(t *testing.T, seam string, sleepMS int) (*workerScenario, *rawWorker, string, []int) {
	t.Helper()
	barrier := t.TempDir()
	if seam != "" {
		rawWorkerEnvExtra = []string{"CURATOR_TEST_EXIT_SEAM=" + seam, "CURATOR_TEST_EXIT_BARRIER=" + barrier}
		t.Cleanup(func() { rawWorkerEnvExtra = nil })
	}
	s := newWorkerScenarioWithScript(t, func(f *workerFixture) stubScript {
		return stubScript{ListStdout: string(encodePackages(t, f.rootPackage())), Artifact: "artifact", BuildSleepMS: sleepMS, SpawnChildren: 2, SpawnPidFile: filepath.Join(barrier, "pids"), SpawnBarrier: barrier}
	})
	if seam == "worker-deadline" {
		s.request.Limits.TimeoutMillis = 4000
	}
	w := s.requireReady()
	w.send(workerMessage{Kind: kindList, Nonce: s.nonce})
	w.receive()
	w.send(workerMessage{Kind: kindPermit, Nonce: s.nonce, Permit: buildPermit(s.request.Secret, s.nonce, s.request.BuildArgv)})
	deadline := time.Now().Add(60 * time.Second)
	var pids []int
	for time.Now().Before(deadline) {
		b, _ := os.ReadFile(filepath.Join(barrier, "pids"))
		f := strings.Fields(string(b))
		if len(f) == 2 {
			for _, x := range f {
				p, _ := strconv.Atoi(x)
				pids = append(pids, p)
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(pids) != 2 {
		t.Fatal("descendants not ready")
	}
	for _, pid := range pids {
		if !processAlive(pid) {
			t.Fatal("acknowledged descendant is not live")
		}
	}
	t.Cleanup(func() {
		_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600)
		for _, pid := range pids {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	return s, w, barrier, pids
}

func exitStatus(err error) string {
	if err == nil {
		return "exit 0"
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		st := exit.Sys().(syscall.WaitStatus)
		if st.Signaled() {
			return "signal " + st.Signal().String()
		}
		return "exit " + strconv.Itoa(st.ExitStatus())
	}
	return err.Error()
}

func exitReport(t *testing.T, w *rawWorker, barrier string, pids []int) (status string, alive int, markers int) {
	t.Helper()
	status = exitStatus(w.command.Wait())
	time.Sleep(1500 * time.Millisecond) // give a successful group kill time to reap
	for _, pid := range pids {
		if processAlive(pid) {
			alive++
		}
	}
	_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600)
	time.Sleep(1500 * time.Millisecond)
	for _, pid := range pids {
		if _, err := os.Stat(filepath.Join(barrier, "survived-"+strconv.Itoa(pid))); err == nil {
			markers++
		}
	}
	t.Logf("domain result %s: worker=%q descendants_alive=%d survival_markers=%d", t.Name(), status, alive, markers)
	if status != "signal killed" || alive != 0 || markers != 0 {
		t.Errorf("domain closure: status=%s alive=%d markers=%d", status, alive, markers)
	}
	return
}

// Closed output and broken input model manager loss after acknowledged live
// descendants. The session-failure row holds stdin open until exit, preventing
// EOF cleanup from masking main-path teardown. Header-only uses no seam;
// held-stdout and plain EOF are controls for the same domain.
func TestWorkerClosedManagerPipes(t *testing.T) {
	for _, mode := range []string{"header", "held-stdout", "plain-EOF", "session-failure", "read-error"} {
		t.Run(mode, func(t *testing.T) {
			seam := ""
			if mode == "session-failure" || mode == "read-error" {
				seam = mode
			}
			_, w, barrier, pids := exitLive(t, seam, 60000)
			if mode != "held-stdout" {
				_ = w.stdout.Close()
			}
			switch mode {
			case "header", "held-stdout":
				w.sendFrame(16, nil)
			case "read-error":
				_, _ = w.stdin.Write([]byte{1})
			case "session-failure":
				w.sendFrame(16, nil)
				if err := os.WriteFile(filepath.Join(barrier, "fail"), nil, 0600); err != nil {
					t.Fatal(err)
				}
				// Hold input until the main failure path exits; EOF cannot mask it.
				exitReport(t, w, barrier, pids)
				_ = w.stdin.Close()
				return
			}
			_ = w.stdin.Close()
			if mode == "held-stdout" {
				m, err := readMessage(w.stdout)
				if err != nil || m.Kind != kindFailure {
					t.Fatalf("control frame: %v %+v", err, m)
				}
			}
			exitReport(t, w, barrier, pids)
		})
	}
}

// The manager stays alive and performs no cancellation. The main worker is
// blocked in a test fault barrier; only its deadline watcher can stop tools.
func TestWorkerIndependentDeadline(t *testing.T) {
	_, w, barrier, pids := exitLive(t, "worker-deadline", 60000)
	// The isolated worker's deadline is shortened before context creation by
	// the raw request. No parent watchdog or pipe close supplies teardown.

	expired := make(chan struct{})
	timer := time.AfterFunc(12*time.Second, func() {
		close(expired)
		_ = syscall.Kill(-w.command.Process.Pid, syscall.SIGKILL)
	})
	defer timer.Stop()
	exitReport(t, w, barrier, pids)
	select {
	case <-expired:
		t.Fatal("worker deadline watcher did not terminate domain")
	default:
	}

}

// A completed launcher leaves tools alive. Shutdown must terminate them even
// when the manager never performs its own domain cleanup or closes stdin.
func TestWorkerSuccessfulShutdownKillsDescendants(t *testing.T) {
	s, w, barrier, pids := exitLive(t, "", 0)
	if m := w.receive(); m.Kind != kindBuildResult {
		t.Fatalf("result=%s", m.Kind)
	}
	w.send(workerMessage{Kind: kindShutdown, Nonce: s.nonce})
	exitReport(t, w, barrier, pids)
}
