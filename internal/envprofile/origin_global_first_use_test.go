package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/transaction"
)

// First-use default records belong to the actual whole-unit journal. An
// ordinary returned fault after source publication restores their absence;
// only a confirmed in-place conflict retains source/L1, with all native
// preimages unchanged and no legacy planner continuation. Account discovery
// and native credentials are outside this fixture's exercised surface.
func TestProvisioningOriginGlobalFirstUse(t *testing.T) {
	for _, mode := range []string{"healthy", "source-fault", "retained", "retained-use", "retained-use-fault"} {
		t.Run(mode, func(t *testing.T) {
			enableV1WritersForTest(t)
			home, _ := migrationRevisionFixture(t, map[string]string{"SKILL.md": "# fixture skill\n"})
			pinOperatorHome(t, t.TempDir())
			homes := pinHomes(t)
			if strings.HasPrefix(mode, "retained") {
				path := filepath.Join(homes[envregistry.CodexCLI], "skills", "binary-skill")
				if err := os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("unmanaged"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := map[string]string{}
			for id, root := range homes {
				before[string(id)] = nativeSnapshot(t, root)
			}
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit, prepared := false, false
			newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
					if e.Point == transaction.PointPrepared {
						raw, err := os.ReadFile(filepath.Join(home, "state", "transactions", "v1", e.TransactionID+".json"))
						if err != nil {
							t.Fatal(err)
						}
						var j transaction.Journal
						if err := json.Unmarshal(raw, &j); err != nil {
							t.Fatal(err)
						}
						if j.Unit != nil && j.Unit.Action == envunit.GlobalExtend {
							prepared = true
							for _, path := range []string{sourcePath(home, DefaultProfile), lockPath(home, DefaultProfile)} {
								if _, err := os.Lstat(path); !os.IsNotExist(err) {
									t.Fatalf("first-use record published before whole-unit commit: %s %v", filepath.Base(path), err)
								}
							}
						}
					}
					if mode == "source-fault" && e.Point == transaction.PointAfterInstall && e.LivePath == sourcePath(home, DefaultProfile) {
						hit = true
						return errors.New("first-use source fault")
					}
					return nil
				}}))...)
			}
			info, err := GlobalInstall(home, Policy{}, envregistry.DefaultMachineConfig(), nil)
			newOperationEngine = saved
			if !prepared {
				t.Fatalf("actual first-use whole-unit journal not reached: %v", err)
			}
			if mode == "healthy" && err != nil {
				t.Fatal(err)
			}
			if mode == "source-fault" && (!hit || err == nil || !strings.Contains(err.Error(), "first-use source fault")) {
				t.Fatalf("actual source stage not reached: %t %v", hit, err)
			}
			if strings.HasPrefix(mode, "retained") {
				var retained *transaction.RetainedGlobalConflict
				if !errors.As(err, &retained) || retained.Cleanup != nil {
					t.Fatalf("typed retained disposition: %v", err)
				}
			}
			for _, path := range []string{lockPath(home, DefaultProfile), sourcePath(home, DefaultProfile)} {
				_, e := os.Lstat(path)
				if mode == "source-fault" && !os.IsNotExist(e) || mode != "source-fault" && e != nil {
					t.Fatalf("first-use record disposition %s: %v", filepath.Base(path), e)
				}
			}
			if mode != "healthy" {
				for id, root := range homes {
					if nativeSnapshot(t, root) != before[string(id)] {
						t.Fatalf("first-use failure changed native %s", id)
					}
				}
			} else if _, e := os.Lstat(filepath.Join(homes[envregistry.CodexCLI], envmarker.Name)); e != nil {
				t.Fatalf("healthy control missed native continuation: %v", e)
			}
			if err := envauthority.Validate(home); err != nil {
				t.Fatal(err)
			}
			identity, err := envunit.ObserveManager(home)
			if err != nil {
				t.Fatal(err)
			}
			ledger, _, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			if (len(ledger.Entries) == 1) != strings.HasPrefix(mode, "retained") {
				t.Fatal("wrong first-use receipt disposition")
			}
			if strings.HasPrefix(mode, "retained") && (len(ledger.Entries[0].Global.RetainedRecords) != 2 || info.LockHash == "") {
				t.Fatal("first-use retained source/L1 capture incomplete")
			}
			if strings.HasPrefix(mode, "retained") {
				// A fresh operation must not infer takeover from the retained lock.
				if _, e := SyncWithPolicy(home, Policy{}); e == nil || !strings.Contains(e.Error(), "global_retained_conflict") {
					t.Fatalf("plain restart lost retained barrier: %v", e)
				}
				for id, root := range homes {
					if nativeSnapshot(t, root) != before[string(id)] {
						t.Fatalf("restart changed %s", id)
					}
				}
				var rows []EntryResult
				var e error
				if strings.HasPrefix(mode, "retained-use") {
					required := "another-profile"
					if _, e := UseWithPolicy(home, DefaultProfile, "", "", false, Policy{Takeover: true, RequireCurrent: &required, RequireCurrentLocked: true}); e == nil || !strings.Contains(e.Error(), "environments.require_current_profile") {
						t.Fatalf("takeover bypassed machine gate: %v", e)
					}
					for id, root := range homes {
						if nativeSnapshot(t, root) != before[string(id)] {
							t.Fatalf("policy refusal changed %s", id)
						}
					}
					activationHit := false
					if mode == "retained-use-fault" {
						newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
							return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
								if event.Point == transaction.PointAfterInstall && event.LivePath == CurrentFile(home) {
									activationHit = true
									return errors.New("retained activation fault")
								}
								return nil
							}}))...)
						}
					}
					rows, e = UseWithPolicy(home, DefaultProfile, "", "", false, Policy{Takeover: true})
					newOperationEngine = saved
					if mode == "retained-use-fault" {
						if !activationHit || e == nil || !strings.Contains(e.Error(), "retained activation fault") {
							t.Fatalf("activation owner not reached: %t %v", activationHit, e)
						}
						for id, root := range homes {
							if nativeSnapshot(t, root) != before[string(id)] {
								t.Fatalf("activation fault changed %s", id)
							}
						}
						if _, e := os.Lstat(CurrentFile(home)); !os.IsNotExist(e) {
							t.Fatalf("failed activation did not restore absence: %v", e)
						}
						after, _, e := envlifecycle.Read(identity)
						if e != nil || len(after.Entries) != 1 || after.Entries[0].ID() != ledger.Entries[0].ID() {
							t.Fatalf("failed takeover lost receipt: %+v %v", after, e)
						}
						return
					}
					current, ce := Current(home)
					if ce != nil || current != DefaultProfile {
						t.Fatalf("successful use did not activate: %q %v (use=%v)", current, ce, e)
					}
				} else {
					rows, e = SyncWithPolicy(home, Policy{Takeover: true})
				}
				if e != nil || len(rows) == 0 {
					t.Fatalf("explicit first-use takeover did not reach native owner: %+v %v", rows, e)
				}
				for _, row := range rows {
					if !row.OK {
						t.Fatalf("takeover row failed: %+v", row)
					}
				}
				_, hash, e := readLock(home, DefaultProfile)
				if e != nil || hash != info.LockHash {
					t.Fatalf("takeover changed L1: %s %v", hash, e)
				}
				after, _, e := envlifecycle.Read(identity)
				if e != nil || len(after.Entries) != 0 {
					t.Fatalf("takeover did not consume exact receipt: %+v %v", after, e)
				}
				if _, e := os.Lstat(filepath.Join(homes[envregistry.CodexCLI], envmarker.Name)); e != nil {
					t.Fatalf("takeover missed actual native publication: %v", e)
				}
			}
		})
	}
}
