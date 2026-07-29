package scopes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/marker"
)

func openProtectedStore(t *testing.T, home string) *buildcache.Store {
	t.Helper()
	store, err := buildcache.New(home)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

// authoritativeGCCase is the published maintenance contract. Only the field
// names are mirrored; the retained roots are read from the authoritative
// document at run time.
type authoritativeGCCase struct {
	Name  string   `json:"name"`
	Roots []string `json:"roots"`
}

func authoritativeGCCases(t *testing.T) []authoritativeGCCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "external-repository-lifecycle.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Cases []authoritativeGCCase `json:"status_repair_gc_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	// Only the retention case belongs to this package. The remaining published
	// cases are external-repository status and repair codes, which this manager
	// does not implement: it accepts skill schemas 1 through 6 and publishes no
	// build_repository_* code at all. They are routed, not bound here.
	var retention []authoritativeGCCase
	for _, published := range document.Cases {
		if len(published.Roots) > 0 {
			retention = append(retention, published)
		}
	}
	if len(retention) == 0 {
		t.Fatal("the authoritative suite publishes no maintenance retention case to bind")
	}
	return retention
}

// TestAuthoritativeGarbageCollectionRootsAreRetained binds every published
// retention root to a real maintenance pass over a real manager home. Each
// root gets its own executable proof, and an unpublished root fails rather
// than passing unasserted.
func TestAuthoritativeGarbageCollectionRootsAreRetained(t *testing.T) {
	for _, published := range authoritativeGCCases(t) {
		published := published
		t.Run(published.Name, func(t *testing.T) {
			for _, root := range published.Roots {
				root := root
				t.Run(root, func(t *testing.T) { assertRootRetained(t, root) })
			}
		})
	}
}

func assertRootRetained(t *testing.T, root string) {
	t.Helper()
	switch root {
	case "artifact-receipts":
		assertReceiptRootRetained(t)
	case "in-flight-journals":
		assertJournalRootRetained(t)
	case "install-markers":
		assertMarkerRootRetained(t)
	case "protected-snapshots":
		assertSnapshotRootRetained(t)
	case "uncertain-entries":
		assertUncertainRootRetained(t)
	default:
		t.Fatalf("published retention root %q has no executable binding", root)
	}
}

// assertReceiptRootRetained proves a published receipt an install marker still
// names survives a pass that sweeps everything it can prove unreferenced.
func assertReceiptRootRetained(t *testing.T) {
	t.Helper()
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	referenced := publishRealEntry(t, store, "kept-tool", "kept artifact")
	orphan := publishRealEntry(t, store, "orphan-tool", "orphan artifact")

	project := t.TempDir()
	installBuildMarker(t, filepath.Join(project, ".agents", "skills"), "skill-p",
		strings.Repeat("a", 40), "kept-tool", referenced)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	backdateEntry(t, home, referenced, 30*24*time.Hour)
	backdateEntry(t, home, orphan, 30*24*time.Hour)

	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("a provable pass warned: %v", result.Warnings)
	}
	entry := realEntryPath(home, referenced)
	if _, err := os.Lstat(entry); err != nil {
		t.Fatalf("a referenced receipt root was swept: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(entry, buildcache.ReceiptFilename)); err != nil {
		t.Fatalf("the retained entry no longer holds its receipt: %v", err)
	}
	// The pass really did sweep, so retention is a decision rather than inaction.
	if _, err := os.Lstat(realEntryPath(home, orphan)); err == nil {
		t.Fatal("the pass swept nothing, so retention proves nothing")
	}
}

