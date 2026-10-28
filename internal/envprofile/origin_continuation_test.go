package envprofile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/scopes"
)

type originSweep struct{ reached int }

func (s *originSweep) Sweep(_ buildcache.SweepRequest, lock buildcache.HomeLock) (buildcache.SweepResult, error) {
	if err := lock.AssertHeld(); err != nil {
		return buildcache.SweepResult{}, err
	}
	s.reached++
	return buildcache.SweepResult{}, nil
}

type originLeaseAuthority struct {
	home         string
	profile      string
	homeRef      string
	observations int
	released     bool
}

func (a *originLeaseAuthority) Resolve(_ context.Context, held managerlock.Held, ref, sha string) (projectlease.Composition, error) {
	if err := held.Validate(a.home); err != nil {
		return projectlease.Composition{}, err
	}
	homeRef := a.homeRef
	if homeRef == "" {
		homeRef = "home:fixture"
	}
	return projectlease.Composition{Ref: ref, SHA256: sha, CheckoutID: "fixture-checkout", Profile: a.profile, Environment: "codex_cli", HomeRef: homeRef, ConcurrentState: true, BuildKeys: []string{}, RuntimeEntries: []string{}, Generated: map[string]projectlease.Surface{}}, nil
}
func (a *originLeaseAuthority) Authorize(context.Context, managerlock.Held, projectlease.Composition, string) error {
	return nil
}
func (a *originLeaseAuthority) Observe(_ context.Context, ref string, g int64) (projectlease.Evidence, error) {
	a.observations++
	return projectlease.Evidence{SessionRef: ref, Generation: g, OwnerRef: "principal:fixture", NoChildStarted: true, OwnershipReconciled: a.released}, nil
}

// These call real GC and lease owners after final removal-journal cleanup.
// Q reaches a real sweep/binding publication; orphan substitution refuses before
// the same continuation, and a composition naming retained P cannot reserve a
// session. Injected composition/session authorities are fixture transport only.
func TestProvisioningOriginContinuationOwners(t *testing.T) {
	for _, owner := range []string{"gc", "lease", "lease-retained", "status"} {
		for _, invalid := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/invalid=%t", owner, invalid), func(t *testing.T) {
				fx, _, _, q, before := removedOriginFixture(t)
				target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
				if invalid {
					if err := os.Remove(filepath.Join(target, envmarker.Name)); err != nil {
						t.Fatal(err)
					}
					before = nativeSnapshot(t, target)
				}
				runtime := filepath.Join(fx.home, "runtime", "unused-fixture", "revision")
				if err := os.MkdirAll(runtime, 0700); err != nil {
					t.Fatal(err)
				}
				sweep := &originSweep{}
				authority := &originLeaseAuthority{profile: q, home: fx.home}
				if owner == "lease-retained" {
					authority.profile = fx.profile
				}
				leases, err := projectlease.New(fx.home, authority, authority)
				if err != nil {
					t.Fatal(err)
				}
				switch owner {
				case "gc":
					lm, e := managerlock.New(fx.home)
					if e != nil {
						t.Fatal(e)
					}
					lock, e := lm.AcquireHomeOnly(context.Background(), false)
					if e != nil {
						t.Fatal(e)
					}
					_, err = scopes.Collect(scopes.MaintenanceRequest{Home: fx.home, Lock: lock, Cache: sweep})
					_ = lock.Close()
				case "status":
					_, err = leases.Status(context.Background())
				default:
					_, err = leases.Acquire(context.Background(), projectlease.Request{LeaseID: "fixture-lease", ContextRef: "composition:fixture", ContextSHA256: strings.Repeat("a", 64), OwnerRef: "principal:fixture", SessionRef: "session:fixture", Generation: 1})
				}
				refuse := invalid || owner == "lease-retained"
				if refuse {
					if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
						t.Fatalf("intended continuation refusal: %v", err)
					}
					if sweep.reached != 0 || authority.observations != 0 {
						t.Fatalf("forbidden continuation reached: sweep=%d sessions=%d", sweep.reached, authority.observations)
					}
					if _, e := os.Stat(runtime); e != nil {
						t.Fatal("refusal swept runtime fixture")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if owner == "gc" {
						if sweep.reached != 1 {
							t.Fatal("collector did not reach sweep")
						}
						if _, e := os.Stat(runtime); !os.IsNotExist(e) {
							t.Fatal("actual runtime collector did not sweep")
						}
					}
					if owner == "lease" {
						rows, e := leases.Status(context.Background())
						if e != nil || len(rows) != 1 || authority.observations == 0 {
							t.Fatalf("actual lease reservation missing: %v %+v", e, rows)
						}
					}
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("continuation changed retained custody")
				}
			})
		}
	}
}

