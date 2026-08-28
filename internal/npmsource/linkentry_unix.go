//go:build unix

package npmsource

import "io/fs"

// linkedPackageEntry reports whether a node_modules child is the link spelling
// of an installed package. npm materializes workspace and file dependencies as
// symlinks on Unix.
func linkedPackageEntry(entry fs.DirEntry, _ string) bool {
	return entry.Type()&fs.ModeSymlink != 0
}

// normalizeTreeLink is the identity on Unix: symlink components are resolved
// by filepath.EvalSymlinks inside the dereferencing copy itself.
func normalizeTreeLink(path string, _ fs.DirEntry) (string, error) {
	return path, nil
}
