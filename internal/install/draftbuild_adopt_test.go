package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
)

// TestExternalAdoptionsListEveryPublishedEntryOnce: the commit adopts exactly
// the entries transactionPlan publishes — one artifact per newly built command
// in its receipt namespace, each snapshot once, nothing for a final-root hit —
// artifacts before snapshots, in key order.
func TestExternalAdoptionsListEveryPublishedEntryOnce(t *testing.T) {
	keyA := "sha256:" + strings.Repeat("a", 64)
	keyB := "sha256:" + strings.Repeat("b", 64)
	keyC := "sha256:" + strings.Repeat("c", 64)
	snapshot := "sha256:" + strings.Repeat("d", 64)
	staged := stagedExternal{root: "/staged", entries: map[string]map[string]externalEntry{
		"tooling": {
			"etool": {result: buildrepo.PipelineResult{CacheKey: keyB, ReceiptSchemaVersion: buildrepo.SourceAwareReceiptSchemaVersion, SnapshotKey: snapshot}, artifactPath: "/staged/x/artifact"},
			"ltool": {result: buildrepo.PipelineResult{CacheKey: keyA, ReceiptSchemaVersion: buildrepo.LegacyReceiptSchemaVersion, SnapshotKey: snapshot}, artifactPath: "/staged/y/artifact"},
			"hit":   {result: buildrepo.PipelineResult{CacheKey: keyC, ReceiptSchemaVersion: buildrepo.SourceAwareReceiptSchemaVersion, SnapshotKey: snapshot}, existing: true},
		},
	}}
	got := staged.adoptions("/final")
	want := []externalAdoption{
		{root: "/final", receiptSchemaVersion: buildrepo.LegacyReceiptSchemaVersion, key: keyA},
		{root: "/final", receiptSchemaVersion: buildrepo.SourceAwareReceiptSchemaVersion, key: keyB},
		{root: "/final", key: snapshot, snapshot: true},
	}
	if len(got) != len(want) {
		t.Fatalf("adoptions = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("adoption %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if list := (stagedExternal{}).adoptions("/final"); len(list) != 0 {
		t.Fatalf("an empty staged set adopts %+v", list)
	}
	// The adoption set is exactly the live paths the transaction publishes.
	plan := staged.transactionPlan("/final")
	if len(plan.Targets) != len(got) {
		t.Fatalf("transaction publishes %d targets but %d entries are adopted", len(plan.Targets), len(got))
	}
}

// TestAdoptExternalPublicationsWarnsWithoutCreatingAnything: a publication the
// store cannot prove is reported as a warning naming the kind and key; the
// committed install is not reverted and the store root is not invented.
func TestAdoptExternalPublicationsWarnsWithoutCreatingAnything(t *testing.T) {
	root := filepath.Join(t.TempDir(), "absent-external-build-cache")
	key := "sha256:" + strings.Repeat("e", 64)
	snapshot := "sha256:" + strings.Repeat("f", 64)
	targets := scopeTargets{adoptions: []externalAdoption{
		{root: root, receiptSchemaVersion: buildrepo.SourceAwareReceiptSchemaVersion, key: key},
		{root: root, key: snapshot, snapshot: true},
	}}
	warnings := adoptExternalPublications(commitRequest{scope: "project"}, targets)
	if len(warnings) != 2 ||
		!strings.Contains(warnings[0], "project: warning: external build cache artifact "+key+" was published but could not be adopted") ||
		!strings.Contains(warnings[1], "project: warning: external build cache snapshot "+snapshot+" was published but could not be adopted") {
		t.Fatalf("warnings = %q", warnings)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("a failed adoption created the store root: %v", err)
	}
	if warnings := adoptExternalPublications(commitRequest{scope: "project"}, scopeTargets{}); len(warnings) != 0 {
		t.Fatalf("no publications warned: %q", warnings)
	}
}
