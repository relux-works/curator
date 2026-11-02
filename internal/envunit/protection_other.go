//go:build !windows

package envunit

import "os"

func metadataProtection(_ string, info os.FileInfo) (uint32, string, error) {
	return uint32(info.Mode().Perm()), "", nil
}

// ValidProtection admits only native metadata from this supported transport.
func ValidProtection(value string) bool { return value == "" }

// ApplyDirectoryProtection changes only the previously pinned directory mode.
// The transaction owns before/after verification and rollback ordering.
func ApplyDirectoryProtection(path string, want Image) error {
	if want.Protection != "" {
		return invalid("foreign directory protection transport")
	}
	return os.Chmod(path, os.FileMode(want.Mode))
}
