package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
)

// failAt hooks the named mutation boundaries. Every listed point fails, every
// other one runs for real, so a case exercises the production path right up to
// the exact step it is about.
func failAt(points ...faultPoint) func(faultPoint) error {
	selected := map[faultPoint]bool{}
	for _, point := range points {
		selected[point] = true
	}
	return func(point faultPoint) error {
		if selected[point] {
			return fmt.Errorf("injected %s failure", point)
		}
		return nil
	}
}

// entryFingerprint digests the live protected entry of one key exactly as it
// is: every member, its mode, and its bytes. A slot with no entry is a distinct
// value rather than an error, because "nothing is live" is one of the states a
// compensation has to be able to restore.
func entryFingerprint(t *testing.T, store *Store, key buildmeta.CacheKey) string {
	t.Helper()
	entryPath, _, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(entryPath); err != nil {
		if os.IsNotExist(err) {
			return "absent"
		}
		t.Fatal(err)
	}
	var records []string
	walkErr := filepath.WalkDir(entryPath, func(path string, _ os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(entryPath, path)
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		record := fmt.Sprintf("%s:%s:%d", filepath.ToSlash(rel), info.Mode().String(), info.Size())
		if info.Mode().IsRegular() {
			payload, err := os.ReadFile(path) // #nosec G304 -- test fixture tree
			if err != nil {
				return err
			}
			digest := sha256.Sum256(payload)
			record += ":" + hex.EncodeToString(digest[:])
		}
		records = append(records, record)
		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	sort.Strings(records)
	return strings.Join(records, "\n")
}

// cacheRootMembers counts the cache root by class: the live entries, what was
// moved aside rather than deleted, and any private staging left behind.
//
// Every compensation in this package is a rename, so quarantined only ever
// grows — that is how a case proves no byte of a displaced entry was destroyed
// on the way back — while staged must always return to zero, because a
// publication that failed has no business leaving a half-built entry in a
// shared root.
func cacheRootMembers(t *testing.T, store *Store, key buildmeta.CacheKey) (live, quarantined, staged int) {
	t.Helper()
	_, base, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		// A cold cache has no root yet, which is the same thing as an empty one.
		return 0, 0, 0
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		switch {
		case strings.HasPrefix(entry.Name(), ".quarantine-"):
			quarantined++
		case strings.HasPrefix(entry.Name(), ".stage-"):
			staged++
		default:
			live++
		}
	}
	return live, quarantined, staged
}

// corruptPredecessor publishes one entry and then breaks it, producing the
// state a repair actually finds: a live entry the manager refuses to reuse and
// therefore quarantines before it selects a replacement.
func corruptPredecessor(t *testing.T, store *Store, publication Publication) {
	t.Helper()
	first, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, first.ArtifactPath, []byte("corrupt predecessor bytes"), 0o700)
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Corrupt {
		t.Fatalf("the fixture predecessor = %+v, want corrupt", live)
	}
}

