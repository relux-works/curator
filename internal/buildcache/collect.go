package buildcache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
)

// DefaultGrace is Curator's documented build-cache grace period. An
// unreferenced protected entry younger than this is always retained.
//
// A swept entry costs a rebuild, never correctness, so the window only has to
// exceed any operation that could still be publishing while holding no journal
// — a crash between publication and journal preparation is the only such case.
// One day is far beyond that and still bounds how long orphaned artifacts
// occupy disk.
const DefaultGrace = 24 * time.Hour

// sweepPrefix names a cache entry this collector already retired. The rename
// is the atomic removal; the tree deletion behind it is resumable cleanup.
const sweepPrefix = ".sweep-"

var entryNameRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

// beforeRetireForTests runs immediately before the first retirement of a pass.
// It is the seam an adversarial test uses to exchange the cache-root pathname
// after validation and prove that the removal still cannot leave the directory
// object this sweep proved. It is nil in every production build.
var beforeRetireForTests func()

// beforeClassifyForTests runs after a candidate's boundary and receipt have
// been proven and immediately before the decisive classification of that same
// candidate. It is the seam an adversarial test uses to exchange the cache-root
// pathname in exactly the window where a pathname-based inspection would
// validate the replacement instead of the entry it is about to remove. It is
// nil in every production build.
var beforeClassifyForTests func()

// SweepRequest describes one locked maintenance sweep of the protected cache.
// The caller owns the mark phase: this package never reads a marker, a journal,
// or a consumer, and never treats receipt content as a live reference.
type SweepRequest struct {
	// Referenced holds every logical build key a valid marker v2 or an
	// in-flight transaction journal names, as "sha256:" + 64 lowercase hex.
	Referenced []string
	// Now is the maintenance clock; the zero value means time.Now().
	Now time.Time
	// Grace retains an unreferenced entry published less than this long ago.
	// The zero value means DefaultGrace.
	Grace time.Duration
}

// SweepResult reports the removed logical keys and every uncertainty an
// operator has to resolve. A warning always means something was retained.
type SweepResult struct {
	Removed  []string
	Warnings []string
}

// Sweep removes unreferenced protected cache entries older than the grace
// period while the caller holds the manager-home mutation lock.
//
// Every decision fails safe. The cache root is revalidated as protected state
// before traversal; an entry is removed only when it is itself protected,
// structurally exact, and self-consistent with the logical key its directory
// name encodes; anything unprovable is retained and reported. Entry content is
// never executed, adopted, or permission-repaired.
func (store *Store) Sweep(request SweepRequest, lock HomeLock) (SweepResult, error) {
	if err := requireHomeLock(lock); err != nil {
		return SweepResult{}, err
	}
	if request.Grace < 0 {
		return SweepResult{}, fmt.Errorf("build cache grace period must not be negative")
	}
	if store == nil || store.supported == nil || !store.supported() {
		return SweepResult{Warnings: []string{
			"build cache sweep skipped: platform protection is unavailable, so no entry can be proven safe to remove",
		}}, nil
	}
	grace := request.Grace
	if grace == 0 {
		grace = DefaultGrace
	}
	now := request.Now
	if now.IsZero() {
		now = time.Now()
	}

	result := SweepResult{}
	referenced := make(map[string]bool, len(request.Referenced))
	for _, key := range request.Referenced {
		name, ok := entryName(key)
		if !ok {
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"build cache sweep: ignored malformed referenced build key %q; repair the marker or journal that names it", key))
			continue
		}
		referenced[name] = true
	}

	base := filepath.Join(store.home, "cache", "build", buildmeta.DriverGoV1)
	if !pathWithin(store.home, base) {
		return result, fmt.Errorf("cache path crosses the manager-home boundary")
	}
	root, err := openSweepRoot(store.home, base)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep skipped: %v; no entry was traversed or removed", err))
		return result, nil
	}
	// The proven root stays open, and every mutation below goes through the
	// handle-relative mutator bound to it, so no removal can follow a pathname
	// that was exchanged after the boundary was proven.
	defer root.close()

	names, err := directoryNames(root.validated.dir)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep skipped: list protected cache root: %v", err))
		return result, nil
	}
	for _, name := range names {
		switch {
		case entryNameRE.MatchString(name):
			if referenced[name] {
				continue
			}
			store.sweepUnreferenced(root, name, now, grace, &result)
		case strings.HasPrefix(name, sweepPrefix):
			// Manager-private wreckage of an interrupted removal: the entry is
			// already unreachable by key, so finishing the deletion is safe.
			if err := root.mutator.RemoveAll(name); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"build cache sweep: could not finish removing retired entry %q: %v", name, err))
			}
		default:
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"build cache sweep: retained unrecognized cache root member %q; inspect it and remove it manually", name))
		}
	}
	return result, nil
}

