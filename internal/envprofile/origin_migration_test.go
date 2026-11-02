package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// Migration's actual inventory and apply owner use recorded A, including Pi's
// old-parent and agent-root paths, under today's B. Returned faults restore the
// complete unit; a fresh engine finishes a genuine forward prefix. This is a
// synthetic restart model, not a provider/account/power-loss witness.
func TestProvisioningOriginMigrationOwner(t *testing.T) {
	enableV1WritersForTest(t)
	for _, mode := range []string{"healthy", "schema1", "legacy", "returned-error", "restart"} {
		t.Run(mode, func(t *testing.T) {
			fx, rr, agent := piOriginRequest(t, true, false, "a")
			if _, err := Resolve(rr); err != nil {
				t.Fatal(err)
			}
			home := ManagedHomeDir(fx.home, fx.profile, "pi")
			origin, err := os.ReadFile(filepath.Join(home, envorigin.Name))
			if err != nil {
				t.Fatal(err)
			}
			if mode == "schema1" {
				marker, err := envmarker.Read(home)
				if err != nil {
					t.Fatal(err)
				}
				marker.Version = envmarker.VersionV1
				entries := []envmarker.Passthrough{}
				for _, entry := range *marker.Passthrough {
					entries = append(entries, envmarker.Passthrough{Path: entry.Path, Strategy: entry.Strategy})
				}
				marker.Passthrough = &entries
				raw, err := marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(home, envmarker.Name), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			link := filepath.Join(home, "auth.json")
			if err = os.Remove(link); err != nil {
				t.Fatal(err)
			}
			old := filepath.Join(filepath.Dir(agent), "auth.json")
			if err = os.Symlink(old, link); err != nil {
				t.Fatal(err)
			}
			if mode == "legacy" {
				if err = os.Remove(filepath.Join(home, envorigin.Name)); err != nil {
					t.Fatal(err)
				}
			}
			b := t.TempDir()
			t.Setenv("PI_CODING_AGENT_DIR", b)
			observations := 0
			req := MigrateRequest{Home: fx.home, Profile: fx.profile, EnvID: "pi", Machine: rr.Machine, Detect: func(envregistry.Adapter) string { observations++; return "unknown" }, NativeHomeOf: func(string) (string, error) {
				t.Error("migration consulted caller native root B")
				return "", fmt.Errorf("forbidden B lookup")
			}}
			before := nativeSnapshot(t, home)
			plan, err := PlanMigration(req)
			if mode == "legacy" {
				if err == nil || !strings.Contains(err.Error(), envorigin.DiagRequired) || observations != 0 {
					t.Fatalf("legacy common migration refusal: observed=%d err=%v", observations, err)
				}
				req.Expect = strings.Repeat("a", 64)
				if _, err = ApplyMigration(req); err == nil || !strings.Contains(err.Error(), envorigin.DiagRequired) {
					t.Fatalf("legacy apply refusal: %v", err)
				}
				if nativeSnapshot(t, home) != before {
					t.Fatal("legacy migration changed home")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Homes) != 1 || plan.Homes[0].PiOldPath != old || plan.Homes[0].PiNewPath != filepath.Join(agent, "auth.json") {
				t.Fatalf("migration Pi authority vector: old=%s new=%s", plan.Homes[0].PiOldPath, plan.Homes[0].PiNewPath)
			}
			if len(plan.Conflicts()) != 0 || len(plan.Ops()) != 1 || plan.Ops()[0].To != filepath.Join(agent, "auth.json") {
				t.Fatalf("recorded migration plan: %+v", plan)
			}
			req.Expect = plan.Hash
			hit := false
			fault := errors.New("fixture credential migration fault")
			if mode != "healthy" && mode != "schema1" {
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if !hit && event.Point == transaction.PointAfterInstall && event.Class == "credential-entry" {
						hit = true
						if mode == "restart" {
							panic("migration-unit-stop")
						}
						return fault
					}
					return nil
				}})}
			}
			var result *MigrateApplyResult
			func() {
				if mode == "restart" {
					defer func() {
						if v := recover(); v != "migration-unit-stop" {
							t.Fatalf("wrong stop: %v", v)
						}
					}()
				}
				result, err = ApplyMigration(req)
			}()
			if mode == "returned-error" {
				if !hit || !errors.Is(err, fault) {
					t.Fatalf("migration fault owner not reached: hit=%t err=%v", hit, err)
				}
				if nativeSnapshot(t, home) != before {
					t.Fatal("migration fault changed whole home preimages")
				}
				return
			}
			if mode == "restart" {
				if !hit {
					t.Fatal("migration restart owner not reached")
				}
				op, e := beginOperation(fx.home)
				if e != nil {
					t.Fatalf("fresh migration unit recovery: %v", e)
				}
				if e = op.close(); e != nil {
					t.Fatal(e)
				}
			} else if err != nil || result == nil || len(result.Applied) != 1 {
				t.Fatalf("actual migration apply: %+v %v", result, err)
			}
			if actual, e := os.Readlink(link); e != nil || actual != filepath.Join(agent, "auth.json") {
				t.Fatalf("migration final target %q %v", actual, e)
			}
			raw, e := os.ReadFile(filepath.Join(home, envorigin.Name))
			if e != nil || string(raw) != string(origin) {
				t.Fatal("migration recaptured origin")
			}
			rr.Repair = false
			if result, e := Resolve(rr); e != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("migrated actual Resolve continuation: %v", e)
			}
		})
	}
}

