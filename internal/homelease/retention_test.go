package homelease_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/homelease"
	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

type retentionAuthority struct{ home, runtime, key string }

func (a retentionAuthority) Admit(_ context.Context, held managerlock.Held, environment, profile, digest string) (homelease.Admission, error) {
	if err := held.Validate(a.home); err != nil {
		return homelease.Admission{}, err
	}
	home := filepath.Join(a.home, "environments", profile, environment)
	return homelease.Admission{Home: home, Protected: []string{home}, RuntimeEntries: []string{a.runtime}, BuildKeys: []string{a.key}}, nil
}

// HL08/16/W18: prepared and persisted unknown holders retain real runtime
// entries absent from every ordinary scope, while the real collector removes
// an unreferenced neighbor. Build-key union is asserted at the real mark
// boundary; this test does not publish or execute a compiler artifact.
func TestHomeLeaseCollectorRetention(t *testing.T) {
	for _, phase := range []string{"prepared", "unknown"} {
		t.Run(phase, func(t *testing.T) {
			identity, err := managerlock.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			home := identity.Home()
			if err := privatedir.Protect(home); err != nil {
				t.Fatal(err)
			}
			retained, neighbor := "leased/"+strings.Repeat("a", 40), "neighbor/"+strings.Repeat("b", 40)
			for _, entry := range []string{retained, neighbor} {
				if err := os.MkdirAll(filepath.Join(home, "runtime", filepath.FromSlash(entry)), 0700); err != nil {
					t.Fatal(err)
				}
			}
			key := "sha256:" + strings.Repeat("c", 64)
			m, err := homelease.New(home, retentionAuthority{home, retained, key}, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := homelease.Request{Environment: "claude", Profile: "sample", LockSHA256: strings.Repeat("d", 64), RequestID: "retention", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "dead-launcher-manager", ManagerRegisteredAt: "2020-01-01T00:00:00Z", RequestID: "retention"}}
			out, _, err := m.Acquire(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			if phase == "unknown" {
				// Persisted uncertainty is a valid cold-start projection. Its
				// initial immutable binding and protected references stay intact.
				l, err := state.Load(home)
				if err != nil {
					t.Fatal(err)
				}
				e := l.Leases[out.Lease.LeaseID]
				e.Record.Phase = "unknown"
				at := "2020-01-01T00:00:01Z"
				e.Record.Observation = state.Observation{State: "unknown", ObservedAt: &at}
				l.Leases[out.Lease.LeaseID] = e
				raw, err := state.Canonical(l)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(state.LedgerPath(home), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := scopes.CollectRuntime(home); err == nil {
				t.Fatal("historical unlocked collector accepted lease state")
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(neighbor))); err != nil {
				t.Fatalf("refusal swept neighbor: %v", err)
			}
			locks, _ := managerlock.New(home)
			lock, err := locks.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			roots, err := homelease.ReferencesUnderLock(home, lock)
			if err != nil || !reflect.DeepEqual(roots.RuntimeEntries, []string{retained}) || !reflect.DeepEqual(roots.BuildKeys, []string{key}) {
				t.Fatalf("retained roots: %+v %v", roots, err)
			}
			result, err := scopes.Collect(scopes.MaintenanceRequest{Home: home, Lock: lock})
			_ = lock.Close()
			if err != nil || !reflect.DeepEqual(result.RemovedRuntime, []string{neighbor}) {
				t.Fatalf("collector: %+v %v", result, err)
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(retained))); err != nil {
				t.Fatalf("live root lost: %v", err)
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(neighbor))); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("unreferenced control retained: %v", err)
			}
			rows, err := m.List(context.Background(), "", "")
			if err != nil || len(rows) != 1 || rows[0].Phase != phase {
				t.Fatalf("collector expired lease: %+v %v", rows, err)
			}
		})
	}
}

