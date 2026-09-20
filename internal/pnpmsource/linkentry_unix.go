//go:build unix

package pnpmsource

import "io/fs"

// admittedLink reports whether a store-registry member or materialized
// dependency entry is pnpm's directory-link spelling. pnpm links directories
// through symlinks on Unix.
func admittedLink(info fs.FileInfo, _ string) bool {
	return info.Mode()&fs.ModeSymlink != 0
}

// normalizeTreeLink is the identity on Unix: symlink components are resolved
// by filepath.EvalSymlinks at the call site.
func normalizeTreeLink(path string, _ fs.FileInfo) (string, error) {
	return path, nil
}
