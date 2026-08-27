// Package snapshot maintains the commit-keyed immutable snapshot cache under
// the machine home: cache/<source>/<commit>/snapshot (Spec §8.2).
package snapshot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/gitops"
)

// Dir returns the conventional first-generation cache location for a source
// at a commit. Get may return a sibling immutable generation after repair.
func Dir(home, source, commit string) string {
	return filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
}

// Get returns the snapshot directory after comparing it with the immutable
// repository commit. Missing, incomplete, and tampered cache entries are
// rebuilt through atomic staging.
func Get(home, source, repo, commit string) (string, error) {
	validated, err := GetValidated(home, source, repo, commit)
	if err != nil {
		return "", err
	}
	defer validated.Close()
	return validated.Path(), nil
}

// GetValidated returns a validation token for the commit-keyed snapshot. A
// cache generation is reused only when it is structurally valid and exactly
// matches a freshly archived immutable repository tree. Repair publishes a
// sibling immutable generation and switches a regular-file pointer, avoiding
// non-portable replacement of a live non-empty directory.
func GetValidated(home, source, repo, commit string) (*buildsource.Token, error) {
	target := Dir(home, source, commit)
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(parent, ".snapshot-*.tmp")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if err := gitops.Archive(repo, commit, tmp); err != nil {
		return nil, err
	}
	authoritative, err := buildsource.Validate(tmp)
	if err != nil {
		return nil, fmt.Errorf("validate repository snapshot: %w", err)
	}
	// Equivalent only reads the frozen identity and tree state. Closing the
	// retained root before publication is required on Windows, where an open
	// directory handle can prevent a rename.
	if err := authoritative.Close(); err != nil {
		return nil, fmt.Errorf("close repository snapshot validation: %w", err)
	}

	lock, err := acquireSnapshotLock(target + ".lock")
	if err != nil {
		return nil, err
	}
	defer lock.Close()

	for _, candidate := range cacheCandidates(target) {
		cached, cacheErr := buildsource.Validate(candidate)
		if cacheErr == nil {
			if cached.Equivalent(authoritative) {
				return cached, nil
			}
			_ = cached.Close()
		}
	}

	validated, err := publish(tmp, target, authoritative)
	if err != nil {
		return nil, err
	}
	return validated, nil
}

func cacheCandidates(target string) []string {
	candidates := make([]string, 0, 2)
	if generation := selectedGeneration(target); generation != "" {
		candidates = append(candidates, generation)
	}
	return append(candidates, target)
}

func selectedGeneration(target string) string {
	pointer := target + ".current"
	pathInfo, err := os.Lstat(pointer)
	if err != nil || !pathInfo.Mode().IsRegular() || pathInfo.Size() < 2 || pathInfo.Size() > 255 {
		return ""
	}
	file, err := os.Open(pointer) // #nosec G304 -- path is manager-derived
	if err != nil {
		return ""
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() || !os.SameFile(pathInfo, openedInfo) {
		return ""
	}
	payload, err := io.ReadAll(io.LimitReader(file, 256))
	if err != nil || int64(len(payload)) != openedInfo.Size() || payload[len(payload)-1] != '\n' {
		return ""
	}
	name := strings.TrimSuffix(string(payload), "\n")
	if strings.ContainsAny(name, "/\\\r\n") || filepath.Base(name) != name ||
		!strings.HasPrefix(name, "snapshot-generation-") {
		return ""
	}
	return filepath.Join(filepath.Dir(target), name)
}

func publish(staged, target string, authoritative *buildsource.Token) (*buildsource.Token, error) {
	// Preserve the historical cache path on a cold miss. Repairs never replace
	// a present path: Windows cannot atomically replace a non-empty directory.
	if _, err := os.Lstat(target); os.IsNotExist(err) && selectedGeneration(target) == "" {
		if err := os.Rename(staged, target); err == nil {
			return validatePublished(target, authoritative)
		}
	}

	parent := filepath.Dir(target)
	oldGeneration := selectedGeneration(target)
	generation, err := os.MkdirTemp(parent, "snapshot-generation-*")
	if err != nil {
		return nil, err
	}
	if err := os.Remove(generation); err != nil {
		return nil, err
	}
	if err := os.Rename(staged, generation); err != nil {
		return nil, err
	}
	published, err := validatePublished(generation, authoritative)
	if err != nil {
		_ = os.RemoveAll(generation)
		return nil, err
	}
	if err := writeCurrent(target+".current", filepath.Base(generation)); err != nil {
		_ = published.Close()
		_ = os.RemoveAll(generation)
		return nil, err
	}
	pruneRetiredGenerations(parent, generation, oldGeneration)
	return published, nil
}

func validatePublished(path string, authoritative *buildsource.Token) (*buildsource.Token, error) {
	validated, err := buildsource.Validate(path)
	if err != nil {
		return nil, fmt.Errorf("validate published snapshot: %w", err)
	}
	if !validated.Equivalent(authoritative) {
		_ = validated.Close()
		return nil, fmt.Errorf("published snapshot does not match repository commit")
	}
	return validated, nil
}

func writeCurrent(path, generation string) error {
	parent := filepath.Dir(path)
	tmp, err := os.CreateTemp(parent, ".snapshot-current-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(generation + "\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err == nil {
		return nil
	}
	// A tampered pointer may be a directory or another non-replaceable type.
	// Snapshot API callers are serialized by the adjacent OS lock, and raw
	// generation paths remain present while this metadata is repaired.
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func pruneRetiredGenerations(parent, current, previous string) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "snapshot-generation-") {
			continue
		}
		path := filepath.Join(parent, entry.Name())
		if path == current || path == previous {
			continue
		}
		// Best effort: Windows correctly refuses while an older validation
		// token still holds the generation open. A later repair retries.
		_ = os.RemoveAll(path)
	}
}
