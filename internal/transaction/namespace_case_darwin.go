//go:build darwin

package transaction

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const pathconfCaseSensitive = 11 // _PC_CASE_SENSITIVE from <sys/unistd.h>.

// existingNamespaceAncestor returns the closest path that exists and can be
// interrogated for filesystem behavior. A symbolic link is skipped rather than
// returned: its destination may be missing or on another filesystem, while the
// question being asked is about the directory that holds it.
//
// It lives beside its callers rather than in namespace.go because Darwin is
// the only platform that interrogates the filesystem for case and
// normalization semantics: Windows answers from the platform contract and the
// remaining unix builds answer from the POSIX one, so neither ever needs an
// ancestor to ask about.
func existingNamespaceAncestor(path string) (string, error) {
	current := filepath.Clean(path)
	for {
		if info, err := os.Lstat(current); err == nil {
			if info.Mode()&os.ModeSymlink == 0 {
				return current, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("path has no existing ancestor: %s", path)
		}
		current = parent
	}
}

func namespaceCaseInsensitive(path string) (bool, error) {
	ancestor, err := existingNamespaceAncestor(path)
	if err != nil {
		return false, err
	}
	caseSensitive, err := unix.Pathconf(ancestor, pathconfCaseSensitive)
	if err != nil {
		return false, err
	}
	return caseSensitive == 0, nil
}

func namespaceNormalizationInsensitive(path string) (bool, error) {
	ancestor, err := existingNamespaceAncestor(path)
	if err != nil {
		return false, err
	}
	var stat unix.Statfs_t
	if err := unix.Statfs(ancestor, &stat); err != nil {
		return false, err
	}
	filesystem := string(bytes.TrimRight(stat.Fstypename[:], "\x00"))
	// APFS and HFS+ compare canonically equivalent Unicode spellings as the
	// same pathname component, including on case-sensitive volumes. Other
	// Darwin-mounted filesystems retain their own component semantics.
	return filesystem == "apfs" || filesystem == "hfs", nil
}
