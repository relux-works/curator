package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/transaction"
)

type projectBinTestImage struct {
	bytes  []byte
	absent bool
}

// Preserve real file bytes or verified absence; a failed read of an existing
// entry must never masquerade as an absent preimage in ownership tests.
func captureProjectBinImages(t *testing.T, paths ...string) map[string]projectBinTestImage {
	t.Helper()
	images := map[string]projectBinTestImage{}
	for _, path := range paths {
		payload, err := os.ReadFile(path)
		if err == nil {
			images[path] = projectBinTestImage{bytes: payload}
			continue
		}
		if _, statErr := os.Lstat(path); !os.IsNotExist(err) || !os.IsNotExist(statErr) {
			t.Fatalf("capture existing state %s: %v", path, err)
		}
		images[path] = projectBinTestImage{absent: true}
	}
	return images
}

func assertProjectBinImages(t *testing.T, images map[string]projectBinTestImage) {
	t.Helper()
	for path, image := range images {
		if image.absent {
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Fatalf("refusal published previously absent state %s: %v", path, err)
			}
			continue
		}
		if payload, err := os.ReadFile(path); err != nil || string(payload) != string(image.bytes) {
			t.Fatalf("refusal changed installed state %s: %v", path, err)
		}
	}
}

// Legacy enforced launchers use the marker-selected manifest and exact sidecar;
// altered/missing evidence refuses with every installed byte preserved. A valid
// ledger and an adopted pair both upgrade; removal deletes both adopted files.
func TestProjectBinLegacyEnforcedEvidence(t *testing.T) {
	for _, scenario := range []string{"ledger", "upgrade", "remove", "sidecar-mismatch", "marker-absent"} {
		t.Run(scenario, func(t *testing.T) {
			e := newEnv(t)
			e.schema8ScriptSkill("enforced-skill", true)
			e.declare("enforced-skill")
			if result := e.install(Options{}); result.Status != "ok" {
				t.Fatalf("positive control: %+v", result)
			}
			bin := filepath.Join(e.project, ".agents", "bin")
			launcher := filepath.Join(bin, nativeLauncherName("enforced-skill-tool"))
			sidecar := launcher + ".curator-shim.json"
			ledger := filepath.Join(filepath.Dir(bin), projectBinLedgerName)
			if scenario != "ledger" {
				if err := os.Remove(ledger); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "sidecar-mismatch" {
				if err := os.WriteFile(sidecar, []byte("{}\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "marker-absent" {
				if err := os.Remove(filepath.Join(filepath.Dir(bin), "skills", "enforced-skill", ".csk-install.json")); err != nil {
					t.Fatal(err)
				}
			}
			before := captureProjectBinImages(t, launcher, sidecar, ledger, sourcelock.PathIn(e.project))
			if scenario == "remove" {
				e.declare()
			}
			result := e.install(Options{})
			if scenario == "sidecar-mismatch" || scenario == "marker-absent" {
				if result.Status == "ok" || !strings.Contains(fmt.Sprint(result.Errors), "project_bin_legacy_enforced") {
					t.Fatalf("intended refusal: %+v", result)
				}
				assertProjectBinImages(t, before)
				return
			}
			if result.Status != "ok" {
				t.Fatalf("reconcile: %+v", result)
			}
			for _, path := range []string{launcher, sidecar} {
				_, err := os.Lstat(path)
				if scenario == "remove" {
					if !os.IsNotExist(err) {
						t.Fatalf("adopted entry survived: %s %v", path, err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// Unrelated unreadable files are reported without reading or fingerprinting;
// a nearby declared directory still refuses as a collision on every platform.
func TestProjectBinUnreadableUnrelated(t *testing.T) {
	e := newEnv(t)
	e.skill("tool")
	e.declare("tool")
	bin := filepath.Join(e.project, ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(bin, "private-tool")
	if err := os.WriteFile(path, []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Run("unreadable", func(t *testing.T) {
		denyProjectLedgerRead(t, path)
		for _, dry := range []bool{true, false} {
			result := e.install(Options{DryRun: dry})
			if result.Status != "ok" || !strings.Contains(fmt.Sprint(result.Messages), projectBinPreserved) || !strings.Contains(fmt.Sprint(result.Messages), path) {
				t.Fatalf("dry=%v unrelated entry was read or report lost: %+v", dry, result)
			}
		}
		// A directory is a platform-neutral collision control, including Windows.
		declared := filepath.Join(bin, "declared")
		if err := os.Mkdir(declared, 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := inspectProjectBin(t.TempDir(), bin, installPlatform(), "", map[string]bool{declared: true}); err == nil || !strings.Contains(err.Error(), projectBinConflict) {
			t.Fatalf("declared collision: %v", err)
		}
		if _, err := os.Lstat(path); err != nil {
			t.Fatal(err)
		}
	})
	if got, err := os.ReadFile(path); err != nil || string(got) != "foreign" {
		t.Fatalf("restored foreign preimage: %q %v", got, err)
	}
}

// Both dry-run and install refuse a colliding path before the staged-build hook;
// removing the collision permits the ordinary installation positive control.
func TestProjectBinDryRunPreflight(t *testing.T) {
	e := newEnv(t)
	e.skill("tool")
	e.declare("tool")
	path := filepath.Join(e.project, ".agents", "bin", shimName("tool-tool"))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("foreign"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, dry := range []bool{true, false} {
		called := false
		result := e.install(Options{DryRun: dry, OnStaged: func(Staged) error { called = true; return errors.New("build hook reached") }})
		if result.Status == "ok" || !strings.Contains(fmt.Sprint(result.Errors), projectBinConflict) || called {
			t.Fatalf("dry=%v %+v build=%v", dry, result, called)
		}
		if got, _ := os.ReadFile(path); string(got) != "foreign" {
			t.Fatal("collision changed")
		}
		if _, err := os.Lstat(filepath.Join(filepath.Dir(filepath.Dir(path)), projectBinLedgerName)); !os.IsNotExist(err) {
			t.Fatalf("preflight published ledger: %v", err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("positive: %+v", result)
	}
}

// Commit-phase ledger faults restore the prior shim and absent ledger; retry
// publishes the pair. This plant must fire after backup, not during prepare.
func TestProjectBinLedgerCommitRollbackRetry(t *testing.T) {
	e := newEnv(t)
	e.skill("tool")
	e.declare("tool")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("positive: %+v", result)
	}
	ledger := filepath.Join(e.project, ".agents", projectBinLedgerName)
	shim := filepath.Join(e.project, ".agents", "bin", shimName("tool-tool"))
	before, err := os.ReadFile(shim)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(ledger); err != nil {
		t.Fatal(err)
	}
	fired := false
	result := e.install(Options{Commit: CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointAfterBackup && strings.Contains(event.Identifier, "project-bin-ownership") {
			fired = true
			return errors.New("ledger commit fault")
		}
		return nil
	}}}})
	if !fired || result.Status == "ok" || !strings.Contains(fmt.Sprint(result.Errors), "ledger commit fault") {
		t.Fatalf("plant did not fire as intended: %+v fired=%v", result, fired)
	}
	if got, _ := os.ReadFile(shim); string(got) != string(before) {
		t.Fatal("rollback changed launcher")
	}
	if _, err := os.Lstat(ledger); !os.IsNotExist(err) {
		t.Fatalf("rollback left ledger: %v", err)
	}
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("retry: %+v", result)
	}
	if _, err := os.Stat(ledger); err != nil {
		t.Fatal(err)
	}
}

// Foreign directories, live links and dangling links refuse without touching
// the entry or link destination. The adjacent absent-path install is valid.
func TestProjectBinCollisionKinds(t *testing.T) {
	for _, kind := range []string{"directory", "symlink", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			e := newEnv(t)
			e.skill("tool")
			e.declare("tool")
			path := filepath.Join(e.project, ".agents", "bin", shimName("tool-tool"))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(t.TempDir(), "foreign")
			if err := os.WriteFile(target, []byte("target"), 0644); err != nil {
				t.Fatal(err)
			}
			if kind == "directory" {
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "keep"), []byte("keep"), 0644); err != nil {
					t.Fatal(err)
				}
			} else {
				link := target
				if kind == "dangling" {
					link += "-missing"
				}
				if err := os.Symlink(link, path); err != nil {
					t.Fatal(err)
				}
			}
			result := e.install(Options{})
			if result.Status == "ok" || !strings.Contains(fmt.Sprint(result.Errors), projectBinConflict) {
				t.Fatalf("refusal: %+v", result)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatal(err)
			}
			if got, _ := os.ReadFile(target); string(got) != "target" {
				t.Fatal("destination changed")
			}
			if kind == "directory" {
				if got, _ := os.ReadFile(filepath.Join(path, "keep")); string(got) != "keep" {
					t.Fatal("directory changed")
				}
			} else {
				want := target
				if kind == "dangling" {
					want += "-missing"
				}
				if got, _ := os.Readlink(path); got != want {
					t.Fatal("link changed")
				}
			}
			if _, err := os.Lstat(filepath.Join(e.project, ".agents", projectBinLedgerName)); !os.IsNotExist(err) {
				t.Fatalf("ledger published: %v", err)
			}
			if kind == "directory" {
				if err := os.Remove(filepath.Join(path, "keep")); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if result := e.install(Options{}); result.Status != "ok" {
				t.Fatalf("absent-path positive control: %+v", result)
			}
		})
	}
}

// A real build-plan fixture refuses before invoking the builder, both during
// dry-run and install. Removing the collision reaches the builder positive
// control; the injected failure keeps publication out of this witness.
func TestProjectBinPreflightBeforeBuilder(t *testing.T) {
	e, deps, builder := transactionFixture(t)
	bin := filepath.Join(e.project, ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(bin, shimName("golden-tool"))
	if err := os.WriteFile(path, []byte("foreign"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, dry := range []bool{true, false} {
		result := e.install(Options{DryRun: dry, Build: deps, External: lifecycleExternalDeps()})
		if result.Status == "ok" || !strings.Contains(fmt.Sprint(result.Errors), projectBinConflict) || len(builder.calls) != 0 {
			t.Fatalf("dry=%v result=%+v calls=%v", dry, result, builder.calls)
		}
		if got, _ := os.ReadFile(path); string(got) != "foreign" {
			t.Fatal("foreign bytes changed")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	builder.failOn["golden-tool"] = errors.New("positive builder reached")
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps()})
	if len(builder.calls) == 0 || !strings.Contains(fmt.Sprint(result.Errors), "positive builder reached") {
		t.Fatalf("positive control: %+v calls=%v", result, builder.calls)
	}
}

// A valid ledger passes; denial of its read fails with the state-read code,
// keeps the ledger in place, and succeeds again after access is restored.
func TestProjectBinUnreadableLedger(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(filepath.Dir(bin), projectBinLedgerName)
	payload := []byte(`{"schema_version":1,"entries":{}}`)
	if err := os.WriteFile(ledger, payload, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectProjectBin(t.TempDir(), bin, installPlatform(), "", nil); err != nil {
		t.Fatalf("positive: %v", err)
	}
	t.Run("denied", func(t *testing.T) {
		denyProjectLedgerRead(t, ledger)
		if _, err := inspectProjectBin(t.TempDir(), bin, installPlatform(), "", nil); err == nil || !strings.Contains(err.Error(), "manager_state_unreadable") {
			t.Fatalf("ledger refusal: %v", err)
		}
		if _, err := os.Lstat(ledger); err != nil {
			t.Fatal(err)
		}
	})
	if got, err := os.ReadFile(ledger); err != nil || string(got) != string(payload) {
		t.Fatalf("ledger changed: %s %v", got, err)
	}
	if _, err := inspectProjectBin(t.TempDir(), bin, installPlatform(), "", nil); err != nil {
		t.Fatalf("restored control: %v", err)
	}
}

// A stale enforced marker with neither live launcher nor sidecar does not load
// an unusable frozen manifest during a real upgrade. A live enforced pair still
// refuses unknown ownership; readable evidence permits upgrade/removal.
func TestProjectBinStaleMarkerManifest(t *testing.T) {
	for _, scenario := range []string{"readable-live", "unusable-stale", "unusable-live"} {
		t.Run(scenario, func(t *testing.T) {
			e := newEnv(t)
			e.schema8ScriptSkill("old", true)
			e.skill("new")
			e.declare("old")
			if r := e.install(Options{}); r.Status != "ok" {
				t.Fatalf("positive install: %+v", r)
			}
			bin := filepath.Join(e.project, ".agents", "bin")
			launcher := filepath.Join(bin, nativeLauncherName("old-tool"))
			sidecar := launcher + ".curator-shim.json"
			installed := filepath.Join(e.project, ".agents", "skills", "old")
			m, _, err := marker.ReadState(installed)
			if err != nil || m == nil {
				t.Fatalf("marker: %v", err)
			}
			source := m.Source
			if m.Git != "" {
				source = identity.Canonical(m.Git)
			}
			frozen := filepath.Join(snapshot.Dir(e.home, source, m.Commit), "agent-skill.json")
			for _, path := range []string{launcher, sidecar, frozen} {
				if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
					t.Fatalf("enforced fixture precondition %s: %v", path, err)
				}
			}
			if scenario != "readable-live" {
				if err := os.Chmod(frozen, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(frozen, []byte("{broken"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Remove(filepath.Join(filepath.Dir(bin), projectBinLedgerName)); err != nil {
				t.Fatal(err)
			}
			if scenario == "unusable-stale" {
				for _, path := range []string{launcher, sidecar} {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				}
			}
			// Schema-1 installs do not create the canonical portable lock.
			lockPath := sourcelock.PathIn(e.project)
			if _, err := os.Lstat(lockPath); !os.IsNotExist(err) {
				t.Fatalf("legacy fixture unexpectedly has a portable lock: %v", err)
			}
			before := captureProjectBinImages(t, launcher, sidecar, filepath.Join(installed, ".csk-install.json"), lockPath)
			e.declare("new")
			r := e.install(Options{})
			if scenario == "unusable-live" {
				if r.Status == "ok" || !strings.Contains(fmt.Sprint(r.Errors), "project_bin_legacy_enforced") || !strings.Contains(fmt.Sprint(r.Errors), "move the legacy launcher and its sidecar aside") {
					t.Fatalf("unknown live evidence: %+v", r)
				}
				assertProjectBinImages(t, before)
				if _, err := os.Lstat(filepath.Join(filepath.Dir(bin), projectBinLedgerName)); !os.IsNotExist(err) {
					t.Fatalf("refusal published ledger: %v", err)
				}
			} else if r.Status != "ok" {
				t.Fatalf("stale/readable upgrade: %+v", r)
			} else {
				for _, path := range []string{launcher, sidecar} {
					if _, err := os.Lstat(path); !os.IsNotExist(err) {
						t.Fatalf("old enforced entry survived removal: %s %v", path, err)
					}
				}
				if _, err := os.Stat(filepath.Join(bin, shimName("new-tool"))); err != nil {
					t.Fatalf("positive upgrade omitted new command: %v", err)
				}
			}
			if _, err := os.Lstat(lockPath); !os.IsNotExist(err) {
				t.Fatalf("legacy install published a portable lock: %v", err)
			}
		})
	}
}

// Desired, ledger-named and legacy candidates retain the stable state-read
// code and one actionable remedy. Readability before and after the isolated
// denial is a positive control; neither bytes nor the ledger may change.
func TestProjectBinRelevantReadDiagnostics(t *testing.T) {
	for _, kind := range []string{"desired", "ledger", "legacy"} {
		t.Run(kind, func(t *testing.T) {
			e := newEnv(t)
			e.skill("tool")
			e.declare("tool")
			if r := e.install(Options{}); r.Status != "ok" {
				t.Fatalf("positive: %+v", r)
			}
			bin := filepath.Join(e.project, ".agents", "bin")
			path := filepath.Join(bin, shimName("tool-tool"))
			ledger := filepath.Join(filepath.Dir(bin), projectBinLedgerName)
			if kind != "ledger" {
				if err := os.Remove(ledger); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.ReadFile(path)
			ledgerBefore, _ := os.ReadFile(ledger)
			desired := map[string]bool{}
			if kind == "desired" {
				desired[path] = true
			}
			t.Run("denied", func(t *testing.T) {
				denyProjectLedgerRead(t, path)
				_, err := inspectProjectBin(e.home, bin, installPlatform(), "", desired)
				if err == nil || !strings.Contains(err.Error(), "manager_state_unreadable") || strings.Count(err.Error(), "restore read access") != 1 {
					t.Fatalf("typed remedy: %v", err)
				}
				got, _ := os.ReadFile(ledger)
				if string(got) != string(ledgerBefore) {
					t.Fatal("ledger changed")
				}
			})
			if got, _ := os.ReadFile(path); string(got) != string(before) {
				t.Fatal("entry changed")
			}
			if _, err := inspectProjectBin(e.home, bin, installPlatform(), "", desired); err != nil {
				t.Fatalf("restored control: %v", err)
			}
		})
	}
}
