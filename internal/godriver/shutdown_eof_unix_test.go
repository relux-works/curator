//go:build darwin || linux

package godriver

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

// A queued shutdown cannot hide parent EOF while a compiler is still running.
// The live-pipe control releases the same acknowledged children and requires
// their markers; the EOF case requires worker/children death and no markers.
func TestWorkerEOFWhileShutdownQueued(t *testing.T) {
	for _, endPipe := range []bool{false, true} {
		name := "live-control"
		if endPipe {
			name = "parent-EOF"
		}
		t.Run(name, func(t *testing.T) {
			barrier := t.TempDir()
			scenario := newWorkerScenarioWithScript(t, func(f *workerFixture) stubScript {
				return stubScript{ListStdout: string(encodePackages(t, f.rootPackage())), Artifact: "artifact", BuildSleepMS: 60000, SpawnChildren: 2, SpawnPidFile: filepath.Join(barrier, "pids"), SpawnBarrier: barrier}
			})
			worker := scenario.requireReady()
			worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
			if message := worker.receive(); message.Kind != kindListResult || message.Result == nil || message.Result.ExitCode != 0 {
				t.Fatalf("list result: %+v", message)
			}
			worker.send(workerMessage{Kind: kindPermit, Nonce: scenario.nonce, Permit: buildPermit(scenario.request.Secret, scenario.nonce, scenario.request.BuildArgv)})
			deadline := time.Now().Add(30 * time.Second)
			var children []string
			for time.Now().Before(deadline) {
				raw, _ := os.ReadFile(filepath.Join(barrier, "pids"))
				children = strings.Fields(string(raw))
				if len(children) == 2 {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if len(children) != 2 {
				t.Fatal("compiler descendants did not acknowledge readiness")
			}
			pids := []int{worker.command.Process.Pid}
			for _, raw := range children {
				pid, err := strconv.Atoi(raw)
				if err != nil {
					t.Fatal(err)
				}
				pids = append(pids, pid)
			}
			for _, pid := range pids {
				if !processAlive(pid) {
					t.Fatalf("positive readiness: process %d not live", pid)
				}
			}
			worker.send(workerMessage{Kind: kindShutdown, Nonce: scenario.nonce})
			if endPipe {
				if err := worker.stdin.Close(); err != nil {
					t.Fatal(err)
				}
				closed := make(chan error, 1)
				go func() { _, err := readMessage(worker.stdout); closed <- err }()
				select {
				case err := <-closed:
					if !errors.Is(err, errProtocolClosed) {
						t.Fatalf("EOF worker response: %v", err)
					}
				case <-time.After(10 * time.Second):
					t.Fatal("queued shutdown hid EOF: worker did not exit")
				}
				err := worker.command.Wait()
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Errorf("EOF did not kill worker: %v", err)
				}
				var status syscall.WaitStatus
				ok := false
				if exit != nil {
					status, ok = exit.Sys().(syscall.WaitStatus)
				}
				if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
					t.Errorf("EOF worker exit: %v", err)
				}
				deadline = time.Now().Add(10 * time.Second)
				for _, pid := range pids {
					for processAlive(pid) && time.Now().Before(deadline) {
						time.Sleep(10 * time.Millisecond)
					}
					if processAlive(pid) {
						t.Errorf("queued shutdown hid EOF: process %d survived", pid)
					}
				}
			}
			if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			for _, raw := range children {
				marker := filepath.Join(barrier, "survived-"+raw)
				if !endPipe {
					deadline = time.Now().Add(10 * time.Second)
					for time.Now().Before(deadline) {
						if _, err := os.Stat(marker); err == nil {
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
					if _, err := os.Stat(marker); err != nil {
						t.Fatalf("live control did not publish: %v", err)
					}
				} else if _, err := os.Stat(marker); !os.IsNotExist(err) {
					t.Errorf("EOF descendant published marker: %v", err)
				}
			}
		})
	}
}