// Identity-changing repair includes inactive Pi and Codex marker/origin bindings
// in the same durable lock transition. Capture facts remain byte-equivalent;
// returned failure restores L0 and a fresh engine completes a genuine prefix.
func TestProvisioningOriginIdentityMigrationOwner(t *testing.T) {
	for _, mode := range []string{"healthy", "returned-error", "restart"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, _ := recordedOriginFixture(t, "a")
			agent := filepath.Join(fx.native["pi"], "agent")
			if err := os.MkdirAll(agent, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(agent, "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			pi := fx.request("pi")
			pi.Repair = true
			if _, err := Resolve(pi); err != nil {
				t.Fatal(err)
			}
			captures := map[string][]byte{}
			for _, id := range []string{"codex_cli", "pi"} {
				raw, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name))
				if err != nil {
					t.Fatal(err)
				}
				captures[id] = raw
			}
			lockBefore, err := os.ReadFile(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
			enableV2WritersForTest(t)
			hit := false
			fault := errors.New("identity origin publication fault")
			req.Repair = true
			if mode != "healthy" {
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if !hit && event.Point == transaction.PointAfterInstall && event.Class == "origin" {
						hit = true
						if mode == "restart" {
							panic("identity-unit-stop")
						}
						return fault
					}
					return nil
				}})}
			}
			var result *ResolveResult
			func() {
				if mode == "restart" {
					defer func() {
						if v := recover(); v != "identity-unit-stop" {
							t.Fatalf("wrong identity stop: %v", v)
						}
					}()
				}
				result, err = Resolve(req)
			}()
			if mode == "returned-error" {
				if !hit || !errors.Is(err, fault) {
					t.Fatalf("identity fault not reached: %t %v", hit, err)
				}
				raw, e := os.ReadFile(lockPath(fx.home, fx.profile))
				if e != nil || string(raw) != string(lockBefore) || nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != before {
					t.Fatal("identity rollback changed group preimages")
				}
				return
			}
			if mode == "restart" {
				if !hit {
					t.Fatal("identity restart not reached")
				}
				req.transactionOptions = nil
				result, err = Resolve(req)
			}
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("actual identity continuation: %v", err)
			}
			_, hash, e := readLock(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			if strings.TrimPrefix(hash, "sha256:") == strings.TrimPrefix(contextlock.HashBytes(lockBefore), "sha256:") {
				t.Fatal("identity owner did not change lock")
			}
			for id, old := range captures {
				target := ManagedHomeDir(fx.home, fx.profile, id)
				_, raw, e := envorigin.Read(fx.home, envorigin.Binding{Profile: fx.profile, ManagedTarget: target, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:")}, id)
				if e != nil {
					t.Fatalf("inactive %s rebind: %v", id, e)
				}
				a, _ := envorigin.CaptureFacts(old)
				b, _ := envorigin.CaptureFacts(raw)
				if string(a) != string(b) {
					t.Fatal("identity migration recaptured original authority")
				}
			}
		})
	}
}

