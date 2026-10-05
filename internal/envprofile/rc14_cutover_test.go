package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/transaction"
)

// Genuine v1 provisioning via Resolve, followed by the production credential
// migration; changing just the marker schema cannot satisfy this regression.
func TestRC14MigrationRehashesLegacyIdentities(t *testing.T) {
	requireLinkCapability(t)
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	rr := fx.request("pi")
	rr.Repair = true
	if _, err := Resolve(rr); err != nil {
		t.Fatal(err)
	}
	legacy, err := envmarker.Read(ManagedHomeDir(fx.home, fx.profile, "pi"))
	if err != nil {
		t.Fatal(err)
	}
	oldPin := legacy.Members[0].StateSHA256
	hashing.EnableV2Writers = true
	req := fx.migrateRequest()
	req.Profile = fx.profile
	req.EnvID = "pi"
	req.Machine.Isolation[fx.profile] = map[string]string{"pi": "isolated"}
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	req.Expect = report.Hash
	if _, err := ApplyMigration(req); err != nil {
		t.Fatal(err)
	}
	published, err := envmarker.Read(ManagedHomeDir(fx.home, fx.profile, "pi"))
	if err != nil {
		t.Fatal(err)
	}
	if published.Version != envmarker.VersionV3 || published.HashVersion != 2 {
		t.Fatalf("marker is not v2: %+v", published)
	}
	if published.Members[0].StateSHA256 == oldPin {
		t.Error("migration relabelled a v1 state identity as v2 without rehashing")
	}
	surface := published.Surfaces[envmarker.SurfaceRootContext]
	files := map[string][]byte{}
	for _, path := range surface.Paths {
		raw, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		files[path] = raw
	}
	want, err := contextmaterialize.SurfaceHashWithVersion(files, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if surface.ContentSHA256 != want {
		t.Errorf("migration surface hash = %s, actual v2 = %s", surface.ContentSHA256, want)
	}
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	if lock.HashVersion != 2 {
		t.Errorf("lock hash version = %d, want 2", lock.HashVersion)
	}
	member, _ := lock.RootMember()
	entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	actual, err := hashing.ContentSHA256WithVersion(entry, map[string]bool{}, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if member.StateHash != hashing.Normalize(actual) {
		t.Errorf("lock state pin = %s, actual v2 = %s", member.StateHash, actual)
	}
}

func legacyIdentityFixture(t *testing.T, envs ...string) *managedFixture {
	requireLinkCapability(t)
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	for _, id := range envs {
		provision(t, fx, id, fx.request(id).Machine)
	}
	hashing.EnableV2Writers = true
	return fx
}

func TestRC14ResolveRepairMigratesSiblingHomes(t *testing.T) {
	fx := legacyIdentityFixture(t, "pi", "claude_code", "opencode")
	before, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	oldMember, _ := before.RootMember()
	// Read-only Resolve still accepts a genuinely v1 profile under v2 writers.
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("legacy reader: %v", err)
	}
	req := fx.request("pi")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"pi", "claude_code", "opencode"} {
		marker := readManagedMarker(t, fx, id)
		if marker.Version != 3 || marker.HashVersion != 2 || marker.Members[0].StateSHA256 == oldMember.StateHash {
			t.Fatalf("sibling %s retained legacy identities: %+v", id, marker)
		}
		if _, err := Resolve(fx.request(id)); err != nil {
			t.Fatalf("migrated %s verification: %v", id, err)
		}
	}
	if _, err := os.Stat(contextstore.EntryDir(fx.home, oldMember.Kind, oldMember.Name, oldMember.PinKey())); err != nil {
		t.Fatalf("shared legacy store entry was removed: %v", err)
	}
}

func TestRC14IdentityMigrationRollsBackEveryEntry(t *testing.T) {
	fx := legacyIdentityFixture(t, "pi", "claude_code")
	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	req.Profile = fx.profile
	req.EnvID = "pi"
	req.Machine.Isolation[fx.profile] = map[string]string{"pi": "isolated"}
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	req.Expect = report.Hash
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointTargetCommitted && event.TargetIndex == 2 {
			return errors.New("injected joint publication failure")
		}
		return nil
	}})}
	if _, err := ApplyMigration(req); err == nil {
		t.Fatal("publication failure was accepted")
	}
	after := snapshotCredentialScope(t, fx)
	added, removed, changed := snapshotDiff(before, after)
	if len(added)+len(removed)+len(changed) != 0 {
		t.Fatalf("rollback changed entries: added=%v removed=%v changed=%v", added, removed, changed)
	}
	req.transactionOptions = nil
	if _, err := ApplyMigration(req); err != nil {
		t.Fatalf("retry of rolled-back plan: %v", err)
	}
}

