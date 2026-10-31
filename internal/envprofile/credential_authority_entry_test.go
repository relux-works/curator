package envprofile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Every listed production entry refuses a schema-valid node home before any
// observable recovery or mutation. A real prepared journal remains byte-identical
// on refusal; entries marked as recovering complete that intent for per-home
// controls. Other entries intentionally leave generic intent pending. The shared
// constructor observes recovery for public entries without transaction options.
// Snapshots have the
// byte/type/link-target limits stated in credential_source_repair_test.go.
func TestCredentialAuthorityEntryPoints(t *testing.T) {
	type invocation struct {
		err       error
		callbacks int
		counted   bool
	}
	type entry struct {
		name     string
		recovers bool
		run      func(*testing.T, *managedFixture, string) invocation
	}
	entries := []entry{
		{"resolve-readonly", false, func(t *testing.T, fx *managedFixture, _ string) invocation {
			req := fx.request(envregistry.ClaudeCode)
			req.Machine.Forms = map[string]string{envregistry.ClaudeCode: envregistry.FormReferenced}
			r, e := Resolve(req)
			if r != nil && len(r.Document) > 0 && e != nil {
				t.Error("refusal emitted fragment")
			}
			return invocation{err: e}
		}},
		{"resolve-repair", true, func(t *testing.T, fx *managedFixture, _ string) invocation {
			req := fx.request(envregistry.ClaudeCode)
			req.Repair = true
			req.Machine.Forms = map[string]string{envregistry.ClaudeCode: envregistry.FormReferenced}
			n := 0
			req.transactionOptions = recoveryCounter(&n)
			r, e := Resolve(req)
			if r != nil && len(r.Document) > 0 && e != nil {
				t.Error("refusal emitted fragment")
			}
			return invocation{e, n, true}
		}},
		{"migration-plan", false, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			req := fx.migrateRequest()
			req.Profile, req.EnvID = fx.profile, envregistry.ClaudeCode
			r, e := PlanMigration(req)
			if e == nil && len(r.Conflicts()) > 0 {
				e = fmt.Errorf("%s", r.Conflicts()[0].Detail)
			}
			return invocation{err: e}
		}},
		{"migration-apply", true, func(t *testing.T, fx *managedFixture, _ string) invocation {
			req := fx.migrateRequest()
			req.Profile, req.EnvID = fx.profile, envregistry.ClaudeCode
			r, e := PlanMigration(req)
			if e != nil {
				t.Fatal(e)
			}
			req.Expect = r.Hash
			n := 0
			req.transactionOptions = recoveryCounter(&n)
			_, e = ApplyMigration(req)
			return invocation{e, n, true}
		}},
		{"operation-recovery", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			n := 0
			op, e := beginOperation(fx.home, recoveryCounter(&n)...)
			if op != nil {
				_ = op.close()
			}
			return invocation{e, n, true}
		}},
		{"transaction-recovery", true, func(t *testing.T, fx *managedFixture, _ string) invocation {
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			n := 0
			engine, e := transaction.New(fx.home, recoveryCounter(&n)...)
			if e != nil {
				t.Fatal(e)
			}
			e = engine.Recover(lock)
			return invocation{e, n, true}
		}},
		{"transaction-commit", true, func(t *testing.T, fx *managedFixture, id string) invocation {
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			engine, e := transaction.New(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			return invocation{err: engine.Commit(lock, id)}
		}},
		{"transaction-prepare", false, func(t *testing.T, fx *managedFixture, _ string) invocation {
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			engine, e := transaction.New(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			staged := filepath.Join(t.TempDir(), "new")
			if e = os.WriteFile(staged, []byte("next"), 0600); e != nil {
				t.Fatal(e)
			}
			target := filepath.Join(fx.home, "state", "next-target")
			_, e = engine.Prepare(lock, transaction.Plan{TransactionID: "next-authority", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: transaction.DigestAbsent}}})
			return invocation{err: e}
		}},
		{"migration-journal-recovery", false, func(t *testing.T, fx *managedFixture, _ string) invocation {
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			engine, e := transaction.New(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			_, e = recoverMigrationJournal(fx.home, &operation{lock: lock, engine: engine}, nil, os.Symlink, os.Rename)
			return invocation{err: e}
		}},
		{"borrowed-global", false, func(t *testing.T, fx *managedFixture, _ string) invocation {
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			calls := 0
			_, e := GlobalSkillTransaction(fx.home, Policy{}, envregistry.DefaultMachineConfig(), nil, fx.request(envregistry.ClaudeCode).NativeHomeOf, lock, func() error { calls++; return nil })
			if e == nil && calls != 1 {
				t.Error("per-home global continuation did not run")
			}
			if e != nil && calls != 0 {
				t.Error("refusal ran global continuation")
			}
			if err := lock.AssertHeld(); err != nil {
				t.Error("borrowed global refusal closed caller lock")
			}
			return invocation{err: e}
		}},

		{"profile-list", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, e := ListWithPolicy(fx.home, Policy{})
			return invocation{err: e}
		}},
		{"ensure-default", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			return invocation{err: EnsureDefaultWithPolicy(fx.home, Policy{})}
		}},
		{"profile-install", true, func(t *testing.T, fx *managedFixture, _ string) invocation {
			s, e := readSource(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			_, _, _, e = Install(fx.home, InstallOptions{Operand: s.Path, As: "second"})
			return invocation{err: e}
		}},
		{"profile-update", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, _, e := UpdateWithPolicy(fx.home, fx.profile, Policy{})
			return invocation{err: e}
		}},
		{"profile-remove", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			return invocation{err: Remove(fx.home, fx.profile, true)}
		}},
		{"profile-use", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, e := UseWithPolicy(fx.home, fx.profile, envregistry.ClaudeCode, "", false, Policy{})
			return invocation{err: e}
		}},
		{"profile-sync", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, e := SyncWithPolicy(fx.home, Policy{})
			return invocation{err: e}
		}},
		{"managed-sync", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, e := SyncManagedHomesWithPolicy(fx.home, Policy{}, envregistry.DefaultMachineConfig())
			return invocation{err: e}
		}},
		{"profile-import", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, _, _, e := Import(fx.home, ImportOptions{As: "imported", AllowLossy: true, NativeHomeOf: fx.request(envregistry.ClaudeCode).NativeHomeOf, GlobalSkillsOf: func() (string, error) { return filepath.Join(fx.xdg, "skills"), nil }})
			return invocation{err: e}
		}},
		{"unmanage", true, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			_, e := Unmanage(UnmanageRequest{Home: fx.home, EnvID: envregistry.ClaudeCode, NativeHomeOf: fx.request(envregistry.ClaudeCode).NativeHomeOf})
			return invocation{err: e}
		}},
		{"set-current", false, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			return invocation{err: SetCurrent(fx.home, "second")}
		}},
		{"set-scoped", false, func(_ *testing.T, fx *managedFixture, _ string) invocation {
			return invocation{err: SetScoped(fx.home, "env:claude_code", "second", false)}
		}},
	}
	for _, entry := range entries {
		for _, node := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/node=%v", entry.name, node), func(t *testing.T) {
				// Provision HOME-derived settings against the same isolated home used by the entry.
				t.Setenv("HOME", t.TempDir())
				t.Setenv("USERPROFILE", os.Getenv("HOME"))
				fx, _, seed := claudeRepairFixture(t, envregistry.FormReferenced)
				isolateAuthorityNativeHomes(t, fx)
				if err := EnsureDefaultWithPolicy(fx.home, Policy{}); err != nil {
					t.Fatal(err)
				}
				if entry.name == "profile-remove" || entry.name == "borrowed-global" {
					if err := SetCurrent(fx.home, DefaultProfile); err != nil {
						t.Fatal(err)
					}
				}
				if entry.name == "profile-import" {
					if err := os.WriteFile(filepath.Join(fx.native[envregistry.ClaudeCode], "CLAUDE.md"), []byte("# imported fixture\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				// A stale context ensures Resolve --repair actually enters the operation.
				if entry.name == "resolve-repair" {
					if err := os.Remove(filepath.Join(filepath.Dir(seed), "CLAUDE.md")); err != nil {
						t.Fatal(err)
					}
				}
				target := filepath.Join(fx.home, "state", "authority-witness")
				id := "authority-pending"
				if entry.name == "migration-journal-recovery" {
					if err := writeMigrationJournal(fx.home, &migrationJournal{Version: 1, Plan: strings.Repeat("a", 64), Complete: true}); err != nil {
						t.Fatal(err)
					}
				}
				prepareAuthorityIntent(t, fx.home, id, target, []byte("recovered intent\n"))
				if node {
					writeNodeAuthorityMarker(t, fx)
				}
				before := nativeSnapshot(t, fx.home)
				nativeBefore := map[string]string{}
				for env, path := range fx.native {
					nativeBefore[env] = nativeSnapshot(t, path)
				}
				sharedCalls := 0
				newEngine := newOperationEngine
				newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
					if len(options) == 0 {
						options = recoveryCounter(&sharedCalls)
					}
					return newEngine(home, options...)
				}
				t.Cleanup(func() { newOperationEngine = newEngine })
				result := entry.run(t, fx, id)
				if node {
					if result.err == nil || !strings.Contains(result.err.Error(), "credential_injection_unavailable:") {
						t.Errorf("entry authority diagnostic: %v", result.err)
					}
					if sharedCalls != 0 || (result.counted && result.callbacks != 0) {
						t.Errorf("entry entered recovery: direct=%d shared=%d", result.callbacks, sharedCalls)
					}
					if nativeSnapshot(t, fx.home) != before {
						t.Error("entry authority refusal mutated manager state")
					}
				} else {
					if result.err != nil {
						t.Fatalf("per-home positive: %v", result.err)
					}
					want := "pending old intent\n"
					if entry.recovers {
						want = "recovered intent\n"
					}
					got, e := os.ReadFile(target)
					if e != nil || string(got) != want {
						t.Fatalf("per-home recovery intent: %q %v; want %q", got, e, want)
					}
					if entry.recovers && !result.counted && entry.name != "transaction-commit" && sharedCalls == 0 {
						t.Error("per-home public entry did not enter observed recovery")
					}
					if result.counted && result.callbacks != 1 {
						t.Errorf("per-home recovery calls: %d", result.callbacks)
					}
				}
				for env, path := range fx.native {
					if node && nativeSnapshot(t, path) != nativeBefore[env] {
						t.Error("entry authority refusal mutated native state")
					}
				}
			})
		}
	}
}

