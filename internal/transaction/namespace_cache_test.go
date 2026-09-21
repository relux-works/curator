package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/staging"
)

// countingResolver wraps canonicalNamespacePath and records every walk by
// exact input path. A cache hit never reaches the wrapped walk.
type countingResolver struct {
	mu      sync.Mutex
	calls   map[string]int
	resolve func(string) (string, error)
}

func newCountingResolver() *countingResolver {
	return &countingResolver{
		calls:   map[string]int{},
		resolve: canonicalNamespacePath,
	}
}

func (counter *countingResolver) fn(path string) (string, error) {
	counter.mu.Lock()
	counter.calls[path]++
	counter.mu.Unlock()
	return counter.resolve(path)
}

func (counter *countingResolver) total() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	total := 0
	for _, count := range counter.calls {
		total += count
	}
	return total
}

func (counter *countingResolver) distinct() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	return len(counter.calls)
}

func (counter *countingResolver) maxPerPath() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	peak := 0
	for _, count := range counter.calls {
		if count > peak {
			peak = count
		}
	}
	return peak
}

// TestCanonicalCacheSharesResolutionsWithinOneRecheckEpoch proves one save
// with N targets walks each distinct path once (not once per validation
// pass), K further saves in the same epoch walk nothing new, and the
// per-write recheck resets the epoch so the next save re-walks.
func TestCanonicalCacheSharesResolutionsWithinOneRecheckEpoch(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
		fileTarget(t, "a", "two", filepath.Join(liveRoot, "two"), filepath.Join(stageRoot, "two"), "old-2", "new-2"),
		fileTarget(t, "b", "three", filepath.Join(liveRoot, "three"), filepath.Join(stageRoot, "three"), "old-3", "new-3"),
	}
	counter := newCountingResolver()
	engine := mustEngine(t, home)
	engine.canonicalResolver = counter.fn
	guard := &recordingGuard{}

	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-cache-epoch",
		ProjectIdentity: "/test/project",
		Targets:         targets,
		BoundaryCheck:   guard.check,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Prepare runs buildJournal validation plus several saves (staging
	// progress, prepared marker), each with two namespace passes. Every
	// distinct path must still have been walked exactly once.
	if got := counter.maxPerPath(); got != 1 {
		t.Fatalf("max walks per path after prepare = %d, want 1 (one walk per distinct path per epoch)", got)
	}
	preparedTotal := counter.total()
	preparedDistinct := counter.distinct()
	t.Logf("prepare walked %d distinct paths, %d total walks", preparedDistinct, preparedTotal)
	if preparedDistinct == 0 || preparedTotal != preparedDistinct {
		t.Fatalf("prepare walks = %d total over %d distinct paths, want equal (no repeated walk)", preparedTotal, preparedDistinct)
	}

	// K further saves in the same epoch walk nothing new.
	stored, err := engine.loadJournal(journal.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	for save := 0; save < 3; save++ {
		if err := engine.saveJournal(stored); err != nil {
			t.Fatal(err)
		}
	}
	if got := counter.total(); got != preparedTotal {
		t.Fatalf("walks after 3 same-epoch saves = %d, want %d (no new walk)", got, preparedTotal)
	}

	// The per-write recheck ends the epoch: the next save re-walks.
	if err := engine.checkBoundary(stored, 0); err != nil {
		t.Fatal(err)
	}
	if err := engine.saveJournal(stored); err != nil {
		t.Fatal(err)
	}
	if got := counter.total(); got <= preparedTotal {
		t.Fatalf("walks after recheck + save = %d, want more than %d (epoch reset re-walks)", got, preparedTotal)
	}
	if got := counter.maxPerPath(); got != 2 {
		t.Fatalf("max walks per path after one reset epoch = %d, want 2", got)
	}

	// Terminal cleanup releases the transaction's cache entry with its guard.
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
	engine.mu.Lock()
	_, cached := engine.nsCache[journal.TransactionID]
	engine.mu.Unlock()
	if cached {
		t.Fatal("canonical cache entry survived terminal journal removal")
	}
}

// TestCanonicalCacheSwappedSymlinkWithinEpochRefusesAtRecheck proves the
// invalidation: a symlink swapped between two saves in one epoch is served
// stale by the second save and still caught at the next per-write recheck,
// and the transaction refuses with the boundary diagnostic as today. The
// post-recheck save re-walks (dropping the invalidation fails this row).
//
// The swap replaces the live file with a link at the same path: the epoch
// cache still serves the pre-swap resolution, while the recheck re-walks
// and refuses the newly introduced final link.
func TestCanonicalCacheSwappedSymlinkWithinEpochRefusesAtRecheck(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)
	live := filepath.Join(liveRoot, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stageRoot, "one"), "old", "new")

	st := staging.Target{LivePath: live}
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, nil) }

	counter := newCountingResolver()
	engine := mustEngine(t, home)
	engine.canonicalResolver = counter.fn
	prepared, err := engine.Prepare(testLock{}, Plan{
		TransactionID: "txn-cache-swap", ProjectIdentity: "/test/project",
		Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable,
	})
	if err != nil {
		t.Fatal(err)
	}
	// loadJournal validates through the same epoch cache; count the epoch
	// from after the load so the swap window is exactly save-swap-save.
	stored, err := engine.loadJournal(prepared.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	beforeSwap := counter.total()

	// Swap the live file for a link at the same path. The sidecar parent
	// keeps no symlink ancestor (the engine's directory sync requires a
	// real directory), while the live spelling itself now resolves
	// elsewhere.
	elsewhere := filepath.Join(root, "elsewhere")
	mustWrite(t, elsewhere, "elsewhere")
	shelved := filepath.Join(root, "shelved-old")
	if err := os.Rename(live, shelved); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, live); err != nil {
		_ = os.Rename(shelved, live)
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := engine.saveJournal(stored); err != nil {
		t.Fatalf("same-epoch save after swap: %v", err)
	}
	if got := counter.total(); got != beforeSwap {
		t.Fatalf("same-epoch save after swap walked %d new paths, want 0 (stale epoch cache)", got-beforeSwap)
	}

	// The next per-write recheck re-walks and refuses the new final link.
	if err := engine.checkBoundary(stored, 0); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("recheck after swap err = %v, want source_output_overlap", err)
	}

	// The post-recheck save re-resolves (the epoch was reset).
	if err := engine.saveJournal(stored); err != nil {
		t.Fatal(err)
	}
	if got := counter.total(); got <= beforeSwap {
		t.Fatal("post-recheck save did not re-walk the filesystem (invalidation missing)")
	}

	// The transaction refuses as today. Commit verifies the preimage
	// before the recheck (commitTarget order), so it stops at the unknown
	// concurrent link with the digest diagnostic; the direct recheck above
	// proves the boundary gate catches the same swap with
	// source_output_overlap when it runs. Either way nothing is written.
	err = engine.Commit(testLock{}, prepared.TransactionID)
	if err == nil {
		t.Fatal("commit after swap succeeded, want refusal")
	}
	t.Logf("commit after swap refuses: %v", err)
	if info, err := os.Lstat(live); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("swapped live link was not preserved: %v %v", info, err)
	}
	if got := mustRead(t, shelved); got != "old" {
		t.Fatalf("shelved preimage = %q, want old", got)
	}
	if _, err := os.Lstat(engine.journalPath(prepared.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestCanonicalCacheIsolatedAcrossTransactions proves no entry crosses
// transactions: a second transaction over the same live paths walks every
// path again instead of hitting the first transaction's cache.
func TestCanonicalCacheIsolatedAcrossTransactions(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	build := func() []Target {
		return []Target{
			fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
			fileTarget(t, "a", "two", filepath.Join(liveRoot, "two"), filepath.Join(stageRoot, "two"), "old-2", "new-2"),
		}
	}
	counter := newCountingResolver()
	engine := mustEngine(t, home)
	engine.canonicalResolver = counter.fn
	guard := &recordingGuard{}

	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID: "txn-cache-first", ProjectIdentity: "/test/project",
		Targets: build(), BoundaryCheck: guard.check,
	}); err != nil {
		t.Fatal(err)
	}
	firstTotal := counter.total()
	if firstTotal == 0 {
		t.Fatal("first transaction walked nothing")
	}
	// Same live paths, new transaction id: every path must be walked
	// again. A cache keyed by path alone would serve the first
	// transaction's entries and this row would fail.
	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID: "txn-cache-second", ProjectIdentity: "/test/project",
		Targets: build(), BoundaryCheck: guard.check,
	}); err != nil {
		t.Fatal(err)
	}
	secondTotal := counter.total()
	if secondTotal < 2*firstTotal {
		t.Fatalf("walks after two same-path transactions = %d, want at least %d (no cross-transaction hit)", secondTotal, 2*firstTotal)
	}
	engine.mu.Lock()
	_, firstCached := engine.nsCache["txn-cache-first"]
	_, secondCached := engine.nsCache["txn-cache-second"]
	engine.mu.Unlock()
	if !firstCached || !secondCached {
		t.Fatal("per-transaction cache entries are not isolated by transaction id")
	}
}

