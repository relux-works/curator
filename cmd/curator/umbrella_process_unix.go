//go:build unix

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

type umbrellaProcessExit struct{ status syscall.WaitStatus }

func (e *umbrellaProcessExit) Error() string { return fmt.Sprintf("provider status: %d", e.status) }
func (e *umbrellaProcessExit) ExitCode() int {
	if e.status.Signaled() {
		return 128 + int(e.status.Signal())
	}
	return e.status.ExitStatus()
}

// runUmbrellaProvider holds signals before Start and waits even if the provider
// ignores them. The provider owns forwarding to any children it launches.
func runUmbrellaProvider(command *exec.Cmd) error {
	signals := make(chan os.Signal, 16)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGCHLD, syscall.SIGCONT)
	defer signal.Stop(signals)

	// A terminal sends Ctrl-C to its foreground group. Give that group to the
	// provider so curator cannot receive and then duplicate a terminal signal.
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var terminal *os.File
	var original int
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		defer tty.Close()
		original, err = unix.IoctlGetInt(int(tty.Fd()), unix.TIOCGPGRP)
		if err == nil && original == syscall.Getpgrp() {
			terminal = tty
			command.SysProcAttr.Foreground = true
			command.SysProcAttr.Ctty = int(tty.Fd())
			// Restoration is a background-group ioctl. Retain a pre-existing ignore.
			ignored := signal.Ignored(syscall.SIGTTOU)
			signal.Ignore(syscall.SIGTTOU)
			defer func() {
				_ = unix.IoctlSetPointerInt(int(tty.Fd()), unix.TIOCSPGRP, original)
				if !ignored {
					signal.Reset(syscall.SIGTTOU)
				}
			}()
		}
	}
	if err := command.Start(); err != nil {
		return err
	}

	// One loop owns both signalling and reaping. A PID cannot be reused while
	// its child is still unreaped; after the terminal wait status we send no
	// further signals. WUNTRACED preserves shell stop/fg semantics now that the
	// provider owns a separate foreground group.
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(command.Process.Pid, &status, syscall.WNOHANG|syscall.WUNTRACED, nil)
		if err == syscall.EINTR {
			continue
		}
		if err != nil || (pid != 0 && (status.Exited() || status.Signaled())) {
			// Wait4 reaps the child. Release its handle before Cmd.Wait closes pipes
			// and joins I/O copying; its process wait error is expected here.
			_ = command.Process.Release()
			_ = command.Wait()
			if err != nil {
				return os.NewSyscallError("wait4", err)
			}
			if status != 0 {
				return &umbrellaProcessExit{status: status}
			}
			return nil
		}
		// WCONTINUED is not requested, so a nonterminal report is a stop.
		// BSD WaitStatus.Stopped excludes SIGSTOP; checking terminal states
		// also handles that real stop without treating it as a child exit.
		if pid != 0 {
			if terminal != nil {
				_ = unix.IoctlSetPointerInt(int(terminal.Fd()), unix.TIOCSPGRP, original)
			}
			// SIGSTOP cannot be ignored: the supervising shell observes curator stop
			// and can resume it with fg. The queued CONT restores provider ownership.
			_ = syscall.Kill(os.Getpid(), syscall.SIGSTOP)
		}
		s := <-signals
		if s == syscall.SIGCHLD {
			continue
		}
		if s == syscall.SIGCONT && terminal != nil {
			_ = unix.IoctlSetPointerInt(int(terminal.Fd()), unix.TIOCSPGRP, command.Process.Pid)
		}
		_ = command.Process.Signal(s)
	}
}
