//go:build !unix && !windows

package pathboundary

import (
	"fmt"
	"io/fs"
	"os"
)

func protectTree(string) error {
	return fmt.Errorf("cannot establish private mutation permissions on this platform")
}

func isLink(_ string, info os.FileInfo) (bool, error) {
	return info.Mode()&fs.ModeSymlink != 0, nil
}

func effectiveOwner() (OwnerIdentity, error) {
	return "", fmt.Errorf("cannot prove operator ownership on this platform")
}

func lookupOwner(path string, _ os.FileInfo) (OwnerIdentity, error) {
	return "", fmt.Errorf("cannot prove operator ownership for %s on this platform", path)
}

func checkMutationPermissions(path string, _ os.FileInfo) error {
	return fmt.Errorf("cannot prove private mutation permissions for %s on this platform", path)
}
