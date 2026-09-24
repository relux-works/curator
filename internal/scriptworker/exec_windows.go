//go:build windows

package scriptworker

import (
	"io"
	"os"
	"path/filepath"
)

// linkFarmEntry exposes one verified executable under the farm directory.
//
// Windows entries are byte copies of the fixed path, never hard links: the
// launch-boundary identity check refuses any executable with more than one
// filesystem link (a manager-made hard link is indistinguishable from a
// substitution), so linking the verified file would make the session refuse
// itself. Symlinks are not used because creating them needs a privilege a
// copy does not. The farm is verified entry by entry after every copy
// lands, by hash for copies.
func linkFarmEntry(farm, entry, target string) error {
	if !filepath.IsAbs(target) {
		return errFarmTargetNotAbsolute
	}
	return copyFarmEntry(target, filepath.Join(farm, entry))
}

func copyFarmEntry(source, destination string) error {
	input, err := os.Open(source) // #nosec G304 -- canonical manager-resolved exec path
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	info, err := input.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxInterpreterBytes {
		return errFarmTargetUnusable
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700) // #nosec G304 -- manager-owned farm entry
	if err != nil {
		return err
	}
	defer func() { _ = output.Close() }()
	if _, err := io.CopyN(output, input, info.Size()); err != nil {
		_ = os.Remove(destination)
		return err
	}
	return nil
}