// assertJournalRootRetained proves an artifact owned only by an in-flight
// transaction journal survives, with no marker naming it.
func assertJournalRootRetained(t *testing.T) {
	t.Helper()
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	inFlight := publishRealEntry(t, store, "in-flight-tool", "in-flight artifact")
	orphan := publishRealEntry(t, store, "orphan-tool", "orphan artifact")
	backdateEntry(t, home, inFlight, 30*24*time.Hour)
	backdateEntry(t, home, orphan, 30*24*time.Hour)

	result, err := Collect(MaintenanceRequest{
		Home: home, Lock: testHomeLock{}, JournalKeys: []string{string(inFlight)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if containsString(result.RemovedBuilds, string(inFlight)) {
		t.Fatalf("a journal-owned root was swept: %v", result.RemovedBuilds)
	}
	if _, err := os.Lstat(realEntryPath(home, inFlight)); err != nil {
		t.Fatalf("a journal-owned root was removed: %v", err)
	}
	if !containsString(result.RemovedBuilds, string(orphan)) {
		t.Fatalf("the pass swept nothing, so journal retention proves nothing: %v", result.RemovedBuilds)
	}
}

// assertMarkerRootRetained proves the runtime an install marker names survives
// the same pass that prunes runtime an unreferenced commit left behind.
func assertMarkerRootRetained(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	project := t.TempDir()
	referenced := strings.Repeat("1", 40)
	unreferenced := strings.Repeat("2", 40)
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-a", referenced)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	for _, commit := range []string{referenced, unreferenced} {
		if err := os.MkdirAll(filepath.Join(home, "runtime", "skill-a", commit), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", referenced)); err != nil {
		t.Fatalf("the runtime an install marker names was swept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", unreferenced)); err == nil {
		t.Fatalf("the pass swept nothing, so marker retention proves nothing: %v", result.RemovedRuntime)
	}
	if !containsString(LoadConsumers(home), project) {
		t.Fatalf("the consumer holding the marker was pruned: %v", LoadConsumers(home))
	}
}

// assertSnapshotRootRetained proves the protected snapshot a scope was
// installed from is never a sweep target.
func assertSnapshotRootRetained(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	project := t.TempDir()
	commit := strings.Repeat("3", 40)
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-a", commit)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	snapshotDir := filepath.Join(home, "cache", "skill-a", commit, "snapshot")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := "---\nname: skill-a\n---\n"
	if err := os.WriteFile(filepath.Join(snapshotDir, "SKILL.md"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	// An unreferenced runtime entry gives the pass something it may remove, so
	// snapshot retention is measured against a pass that actually swept.
	if err := os.MkdirAll(filepath.Join(home, "runtime", "skill-a", strings.Repeat("4", 40)), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedRuntime) == 0 {
		t.Fatal("the pass swept nothing, so snapshot retention proves nothing")
	}
	kept, err := os.ReadFile(filepath.Join(snapshotDir, "SKILL.md")) // #nosec G304 -- test fixture path
	if err != nil || string(kept) != payload {
		t.Fatalf("the protected snapshot root did not survive maintenance: %q, %v", kept, err)
	}
}

// assertUncertainRootRetained proves an entry the pass cannot account for stays
// on the machine: one unreadable marker makes the reference set unprovable, and
// the build sweep is skipped rather than guessed.
func assertUncertainRootRetained(t *testing.T) {
	t.Helper()
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	orphan := publishRealEntry(t, store, "orphan-tool", "orphan artifact")
	backdateEntry(t, home, orphan, 30*24*time.Hour)

	project := t.TempDir()
	broken := filepath.Join(project, ".agents", "skills", "skill-broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, marker.Name), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}

	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedBuilds) != 0 {
		t.Fatalf("an unprovable reference set still swept: %v", result.RemovedBuilds)
	}
	if _, err := os.Lstat(realEntryPath(home, orphan)); err != nil {
		t.Fatalf("an uncertain entry was removed: %v", err)
	}
	if !warned(result, "could not be proven complete") {
		t.Fatalf("the skipped sweep was not reported: %v", result.Warnings)
	}
	// The uncertain consumer is not forgotten either, so a second pass sees the
	// same uncertainty instead of inheriting a quietly emptied registry.
	if !containsString(LoadConsumers(home), project) {
		t.Fatalf("the uncertain consumer was pruned: %v", LoadConsumers(home))
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