// sweepUnreferenced classifies one unreferenced candidate and removes it only
// when every safety property holds.
func (store *Store) sweepUnreferenced(root *sweepRoot, name string, now time.Time, grace time.Duration, result *SweepResult) {
	key := buildmeta.CacheKey("sha256:" + name)
	published, inspected, err := store.inspectUnexpected(root, name, key)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep: retained %s: %v", key, err))
		return
	}
	if inspected.Status != Hit {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep: retained %s (%s): %s", key, inspected.Status, inspected.Reason))
		return
	}
	if published.After(now) {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep: retained %s: its publication time %s is in the future; check the system clock",
			key, published.UTC().Format(time.RFC3339)))
		return
	}
	if now.Sub(published) <= grace {
		return
	}
	if hook := beforeRetireForTests; hook != nil {
		hook()
	}
	if err := retireEntry(root, name); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"build cache sweep: could not remove %s: %v", key, err))
		return
	}
	result.Removed = append(result.Removed, string(key))
}

// inspectUnexpected validates an entry whose logical input maintenance does not
// know in advance, and reports when it was published.
//
// Provenance stays with the protected boundary. The entry is re-resolved
// without following a link at any component, its parent has to be the very
// directory object this sweep proved, the receipt is read only to name the
// input the entry claims, and that input is then independently re-derived into
// a cache key which must equal the directory name. The full read-only
// inspection re-checks the exact structure, receipt, artifact hash, and size,
// so a receipt that is merely internally consistent proves nothing.
//
// The decisive inspection runs on the descriptor proved here, never on the
// pathname it came from. That is what makes the classification and the removal
// speak about the same directory object: an exchange of the cache-root pathname
// in this window can no longer let a planted replacement validate while the
// unproven original is the one that gets retired.
func (store *Store) inspectUnexpected(root *sweepRoot, name string, key buildmeta.CacheKey) (time.Time, Result, error) {
	entryPath := filepath.Join(root.path, name)
	opened, err := openProtectedDir(store.home, entryPath)
	if err != nil {
		return time.Time{}, Result{}, err
	}
	defer opened.close()
	if err := root.assertParentOf(opened); err != nil {
		return time.Time{}, Result{}, err
	}
	info, err := opened.dir.Stat()
	if err != nil {
		return time.Time{}, Result{}, fmt.Errorf("stat cache entry: %w", err)
	}
	receiptBytes, err := readProtectedChild(opened.dir, ReceiptFilename)
	if err != nil {
		return time.Time{}, Result{}, err
	}
	receipt, err := buildmeta.DecodeReceipt(receiptBytes)
	if err != nil {
		return time.Time{}, Result{}, fmt.Errorf("invalid receipt: %w", err)
	}
	executionBytes, err := readProtectedChild(opened.dir, ExecutionReceiptFilename)
	if err != nil {
		return time.Time{}, Result{}, err
	}
	execution, err := closureexec.DecodeBuildSessionReceipt(executionBytes)
	if err != nil {
		return time.Time{}, Result{}, fmt.Errorf("invalid execution receipt: %w", err)
	}
	assuredID, err := (closureexec.AssuredBuildCacheInput{BuildInput: receipt.Input, Binding: execution.Binding}).ID()
	if err != nil || buildmeta.CacheKey(assuredID) != key {
		return time.Time{}, Result{}, fmt.Errorf("assured cache input does not match the entry key")
	}
	if hook := beforeClassifyForTests; hook != nil {
		hook()
	}
	return info.ModTime(), inspectProvenEntry(opened, entryPath, Expectation{Input: receipt.Input, Assurance: execution.Binding}, key), nil
}

func readProtectedChild(dir *os.File, name string) ([]byte, error) {
	file, err := openProtectedChildFile(dir, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	payload, err := readExactFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return payload, nil
}

// retireEntry removes one direct child of the proven cache root through the
// mutator bound to that root. The rename is the atomic removal: afterwards no
// reader can reach the entry by its key, and because both names are resolved
// relative to the proven directory object rather than by pathname, neither the
// rename nor the deletion behind it can act on a directory that replaced the
// cache root after it was validated.
func retireEntry(root *sweepRoot, name string) error {
	if root == nil || root.mutator == nil {
		return fmt.Errorf("the protected cache root is not open for mutation")
	}
	if !isDirectChildName(name) {
		return fmt.Errorf("cache entry name is not a direct root child")
	}
	retired, err := reserveRetiredName(root.mutator, name)
	if err != nil {
		return err
	}
	if err := root.mutator.Rename(name, retired); err != nil {
		return fmt.Errorf("retire cache entry: %w", err)
	}
	if err := syncDirHandle(root.validated.dir); err != nil {
		return fmt.Errorf("sync cache root: %w", err)
	}
	if err := root.mutator.RemoveAll(retired); err != nil {
		// The entry is already unreachable by key. The leftover is named for
		// the next sweep to finish, so this is reported, not failed.
		return fmt.Errorf("remove retired cache entry %q: %w", retired, err)
	}
	return syncDirHandle(root.validated.dir)
}

// isDirectChildName reports whether name addresses one member of a directory
// and nothing else — no traversal, no separator, no self or parent reference.
func isDirectChildName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsRune(name, '/') && !strings.ContainsRune(name, filepath.Separator)
}

