//go:build unix

package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
)

// TestSweepRetainsUntrustedUnixState proves that ownership and permission
// failures anywhere on the traversal path stop the sweep instead of widening
// it, and that nothing is repaired on the way.
func TestSweepRetainsUntrustedUnixState(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(t *testing.T, store *Store, entryPath string)
		warning  string
		rootWide bool
	}{
		{
			name: "entry group writable",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				chmodTest(t, entryPath, 0o770)
			},
			warning: "writable by group or other",
		},
		{
			name: "artifact directory other writable",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				chmodTest(t, filepath.Join(entryPath, "bin"), 0o707)
			},
			warning: "writable by group or other",
		},
		{
			name: "receipt group writable",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				chmodTest(t, filepath.Join(entryPath, ReceiptFilename), 0o660)
			},
			warning: "writable by group or other",
		},
		{
			name: "cache root group writable",
			mutate: func(t *testing.T, store *Store, _ string) {
				chmodTest(t, cacheRoot(store), 0o770)
			},
			warning:  "build cache sweep skipped",
			rootWide: true,
		},
		{
			name: "manager home other writable",
			mutate: func(t *testing.T, store *Store, _ string) {
				chmodTest(t, store.Home(), 0o707)
			},
			warning:  "build cache sweep skipped",
			rootWide: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newSweepStore(t)
			key := publishTestEntry(t, store, "tool", "artifact")
			entryPath := entryPathOf(store, key)
			backdate(t, entryPath, 30*24*time.Hour)
			test.mutate(t, store, entryPath)
			t.Cleanup(func() { chmodTest(t, store.Home(), 0o700) })

			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Removed) != 0 {
				t.Fatalf("untrusted state was swept: %v", result.Removed)
			}
			if !entryExists(t, store, key) {
				t.Fatal("untrusted state was deleted")
			}
			if !warningsContaining(result, test.warning) {
				t.Fatalf("missing warning %q: %v", test.warning, result.Warnings)
			}
			if test.rootWide && len(result.Warnings) != 1 {
				t.Fatalf("an unprovable root reported per-entry detail: %v", result.Warnings)
			}
		})
	}
}

// TestSweepDoesNotFollowSymlinkedEntries proves a link planted in the cache
// root is never traversed and never deletes its target.
func TestSweepDoesNotFollowSymlinkedEntries(t *testing.T) {
	store := newSweepStore(t)
	published := publishTestEntry(t, store, "tool", "artifact")
	publishedPath := entryPathOf(store, published)
	backdate(t, publishedPath, 30*24*time.Hour)

	// A second, well-formed entry name that is only a link to the first.
	linkName := "b" + filepath.Base(publishedPath)[1:]
	if linkName == filepath.Base(publishedPath) {
		linkName = "c" + filepath.Base(publishedPath)[1:]
	}
	linkPath := filepath.Join(cacheRoot(store), linkName)
	if err := os.Symlink(publishedPath, linkPath); err != nil {
		t.Fatal(err)
	}

	result, err := store.Sweep(SweepRequest{Referenced: []string{string(published)}}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("a symlinked entry was swept: %v", result.Removed)
	}
	if !entryExists(t, store, published) {
		t.Fatal("the referenced link target was removed")
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("the link itself was removed without proof: %v", err)
	}
	if !warningsContaining(result, "retained") {
		t.Fatalf("the symlinked entry was not reported: %v", result.Warnings)
	}
}

// TestSweepDoesNotFollowASymlinkedCacheRoot proves a redirected cache root is
// refused rather than traversed.
func TestSweepDoesNotFollowASymlinkedCacheRoot(t *testing.T) {
	store := newSweepStore(t)
	elsewhere := filepath.Join(store.Home(), "elsewhere")
	if err := os.MkdirAll(filepath.Join(elsewhere, "victim"), 0o700); err != nil {
		t.Fatal(err)
	}
	driverRoot := cacheRoot(store)
	if err := os.MkdirAll(filepath.Dir(driverRoot), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, driverRoot); err != nil {
		t.Fatal(err)
	}

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("a symlinked cache root was swept: %v", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(elsewhere, "victim")); err != nil {
		t.Fatalf("a directory behind the redirected root was removed: %v", err)
	}
	if !warningsContaining(result, "build cache sweep skipped") {
		t.Fatalf("the redirected root was not reported: %v", result.Warnings)
	}
}

