package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// Shared production-owner fixture for migration's distinct publication/failure
// boundaries. All native bytes are synthetic. Only raw syscalls/fault points are
// replaced; planning, origin admission, whole-unit journal and recovery are real.
func migrationTransactionCase(t *testing.T, mode string) {
	t.Helper()
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	const credential = "fixture-migration-codex-sentinel"
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte(credential), 0600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	req := fx.migrateRequest()
	req.Machine.Isolation = map[string]map[string]string{fx.profile: {"codex_cli": "isolated"}}
	req.NativeHomeOf = func(string) (string, error) {
		t.Fatal("migration consulted current caller root")
		return "", os.ErrInvalid
	}
	plan, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Ops()) != 2 || len(plan.Conflicts()) != 0 {
		t.Fatalf("two-home healthy plan: %+v %+v", plan.Ops(), plan.Conflicts())
	}
	req.Expect = plan.Hash
	captures := map[string]string{}
	for _, id := range []string{"pi", "codex_cli"} {
		raw, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name))
		if err != nil {
			t.Fatal(err)
		}
		captures[id] = string(raw)
	}
	before := nativeSnapshot(t, EnvRoot(fx.home))
	nativeBefore := snapshotCredentialScope(t, fx)
	hit := false
	fault := errors.New("fixture migration boundary refusal")
	stopped := mode == "restart" || mode == "wrong-link" || mode == "marker-drift" || mode == "unrelated-temp"
	if mode == "symlink-error" || mode == "oversized-symlink" {
		req.transactionOptions = []transaction.Option{transaction.WithSymlink(func(target, path string) error {
			hit = true
			if mode == "oversized-symlink" {
				return os.Symlink(strings.Repeat("x", 100000), path)
			}
			return fault
		})}
	} else {
		req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
			point, class := transaction.PointAfterInstall, "credential-entry"
			if mode == "before-install" {
				point = transaction.PointBeforeInstall
			}
			if mode == "after-marker" {
				class = "marker"
			}
			if !hit && event.Point == point && event.Class == class {
				hit = true
				if stopped {
					panic("migration-whole-unit-stop")
				}
				return fault
			}
			return nil
		}})}
	}
	var result *MigrateApplyResult
	func() {
		if stopped {
			defer func() {
				if v := recover(); v != "migration-whole-unit-stop" {
					t.Fatalf("wrong restart stop: %v", v)
				}
			}()
		}
		result, err = ApplyMigration(req)
	}()
	if !hit {
		t.Fatalf("actual %s boundary not reached: result=%+v err=%v", mode, result, err)
	}
	assertCapture := func() {
		for id, want := range captures {
			raw, e := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name))
			if e != nil || string(raw) != want {
				t.Fatalf("%s origin recaptured: %v", id, e)
			}
		}
	}
	assertNoCopy := func() {
		holders := credentialHolders(t, fx, credential)
		if len(holders) != 1 || !holders[codexAuth] {
			t.Fatalf("credential copied: %v", holders)
		}
	}
	assertNoCopy()
	if !stopped {
		if mode == "oversized-symlink" {
			if err == nil || !strings.Contains(err.Error(), "file name too long") {
				t.Fatalf("expected actual oversized target refusal: %v", err)
			}
		} else if !errors.Is(err, fault) {
			t.Fatalf("wrong boundary refusal: %v", err)
		}
		if nativeSnapshot(t, EnvRoot(fx.home)) != before {
			t.Fatal("returned migration failure changed whole-home preimages")
		}
		added, removed, changed := snapshotDiff(nativeBefore, snapshotCredentialScope(t, fx))
		if len(added)+len(removed)+len(changed) != 0 {
			t.Fatalf("returned failure changed native/home entries: %v %v %v", added, removed, changed)
		}
		assertCapture()
		return
	}
	journals, err := filepath.Glob(filepath.Join(fx.home, "state", "transactions", "v1", "*.json"))
	if err != nil || len(journals) != 1 {
		t.Fatalf("actual pending whole-unit journal: %v %v", journals, err)
	}
	var stray string
	switch mode {
	case "wrong-link":
		if err = os.Remove(piLink); err != nil {
			t.Fatal(err)
		}
		if err = os.Symlink(filepath.Join(t.TempDir(), "untrusted-auth.json"), piLink); err != nil {
			t.Fatal(err)
		}
	case "marker-drift":
		markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
		marker, err := envmarker.Read(filepath.Dir(markerPath))
		if err != nil {
			t.Fatal(err)
		}
		marker.Profile.LockSHA256 = strings.Repeat("a", 64)
		raw, err := marker.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(markerPath, raw, 0600); err != nil {
			t.Fatal(err)
		}
	case "unrelated-temp":
		stray = filepath.Join(filepath.Dir(piLink), ".migrate-operator-backup.tmp")
		if err = os.WriteFile(stray, []byte("synthetic unrelated backup"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pendingBefore := nativeSnapshot(t, fx.home)
	op, err := beginOperation(fx.home)
	if mode == "wrong-link" || mode == "marker-drift" {
		if op != nil {
			_ = op.close()
		}
		if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) && !strings.Contains(err.Error(), "transaction_corrupt") {
			t.Fatalf("expected authority/preimage refusal: %v", err)
		}
		if nativeSnapshot(t, fx.home) != pendingBefore {
			t.Fatal("recovery refusal changed pending journal or substituted state")
		}
		assertNoCopy()
		return
	}
	if err != nil {
		t.Fatalf("genuine migration recovery: %v", err)
	}
	if err = op.close(); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != agentAuth {
		t.Fatalf("migration forward target %q, want %q (old %q): %v", target, agentAuth, oldTarget, err)
	}
	if _, err = os.Lstat(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), "auth.json")); !os.IsNotExist(err) {
		t.Fatalf("codex unlink missing: %v", err)
	}
	if stray != "" {
		raw, err := os.ReadFile(stray)
		if err != nil || string(raw) != "synthetic unrelated backup" {
			t.Fatalf("unrelated temp changed: %v", err)
		}
	}
	assertCapture()
	assertNoCopy()
	for _, id := range []string{"pi", "codex_cli"} {
		rr := fx.request(id)
		rr.Machine = req.Machine
		resolved, err := Resolve(rr)
		if err != nil || resolved == nil || len(resolved.Document) == 0 {
			t.Fatalf("recovered %s continuation: %v", id, err)
		}
	}
}
