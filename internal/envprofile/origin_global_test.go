package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/managerlock"
)

// All account and native roots are synthetic before invoking the actual global
// owner. Global maintenance must use these records and may never obtain a real
// account home. The fixture lock has a deliberate root weight that resolution
// changes, without fetching any remote Git source or invoking a build.
func recordedGlobalOriginFixture(t *testing.T) (*managedFixture, func(string) (string, error)) {
	t.Helper()
	enableV1WritersForTest(t)
	fx := writeManagedFixture(t, "global-origin-fixture")
	lock, _, e := readLock(fx.home, fx.profile)
	if e != nil {
		t.Fatal(e)
	}
	root, ok := lock.RootMember()
	if !ok {
		t.Fatal("missing root")
	}
	root.Weight = 123
	lock.Members = []contextlock.Member{root}
	fx.lockHash, e = contextlock.Write(lockPath(fx.home, fx.profile), lock)
	if e != nil {
		t.Fatal(e)
	}
	anchor := t.TempDir()
	t.Setenv("CODEX_HOME", fx.native["codex_cli"])
	t.Setenv("CLAUDE_CONFIG_DIR", fx.native["claude_code"])
	t.Setenv("PI_CODING_AGENT_DIR", fx.native["pi"])
	t.Setenv("XDG_CONFIG_HOME", fx.xdg)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	for _, id := range []string{"codex_cli", "pi"} {
		if e = os.WriteFile(filepath.Join(fx.native[id], "auth.json"), nil, 0600); e != nil {
			t.Fatal(e)
		}
	}
	resolver := fx.request("codex_cli").NativeHomeOf
	for _, id := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
		req := fx.request(id)
		req.Repair = true
		req.originAccountLookup = func() (envorigin.Account, error) {
			return envorigin.Account{UID: uint64(os.Geteuid()), RealUID: uint64(os.Getuid()), EffectiveUID: uint64(os.Geteuid()), Home: anchor}, nil
		}
		if result, e := Resolve(req); e != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("synthetic recorded %s provisioning: %v", id, e)
		}
	}
	return fx, resolver
}

