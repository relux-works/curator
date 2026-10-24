//go:build !unix && !windows

package envprofile

import (
	"errors"
)

func makeFIFOVectorForTest(string) error {
	return errors.New("host platform does not expose POSIX named pipes")
}

func makeWorldWritableDirectoryForTest(string) (func() error, error) {
	return nil, errors.New("host platform does not expose a writable-boundary test helper")
}

func makeWorldWritableSingleDirectoryForTest(path string) (func() error, error) {
	return makeWorldWritableDirectoryForTest(path)
}

func makeWorldReadableFileForTest(string) (func() error, error) {
	return nil, fmt.Errorf("foreign file permissions unavailable on this platform")
}
