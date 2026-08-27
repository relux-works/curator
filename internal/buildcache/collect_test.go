package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
)

// newSweepStore builds a store whose manager home is provable protected state.
//
// Every sweep decision is derived from that boundary, so a host that cannot
// produce it — an elevated Windows session, where the OS assigns the
// Administrators group as owner of everything the process creates — can host no
// fixture at all. That is a property of the environment, not of the sweep, so
// the reason is reported and the test is skipped rather than failed.
func newSweepStore(t *testing.T) *Store {
	t.Helper()
	store := newTestStore(t)
	_, err := openProtectedDir(store.Home(), filepath.Join(store.Home(), "protected-state-probe"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Skipf("this host cannot create manager-protected state: %v", err)
	}
	return store
}

// publishTestEntry publishes one protected winner and returns its logical key.
func publishTestEntry(t *testing.T, store *Store, command, artifact string) buildmeta.CacheKey {
	t.Helper()
	input := testInput(command)
	publication, _ := testPublication(t, store.Home(), input, []byte(artifact))
	if _, err := store.Publish(publication, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func cacheRoot(store *Store) string {
	return filepath.Join(store.Home(), "cache", "build", buildmeta.DriverGoV1)
}

// openTestSweepRoot proves the cache-root boundary the same way a sweep does
// and keeps it open for the test, so a removal primitive can be exercised
// against a real handle-bound root instead of a bare pathname.
func openTestSweepRoot(t *testing.T, store *Store) *sweepRoot {
	t.Helper()
	root, err := openSweepRoot(store.Home(), cacheRoot(store))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(root.close)
	return root
}

// stageReplacementEntry publishes the same logical entry in a second protected
// home and moves it into decoy, so an adversarial test has a fully valid
// replacement to plant at the cache-root pathname mid-pass. It returns the
// entry's directory name, which is the logical key both stores derive.
func stageReplacementEntry(t *testing.T, decoy, command, artifact string) string {
	t.Helper()
	other := newSweepStore(t)
	published := entryPathOf(other, publishTestEntry(t, other, command, artifact))
	if err := createProtectedDir(decoy); err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(published)
	if err := os.Rename(published, filepath.Join(decoy, name)); err != nil {
		t.Fatal(err)
	}
	return name
}

func entryPathOf(store *Store, key buildmeta.CacheKey) string {
	return filepath.Join(cacheRoot(store), strings.TrimPrefix(string(key), "sha256:"))
}

func entryExists(t *testing.T, store *Store, key buildmeta.CacheKey) bool {
	t.Helper()
	_, err := os.Lstat(entryPathOf(store, key))
	return err == nil
}

// backdate moves an entry's recorded publication time into the past.
func backdate(t *testing.T, path string, age time.Duration) {
	t.Helper()
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func warningsContaining(result SweepResult, needle string) bool {
	for _, warning := range result.Warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}

// TestSweepRetainsReferencedAndYoungEntries proves the two retention rules a
// marker or an in-flight journal depends on.
func TestSweepRetainsReferencedAndYoungEntries(t *testing.T) {
	store := newSweepStore(t)
	referenced := publishTestEntry(t, store, "referenced-tool", "referenced artifact")
	young := publishTestEntry(t, store, "young-tool", "young artifact")
	stale := publishTestEntry(t, store, "stale-tool", "stale artifact")

	// Only the stale entry is older than the grace window.
	backdate(t, entryPathOf(store, referenced), 30*24*time.Hour)
	backdate(t, entryPathOf(store, stale), 30*24*time.Hour)

	result, err := store.Sweep(SweepRequest{Referenced: []string{string(referenced)}}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != string(stale) {
		t.Fatalf("removed = %v, warnings = %v", result.Removed, result.Warnings)
	}
	if !entryExists(t, store, referenced) {
		t.Fatal("a referenced entry was swept")
	}
	if !entryExists(t, store, young) {
		t.Fatal("an entry inside the grace window was swept")
	}
	if entryExists(t, store, stale) {
		t.Fatal("an unreferenced entry older than grace survived")
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("a clean sweep warned: %v", result.Warnings)
	}
}

// TestSweepUsesTheDocumentedDefaultGrace proves the retention window is the
// documented constant when the caller sets none.
func TestSweepUsesTheDocumentedDefaultGrace(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	backdate(t, entryPathOf(store, key), DefaultGrace-time.Hour)

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 || !entryExists(t, store, key) {
		t.Fatalf("an entry inside the default grace window was swept: %+v", result)
	}

	backdate(t, entryPathOf(store, key), DefaultGrace+time.Hour)
	result, err = store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 1 || entryExists(t, store, key) {
		t.Fatalf("an entry past the default grace window survived: %+v", result)
	}
}

// TestSweepRequiresTheHomeLock proves maintenance never mutates the cache on a
// missing or released lock witness.
func TestSweepRequiresTheHomeLock(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	backdate(t, entryPathOf(store, key), 30*24*time.Hour)
	before := treeFingerprint(t, store.Home())

	if _, err := store.Sweep(SweepRequest{}, nil); err == nil {
		t.Fatal("sweep accepted a missing lock witness")
	}
	if _, err := store.Sweep(SweepRequest{}, (*pointerTestHomeLock)(nil)); err == nil {
		t.Fatal("sweep accepted a nil pointer lock witness")
	}
	if _, err := store.Sweep(SweepRequest{}, testHomeLock{err: os.ErrClosed}); err == nil {
		t.Fatal("sweep accepted a released lock witness")
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatalf("an unlocked sweep changed manager home\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestSweepRetainsUnprovableEntries proves every uncertainty fails safe and is
// reported instead of deleted.
func TestSweepRetainsUnprovableEntries(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, store *Store, entryPath string)
		warning string
	}{
		{
			name: "corrupt receipt",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				writeFile(t, filepath.Join(entryPath, ReceiptFilename), []byte("{not a receipt}"), 0o600)
			},
			warning: "invalid receipt",
		},
		{
			name: "missing receipt",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				if err := os.Remove(filepath.Join(entryPath, ReceiptFilename)); err != nil {
					t.Fatal(err)
				}
			},
			// Unix names the absent member, Windows reports the wrapped
			// os.ErrNotExist; both say the entry is incomplete.
			warning: "cache entry is incomplete",
		},
		{
			name: "tampered artifact",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				names, err := os.ReadDir(filepath.Join(entryPath, "bin"))
				if err != nil || len(names) != 1 {
					t.Fatalf("artifact directory = %v, %v", names, err)
				}
				writeFile(t, filepath.Join(entryPath, "bin", names[0].Name()), []byte("tampered"), 0o700)
			},
			warning: "artifact",
		},
		{
			name: "unexpected member",
			mutate: func(t *testing.T, _ *Store, entryPath string) {
				writeFile(t, filepath.Join(entryPath, "unexpected"), []byte("x"), 0o600)
			},
			warning: "unexpected contents",
		},
		{
			name: "receipt of another key",
			mutate: func(t *testing.T, store *Store, entryPath string) {
				other, _ := testPublication(t, store.Home(), testInput("other-tool"), []byte("other artifact"))
				writeFile(t, filepath.Join(entryPath, ReceiptFilename), other.ReceiptBytes, 0o600)
			},
			warning: "receipt cache key does not match",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newSweepStore(t)
			key := publishTestEntry(t, store, "tool", "artifact")
			entryPath := entryPathOf(store, key)
			test.mutate(t, store, entryPath)
			backdate(t, entryPath, 30*24*time.Hour)

			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Removed) != 0 {
				t.Fatalf("an unprovable entry was removed: %v", result.Removed)
			}
			if !entryExists(t, store, key) {
				t.Fatal("an unprovable entry was deleted")
			}
			if !warningsContaining(result, test.warning) {
				t.Fatalf("missing an actionable warning for %q: %v", test.warning, result.Warnings)
			}
		})
	}
}

// TestSweepRetainsForeignAndPrivateRootMembers proves the sweep only ever
// removes something it recognizes as a complete cache entry.
func TestSweepRetainsForeignAndPrivateRootMembers(t *testing.T) {
	store := newSweepStore(t)
	publishTestEntry(t, store, "tool", "artifact")
	root := cacheRoot(store)

	foreign := filepath.Join(root, "not-a-cache-key")
	if err := os.Mkdir(foreign, 0o700); err != nil {
		t.Fatal(err)
	}
	quarantined := filepath.Join(root, ".quarantine-abc-123")
	if err := os.Mkdir(quarantined, 0o700); err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(root, ".stage-abc-123")
	if err := os.Mkdir(staging, 0o700); err != nil {
		t.Fatal(err)
	}
	shortName := filepath.Join(root, strings.Repeat("a", 63))
	if err := os.Mkdir(shortName, 0o700); err != nil {
		t.Fatal(err)
	}
	uppercase := filepath.Join(root, strings.ToUpper(strings.Repeat("a", 64)))
	if err := os.Mkdir(uppercase, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{foreign, quarantined, staging, shortName, uppercase} {
		backdate(t, path, 30*24*time.Hour)
	}

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("an unrecognized root member was removed: %v", result.Removed)
	}
	for _, path := range []string{foreign, quarantined, staging, shortName} {
		if _, err := os.Lstat(path); err != nil {
			t.Fatalf("%s was removed: %v", filepath.Base(path), err)
		}
		if !warningsContaining(result, filepath.Base(path)) {
			t.Fatalf("no warning named %s: %v", filepath.Base(path), result.Warnings)
		}
	}
	// A case-insensitive filesystem can collapse the uppercase name onto the
	// lowercase one; only assert on it where it is a distinct member.
	if _, err := os.Lstat(uppercase); err == nil && !strings.EqualFold(uppercase, shortName) {
		if !entryNameRE.MatchString(filepath.Base(uppercase)) && !warningsContaining(result, filepath.Base(uppercase)) {
			t.Fatalf("no warning named the uppercase member: %v", result.Warnings)
		}
	}
}

// TestSweepFinishesInterruptedRemovals proves an interrupted removal leaves no
// permanently orphaned tree: the entry is already unreachable by key, and the
// next sweep completes the deletion.
func TestSweepFinishesInterruptedRemovals(t *testing.T) {
	store := newSweepStore(t)
	root := cacheRoot(store)
	if err := ensureProtectedBase(store.Home(), root); err != nil {
		t.Fatal(err)
	}
	retired := filepath.Join(root, sweepPrefix+strings.Repeat("a", 64)+"-123")
	if err := os.MkdirAll(filepath.Join(retired, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}

	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(retired); err == nil {
		t.Fatal("an interrupted removal was not finished")
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("finishing an interrupted removal warned: %v", result.Warnings)
	}
}

// TestSweepReportsMalformedReferencesAndSkewedClocks proves the two inputs a
// sweep cannot trust are reported rather than silently ignored.
func TestSweepReportsMalformedReferencesAndSkewedClocks(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")

	result, err := store.Sweep(SweepRequest{
		Referenced: []string{"not-a-key", "sha256:" + strings.Repeat("A", 64), ""},
		Now:        time.Now().Add(-365 * 24 * time.Hour),
	}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 || !entryExists(t, store, key) {
		t.Fatalf("a future-dated entry was swept: %+v", result)
	}
	if !warningsContaining(result, "malformed referenced build key") {
		t.Fatalf("no malformed-reference warning: %v", result.Warnings)
	}
	if !warningsContaining(result, "in the future") {
		t.Fatalf("no clock-skew warning: %v", result.Warnings)
	}
}

// TestSweepWithoutACacheRootIsANoop proves maintenance creates no cache state.
func TestSweepWithoutACacheRootIsANoop(t *testing.T) {
	store := newSweepStore(t)
	before := treeFingerprint(t, store.Home())
	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 || len(result.Warnings) != 0 {
		t.Fatalf("sweep of an absent cache root = %+v", result)
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatalf("sweep created cache state\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestSweepRejectsUnsupportedProtectionAndNegativeGrace proves both fail closed.
func TestSweepRejectsUnsupportedProtectionAndNegativeGrace(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "tool", "artifact")
	backdate(t, entryPathOf(store, key), 30*24*time.Hour)

	if _, err := store.Sweep(SweepRequest{Grace: -time.Second}, testHomeLock{}); err == nil {
		t.Fatal("sweep accepted a negative grace period")
	}
	store.supported = func() bool { return false }
	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 0 || !entryExists(t, store, key) {
		t.Fatalf("an unsupported platform swept an entry: %+v", result)
	}
	if !warningsContaining(result, "platform protection is unavailable") {
		t.Fatalf("no unsupported-platform warning: %v", result.Warnings)
	}
}

// TestRetireEntryCannotEscapeTheCacheRoot proves the removal primitive refuses
// any name that is not a direct child of the validated cache root.
func TestRetireEntryCannotEscapeTheCacheRoot(t *testing.T) {
	store := newSweepStore(t)
	if err := ensureProtectedBase(store.Home(), cacheRoot(store)); err != nil {
		t.Fatal(err)
	}
	root := openTestSweepRoot(t, store)
	outside := filepath.Join(store.Home(), "outside")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"..", ".", "", "../outside", "sub/child", filepath.Join("..", "outside")} {
		if err := retireEntry(root, name); err == nil {
			t.Fatalf("retire accepted the escaping name %q", name)
		}
	}
	if _, err := os.Lstat(outside); err != nil {
		t.Fatalf("a directory outside the cache root was removed: %v", err)
	}
}

// TestRetireEntryRefusesAnUnboundRoot proves the removal primitive does nothing
// at all without a proven, mutation-bound cache root.
func TestRetireEntryRefusesAnUnboundRoot(t *testing.T) {
	name := strings.Repeat("a", 64)
	if err := retireEntry(nil, name); err == nil {
		t.Fatal("retire accepted a missing cache root")
	}
	if err := retireEntry(&sweepRoot{path: t.TempDir()}, name); err == nil {
		t.Fatal("retire accepted a cache root that is not open for mutation")
	}
}

// TestSweepRefusesACacheRootExchangedBeforeMutation proves the mutator is
// accepted only when it names the very directory object the sweep proved.
func TestSweepRefusesACacheRootExchangedBeforeMutation(t *testing.T) {
	store := newSweepStore(t)
	if err := ensureProtectedBase(store.Home(), cacheRoot(store)); err != nil {
		t.Fatal(err)
	}
	root := openTestSweepRoot(t, store)
	other, err := openSweepRoot(store.Home(), cacheRoot(store))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(other.close)

	// A cache root whose recorded identity belongs to a different directory is
	// exactly what a mid-pass pathname exchange looks like from the inside.
	elsewhere := filepath.Join(store.Home(), "elsewhere")
	if err := os.MkdirAll(elsewhere, 0o700); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(elsewhere)
	if err != nil {
		t.Fatal(err)
	}
	other.identity = info
	if err := other.assertParentOf(&protectedDir{parents: []*os.File{root.validated.dir}}); err == nil {
		t.Fatal("an entry outside the proven root was accepted")
	}
	if err := other.assertParentOf(&protectedDir{}); err == nil {
		t.Fatal("an entry without a proven parent was accepted")
	}
}

// TestEntryNameRejectsNonCanonicalKeys pins the logical-key spelling the mark
// phase and the entry layout have to agree on.
func TestEntryNameRejectsNonCanonicalKeys(t *testing.T) {
	valid := "sha256:" + strings.Repeat("a", 64)
	if name, ok := entryName(valid); !ok || name != strings.Repeat("a", 64) {
		t.Fatalf("entryName(%q) = %q, %v", valid, name, ok)
	}
	for _, key := range []string{
		"", "sha256:", strings.Repeat("a", 64),
		"sha256:" + strings.Repeat("A", 64),
		"sha256:" + strings.Repeat("a", 63),
		"sha256:" + strings.Repeat("a", 65),
		"sha512:" + strings.Repeat("a", 64),
		"sha256:" + strings.Repeat("a", 63) + "/",
	} {
		if _, ok := entryName(key); ok {
			t.Fatalf("entryName accepted %q", key)
		}
	}
}
