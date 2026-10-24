package envprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// Snapshots below cover bytes, entry inventory/types and symlink targets;
// filesystem timestamps and permission bits are not asserted.

// A full, schema-valid Claude node carrier refuses before either damaged-pin
// or absent-entry repair, for both repair values, without output or changes to
// manager/native state. Per-home neighbors refuse read-only and repair normally.
func TestNodeAuthorityBeforeStoreRepair(t *testing.T) {
	for _, damage := range []string{"damaged", "missing"} {
		for _, node := range []bool{true, false} {
			for _, repair := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/node=%v/repair=%v", damage, node, repair), func(t *testing.T) {
					fx := writeManagedFixture(t, "example")
					req := fx.request(envregistry.ClaudeCode)
					req.Repair = true
					if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
						t.Fatalf("provision: %v", err)
					}
					if node {
						writeNodeAuthorityMarker(t, fx)
					}
					entry := storeBoundaryEntry(t, fx, "mcp")
					pinFile := filepath.Join(entry, "unused", "pin-only.txt")
					if damage == "missing" {
						if err := os.RemoveAll(entry); err != nil {
							t.Fatal(err)
						}
					} else if err := os.WriteFile(pinFile, []byte("damaged\n"), 0600); err != nil {
						t.Fatal(err)
					}
					before := nativeSnapshot(t, fx.home)
					nativeBefore := nativeSnapshot(t, fx.native[envregistry.ClaudeCode])
					recoveries := 0
					req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
					req.Repair = repair
					result, err := Resolve(req)
					if node {
						if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
							t.Errorf("node authority diagnostic: %v", err)
						}
					} else if repair {
						if err != nil || result == nil || len(result.Document) == 0 {
							t.Fatalf("per-home store repair: %v", err)
						}
						restored, err := os.ReadFile(pinFile)
						if err != nil || string(restored) != "not surfaced\n" {
							t.Fatalf("per-home store not restored: %q %v", restored, err)
						}
					} else if err == nil || !strings.HasPrefix(err.Error(), "environment_store_untrusted:") {
						t.Errorf("per-home read-only store refusal: %v", err)
					}
					if node || !repair {
						if result != nil && len(result.Document) != 0 {
							t.Error("store refusal emitted a fragment")
						}
						if recoveries != 0 {
							t.Error("node/store refusal entered recovery")
						}
						if nativeSnapshot(t, fx.home) != before {
							t.Error("authority/store refusal mutated manager state")
						}
					}
					if nativeSnapshot(t, fx.native[envregistry.ClaudeCode]) != nativeBefore {
						t.Error("store resolution changed native state")
					}
				})
			}
		}
	}
}

func writeNodeAuthorityMarker(t *testing.T, fx *managedFixture) {
	t.Helper()
	marker := readManagedMarker(t, fx, envregistry.ClaudeCode)
	marker.Version = envmarker.VersionV7
	marker.HashVersion = 2
	marker.CodexSettings = json.RawMessage(`{}`)
	emptyPass := []envmarker.Passthrough{}
	emptySeeds := []string{}
	marker.Passthrough = &emptyPass
	marker.Seeds = &emptySeeds
	marker.SeedLinks = nil
	marker.CredentialSource = &envmarker.CredentialSource{Source: &envmarker.NodeSource{Mode: "node", SourceRef: "local:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", HarnessChannel: "CLAUDE_CODE_OAUTH_TOKEN"}, SourceGeneration: 1, PolicyGeneration: 1}
	raw, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := envmarker.Parse(raw); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), envmarker.Name), raw, 0600); err != nil {
		t.Fatal(err)
	}
}

