//go:build !darwin && !linux

package testtoolchain

import "os"

// Windows production acquisition is a no-op. Sequential metadata controls use
// ordinary files; reject nonregular entries before opening them.
func openHostGOROOTRecord(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, os.ErrInvalid
	}
	return os.Open(path) // #nosec G304 -- diagnostic regular file, rechecked after opening.
}
