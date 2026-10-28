//go:build !windows

package transaction

import "github.com/relux-works/curator/internal/envunit"

func prepareExistingDirectoryProtection(dir DirectoryRecord, index int, id string) (envunit.Image, error) {
	return dir.After, nil
}