// Authority changed at the lock seam is revalidated before real prepared intent
// can recover. The per-home control recovers the same journal and emits a fragment.
// This checks ordering under the manager lock, not exclusion of external writers.
func TestNodeAuthorityBeforeRecoveryUnderLock(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, req, seed := claudeRepairFixture(t, envregistry.FormReferenced)
			target := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
			desired := readClaudeState(t, target)
			if err := os.WriteFile(target, []byte("pending old context"), 0600); err != nil {
				t.Fatal(err)
			}
			manager, err := managerlock.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			staged := filepath.Join(t.TempDir(), "target")
			if err := os.WriteFile(staged, desired, 0600); err != nil {
				t.Fatal(err)
			}
			digest, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil {
				t.Fatal(err)
			}
			_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "node-authority-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: digest}}})
			if err != nil {
				t.Fatal(err)
			}
			if pending, err := engine.PendingRecovery(lock); err != nil || !pending {
				t.Fatalf("pending setup: %v %v", pending, err)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			reached, recoveries := false, 0
			var before string
			nativeBefore := nativeSnapshot(t, fx.native[envregistry.ClaudeCode])
			req.beforeRepairLock = func() {
				reached = true
				if node {
					writeNodeAuthorityMarker(t, fx)
				}
				before = nativeSnapshot(t, fx.home)
			}
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
			result, err := Resolve(req)
			if !reached {
				t.Fatal("lock seam not reached")
			}
			if node {
				if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
					t.Errorf("locked node authority diagnostic: %v", err)
				}
				if result != nil && len(result.Document) != 0 {
					t.Error("locked node refusal emitted fragment")
				}
				if recoveries != 0 {
					t.Error("locked node refusal entered recovery")
				}
				if nativeSnapshot(t, fx.home) != before {
					t.Error("locked authority refusal mutated pending state")
				}
			} else {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("per-home recovery: %v", err)
				}
				if recoveries != 1 || string(readClaudeState(t, target)) != string(desired) {
					t.Fatal("per-home control did not recover real intent")
				}
			}
			if nativeSnapshot(t, fx.native[envregistry.ClaudeCode]) != nativeBefore {
				t.Error("recovery resolution changed native state")
			}
		})
	}
}

// Both store-repair entry points recheck a marker replaced at the lock seam,
// preserve the normative refusal diagnostic, and leave damaged/missing state
// untouched. Per-home repairs for these same entries are controlled above.
func TestNodeAuthorityAtStoreRepairLock(t *testing.T) {
	for _, damage := range []string{"damaged", "missing"} {
		t.Run(damage, func(t *testing.T) {
			fx := writeManagedFixture(t, "example")
			req := fx.request(envregistry.ClaudeCode)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			entry := storeBoundaryEntry(t, fx, "mcp")
			if damage == "missing" {
				if err := os.RemoveAll(entry); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(filepath.Join(entry, "unused", "pin-only.txt"), []byte("damaged\n"), 0600); err != nil {
				t.Fatal(err)
			}
			var before string
			nativeBefore := nativeSnapshot(t, fx.native[envregistry.ClaudeCode])
			req.beforeRepairLock = func() { writeNodeAuthorityMarker(t, fx); before = nativeSnapshot(t, fx.home) }
			recoveries := 0
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
			result, err := Resolve(req)
			if before == "" {
				t.Fatal("store repair lock seam not reached")
			}
			if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
				t.Errorf("store lock authority diagnostic: %v", err)
			}
			if result != nil && len(result.Document) != 0 {
				t.Error("store lock refusal emitted fragment")
			}
			if recoveries != 0 {
				t.Error("store lock refusal entered recovery")
			}
			if nativeSnapshot(t, fx.home) != before || nativeSnapshot(t, fx.native[envregistry.ClaudeCode]) != nativeBefore {
				t.Error("store lock authority refusal mutated state")
			}
		})
	}
}

