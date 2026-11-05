//go:build darwin || linux

package godriver

import (
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// These seams are assigned only in isolated test subprocesses.
var orphanKill = unix.Kill
var orphanSignalWait = func() { time.Sleep(workerShutdownGrace) }

// Never interpret an ignored kill error or a returned syscall as teardown.
// Only our verified private group may be signalled. Successful self-group
// SIGKILL normally never returns; allow a bounded delivery window if it does.
func stopOrphanedWorkerDomain() {
	group, err := unix.Getpgid(0)
	if err != nil || group != os.Getpid() {
		os.Exit(exitWorkerTeardownFailed)
	}
	if err := orphanKill(-os.Getpid(), unix.SIGKILL); err != nil {
		os.Exit(exitWorkerTeardownFailed)
	}
	orphanSignalWait()
	os.Exit(exitWorkerTeardownFailed)
}

// A handler makes fd 1/2 EPIPE return to the caller instead of terminating the
// worker before teardown. Unlike SIG_IGN, this disposition is reset on exec.
func protectWorkerPipeWrites() func() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPIPE)
	return func() { signal.Stop(c) }
}

// Even a successful launcher may leave tools alive after publishing its result.
func finishWorkerDomain() { stopOrphanedWorkerDomain() }
