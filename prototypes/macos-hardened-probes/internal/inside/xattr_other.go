//go:build !darwin

package inside

import "syscall"

// setxattr is unimplemented outside darwin. This module is a macOS prototype;
// the stub exists so the package still compiles when a portable test runs on
// another host, and it reports ENOSYS so a caller can never mistake a missing
// implementation for a kernel refusal.
func setxattr(path, name string, value []byte) error {
	return syscall.ENOSYS
}
