//go:build unix

package pathboundary

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func protectTree(string) error { return nil }

func isLink(_ string, info os.FileInfo) (bool, error) {
	return info.Mode()&fs.ModeSymlink != 0, nil
}

func effectiveOwner() (OwnerIdentity, error) {
	return OwnerIdentity(fmt.Sprintf("uid:%d", os.Geteuid())), nil
}

func lookupOwner(path string, info os.FileInfo) (OwnerIdentity, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("cannot prove owner of %s", path)
	}
	return OwnerIdentity(fmt.Sprintf("uid:%d", stat.Uid)), nil
}

func checkMutationPermissions(path string, info os.FileInfo) error {
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("%s grants group or other identities mutation permission", path)
	}
	return nil
}
