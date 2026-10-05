package envprofile

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

const diagPinHash = "pin_hash"

type protectedRoots struct {
	environments bool
	store        bool
}

type storeEntryFailure struct {
	Member contextlock.Member
	Path   string
	Check  string
	Err    error
}

func (f *storeEntryFailure) Error() string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%s: store entry %s/%s failed %s check: %v", envregistry.DiagStoreUntrusted, f.Member.Kind, f.Member.Name, f.Check, f.Err)
}

func boundaryOwner(lookup pathboundary.OwnerLookup) pathboundary.OwnerLookup {
	if lookup != nil {
		return lookup
	}
	return pathboundary.DefaultOwnerLookup()
}

func validateResolveRoots(home string, owner pathboundary.OwnerLookup) (protectedRoots, error) {
	var roots protectedRoots
	var err error
	roots.environments, err = validateOptionalProtectedRoot(home, EnvRoot(home), "environments root", owner)
	if err != nil {
		return roots, err
	}
	roots.store, err = validateOptionalProtectedRoot(home, contextstore.Root(home), "profile store root", owner)
	if err != nil {
		return roots, err
	}
	return roots, nil
}

func validateOptionalProtectedRoot(home, root, label string, owner pathboundary.OwnerLookup) (bool, error) {
	state, err := stateread.Lstat(root)
	if err != nil {
		return false, boundaryDiagnostic(label, "boundary", err)
	}
	if state.Kind == stateread.KindAbsent {
		return false, nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil {
		return false, boundaryDiagnostic(label, "boundary", fmt.Errorf("root metadata is unavailable"))
	}
	if err := pathboundary.ValidateLeafWithOwner(home, root, boundaryOwner(owner)); err != nil {
		return false, boundaryDiagnostic(label, boundaryCheck(err), err)
	}
	return true, nil
}

func validateLockBoundary(home, profile string, owner pathboundary.OwnerLookup) error {
	path := lockPath(home, profile)
	state, err := stateread.Lstat(path)
	if err != nil {
		return boundaryDiagnostic("profile lock", "boundary", err)
	}
	if state.Kind == stateread.KindAbsent {
		return nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil {
		return boundaryDiagnostic("profile lock", "boundary", fmt.Errorf("lock metadata is unavailable"))
	}
	if err := pathboundary.ValidateLeafWithOwner(home, path, boundaryOwner(owner)); err != nil {
		return boundaryDiagnostic("profile lock", boundaryCheck(err), err)
	}
	return nil
}

func validateMarkerBoundary(home, profile, envID string, owner pathboundary.OwnerLookup) error {
	root := EnvRoot(home)
	path := filepath.Join(ManagedHomeDir(home, profile, envID), envmarker.Name)
	state, err := stateread.Lstat(path)
	if err != nil {
		return boundaryDiagnostic("environment marker", "boundary", err)
	}
	if state.Kind == stateread.KindAbsent {
		return nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil {
		return boundaryDiagnostic("environment marker", "boundary", fmt.Errorf("marker metadata is unavailable"))
	}
	if err := pathboundary.ValidateWithinWithOwner(root, path, boundaryOwner(owner)); err != nil {
		return boundaryDiagnostic("environment marker", boundaryCheck(err), err)
	}
	return nil
}

func boundaryCheck(err error) string {
	var failure *pathboundary.Failure
	if errors.As(err, &failure) && failure.Check != "" {
		return failure.Check
	}
	return "boundary"
}

func boundaryDiagnostic(label, check string, err error) error {
	return fmt.Errorf("%s: %s failed %s check: %v", envregistry.DiagStoreUntrusted, label, check, err)
}

func wouldRebuildUntrustedStore(err error) error {
	return fmt.Errorf("%s: %v", envregistry.DiagWouldRebuildUntrustedStore, err)
}

func dryRunStoreEntryFailure(failure *storeEntryFailure) error {
	if failure == nil {
		return fmt.Errorf("%s: store entry failure is unavailable", DiagRepairFailed)
	}
	if pinnedCommitObjectUnavailable(failure) {
		return failure
	}
	if failure.Member.Commit == "" {
		return fmt.Errorf("%s: %w; this state pin has no revalidatable Git snapshot", DiagRepairFailed, failure)
	}
	return wouldRebuildUntrustedStore(failure)
}

func validateNamedStoreBoundaries(home string, lock *contextlock.Lock, owner pathboundary.OwnerLookup) *storeEntryFailure {
	root := contextstore.Root(home)
	for _, member := range lock.Members {
		entry := contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey())
		if err := pathboundary.ValidateWithinWithOwner(root, entry, boundaryOwner(owner)); err != nil {
			return &storeEntryFailure{Member: member, Path: entry, Check: boundaryCheck(err), Err: err}
		}
	}
	return nil
}

func validateNamedStorePins(home string, lock *contextlock.Lock) *storeEntryFailure {
	for _, member := range lock.Members {
		entry := contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey())
		actual, expected, err := storeEntryPinHashes(home, member, entry, lock.ContentHashVersion())
		if err != nil {
			return &storeEntryFailure{Member: member, Path: entry, Check: diagPinHash, Err: err}
		}
		if actual != expected {
			return &storeEntryFailure{Member: member, Path: entry, Check: diagPinHash, Err: fmt.Errorf("recorded pin %q does not match store hash %q", member.Pin(), actual)}
		}
	}
	return nil
}

func validateProfileStoreState(home, profile string, readRegularFile func(string) (stateread.File, error), owner pathboundary.OwnerLookup) error {
	roots, err := validateResolveRoots(home, owner)
	if err != nil {
		return err
	}
	_, lock, _, err := loadResolveInputsWithOwner(home, profile, readRegularFile, owner)
	if err != nil {
		return err
	}
	if !roots.store {
		return boundaryDiagnostic("profile store root", "boundary", fmt.Errorf("root %s is absent", contextstore.Root(home)))
	}
	if failure := validateNamedStoreBoundaries(home, lock, owner); failure != nil {
		return failure
	}
	if failure := validateNamedStorePins(home, lock); failure != nil {
		return failure
	}
	return nil
}

func expectedStoreHash(home string, member contextlock.Member) (string, string, error) {
	if member.Commit != "" {
		return pinnedCommitTree(home, member)
	}
	if member.StateHash != "" {
		return "sha256:" + member.StateHash, "", nil
	}
	return "", "", fmt.Errorf("member %s/%s has no expected state or commit pin", member.Kind, member.Name)
}

func storeEntryPinHashes(home string, member contextlock.Member, entry string, version hashing.Version) (actual, expected string, err error) {
	expected, format, err := expectedStoreHash(home, member)
	if err != nil {
		return "", "", err
	}
	if member.Commit != "" {
		actual, err = gitops.TreeObjectIDFromDir(entry, format)
		if err != nil {
			return "", "", err
		}
		return actual, expected, nil
	}
	actual, err = hashing.ContentSHA256WithVersion(entry, map[string]bool{}, version)
	if err != nil {
		return "", "", err
	}
	return actual, expected, nil
}

func pinnedCommitObjectUnavailable(failure *storeEntryFailure) bool {
	return failure != nil && failure.Check == diagPinHash &&
		strings.Contains(failure.Err.Error(), "pinned commit object unavailable")
}

func pinnedCommitTree(home string, member contextlock.Member) (string, string, error) {
	if member.Commit == "" || member.Source == "" {
		return "", "", fmt.Errorf("member %s/%s has no recomputable commit pin", member.Kind, member.Name)
	}
	repo := newGitManager(home).repoDir(member.Source)
	treeID, format, err := gitops.PinnedCommitTree(repo, member.Commit)
	if err != nil {
		return "", "", fmt.Errorf("pinned commit object unavailable for member %s/%s: %w", member.Kind, member.Name, err)
	}
	return treeID, format, nil
}

func extractPinnedSnapshot(home string, member contextlock.Member) (string, func(), string, error) {
	if member.Commit == "" || member.Source == "" {
		return "", nil, "", fmt.Errorf("member %s/%s has no recomputable pin", member.Kind, member.Name)
	}
	snapshot, err := os.MkdirTemp("", "curator-store-pin-*")
	if err != nil {
		return "", nil, "", err
	}
	cleanup := func() { _ = os.RemoveAll(snapshot) }
	repo := newGitManager(home).repoDir(member.Source)
	if err := gitops.Extract(repo, member.Commit, snapshot); err != nil {
		cleanup()
		return "", nil, "", fmt.Errorf("extract pinned snapshot %s: %w", member.Commit, err)
	}
	hash, err := contextstore.ContentHash(snapshot)
	if err != nil {
		cleanup()
		return "", nil, "", fmt.Errorf("hash pinned snapshot %s: %w", member.Commit, err)
	}
	return snapshot, cleanup, hash, nil
}

func rebuildUntrustedStoreEntry(req *ResolveRequest, expectedLockHash string, failure *storeEntryFailure) error {
	if failure == nil {
		return fmt.Errorf("%s: store entry failure is unavailable", DiagRepairFailed)
	}
	member := failure.Member
	if member.Commit == "" {
		return fmt.Errorf("%s: %w; this state pin has no revalidatable Git snapshot", DiagRepairFailed, failure)
	}
	op, err := beginOperation(req.Home, req.transactionOptions...)
	if err != nil {
		return fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	defer func() { _ = op.close() }()

	roots, err := validateResolveRoots(req.Home, req.boundaryOwnerLookup)
	if err != nil {
		return err
	}
	if !roots.store {
		return boundaryDiagnostic("profile store root", "boundary", fmt.Errorf("root %s is absent", contextstore.Root(req.Home)))
	}
	source, currentLock, currentHash, err := loadResolveInputsWithOwner(req.Home, req.Profile, req.readRegularFile, req.boundaryOwnerLookup)
	if err != nil {
		return err
	}
	if currentHash != expectedLockHash || !sameLockedMember(currentLock, member) {
		return fmt.Errorf("%s: profile lock changed before the store entry could be rebuilt", DiagRepairFailed)
	}
	if err := validateProfilePathSources(req.Profile, source, req.Policy); err != nil {
		return err
	}
	if err := validateMarkerBoundary(req.Home, req.Profile, req.EnvID, req.boundaryOwnerLookup); err != nil {
		return err
	}
	if other := validateNamedStoreBoundariesExcept(req.Home, currentLock, req.boundaryOwnerLookup, member); other != nil {
		return fmt.Errorf("%s: another store entry is untrusted: %w", DiagRepairFailed, other)
	}
	if other := validateNamedStorePinsExcept(req.Home, currentLock, member); other != nil {
		return fmt.Errorf("%s: another store entry is untrusted: %w", DiagRepairFailed, other)
	}
	snapshot, cleanup, expectedHash, err := extractPinnedSnapshot(req.Home, member)
	if err != nil {
		return fmt.Errorf("%s: revalidate pinned snapshot: %w", DiagRepairFailed, err)
	}
	defer cleanup()
	if err := publishRebuiltStoreEntry(req.Home, member, snapshot, expectedHash, req.boundaryOwnerLookup); err != nil {
		return fmt.Errorf("%s: rebuild store entry %s/%s: %w", DiagRepairFailed, member.Kind, member.Name, err)
	}
	if failure := validateNamedStoreBoundaries(req.Home, currentLock, req.boundaryOwnerLookup); failure != nil {
		return fmt.Errorf("%s: rebuilt store entry is still untrusted: %w", DiagRepairFailed, failure)
	}
	if failure := validateNamedStorePins(req.Home, currentLock); failure != nil {
		return fmt.Errorf("%s: rebuilt store entry did not match its pin: %w", DiagRepairFailed, failure)
	}
	return nil
}

func sameLockedMember(lock *contextlock.Lock, member contextlock.Member) bool {
	if lock == nil {
		return false
	}
	current, ok := lock.Find(member.Kind, member.Name)
	return ok && current.PinKey() == member.PinKey() && current.Source == member.Source
}

func validateNamedStoreBoundariesExcept(home string, lock *contextlock.Lock, owner pathboundary.OwnerLookup, skip contextlock.Member) *storeEntryFailure {
	remaining := &contextlock.Lock{Root: lock.Root}
	for _, member := range lock.Members {
		if member.Kind != skip.Kind || member.Name != skip.Name || member.PinKey() != skip.PinKey() {
			remaining.Members = append(remaining.Members, member)
		}
	}
	return validateNamedStoreBoundaries(home, remaining, owner)
}

func validateNamedStorePinsExcept(home string, lock *contextlock.Lock, skip contextlock.Member) *storeEntryFailure {
	remaining := &contextlock.Lock{Root: lock.Root}
	for _, member := range lock.Members {
		if member.Kind != skip.Kind || member.Name != skip.Name || member.PinKey() != skip.PinKey() {
			remaining.Members = append(remaining.Members, member)
		}
	}
	return validateNamedStorePins(home, remaining)
}

func publishRebuiltStoreEntry(home string, member contextlock.Member, snapshot, expectedHash string, owner pathboundary.OwnerLookup) error {
	root := contextstore.Root(home)
	target := contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey())
	parent := filepath.Dir(target)
	if err := ensureStoreEntryParent(root, parent, owner); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".curator-rebuild-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err := privatedir.Protect(stage); err != nil {
		return err
	}
	if err := copyPinnedSnapshot(snapshot, stage); err != nil {
		return err
	}
	if err := privatedir.ProtectTree(stage); err != nil {
		return err
	}
	if err := pathboundary.ValidateWithinWithOwner(root, stage, boundaryOwner(owner)); err != nil {
		return err
	}
	stageHash, err := contextstore.ContentHash(stage)
	if err != nil {
		return err
	}
	if stageHash != expectedHash {
		return fmt.Errorf("staged snapshot hash %q differs from revalidated hash %q", stageHash, expectedHash)
	}

	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	backup := filepath.Join(parent, ".curator-untrusted-"+hex.EncodeToString(nonce[:]))
	state, err := stateread.Lstat(target)
	if err != nil {
		return err
	}
	movedOld := state.Kind == stateread.KindPresent
	if movedOld {
		if err := renameManagedEntry(target, backup); err != nil {
			return err
		}
	}
	if err := renameManagedEntry(stage, target); err != nil {
		if movedOld {
			_ = renameManagedEntry(backup, target)
		}
		return err
	}
	if movedOld {
		if err := os.RemoveAll(backup); err != nil {
			return err
		}
	}
	return nil
}

