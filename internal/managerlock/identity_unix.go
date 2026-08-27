//go:build unix

package managerlock

import "path/filepath"

func canonicalWithExistingPrefix(existing string, missing []string, _ string) (string, error) {
	canonical := existing
	for index := len(missing) - 1; index >= 0; index-- {
		canonical = filepath.Join(canonical, missing[index])
	}
	return filepath.Clean(canonical), nil
}
