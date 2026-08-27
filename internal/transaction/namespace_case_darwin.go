//go:build darwin

package transaction

import (
	"bytes"

	"golang.org/x/sys/unix"
)

const pathconfCaseSensitive = 11 // _PC_CASE_SENSITIVE from <sys/unistd.h>.

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
