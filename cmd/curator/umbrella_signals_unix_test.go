//go:build unix

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/testtoolchain"
)

const umbrellaSignalProvider = `package main
import ("fmt"; "io"; "os"; "os/signal"; "strconv"; "syscall")
func main() {
 signals := make(chan os.Signal, 16)
 signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
 mode := os.Args[2]
 var input []byte
 if mode != "count" { input, _ = io.ReadAll(os.Stdin) }
 fmt.Printf("ready %d %q %q %q\n", os.Getpid(), os.Args[1:], os.Getenv("UMBRELLA_TRANSPORT"), input)
 if code, err := strconv.Atoi(mode); err == nil { os.Exit(code) }
 for s := range signals {
  n := s.(syscall.Signal)
  fmt.Printf("signal %d\n", n)
  if mode == "ignore-term" && n == syscall.SIGTERM { continue }
  if mode == "count" { continue }
  if n == syscall.SIGTERM {
   signal.Reset(n)
   _ = syscall.Kill(os.Getpid(), n)
   select {}
  }
  os.Exit(128 + int(n))
 }
}
`

func buildUmbrellaSignalFixture(t *testing.T) (string, string) {
	t.Helper()
	binary := buildHelpCLI(t)
	dir := t.TempDir()
	source := filepath.Join(dir, "provider.go")
	if err := os.WriteFile(source, []byte(umbrellaSignalProvider), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testtoolchain.Go(t), "build", "-o", filepath.Join(dir, "curator-run"), source)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake: %v\n%s", err, out)
	}
	return binary, dir
}

type umbrellaSignalLane struct {
	command *exec.Cmd
	lines   <-chan string
	done    <-chan error
	pid     int
	stderr  *bytes.Buffer
}

