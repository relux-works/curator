//go:build windows

package buildcache

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// linkDirectoryForTest creates a directory reparse point, preferring a symbolic
// link and falling back to a junction. A junction needs no privilege, so the
// reparse cases below run on a plain local account instead of skipping the
// coverage the acceptance criteria ask for.
func linkDirectoryForTest(t *testing.T, target, link string) {
	t.Helper()
	symlinkErr := os.Symlink(target, link)
	if symlinkErr == nil {
		return
	}
	t.Logf("symbolic links are unavailable to this account (%v); using a junction", symlinkErr)
	output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Skipf("this host cannot create a directory reparse point: %v: %s", err, output)
	}
}

// TestSweepRetainsUntrustedWindowsState proves an ownership or DACL failure
// anywhere on the traversal path stops the sweep instead of widening it, and
// that nothing is repaired on the way.
func TestSweepRetainsUntrustedWindowsState(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(t *testing.T, store *Store, entryPath string)
		warning  string
		rootWide bool
	}{
		{
			name: "entry grants another principal mutation rights",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				grantWorldMutation(t, entryPath)
			},
			warning: "another principal",
		},
		{
			name: "receipt grants another principal mutation rights",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				grantWorldMutation(t, filepath.Join(entryPath, ReceiptFilename))
			},
			warning: "another principal",
		},
		{
			name: "cache root grants another principal mutation rights",
			mutate: func(t *testing.T, store *Store, _ string) {
				grantWorldMutation(t, cacheRoot(store))
			},
			warning:  "build cache sweep skipped",
			rootWide: true,
		},
		{
			name: "manager home grants another principal mutation rights",
			mutate: func(t *testing.T, store *Store, _ string) {
				grantWorldMutation(t, store.Home())
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

// TestSweepDoesNotFollowAReparsedEntry proves a reparse point planted in the
// cache root is never traversed and never deletes its target.
func TestSweepDoesNotFollowAReparsedEntry(t *testing.T) {
	store := newSweepStore(t)
	published := publishTestEntry(t, store, "tool", "artifact")
	publishedPath := entryPathOf(store, published)
	backdate(t, publishedPath, 30*24*time.Hour)

	linkName := "b" + filepath.Base(publishedPath)[1:]
	if linkName == filepath.Base(publishedPath) {
		linkName = "c" + filepath.Base(publishedPath)[1:]
	}
	linkPath := filepath.Join(cacheRoot(store), linkName)
	linkDirectoryForTest(t, publishedPath, linkPath)

	result, err := store.Sweep(SweepRequest{Referenced: []string{string(published)}}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("a reparsed entry was swept: %v", result.Removed)
	}
	if !entryExists(t, store, published) {
		t.Fatal("the referenced reparse target was removed")
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("the reparse point itself was removed without proof: %v", err)
	}
	if !warningsContaining(result, "retained") {
		t.Fatalf("the reparsed entry was not reported: %v", result.Warnings)
	}
}

// TestSweepRemovalSurvivesACacheRootExchangedMidPass is the Windows half of the
// adversarial boundary swap: the cache-root pathname is exchanged for a
// replacement tree after the boundary was proven but before the first removal.
//
// Windows answers this one component earlier than Unix does. The proven
// handles are opened without FILE_SHARE_DELETE, so the cache root and every
// component above it cannot be renamed or deleted while the sweep holds them,
// and the swap is refused by the OS. The test accepts either outcome and
// insists on the same guarantee in both: nothing in the replacement tree is
// touched.
func TestSweepRemovalSurvivesACacheRootExchangedMidPass(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	entryPath := entryPathOf(store, key)
	name := filepath.Base(entryPath)
	backdate(t, entryPath, 30*24*time.Hour)

	root := cacheRoot(store)
	moved := filepath.Join(filepath.Dir(root), "moved-away")
	decoy := filepath.Join(store.Home(), "decoy")
	if err := os.MkdirAll(filepath.Join(decoy, name, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(decoy, name, "keep.txt"), []byte("keep"), 0o600)

	var swapErr error
	swapped := false
	beforeRetireForTests = func() {
		if swapped {
			return
		}
		swapped = true
		if swapErr = os.Rename(root, moved); swapErr != nil {
			return
		}
		swapErr = os.Rename(decoy, root)
	}
	t.Cleanup(func() { beforeRetireForTests = nil })

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if !swapped {
		t.Fatal("no removal was attempted, so the swap window was never reached")
	}
	if swapErr != nil {
		t.Logf("the proven cache root could not be exchanged at all: %v", swapErr)
		if _, err := os.Lstat(filepath.Join(decoy, name, "keep.txt")); err != nil {
			t.Fatalf("the replacement tree was touched: %v", err)
		}
	} else if _, err := os.Lstat(filepath.Join(root, name, "keep.txt")); err != nil {
		t.Fatalf("the sweep acted inside the replacement tree: %v", err)
	}
	if len(result.Removed) != 1 {
		t.Fatalf("removed = %v, warnings = %v", result.Removed, result.Warnings)
	}
	if _, err := os.Lstat(entryPathOf(store, key)); err == nil && swapErr != nil {
		t.Fatal("the proven entry survived its own removal")
	}
}

// TestSweepClassificationSurvivesACacheRootExchangedMidPass is the Windows half
// of the adversarial classification swap: the cache-root pathname is exchanged
// for a replacement tree in the window between proving a candidate and running
// the decisive inspection on it.
//
// Windows answers this one layer lower than Unix: every component down to the
// candidate is held open without FILE_SHARE_DELETE, so the OS usually refuses
// the exchange outright. The test accepts either outcome and insists on the
// same guarantee in both — the structurally unexpected original is retained and
// reported, and nothing in the replacement tree is touched.
func TestSweepClassificationSurvivesACacheRootExchangedMidPass(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	entryPath := entryPathOf(store, key)
	name := filepath.Base(entryPath)

	// Backdated after the write, so the entry cannot be retained merely for
	// being young.
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
	if !swapped {
		t.Fatal("no classification was attempted, so the swap window was never reached")
	}
	if len(result.Removed) != 0 {
		t.Fatalf("an entry was retired on someone else's proof: %v", result.Removed)
	}
	if swapErr != nil {
		t.Logf("the proven cache root could not be exchanged at all: %v", swapErr)
		if _, err := os.Lstat(filepath.Join(decoy, name, ReceiptFilename)); err != nil {
			t.Fatalf("the replacement tree was touched: %v", err)
		}
		if _, err := os.Lstat(filepath.Join(entryPath, ReceiptFilename)); err != nil {
			t.Fatalf("the unproven original was removed: %v", err)
		}
	} else {
		if _, err := os.Lstat(filepath.Join(root, name, ReceiptFilename)); err != nil {
			t.Fatalf("the sweep acted inside the replacement tree: %v", err)
		}
		if _, err := os.Lstat(filepath.Join(moved, name, ReceiptFilename)); err != nil {
			t.Fatalf("the unproven original was removed: %v", err)
		}
	}
	if !warningsContaining(result, "unexpected contents") {
		t.Fatalf("the retained candidate was not reported: %v", result.Warnings)
	}
}

// TestSweepDoesNotFollowAReparsedCacheRoot proves a redirected cache root is
// refused rather than traversed.
func TestSweepDoesNotFollowAReparsedCacheRoot(t *testing.T) {
	store := newSweepStore(t)
	elsewhere := filepath.Join(store.Home(), "elsewhere")
	if err := os.MkdirAll(filepath.Join(elsewhere, "victim"), 0o700); err != nil {
		t.Fatal(err)
	}
	driverRoot := cacheRoot(store)
	if err := os.MkdirAll(filepath.Dir(driverRoot), 0o700); err != nil {
		t.Fatal(err)
	}
	linkDirectoryForTest(t, elsewhere, driverRoot)

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("a reparsed cache root was swept: %v", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(elsewhere, "victim")); err != nil {
		t.Fatalf("a directory behind the redirected root was removed: %v", err)
	}
	if !warningsContaining(result, "build cache sweep skipped") {
		t.Fatalf("the redirected root was not reported: %v", result.Warnings)
	}
}