func isolateAuthorityNativeHomes(t *testing.T, fx *managedFixture) {
	t.Helper()
	for _, adapter := range envregistry.Registry {
		path := fx.native[adapter.ID]
		if adapter.ID == envregistry.OpenCode {
			path = fx.xdg
		}
		if path == "" {
			path = t.TempDir()
		}
		t.Setenv(adapter.EnvVar, path)
	}
}

func recoveryCounter(n *int) []transaction.Option {
	return []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { *n++ }})}
}

func authorityTestLock(t *testing.T, home string) *managerlock.HomeLock {
	t.Helper()
	manager, e := managerlock.New(home)
	if e != nil {
		t.Fatal(e)
	}
	lock, e := manager.AcquireHomeOnly(context.Background(), false)
	if e != nil {
		t.Fatal(e)
	}
	return lock
}

func prepareAuthorityIntent(t *testing.T, home, id, target string, desired []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(target); os.IsNotExist(err) {
		if err := os.WriteFile(target, []byte("pending old intent\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	lock := authorityTestLock(t, home)
	defer func() { _ = lock.Close() }()
	engine, e := transaction.New(home)
	if e != nil {
		t.Fatal(e)
	}
	staged := filepath.Join(t.TempDir(), "intent")
	if e = os.WriteFile(staged, desired, 0600); e != nil {
		t.Fatal(e)
	}
	digest, e := transaction.DigestTarget(transaction.KindBytes, target)
	if e != nil {
		t.Fatal(e)
	}
	_, e = engine.Prepare(lock, transaction.Plan{TransactionID: id, ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: digest}}})
	if e != nil {
		t.Fatal(e)
	}
	pending, e := engine.PendingRecovery(lock)
	if e != nil || !pending {
		t.Fatalf("real pending setup: %v %v", pending, e)
	}
}

// The shared operation rechecks recovered authority, reports the recovery that
// already completed, and returns no operation that could publish more state.
func TestOperationRecoveredNodeAuthority(t *testing.T) {
	fx, _, _ := claudeRepairFixture(t, envregistry.FormReferenced)
	markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), envmarker.Name)
	prior := readClaudeState(t, markerPath)
	writeNodeAuthorityMarker(t, fx)
	node := readClaudeState(t, markerPath)
	if err := os.WriteFile(markerPath, prior, 0600); err != nil {
		t.Fatal(err)
	}
	prepareAuthorityIntent(t, fx.home, "authority-recovery", markerPath, node)
	n := 0
	op, err := beginOperation(fx.home, recoveryCounter(&n)...)
	if op != nil {
		_ = op.close()
		t.Error("recovered node authority returned a writable operation")
	}
	if err == nil || !strings.Contains(err.Error(), "credential_injection_unavailable:") || !strings.Contains(err.Error(), "previously pending transaction recovery completed") {
		t.Fatalf("recovered authority diagnostic: %v", err)
	}
	if n != 1 || string(readClaudeState(t, markerPath)) != string(node) {
		t.Fatal("authorized recovery did not publish the node marker")
	}
}

