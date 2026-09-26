// Production entry points under test: Resolve and ApplyMigration for
// the Decision 0017 follow-through rows (environments §7.4, §10.1;
// manager §12.4, §12.5) — the repair_failed refusal class and the
// migration lock. Every row drives a production entry on a temporary
// store; helper-direct assertions are bounds, not rows.
//
// This file also carries requireLinkCapability, the Windows symlink
// privilege probe shared by every 0017 row whose fixture or production
// repair needs symlinks.
package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
)

// requireLinkCapability skips the row when the host cannot create the
// symlinks the fixture or the production repair needs — the Windows
// privilege pattern shared with TestSymlinkInPathIsSourceInvalid. The
// reason classifies host-capability in .github/ci/skip-classes.tsv via
// the existing "this host cannot create" entry; ledger rows tolerate
// it on Windows only, so a skip on a unix runner still fails the gate.
// A symlink failure after this probe passed is a real defect and
// fails, never skips: direct os.Symlink sites keep t.Fatal on purpose.
func requireLinkCapability(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("probe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "link")); err != nil {
		t.Skipf("this host cannot create symlinks: %v", err)
	}
}

// TestCredentialLinkInspectionRepairFails is the
// environment_repair_failed row: a correctly targeted link whose native
// target cannot be inspected stays stale past repair — there is no
// re-link that heals it — so repair fails with
// environment_repair_failed carrying the conflict/inspection reason,
// never the bare conflict and never silence, with the link untouched.
// Restoring the native side heals the home and repair converges. A
// mutant that returns the bare conflict instead of wrapping it fails
// the repair_failed assertion; one that reports the state current
// fails the staleness. A regular file where the agent directory should
// be fails the target stat deterministically on Unix, even as root;
// Windows maps a stat through a file to path-not-found, so the row is
// Unix-only like TestCredentialLinkTargetInspectionFailure.
func TestCredentialLinkInspectionRepairFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this host cannot create an uninspectable link target: Windows maps a stat through a file to path-not-found, so the inspection diagnostic runs on the unix runners")
	}
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("the live home is current: %v", err)
	}
	// Break the target's parent: a file where the agent directory was.
	if err := os.RemoveAll(filepath.Join(native, "agent")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "agent"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale, err := Resolve(fx.request("pi"))
	if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("an uninspectable target must be stale, got %v", err)
	}
	if reasons := strings.Join(stale.StaleReasons, "; "); !strings.Contains(reasons, "cannot be inspected") {
		t.Fatalf("stale reasons carry the inspection diagnostic: %q", reasons)
	}
	req := fx.request("pi")
	req.Repair = true
	_, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("repair of an uninspectable target must fail with %s, got %v", DiagRepairFailed, err)
	}
	for _, want := range []string{"cannot be inspected", envregistry.DiagCredentialConflict, "auth.json", agentAuth} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the repair failure carries the inspection reason, want %q in %v", want, err)
		}
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("repair touches nothing: %q (%v), want %q", target, err, agentAuth)
	}
	// Restoring the native side heals the home: repair converges and
	// the link reads through.
	if err := os.Remove(filepath.Join(native, "agent")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(native, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("repair converges once the native target is back: %v", err)
	}
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("the healed home is current: %v", err)
	}
	if payload, err := os.ReadFile(link); err != nil || string(payload) != "{\"t\":\"operator-pi\"}\n" {
		t.Fatalf("the link serves the restored bytes: %q (%v)", payload, err)
	}
}

// TestMigrateApplyLockContention proves the migration apply runs under
// the manager-home mutation lock like every other profile mutation: a
// contended apply refuses with the distinct lock-acquisition
// diagnostic — never repair_failed, never a partial apply — with zero
// writes and no journal. A mutant that drops the lock acquisition
// applies cleanly and fails the refusal; one that swallows only the
// timeout still refuses here and passes, so the row pins the timeout
// shape production reports.
func TestMigrateApplyLockContention(t *testing.T) {
	requireLinkCapability(t)
	fx, link, oldTarget, _ := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 {
		t.Fatalf("one relink op: %+v", report.Ops())
	}
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := manager.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	previous := lockTimeout
	lockTimeout = 0
	defer func() { lockTimeout = previous }()
	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), DiagLockUnavailable) {
		t.Fatalf("a contended apply must fail with %s, got %v", DiagLockUnavailable, err)
	}
	if strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("lock contention must not masquerade as %s: %v", DiagRepairFailed, err)
	}
	if result != nil && len(result.Applied) != 0 {
		t.Fatalf("a contended apply executes nothing: %+v", result)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("a contended apply journals nothing: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a contended apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}
