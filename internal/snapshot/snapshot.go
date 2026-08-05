// Package snapshot maintains the commit-keyed immutable snapshot cache under
// the machine home: cache/<source>/<commit>/snapshot (Spec §8.2).
package snapshot

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/transaction"
)

// ErrDestinationConflict reports that an existing commit-keyed cache entry
// could not be authenticated as the exact snapshot being published.
var ErrDestinationConflict = errors.New("snapshot destination conflicts with immutable commit")

// Dir returns the cache location for a source at a commit.
func Dir(home, source, commit string) string {
	return filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
}

// Get returns the snapshot directory, producing it from the repository on a
// cache miss. Staging is atomic: a concurrent producer of the same snapshot
// wins harmlessly.
func Get(home, source, repo, commit string) (string, error) {
	target := Dir(home, source, commit)
	if _, err := os.Lstat(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(parent, ".snapshot-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	// MkdirTemp deliberately creates a private directory. Restore the cache
	// root mode used by the original archive publication before comparing or
	// publishing it.
	if err := os.Chmod(tmp, 0o755); err != nil { // #nosec G302 -- published immutable snapshots intentionally preserve the historical world-readable cache-root mode.
		return "", err
	}
	if err := gitops.Archive(repo, commit, tmp); err != nil {
		return "", err
	}
	// A commit-shaped directory name is not evidence that its contents came
	// from that commit. Authenticate hits against a freshly prepared archive
	// just like concurrent publication winners; never pass tampered or partial
	// cache bytes to installers.
	if _, err := os.Lstat(target); err == nil {
		if err := authenticateDestination(tmp, target); err != nil {
			return "", err
		}
		return target, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err := publishPreparedSnapshot(tmp, target); err != nil {
		return "", err
	}
	return target, nil
}

// publishPreparedSnapshot atomically installs expected without replacing any
// destination that appeared while the caller prepared it. A concurrent winner
// is reusable only when its complete tree authenticates as expected.
func publishPreparedSnapshot(expected, target string) error {
	if _, err := os.Lstat(target); err == nil {
		return authenticateDestination(expected, target)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := renameNoReplace(expected, target); err != nil {
		// Another process may have published the same immutable commit after
		// our final pre-rename check. Reuse its destination only after proving
		// the complete link-free tree, file bytes, and permission bits equal
		// the archive we prepared for this commit.
		authErr := authenticateDestination(expected, target)
		if authErr == nil {
			return nil
		}
		return errors.Join(fmt.Errorf("publish snapshot: %w", err), authErr)
	}
	return nil
}

func authenticateDestination(expected, destination string) error {
	info, err := os.Lstat(destination)
	if err != nil {
		return fmt.Errorf("%w: inspect destination: %v", ErrDestinationConflict, err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("%w: destination is not a directory", ErrDestinationConflict)
	}
	expectedDigest, err := transaction.DigestPath(expected)
	if err != nil {
		return fmt.Errorf("%w: authenticate expected snapshot: %v", ErrDestinationConflict, err)
	}
	destinationDigest, err := transaction.DigestPath(destination)
	if err != nil {
		return fmt.Errorf("%w: authenticate destination: %v", ErrDestinationConflict, err)
	}
	if destinationDigest != expectedDigest {
		return fmt.Errorf("%w: destination tree does not match expected snapshot", ErrDestinationConflict)
	}
	return nil
}
