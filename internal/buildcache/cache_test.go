package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

type testHomeLock struct{ err error }

func (lock testHomeLock) AssertHeld() error { return lock.err }

type pointerTestHomeLock struct{}

func (*pointerTestHomeLock) AssertHeld() error { return nil }

func TestInspectMissAndUnsupportedAreReadOnly(t *testing.T) {
	store := newTestStore(t)
	input := testInput("tool")
	before := treeFingerprint(t, store.Home())
	result := store.Inspect(Expectation{Input: input})
	if result.Status != Miss || result.DryRunOutcome() != "would-preflight-and-build" {
		t.Fatalf("miss = %+v", result)
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatalf("read-only miss changed manager home\nbefore: %s\nafter:  %s", before, after)
	}

	store.supported = func() bool { return false }
	result = store.Inspect(Expectation{Input: input})
	if result.Status != Unsupported || result.DryRunOutcome() != "unsupported" {
		t.Fatalf("unsupported = %+v", result)
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatalf("unsupported inspection changed manager home\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestPublishAndInspectExactProtectedHit(t *testing.T) {
	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("golden-tool"), []byte("verified executable"))
	result, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Published || result.ReceiptHash != receiptHash {
		t.Fatalf("publication = %+v", result)
	}

	hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
	if hit.Status != Hit || hit.DryRunOutcome() != "cache-hit" {
		t.Fatalf("inspection = %+v", hit)
	}
	if string(hit.ReceiptBytes) != string(publication.ReceiptBytes) {
		t.Fatal("receipt bytes changed during publication")
	}
	if got := readFile(t, hit.ArtifactPath); got != "verified executable" {
		t.Fatalf("artifact = %q", got)
	}

	again, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != ReusedWinner || again.ArtifactPath != hit.ArtifactPath {
		t.Fatalf("identical publication = %+v", again)
	}
}

func TestInspectRejectsCorruptReceiptAndArtifactState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, store *Store, publication Publication, hit Result)
		hash   func(buildmeta.ReceiptHash) buildmeta.ReceiptHash
	}{
		{
			name: "receipt trailing newline",
			mutate: func(t *testing.T, _ *Store, publication Publication, hit Result) {
				writeFile(t, filepath.Join(filepath.Dir(hit.ArtifactPath), "..", ReceiptFilename), append(publication.ReceiptBytes, '\n'), 0o600)
			},
		},
		{
			name: "artifact bytes",
			mutate: func(t *testing.T, _ *Store, _ Publication, hit Result) {
				writeFile(t, hit.ArtifactPath, []byte("tampered artifact"), 0o700)
			},
		},
		{
			name: "missing receipt",
			mutate: func(t *testing.T, _ *Store, _ Publication, hit Result) {
				if err := os.Remove(filepath.Join(filepath.Dir(hit.ArtifactPath), "..", ReceiptFilename)); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "oversized receipt",
			mutate: func(t *testing.T, _ *Store, _ Publication, hit Result) {
				receipt := filepath.Join(filepath.Dir(hit.ArtifactPath), "..", ReceiptFilename)
				if err := os.Truncate(receipt, maxReceiptBytes+1); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "unexpected entry",
			mutate: func(t *testing.T, _ *Store, _ Publication, hit Result) {
				writeFile(t, filepath.Join(filepath.Dir(hit.ArtifactPath), "unexpected"), []byte("x"), 0o600)
			},
		},
		{
			name:   "recorded receipt hash",
			mutate: func(*testing.T, *Store, Publication, Result) {},
			hash: func(buildmeta.ReceiptHash) buildmeta.ReceiptHash {
				return buildmeta.ReceiptHash("sha256:" + strings.Repeat("0", 64))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			if _, err := store.Publish(publication, testHomeLock{}); err != nil {
				t.Fatal(err)
			}
			hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
			if hit.Status != Hit {
				t.Fatalf("initial inspection = %+v", hit)
			}
			test.mutate(t, store, publication, hit)
			wantHash := receiptHash
			if test.hash != nil {
				wantHash = test.hash(receiptHash)
			}
			result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: wantHash})
			if result.Status != Corrupt || result.DryRunOutcome() != "corrupt" {
				t.Fatalf("corrupt inspection = %+v", result)
			}
		})
	}
}

func TestPublicationRequiresHeldHomeLock(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	before := treeFingerprint(t, store.Home())
	var nilPointerLock *pointerTestHomeLock
	for _, lock := range []HomeLock{nil, nilPointerLock, testHomeLock{err: errors.New("released")}} {
		if _, err := store.Publish(publication, lock); err == nil {
			t.Fatal("publication without a held manager-home lock succeeded")
		}
		if after := treeFingerprint(t, store.Home()); after != before {
			t.Fatalf("rejected publication changed manager home\nbefore: %s\nafter:  %s", before, after)
		}
	}
	key, err := publication.Input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Quarantine(key, nil); err == nil {
		t.Fatal("quarantine without a held manager-home lock succeeded")
	}
}