// Node authority refuses before waiting for a busy mutation lock. The per-home
// neighbor reaches the real contention refusal and succeeds after release.
// The 25ms correctness deadline applies only while the fixture holds the lock;
// the released positive uses the unchanged production lock budget.
// Full snapshots, including lock bytes, bracket acquisition/release because
// Windows forbids reading the file while its exclusive byte-range lock is held.
func TestOperationAuthorityBeforeLock(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, _, _ := claudeRepairFixture(t, envregistry.FormReferenced)
			lock := authorityTestLock(t, fx.home)
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			priorTimeout := lockTimeout
			lockTimeout = 25 * time.Millisecond
			defer func() { lockTimeout = priorTimeout }()
			if node {
				writeNodeAuthorityMarker(t, fx)
			}
			before := nativeSnapshot(t, fx.home)
			lock = authorityTestLock(t, fx.home)
			n := 0
			op, err := beginOperation(fx.home, recoveryCounter(&n)...)
			if op != nil {
				_ = op.close()
				t.Fatal("busy lock returned an operation")
			}
			want := "environment_lock_unavailable:"
			if node {
				want = "credential_injection_unavailable:"
			}
			if err == nil || !strings.HasPrefix(err.Error(), want) {
				t.Fatalf("node operation diagnostic: %v; want %s", err, want)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			if n != 0 || nativeSnapshot(t, fx.home) != before {
				t.Fatal("busy authority refusal mutated or entered recovery")
			}
			if !node {
				lockTimeout = priorTimeout
				op, err = beginOperation(fx.home)
				if err != nil {
					t.Fatalf("released per-home positive: %v", err)
				}
				_ = op.close()
			}
		})
	}
}

