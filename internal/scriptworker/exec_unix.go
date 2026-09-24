//go:build !windows

package scriptworker

import (
	"os"
	"path/filepath"
)

// linkFarmEntry exposes one verified executable under the farm directory.
// Unix entries are symbolic links to the fixed path: the target was
// identity-verified before the link was made, the farm directory is
// operation-private, and the farm is verified entry by entry after every
// link lands.
func linkFarmEntry(farm, entry, target string) error {
	if !filepath.IsAbs(target) {
		return errFarmTargetNotAbsolute
	}
	return os.Symlink(target, filepath.Join(farm, entry))
}
