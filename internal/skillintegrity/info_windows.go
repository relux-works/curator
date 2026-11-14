//go:build windows

package skillintegrity

import (
	"io/fs"
	"os"
)

// File.Stat eagerly retains the volume/index from a no-follow handle. WalkDir
// path-backed FileInfo defers that lookup until SameFile, after a possible swap.
func captureEntryInfo(path string, entry fs.DirEntry) (os.FileInfo, error) {
	info, err := entry.Info()
	if err != nil || !info.Mode().IsRegular() {
		return info, err
	}
	f, err := openCaptureFile(path)
	if err != nil {
		return nil, failure("skill_integrity_boundary_invalid", "file identity could not be retained")
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}
