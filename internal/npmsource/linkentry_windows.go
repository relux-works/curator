//go:build windows

package npmsource

import (
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

// linkedPackageEntry reports whether a node_modules child is the link spelling
// of an installed package. npm materializes workspace and file dependencies as
// directory junctions on Windows, which Go reports as ModeIrregular rather
// than ModeSymlink, so the junction shape is read back from the attributes:
// a reparse point that is also a directory. Any other irregular node stays
// refused.
func linkedPackageEntry(entry fs.DirEntry, path string) bool {
	if entry.Type()&fs.ModeSymlink != 0 {
		return true
	}
	return entry.Type() == fs.ModeIrregular && directoryJunction(path)
}

func directoryJunction(path string) bool {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attributes, err := windows.GetFileAttributes(path16)
	if err != nil {
		return false
	}
	return attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 &&
		attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0
}

// normalizeTreeLink resolves npm's junction spelling of a workspace link to
// its stated target before the dereferencing copy descends through it.
// filepath.EvalSymlinks does not evaluate junctions and refuses any path that
// still carries one as a component, while os.Readlink resolves the junction
// itself; the containment check downstream still decides whether the target
// is admissible.
func normalizeTreeLink(path string, entry fs.DirEntry) (string, error) {
	if entry.Type() == fs.ModeIrregular && directoryJunction(path) {
		return os.Readlink(path)
	}
	return path, nil
}