func TestRC14IdentityMigrationRefusesSiblingPlanDrift(t *testing.T) {
	fx := legacyIdentityFixture(t, "pi", "claude_code")
	req := fx.migrateRequest()
	req.Profile = fx.profile
	req.EnvID = "pi"
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	req.Expect = report.Hash
	marker := readManagedMarker(t, fx, "claude_code")
	path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "claude_code"), marker.Surfaces[envmarker.SurfaceRootContext].Paths[0])
	if err := os.WriteFile(path, []byte("edited sibling\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := snapshotCredentialScope(t, fx)
	if _, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), "plan drift") {
		t.Fatalf("sibling drift error = %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added)+len(removed)+len(changed) != 0 {
		t.Fatal("refused plan mutated the profile")
	}
}

func TestRC14IdentityMigrationRefusesVersionAndPinMismatch(t *testing.T) {
	for _, scenario := range []string{"marker-version", "lock-version", "store-pin"} {
		t.Run(scenario, func(t *testing.T) {
			fx := legacyIdentityFixture(t, "pi")
			switch scenario {
			case "marker-version":
				marker := readManagedMarker(t, fx, "pi")
				marker.Version = 3
				marker.HashVersion = 2
				payload, err := marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), envmarker.Name), payload, 0o600); err != nil {
					t.Fatal(err)
				}
			case "lock-version":
				lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
				if err != nil {
					t.Fatal(err)
				}
				lock.SchemaVersion = 2
				lock.HashVersion = 2
				if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
					t.Fatal(err)
				}
			case "store-pin":
				lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
				if err != nil {
					t.Fatal(err)
				}
				member, _ := lock.RootMember()
				if err := os.WriteFile(filepath.Join(contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey()), "context", "a.md"), []byte("tampered\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			req := fx.migrateRequest()
			req.Profile = fx.profile
			if _, err := PlanMigration(req); err == nil {
				t.Fatal("mismatched identity was accepted")
			}
		})
	}
}

func TestRC14IdentityMigrationRecoversInterruptedCommit(t *testing.T) {
	fx := legacyIdentityFixture(t, "pi", "claude_code")
	req := fx.migrateRequest()
	req.Profile = fx.profile
	req.EnvID = "pi"
	req.Machine.Isolation[fx.profile] = map[string]string{"pi": "isolated"}
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	req.Expect = report.Hash
	crash := false
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointTargetCommitted && event.TargetIndex == 2 {
			crash = true
			panic("simulated process interruption")
		}
		return nil
	}})}
	func() {
		defer func() {
			if recovered := recover(); recovered != "simulated process interruption" {
				t.Fatalf("unexpected interruption: %v", recovered)
			}
		}()
		_, _ = ApplyMigration(req)
	}()
	if !crash {
		t.Fatal("production migration never entered the transaction")
	}
	// A subsequent mutation recovers the prepared sibling staging even though
	// the original assembly scratch directory no longer exists.
	rr := fx.request("pi")
	rr.Repair = true
	rr.Machine = req.Machine
	if _, err := Resolve(rr); err != nil {
		t.Fatalf("restart recovery: %v", err)
	}
	for _, id := range []string{"pi", "claude_code"} {
		marker := readManagedMarker(t, fx, id)
		if marker.Version != 3 || marker.HashVersion != 2 {
			t.Fatalf("recovery stranded %s in v1", id)
		}
	}
	if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), "auth.json")); !os.IsNotExist(err) {
		t.Fatalf("credential unlink was not recovered: %v", err)
	}
}