// TestAFailedPublicationRestoresTheCacheItDisplaced pins the guarantee the
// caller depends on: an error from Publish means the protected cache root is
// exactly what Publish found.
//
// Selecting a winner is three live moves, not one. The unusable predecessor is
// quarantined, the staged directory is renamed into the freed slot, and only
// then is the winner validated and the cache root synced. A fault at any of
// those steps used to return an empty result with an error while the shared
// slot already held something else — and an empty result gives the caller
// nothing to compensate with, so a failed install would leave the predecessor
// quarantined and a replacement live for an operation that committed nothing.
//
// Each case faults one boundary on the production path and asserts the live
// entry is byte-identical to what it was, that the error does not claim a
// changed cache, and that nothing was deleted to get there.
func TestAFailedPublicationRestoresTheCacheItDisplaced(t *testing.T) {
	for name, testCase := range map[string]struct {
		// displaced seeds a corrupt live entry the publication has to move aside
		// first. Without it the slot is free and the quarantine step never runs.
		displaced bool
		at        faultPoint
		wantLive  Status
	}{
		"quarantining the predecessor fails": {displaced: true, at: faultQuarantine, wantLive: Corrupt},
		// A selection that never succeeds runs the retry loop out after the
		// predecessor was already moved aside, which is its own exit.
		"selecting the replacement never succeeds":  {displaced: true, at: faultSelect, wantLive: Corrupt},
		"validating the replacement fails":          {displaced: true, at: faultValidate, wantLive: Corrupt},
		"syncing the replacement fails":             {displaced: true, at: faultSync, wantLive: Corrupt},
		"selecting into a free slot never succeeds": {at: faultSelect, wantLive: Miss},
		"validating a cold publication fails":       {at: faultValidate, wantLive: Miss},
		"syncing a cold publication fails":          {at: faultSync, wantLive: Miss},
		"an unfaulted publication is unchanged":     {displaced: true, at: "no-such-boundary", wantLive: Hit},
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestStore(t)
			publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
			if err != nil {
				t.Fatal(err)
			}
			if testCase.displaced {
				corruptPredecessor(t, store, publication)
			}
			before := entryFingerprint(t, store, key)
			liveBefore, quarantinedBefore, _ := cacheRootMembers(t, store, key)

			replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			store.faults = failAt(testCase.at)
			result, publishErr := store.Publish(replacement, testHomeLock{})
			store.faults = nil

			// The control case proves the fixture really does publish when nothing
			// is faulted, so a passing failure case cannot be a publication that
			// never happened.
			if testCase.wantLive == Hit {
				if publishErr != nil {
					t.Fatalf("an unfaulted publication failed: %v", publishErr)
				}
				if result.Status != Published || result.Quarantined == "" {
					t.Fatalf("publication = %+v, want a published winner reporting its quarantine", result)
				}
				if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Hit {
					t.Fatalf("live entry = %+v, want a hit", live)
				}
				return
			}

			if publishErr == nil {
				t.Fatalf("a publication faulted at %s succeeded: %+v", testCase.at, result)
			}
			if StateChanged(publishErr) {
				t.Fatalf("a publication that put the cache back reported a changed cache: %v", publishErr)
			}
			if result != (PublicationResult{}) {
				t.Fatalf("a failed publication returned a usable result: %+v", result)
			}
			if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != testCase.wantLive {
				t.Fatalf("live verdict = %s, want the prior %s", live.Status, testCase.wantLive)
			}
			if after := entryFingerprint(t, store, key); after != before {
				t.Fatalf("a failed publication changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
			}
			live, quarantined, staged := cacheRootMembers(t, store, key)
			switch {
			case live != liveBefore:
				t.Fatalf("live cache entries = %d, want the prior %d", live, liveBefore)
			case quarantined < quarantinedBefore:
				t.Fatal("a compensation deleted quarantined bytes instead of leaving them for the sweep")
			case staged != 0:
				t.Fatalf("a failed publication left %d private staging directories in the shared root", staged)
			}
		})
	}
}

// TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError covers
// the interior failure that the ordinary quarantine fault case cannot reach.
// quarantinePath renames the live entry and then syncs its parent; if that sync
// fails, it returns an error without the path it already moved. From Publish's
// point of view that is indistinguishable from this deterministic seam: the
// production quarantine helper has moved the predecessor, but the boundary
// returns an error before Publish records displaced.
//
// Publish documents that an ordinary error means the cache is exactly as the
// call found it. The test therefore requires the corrupt predecessor to remain
// live and byte-identical. A StateChangedError with a complete recovery record
// would be the alternative contract, but the current result is empty and the
// current error does not carry that record.
func TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	corruptPredecessor(t, store, publication)
	before := entryFingerprint(t, store, key)
	entryPath, _, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}

	store.faults = func(point faultPoint) error {
		if point != faultQuarantine {
			return nil
		}
		moved, moveErr := store.quarantinePath(entryPath, testHomeLock{})
		if moveErr != nil {
			return moveErr
		}
		if moved == "" {
			return fmt.Errorf("the post-mutation quarantine fixture moved nothing")
		}
		return fmt.Errorf("injected sync quarantined cache root failure")
	}
	replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	result, publishErr := store.Publish(replacement, testHomeLock{})
	store.faults = nil

	if publishErr == nil {
		t.Fatalf("a quarantine that failed after moving the predecessor succeeded: %+v", result)
	}
	if StateChanged(publishErr) {
		t.Fatalf("Publish returned a changed-state error without a recovery record: %v", publishErr)
	}
	if result != (PublicationResult{}) {
		t.Fatalf("a failed publication returned a usable result: %+v", result)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Corrupt {
		t.Fatalf("live verdict = %s, want the corrupt predecessor restored", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a quarantine error changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack covers the production
// helper's own interior window: quarantinePath renames the live entry aside and
// then syncs the cache root, and only that sync can fail with the slot already
// empty.
//
// This is the boundary the two seam-driven cases above model, exercised here on
// the real helper. Every caller of it reads the same guarantee — an error with
// no reported path means the slot is untouched — so the case asserts both halves
// of that claim rather than only the surviving bytes.
func TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Publish(publication, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	before := entryFingerprint(t, store, key)

	store.faults = failAt(faultQuarantineSync)
	moved, quarantineErr := store.Quarantine(key, testHomeLock{})
	store.faults = nil

	if quarantineErr == nil {
		t.Fatalf("a quarantine whose sync failed succeeded: %q", moved)
	}
	if moved != "" {
		t.Fatalf("a quarantine that put the entry back reported it as moved to %q", moved)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Hit {
		t.Fatalf("live verdict = %s, want the entry put back", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a failed quarantine changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState covers the
// second durability boundary in a failed quarantine. Returning the entry to
// the live pathname repairs the visible namespace, but that compensating rename
// must itself be synced before Publish can promise that the cache is unchanged.
//
// If the compensating sync also fails, the predecessor bytes remain recoverable
// and live, while the error must truthfully report uncertain/changed durable
// state so install can set BuildCacheRetained instead of claiming a clean
// rollback.
func TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	corruptPredecessor(t, store, publication)
	before := entryFingerprint(t, store, key)

	initialSyncFaulted := false
	rollbackSyncFaulted := false
	store.faults = func(point faultPoint) error {
		switch point {
		case faultQuarantineSync:
			initialSyncFaulted = true
			return fmt.Errorf("injected quarantine sync failure")
		case faultPoint("quarantine-rollback-sync"):
			rollbackSyncFaulted = true
			return fmt.Errorf("injected quarantine rollback sync failure")
		default:
			return nil
		}
	}
	replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	result, publishErr := store.Publish(replacement, testHomeLock{})
	store.faults = nil

	if publishErr == nil {
		t.Fatalf("a quarantine with two failed durability boundaries succeeded: %+v", result)
	}
	if !initialSyncFaulted {
		t.Fatal("the initial post-rename sync fault never ran")
	}
	if !rollbackSyncFaulted {
		t.Fatal("the compensating rename was not followed by a faultable durability sync")
	}
	if !StateChanged(publishErr) {
		t.Fatalf("a failed compensating sync did not report changed durable state: %v", publishErr)
	}
	if result != (PublicationResult{}) {
		t.Fatalf("a failed publication returned a usable result: %+v", result)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Corrupt {
		t.Fatalf("live verdict = %s, want the recoverable corrupt predecessor", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a failed compensating sync changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
	live, _, staged := cacheRootMembers(t, store, key)
	if live != 1 {
		t.Fatalf("live cache entries = %d, want one recoverable predecessor", live)
	}
	if staged != 0 {
		t.Fatalf("a failed publication left %d private staging directories in the shared root", staged)
	}
}

// TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller runs the same
// interior fault underneath a publication and a reversal, on the production
// path, with no seam standing in for the rename.
//
// The two paths answer it differently on purpose. A publication that never
// records a displaced predecessor has nothing to unwind and leaves the entry
// the helper put back; a reversal reports a changed cache whatever happens,
// because the entry this run published is still live and a caller reading an
// ordinary failure would go on claiming otherwise.
func TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller(t *testing.T) {
	for name, testCase := range map[string]struct {
		// revert runs the fault under Revert instead of under Publish.
		revert       bool
		wantLive     Status
		wantChanged  bool
		wantArtifact string
	}{
		"a publication cannot quarantine its predecessor durably": {
			wantLive: Corrupt, wantArtifact: "corrupt predecessor bytes",
		},
		"a reversal cannot withdraw its winner durably": {
			revert: true, wantLive: Hit, wantChanged: true, wantArtifact: "artifact",
		},
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestStore(t)
			publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
			if err != nil {
				t.Fatal(err)
			}
			corruptPredecessor(t, store, publication)

			replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			var published PublicationResult
			if testCase.revert {
				if published, err = store.Publish(replacement, testHomeLock{}); err != nil {
					t.Fatal(err)
				}
				if published.Status != Published || published.Quarantined == "" {
					t.Fatalf("publication = %+v, want a published winner reporting its quarantine", published)
				}
			}
			before := entryFingerprint(t, store, key)

			store.faults = failAt(faultQuarantineSync)
			var faultErr error
			if testCase.revert {
				faultErr = store.Revert(key, published, testHomeLock{})
			} else {
				_, faultErr = store.Publish(replacement, testHomeLock{})
			}
			store.faults = nil

			if faultErr == nil {
				t.Fatal("an operation whose quarantine could not be made durable succeeded")
			}
			if StateChanged(faultErr) != testCase.wantChanged {
				t.Fatalf("changed-cache report = %t, want %t: %v",
					StateChanged(faultErr), testCase.wantChanged, faultErr)
			}
			if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != testCase.wantLive {
				t.Fatalf("live verdict = %s, want %s", live.Status, testCase.wantLive)
			}
			if after := entryFingerprint(t, store, key); after != before {
				t.Fatalf("a failed quarantine changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
			}
			entryPath, _, err := store.paths(key)
			if err != nil {
				t.Fatal(err)
			}
			relative, err := buildmeta.ArtifactPath(publication.Input.Command, publication.Input.Target.GOOS)
			if err != nil {
				t.Fatal(err)
			}
			if got := readFile(t, filepath.Join(entryPath, filepath.FromSlash(relative))); got != testCase.wantArtifact {
				t.Fatalf("live artifact = %q, want %q", got, testCase.wantArtifact)
			}
			if _, _, staged := cacheRootMembers(t, store, key); staged != 0 {
				t.Fatalf("a failed operation left %d private staging directories in the shared root", staged)
			}
		})
	}
}

// TestAQuarantineThatCannotPutTheEntryBackHandsItsCallerTheRecord covers the
// last exit of the helper: the sync failed, the entry could not be returned to
// the live slot either, and the only thing that keeps the move recoverable is
// the quarantine path it reports alongside the error.
//
// The fixture takes the slot with an unremovable directory so the return rename
// really fails, then Publish has to use the reported path to put the predecessor
// back — which is the whole reason the path is returned with an error at all.
func TestAQuarantineThatCannotPutTheEntryBackHandsItsCallerTheRecord(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	corruptPredecessor(t, store, publication)
	before := entryFingerprint(t, store, key)
	entryPath, _, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}

	// Fire once: the compensation quarantines this squatter and restores the
	// predecessor over it, and a second firing would only re-take the slot.
	fired := false
	store.faults = func(point faultPoint) error {
		if point != faultQuarantineSync || fired {
			return nil
		}
		fired = true
		if err := os.MkdirAll(filepath.Join(entryPath, "bin"), 0o700); err != nil {
			return err
		}
		return fmt.Errorf("injected quarantine sync failure")
	}
	replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	result, publishErr := store.Publish(replacement, testHomeLock{})
	store.faults = nil

	if publishErr == nil {
		t.Fatalf("a publication whose quarantine could not be undone succeeded: %+v", result)
	}
	if !fired {
		t.Fatal("the interior quarantine fault never ran, so the case proves nothing")
	}
	if StateChanged(publishErr) {
		t.Fatalf("a publication that put the cache back reported a changed cache: %v", publishErr)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Corrupt {
		t.Fatalf("live verdict = %s, want the corrupt predecessor restored", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a failed publication changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestAPublicationThatCannotRestoreReportsAChangedCache covers the one case the
// compensation cannot fix: a fault inside the compensation itself.
//
// Silence here would be the worst outcome of all, because the caller would read
// an ordinary failure and go on telling the operator the live build cache is
// unchanged. Each case asserts the error says otherwise, and asserts exactly
// which entry is live — a usable one in every case, never an empty slot a live
// launcher would resolve to nothing.
func TestAPublicationThatCannotRestoreReportsAChangedCache(t *testing.T) {
	for name, testCase := range map[string]struct {
		at []faultPoint
		// wantLive is the verdict left behind. Hit means the replacement this
		// publication selected is still what the slot holds; Corrupt means the
		// predecessor was restored but the restoration is not durable.
		wantLive Status
	}{
		"the withdrawal of the replacement fails": {
			at: []faultPoint{faultValidate, faultWithdraw}, wantLive: Hit,
		},
		"the predecessor cannot be put back": {
			at: []faultPoint{faultValidate, faultRestore}, wantLive: Hit,
		},
		"the restoration cannot be synced": {
			at: []faultPoint{faultValidate, faultRestoreSync}, wantLive: Corrupt,
		},
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestStore(t)
			publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
			if err != nil {
				t.Fatal(err)
			}
			corruptPredecessor(t, store, publication)

			replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			store.faults = failAt(testCase.at...)
			_, publishErr := store.Publish(replacement, testHomeLock{})
			store.faults = nil

			if publishErr == nil {
				t.Fatal("a publication faulted through its own compensation succeeded")
			}
			if !StateChanged(publishErr) {
				t.Fatalf("a publication that could not restore did not say so: %v", publishErr)
			}
			if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != testCase.wantLive {
				t.Fatalf("live verdict = %s, want %s", live.Status, testCase.wantLive)
			}
			// A slot a launcher already points at must never be left empty, so the
			// compensation returns the entry it withdrew when it cannot do better.
			if entryFingerprint(t, store, key) == "absent" {
				t.Fatal("a failed compensation left the live slot empty")
			}
		})
	}
}

// TestAFailedReversalIsFailClosedAndReportsAChangedCache is the mirror image
// for the caller-driven half.
//
// Reversal withdraws the published winner and renames the predecessor back, and
// a fault between those two renames would otherwise leave the slot empty while
// a perfectly usable entry sat in quarantine — strictly worse than either end
// state. Every failing path reports a changed cache, because every one of them
// leaves the cache holding something other than what the run found.
func TestAFailedReversalIsFailClosedAndReportsAChangedCache(t *testing.T) {
	for name, testCase := range map[string]struct {
		at faultPoint
		// wantLive is the verdict left behind, and wantBytes the artifact bytes
		// that go with it.
		wantLive  Status
		wantBytes string
	}{
		"the withdrawal fails": {at: faultWithdraw, wantLive: Hit, wantBytes: "artifact"},
		"the predecessor cannot be put back": {
			at: faultRestore, wantLive: Hit, wantBytes: "artifact",
		},
		"the restoration cannot be synced": {
			at: faultRestoreSync, wantLive: Corrupt, wantBytes: "corrupt predecessor bytes",
		},
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestStore(t)
			publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
			if err != nil {
				t.Fatal(err)
			}
			corruptPredecessor(t, store, publication)

			replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			replaced, err := store.Publish(replacement, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			if replaced.Status != Published || replaced.Quarantined == "" {
				t.Fatalf("publication = %+v, want a published winner reporting its quarantine", replaced)
			}

			store.faults = failAt(testCase.at)
			revertErr := store.Revert(key, replaced, testHomeLock{})
			store.faults = nil
			if revertErr == nil {
				t.Fatalf("a reversal faulted at %s succeeded", testCase.at)
			}
			if !StateChanged(revertErr) {
				t.Fatalf("a reversal that did not complete did not report a changed cache: %v", revertErr)
			}

			live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance})
			if live.Status != testCase.wantLive {
				t.Fatalf("live verdict = %s, want %s", live.Status, testCase.wantLive)
			}
			entryPath, _, err := store.paths(key)
			if err != nil {
				t.Fatal(err)
			}
			relative, err := buildmeta.ArtifactPath(publication.Input.Command, publication.Input.Target.GOOS)
			if err != nil {
				t.Fatal(err)
			}
			artifact := filepath.Join(entryPath, filepath.FromSlash(relative))
			if got := readFile(t, artifact); got != testCase.wantBytes {
				t.Fatalf("live artifact = %q, want %q", got, testCase.wantBytes)
			}
		})
	}
}

// TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError pins the mirror
// interior failure. restoreDisplaced asks quarantinePath to withdraw the
// published winner. If the rename succeeds and the following parent sync
// fails, quarantinePath returns no moved path, so the reversal must still put
// the winner back or return a complete recovery record. Leaving the slot empty
// is not fail-closed for launchers that already resolve through it.
func TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	corruptPredecessor(t, store, publication)
	replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	replaced, err := store.Publish(replacement, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Status != Published || replaced.Quarantined == "" {
		t.Fatalf("publication = %+v, want a published winner reporting its quarantine", replaced)
	}
	before := entryFingerprint(t, store, key)
	entryPath, _, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}

	store.faults = func(point faultPoint) error {
		if point != faultWithdraw {
			return nil
		}
		moved, moveErr := store.quarantinePath(entryPath, testHomeLock{})
		if moveErr != nil {
			return moveErr
		}
		if moved == "" {
			return fmt.Errorf("the post-mutation withdrawal fixture moved nothing")
		}
		return fmt.Errorf("injected sync quarantined cache root failure")
	}
	revertErr := store.Revert(key, replaced, testHomeLock{})
	store.faults = nil

	if revertErr == nil {
		t.Fatal("a withdrawal that failed after moving the winner succeeded")
	}
	if !StateChanged(revertErr) {
		t.Fatalf("a reversal that did not restore the winner did not report changed state: %v", revertErr)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Hit {
		t.Fatalf("live verdict = %s, want the published winner returned", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a withdrawal error changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestAQuarantineReportsChangedStateOnlyWhenItsRollbackIsNotDurable pins the
// discrimination the typed failure is worth anything for.
//
// A quarantine that puts the entry back and syncs it has compensated its own
// mutation completely, and an untyped error is the guarantee the caller reads
// off that. Typing every failed quarantine as changed state would satisfy the
// changed-state case above while telling the install layer to retain the cache
// after a rollback that fully succeeded — so the two outcomes are asserted
// against each other, on the same production path, in the same shape.
func TestAQuarantineReportsChangedStateOnlyWhenItsRollbackIsNotDurable(t *testing.T) {
	for name, testCase := range map[string]struct {
		faults      []faultPoint
		wantChanged bool
	}{
		"a durable rollback owes the caller nothing": {
			faults:      []faultPoint{faultQuarantineSync},
			wantChanged: false,
		},
		"a rollback that is not durable reports changed state": {
			faults:      []faultPoint{faultQuarantineSync, faultQuarantineRollbackSync},
			wantChanged: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestStore(t)
			publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.Publish(publication, testHomeLock{}); err != nil {
				t.Fatal(err)
			}
			before := entryFingerprint(t, store, key)

			store.faults = failAt(testCase.faults...)
			moved, quarantineErr := store.Quarantine(key, testHomeLock{})
			store.faults = nil

			if quarantineErr == nil {
				t.Fatalf("a quarantine whose sync failed succeeded: %q", moved)
			}
			// The path stays empty in both outcomes: the caller owns nothing
			// either way, because the entry is back in the live slot.
			if moved != "" {
				t.Fatalf("a quarantine that put the entry back reported it as moved to %q", moved)
			}
			if StateChanged(quarantineErr) != testCase.wantChanged {
				t.Fatalf("StateChanged = %v, want %v: %v",
					StateChanged(quarantineErr), testCase.wantChanged, quarantineErr)
			}
			if testCase.wantChanged {
				var changed *StateChangedError
				if !errors.As(quarantineErr, &changed) {
					t.Fatal("a changed-state quarantine did not carry a *StateChangedError")
				}
				if changed.Key != key {
					t.Fatalf("changed state reported key %s, want %s", changed.Key, key)
				}
			}
			// The bytes are live and unchanged on both paths. Only the promise
			// about them differs, which is the whole point of the typing.
			if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Hit {
				t.Fatalf("live verdict = %s, want the entry put back", live.Status)
			}
			if after := entryFingerprint(t, store, key); after != before {
				t.Fatalf("a failed quarantine changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
			}
			live, quarantined, staged := cacheRootMembers(t, store, key)
			if live != 1 || quarantined != 0 {
				t.Fatalf("cache root = %d live, %d quarantined; want the entry back in the live slot",
					live, quarantined)
			}
			if staged != 0 {
				t.Fatalf("a failed quarantine left %d private staging directories in the shared root", staged)
			}
		})
	}
}

// TestAPublicationWhoseQuarantineRolledBackDurablyDoesNotRetainTheCache is the
// same discrimination one layer up, where the consequence lives.
//
// The install layer sets BuildCacheRetained straight from StateChanged, so a
// publication that failed on a quarantine which then rolled itself back durably
// has to report an ordinary error. The changed-state half of this pairing is
// TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState.
func TestAPublicationWhoseQuarantineRolledBackDurablyDoesNotRetainTheCache(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := testAssuredCacheKey(publication.Input, publication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	corruptPredecessor(t, store, publication)
	before := entryFingerprint(t, store, key)

	store.faults = failAt(faultQuarantineSync)
	replacement, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	result, publishErr := store.Publish(replacement, testHomeLock{})
	store.faults = nil

	if publishErr == nil {
		t.Fatalf("a publication whose quarantine sync failed succeeded: %+v", result)
	}
	if StateChanged(publishErr) {
		t.Fatalf("a durably rolled-back quarantine reported changed state: %v", publishErr)
	}
	if result != (PublicationResult{}) {
		t.Fatalf("a failed publication returned a usable result: %+v", result)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, Assurance: publication.Assurance}); live.Status != Corrupt {
		t.Fatalf("live verdict = %s, want the recoverable corrupt predecessor", live.Status)
	}
	if after := entryFingerprint(t, store, key); after != before {
		t.Fatalf("a failed publication changed the live entry\nbefore:\n%s\nafter:\n%s", before, after)
	}
	live, _, staged := cacheRootMembers(t, store, key)
	if live != 1 {
		t.Fatalf("live cache entries = %d, want one recoverable predecessor", live)
	}
	if staged != 0 {
		t.Fatalf("a failed publication left %d private staging directories in the shared root", staged)
	}
}