// Adopted global handled conflict retains L1/store, with every managed home
// (including inactive adapters), origin capture and native entry at its exact
// old state. A durable receipt keeps this intentionally nonlaunchable after the
// journal is cleaned; the legacy continuation must not run or emit a fragment.
func TestProvisioningOriginGlobalRetainedConflict(t *testing.T) {
	fx, resolver := recordedGlobalOriginFixture(t)
	native := fx.native["codex_cli"]
	// Explicitly create an unrecorded regular entry; WriteFile alone would
	// follow a fixture's existing context alias rather than plant this conflict.
	for _, name := range []string{"AGENTS.md", envmarker.Name} {
		if er := os.Remove(filepath.Join(native, name)); er != nil && !os.IsNotExist(er) {
			t.Fatal(er)
		}
	}
	if e := os.WriteFile(filepath.Join(native, "AGENTS.md"), []byte("unmanaged native fixture"), 0644); e != nil {
		t.Fatal(e)
	}
	scoped, er := ScopedCurrents(fx.home)
	if er != nil {
		t.Fatal(er)
	}
	t.Logf("fixture scoped selections: %v", scoped)
	if er = preflightInPlaceScope(fx.home, fx.profile, "", Policy{}, resolver); er == nil || !strings.Contains(er.Error(), DiagUnmanagedConflict) {
		t.Fatalf("fixture must reach native unmanaged preflight: %v", er)
	}
	beforeManaged := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
	beforeNative := map[string]string{}
	for id, root := range fx.native {
		beforeNative[id] = nativeSnapshot(t, root)
	}
	_, oldHash, e := readLock(fx.home, fx.profile)
	if e != nil {
		t.Fatal(e)
	}
	manager, e := managerlock.New(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	held, e := manager.AcquireHomeOnly(context.Background(), false)
	if e != nil {
		t.Fatal(e)
	}
	defer held.Close()
	calls := 0
	info, e := publishGlobalContinuationFixture(t, fx.home, resolver, held, &calls)
	if e == nil || !strings.Contains(e.Error(), DiagUnmanagedConflict) {
		t.Fatalf("intended in-place handled conflict: %v info=%+v lock=%+v", e, info, info.Lock)
	}
	if calls != 0 {
		t.Fatal("handled conflict ran legacy continuation")
	}
	_, newHash, er := readLock(fx.home, fx.profile)
	if er != nil || newHash == oldHash || info.LockHash != newHash {
		t.Fatalf("handled conflict dropped retained L1: %v old=%s new=%s info=%s", er, oldHash, newHash, info.LockHash)
	}
	if nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != beforeManaged {
		t.Fatal("handled conflict changed mutable managed home preimages")
	}
	for id, root := range fx.native {
		if nativeSnapshot(t, root) != beforeNative[id] {
			t.Fatalf("handled conflict changed native %s preimages", id)
		}
	}
	identity, e := envunit.ObserveManager(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	ledger, _, e := envlifecycle.Read(identity)
	if e != nil {
		t.Fatal(e)
	}
	if len(ledger.Entries) != 1 || ledger.Entries[0].Global == nil {
		t.Fatal("handled conflict lacks durable retained-L1 receipt")
	}
	if e = envauthority.Validate(fx.home); e != nil {
		t.Fatalf("valid retained custody admission: %v", e)
	}
	if e = envauthority.SelectedLifecycle(fx.home, fx.profile); e == nil || !strings.Contains(e.Error(), "global_retained_conflict") {
		t.Fatalf("retained L1 falsely launchable: %v", e)
	}
}

// The healthy global neighbor must reach its real legacy continuation and
// converge every inactive managed adapter. A retained conflict instead survives
// a new engine, refuses plain sync, then explicitly converges the whole group
// while preserving the original A capture even under synthetic B selectors.
func TestProvisioningOriginGlobalContinuation(t *testing.T) {
	for _, retained := range []bool{false, true} {
		t.Run(fmt.Sprintf("retained=%t", retained), func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			beforeOrigin := map[string]*envorigin.Record{}
			for _, id := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
				raw, e := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name))
				if e != nil {
					t.Fatal(e)
				}
				beforeOrigin[id], e = envorigin.ParseCaptured(raw, envorigin.Binding{Profile: fx.profile, LockSHA256: fx.lockHash, ManagedTarget: ManagedHomeDir(fx.home, fx.profile, id)}, id, fx.home)
				if e != nil {
					t.Fatal(e)
				}
			}
			if retained {
				plantGlobalUnmanaged(t, fx)
			}
			manager, e := managerlock.New(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			held, e := manager.AcquireHomeOnly(context.Background(), false)
			if e != nil {
				t.Fatal(e)
			}
			calls := 0
			info, e := publishGlobalContinuationFixture(t, fx.home, resolver, held, &calls)
			if er := held.Close(); er != nil {
				t.Fatal(er)
			}
			if retained {
				var conflict *transaction.RetainedGlobalConflict
				if !errors.As(e, &conflict) || calls != 0 {
					t.Fatalf("typed retained cut: %v calls=%d", e, calls)
				}
				engine, e := transaction.New(fx.home)
				if e != nil {
					t.Fatal(e)
				}
				held, e = manager.AcquireHomeOnly(context.Background(), false)
				if e != nil {
					t.Fatal(e)
				}
				if e = engine.Recover(held); e != nil {
					t.Fatal(e)
				}
				if e = held.Close(); e != nil {
					t.Fatal(e)
				}
				baseline := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
				if _, e = SyncWithPolicy(fx.home, Policy{}); e == nil || !strings.Contains(e.Error(), "global_retained_conflict") {
					t.Fatalf("plain restart/sync implicitly took over: %v", e)
				}
				if nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != baseline {
					t.Fatal("plain sync changed retained homes")
				}
				t.Setenv("CODEX_HOME", t.TempDir())
				t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
				t.Setenv("PI_CODING_AGENT_DIR", t.TempDir())
				t.Setenv("XDG_CONFIG_HOME", t.TempDir())
				if _, e = SyncWithPolicy(fx.home, Policy{Takeover: true}); e != nil {
					t.Fatalf("explicit whole-group takeover: %v", e)
				}
			} else if e != nil || calls != 1 {
				t.Fatalf("healthy continuation not reached: %v calls=%d", e, calls)
			}
			_, hash, e := readLock(fx.home, fx.profile)
			if e != nil || hash != info.LockHash || hash == fx.lockHash {
				t.Fatalf("L1 changed/lost: %v hash=%s info=%s", e, hash, info.LockHash)
			}
			for id, old := range beforeOrigin {
				raw, e := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name))
				if e != nil {
					t.Fatal(e)
				}
				record, e := envorigin.Parse(raw, envorigin.Binding{Profile: fx.profile, LockSHA256: hash, ManagedTarget: ManagedHomeDir(fx.home, fx.profile, id)}, id, fx.home)
				if e != nil {
					t.Fatalf("inactive %s not rebound: %v", id, e)
				}
				a, _ := json.Marshal(old.Capture)
				b, _ := json.Marshal(record.Capture)
				if !bytes.Equal(a, b) || old.NativeRoot != record.NativeRoot {
					t.Fatalf("%s recaptured B", id)
				}
			}
			identity, e := envunit.ObserveManager(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			ledger, _, e := envlifecycle.Read(identity)
			if e != nil || len(ledger.Entries) != 0 {
				t.Fatalf("completed group retains pending receipt: %v", e)
			}
			if e = envauthority.Validate(fx.home); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func plantGlobalUnmanaged(t *testing.T, fx *managedFixture) {
	t.Helper()
	root := fx.native["codex_cli"]
	for _, name := range []string{"AGENTS.md", envmarker.Name} {
		if e := os.Remove(filepath.Join(root, name)); e != nil && !os.IsNotExist(e) {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("unmanaged native fixture"), 0644); e != nil {
		t.Fatal(e)
	}
}

// Returned faults compensate; modeled process stops lose all callbacks and
// recover from the real journal. L1 is retained only after typed post-lock
// conflict confirmation. These are in-process stop models, not power-loss proof.
func TestProvisioningOriginGlobalRecovery(t *testing.T) {
	for _, row := range []struct {
		name            string
		point           transaction.Point
		crash, retained bool
	}{
		{"lock-returned", transaction.PointAfterInstall, false, false},
		{"lock-stop", transaction.PointAfterInstall, true, false},
		{"retained-stop", transaction.PointGlobalRetained, true, true},
		{"cleanup-stop", transaction.PointBeforeCleanup, true, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			plantGlobalUnmanaged(t, fx)
			before := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
			native := map[string]string{}
			for id, root := range fx.native {
				native[id] = nativeSnapshot(t, root)
			}
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit := false
			newOperationEngine = func(home string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(home, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
					match := e.Point == row.point
					if row.point == transaction.PointAfterInstall {
						match = match && e.LivePath == lockPath(fx.home, fx.profile)
					}
					if match && !hit {
						hit = true
						if row.crash {
							panic("global-recovery-stop")
						}
						return errors.New("global-lock-returned")
					}
					return nil
				}}))...)
			}
			manager, e := managerlock.New(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			held, e := manager.AcquireHomeOnly(context.Background(), false)
			if e != nil {
				t.Fatal(e)
			}
			calls := 0
			func() {
				defer func() {
					if value := recover(); value != nil {
						if value != "global-recovery-stop" {
							panic(value)
						}
					}
				}()
				_, e = publishGlobalContinuationFixture(t, fx.home, resolver, held, &calls)
			}()
			if er := held.Close(); er != nil {
				t.Fatal(er)
			}
			newOperationEngine = saved
			if !hit || calls != 0 {
				t.Fatalf("fault did not reach actual prefix/cut: hit=%t calls=%d err=%v", hit, calls, e)
			}
			held, e = manager.AcquireHomeOnly(context.Background(), false)
			if e != nil {
				t.Fatal(e)
			}
			engine, e := saved(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			recovery := engine.Recover(held)
			if er := held.Close(); er != nil {
				t.Fatal(er)
			}
			// A retained disposition may report its committed conflict once while
			// finishing receipt publication; it must have completed all cleanup.
			var conflict *transaction.RetainedGlobalConflict
			if recovery != nil && (!row.retained || !errors.As(recovery, &conflict) || conflict.Cleanup != nil) {
				t.Fatalf("fresh journal recovery: %v", recovery)
			}
			_, hash, e := readLock(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			if (hash != fx.lockHash) != row.retained {
				t.Fatalf("wrong retained/compensated lock: %s", hash)
			}
			if nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != before {
				t.Fatal("failure/restart changed managed preimages")
			}
			for id, root := range fx.native {
				if nativeSnapshot(t, root) != native[id] {
					t.Fatalf("failure/restart changed native %s", id)
				}
			}
			if e = envauthority.Validate(fx.home); e != nil {
				t.Fatal(e)
			}
			entries, e := os.ReadDir(filepath.Join(fx.home, "state", "transactions", "v1"))
			if e != nil {
				t.Fatal(e)
			}
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".json") || strings.HasSuffix(entry.Name(), ".delete") {
					t.Fatal("completed recovery left a journal")
				}
			}
		})
	}
}