func TestRC14Schema1IdentityMigrationWithoutCredentialOperations(t *testing.T) {
	for _, operation := range []string{"resolve", "migrate"} {
		t.Run(operation, func(t *testing.T) {
			fx := legacyIdentityFixture(t, "pi")
			marker := readManagedMarker(t, fx, "pi")
			marker.Version = 1
			marker.HashVersion = 0
			records := []envmarker.Passthrough{}
			for _, record := range *marker.Passthrough {
				if record.Path != "" {
					records = append(records, envmarker.Passthrough{Path: record.Path, Strategy: record.Strategy})
				}
			}
			marker.Passthrough = &records
			payload, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), envmarker.Name), payload, 0o600); err != nil {
				t.Fatal(err)
			}
			if operation == "resolve" {
				req := fx.request("pi")
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
			} else {
				req := fx.migrateRequest()
				req.Profile = fx.profile
				req.EnvID = "pi"
				report, err := PlanMigration(req)
				if err != nil {
					t.Fatal(err)
				}
				req.Expect = report.Hash
				if len(report.Ops()) != 0 {
					t.Fatalf("unexpected credential operations: %+v", report.Ops())
				}
				if _, err := ApplyMigration(req); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := Resolve(fx.request("pi")); err != nil {
				t.Fatal(err)
			}
			published := readManagedMarker(t, fx, "pi")
			if published.Version != 3 || (*published.Passthrough)[0].BackendVersion == "" {
				t.Fatalf("schema-1 metadata not upgraded: %+v", published)
			}
		})
	}
}

func TestRC14IdentityMigrationPreservesFallbackCopies(t *testing.T) {
	fx := legacyIdentityFixture(t, "pi")
	marker := readManagedMarker(t, fx, "pi")
	surface := marker.Surfaces[envmarker.SurfaceSkills]
	path := surface.Paths[0]
	live := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), filepath.FromSlash(path))
	target, err := os.Readlink(live)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(live); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(target, ManagedHomeDir(fx.home, fx.profile, "pi"), path); err != nil {
		t.Fatal(err)
	}
	copies := append([]envmarker.Copy{}, *surface.Copies...)
	copies = append(copies, envmarker.Copy{Path: path, Reason: envmarker.ReasonSymlinkFallback})
	surface.Copies = &copies
	marker.Surfaces[envmarker.SurfaceSkills] = surface
	payload, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), envmarker.Name), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	rr := fx.request("pi")
	rr.Repair = true
	if _, err := Resolve(rr); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(live)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("fallback copy became a link: %v", err)
	}
	published := readManagedMarker(t, fx, "pi")
	got := published.Surfaces[envmarker.SurfaceSkills]
	if len(*got.Copies) != 1 || (*got.Copies)[0].Reason != envmarker.ReasonSymlinkFallback {
		t.Fatalf("fallback metadata lost: %+v", got)
	}
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("migrated copy verification: %v", err)
	}
}

func TestRC14UseMigratesLegacyProfileBeforeNativePublication(t *testing.T) {
	requireLinkCapability(t)
	native := pinHomes(t)
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	fx := writeManagedFixture(t, "sample")
	before, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	old, _ := before.RootMember()
	hashing.EnableV2Writers = true
	if _, err := UseWithPolicy(fx.home, fx.profile, "pi", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	member, _ := lock.RootMember()
	if lock.ContentHashVersion() != 2 || member.StateHash == old.StateHash {
		t.Fatal("Use did not migrate the old profile identity")
	}
	marker, err := envmarker.Read(native["pi"])
	if err != nil {
		t.Fatal(err)
	}
	if marker.Version != 3 || marker.HashVersion != 2 || marker.Members[0].StateSHA256 != member.StateHash {
		t.Fatalf("native marker retains a legacy identity: %+v", marker)
	}
	files := map[string][]byte{}
	for _, path := range marker.Surfaces[envmarker.SurfaceRootContext].Paths {
		data, err := os.ReadFile(filepath.Join(native["pi"], path))
		if err != nil {
			t.Fatal(err)
		}
		files[path] = data
	}
	want, err := contextmaterialize.SurfaceHashWithVersion(files, 2)
	if err != nil {
		t.Fatal(err)
	}
	if marker.Surfaces[envmarker.SurfaceRootContext].ContentSHA256 != want {
		t.Fatal("native surface hash was relabelled")
	}
}