func TestPublishQuarantinesCorruptEntryBeforeReplacement(t *testing.T) {
	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	published, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, published.ArtifactPath, []byte("corrupt"), 0o700)

	replaced, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Status != Published {
		t.Fatalf("replacement = %+v", replaced)
	}
	if hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash}); hit.Status != Hit {
		t.Fatalf("replacement inspection = %+v", hit)
	}
	base := filepath.Dir(filepath.Dir(replaced.ArtifactPath))
	entries, err := os.ReadDir(filepath.Dir(base))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".quarantine-") {
			found = true
		}
	}
	if !found {
		t.Fatal("corrupt entry was replaced without quarantine")
	}
}

// TestRevertRestoresExactlyWhatAPublicationDisplaced proves publication is
// reversible against real protected directories.
//
// A caller that publishes for an installation it cannot yet make durable has to
// be able to put the cache back byte for byte, including the unusable entry it
// quarantined. Both directions are pinned: a replacement goes back to the
// corrupt predecessor, and a publication into a free slot leaves no entry at
// all behind.
func TestRevertRestoresExactlyWhatAPublicationDisplaced(t *testing.T) {
	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := publication.Input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}

	// A cold publication is withdrawn completely: nothing was displaced, so the
	// cache goes back to holding no entry for the key at all.
	cold := treeFingerprint(t, store.Home())
	published, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if published.Quarantined != "" {
		t.Fatalf("a publication into a free slot reported a quarantine %q", published.Quarantined)
	}
	if err := store.Revert(key, published, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	if live := store.Inspect(Expectation{Input: publication.Input}); live.Status != Miss {
		t.Fatalf("withdrawn publication = %+v, want a miss", live)
	}

	// A replacement is reversible too, and puts the exact corrupt bytes back.
	first, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, first.ArtifactPath, []byte("corrupt"), 0o700)
	corrupt := treeFingerprint(t, store.Home())
	if live := store.Inspect(Expectation{Input: publication.Input}); live.Status != Corrupt {
		t.Fatalf("tampered entry = %+v, want corrupt", live)
	}

	replacement, source := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	if source != receiptHash {
		t.Fatalf("the replacement receipt %s is not the original %s", source, receiptHash)
	}
	replaced, err := store.Publish(replacement, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Status != Published || replaced.Quarantined == "" {
		t.Fatalf("replacement = %+v, want a published winner that reports its quarantine", replaced)
	}
	if live := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash}); live.Status != Hit {
		t.Fatalf("replacement inspection = %+v", live)
	}

	if err := store.Revert(key, replaced, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	if live := store.Inspect(Expectation{Input: publication.Input}); live.Status != Corrupt {
		t.Fatalf("reverted entry = %+v, want the corrupt predecessor back", live)
	}
	if readFile(t, first.ArtifactPath) != "corrupt" {
		t.Fatal("the restored entry does not hold the bytes the publication displaced")
	}
	// Only the reversal's own quarantine of the withdrawn winner may differ:
	// no byte of the cold state or of the corrupt entry is deleted here.
	if cold == corrupt {
		t.Fatal("the fixture never actually changed the cache, so the case proves nothing")
	}
}

// TestRevertFailsClosed pins the boundaries of the reversal: it needs the same
// held home lock as the publication, it refuses to import bytes from outside
// the protected cache root, and a publication that selected no new winner has
// nothing to undo.
//
// Every refusal is reported as a changed cache even though it moves nothing.
// That is not a contradiction: a refused reversal leaves the entry this run
// published exactly where it is, so a caller that read the refusal as an
// ordinary failure would go on claiming the live cache is unchanged.
func TestRevertFailsClosed(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := publication.Input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	published, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}

	unsupported := &Store{home: t.TempDir(), supported: func() bool { return false }}
	for name, refused := range map[string]error{
		"without a held manager-home lock": store.Revert(key, published, nil),
		"without supported protection":     unsupported.Revert(key, published, testHomeLock{}),
		"with a malformed cache key": store.Revert(
			buildmeta.CacheKey("not-a-cache-key"), published, testHomeLock{}),
	} {
		if refused == nil {
			t.Fatalf("reversal %s succeeded", name)
		}
		if !StateChanged(refused) {
			t.Fatalf("reversal %s did not report the published entry as still live: %v", name, refused)
		}
	}
	if err := store.Revert(key, PublicationResult{Status: ReusedWinner}, testHomeLock{}); err != nil {
		t.Fatalf("reverting a reused winner failed: %v", err)
	}
	if live := store.Inspect(Expectation{Input: publication.Input}); live.Status != Hit {
		t.Fatalf("a reused-winner reversal changed the cache: %+v", live)
	}

	outside := filepath.Join(t.TempDir(), "planted")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	foreign := PublicationResult{Status: Published, Quarantined: outside}
	foreignErr := store.Revert(key, foreign, testHomeLock{})
	if foreignErr == nil {
		t.Fatal("reversal restored an entry from outside the protected cache root")
	}
	if !StateChanged(foreignErr) {
		t.Fatalf("a refused reversal did not report the published entry as still live: %v", foreignErr)
	}
	if _, err := os.Lstat(outside); err != nil {
		t.Fatalf("the refused reversal moved foreign bytes: %v", err)
	}
	// A refusal is not allowed to be half-applied: the entry it would have
	// replaced is still exactly where the publication left it.
	if live := store.Inspect(Expectation{Input: publication.Input}); live.Status != Hit {
		t.Fatalf("a refused reversal withdrew the live entry anyway: %+v", live)
	}
}

