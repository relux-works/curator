//go:build !windows

package transaction

import "os"

func makePublicationDirectory(path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}

func createPublicationFile(path string, mode os.FileMode) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode) // #nosec G304 -- exclusive transaction-owned staging path.
}