func ensureStoreEntryParent(root, parent string, owner pathboundary.OwnerLookup) error {
	rootState, err := stateread.Lstat(root)
	if err != nil {
		return err
	}
	if rootState.Kind != stateread.KindPresent || rootState.Info == nil {
		return fmt.Errorf("profile store root is unavailable")
	}
	if err := pathboundary.ValidateRootWithOwner(root, boundaryOwner(owner)); err != nil {
		return err
	}
	relative, err := filepath.Rel(root, parent)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return fmt.Errorf("store entry parent %s is outside %s", parent, root)
	}
	current := root
	for _, component := range strings.Split(relative, string(os.PathSeparator)) {
		current = filepath.Join(current, component)
		state, err := stateread.Lstat(current)
		if err != nil {
			return err
		}
		if state.Kind == stateread.KindAbsent {
			if err := privatedir.Make(current); err != nil && !os.IsExist(err) {
				return err
			}
			state, err = stateread.Lstat(current)
			if err != nil {
				return err
			}
		}
		if state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.IsDir() {
			return fmt.Errorf("store entry parent component %s is not a directory", current)
		}
		if err := pathboundary.ValidateRootWithOwner(current, boundaryOwner(owner)); err != nil {
			return err
		}
	}
	return nil
}

func copyPinnedSnapshot(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			_, err := managedDirectory(destination, relative)
			return err
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("pinned snapshot component %s is not a regular file", relative)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- path is from the operation-private pinned snapshot
		if err != nil {
			return err
		}
		return atomicManagedFile(destination, relative, payload, info.Mode().Perm())
	})
}
