//go:build unix

package testhomeguard

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// Signals target the real test child, never the observer: a diagnostic
// SIGQUIT must dump the blocked child, and an operator's INT/TERM/HUP must
// reach the child whose wait status the caller observes. The observer
// survives to join it and check the postimage; its own death by one of
// these signals is preserved by re-raising the child's signal afterwards.
// SIGKILL/observer death remain residuals.
func forwardDiagnosticSignals(child *os.Process) func() {
	ch := make(chan os.Signal, 4)
	done := make(chan struct{})
	signal.Notify(ch, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for {
			select {
			case s := <-ch:
				_ = child.Signal(s)
			case <-done:
				return
			}
		}
	}()
	// A session-leader supervisor owns a controlling terminal its caller
	// job-controls through this process: when the observed child stops,
	// stop with it so the caller's wait observes the stop, and pass the
	// caller's continue through on resume. Other supervisors never stop.
	if sid, err := unix.Getsid(0); err == nil && sid == os.Getpid() {
		go func() {
			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()
			stopped := false
			for {
				select {
				case <-ticker.C:
					now := childStopped(child.Pid)
					if now && !stopped {
						stopped = true
						_ = syscall.Kill(os.Getpid(), syscall.SIGSTOP)
						_ = syscall.Kill(child.Pid, syscall.SIGCONT)
					}
					if !now {
						stopped = false
					}
				case <-done:
					return
				}
			}
		}()
	}
	return func() { signal.Stop(ch); close(done) }
}

func signaledExit(err error) (os.Signal, bool) {
	exit, ok := err.(*exec.ExitError)
	if !ok {
		return nil, false
	}
	status, ok := exit.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() {
		return nil, false
	}
	return status.Signal(), true
}

// reraiseSignal restores the default disposition and kills this supervisor
// with the child's signal, preserving the caller's observed wait status.
func reraiseSignal(sig os.Signal) {
	signal.Reset(sig)
	if s, ok := sig.(syscall.Signal); ok {
		_ = syscall.Kill(os.Getpid(), s)
	}
}
