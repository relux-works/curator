//go:build unix

package pathboundary

import "os"

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
