//go:build !windows

package godriver

import "syscall"

// processAlive reports whether the process still exists. Signal 0 performs the
// permission and existence check only.
func processAlive(pid int) bool { return syscall.Kill(pid, 0) == nil }