// A live/uncertain lease forbids removal without consulting session liveness.
// An explicitly reconciled released lease loses its live reference in the SAME
// removal unit/receipt; immutable binding history and retained home bytes stay.
func TestProvisioningOriginRemoveLeaseReferences(t *testing.T) {
	for _, released := range []bool{false, true} {
		t.Run(fmt.Sprintf("released=%t", released), func(t *testing.T) {
			fx, _, _, _, before := prepareRemovalOriginFixture(t)
			a := &originLeaseAuthority{profile: fx.profile, home: fx.home}
			m, e := projectlease.New(fx.home, a, a)
			if e != nil {
				t.Fatal(e)
			}
			_, e = m.Acquire(context.Background(), projectlease.Request{LeaseID: "removal-lease", ContextRef: "composition:removal", ContextSHA256: strings.Repeat("b", 64), OwnerRef: "principal:fixture", SessionRef: "session:removal", Generation: 1})
			if e != nil {
				t.Fatal(e)
			}
			if released {
				a.released = true
				if e = m.Reconcile(context.Background(), "removal-lease"); e != nil {
					t.Fatal(e)
				}
			}
			observed := a.observations
			e = Remove(fx.home, fx.profile, false)
			if !released {
				if e == nil || !strings.Contains(e.Error(), projectlease.Busy) {
					t.Fatalf("intended active-lease refusal: %v", e)
				}
				if _, er := os.Stat(sourcePath(fx.home, fx.profile)); er != nil {
					t.Fatal("refusal removed installed profile")
				}
			} else {
				if e != nil {
					t.Fatal(e)
				}
				rows, er := m.Status(context.Background())
				if er != nil || len(rows) != 0 {
					t.Fatalf("removed lease still a live ledger reference: %v %+v", er, rows)
				}
			}
			if a.observations != observed {
				t.Fatal("removal inferred permission from session observation")
			}
			if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
				t.Fatal("removal changed retained data")
			}
		})
	}
}

