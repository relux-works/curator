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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/pathboundary"
)

// MarkerName is excluded from hashing by default.
const MarkerName = ".csk-install.json"

// Version identifies the framing used to compute a tree identity. A missing
// version on legacy state is interpreted as V1 by its carrier reader.
type Version uint64

const (
	// VersionV1 is the byte-frozen legacy content framing.
	VersionV1 Version = 1
	// VersionV2 is the domain-separated, length-framed content framing.
	VersionV2 Version = 2
)

// EnableV2Writers is the sole manager write cut-over for content-hash v2 and
// its versioned carrier shapes. SPEC_PIN remains rc.13, so production writes
// stay at v1 until that pin advances. TODO: enable with the rc.14 SPEC_PIN
// bump after the v2 carrier schemas are released.
//
// Tests that exercise v2 writer shapes may set this explicitly and restore it
// before returning.
var EnableV2Writers = false

// WriteVersion returns the content-hash framing selected for new manager
// state. Readers continue to accept explicitly versioned v2 state regardless
// of this write switch.
func WriteVersion() Version {
	if EnableV2Writers {
		return VersionV2
	}
	return VersionV1
}

const contentV2Domain = "curator-content-v2\x00"

// Identity keeps a digest and its framing version inseparable at comparison
// sites. Identical digest text from different versions is a different
// identity.
type Identity struct {
	HashVersion Version `json:"hash_version"`
	SHA256      string  `json:"content_sha256"`
}

// Equal reports whether both identities have the same framing version and
// normalized digest.
func (identity Identity) Equal(other Identity) bool {
	return identity.HashVersion == other.HashVersion && Normalize(identity.SHA256) == Normalize(other.SHA256)
}

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

// ContentSHA256WithVersion hashes root with the explicitly selected framing.
// ContentSHA256 remains the frozen v1 entry point for legacy readers.
func ContentSHA256WithVersion(root string, exclude map[string]bool, version Version) (string, error) {
	switch version {
	case VersionV1:
		return ContentSHA256(root, exclude)
	case VersionV2:
		return contentSHA256V2(root, exclude)
	default:
		return "", fmt.Errorf("unsupported content hash version %d", version)
	}
}

// ContentIdentity computes a versioned tree identity.
func ContentIdentity(root string, exclude map[string]bool, version Version) (Identity, error) {
	digest, err := ContentSHA256WithVersion(root, exclude, version)
	if err != nil {
		return Identity{}, err
	}
	return Identity{HashVersion: version, SHA256: digest}, nil
}

// ContentSHA256Files hashes an already materialized, path-keyed file set with
// the selected content framing. It is used for surfaces assembled in memory.
func ContentSHA256Files(files map[string][]byte, version Version) (string, error) {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	digest := sha256.New()
	switch version {
	case VersionV1:
		for index, path := range paths {
			if index > 0 {
				_, _ = digest.Write([]byte{0})
			}
			_, _ = io.WriteString(digest, path)
			_, _ = digest.Write([]byte{0})
			_, _ = digest.Write(files[path])
		}
	case VersionV2:
		_, _ = io.WriteString(digest, contentV2Domain)
		var length [8]byte
		for _, path := range paths {
			_, _ = digest.Write([]byte{'F'})
			binary.BigEndian.PutUint64(length[:], uint64(len(path)))
			_, _ = digest.Write(length[:])
			_, _ = io.WriteString(digest, path)
			binary.BigEndian.PutUint64(length[:], uint64(len(files[path])))
			_, _ = digest.Write(length[:])
			_, _ = digest.Write(files[path])
		}
	default:
		return "", fmt.Errorf("unsupported content hash version %d", version)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func contentSHA256V2(root string, exclude map[string]bool) (string, error) {
	if exclude == nil {
		exclude = map[string]bool{MarkerName: true}
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// V2 identities cover regular files only. In particular, symlinks
		// are not dereferenced into content supplied by another path.
		if !entry.Type().IsRegular() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		posix := filepath.ToSlash(rel)
		if !exclude[posix] {
			files = append(files, posix)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)

	digest := sha256.New()
	_, _ = io.WriteString(digest, contentV2Domain)
	var length [8]byte
	for _, posix := range files {
		file, err := pathboundary.OpenReadNoFollow(filepath.Join(root, filepath.FromSlash(posix)))
		if err != nil {
			return "", err
		}
		info, statErr := file.Stat()
		if statErr != nil {
			_ = file.Close()
			return "", statErr
		}
		if !info.Mode().IsRegular() {
			_ = file.Close()
			return "", fmt.Errorf("content hash v2 entry %q is not a regular file", posix)
		}
		content, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			return "", readErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		_, _ = digest.Write([]byte{'F'})
		binary.BigEndian.PutUint64(length[:], uint64(len(posix)))
		_, _ = digest.Write(length[:])
		_, _ = io.WriteString(digest, posix)
		binary.BigEndian.PutUint64(length[:], uint64(len(content)))
		_, _ = digest.Write(length[:])
		_, _ = digest.Write(content)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

// Normalize lowercases and strips the optional prefix so hashes compare
// reliably.
func Normalize(hash string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(hash), "sha256:"))
}
