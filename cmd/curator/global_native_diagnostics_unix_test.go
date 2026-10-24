//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Timeout diagnostics run only after the existing deadline has failed. Capture
// only this CLI's descendants, then join the CLI before reading its output.
// SIGQUIT supplies Go stacks without adding a production tracing hook. Any
// surviving fixture descendants are killed by their recorded individual PIDs.
func (child *nativeCommand) timeoutDiagnostic() string {
	tree, diagnostic := nativeOwnedProcesses(child.cmd.Process.Pid)
	_ = child.cmd.Process.Signal(syscall.SIGQUIT)
	select {
	case <-child.done:
	case <-time.After(5 * time.Second):
		_ = child.cmd.Process.Kill()
		<-child.done
	}
	child.finished = true
	for i := len(tree) - 1; i >= 0; i-- {
		tree[i].killIfCurrent()
	}
	return fmt.Sprintf("production CLI did not finish after release\nowned process trace:\n%s\njoined CLI output and stacks:\n%s", diagnostic, child.output.String())
}

type nativeOwnedProcess struct {
	pid     int
	started string
}

func (process nativeOwnedProcess) killIfCurrent() bool {
	if nativeProcessStart(process.pid) != process.started {
		return false
	}
	child, err := os.FindProcess(process.pid)
	return err == nil && child.Kill() == nil
}

func nativeProcessStart(pid int) string {
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil {
		return ""
	}
	return strings.Join(strings.Fields(string(output)), " ")
}

func nativeOwnedProcesses(root int) ([]nativeOwnedProcess, string) {
	output, err := exec.Command("ps", "-axo", "pid=,ppid=,stat=,time=,lstart=,comm=").Output()
	if err != nil {
		return nil, fmt.Sprintf("process trace failed: %v", err)
	}
	descendants := map[int]bool{root: true}
	rows := strings.Split(string(output), "\n")
	pids := []nativeOwnedProcess{}
	var trace strings.Builder
	for changed := true; changed; {
		changed = false
		for _, row := range rows {
			fields := strings.Fields(row)
			if len(fields) < 10 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			parent, _ := strconv.Atoi(fields[1])
			if pid == root || !descendants[parent] || descendants[pid] {
				continue
			}
			descendants[pid] = true
			changed = true
			pids = append(pids, nativeOwnedProcess{pid: pid, started: strings.Join(fields[4:9], " ")})
			fmt.Fprintf(&trace, "pid=%d parent=%d state=%s cpu=%s command=%s\n", pid, parent, fields[2], fields[3], filepath.Base(strings.Join(fields[9:], " ")))
		}
	}
	return pids, trace.String()
}

// A normally completed child succeeds through wait. A blocked child supplies
// joined Go stacks, and the real watchdog must fail its subprocess test with
// that diagnostic. Output is read only after its exec writer has joined.
func TestNativeCommandTimeoutDiagnostics(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		t.Run(fmt.Sprint(blocked), func(t *testing.T) {
			child := &nativeCommand{done: make(chan error, 1)}
			child.cmd = exec.Command(os.Args[0], "-test.run=^TestNativeCommandDiagnosticProcess$")
			child.cmd.Env = append(os.Environ(), cliHelperProcess+"=1", "CURATOR_NATIVE_DIAGNOSTIC_HELPER=1")
			if blocked {
				child.cmd.Env = append(child.cmd.Env, "CURATOR_NATIVE_DIAGNOSTIC_BLOCK=1")
			}
			ready, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer ready.Close()
			defer writer.Close()
			child.cmd.ExtraFiles = []*os.File{writer}
			child.cmd.Stdout = &child.output
			child.cmd.Stderr = &child.output
			if err := child.cmd.Start(); err != nil {
				t.Fatal(err)
			}
			go func() { child.done <- child.cmd.Wait() }()
			t.Cleanup(func() {
				if !child.finished {
					_ = child.cmd.Process.Kill()
					<-child.done
					child.finished = true
				}
			})
			if !blocked {
				child.wait(t)
				return
			}
			_ = writer.Close()
			var started [1]byte
			if _, err := io.ReadFull(ready, started[:]); err != nil || started[0] != 1 {
				t.Fatalf("diagnostic helper did not reach readiness: %v", err)
			}
			if os.Getenv("CURATOR_NATIVE_DIAGNOSTIC_WATCHDOG") == "1" {
				child.waitWithin(t, time.Millisecond)
				t.Fatal("watchdog returned success")
			}
			diagnostic := child.timeoutDiagnostic()
			if !child.finished || !strings.Contains(diagnostic, "production CLI did not finish after release") || !strings.Contains(diagnostic, "TestNativeCommandDiagnosticProcess") || !strings.Contains(diagnostic, "SIGQUIT") {
				t.Fatalf("timeout lost failure or blocked child stack: %s", diagnostic)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			watchdog := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeCommandTimeoutDiagnostics$/^true$")
			watchdog.Env = append(os.Environ(), cliHelperProcess+"=1", "CURATOR_NATIVE_DIAGNOSTIC_WATCHDOG=1")
			output, err := watchdog.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || ctx.Err() != nil ||
				!bytes.Contains(output, []byte("production CLI did not finish after release")) ||
				!bytes.Contains(output, []byte("SIGQUIT")) ||
				bytes.Contains(output, []byte("watchdog returned success")) {
				t.Fatalf("timeout lost failing verdict: %v\n%s", err, output)
			}
		})
	}
}

func TestNativeCommandDiagnosticProcess(_ *testing.T) {
	if os.Getenv("CURATOR_NATIVE_DIAGNOSTIC_HELPER") != "1" {
		return
	}
	ready := os.NewFile(3, "diagnostic-ready")
	if ready == nil {
		os.Exit(94)
	}
	if _, err := ready.Write([]byte{1}); err != nil {
		os.Exit(94)
	}
	_ = ready.Close()
	if os.Getenv("CURATOR_NATIVE_DIAGNOSTIC_BLOCK") == "1" {
		time.Sleep(time.Minute)
	}
}

// A stale start identity cannot authorize killing even a fixture-owned PID;
// the matching identity can terminate that same live fixture and it is joined.
func TestNativeCommandCleanupIdentity(t *testing.T) {
	command := exec.Command("sleep", "60")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = command.Process.Kill(); _ = command.Wait() }()
	started := nativeProcessStart(command.Process.Pid)
	if started == "" {
		t.Fatal("fixture process start unavailable")
	}
	stale := nativeOwnedProcess{pid: command.Process.Pid, started: started + " stale"}
	if stale.killIfCurrent() || command.Process.Signal(syscall.Signal(0)) != nil {
		t.Fatal("stale process identity admitted")
	}
	current := nativeOwnedProcess{pid: command.Process.Pid, started: started}
	if !current.killIfCurrent() {
		t.Fatal("matching process identity refused")
	}
}
