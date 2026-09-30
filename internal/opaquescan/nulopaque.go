// Package opaquescan finds snapshot files whose NUL bytes make the v1 tree
// framing ambiguous. It walks directories without following symlinks and
// reads manager-owned file bytes through stateread.
package opaquescan

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/stateread"
)

// NULPaths returns the POSIX-style paths of regular files below root that
// contain at least one NUL byte. Symlinks and other non-regular entries are
// ignored. A missing root is an empty snapshot; failed or partial reads are
// errors, never absence.
func NULPaths(root string) ([]string, error) {
	metadata, err := stateread.Lstat(root)
	if err != nil {
		return nil, err
	}
	if metadata.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return nil, stateread.UnusableError(root, fmt.Errorf("unknown snapshot metadata state %q", metadata.Kind))
	}
	if metadata.Info.Mode()&fs.ModeSymlink != 0 {
		return nil, nil
	}
	if !metadata.Info.IsDir() {
		return nil, stateread.UnusableError(root, fmt.Errorf("snapshot root is not a directory"))
	}

	var paths []string
	if err := walk(root, root, &paths); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func walk(root, dir string, paths *[]string) error {
	listing, err := stateread.ReadDir(dir)
	if err != nil {
		return err
	}
	if listing.Kind == stateread.KindAbsent {
		return stateread.UnusableError(dir, fmt.Errorf("snapshot directory disappeared during scan"))
	}
	if listing.Kind != stateread.KindPresent {
		return stateread.UnusableError(dir, fmt.Errorf("unknown snapshot directory state %q", listing.Kind))
	}

	for _, entry := range listing.Entries {
		path := filepath.Join(dir, entry.Name())
		metadata, err := stateread.Lstat(path)
		if err != nil {
			return err
		}
		if metadata.Kind == stateread.KindAbsent {
			return stateread.UnusableError(path, fmt.Errorf("snapshot entry disappeared during scan"))
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return stateread.UnusableError(path, fmt.Errorf("unknown snapshot entry state %q", metadata.Kind))
		}
		if metadata.Info.IsDir() {
			if err := walk(root, path, paths); err != nil {
				return err
			}
			continue
		}
		if !metadata.Info.Mode().IsRegular() {
			continue
		}

		file, err := stateread.ReadRegularFile(path)
		if err != nil {
			return err
		}
		if file.Kind != stateread.KindPresent {
			return stateread.UnusableError(path, fmt.Errorf("snapshot file is no longer present"))
		}
		if bytes.IndexByte(file.Bytes, 0) < 0 {
			continue
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return stateread.UnusableError(path, err)
		}
		*paths = append(*paths, filepath.ToSlash(rel))
	}
	return nil
}