// Faults at reference publication and committed cleanup preserve the atomic
// remove/receipt result. A fresh engine loses callbacks yet finishes the exact
// unit; it never reinterprets P as a new account/discovery request.
func TestProvisioningOriginRemoveReferenceRecovery(t *testing.T) {
	for _, mode := range []string{"rollback", "restart", "cleanup-return", "cleanup-restart"} {
		t.Run(mode, func(t *testing.T) {
			fx, _, _, _, before := prepareRemovalOriginFixture(t)
			a := &originLeaseAuthority{profile: fx.profile, home: fx.home, released: true}
			m, e := projectlease.New(fx.home, a, a)
			if e != nil {
				t.Fatal(e)
			}
			_, e = m.Acquire(context.Background(), projectlease.Request{LeaseID: "removal-fault", ContextRef: "composition:removal", ContextSHA256: strings.Repeat("b", 64), OwnerRef: "principal:fixture", SessionRef: "session:removal", Generation: 1})
			if e != nil {
				t.Fatal(e)
			}
			if e = m.Reconcile(context.Background(), "removal-fault"); e != nil {
				t.Fatal(e)
			}
			profileBefore := nativeSnapshot(t, ProfileDir(fx.home, fx.profile))
			leasePath := filepath.Join(fx.home, "state", "project-leases", "v1", "ledger.json")
			ledgerBefore, e := os.ReadFile(leasePath)
			if e != nil {
				t.Fatal(e)
			}
			saved := newOperationEngine
			reached := false
			fault := fmt.Errorf("remove-reference-fault")
			newOperationEngine = func(home string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(home, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					hit := event.Point == transaction.PointAfterInstall && event.Class == "reference-record"
					if strings.HasPrefix(mode, "cleanup") {
						hit = event.Point == transaction.PointBeforeCleanup && event.Class == ""
					}
					if hit {
						reached = true
						if strings.Contains(mode, "restart") {
							panic("remove-reference-stop")
						}
						return fault
					}
					return nil
				}}))...)
			}
			func() {
				defer func() {
					newOperationEngine = saved
					if strings.Contains(mode, "restart") {
						if got := recover(); got != "remove-reference-stop" {
							t.Fatalf("wrong stop: %v", got)
						}
					}
				}()
				e = Remove(fx.home, fx.profile, false)
			}()
			if !reached {
				t.Fatal("actual removal boundary not reached")
			}
			if mode == "rollback" {
				if !errors.Is(e, fault) {
					t.Fatalf("intended fault: %v", e)
				}
				raw, er := os.ReadFile(leasePath)
				if er != nil || !bytes.Equal(raw, ledgerBefore) || nativeSnapshot(t, ProfileDir(fx.home, fx.profile)) != profileBefore {
					t.Fatal("rollback split installed/reference preimages")
				}
			} else {
				if e = Remove(fx.home, fx.profile, false); e != nil {
					t.Fatalf("fresh removal cleanup continuation: %v", e)
				}
				rows, er := m.Status(context.Background())
				if er != nil || len(rows) != 0 {
					t.Fatalf("committed removal reference survived: %v %+v", er, rows)
				}
				if _, er := os.Lstat(ProfileDir(fx.home, fx.profile)); !os.IsNotExist(er) {
					t.Fatal("committed removal resurrected installed profile")
				}
			}
			if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
				t.Fatal("fault changed retained home")
			}
		})
	}
}

// Reintroducing a formerly valid released reference is not disjoint just
// because its shared ledger path differs from P. Common admission must detect
// this substituted semantic dependency before any new GC/lease transaction.
func TestProvisioningOriginRestoredLeaseReference(t *testing.T) {
	fx, _, _, _, before := prepareRemovalOriginFixture(t)
	a := &originLeaseAuthority{profile: fx.profile, home: fx.home, released: true}
	m, e := projectlease.New(fx.home, a, a)
	if e != nil {
		t.Fatal(e)
	}
	_, e = m.Acquire(context.Background(), projectlease.Request{LeaseID: "restored-lease", ContextRef: "composition:restored", ContextSHA256: strings.Repeat("b", 64), OwnerRef: "principal:fixture", SessionRef: "session:restored", Generation: 1})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.Reconcile(context.Background(), "restored-lease"); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(fx.home, "state", "project-leases", "v1", "ledger.json")
	prior, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = Remove(fx.home, fx.profile, false); e != nil {
		t.Fatal(e)
	}
	if e = envauthority.Validate(fx.home); e != nil {
		t.Fatalf("valid revoked-reference control: %v", e)
	}
	// Even a typed journal with a permissive caller hook cannot reintroduce the
	// revoked reference. The common executor must inspect its actual postimage.
	op, e := beginOperation(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	staged := filepath.Join(t.TempDir(), "restored-ledger")
	if e = os.WriteFile(staged, prior, 0600); e != nil {
		t.Fatal(e)
	}
	current, e := transaction.DigestTarget(transaction.KindBytes, path)
	if e != nil {
		t.Fatal(e)
	}
	_, e = op.engine.Prepare(op.lock, transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "project-leases", Targets: []transaction.Target{{Class: "project-lease", Identifier: path, Kind: transaction.KindBytes, LivePath: path, StagedSource: staged, PreimageDigest: current, PrivateFile: true}}})
	_ = op.close()
	if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
		t.Fatalf("intended desired semantic reference refusal: %v", e)
	}
	if e = os.WriteFile(path, prior, 0600); e != nil {
		t.Fatal(e)
	}
	e = envauthority.Validate(fx.home)
	if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
		t.Fatalf("intended semantic reference refusal: %v", e)
	}
	if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
		t.Fatal("common admission changed retained home")
	}
}