// TestSweepRemovalDoesNotFollowLinksInsideAnEntry proves that removing one
// entry cannot delete anything outside the Curator cache root.
func TestSweepRemovalDoesNotFollowLinksInsideAnEntry(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	entryPath := entryPathOf(store, key)

	outside := filepath.Join(store.Home(), "outside")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(outside, "keep.txt"), []byte("keep"), 0o600)

	// The entry is no longer structurally exact, so it must be retained; the
	// planted link must survive untraversed either way.
	if err := os.Symlink(outside, filepath.Join(entryPath, "escape")); err != nil {
		t.Fatal(err)
	}
	backdate(t, entryPath, 30*24*time.Hour)

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("an entry holding a link was swept: %v", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(outside, "keep.txt")); err != nil {
		t.Fatalf("a file outside the cache root was removed: %v", err)
	}
	if !warningsContaining(result, "unexpected contents") {
		t.Fatalf("the deviating entry was not reported: %v", result.Warnings)
	}
}

// TestRetireEntryRemovesLinksWithoutFollowingThem proves the removal primitive
// itself, on an entry that does contain a link, never deletes the link target.
func TestRetireEntryRemovesLinksWithoutFollowingThem(t *testing.T) {
	store := newSweepStore(t)
	root := cacheRoot(store)
	if err := ensureProtectedBase(store.Home(), root); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(store.Home(), "outside")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(outside, "keep.txt"), []byte("keep"), 0o600)

	name := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	entryPath := filepath.Join(root, name)
	if err := os.Mkdir(entryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(entryPath, "escape")); err != nil {
		t.Fatal(err)
	}
	if err := retireEntry(openTestSweepRoot(t, store), name); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(entryPath); err == nil {
		t.Fatal("the entry survived its removal")
	}
	if _, err := os.Lstat(filepath.Join(outside, "keep.txt")); err != nil {
		t.Fatalf("removal followed the link out of the cache root: %v", err)
	}
}

// TestSweepRemovalSurvivesACacheRootExchangedMidPass is the adversarial
// boundary-swap case: the cache-root pathname is exchanged for a replacement
// tree after the boundary was proven but before the first removal.
//
// An open descriptor does not pin later pathname resolution on Unix, so this is
// exactly the window in which a pathname-based rename or deletion would act
// inside the replacement. The removal has to stay with the directory object the
// sweep proved: the replacement must be untouched, the proven entry must be the
// one that goes, and every candidate that now resolves outside the proven root
// must be retained and reported.
func TestSweepRemovalSurvivesACacheRootExchangedMidPass(t *testing.T) {
	store := newSweepStore(t)
	first := publishTestEntry(t, store, "first-tool", "first artifact")
	second := publishTestEntry(t, store, "second-tool", "second artifact")
	names := []string{filepath.Base(entryPathOf(store, first)), filepath.Base(entryPathOf(store, second))}
	for _, key := range []buildmeta.CacheKey{first, second} {
		backdate(t, entryPathOf(store, key), 30*24*time.Hour)
	}

	root := cacheRoot(store)
	moved := filepath.Join(filepath.Dir(root), "moved-away")
	decoy := filepath.Join(store.Home(), "decoy")
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(decoy, name, "bin"), 0o700); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(decoy, name, "keep.txt"), []byte("keep"), 0o600)
	}

	var swapErr error
	swapped := false
	beforeRetireForTests = func() {
		if swapped {
			return
		}
		swapped = true
		if err := os.Rename(root, moved); err != nil {
			swapErr = err
			return
		}
		if err := os.Rename(decoy, root); err != nil {
			swapErr = err
		}
	}
	t.Cleanup(func() { beforeRetireForTests = nil })

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if swapErr != nil {
		t.Fatalf("the boundary swap could not be staged: %v", swapErr)
	}
	if !swapped {
		t.Fatal("no removal was attempted, so the swap window was never reached")
	}
	// Nothing in the replacement tree, now sitting at the cache-root pathname,
	// was renamed or deleted.
	for _, name := range names {
		if _, err := os.Lstat(filepath.Join(root, name, "keep.txt")); err != nil {
			t.Fatalf("the sweep acted inside the replacement tree: %v", err)
		}
	}
	// The removal followed the proven directory object instead.
	if len(result.Removed) != 1 {
		t.Fatalf("removed = %v, warnings = %v", result.Removed, result.Warnings)
	}
	removedName := strings.TrimPrefix(result.Removed[0], "sha256:")
	if _, err := os.Lstat(filepath.Join(moved, removedName)); err == nil {
		t.Fatal("the proven entry survived while the replacement was at risk")
	}
	if !warningsContaining(result, "no longer lives in the proven cache root") {
		t.Fatalf("the exchanged boundary was not reported: %v", result.Warnings)
	}
}