// If already-authorized durable recovery replaces a per-home marker with node
// authority, repair must stop before rebuilding the damaged store. Recovery is
// reported; the damaged store and native bytes remain unchanged.
func TestRecoveredNodeAuthorityBeforeStorePublication(t *testing.T) {
	fx := writeManagedFixture(t, "example")
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), envmarker.Name)
	original := readClaudeState(t, path)
	writeNodeAuthorityMarker(t, fx)
	desired := readClaudeState(t, path)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	entry := storeBoundaryEntry(t, fx, "mcp")
	if err := os.WriteFile(filepath.Join(entry, "unused", "pin-only.txt"), []byte("damaged\n"), 0600); err != nil {
		t.Fatal(err)
	}
	storeBefore := nativeSnapshot(t, entry)
	nativeBefore := nativeSnapshot(t, fx.native[envregistry.ClaudeCode])
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := transaction.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(t.TempDir(), "marker")
	if err := os.WriteFile(staged, desired, 0600); err != nil {
		t.Fatal(err)
	}
	digest, err := transaction.DigestTarget(transaction.KindBytes, path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "recovered-node-authority", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: path, Kind: transaction.KindBytes, LivePath: path, StagedSource: staged, PreimageDigest: digest}}})
	if err != nil {
		t.Fatal(err)
	}
	if pending, err := engine.PendingRecovery(lock); err != nil || !pending {
		t.Fatalf("pending marker setup: %v %v", pending, err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	recoveries := 0
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
	result, err := Resolve(req)
	if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") || !strings.Contains(err.Error(), "previously pending transaction recovery completed") {
		t.Errorf("recovered node authority diagnostic: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Error("recovered node authority emitted fragment")
	}
	if recoveries != 1 || string(readClaudeState(t, path)) != string(desired) {
		t.Error("durable marker recovery did not complete")
	}
	if nativeSnapshot(t, entry) != storeBefore {
		t.Error("recovered authority refusal mutated store")
	}
	if nativeSnapshot(t, fx.native[envregistry.ClaudeCode]) != nativeBefore {
		t.Error("recovered authority changed native state")
	}
}

// With main's borrowed-operation path, post-validation refusal must leave the
// outer mutation lock held. The per-home control returns that same operation.
// The read seam models an authority change between the two locked validations.
func TestNodeAuthorityKeepsBorrowedRepairLock(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, req, _ := claudeRepairFixture(t, envregistry.FormReferenced)
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = op.close() }()
			req.borrowedOperation = op
			reads := 0
			req.readStateFile = func(path string) (stateread.File, error) {
				reads++
				if node && reads == 2 {
					writeNodeAuthorityMarker(t, fx)
				}
				return stateread.ReadFile(path)
			}
			result, err := req.beginRepairOperation()
			if reads != 2 {
				t.Fatalf("locked validations: %d", reads)
			}
			if node {
				if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") || result != nil {
					t.Errorf("borrowed authority refusal: %v", err)
				}
			} else if err != nil || result != op {
				t.Errorf("per-home borrowed operation: %v", err)
			}
			if err := op.lock.AssertHeld(); err != nil {
				t.Errorf("authority refusal closed the borrowed lock: %v", err)
			}
		})
	}
}

// ApplyMigration refuses a full node carrier before recovering real prepared context
// intent; the per-home control recovers that same intent and succeeds.
func TestMigrationNodeAuthorityBeforeRecovery(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, _, seed := claudeRepairFixture(t, envregistry.FormReferenced)
			target := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
			desired := readClaudeState(t, target)
			if err := os.WriteFile(target, []byte("pending old context"), 0600); err != nil {
				t.Fatal(err)
			}
			manager, err := managerlock.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			staged := filepath.Join(t.TempDir(), "target")
			if err := os.WriteFile(staged, desired, 0600); err != nil {
				t.Fatal(err)
			}
			digest, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil {
				t.Fatal(err)
			}
			_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "node-authority-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: digest}}})
			if err != nil {
				t.Fatal(err)
			}
			if pending, err := engine.PendingRecovery(lock); err != nil || !pending {
				t.Fatalf("pending setup: %v %v", pending, err)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			if node {
				writeNodeAuthorityMarker(t, fx)
			}
			req := fx.migrateRequest()
			req.Profile, req.EnvID = fx.profile, envregistry.ClaudeCode
			plan, err := PlanMigration(req)
			if err != nil {
				t.Fatal(err)
			}
			req.Expect = plan.Hash
			before := nativeSnapshot(t, fx.home)
			nativeBefore := nativeSnapshot(t, fx.native[envregistry.ClaudeCode])
			recoveries := 0
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
			result, err := ApplyMigration(req)
			t.Logf("node=%v recovery_calls=%d refusal=%v", node, recoveries, err)

			if node {
				if err == nil || !strings.Contains(err.Error(), "credential_injection_unavailable:") {
					t.Errorf("locked node authority diagnostic: %v", err)
				}

				if recoveries != 0 {
					t.Error("migration node refusal entered recovery")
				}
				if nativeSnapshot(t, fx.home) != before {
					t.Error("migration authority refusal mutated pending state")
				}
			} else {
				if err != nil || result == nil {
					t.Fatalf("per-home recovery: %v", err)
				}
				if recoveries != 1 || string(readClaudeState(t, target)) != string(desired) {
					t.Fatal("per-home control did not recover real intent")
				}
			}
			if nativeSnapshot(t, fx.native[envregistry.ClaudeCode]) != nativeBefore {
				t.Error("recovery resolution changed native state")
			}
		})
	}
}