// Recovery admitted under per-home authority may complete its durable marker
// intent. If that intent introduces node authority, no subsequent journal may
// run. The matching per-home marker intent permits both journals to complete.
func TestTransactionRecoveredAuthorityStopsNextJournal(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, _, _ := claudeRepairFixture(t, envregistry.FormReferenced)
			markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), envmarker.Name)
			prior := readClaudeState(t, markerPath)
			desired := prior
			if node {
				writeNodeAuthorityMarker(t, fx)
				desired = readClaudeState(t, markerPath)
				if err := os.WriteFile(markerPath, prior, 0600); err != nil {
					t.Fatal(err)
				}
			}
			prepareAuthorityIntent(t, fx.home, "authority-a", markerPath, desired)
			target := filepath.Join(fx.home, "state", "authority-witness")
			prepareAuthorityIntent(t, fx.home, "authority-z", target, []byte("recovered intent\n"))
			lock := authorityTestLock(t, fx.home)
			defer func() { _ = lock.Close() }()
			n := 0
			engine, err := transaction.New(fx.home, recoveryCounter(&n)...)
			if err != nil {
				t.Fatal(err)
			}
			recovered, err := engine.RecoverWithOutcome(lock)
			if !recovered || n != 1 {
				t.Fatalf("real recovery outcome: %v %d %v", recovered, n, err)
			}
			if node {
				if err == nil || !strings.Contains(err.Error(), "credential_injection_unavailable:") || !strings.Contains(err.Error(), "previously pending transaction recovery completed") {
					t.Errorf("post-recovery authority: %v", err)
				}
				if string(readClaudeState(t, target)) != "pending old intent\n" {
					t.Error("recovered authority ran a later journal")
				}
				if pending, err := engine.PendingRecovery(lock); err != nil || !pending {
					t.Fatalf("later intent was not retained: %v %v", pending, err)
				}
			} else if err != nil || string(readClaudeState(t, target)) != "recovered intent\n" {
				t.Fatalf("per-home multi-journal positive: %v", err)
			}
			if string(readClaudeState(t, markerPath)) != string(desired) {
				t.Fatal("first durable marker intent did not complete")
			}
		})
	}
}

