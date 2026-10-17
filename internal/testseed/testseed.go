// Package testseed constructs filesystem-readable native seed fixtures. It must
// not be used to repair an input at resolution time or to construct negative
// ownership/permission cases.
package testseed

import (
	"os"
	"runtime"

	"github.com/relux-works/curator/internal/privatedir"
)

// WriteFile preserves the requested Unix mode. On Windows it explicitly sets
// current-user ownership and an owner-only DACL after fixture creation, since
// mode bits and parent directory ownership cannot establish that contract.
func WriteFile(path string, data []byte, mode os.FileMode) error {
	if err := os.WriteFile(path, data, mode); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return privatedir.ProtectFile(path)
	}
	return nil
}
