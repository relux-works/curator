//go:build !unix

package main

import "errors"

// makeFifo answers the STUB_MKFIFO_TARGETS probe where the platform has
// no FIFO primitive. The FIFO rows are Linux-only; this stub keeps the
// test double portable.
func makeFifo(_ string) error {
	return errors.New("named pipes are unavailable on this platform")
}
