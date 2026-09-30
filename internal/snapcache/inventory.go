package snapcache

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/stateread"
)

// reservedNamespaces are the top-level cache directories that belong to other
// stores. They are never traversed as snapshot sources, so a skill whose
// source key collides with one of them is never pruned (retention is always
// safe).
var reservedNamespaces = map[string]bool{"build": true, "registry": true}

// prunePrefix names an entry that removal moved out of the store namespace.
// One left behind by an interrupted removal is deleted by the next real run.
const prunePrefix = ".curator-prune-"

// maxSourceDepth bounds the directory depth a source key may span below the
// cache root (host/owner/repository and a few more segments).
const maxSourceDepth = 8

// Inventory is the snapshot store's content.
type Inventory struct {
	Entries []Entry
	// Leftovers are entries an interrupted removal moved out of the store
	// namespace and did not finish deleting.
	Leftovers []string
	// Notes describe ignored store content, such as links, which retention
	// never follows or removes.
	Notes []string
}

// CacheRoot is the snapshot store root under a manager home.
func CacheRoot(home string) string { return filepath.Join(home, "cache") }

// ReadInventory lists every commit-keyed snapshot entry with its sizes and
// last-use time. It never follows a link. A cache root that exists but is
// not a plain directory fails, because its content cannot be proven.
//
// Last-use time is the modification time of the entry directory
// cache/<source>/<commit>. Every publication and every serving of the entry
// (snapshot.Get and snapshot.AuthenticateGit) creates and removes a private
// staging directory inside it, which advances that time.
func ReadInventory(home string) (Inventory, error) {
	var inventory Inventory
	root := CacheRoot(home)
	metadata, err := stateread.Lstat(root)
	if err != nil {
		return inventory, err
	}
	if metadata.Kind == stateread.KindAbsent {
		return inventory, nil
	}
	if metadata.Info.Mode()&fs.ModeSymlink != 0 || !metadata.Info.IsDir() {
		return inventory, fmt.Errorf("snapshot cache root %s is not a plain directory", root)
	}
	err = inventory.walk(root, nil)
	return inventory, err
}

func (inventory *Inventory) walk(dir string, source []string) error {
	listing, err := stateread.ReadDir(dir)
	if err != nil {
		return err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil
	}
	for _, child := range listing.Entries {
		name := child.Name()
		path := filepath.Join(dir, name)
		if len(source) == 0 && reservedNamespaces[name] {
			continue
		}
		if strings.HasPrefix(name, prunePrefix) {
			inventory.Leftovers = append(inventory.Leftovers, path)
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue // staging and other private names are never sources
		}
		info, err := os.Lstat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 || isReparsePoint(info) {
			inventory.Notes = append(inventory.Notes, fmt.Sprintf("ignored link %s in the snapshot cache", path))
			continue
		}
		if !info.IsDir() {
			continue
		}
		if len(source) > 0 && isCommit(name) {
			entry, ok, err := readEntry(path, strings.Join(source, "/"), name, info)
			if err != nil {
				return err
			}
			if ok {
				inventory.Entries = append(inventory.Entries, entry)
			}
			continue
		}
		if len(source) >= maxSourceDepth {
			continue
		}
		if err := inventory.walk(path, append(append([]string(nil), source...), name)); err != nil {
			return err
		}
	}
	return nil
}

func readEntry(path, source, commit string, info fs.FileInfo) (Entry, bool, error) {
	snapshotInfo, err := os.Lstat(filepath.Join(path, "snapshot"))
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, false, nil // no published snapshot, only staging residue
		}
		return Entry{}, false, err
	}
	if !snapshotInfo.IsDir() || snapshotInfo.Mode()&fs.ModeSymlink != 0 {
		return Entry{}, false, nil
	}
	allocated, logical, err := measure(path)
	if err != nil {
		return Entry{}, false, fmt.Errorf("measure snapshot %s: %w", path, err)
	}
	return Entry{
		Source:         source,
		Commit:         commit,
		LastUsed:       info.ModTime().UTC(),
		AllocatedBytes: allocated,
		LogicalBytes:   logical,
		Path:           path,
	}, true, nil
}

// measure sums the regular files below path without following links. Each
// file identity counts once, so hard links inside one entry are not counted
// twice.
func measure(path string) (allocated, logical int64, err error) {
	seen := map[string]bool{}
	err = filepath.WalkDir(path, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if id, ok := fileIdentity(info); ok {
			if seen[id] {
				return nil
			}
			seen[id] = true
		}
		logical += info.Size()
		allocated += allocatedSize(info)
		return nil
	})
	return allocated, logical, err
}

// isCommit reports whether name is a full lowercase SHA-1 or SHA-256 commit.
func isCommit(name string) bool {
	if len(name) != 40 && len(name) != 64 {
		return false
	}
	for _, r := range name {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
