//go:build unix

package main

import "golang.org/x/sys/unix"

// makeFifo creates a named pipe for the STUB_MKFIFO_TARGETS probe, so
// Linux tests prove whether the write confinement allows FIFO creation
// at the target.
func makeFifo(path string) error {
	return unix.Mkfifo(path, 0o644)
}