// This fixture joins a real env-file replacement to the real profile publisher.
// Its counter is incremented only after verifying the actual committed bytes;
// refusal and saved precommit phases must leave the old consumer intact.
func publishGlobalContinuationFixture(t *testing.T, home string, resolver func(string) (string, error), held transaction.HomeLock, calls *int) (Info, error) {
	t.Helper()
	live := filepath.Join(home, "global", "env.sh")
	old, oldErr := os.ReadFile(live)
	if oldErr != nil && !os.IsNotExist(oldErr) {
		t.Fatal(oldErr)
	}
	stage := filepath.Join(t.TempDir(), "env.sh")
	if err := os.WriteFile(stage, []byte("new continuation"), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := transaction.DigestTarget(transaction.KindBytes, live)
	if err != nil {
		t.Fatal(err)
	}
	plan := transaction.Plan{Legacy: &envunit.LegacyPublication{UserHome: t.TempDir(), Entries: []envunit.LegacyTarget{{Class: "50-env-file", Path: live}}}, Targets: []transaction.Target{{Class: "50-env-file", Identifier: live, Kind: transaction.KindBytes, LivePath: live, StagedSource: stage, PreimageDigest: before}}}
	engine, err := transaction.New(home)
	if err != nil {
		t.Fatal(err)
	}
	plan.Directories, err = engine.PublicationDirectories(live)
	if err != nil {
		t.Fatal(err)
	}
	info, err := PublishGlobalSkillPlan(home, Policy{}, envregistry.DefaultMachineConfig(), nil, resolver, held, plan)
	raw, readErr := os.ReadFile(live)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatal(readErr)
	}
	if err == nil {
		if string(raw) != "new continuation" {
			t.Fatal("healthy global owner omitted consumer publication")
		}
		*calls++
	} else if !errors.Is(err, transaction.ErrImplementationCorruption) && (string(raw) != string(old) || os.IsNotExist(readErr) != os.IsNotExist(oldErr)) {
		t.Fatal("refusal changed global consumer preimage")
	}
	return info, err
}