// TestLegacyJournalBypassesCanonicalCache proves journals without a recheck
// (no guard, nil proof) never use the cache: every save walks every path,
// exactly as before the cache existed.
func TestLegacyJournalBypassesCanonicalCache(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
	}
	counter := newCountingResolver()
	engine := mustEngine(t, home)
	engine.canonicalResolver = counter.fn

	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID: "txn-cache-legacy", ProjectIdentity: "/test/project",
		Targets: targets,
	})
	if err != nil {
		t.Fatal(err)
	}
	afterPrepare := counter.total()
	stored, err := engine.loadJournal(journal.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	loadWalks := counter.total() - afterPrepare
	if loadWalks == 0 {
		t.Fatal("legacy loadJournal walked nothing (expected a full per-save walk)")
	}
	before := counter.total()
	if err := engine.saveJournal(stored); err != nil {
		t.Fatal(err)
	}
	after := counter.total()
	if after <= before {
		t.Fatal("legacy saveJournal served a cache entry (legacy must walk every save)")
	}
	engine.mu.Lock()
	_, cached := engine.nsCache[journal.TransactionID]
	engine.mu.Unlock()
	if cached {
		t.Fatal("legacy transaction populated the canonical cache")
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
}

// TestCachedTargetPathMirrorsFreeFunction proves the cached entry-path join
// matches canonicalNamespaceTargetPath on the same inputs, so the cache
// changes only when paths are walked, never what they resolve to.
func TestCachedTargetPathMirrorsFreeFunction(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	engine := mustEngine(t, home)
	engine.canonicalResolver = canonicalNamespacePath

	live := filepath.Join(root, "live", "one")
	mustWrite(t, live, "old")
	cases := []struct {
		path  string
		entry bool
	}{
		{path: live, entry: false},
		{path: live, entry: true},
		{path: filepath.Join(root, "live", "missing"), entry: false},
		{path: filepath.Join(root, "live", "missing"), entry: true},
	}
	for _, testCase := range cases {
		want, wantErr := canonicalNamespaceTargetPath(testCase.path, testCase.entry)
		got, gotErr := engine.cachedCanonicalTargetPath("txn-mirror", testCase.path, testCase.entry)
		if (wantErr == nil) != (gotErr == nil) || want != got {
			t.Fatalf("path %q entry=%v: cached = (%q, %v), want (%q, %v)",
				testCase.path, testCase.entry, got, gotErr, want, wantErr)
		}
		// Second call hits the cache and returns the identical answer.
		again, againErr := engine.cachedCanonicalTargetPath("txn-mirror", testCase.path, testCase.entry)
		if (wantErr == nil) != (againErr == nil) || want != again {
			t.Fatalf("cached repeat for %q entry=%v = (%q, %v), want (%q, %v)",
				testCase.path, testCase.entry, again, againErr, want, wantErr)
		}
	}
}

var _ = fmt.Sprintf
