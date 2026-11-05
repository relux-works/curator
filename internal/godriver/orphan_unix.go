//go:build darwin || linux

package godriver

import (
	"golang.org/x/sys/unix"
	"os"
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
