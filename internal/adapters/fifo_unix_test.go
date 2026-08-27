//go:build unix

package adapters

import "golang.org/x/sys/unix"

// makeFIFO creates a named pipe, which is the cheapest portable stand-in for
// "a managed path that is neither a directory, a regular file, nor a link".
func makeFIFO(path string) error {
	return unix.Mkfifo(path, 0o600)
}