// HL03/16: the same physical manager has the same retained roots through a
// directory alias. A foreign or closed acquisition cannot authorize that read;
// both refusals and successful reference reads preserve the complete preimage.
func TestHomeLeaseCollectorManagerAlias(t *testing.T) {
	for _, spelling := range []string{"canonical", "alias"} {
		t.Run(spelling, func(t *testing.T) {
			base := t.TempDir()
			configured := filepath.Join(base, "manager")
			if err := privatedir.MakeAll(configured); err != nil {
				t.Fatal(err)
			}
			locks, err := managerlock.New(configured)
			if err != nil {
				t.Fatal(err)
			}
			home := locks.Home()
			query := home
			if spelling == "alias" {
				query = filepath.Join(base, "manager-alias")
				if err := os.Symlink(home, query); err != nil {
					t.Fatal(err)
				}
			}
			retained := "leased/" + strings.Repeat("a", 40)
			key := "sha256:" + strings.Repeat("c", 64)
			m, err := homelease.New(query, retentionAuthority{home, retained, key}, nil)
			if err != nil {
				t.Fatal(err)
			}
			req := homelease.Request{Environment: "claude", Profile: "sample", LockSHA256: strings.Repeat("d", 64), RequestID: "alias-retention", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "dead-launcher-manager", ManagerRegisteredAt: "2020-01-01T00:00:00Z", RequestID: "alias-retention"}}
			if _, _, err := m.Acquire(context.Background(), req); err != nil {
				t.Fatal(err)
			}
			if _, err := state.Load(home); err != nil {
				t.Fatalf("canonical ledger control: %v", err)
			}
			lock, err := locks.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			before := boundarySnapshot(t, home)
			roots, err := homelease.ReferencesUnderLock(query, lock)
			if err != nil || !reflect.DeepEqual(roots.RuntimeEntries, []string{retained}) || !reflect.DeepEqual(roots.BuildKeys, []string{key}) {
				t.Errorf("same-root retained references: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("reference read changed preimage")
			}
			foreignManager, err := managerlock.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			foreign, err := foreignManager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			roots, err = homelease.ReferencesUnderLock(query, foreign)
			if closeErr := foreign.Close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if !errors.Is(err, managerlock.ErrNotHeld) || !reflect.DeepEqual(roots, homelease.Roots{}) {
				t.Errorf("foreign lock: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("foreign-lock refusal changed preimage")
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			roots, err = homelease.ReferencesUnderLock(query, lock)
			if !errors.Is(err, managerlock.ErrNotHeld) || !reflect.DeepEqual(roots, homelease.Roots{}) {
				t.Errorf("closed lock: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("closed-lock refusal changed preimage")
			}
		})
	}
}

// HL16/W18: synthetic independent home/project authority retains real protected
// cache publications and runtime directories with no ordinary install markers.
// Prepared and unknown home holders join project roots; a real pending acquire
// refuses the whole collector without changing any fixture-owned preimage.
// Ending home protection permits its artifacts to be swept while the durable
// project lease/view still retains its own roots, even after reconciliation.
// Artifact bytes,
// build inputs and portable receipts are fixtures; no compiler, runtime bridge
// or production observer is qualified by this test.
func TestHomeLeaseCollectorArtifactUnion(t *testing.T) {
	for _, phase := range []string{"prepared", "unknown", "pending-acquire"} {
		t.Run(phase, func(t *testing.T) {
			identity, err := managerlock.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			home := identity.Home()
			if err := privatedir.Protect(home); err != nil {
				t.Fatal(err)
			}
			store, err := buildcache.New(home)
			if err != nil {
				t.Fatal(err)
			}
			// No native process observation is claimed; fixture artifacts are inert.
			store.ExecutablePaths = func() ([]string, error) { return nil, nil }
			lock, err := identity.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			homeArtifact := publishLeaseFixtureArtifact(t, store, lock, "home-retained")
			projectArtifact := publishLeaseFixtureArtifact(t, store, lock, "project-retained")
			orphanArtifact := publishLeaseFixtureArtifact(t, store, lock, "orphan")
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			for _, entry := range []string{"tool/home", "tool/project", "tool/orphan"} {
				if err := privatedir.MakeAll(filepath.Join(home, "runtime", filepath.FromSlash(entry))); err != nil {
					t.Fatal(err)
				}
			}
			pa := projectAuthority{home: home, composition: projectlease.Composition{Ref: "composition:retained", SHA256: strings.Repeat("e", 64), CheckoutID: "checkout", Profile: "sample", Environment: "codex_cli", HomeRef: "home:project", ConcurrentState: true, BuildKeys: []string{string(projectArtifact.CacheKey)}, RuntimeEntries: []string{"tool/project"}, Generated: map[string]projectlease.Surface{}}}
			pm, err := projectlease.New(home, pa, pa)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := pm.Acquire(context.Background(), projectlease.Request{LeaseID: "project-retained", ContextRef: pa.composition.Ref, ContextSHA256: pa.composition.SHA256, OwnerRef: "principal:operator", SessionRef: "session:fixture", Generation: 1}); err != nil {
				t.Fatal(err)
			}
			authority := retentionAuthority{home, "tool/home", string(homeArtifact.CacheKey)}
			m, err := homelease.New(home, authority, nil)
			if err != nil {
				t.Fatal(err)
			}
			fired := false
			if phase == "pending-acquire" {
				m, err = homelease.NewPublicationFaultManagerForTest(home, authority, transaction.Hooks{Fault: func(e transaction.Event) error {
					if !fired && e.Point == transaction.PointBeforeCleanup {
						fired = true
						return errors.New("fixture pending acquire")
					}
					return nil
				}})
				if err != nil {
					t.Fatal(err)
				}
			}
			req := homelease.Request{Environment: "claude", Profile: "sample", LockSHA256: strings.Repeat("d", 64), RequestID: "artifact-retention", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2020-01-01T00:00:00Z", RequestID: "artifact-retention"}}
			out, _, acquireErr := m.Acquire(context.Background(), req)
			if phase == "pending-acquire" {
				var refusal *state.Refusal
				if !fired || out.Changed || !errors.As(acquireErr, &refusal) || refusal.Code != "environment_lease_write_failed" || refusal.Reason != "publication_failed" {
					t.Fatalf("pending acquisition not exercised: %+v %v", out, acquireErr)
				}
			} else if acquireErr != nil {
				t.Fatal(acquireErr)
			}
			collect := func() (scopes.MaintenanceResult, error) {
				lock, err := identity.AcquireHomeOnly(context.Background(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
				return scopes.Collect(scopes.MaintenanceRequest{Home: home, Lock: lock, Cache: store, Now: time.Now().Add(48 * time.Hour), Grace: time.Hour})
			}
			if phase == "pending-acquire" {
				before := boundarySnapshot(t, home)
				result, err := collect()
				var refusal *state.Refusal
				if !errors.As(err, &refusal) || refusal.Code != "environment_lease_recovery_required" || refusal.Reason != "pending_mutation" || len(result.RemovedBuilds)+len(result.RemovedRuntime) != 0 {
					t.Fatalf("pending collector: %+v %v", result, err)
				}
				if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
					t.Fatal("pending collector changed fixture preimage")
				}
				lock, err := identity.AcquireHomeOnly(context.Background(), false)
				if err != nil {
					t.Fatal(err)
				}
				engine, err := transaction.New(home)
				if err != nil {
					t.Fatal(err)
				}
				if err := engine.Recover(lock); err != nil {
					t.Fatal(err)
				}
				if err := lock.Close(); err != nil {
					t.Fatal(err)
				}
				m, err = homelease.New(home, authority, nil)
				if err != nil {
					t.Fatal(err)
				}
				out, _, err = m.Acquire(context.Background(), req)
				if err != nil || out.Changed || out.Lease == nil {
					t.Fatalf("recovered original acquisition: %+v %v", out, err)
				}
			}
			if phase == "unknown" {
				ledger, err := state.Load(home)
				if err != nil {
					t.Fatal(err)
				}
				entry := ledger.Leases[out.Lease.LeaseID]
				at := "2020-01-01T00:00:01Z"
				entry.Record.Phase = "unknown"
				entry.Record.Observation = state.Observation{State: "unknown", ObservedAt: &at}
				ledger.Leases[out.Lease.LeaseID] = entry
				raw, err := state.Canonical(ledger)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(state.LedgerPath(home), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			assertSweep := func(want buildcache.PublicationResult, runtimeEntry string) {
				t.Helper()
				result, err := collect()
				if err != nil || !reflect.DeepEqual(result.RemovedBuilds, []string{string(want.CacheKey)}) || !reflect.DeepEqual(result.RemovedRuntime, []string{runtimeEntry}) {
					t.Fatalf("collector result: %+v %v", result, err)
				}
				if _, err := os.Lstat(want.ArtifactPath); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("unreferenced build artifact retained", err)
				}
			}
			assertPresent := func(artifact buildcache.PublicationResult, runtimeEntry string) {
				raw, err := os.ReadFile(artifact.ArtifactPath)
				if err != nil || string(raw) != "isolated artifact fixture" {
					t.Fatal("retained build lost or changed", err)
				}
				if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(runtimeEntry))); err != nil {
					t.Fatal("retained runtime lost", err)
				}
			}
			assertSweep(orphanArtifact, "tool/orphan")
			assertPresent(homeArtifact, "tool/home")
			assertPresent(projectArtifact, "tool/project")
			if _, _, err := m.Break(context.Background(), out.Lease.LeaseID, 1, true, "fixture protection explicitly abandoned"); err != nil {
				t.Fatal(err)
			}
			assertSweep(homeArtifact, "tool/home")
			assertPresent(projectArtifact, "tool/project")
			if err := pm.Reconcile(context.Background(), "project-retained"); err != nil {
				t.Fatal(err)
			}
			// Reconciliation releases liveness, not the durable composition roots.
			result, err := collect()
			if err != nil || len(result.RemovedBuilds)+len(result.RemovedRuntime) != 0 || len(result.Warnings) != 0 {
				t.Fatalf("durable project roots after reconciliation: %+v %v", result, err)
			}
			assertPresent(projectArtifact, "tool/project")
		})
	}
}

func publishLeaseFixtureArtifact(t *testing.T, store *buildcache.Store, lock *managerlock.HomeLock, command string) buildcache.PublicationResult {
	t.Helper()
	tuning := map[string]string{}
	switch runtime.GOARCH {
	case "amd64":
		tuning["GOAMD64"] = "v1"
	case "arm64":
		tuning["GOARM64"] = "v8.0"
	case "386":
		tuning["GO386"] = "sse2"
	}
	input := buildmeta.Input{SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1, BuildSource: buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}, BuildRoot: "build", Command: command, SourceDir: "build/cmd/" + command, Target: buildmeta.Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Tuning: tuning}, Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath, GoVersion: "go version " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, ContentSHA256: "sha256:" + strings.Repeat("c", 64)}, Policy: buildmeta.FixedPolicy()}
	artifact := []byte("isolated artifact fixture")
	path, err := buildmeta.ArtifactPath(command, runtime.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(artifact)
	record := buildmeta.Artifact{Path: path, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(artifact))}
	receipt, err := buildmeta.NewReceipt(input, record)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	execution, err := closureexec.NewPortableBuildSessionReceipt(input, record, nil)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "artifact")
	if err := os.WriteFile(source, artifact, 0600); err != nil {
		t.Fatal(err)
	}
	published, err := store.Publish(buildcache.Publication{Input: input, ReceiptBytes: raw, Assurance: closureexec.PortableAssuranceBinding(), ExecutionReceipt: execution, ArtifactSource: source}, lock)
	if err != nil {
		t.Fatal(err)
	}
	return published
}