// A marker changed by locked validation is checked again by the transaction
// consumer before its recovery callback or real prepared intent can run.
// The expected snapshot includes that intentional marker change and is captured
// before acquisition; the actual snapshot follows release and omits no files.
func TestAuthorityChangeBeforeTransactionRecovery(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, _, _ := claudeRepairFixture(t, envregistry.FormReferenced)
			target := filepath.Join(fx.home, "state", "authority-witness")
			prepareAuthorityIntent(t, fx.home, "authority-pending", target, []byte("recovered intent\n"))
			markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), envmarker.Name)
			prior := readClaudeState(t, markerPath)
			if node {
				writeNodeAuthorityMarker(t, fx)
			}
			before := nativeSnapshot(t, fx.home)
			if node {
				if err := os.WriteFile(markerPath, prior, 0600); err != nil {
					t.Fatal(err)
				}
			}
			n := 0
			op, err := beginOperationValidated(fx.home, func() error {
				if node {
					writeNodeAuthorityMarker(t, fx)
				}
				return nil
			}, recoveryCounter(&n)...)
			if op != nil {
				if closeErr := op.close(); closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			if node {
				if err == nil || !strings.Contains(err.Error(), "credential_injection_unavailable:") || op != nil {
					t.Errorf("locked transaction authority: %v", err)
				}
				if n != 0 || nativeSnapshot(t, fx.home) != before {
					t.Error("locked transaction authority entered recovery or mutated")
				}
			} else if err != nil || n != 1 || string(readClaudeState(t, target)) != "recovered intent\n" {
				t.Fatalf("per-home locked recovery: %v; calls=%d", err, n)
			}
		})
	}
}

// Capture never recovers. It refuses node authority before pending-journal
// inventory; a valid per-home capture emits bytes, while a pending per-home
// capture keeps the existing explicit-recovery-required refusal.
func TestCaptureCredentialAuthority(t *testing.T) {
	for _, pending := range []bool{false, true} {
		for _, node := range []bool{true, false} {
			t.Run(fmt.Sprintf("pending=%v/node=%v", pending, node), func(t *testing.T) {
				fx, req := captureFixture(t)
				// The shared gate covers all adapters even when capture selects Codex.
				clause := fx.request(envregistry.ClaudeCode)
				clause.Repair = true
				if _, err := Resolve(clause); err != nil {
					t.Fatal(err)
				}
				if pending {
					prepareAuthorityIntent(t, fx.home, "capture-pending", filepath.Join(fx.home, "state", "authority-witness"), []byte("recovered intent\n"))
				}
				if node {
					writeNodeAuthorityMarker(t, fx)
				}
				before := nativeSnapshot(t, fx.home)
				n := 0
				req.Resolve.transactionOptions = recoveryCounter(&n)
				result, err := CaptureMaterialize(req)
				if node || pending {
					want := "credential_injection_unavailable:"
					if !node {
						want = DiagMaterializeRecoveryRequired
					}
					if err == nil || !strings.Contains(err.Error(), want) || result != nil {
						t.Fatalf("capture authority diagnostic: %v", err)
					}
				} else if err != nil || result == nil {
					t.Fatalf("per-home capture positive: %v", err)
				}
				if n != 0 || nativeSnapshot(t, fx.home) != before {
					t.Fatal("capture changed or recovered manager state")
				}
			})
		}
	}
}
