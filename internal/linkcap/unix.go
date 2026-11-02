//go:build unix

// Package linkcap classifies actual symlink syscall limitations. Permission,
// I/O and metadata errors are not evidence that copying is authorized.
package linkcap

import (
	"errors"
	"syscall"
)

// Unavailable classifies only supported filesystem link-capability errors.
func Unavailable(err error) bool {
	return errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.ENOSYS)
}