func startUmbrellaSignalLane(t *testing.T, binary, provider, mode string) *umbrellaSignalLane {
	t.Helper()
	sandbox := newHelpSandbox(t)
	command := exec.Command(binary, "run", "example", mode, "--", "literal argument")
	command.Dir = sandbox.cwd
	command.Env = append(sandbox.env, "PATH="+provider+string(os.PathListSeparator)+os.Getenv("PATH"), "UMBRELLA_TRANSPORT=unchanged")
	command.Stdin = strings.NewReader("raw\x00stdin\n")
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	stderr := new(bytes.Buffer)
	command.Stderr = stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	lines := make(chan string, 32)
	scanned := make(chan struct{})
	go func() {
		defer close(scanned)
		defer close(lines)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	// Drain stdout before Wait closes its pipe.
	done := make(chan error, 1)
	lane := &umbrellaSignalLane{command: command, lines: lines, done: done, stderr: stderr}
	t.Cleanup(func() {
		if lane.pid != 0 {
			_ = syscall.Kill(lane.pid, syscall.SIGKILL)
		}
		_ = command.Process.Kill()
		select {
		case <-lane.done:
		case <-time.After(5 * time.Second):
			t.Error("fixture did not stop")
		}
	})
	go func() { <-scanned; done <- command.Wait() }()
	ready := lane.line(t)
	if _, err := fmt.Sscanf(ready, "ready %d", &lane.pid); err != nil {
		t.Fatalf("provider not ready: %q", ready)
	}
	if !strings.Contains(ready, `"unchanged" "raw\x00stdin\n"`) || !strings.Contains(ready, `["example" "`+mode+`" "--" "literal argument"]`) {
		t.Fatalf("transport changed: %s", ready)
	}
	return lane
}

func (lane *umbrellaSignalLane) line(t *testing.T) string {
	t.Helper()
	select {
	case line, ok := <-lane.lines:
		if !ok {
			t.Fatal("provider stdout closed unexpectedly")
		}
		return line
	case <-time.After(5 * time.Second):
		t.Fatal("provider output timeout")
	}
	return ""
}

func (lane *umbrellaSignalLane) wait(t *testing.T, want int) {
	t.Helper()
	select {
	case err := <-lane.done:
		// Preserve the result for cleanup and subsequent assertions.
		lane.done = closedUmbrellaResult(err)
		got := 0
		if err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			got = exit.ExitCode()
		}
		if got != want {
			t.Fatalf("curator exit=%d, want %d; stderr=%s", got, want, lane.stderr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("curator did not reap provider")
	}
	if err := syscall.Kill(lane.pid, 0); err != syscall.ESRCH {
		t.Fatalf("provider PID %d still exists: %v", lane.pid, err)
	}
	lane.pid = 0
}

func closedUmbrellaResult(err error) <-chan error {
	result := make(chan error, 1)
	result <- err
	return result
}

// Positive: PID-only signals reach the compiled provider through cmdUmbrella;
// real TERM death maps to 143 and the provider is reaped. Other handlers exit
// explicitly, avoiding Go's special default SIGQUIT stack-dump behavior.
func TestUmbrellaProviderSignals(t *testing.T) {
	binary, provider := buildUmbrellaSignalFixture(t)
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT} {
		t.Run(strconv.Itoa(int(sig)), func(t *testing.T) {
			lane := startUmbrellaSignalLane(t, binary, provider, "signal")
			if err := lane.command.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			if got := lane.line(t); got != fmt.Sprintf("signal %d", sig) {
				t.Fatalf("received %q", got)
			}
			lane.wait(t, 128+int(sig))
		})
	}
}

// Negative controls paired with normal exits: ignored TERM must keep curator
// waiting, never report success; a subsequent HUP ends both with status 129.
func TestUmbrellaProviderExitControls(t *testing.T) {
	binary, provider := buildUmbrellaSignalFixture(t)
	for _, code := range []int{0, 1, 42} {
		t.Run(strconv.Itoa(code), func(t *testing.T) { startUmbrellaSignalLane(t, binary, provider, strconv.Itoa(code)).wait(t, code) })
	}
	t.Run("ignored-TERM", func(t *testing.T) {
		lane := startUmbrellaSignalLane(t, binary, provider, "ignore-term")
		if err := lane.command.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatal(err)
		}
		if got := lane.line(t); got != "signal 15" {
			t.Fatalf("received %q", got)
		}
		select {
		case err := <-lane.done:
			lane.done = closedUmbrellaResult(err)
			t.Fatalf("curator stopped after ignored TERM: %v", err)
		case <-time.After(200 * time.Millisecond):
		}
		if err := syscall.Kill(lane.pid, 0); err != nil {
			t.Fatalf("ignoring provider exited: %v", err)
		}
		if err := lane.command.Process.Signal(syscall.SIGHUP); err != nil {
			t.Fatal(err)
		}
		if got := lane.line(t); got != "signal 1" {
			t.Fatalf("received %q", got)
		}
		lane.wait(t, 129)
	})
}

// Negative: after Wait the completed process refuses signals, registration is
// removed, and a later TERM retains its normal default action without panic.
// Start failures must remove registration too.
func TestUmbrellaProviderSignalCleanup(t *testing.T) {
	if mode := os.Getenv("UMBRELLA_SIGNAL_CLEANUP"); mode != "" {
		command := exec.Command("/bin/sh", "-c", "exit 0")
		if mode == "start-failure" {
			command = exec.Command("/nonexistent-curator-provider")
		}
		err := runUmbrellaProvider(command)
		if mode == "start-failure" {
			if err == nil {
				t.Fatal("missing executable succeeded")
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if err := command.Process.Signal(syscall.SIGTERM); err == nil {
				t.Fatalf("late Signal = %v", err)
			}
		}
		if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)
		t.Fatal("TERM was swallowed after provider completion")
	}
	for _, mode := range []string{"exited", "start-failure"} {
		t.Run(mode, func(t *testing.T) {
			child := exec.Command(os.Args[0], "-test.run=^TestUmbrellaProviderSignalCleanup$")
			child.Env = append(os.Environ(), cliHelperProcess+"=1", "UMBRELLA_SIGNAL_CLEANUP="+mode)
			out, err := child.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("helper: %v %s", err, out)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGTERM {
				t.Fatalf("helper was not killed by late TERM: %v %s", err, out)
			}
		})
	}
}