func TestAtomicPublicationIdenticalRace(t *testing.T) {
	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("same artifact"))
	const publishers = 12
	start := make(chan struct{})
	results := make(chan PublicationResult, publishers)
	errs := make(chan error, publishers)
	var wait sync.WaitGroup
	for index := 0; index < publishers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			result, err := store.Publish(publication, testHomeLock{})
			results <- result
			errs <- err
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	published := 0
	for result := range results {
		if result.Status == Published {
			published++
		} else if result.Status != ReusedWinner {
			t.Fatalf("race result = %+v", result)
		}
	}
	if published != 1 {
		t.Fatalf("published winners = %d, want 1", published)
	}
	if hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash}); hit.Status != Hit {
		t.Fatalf("race winner = %+v", hit)
	}
}

func TestAtomicPublicationConflictingRace(t *testing.T) {
	store := newTestStore(t)
	input := testInput("tool")
	first, _ := testPublication(t, store.Home(), input, []byte("first artifact"))
	second, _ := testPublication(t, store.Home(), input, []byte("second artifact"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, publication := range []Publication{first, second} {
		publication := publication
		go func() {
			<-start
			_, err := store.Publish(publication, testHomeLock{})
			errs <- err
		}()
	}
	close(start)
	var success, conflicts int
	for index := 0; index < 2; index++ {
		err := <-errs
		if err == nil {
			success++
			continue
		}
		var conflict *ConflictError
		if errors.As(err, &conflict) {
			conflicts++
			continue
		}
		t.Fatalf("unexpected publication error: %v", err)
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
}

func TestUnsupportedPublicationFailsWithoutPersistentState(t *testing.T) {
	store := newTestStore(t)
	store.supported = func() bool { return false }
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	before := treeFingerprint(t, store.Home())
	if _, err := store.Publish(publication, testHomeLock{}); err == nil {
		t.Fatal("unsupported platform published persistent cache state")
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatalf("unsupported publication changed manager home\nbefore: %s\nafter:  %s", before, after)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	if !protectionSupported() {
		t.Skip("persistent protection is unsupported on this platform")
	}
	home := t.TempDir()
	protectTestHome(t, home)
	store, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func testInput(command string) buildmeta.Input {
	tuning := map[string]string{}
	switch runtime.GOARCH {
	case "amd64":
		tuning["GOAMD64"] = "v1"
	case "arm64":
		tuning["GOARM64"] = "v8.0"
	case "386":
		tuning["GO386"] = "sse2"
	}
	return buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource: buildsource.Identity{
			Algorithm:     buildsource.Algorithm,
			ContentSHA256: "sha256:" + strings.Repeat("b", 64),
		},
		BuildRoot: "build",
		Command:   command,
		SourceDir: "build/cmd/" + command,
		Target: buildmeta.Target{
			GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Tuning: tuning,
		},
		Toolchain: buildmeta.Toolchain{
			Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
			GoVersion: "go version go1.26.1 " + runtime.GOOS + "/" + runtime.GOARCH, ContentSHA256: "sha256:" + strings.Repeat("c", 64),
		},
		Policy: buildmeta.FixedPolicy(),
	}
}

func testPublication(t *testing.T, root string, input buildmeta.Input, artifact []byte) (Publication, buildmeta.ReceiptHash) {
	t.Helper()
	artifactPath, err := buildmeta.ArtifactPath(input.Command, input.Target.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(artifact)
	receipt, err := buildmeta.NewReceipt(input, buildmeta.Artifact{
		Path: artifactPath, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(artifact)),
	})
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	receiptHash, err := buildmeta.HashReceiptBytes(receiptBytes)
	if err != nil {
		t.Fatal(err)
	}
	sourceDir := filepath.Join(root, "private-builds")
	if err := os.MkdirAll(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	source, err := os.CreateTemp(sourceDir, "artifact-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Write(artifact); err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	return Publication{Input: input, ReceiptBytes: receiptBytes, ArtifactSource: source.Name()}, receiptHash
}

func writeFile(t *testing.T, path string, payload []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, payload, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- test path
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func treeFingerprint(t *testing.T, root string) string {
	t.Helper()
	var records []string
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
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
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(records)
	return strings.Join(records, "|")
}