// An installed but never materialized profile still owns installed records and
// lease references. Removal must not bypass live-lease refusal or durable
// revocation merely because the managed-home inventory is empty.
func TestProvisioningOriginRemoveUnmaterializedReferences(t *testing.T) {
	for _, released := range []bool{false, true} {
		t.Run(fmt.Sprintf("released=%t", released), func(t *testing.T) {
			fx, _, _, q, _ := prepareRemovalOriginFixture(t)
			profile := "unmaterialized-fixture"
			if e := os.Mkdir(ProfileDir(fx.home, profile), 0700); e != nil {
				t.Fatal(e)
			}
			for _, name := range []string{"source.json", "lock.json"} {
				raw, e := os.ReadFile(filepath.Join(ProfileDir(fx.home, q), name))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(ProfileDir(fx.home, profile), name), raw, 0600); e != nil {
					t.Fatal(e)
				}
			}
			a := &originLeaseAuthority{profile: profile, home: fx.home, released: released}
			m, e := projectlease.New(fx.home, a, a)
			if e != nil {
				t.Fatal(e)
			}
			_, e = m.Acquire(context.Background(), projectlease.Request{LeaseID: "unmaterialized-lease", ContextRef: "composition:unmaterialized", ContextSHA256: strings.Repeat("b", 64), OwnerRef: "principal:fixture", SessionRef: "session:unmaterialized", Generation: 1})
			if e != nil {
				t.Fatal(e)
			}
			if released {
				if e = m.Reconcile(context.Background(), "unmaterialized-lease"); e != nil {
					t.Fatal(e)
				}
			}
			installedBefore := nativeSnapshot(t, ProfileDir(fx.home, profile))
			leasePath := filepath.Join(fx.home, "state", "project-leases", "v1", "ledger.json")
			leaseBefore, e := os.ReadFile(leasePath)
			if e != nil {
				t.Fatal(e)
			}
			e = Remove(fx.home, profile, false)
			if !released {
				if e == nil || !strings.Contains(e.Error(), projectlease.Busy) {
					t.Fatalf("intended active lease refusal: %v", e)
				}
				after, er := os.ReadFile(leasePath)
				if er != nil || !bytes.Equal(after, leaseBefore) || nativeSnapshot(t, ProfileDir(fx.home, profile)) != installedBefore {
					t.Fatal("refusal changed installed/reference preimages")
				}
				return
			}
			if e != nil {
				t.Fatalf("released reference removal: %v", e)
			}
			rows, e := m.Status(context.Background())
			if e != nil || len(rows) != 0 {
				t.Fatalf("reference was not revoked: %v %+v", e, rows)
			}
			if e = Remove(fx.home, profile, false); e != nil {
				t.Fatalf("repeat removal completion: %v", e)
			}
			if e = Remove(fx.home, profile, true); e != nil {
				t.Fatalf("empty-inventory receipt purge: %v", e)
			}
			if e = Remove(fx.home, profile, true); e != nil {
				t.Fatalf("repeat empty purge: %v", e)
			}
			if e = envauthority.Validate(fx.home); e != nil {
				t.Fatal(e)
			}
		})
	}
}