// TestSweepClassificationSurvivesACacheRootExchangedMidPass is the adversarial
// classification swap: the cache-root pathname is exchanged for a replacement
// tree after the candidate's boundary and receipt are proven, in the window
// where the decisive inspection runs.
//
// The candidate on disk is structurally unexpected, so an honest classification
// must retain it. The replacement carries a perfectly valid entry under the same
// name. A classification that re-resolved the pathname would inspect the
// replacement, call the candidate a hit, and then retire the unproven original
// through the mutator still bound to the proven root — validating one directory
// and deleting another. Reading the classification through the proven entry
// descriptor is what makes both halves speak about the same object.
func TestSweepClassificationSurvivesACacheRootExchangedMidPass(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	entryPath := entryPathOf(store, key)
	name := filepath.Base(entryPath)

	// The original is no longer structurally exact, so it must be retained. It
	// is backdated afterwards: writing into the directory refreshes its
	// publication time, and an entry inside the grace window would be retained
	// for a reason that proves nothing about classification.
	writeFile(t, filepath.Join(entryPath, "unexpected.txt"), []byte("x"), 0o600)
	backdate(t, entryPath, 30*24*time.Hour)

	root := cacheRoot(store)
	moved := filepath.Join(filepath.Dir(root), "moved-away")
	decoy := filepath.Join(store.Home(), "decoy")
	if replacement := stageReplacementEntry(t, decoy, "tool", "artifact"); replacement != name {
		t.Fatalf("the replacement entry is named %q, not %q", replacement, name)
	}

	var swapErr error
	swapped := false
	beforeClassifyForTests = func() {
		if swapped {
			return
		}
		swapped = true
		if swapErr = os.Rename(root, moved); swapErr != nil {
			return
		}
		swapErr = os.Rename(decoy, root)
	}
	t.Cleanup(func() { beforeClassifyForTests = nil })

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if swapErr != nil {
		t.Fatalf("the boundary swap could not be staged: %v", swapErr)
	}
	if !swapped {
		t.Fatal("no classification was attempted, so the swap window was never reached")
	}
	if len(result.Removed) != 0 {
		t.Fatalf("an entry was retired on someone else's proof: %v", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(moved, name, ReceiptFilename)); err != nil {
		t.Fatalf("the unproven original was removed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, name, ReceiptFilename)); err != nil {
		t.Fatalf("the sweep acted inside the replacement tree: %v", err)
	}
	if !warningsContaining(result, "unexpected contents") {
		t.Fatalf("the retained candidate was not reported: %v", result.Warnings)
	}

	// Negative control: the pathname a re-resolving classification would have
	// opened now names the replacement, and the replacement classifies as a
	// reusable hit. That is exactly the verdict the handle-bound classification
	// refused to borrow for the entry it was about to remove.
	control := store.inspectEntry(entryPathOf(store, key), Expectation{Input: testInput("tool")}, key)
	if control.Status != Hit {
		t.Fatalf("the control did not reproduce the reopen hazard: %s: %s", control.Status, control.Reason)
	}
}

func chmodTest(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}