// reserveRetiredName picks a private name that does not exist inside the proven
// root. The name is only ever resolved through the root handle, so it cannot
// collide with anything outside the boundary.
func reserveRetiredName(root *os.Root, name string) (string, error) {
	for attempt := 0; attempt < 16; attempt++ {
		var suffix [8]byte
		if _, err := rand.Read(suffix[:]); err != nil {
			return "", fmt.Errorf("reserve a private cache name: %w", err)
		}
		candidate := sweepPrefix + name + "-" + hex.EncodeToString(suffix[:])
		_, err := root.Lstat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("reserve a private cache name: %w", err)
		}
	}
	return "", fmt.Errorf("could not reserve a private cache name")
}

// entryName maps a logical build key to its cache directory name.
func entryName(key string) (string, bool) {
	name, found := strings.CutPrefix(key, "sha256:")
	if !found || !entryNameRE.MatchString(name) {
		return "", false
	}
	return name, true
}

// sweepRoot binds one proven protected cache root to the handle-relative
// mutator that every removal of the same pass has to go through.
//
// The validated handles prove the boundary — type, owner, and mode or DACL of
// every component, resolved without following a link. The mutator is opened on
// the same directory and accepted only when it names the very same directory
// object, so an exchange of the cache-root pathname between validation and
// mutation is refused rather than followed.
type sweepRoot struct {
	path      string
	validated *protectedDir
	identity  os.FileInfo
	mutator   *os.Root
}

func openSweepRoot(home, base string) (*sweepRoot, error) {
	validated, err := openProtectedDir(home, base)
	if err != nil {
		return nil, err
	}
	identity, err := validated.dir.Stat()
	if err != nil {
		validated.close()
		return nil, untrustedf("stat the protected cache root: %v", err)
	}
	mutator, err := os.OpenRoot(base)
	if err != nil {
		validated.close()
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, untrustedf("open the protected cache root for mutation: %v", err)
	}
	mutatorInfo, err := mutator.Stat(".")
	if err != nil {
		_ = mutator.Close()
		validated.close()
		return nil, untrustedf("stat the protected cache root for mutation: %v", err)
	}
	if !os.SameFile(identity, mutatorInfo) {
		_ = mutator.Close()
		validated.close()
		return nil, untrustedf("the protected cache root changed identity between validation and mutation")
	}
	return &sweepRoot{path: base, validated: validated, identity: identity, mutator: mutator}, nil
}

// assertParentOf refuses an entry whose parent is not the directory object this
// sweep proved, which is what a cache root exchanged mid-pass looks like.
func (root *sweepRoot) assertParentOf(entry *protectedDir) error {
	parent := entry.parent()
	if parent == nil {
		return untrustedf("cache entry has no proven parent directory")
	}
	info, err := parent.Stat()
	if err != nil {
		return untrustedf("stat the parent of a cache entry: %v", err)
	}
	if !os.SameFile(info, root.identity) {
		return untrustedf("cache entry no longer lives in the proven cache root")
	}
	return nil
}

func (root *sweepRoot) close() {
	if root == nil {
		return
	}
	if root.mutator != nil {
		_ = root.mutator.Close()
	}
	root.validated.close()
}

// protectedDir is a validated directory handle together with the parent handles
// that prove the path was resolved without following a link at any component.
type protectedDir struct {
	dir     *os.File
	parents []*os.File
}

// parent returns the handle of the directory that directly contains dir.
func (opened *protectedDir) parent() *os.File {
	if opened == nil || len(opened.parents) == 0 {
		return nil
	}
	return opened.parents[len(opened.parents)-1]
}

func (opened *protectedDir) close() {
	if opened == nil {
		return
	}
	if opened.dir != nil {
		_ = opened.dir.Close()
	}
	for index := len(opened.parents) - 1; index >= 0; index-- {
		if opened.parents[index] != nil {
			_ = opened.parents[index].Close()
		}
	}
}
