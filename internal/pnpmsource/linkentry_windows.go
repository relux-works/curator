//go:build windows

package pnpmsource

import (
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

// admittedLink reports whether a store-registry member or materialized
// dependency entry is pnpm's directory-link spelling. pnpm links directories
// through symlink-dir, which creates junctions on Windows, and Go reports a
// junction as ModeIrregular rather than ModeSymlink, so the junction shape is
// read back from the attributes: a reparse point that is also a directory.
// Any other irregular node stays refused.
func admittedLink(info fs.FileInfo, path string) bool {
	if info.Mode()&fs.ModeSymlink != 0 {
		return true
	}
	return info.Mode().Type() == fs.ModeIrregular && directoryJunction(path)
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

// normalizeTreeLink resolves pnpm's junction spelling of a directory link to
// its stated target before comparison. filepath.EvalSymlinks does not
// evaluate junctions, while os.Readlink resolves the junction itself; the
// target comparison downstream still decides whether the link is admissible.
func normalizeTreeLink(path string, info fs.FileInfo) (string, error) {
	if info.Mode().Type() == fs.ModeIrregular && directoryJunction(path) {
		return os.Readlink(path)
	}
	return path, nil
}