// Common admission sees a legacy migration before every continuation. Only a
// sealed, independently recorded, exact completed postimage permits cleanup;
// an unsealed or substituted record never grants a forward/rollback exception.
func TestProvisioningOriginLegacyMigrationBarrier(t *testing.T) {
	enableV1WritersForTest(t)
	for _, state := range []string{"sealed", "unsealed", "wrong-target", "foreign-path", "duplicate", "marker-edit", "unknown-field"} {
		t.Run(state, func(t *testing.T) {
			fx, rr, agent := piOriginRequest(t, true, false, "a")
			if _, err := Resolve(rr); err != nil {
				t.Fatal(err)
			}
			target := ManagedHomeDir(fx.home, fx.profile, "pi")
			marker, err := os.ReadFile(filepath.Join(target, envmarker.Name))
			if err != nil {
				t.Fatal(err)
			}
			j := &migrationJournal{Version: 1, Plan: strings.Repeat("a", 64), Complete: true, MarkersDone: true,
				Ops:     []migrationJournalOp{{Profile: fx.profile, EnvID: "pi", Kind: MigrateOpRelink, Path: "auth.json", From: filepath.Join(filepath.Dir(agent), "auth.json"), To: filepath.Join(agent, "auth.json"), Done: true}},
				Markers: []migrationJournalMarker{{Live: filepath.Join(target, envmarker.Name), Prior: marker, New: marker}}}
			switch state {
			case "unsealed":
				j.Complete = false
			case "wrong-target":
				j.Ops[0].To = filepath.Join(t.TempDir(), "auth.json")
			case "foreign-path":
				j.Ops[0].Path = "skills/foreign"
			case "duplicate":
				j.Ops = append(j.Ops, j.Ops[0])
			case "marker-edit":
				j.Markers[0].New = append([]byte(nil), append(marker, ' ')...)
			}
			if err := writeMigrationJournal(fx.home, j); err != nil {
				t.Fatal(err)
			}
			if state == "unknown-field" {
				path := migrationJournalPath(fx.home)
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				raw = append([]byte(`{"unexpected":true,`), raw[1:]...)
				if err = os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := nativeSnapshot(t, fx.home)
			homeBefore := nativeSnapshot(t, target)
			rr.originAccountLookup = func() (envorigin.Account, error) {
				t.Fatal("legacy journal triggered discovery")
				return envorigin.Account{}, nil
			}
			seedReads := 0
			rr.NativeHomeOf = func(id string) (string, error) { seedReads++; return fx.native[id], nil }
			backendReads := 0
			rr.passthroughTargetStat = func(path string) (os.FileInfo, error) {
				backendReads++
				if path != filepath.Join(agent, "auth.json") {
					t.Fatal("credential lookup escaped A")
				}
				return os.Stat(path)
			}
			result, err := Resolve(rr)
			if state != "sealed" {
				if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || result != nil && len(result.Document) > 0 {
					t.Fatalf("legacy journal continuation refusal: %v", err)
				}
				if seedReads != 0 || backendReads != 0 {
					t.Fatal("legacy refusal reached seed/backend lookup")
				}
				if nativeSnapshot(t, fx.home) != before {
					t.Fatal("legacy journal refusal changed preimages")
				}
				return
			}
			if err != nil || len(result.Document) == 0 {
				t.Fatalf("sealed exact actual continuation: %v", err)
			}
			req := MigrateRequest{Home: fx.home, Profile: fx.profile, EnvID: "pi", Machine: rr.Machine, Detect: rr.Detect}
			plan, err := PlanMigration(req)
			if err != nil {
				t.Fatal(err)
			}
			req.Expect = plan.Hash
			applied, err := ApplyMigration(req)
			if err == nil || !strings.Contains(err.Error(), "migration plan drift") || applied.Recovered == nil {
				t.Fatalf("sealed cleanup must invalidate the prior recovery-bound plan: %v", err)
			}
			plan, err = PlanMigration(req)
			if err != nil {
				t.Fatal(err)
			}
			req.Expect = plan.Hash
			if _, err = ApplyMigration(req); err != nil {
				t.Fatalf("explicit fresh plan after cleanup: %v", err)
			}
			if _, err = os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
				t.Fatalf("sealed journal cleanup: %v", err)
			}
			if nativeSnapshot(t, target) != homeBefore {
				t.Fatal("sealed cleanup changed managed preimages")
			}
		})
	}
}
