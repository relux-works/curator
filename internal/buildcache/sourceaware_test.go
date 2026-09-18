package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
)

func sourceAwareTestInput(command string) buildmeta.Input {
	input := testInput(command)
	input.Package = &buildmeta.Package{Kind: buildmeta.PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("ab", 32)}
	return input
}

func entryDirs(t *testing.T, home string) (legacy, sourceAware []string) {
	t.Helper()
	list := func(name string) []string {
		entries, err := os.ReadDir(filepath.Join(home, "cache", "build", name))
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			t.Fatal(err)
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		return names
	}
	return list(Namespace), list(SourceAwareNamespace)
}

// TestSourceAwarePublicationUsesTheReceipt3Namespace is the positive row:
// a package-bearing input publishes under cache/build/go-v1-receipt-3 with a
// schema-3 receipt, is found again by the same expectation, and leaves the
// legacy namespace untouched.
func TestSourceAwarePublicationUsesTheReceipt3Namespace(t *testing.T) {
	store := newTestStore(t)
	input := sourceAwareTestInput("tool")
	publication, receiptHash := testPublication(t, store.Home(), input, []byte("source-aware executable"))
	result, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Published || !result.SourceAware || result.ReceiptHash != receiptHash {
		t.Fatalf("publication = %+v", result)
	}
	legacy, sourceAware := entryDirs(t, store.Home())
	if len(legacy) != 0 || len(sourceAware) != 1 || "sha256:"+sourceAware[0] != string(result.CacheKey) {
		t.Fatalf("namespaces: legacy=%v receipt-3=%v key=%s", legacy, sourceAware, result.CacheKey)
	}
	hit := store.Inspect(Expectation{Input: input, ReceiptHash: receiptHash, Assurance: publication.Assurance})
	if hit.Status != Hit || hit.Receipt.SchemaVersion != buildmeta.SourceAwareSchemaVersion || hit.Receipt.Input.Package == nil {
		t.Fatalf("inspection = %+v", hit)
	}
	if !strings.Contains(hit.ArtifactPath, string(filepath.Separator)+SourceAwareNamespace+string(filepath.Separator)) {
		t.Fatalf("artifact path outside the receipt-3 namespace: %s", hit.ArtifactPath)
	}
	again, err := store.Publish(publication, testHomeLock{})
	if err != nil || again.Status != ReusedWinner || !again.SourceAware {
		t.Fatalf("identical publication = %+v, %v", again, err)
	}
	// Revert finds the same slot through the recorded namespace.
	if err := store.Revert(result.CacheKey, result, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	if after := store.Inspect(Expectation{Input: input, Assurance: publication.Assurance}); after.Status == Hit {
		t.Fatalf("reverted receipt-3 publication still live: %+v", after)
	}
}

// TestNoLegacyHitSatisfiesASourceAwareLookup is the namespace negative row:
// the same driver input published without a package is a miss for the
// package-bearing expectation, and vice versa; an entry copied across the
// namespace boundary is refused rather than adopted.
func TestNoLegacyHitSatisfiesASourceAwareLookup(t *testing.T) {
	store := newTestStore(t)
	legacyInput := testInput("tool")
	legacyPublication, _ := testPublication(t, store.Home(), legacyInput, []byte("legacy executable"))
	if _, err := store.Publish(legacyPublication, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	sourceAwareInput := sourceAwareTestInput("tool")
	if miss := store.Inspect(Expectation{Input: sourceAwareInput, Assurance: legacyPublication.Assurance}); miss.Status != Miss {
		t.Fatalf("legacy entry answered a source-aware lookup: %+v", miss)
	}
	publication, _ := testPublication(t, store.Home(), sourceAwareInput, []byte("source-aware executable"))
	published, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	other := sourceAwareTestInput("tool")
	other.Package = &buildmeta.Package{Kind: buildmeta.PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("cd", 32)}
	if miss := store.Inspect(Expectation{Input: other, Assurance: publication.Assurance}); miss.Status != Miss {
		t.Fatalf("a different package identity reused the entry: %+v", miss)
	}
	if hit := store.Inspect(Expectation{Input: legacyInput, Assurance: legacyPublication.Assurance}); hit.Status != Hit || hit.Receipt.SchemaVersion != buildmeta.SchemaVersion {
		t.Fatalf("legacy lookup no longer exact: %+v", hit)
	}

	// Plant the receipt-3 winner under the legacy namespace at the legacy
	// key: the legacy expectation must not adopt a wrapper receipt.
	legacyKey, err := testAssuredCacheKey(legacyInput, legacyPublication.Assurance)
	if err != nil {
		t.Fatal(err)
	}
	sourceAwareEntry, _, err := store.pathsIn(published.CacheKey, true)
	if err != nil {
		t.Fatal(err)
	}
	plantedEntry, _, err := store.paths(legacyKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(plantedEntry); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(sourceAwareEntry, plantedEntry); err != nil {
		t.Fatal(err)
	}
	planted := store.Inspect(Expectation{Input: legacyInput, Assurance: legacyPublication.Assurance})
	if planted.Status == Hit {
		t.Fatalf("legacy lookup adopted a planted receipt-3 entry: %+v", planted)
	}
	// And the legacy winner planted under the receipt-3 namespace at the
	// source-aware key is refused for the source-aware expectation.
	legacyEntry, _, err := store.paths(legacyKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(legacyEntry); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(sourceAwareEntry, legacyEntry); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(sourceAwareEntry); err != nil {
		t.Fatal(err)
	}
	legacyWinner, _ := testPublication(t, store.Home(), legacyInput, []byte("legacy executable"))
	if _, err := store.Publish(legacyWinner, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	legacyEntry, _, err = store.paths(legacyKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyTree(legacyEntry, sourceAwareEntry); err != nil {
		t.Fatal(err)
	}
	if planted := store.Inspect(Expectation{Input: sourceAwareInput, Assurance: publication.Assurance}); planted.Status == Hit {
		t.Fatalf("source-aware lookup adopted a planted legacy entry: %+v", planted)
	}
}

func copyTree(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- test fixture copy
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, info.Mode().Perm())
	})
}

// TestSweepCollectsTheReceipt3Namespace: unreferenced receipt-3 entries are
// removed after grace and referenced ones retained, with the same key set
// that governs the legacy namespace.
func TestSweepCollectsTheReceipt3Namespace(t *testing.T) {
	store := newTestStore(t)
	binding := closureexec.PortableAssuranceBinding()
	keep, _ := testPublication(t, store.Home(), sourceAwareTestInput("keep"), []byte("kept"))
	kept, err := store.Publish(keep, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	drop, _ := testPublication(t, store.Home(), sourceAwareTestInput("drop"), []byte("dropped"))
	dropped, err := store.Publish(drop, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	legacy, _ := testPublication(t, store.Home(), testInput("legacy"), []byte("legacy"))
	legacyPublished, err := store.Publish(legacy, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.Sweep(SweepRequest{Referenced: []string{string(kept.CacheKey)}, Now: time.Now().Add(48 * time.Hour), Grace: time.Hour}, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	removed := strings.Join(result.Removed, "\n")
	if !strings.Contains(removed, strings.TrimPrefix(string(dropped.CacheKey), "sha256:")) ||
		!strings.Contains(removed, strings.TrimPrefix(string(legacyPublished.CacheKey), "sha256:")) ||
		strings.Contains(removed, strings.TrimPrefix(string(kept.CacheKey), "sha256:")) {
		t.Fatalf("sweep removed %v, warnings %v", result.Removed, result.Warnings)
	}
	if hit := store.Inspect(Expectation{Input: keep.Input, Assurance: binding}); hit.Status != Hit {
		t.Fatalf("referenced receipt-3 entry swept: %+v", hit)
	}
	if miss := store.Inspect(Expectation{Input: drop.Input, Assurance: binding}); miss.Status != Miss {
		t.Fatalf("unreferenced receipt-3 entry retained: %+v", miss)
	}
}
