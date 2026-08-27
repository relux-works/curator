package buildcache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
)

// HomeLock is a caller-owned witness for the exclusive manager-home lock.
// Implementations must fail AssertHeld after release. Buildcache never creates
// a lock, which keeps dry-run inspection strictly read-only.
type HomeLock interface {
	AssertHeld() error
}

// Publication contains one fully built private artifact and its exact receipt.
// ArtifactSource is copied from its open file handle and is never executed.
type Publication struct {
	Input          buildmeta.Input
	ReceiptBytes   []byte
	ArtifactSource string
}

// PublicationStatus distinguishes a newly selected directory winner from an
// identical winner selected by another publisher.
type PublicationStatus string

// Publication outcomes: a newly protected winner, or an identical existing one.
const (
	Published    PublicationStatus = "published"
	ReusedWinner PublicationStatus = "reused-winner"
)

// PublicationResult identifies the protected immutable winner.
type PublicationResult struct {
	Status       PublicationStatus
	ArtifactPath string
	ReceiptHash  buildmeta.ReceiptHash
}

// ConflictError reports different bytes for the same logical cache key.
// The existing protected winner is left unchanged.
type ConflictError struct {
	Key buildmeta.CacheKey
}

func (err *ConflictError) Error() string {
	return fmt.Sprintf("cache publication conflict for %s", err.Key)
}

// Publish verifies a private build, creates a protected staging directory, and
// selects one complete directory winner atomically. Existing corrupt or
// untrusted entries are quarantined rather than adopted or permission-repaired.
func (store *Store) Publish(publication Publication, lock HomeLock) (PublicationResult, error) {
	if err := requireHomeLock(lock); err != nil {
		return PublicationResult{}, err
	}
	if store == nil || store.supported == nil || !store.supported() {
		return PublicationResult{}, fmt.Errorf("persistent build cache protection is unsupported")
	}
	receipt, err := buildmeta.DecodeExpectedReceipt(publication.ReceiptBytes, publication.Input)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("validate publication receipt: %w", err)
	}
	key, err := publication.Input.CacheKey()
	if err != nil {
		return PublicationResult{}, fmt.Errorf("derive publication key: %w", err)
	}
	if receipt.CacheKey != key {
		return PublicationResult{}, fmt.Errorf("publication receipt key mismatch")
	}
	receiptHash, err := buildmeta.HashReceiptBytes(publication.ReceiptBytes)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("hash publication receipt: %w", err)
	}

	source, err := openRegularSource(publication.ArtifactSource)
	if err != nil {
		return PublicationResult{}, err
	}
	defer func() { _ = source.Close() }()
	artifactHash, artifactSize, err := hashOpenFile(source)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("hash publication artifact: %w", err)
	}
	if artifactHash != receipt.Artifact.SHA256 || artifactSize != receipt.Artifact.Size {
		return PublicationResult{}, fmt.Errorf("publication artifact does not match receipt")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return PublicationResult{}, fmt.Errorf("rewind publication artifact: %w", err)
	}

	entryPath, base, err := store.paths(key)
	if err != nil {
		return PublicationResult{}, err
	}
	if err := ensureProtectedBase(store.home, base); err != nil {
		return PublicationResult{}, fmt.Errorf("prepare protected cache root: %w", err)
	}
	stage, err := makeProtectedTempDir(base, ".stage-"+strings.TrimPrefix(string(key), "sha256:")+"-")
	if err != nil {
		return PublicationResult{}, fmt.Errorf("create publication staging: %w", err)
	}
	keepStage := false
	defer func() {
		if !keepStage {
			_ = os.RemoveAll(stage)
		}
	}()

	artifactRel := filepath.FromSlash(receipt.Artifact.Path)
	binDir := filepath.Join(stage, "bin")
	if err := createProtectedDir(binDir); err != nil {
		return PublicationResult{}, fmt.Errorf("create staged artifact directory: %w", err)
	}
	stagedArtifact := filepath.Join(stage, artifactRel)
	if err := writeProtectedFile(stagedArtifact, 0o700, source); err != nil {
		return PublicationResult{}, fmt.Errorf("write staged artifact: %w", err)
	}
	if err := writeProtectedFile(filepath.Join(stage, ReceiptFilename), 0o600, bytes.NewReader(publication.ReceiptBytes)); err != nil {
		return PublicationResult{}, fmt.Errorf("write staged receipt: %w", err)
	}
	if err := syncDirectory(binDir); err != nil {
		return PublicationResult{}, fmt.Errorf("sync staged artifact directory: %w", err)
	}
	if err := syncDirectory(stage); err != nil {
		return PublicationResult{}, fmt.Errorf("sync staged cache entry: %w", err)
	}

	expect := Expectation{Input: publication.Input, ReceiptHash: receiptHash}
	winnerExpect := Expectation{Input: publication.Input}
	staged := store.inspectEntry(stage, expect, key)
	if staged.Status != Hit {
		return PublicationResult{}, fmt.Errorf("staged cache entry failed protected validation: %s", staged.Reason)
	}

	for attempts := 0; attempts < 3; attempts++ {
		if err := requireHomeLock(lock); err != nil {
			return PublicationResult{}, err
		}
		winner := store.Inspect(winnerExpect)
		switch winner.Status {
		case Hit:
			if bytes.Equal(winner.ReceiptBytes, publication.ReceiptBytes) &&
				winner.Receipt.Artifact.SHA256 == artifactHash && winner.Receipt.Artifact.Size == artifactSize {
				return PublicationResult{Status: ReusedWinner, ArtifactPath: winner.ArtifactPath, ReceiptHash: winner.ReceiptHash}, nil
			}
			return PublicationResult{}, &ConflictError{Key: key}
		case Unsupported:
			return PublicationResult{}, fmt.Errorf("persistent build cache protection is unsupported")
		case Corrupt, UntrustedProvenance:
			if _, err := store.quarantinePath(entryPath, lock); err != nil {
				return PublicationResult{}, err
			}
		case Miss:
		}

		if err := renameDirectoryNoReplace(stage, entryPath); err == nil {
			keepStage = true
			winner = store.Inspect(expect)
			if winner.Status != Hit {
				return PublicationResult{}, fmt.Errorf("published cache winner failed validation: %s", winner.Reason)
			}
			if err := syncDirectory(base); err != nil {
				return PublicationResult{}, fmt.Errorf("sync cache root: %w", err)
			}
			return PublicationResult{Status: Published, ArtifactPath: winner.ArtifactPath, ReceiptHash: winner.ReceiptHash}, nil
		}
		// A racing publisher may have selected a winner. Loop to validate it;
		// no directory is merged and this staging directory remains private.
	}
	return PublicationResult{}, fmt.Errorf("cache publication could not select or validate a winner")
}

