// Package hashing computes the deterministic content hash of an installed
// tree (Spec §8.5).
//
// The hash is SHA-256 over a byte stream assembled from the tree's files,
// sorted by POSIX-style relative path. Each file contributes
// "relpath NUL content"; records are joined with NUL. The result is prefixed
// "sha256:". The install marker itself is excluded so the marker can carry
// the hash of everything else.
package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MarkerName is excluded from hashing by default.
const MarkerName = ".csk-install.json"

// ContentSHA256 hashes the tree rooted at root, excluding the given
// root-relative POSIX paths (defaults to the install marker when nil).
func ContentSHA256(root string, exclude map[string]bool) (string, error) {
	if exclude == nil {
		exclude = map[string]bool{MarkerName: true}
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() && entry.Type()&fs.ModeSymlink == 0 {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		posix := filepath.ToSlash(rel)
		if exclude[posix] {
			return nil
		}
		files = append(files, posix)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)

	digest := sha256.New()
	for index, posix := range files {
		if index > 0 {
			digest.Write([]byte{0})
		}
		digest.Write([]byte(posix))
		digest.Write([]byte{0})
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(posix))) // #nosec G304 -- paths come from the walked tree
		if err != nil {
			return "", err
		}
		digest.Write(content)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

// Normalize lowercases and strips the optional prefix so hashes compare
// reliably.
func Normalize(hash string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(hash), "sha256:"))
}
