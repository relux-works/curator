//go:build darwin || linux

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Positive/negative TTY pair: a terminal-generated Ctrl-C and a PID-only INT
// each reach the compiled provider exactly once, on a real controlling PTY.
// Provider stop/continue follows the wrapper, and foreground is restored after reaping.
func TestUmbrellaProviderTerminal(t *testing.T) {
	if provider := os.Getenv("UMBRELLA_PTY_PROVIDER"); provider != "" {
		command := exec.Command(provider, "example", "count")
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := runUmbrellaProvider(command)
		var exit *umbrellaProcessExit
		if !errors.As(err, &exit) || exit.ExitCode() != 137 {
			t.Fatalf("provider exit: %v", err)
		}
		fg, err := unix.IoctlGetInt(int(os.Stdin.Fd()), unix.TIOCGPGRP)
		if err != nil || fg != syscall.Getpgrp() {
			t.Fatalf("foreground not restored: %d %v", fg, err)
		}
		fmt.Println("restored")
		return
	}
	_, providerDir := buildUmbrellaSignalFixture(t)
	master, slave := openUmbrellaPTY(t)
	child := exec.Command(os.Args[0], "-test.run=^TestUmbrellaProviderTerminal$")
	// TERM=dumb prevents termenv's startup OSC queries from consuming the PTY ready timeout.
	child.Env = append(os.Environ(), cliHelperProcess+"=1", "UMBRELLA_PTY_PROVIDER="+providerDir+"/curator-run", "TERM=dumb")
	child.Stdin, child.Stdout, child.Stderr = slave, slave, slave
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	var pid int
	done := make(chan error, 1)
	waitStarted := false
	wait := func() {
		if !waitStarted {
			waitStarted = true
			go func() { done <- child.Wait() }()
		}
	}
	t.Cleanup(func() {
		_ = child.Process.Signal(syscall.SIGCONT)
		if pid != 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
		_ = child.Process.Kill()
		wait()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("PTY helper did not stop")
		}
	})
	lines := make(chan string, 32)
	go func() {
		scanner := bufio.NewScanner(master)
		for scanner.Scan() {
			lines <- strings.TrimSpace(scanner.Text())
		}
		close(lines)
	}()
	next := func() string {
		t.Helper()
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatal("PTY closed")
			}
			return line
		case <-time.After(5 * time.Second):
			state, _ := exec.Command("ps", "-o", "pid=,ppid=,stat=", "-p", fmt.Sprintf("%d,%d", child.Process.Pid, pid)).Output()
			fg, _ := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPGRP)
			t.Fatalf("PTY output timeout; foreground=%d state=%s", fg, state)
		}
		return ""
	}
	for {
		line := next()
		if strings.HasPrefix(line, "ready ") {
			if _, err := fmt.Sscanf(line, "ready %d", &pid); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	fg, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPGRP)
	if err != nil || fg != pid || pid == child.Process.Pid {
		t.Fatalf("foreground=%d provider=%d helper=%d error=%v", fg, pid, child.Process.Pid, err)
	}
	for _, terminal := range []bool{true, false} {
		if terminal {
			_, err = master.Write([]byte{3})
		} else {
			err = child.Process.Signal(syscall.SIGINT)
		}
		if err != nil {
			t.Fatal(err)
		}
		if line := next(); !strings.Contains(line, "signal 2") {
			t.Fatalf("INT output: %q", line)
		}
		select {
		case line := <-lines:
			t.Fatalf("duplicate or unexpected signal output: %q", line)
		case <-time.After(200 * time.Millisecond):
		}
	}

	// A terminal-owned provider stop must also stop the curator wrapper, so
	// its supervising shell can observe and resume the complete foreground job.
	if err := syscall.Kill(pid, syscall.SIGSTOP); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		var status syscall.WaitStatus
		stopped, err := syscall.Wait4(child.Process.Pid, &status, syscall.WNOHANG|syscall.WUNTRACED, nil)
		if err != nil {
			t.Fatal(err)
		}
		if stopped != 0 {
			// BSD WaitStatus.Stopped classifies the SIGSTOP encoding as
			// continued, even though this wait requested no WCONTINUED.
			if status != syscall.WaitStatus(uint32(syscall.SIGSTOP)<<8|0x7f) {
				time.Sleep(50 * time.Millisecond)
				var output []string
			drain:
				for i := 0; i < 32; i++ {
					select {
					case line, ok := <-lines:
						if !ok {
							break drain
						}
						output = append(output, line)
					default:
						break drain
					}
				}
				t.Fatalf("curator stop status: %v; output=%q", status, output)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("curator did not follow provider stop")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := child.Process.Signal(syscall.SIGCONT); err != nil {
		t.Fatal(err)
	}
	wait()
	// CONT delivery is asynchronous. Await the provider's actual receipt before
	// testing a later INT: on Darwin an INT sent while stopped can be lost on
	// CONT even for a directly supervised child, without curator in the path.
	if line := next(); line != "continued" {
		t.Fatalf("CONT receipt: %q", line)
	}
	if err := child.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	if line := next(); !strings.Contains(line, "signal 2") {
		t.Fatalf("resumed INT output: %q", line)
	}
	fg, err = unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPGRP)
	if err != nil || fg != pid {
		t.Fatalf("resumed foreground=%d provider=%d error=%v", fg, pid, err)
	}
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	if line := next(); line != "restored" {
		t.Fatalf("restoration: %q", line)
	}
	select {
	case err := <-done:
		done <- err
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PTY helper did not exit")
	}
	if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("PTY provider remains: %v", err)
	}
	pid = 0
}