// Quarantine atomically moves a present entry aside under the caller-held home
// lock. It never traverses or repairs the entry and returns an empty path for a
// missing key.
func (store *Store) Quarantine(key buildmeta.CacheKey, lock HomeLock) (string, error) {
	if err := requireHomeLock(lock); err != nil {
		return "", err
	}
	if store == nil || store.supported == nil || !store.supported() {
		return "", fmt.Errorf("persistent build cache protection is unsupported")
	}
	entryPath, _, err := store.paths(key)
	if err != nil {
		return "", err
	}
	return store.quarantinePath(entryPath, lock)
}

func (store *Store) quarantinePath(entryPath string, lock HomeLock) (string, error) {
	if err := requireHomeLock(lock); err != nil {
		return "", err
	}
	if _, err := os.Lstat(entryPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("inspect cache entry for quarantine: %w", err)
	}
	parent := filepath.Dir(entryPath)
	name := ".quarantine-" + filepath.Base(entryPath) + "-"
	placeholder, err := os.CreateTemp(parent, name)
	if err != nil {
		return "", fmt.Errorf("reserve quarantine name: %w", err)
	}
	quarantinePath := placeholder.Name()
	if err := placeholder.Close(); err != nil {
		_ = os.Remove(quarantinePath)
		return "", err
	}
	if err := os.Remove(quarantinePath); err != nil {
		return "", err
	}
	if err := os.Rename(entryPath, quarantinePath); err != nil {
		return "", fmt.Errorf("quarantine cache entry: %w", err)
	}
	if err := syncDirectory(parent); err != nil {
		return "", fmt.Errorf("sync quarantined cache root: %w", err)
	}
	return quarantinePath, nil
}

func requireHomeLock(lock HomeLock) error {
	if lock == nil || (reflect.ValueOf(lock).Kind() == reflect.Pointer && reflect.ValueOf(lock).IsNil()) {
		return fmt.Errorf("caller-held manager-home lock is required")
	}
	if err := lock.AssertHeld(); err != nil {
		return fmt.Errorf("manager-home lock is not held: %w", err)
	}
	return nil
}

func openRegularSource(path string) (*os.File, error) {
	if path == "" {
		return nil, fmt.Errorf("publication artifact source is empty")
	}
	lstat, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect publication artifact: %w", err)
	}
	if !lstat.Mode().IsRegular() {
		return nil, fmt.Errorf("publication artifact is not a regular file")
	}
	file, err := os.Open(path) // #nosec G304 -- caller-provided private staging input is verified below
	if err != nil {
		return nil, fmt.Errorf("open publication artifact: %w", err)
	}
	opened, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(lstat, opened) {
		_ = file.Close()
		return nil, fmt.Errorf("publication artifact changed while opening")
	}
	return file, nil
}
