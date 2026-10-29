package envlifecycle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envunit"
)

func fixtureLedger(t *testing.T) (*Ledger, Entry) {
	t.Helper()
	manager := filepath.Join(t.TempDir(), "manager")
	if err := os.Mkdir(manager, 0700); err != nil {
		t.Fatal(err)
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		t.Fatal(err)
	}
	entry := Entry{Purged: &Purged{ID: "purged-p", Kind: "purged", Phase: "complete", Manager: identity, PriorRemovalID: "removed-p", ProfileName: "p", TargetNamespaces: []string{filepath.Join(manager, "environments", "p", "codex_cli")}, RemovalCommit: Commit{Unit: "remove-p", Evidence: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, PurgeCommit: Commit{Unit: "purge-p", Evidence: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, ExpectedAbsence: []string{filepath.Join(manager, "profiles", "p"), filepath.Join(manager, "environments", "p", "codex_cli")}}}
	return &Ledger{Record: Version, Manager: identity, Revision: 1, Entries: []Entry{entry}}, entry
}

// Proves closed variant discrimination and absence-only purged classification;
// it does not exercise remove, quarantine or a profile continuation owner.
func TestProvisioningOriginLifecycleLedger(t *testing.T) {
	ledger, _ := fixtureLedger(t)
	raw, err := ledger.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(raw, ledger.Manager); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{
		bytes.Replace(raw, []byte(`"revision":`), []byte(`"Revision":`), 1),
		bytes.Replace(raw, []byte(`"kind":"purged"`), []byte(`"kind":"removed-retained"`), 1),
		bytes.Replace(raw, []byte(`"phase":"complete"`), []byte(`"phase":"retained"`), 1),
		bytes.Replace(raw, []byte(`"prior_removal_id":`), []byte(`"unexpected":`), 1),
		append(raw, []byte(`{}`)...),
	} {
		if _, err := Parse(bad, ledger.Manager); err == nil {
			t.Fatalf("accepted invalid lifecycle %s", bad)
		}
	}
	if err := ledger.VerifyAbsence(); err != nil {
		t.Fatal(err)
	}
	target := ledger.Entries[0].Purged.TargetNamespaces[0]
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ledger.VerifyAbsence(); err == nil {
		t.Fatal("new namespace was exempted by historical purge")
	}
}

// Proves structured CAS cannot silently delete/replace protected P for Q. The
// returned bytes are only a plan; publication still requires a real journal CAS.
func TestProvisioningOriginLifecycleStructuredCAS(t *testing.T) {
	ledger, p := fixtureLedger(t)
	raw, err := ledger.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PlanReplace(raw, ledger.Manager, 1, []Entry{}, nil); err == nil {
		t.Fatal("dropped protected entry")
	}
	q := p
	q.Purged = new(Purged)
	*q.Purged = *p.Purged
	q.Purged.ID = "purged-q"
	q.Purged.ProfileName = "q"
	q.Purged.TargetNamespaces = []string{filepath.Join(ledger.Manager.Path, "environments", "q", "pi")}
	q.Purged.ExpectedAbsence = []string{filepath.Join(ledger.Manager.Path, "profiles", "q"), q.Purged.TargetNamespaces[0]}
	next, err := PlanReplace(raw, ledger.Manager, 1, []Entry{p, q}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(next, ledger.Manager)
	if err != nil || got.Revision != 2 || len(got.Entries) != 2 {
		t.Fatalf("CAS result %+v: %v", got, err)
	}
	before, _ := p.Marshal()
	after, _ := got.Entries[0].Marshal()
	if !bytes.Equal(before, after) {
		t.Fatal("disjoint Q rewrote P")
	}
	if _, err := PlanReplace(raw, ledger.Manager, 2, []Entry{p, q}, nil); err == nil {
		t.Fatal("stale ledger revision accepted")
	}
}
