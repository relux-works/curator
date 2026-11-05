//go:build darwin || linux

package godriver

import (
	"os"

	"golang.org/x/sys/unix"
)

// Kill the verified private worker group, including this process. Killing only
// exec.Cmd.Process would leave compiler tool descendants alive after manager death.
func stopOrphanedWorkerDomain() {
	_ = unix.Kill(-os.Getpid(), unix.SIGKILL)
	os.Exit(exitWorkerFailure)
}
