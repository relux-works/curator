//go:build darwin || linux

package godriver

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/processcontext"
	"golang.org/x/sys/unix"
)

// A separate manager drives the production worker and signal context. Children
// acknowledge readiness before the manager is killed/interrupted. The release
// control proves the same live children can publish markers when left running.
func TestBuildManagerLifecycle(t *testing.T) {
	if barrier := os.Getenv("CURATOR_TEST_MANAGER_BARRIER"); barrier != "" {
		fixture := newSnapshotFixture(t)
		fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact", BuildSleepMS: 60000, SpawnChildren: 2, SpawnPidFile: filepath.Join(barrier, "pids"), SpawnBarrier: barrier})
		ctx, cancel := processcontext.Notify(t.Context())
		defer cancel()
		result, err := Build(ctx, fixture.request(ResourceLimits{Timeout: 60 * time.Second}))
		if DiagnosticCode(err) != "process_canceled" || result.Artifact.StagedPath != "" {
			t.Fatalf("manager interrupt: result=%+v err=%v", result, err)
		}
		return
	}
	for _, mode := range []string{"killed", "interrupted", "terminated", "release-control"} {
		t.Run(mode, func(t *testing.T) {
			barrier := t.TempDir()
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command(executable, "-test.run=^TestBuildManagerLifecycle$", "-test.v")
			command.Env = append(os.Environ(), "CURATOR_TEST_MANAGER_BARRIER="+barrier, "TMPDIR="+barrier)
			output, err := os.Create(filepath.Join(barrier, "manager.log"))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := output.Close(); err != nil {
					t.Errorf("close test file: %v", err)
				}
			}()
			command.Stdout, command.Stderr = output, output
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			waited := make(chan error, 1)
			go func() { waited <- command.Wait() }()
			joined := false
			owned := []int{command.Process.Pid}
			t.Cleanup(func() {
				for _, pid := range owned {
					_ = syscall.Kill(pid, syscall.SIGKILL)
				}
				_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600)
				if !joined {
					select {
					case <-waited:
					case <-time.After(10 * time.Second):
						t.Error("manager cleanup did not join")
					}
				}
			})
			guard, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			var children []string
			for guard.Err() == nil {
				raw, _ := os.ReadFile(filepath.Join(barrier, "pids"))
				children = strings.Fields(string(raw))
				if len(children) == 2 {
					break
				}
				select {
				case err := <-waited:
					joined = true
					raw, _ := os.ReadFile(filepath.Join(barrier, "manager.log"))
					t.Fatalf("manager ended before readiness: %v\n%s", err, raw)
				default:
				}
				time.Sleep(10 * time.Millisecond)
			}
			if len(children) != 2 {
				raw, _ := os.ReadFile(filepath.Join(barrier, "manager.log"))
				t.Fatalf("manager descendant readiness: %v\n%s", children, raw)
			}
			for _, raw := range children {
				pid, err := strconv.Atoi(raw)
				if err != nil {
					t.Fatal(err)
				}
				owned = append(owned, pid)
				if !processAlive(pid) {
					t.Fatalf("child %d was not live", pid)
				}
			}
			// The compiler is the children's parent; all inherit the verified worker's
			// private process group. Record each PID for postconditions and cleanup.
			raw, err := exec.Command("ps", "-o", "ppid=", "-p", children[0]).Output()
			if err != nil {
				t.Fatal(err)
			}
			compiler, err := strconv.Atoi(strings.TrimSpace(string(raw)))
			if err != nil {
				t.Fatal(err)
			}
			group, err := unix.Getpgid(compiler)
			if err != nil {
				t.Fatal(err)
			}
			if group == command.Process.Pid || group == unix.Getpgrp() {
				t.Fatal("worker domain is not private")
			}
			owned = append(owned, compiler, group)
			if mode == "release-control" {
				if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0600); err != nil {
					t.Fatal(err)
				}
				for _, pid := range children {
					for guard.Err() == nil {
						if _, err := os.Stat(filepath.Join(barrier, "survived-"+pid)); err == nil {
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
					if _, err := os.Stat(filepath.Join(barrier, "survived-"+pid)); err != nil {
						t.Fatalf("live release control did not publish: %v", err)
					}
				}
			}
			signal := syscall.SIGINT
			if mode == "killed" {
				signal = syscall.SIGKILL
			}
			if mode == "terminated" {
				signal = syscall.SIGTERM
			}
			if err := command.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-waited:
				joined = true
				if mode != "killed" && err != nil {
					raw, _ := os.ReadFile(filepath.Join(barrier, "manager.log"))
					t.Fatalf("manager signal result: %v\n%s", err, raw)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("manager did not cancel promptly")
			}
			for _, pid := range owned[1:] {
				deadline := time.Now().Add(10 * time.Second)
				for processAlive(pid) && time.Now().Before(deadline) {
					time.Sleep(10 * time.Millisecond)
				}
				if processAlive(pid) {
					t.Fatalf("manager %s left process %d running", mode, pid)
				}
			}
			if mode != "release-control" {
				_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600)
				for _, pid := range children {
					if _, err := os.Stat(filepath.Join(barrier, "survived-"+pid)); !os.IsNotExist(err) {
						t.Fatalf("manager %s descendant survived: %s %v", mode, pid, err)
					}
				}
			}
			t.Logf("manager=%s; compiler, worker and two descendants terminated", mode)
		})
	}
}
