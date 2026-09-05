// Package contextstore keeps the profile store of environments §4: one
// immutable regular-file tree per closure member below the manager home,
// keyed by the member's pin — the resolved commit of a git package, the state
// hash of a path package or the synthesized local root — on the runtime-store
// pattern, so two profiles whose locks name one member at one pin share its
// entry.
//
// A git entry is extracted from the object database through gitops.Extract,
// so its bytes are a function of the commit alone (environments §1.2). A
// state entry is a copied regular-file tree keyed by its core §8 content hash.
package contextstore

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
)

// DiagSourceInvalid is the diagnostic for a state tree that violates the
// core §6.2 archive discipline (environments §1).
const DiagSourceInvalid = "profile_source_invalid"

// Root is the store root below the manager home.
func Root(home string) string { return filepath.Join(home, "contexts") }

// EntryDir is the entry path of a member at a pin key (bare commit or hash).
func EntryDir(home, kind, name, pinKey string) string {
	return filepath.Join(Root(home), kind, name, pinKey)
}

// Exists reports whether the entry is present.
func Exists(home, kind, name, pinKey string) bool {
	info, err := os.Stat(EntryDir(home, kind, name, pinKey))
	return err == nil && info.IsDir()
}

// ContentHash is the core §8 content hash of a store entry (no exclusions:
// a store entry carries no marker).
func ContentHash(dir string) (string, error) {
	return hashing.ContentSHA256(dir, map[string]bool{})
}

// EnsureGit installs the snapshot of commit from repo as the entry of
// (kind, name) and returns its path. An existing entry is reused; a new one
// is extracted beside its final path and renamed into place only after the
// extraction succeeded.
func EnsureGit(home, kind, name, repo, commit string) (string, error) {
	if !identifiers.Valid(name) {
		return "", fmt.Errorf("store entry name %q is not a portable identifier", name)
	}
	target := EntryDir(home, kind, name, commit)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return target, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp(filepath.Dir(target), "."+commit+".extract-*")
	if err != nil {
		return "", err
	}
	if err := gitops.Extract(repo, commit, staging); err != nil {
		_ = os.RemoveAll(staging)
		return "", err
	}
	return publish(staging, target)
}

// EnsureState copies the regular-file tree at source into the store as a
// state entry of name, keyed by the tree's content hash, and returns the
// entry path and the bare 64-hex state hash. A root-level .git entry is
// excluded; a link, special file, or nested .git entry is
// profile_source_invalid.
func EnsureState(home, kind, name, source string) (string, string, error) {
	if !identifiers.Valid(name) {
		return "", "", fmt.Errorf("store entry name %q is not a portable identifier", name)
	}
	parent := filepath.Join(Root(home), kind, name)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", "", err
	}
	staging, err := os.MkdirTemp(parent, ".state-*")
	if err != nil {
		return "", "", err
	}
	if err := copyStateTree(source, staging); err != nil {
		_ = os.RemoveAll(staging)
		return "", "", err
	}
	hash, err := ContentHash(staging)
	if err != nil {
		_ = os.RemoveAll(staging)
		return "", "", err
	}
	key := hashing.Normalize(hash)
	target := filepath.Join(parent, key)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		_ = os.RemoveAll(staging)
		return target, key, nil
	}
	dir, err := publish(staging, target)
	if err != nil {
		return "", "", err
	}
	return dir, key, nil
}

func publish(staging, target string) (string, error) {
	if err := os.Rename(staging, target); err != nil {
		if info, statErr := os.Stat(target); statErr == nil && info.IsDir() {
			_ = os.RemoveAll(staging)
			return target, nil
		}
		_ = os.RemoveAll(staging)
		return "", err
	}
	return target, nil
}

func copyStateTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.Name() == ".git" {
			if filepath.Dir(rel) == "." {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			return fmt.Errorf("%s: %s carries a .git entry below its root", DiagSourceInvalid, rel)
		}
		target := filepath.Join(destination, rel)
		switch {
		case entry.IsDir():
			return os.MkdirAll(target, 0o755)
		case entry.Type().IsRegular():
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Sys() != nil && hardLinked(info) {
				return fmt.Errorf("%s: %s is a hard link", DiagSourceInvalid, rel)
			}
			payload, err := os.ReadFile(path) // #nosec G304 -- walked below the source root
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			return os.WriteFile(target, payload, 0o644)
		default:
			return fmt.Errorf("%s: %s is not a directory or a regular file", DiagSourceInvalid, rel)
		}
	})
}
