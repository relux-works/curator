//go:build windows

package transaction

import (
	"os"

	"github.com/relux-works/curator/internal/privatedir"
)

func makePublicationDirectory(path string, _ os.FileMode) error {
	return privatedir.Make(path)
}

func createPublicationFile(path string, mode os.FileMode) (*os.File, error) {
	return privatedir.Create(path, mode)
}
