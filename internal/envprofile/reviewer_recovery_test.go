// Committed reviewer recovery probes for TASK-260922-1t2w1q revision 3
// (TestReviewerRecoveryMarkerDrift, TestReviewerRecoveryPreservesRegularTemp).
// Production entry: ApplyMigration on temporary stores.
package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

func TestReviewerRecoveryMarkerDrift(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 {
		t.Fatalf("the plan carries two operations: %+v", report.Ops())
	}
	req.Expect = report.Hash
	// Kill after the first operation: registry order puts the codex
	// unlink first and the Pi relink second.
	crash := req
	crash.InjectFault = func(point string, applied int) error {
		if point == MigrateFaultCrash && applied >= 1 {
			return errors.New("kill -9")
		}
		return nil
	}
	interrupted, err := ApplyMigration(crash)
	if err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("the crash must interrupt, got %v", err)
	}
	if interrupted == nil || len(interrupted.Applied) != 1 {
		t.Fatalf("one operation stands at the kill: %+v", interrupted)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the first operation landed before the kill: %v", err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != oldTarget {
		t.Fatalf("the second operation never ran: %q (%v)", target, err)
	}
	journal, err := readMigrationJournal(fx.home)
	if err != nil || journal == nil {
		t.Fatalf("the journal stands after the kill: %+v (%v)", journal, err)
	}
	if journal.Plan != report.Hash || len(journal.Ops) != 2 || !journal.Ops[0].Done || journal.Ops[1].Done {
		t.Fatalf("the journal records one of two done: %+v", journal)
	}
	mp := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)
	raw, err := os.ReadFile(mp)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := envmarker.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	marker.Profile.LockSHA256 = strings.Repeat("a", 64)
	edited, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mp, edited, 0o600); err != nil {
		t.Fatal(err)
	}
	req.InjectFault = nil
	_, applyErr := ApplyMigration(req)
	after, err := os.ReadFile(mp)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(edited) {
		t.Errorf("recovery overwrote drifted marker before refusing: apply error=%v", applyErr)
	}
	if applyErr == nil {
		t.Error("stale plan accepted after recovery erased marker drift")
	}
}

func TestReviewerRecoveryPreservesRegularTemp(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 {
		t.Fatalf("the plan carries two operations: %+v", report.Ops())
	}
	req.Expect = report.Hash
	// Kill after the first operation: registry order puts the codex
	// unlink first and the Pi relink second.
	crash := req
	crash.InjectFault = func(point string, applied int) error {
		if point == MigrateFaultCrash && applied >= 1 {
			return errors.New("kill -9")
		}
		return nil
	}
	interrupted, err := ApplyMigration(crash)
	if err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("the crash must interrupt, got %v", err)
	}
	if interrupted == nil || len(interrupted.Applied) != 1 {
		t.Fatalf("one operation stands at the kill: %+v", interrupted)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the first operation landed before the kill: %v", err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != oldTarget {
		t.Fatalf("the second operation never ran: %q (%v)", target, err)
	}
	journal, err := readMigrationJournal(fx.home)
	if err != nil || journal == nil {
		t.Fatalf("the journal stands after the kill: %+v (%v)", journal, err)
	}
	if journal.Plan != report.Hash || len(journal.Ops) != 2 || !journal.Ops[0].Done || journal.Ops[1].Done {
		t.Fatalf("the journal records one of two done: %+v", journal)
	}
	stray := filepath.Join(filepath.Dir(piLink), ".migrate-operator-backup.tmp")
	if err := os.WriteFile(stray, []byte("operator credential backup"), 0o600); err != nil {
		t.Fatal(err)
	}
	req.InjectFault = nil
	_, err = ApplyMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stray); err != nil {
		t.Fatalf("recovery deleted unrelated regular file: %v", err)
	}
}
