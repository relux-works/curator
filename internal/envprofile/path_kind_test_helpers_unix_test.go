//go:build unix

package envprofile

import (
	"fmt"
	"os"
	"syscall"
)

func makeFIFOVectorForTest(path string) error {
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		return fmt.Errorf("mkfifo: %w", err)
	}
	return nil
}

func makeWorldWritableDirectoryForTest(path string) (func() error, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o777); err != nil {
		return nil, err
	}
	return func() error { return os.Chmod(path, info.Mode().Perm()) }, nil
}
