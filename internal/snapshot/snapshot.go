// Package snapshot maintains the commit-keyed immutable snapshot cache under
// the machine home: cache/<source>/<commit>/snapshot (Spec §8.2).
package snapshot

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/transaction"
)

// ErrDestinationConflict reports that an existing commit-keyed cache entry
// could not be authenticated as the exact snapshot being published.
var ErrDestinationConflict = errors.New("snapshot destination conflicts with immutable commit")

// digestSnapshotDestination is a narrow seam for exercising transient
// destination-open failures. Production uses transaction.DigestPath.
var digestSnapshotDestination = transaction.DigestPath

const destinationSharingViolationRetries = 3
const destinationSharingViolationRetryDelay = 10 * time.Millisecond

// Dir returns the cache location for a source at a commit.
func Dir(home, source, commit string) string {
	return filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
}

// AuthenticateGit proves the cached snapshot for source at commit is exactly
// the tree of commit in repo and returns its directory. It is the read-only
// frozen-consumption counterpart of Get: a missing cache entry fails with
// source_snapshot_unavailable and never triggers extraction-into-place, while
// a cache tree whose complete byte inventory (every regular file including
// runtime and build roots, plus permission bits) differs from the locked
// commit fails with source_snapshot_changed. Only local repository reads back
// the pinned commit; no fetch, no ref resolution, and no live-source adoption
// occurs. Callers authenticate the shared commit tree once, then serve locked
// member subtrees from it.
func AuthenticateGit(home, source, repo, commit string) (string, error) {
	target := Dir(home, source, commit)
	info, err := os.Lstat(target)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("source_snapshot_unavailable: snapshot for %s at %s is not in the store; capture it with an explicit attempt", source, commit)
		}
		return "", err
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s at %s is not a directory", source, commit)
	}
	if err := gitops.EnsureRepo(repo); err != nil {
		return "", fmt.Errorf("source_snapshot_unavailable: cannot authenticate snapshot for %s at %s without its locked repository; capture it with an explicit attempt", source, commit)
	}
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(parent, ".snapshot-auth-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	// MkdirTemp deliberately creates a private directory. Restore the cache
	// root mode before comparing, exactly as publication does, so the root
	// mode itself never reports a false mismatch.
	if err := os.Chmod(tmp, 0o755); err != nil { // #nosec G302 -- authentication compares against published immutable snapshots that preserve the historical world-readable cache-root mode.
		return "", err
	}
	if err := gitops.Extract(repo, commit, tmp); err != nil {
		return "", fmt.Errorf("source_snapshot_unavailable: cannot authenticate snapshot for %s at %s without its locked commit; capture it with an explicit attempt", source, commit)
	}
	expectedDigest, err := transaction.DigestPath(tmp)
	if err != nil {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s at %s failed authentication: %v", source, commit, err)
	}
	actualDigest, err := transaction.DigestPath(target)
	if err != nil {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s at %s failed authentication: %v", source, commit, err)
	}
	if actualDigest != expectedDigest {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s at %s does not match its locked commit", source, commit)
	}
	return target, nil
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
	if err := gitops.Extract(repo, commit, tmp); err != nil {
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
	destinationDigest, err := digestDestinationWithSharingRetry(destination)
	if err != nil {
		return fmt.Errorf("%w: authenticate destination: %v", ErrDestinationConflict, err)
	}
	if destinationDigest != expectedDigest {
		return fmt.Errorf("%w: destination tree does not match expected snapshot", ErrDestinationConflict)
	}
	return nil
}

// digestDestinationWithSharingRetry handles a short Windows sharing violation
// while another Get is publishing the immutable destination. All other read
// errors, and a sharing violation that outlasts this bound, fail closed.
func digestDestinationWithSharingRetry(destination string) (string, error) {
	for retry := 0; ; retry++ {
		digest, err := digestSnapshotDestination(destination)
		if err == nil || !isDestinationSharingViolation(err) || retry == destinationSharingViolationRetries {
			return digest, err
		}
		time.Sleep(destinationSharingViolationRetryDelay)
	}
}
