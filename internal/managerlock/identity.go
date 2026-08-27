package managerlock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// ProjectIdentity is the canonical absolute identity used to order and name a
// project operation lock. Its UTF-8 bytes are the ordering key.
type ProjectIdentity string

// CanonicalProject resolves path to a clean absolute project identity. Existing
// symlinks are evaluated so aliases of one project contend on the same lock.
func CanonicalProject(path string) (ProjectIdentity, error) {
	canonical, err := canonicalAbsolute(path, "project path")
	if err != nil {
		return "", err
	}
	return ProjectIdentity(canonical), nil
}

// CanonicalProjects returns unique canonical identities in unsigned UTF-8 byte
// order. Duplicate physical or textual project identities are rejected.
func CanonicalProjects(paths ...string) ([]ProjectIdentity, error) {
	identities := make([]ProjectIdentity, 0, len(paths))
	for _, path := range paths {
		identity, err := CanonicalProject(path)
		if err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(i, j int) bool {
		return bytes.Compare([]byte(identities[i]), []byte(identities[j])) < 0
	})
	for index := 1; index < len(identities); index++ {
		if identities[index-1] == identities[index] {
			return nil, fmt.Errorf("duplicate canonical project identity %q", identities[index])
		}
	}
	return identities, nil
}

func canonicalAbsolute(path, label string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s is empty", label)
	}
	if !utf8.ValidString(path) || strings.IndexByte(path, 0) >= 0 {
		return "", fmt.Errorf("%s is not valid UTF-8 filesystem text", label)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	absolute = filepath.Clean(absolute)
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return canonicalWithExistingPrefix(filepath.Clean(resolved), nil, label)
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("resolve %s symlinks: %w", label, err)
	}

	// A first mutation may address a not-yet-created manager home. Resolve the
	// longest existing prefix so its identity does not change after the missing
	// suffix is created through a symlinked ancestor.
	prefix := absolute
	suffix := make([]string, 0, 4)
	for {
		parent := filepath.Dir(prefix)
		if parent == prefix {
			return "", fmt.Errorf("resolve existing prefix for %s: %w", label, err)
		}
		suffix = append(suffix, filepath.Base(prefix))
		prefix = parent

		resolved, err = filepath.EvalSymlinks(prefix)
		if err == nil {
			return canonicalWithExistingPrefix(filepath.Clean(resolved), suffix, label)
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve %s symlinks: %w", label, err)
		}
	}
}

func lockName(domain, identity string) string {
	digest := sha256.Sum256([]byte(domain + "\x00" + identity))
	return hex.EncodeToString(digest[:]) + ".lock"
}
