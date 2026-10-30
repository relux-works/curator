package envprofile

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/homelease"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
	"path/filepath"
)

func leaseManagedFixture(t *testing.T, fx *managedFixture, id string) (*homelease.Manager, *homelease.Record) {
	t.Helper()
	m, err := HomeLeaseManager(fx.request(id), nil)
	if err != nil {
		t.Fatal(err)
	}
	r := homelease.Request{Environment: id, Profile: fx.profile, LockSHA256: strings.TrimPrefix(fx.lockHash, "sha256:"), RequestID: "writer-request", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "writer-request"}}
	out, reason, err := m.Acquire(context.Background(), r)
	if err != nil || reason != "acquired" || !out.Changed {
		t.Fatalf("admit: %+v %s %v", out, reason, err)
	}
	return m, out.Lease
}

// W02: a linked, pinned store entry cannot be reconstructed while its consumer
// is leased. Releasing protection permits reconstruction from the real pinned
// Git object, followed by successful ordinary materialization verification.
func TestHomeLeaseStoreReconstruction(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("skill", "myskill")
	if !ok {
		t.Fatal("missing pinned fixture member")
	}
	entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	if err := os.WriteFile(filepath.Join(entry, "SKILL.md"), []byte("corrupt pinned source"), 0600); err != nil {
		t.Fatal(err)
	}
	r := fx.request(envregistry.CodexCLI)
	r.Repair = true
	before, _ := captureFixtureState(t, fx)
	_, err = Resolve(r)
	requireHomeLeaseRefusal(t, err, lease.LeaseID)
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("store repair refusal changed preimage")
	}
	breakFixtureLease(t, m, lease)
	if _, err := Resolve(r); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(fx.request(envregistry.CodexCLI)); err != nil {
		t.Fatalf("rebuilt positive control: %v", err)
	}
}

// W08: the standalone global library entry refuses before extending a lock or
// materializing a home. The unleased control completes both native and managed
// phases using isolated provider homes and a PATH without provider executables.
func TestHomeLeaseStandaloneGlobalInstall(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	write := func() error {
		_, err := GlobalInstall(fx.home, Policy{}, envregistry.DefaultMachineConfig(), fx.request(envregistry.CodexCLI).NativeHomeOf)
		return err
	}
	before, _ := captureFixtureState(t, fx)
	requireHomeLeaseRefusal(t, write(), lease.LeaseID)
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("global refusal changed preimage")
	}
	breakFixtureLease(t, m, lease)
	if err := write(); err != nil {
		t.Fatalf("unleased global control: %v", err)
	}
}

// W06/14: reinstall and native unmanage refuse before records, native aliases
// or the global selection change. After an explicit break the same entry point
// changes real fixture state, rather than merely returning a no-op success.
func TestHomeLeaseReinstallAndUnmanage(t *testing.T) {
	for _, action := range []string{"reinstall", "unmanage", "restore-backups"} {
		t.Run(action, func(t *testing.T) {
			fx, m, lease := leasedManagedFixture(t)
			write := func() error {
				if action == "reinstall" {
					source, err := readSource(fx.home, fx.profile)
					if err != nil {
						return err
					}
					_, _, _, err = Install(fx.home, InstallOptions{Operand: source.Path, As: fx.profile, ConfirmSystemDelta: true})
					return err
				}
				_, err := Unmanage(UnmanageRequest{Home: fx.home, EnvID: envregistry.CodexCLI, RestoreBackups: action == "restore-backups", NativeHomeOf: fx.request(envregistry.CodexCLI).NativeHomeOf})
				return err
			}
			before, _ := captureFixtureState(t, fx)
			requireHomeLeaseRefusal(t, write(), lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("refusal changed preimage")
			}
			breakFixtureLease(t, m, lease)
			before, _ = captureFixtureState(t, fx)
			if err := write(); err != nil {
				t.Fatal(err)
			}
			after, _ = captureFixtureState(t, fx)
			if bytes.Equal(before, after) {
				t.Fatal("positive control made no real change")
			}
		})
	}
}

// W15: a lease protects even a non-current profile's authority. A disjoint
// non-current profile can actually be removed while that reservation survives.
func TestHomeLeaseNoncurrentRemoval(t *testing.T) {
	for _, purge := range []bool{false, true} {
		t.Run(map[bool]string{false: "remove", true: "purge"}[purge], func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			other := t.TempDir()
			writePackage(t, other, "other", "1.0.0", "other\n")
			if _, _, _, err := Install(fx.home, InstallOptions{Operand: other}); err != nil {
				t.Fatal(err)
			}
			if err := SetCurrent(fx.home, "other"); err != nil {
				t.Fatal(err)
			}
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			before, _ := captureFixtureState(t, fx)
			requireHomeLeaseRefusal(t, Remove(fx.home, fx.profile, purge), lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("refused removal changed preimage")
			}
			breakFixtureLease(t, m, lease)
			if err := Remove(fx.home, fx.profile, purge); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(ProfileDir(fx.home, fx.profile)); !os.IsNotExist(err) {
				t.Fatalf("released profile was not removed: %v", err)
			}
		})
	}
}

// HL14/W04: a batch whose second profile is leased cannot update its first
// profile or even invoke the per-row callback. A disjoint profile installation
// remains writable while the same lease blocks the batch.
func TestHomeLeaseBatchAndDisjointWriter(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	source := t.TempDir()
	writePackage(t, source, "unrelated", "1.0.0", "unrelated\n")
	protected := captureTree(t, ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))
	if _, _, _, err := Install(fx.home, InstallOptions{Operand: source}); err != nil {
		t.Fatalf("disjoint install: %v", err)
	}
	if _, err := os.Stat(lockPath(fx.home, "unrelated")); err != nil {
		t.Fatalf("disjoint install did not publish: %v", err)
	}
	if !reflect.DeepEqual(protected, captureTree(t, ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))) {
		t.Fatal("disjoint writer changed leased home")
	}
	called := 0
	visit := func(string, Info, bool, error) bool { called++; return true }
	before, _ := captureFixtureState(t, fx)
	requireHomeLeaseRefusal(t, UpdateProfilesWithOptions(fx.home, []string{"unrelated", fx.profile}, UpdateOptions{}, visit), lease.LeaseID)
	after, _ := captureFixtureState(t, fx)
	if called != 0 || !bytes.Equal(before, after) {
		t.Fatal("batch wrote or visited its first target before checking the second")
	}
	breakFixtureLease(t, m, lease)
	if err := UpdateProfilesWithOptions(fx.home, []string{"unrelated", fx.profile}, UpdateOptions{}, visit); err != nil || called != 2 {
		t.Fatalf("unleased batch control: calls=%d err=%v", called, err)
	}
}

// W03/13: a genuinely v1 home is admitted before enabling v2 writers.
// Resolve, use and explicit migration refuse, then publish v2 identities after
// break while a disjoint profile remains leased. No credential payload is read
// by the lease guard.
func TestHomeLeaseIdentityMigration(t *testing.T) {
	for _, action := range []string{"resolve", "explicit", "explicit-sibling-env", "use"} {
		t.Run(action, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			var fx *managedFixture
			if action == "explicit-sibling-env" {
				fx = legacyIdentityFixture(t, envregistry.CodexCLI, envregistry.ClaudeCode)
			} else {
				fx = legacyIdentityFixture(t, envregistry.CodexCLI)
			}
			hashing.EnableV2Writers = false
			checkDisjoint := preserveDisjointHomeLease(t, fx)
			t.Setenv("CODEX_HOME", fx.native[envregistry.CodexCLI])
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			hashing.EnableV2Writers = true
			write := func() error {
				if action == "use" {
					_, err := UseWithPolicy(fx.home, fx.profile, envregistry.CodexCLI, "", false, Policy{})
					return err
				}
				if action == "explicit" || action == "explicit-sibling-env" {
					req := fx.migrateRequest()
					req.Profile, req.EnvID = fx.profile, envregistry.CodexCLI
					if action == "explicit-sibling-env" {
						req.EnvID = envregistry.ClaudeCode
					}
					plan, err := PlanMigration(req)
					if err != nil {
						return err
					}
					req.Expect = plan.Hash
					_, err = ApplyMigration(req)
					return err
				}
				req := fx.request(envregistry.CodexCLI)
				req.Repair = true
				_, err := Resolve(req)
				return err
			}
			before, _ := captureFixtureState(t, fx)
			requireHomeLeaseRefusal(t, write(), lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("migration refusal changed preimage")
			}
			breakFixtureLease(t, m, lease)
			if err := write(); err != nil {
				t.Fatal(err)
			}
			marker, err := envmarker.Read(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))
			if err != nil || marker == nil || marker.HashVersion != 2 {
				t.Fatalf("positive migration did not rehash: %+v %v", marker, err)
			}
			if action == "explicit-sibling-env" {
				sibling, err := envmarker.Read(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode))
				if err != nil || sibling == nil || sibling.HashVersion != 2 {
					t.Fatalf("selected sibling not migrated: %+v %v", sibling, err)
				}
			}
			checkDisjoint()
		})
	}
}

// HL02: a borrowed writer cannot use an already closed real lock witness even
// if its earlier admission succeeded. A current witness admits disjoint work.
func TestHomeLeaseBorrowedWitness(t *testing.T) {
	fx, _, _ := leasedManagedFixture(t)
	op, err := beginOperationGuarded(fx.home, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if err := homeLeaseGuard(fx.home, op.lock, []string{}); err != nil {
		t.Fatal(err)
	}
	_ = op.close()
	before := captureTree(t, fx.home)
	if err := homeLeaseGuard(fx.home, op.lock, []string{}); err == nil {
		t.Fatal("closed borrowed witness admitted a write")
	}
	if !reflect.DeepEqual(before, captureTree(t, fx.home)) {
		t.Fatal("stale witness refusal changed state")
	}
}

// W17: direct legacy migration recovery is guarded even when entered with a
// borrowed lock. A completed journal still requires a write to remove it. The
// fixture owner withdraws and restores the planted intent around an explicit
// lease break; the unleased control then performs real recovery cleanup.
func TestHomeLeaseLegacyMigrationRecovery(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	journal := &migrationJournal{Version: 1, Plan: strings.Repeat("a", 64), Complete: true}
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	op, err := beginOperationGuarded(fx.home, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = recoverMigrationJournal(fx.home, op, nil, os.Symlink, os.Rename)
	requireHomeLeaseRefusal(t, err, lease.LeaseID)
	if err := op.close(); err != nil {
		t.Fatal(err)
	}
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("legacy recovery refusal changed preimage")
	}
	if err := os.Remove(migrationJournalPath(fx.home)); err != nil {
		t.Fatal(err)
	}
	breakFixtureLease(t, m, lease)
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	op, err = beginOperationGuarded(fx.home, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = op.close() }()
	if _, err := recoverMigrationJournal(fx.home, op, nil, os.Symlink, os.Rename); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("unleased recovery did not remove completed intent: %v", err)
	}
}

// HL17: no-repair resolve detects pending legacy intent even before the first
// lease namespace exists, emits no result, and does not reconcile it. Removing
// only that fixture-owned pending intent permits the same read-only resolve.
func TestHomeLeaseReadOnlyLegacyIntent(t *testing.T) {
	isolateHomeLeaseWriters(t)
	fx, _ := captureFixture(t)
	if err := writeMigrationJournal(fx.home, &migrationJournal{Version: 1, Plan: strings.Repeat("a", 64), Complete: true}); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	result, err := Resolve(fx.request(envregistry.CodexCLI))
	var refusal *homelease.Refusal
	if result != nil || !errors.As(err, &refusal) || refusal.Code != "environment_lease_recovery_required" || refusal.Reason != "pending_mutation" {
		t.Fatalf("read-only pending intent: result=%+v err=%v", result, err)
	}
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("read-only pending refusal changed preimage")
	}
	if err := os.Remove(migrationJournalPath(fx.home)); err != nil {
		t.Fatal(err)
	}
	before, _ = captureFixtureState(t, fx)
	if _, err := Resolve(fx.request(envregistry.CodexCLI)); err != nil {
		t.Fatal(err)
	}
	after, _ = captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("positive read-only resolve changed preimage")
	}
}

// HL17: pending home publication refuses no-repair resolve, while an unrelated
// valid state publication remains a read-only positive control. Both calls
// preserve the complete preimage, including the still-pending journal.
func TestHomeLeaseReadOnlyPendingTargets(t *testing.T) {
	for _, relevant := range []bool{false, true} {
		t.Run(map[bool]string{false: "unrelated-control", true: "managed-home-refusal"}[relevant], func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			target := filepath.Join(fx.home, "state", "unrelated-note")
			if relevant {
				target = filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), ".agent-environment.json")
			}
			prepareAuthorityIntent(t, fx.home, "pending-read", target, []byte("pending replacement"))
			before, _ := captureFixtureState(t, fx)
			result, err := Resolve(fx.request(envregistry.CodexCLI))
			if relevant {
				var refusal *homelease.Refusal
				if result != nil || !errors.As(err, &refusal) || refusal.Code != "environment_lease_recovery_required" || refusal.Reason != "pending_mutation" {
					t.Fatalf("pending home publication: %+v %v", result, err)
				}
			} else if err != nil || result == nil {
				t.Fatalf("unrelated pending control: %+v %v", result, err)
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("read-only pending observation changed preimage")
			}
		})
	}
}

// W09: a recorded managed skill copy remains current without writes, then a
// prepared lease blocks its actual repair before any fixture-owned change.
// After an explicit fixture break the same repair restores the full copied
// tree in each managed environment, including removing a drifted extra file.
// Initial symlink-failure selection and native copy fallback remain unqualified.
func TestHomeLeaseRecordedCopyFallback(t *testing.T) {
	for _, tc := range []struct{ environment, drift string }{
		{envregistry.CodexCLI, "changed-bytes"}, {envregistry.CodexCLI, "extra-file"},
		{envregistry.ClaudeCode, "changed-bytes"}, {envregistry.ClaudeCode, "extra-file"},
		{envregistry.Muse, "changed-bytes"}, {envregistry.Muse, "extra-file"},
	} {
		t.Run(tc.environment+"/"+tc.drift, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			if tc.environment == envregistry.Muse {
				fx.native[envregistry.Muse] = t.TempDir()
				if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic-native\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.environment != envregistry.CodexCLI {
				provision := fx.request(tc.environment)
				provision.Repair = true
				if _, err := Resolve(provision); err != nil {
					t.Fatal(err)
				}
			}
			root := ManagedHomeDir(fx.home, fx.profile, tc.environment)
			adapter, err := envregistry.ByID(tc.environment)
			if err != nil {
				t.Fatal(err)
			}
			rel := adapter.SkillsDir + "/myskill"
			live := filepath.Join(root, filepath.FromSlash(rel))
			source, err := os.Readlink(live)
			if err != nil {
				t.Fatal(err)
			}
			want := projectionFiles(t, source)
			if err := os.Remove(live); err != nil {
				t.Fatal(err)
			}
			if err := copyTree(source, root, rel); err != nil {
				t.Fatal(err)
			}
			marker, err := envmarker.Read(root)
			if err != nil {
				t.Fatal(err)
			}
			surface := marker.Surfaces[envmarker.SurfaceSkills]
			copies := []envmarker.Copy{}
			if surface.Copies != nil {
				copies = append(copies, (*surface.Copies)...)
			}
			copies = append(copies, envmarker.Copy{Path: rel, Reason: envmarker.ReasonSymlinkFallback})
			surface.Copies = &copies
			marker.Surfaces[envmarker.SurfaceSkills] = surface
			raw, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, envmarker.Name), raw, 0600); err != nil {
				t.Fatal(err)
			}
			before, _ := captureFixtureState(t, fx)
			if _, err := Resolve(fx.request(tc.environment)); err != nil {
				t.Fatalf("recorded copy is not current: %v", err)
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("current recorded-copy read changed state")
			}
			m, lease := leaseManagedFixture(t, fx, tc.environment)
			path := filepath.Join(live, "SKILL.md")
			if tc.drift == "extra-file" {
				path = filepath.Join(live, "extra.md")
			}
			if err := os.WriteFile(path, []byte("isolated copy drift\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before, _ = captureFixtureState(t, fx)
			req := fx.request(tc.environment)
			req.Repair = true
			_, err = Resolve(req)
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ = captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("leased copy-fallback repair changed refusal preimage")
			}
			breakFixtureLease(t, m, lease)
			if _, err := Resolve(req); err != nil {
				t.Fatalf("unleased copy repair: %v", err)
			}
			info, err := os.Lstat(live)
			if err != nil || !info.IsDir() {
				t.Fatalf("recorded fallback no longer a copied directory: %v", err)
			}
			if !reflect.DeepEqual(projectionFiles(t, live), want) {
				t.Fatal("copy repair did not restore the complete admitted projection")
			}
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("repaired copy is not current: %v", err)
			}
		})
	}
}

// preserveDisjointHomeLease installs and verifies an unrelated profile through
// the real APIs. Its prepared reservation and protected surfaces must survive
// the target writer's successful control, not merely its refusal.
func preserveDisjointHomeLease(t *testing.T, fx *managedFixture) func() {
	t.Helper()
	source := t.TempDir()
	writePackage(t, source, "lease-neighbor", "1.0.0", "independent fixture\n")
	if _, _, _, err := Install(fx.home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	neighbor := *fx
	neighbor.profile = "lease-neighbor"
	_, _, hash, err := loadResolveInputs(fx.home, neighbor.profile)
	if err != nil {
		t.Fatal(err)
	}
	neighbor.lockHash = hash
	req := neighbor.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	m, err := HomeLeaseManager(neighbor.request(envregistry.CodexCLI), nil)
	if err != nil {
		t.Fatal(err)
	}
	request := homelease.Request{Environment: envregistry.CodexCLI, Profile: neighbor.profile, LockSHA256: strings.TrimPrefix(hash, "sha256:"), RequestID: "disjoint-fixture", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "disjoint-fixture"}}
	out, reason, err := m.Acquire(context.Background(), request)
	if err != nil || !out.Changed || out.Lease == nil || reason != "acquired" {
		t.Fatalf("disjoint admission: %+v %s %v", out, reason, err)
	}
	before := map[string]map[string]string{"profile": captureTree(t, ProfileDir(fx.home, neighbor.profile)), "home": captureTree(t, ManagedHomeDir(fx.home, neighbor.profile, envregistry.CodexCLI)), "source": captureTree(t, source)}
	return func() {
		t.Helper()
		rows, err := m.List(context.Background(), envregistry.CodexCLI, neighbor.profile)
		if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0], *out.Lease) || !rows[0].Blocking() {
			t.Fatalf("disjoint reservation changed: %+v %v", rows, err)
		}
		after := map[string]map[string]string{"profile": captureTree(t, ProfileDir(fx.home, neighbor.profile)), "home": captureTree(t, ManagedHomeDir(fx.home, neighbor.profile, envregistry.CodexCLI)), "source": captureTree(t, source)}
		if !reflect.DeepEqual(before, after) {
			t.Fatal("target writer changed disjoint protected surfaces")
		}
	}
}

// W02: two independently verified profile aliases consume the same pinned
// source. Repair through the unleased consumer must still honor the other
// consumer's reservation before changing any source or projection. After that
// holder ends, the pinned content is restored and both consumers verify while
// a separate disjoint profile remains leased. All credentials are synthetic.
func TestHomeLeaseStoreTwoConsumers(t *testing.T) {
	isolateHomeLeaseWriters(t)
	fx, _ := captureFixture(t)
	checkDisjoint := preserveDisjointHomeLease(t, fx)
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(sourcePath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	other := *fx
	other.profile = "second-consumer"
	if err := os.MkdirAll(ProfileDir(fx.home, other.profile), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath(fx.home, other.profile), source, 0600); err != nil {
		t.Fatal(err)
	}
	other.lockHash, err = contextlock.Write(lockPath(fx.home, other.profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	req := other.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatalf("second consumer admission: %v", err)
	}
	member, ok := lock.Find("skill", "myskill")
	if !ok {
		t.Fatal("missing shared skill")
	}
	_, secondLock, _, err := loadResolveInputs(fx.home, other.profile)
	if err != nil {
		t.Fatal(err)
	}
	secondMember, ok := secondLock.Find("skill", "myskill")
	if !ok || member.PinKey() != secondMember.PinKey() {
		t.Fatal("fixture consumers do not share a pin")
	}
	entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	pinned, err := os.ReadFile(filepath.Join(entry, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
	if err := os.WriteFile(filepath.Join(entry, "SKILL.md"), []byte("shared fixture drift"), 0600); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	_, err = Resolve(req)
	requireHomeLeaseRefusal(t, err, lease.LeaseID)
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("shared consumer refusal changed preimage")
	}
	breakFixtureLease(t, m, lease)
	if _, err := Resolve(req); err != nil {
		t.Fatalf("shared reconstruction: %v", err)
	}
	actual, err := os.ReadFile(filepath.Join(entry, "SKILL.md"))
	if err != nil || !bytes.Equal(actual, pinned) {
		t.Fatal("shared pin not restored", err)
	}
	for _, consumer := range []*managedFixture{fx, &other} {
		if _, err := Resolve(consumer.request(envregistry.CodexCLI)); err != nil {
			t.Fatalf("reconstructed consumer %s: %v", consumer.profile, err)
		}
	}
	checkDisjoint()
}

// W01: takeover of an existing managed home and first provisioning through a
// shared dependency both refuse without writes while that dependency is leased.
// Ending that holder permits actual takeover/provisioning while disjoint
// protection survives. No initial OS link failure or native fallback is claimed.
func TestHomeLeaseProvisionAndTakeover(t *testing.T) {
	for _, mode := range []string{"takeover", "first-provision"} {
		t.Run(mode, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			checkDisjoint := preserveDisjointHomeLease(t, fx)
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			if mode == "takeover" {
				req.Policy.Takeover = true
				path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.EnvID), "AGENTS.md")
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("takeover fixture drift"), 0600); err != nil {
					t.Fatal(err)
				}
				// A valid older marker does not own this newly desired surface.
				root := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
				marker, err := envmarker.Read(root)
				if err != nil {
					t.Fatal(err)
				}
				delete(marker.Surfaces, envmarker.SurfaceRootContext)
				raw, err := marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, envmarker.Name), raw, 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := envmarker.Read(root); err != nil {
					t.Fatalf("takeover marker invalid: %v", err)
				}
			} else {
				req.EnvID = envregistry.ClaudeCode
				if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, req.EnvID)); !os.IsNotExist(err) {
					t.Fatal("first home already provisioned", err)
				}
				_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				member, ok := lock.Find("skill", "myskill")
				if !ok {
					t.Fatal("shared dependency missing")
				}
				if err := os.WriteFile(filepath.Join(contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey()), "SKILL.md"), []byte("first provision shared drift"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := captureFixtureState(t, fx)
			_, err := Resolve(req)
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("provision/takeover refusal changed preimage")
			}
			breakFixtureLease(t, m, lease)
			if mode == "takeover" {
				withoutConsent := req
				withoutConsent.Policy.Takeover = false
				before, _ := captureFixtureState(t, fx)
				_, err := Resolve(withoutConsent)
				if err == nil || !strings.Contains(err.Error(), DiagUnmanagedConflict) {
					t.Fatalf("takeover consent not required: %v", err)
				}
				after, _ := captureFixtureState(t, fx)
				if !bytes.Equal(before, after) {
					t.Fatal("missing takeover consent changed preimage")
				}
			}
			if _, err := Resolve(req); err != nil {
				t.Fatalf("unleased provision/takeover: %v", err)
			}
			if mode == "takeover" {
				backups, err := filepath.Glob(filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.EnvID), ".agent-environment-backup", "*", "AGENTS.md"))
				if err != nil || len(backups) != 1 {
					t.Fatalf("takeover backup inventory: %v %v", backups, err)
				}
				raw, err := os.ReadFile(backups[0])
				if err != nil || string(raw) != "takeover fixture drift" {
					t.Fatal("takeover did not preserve original bytes", err)
				}
			}
			if _, err := Resolve(fx.request(req.EnvID)); err != nil {
				t.Fatalf("written home not current: %v", err)
			}
			if mode == "first-provision" {
				if _, err := os.Stat(ManagedHomeDir(fx.home, fx.profile, req.EnvID)); err != nil {
					t.Fatal("home not provisioned", err)
				}
			}
			checkDisjoint()
		})
	}
}

// W05: actual native selection distinguishes the old, new and scoped-clear
// targets. Each entry refuses before changing any fixture surface; after the
// relevant break, the pointer/native document really changes while an unrelated
// profile remains leased. Native homes and credentials are synthetic fixtures.
func TestHomeLeaseSelectionAliases(t *testing.T) {
	for _, action := range []string{"use-from", "use-to", "clear-scoped", "current", "scoped"} {
		t.Run(action, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			t.Setenv("CODEX_HOME", fx.native[envregistry.CodexCLI])
			checkDisjoint := preserveDisjointHomeLease(t, fx)
			source := t.TempDir()
			writePackage(t, source, "selection-target", "1.0.0", "selection fixture\n")
			if _, _, _, err := Install(fx.home, InstallOptions{Operand: source}); err != nil {
				t.Fatal(err)
			}
			other := *fx
			other.profile = "selection-target"
			provision := other.request(envregistry.CodexCLI)
			provision.Repair = true
			if _, err := Resolve(provision); err != nil {
				t.Fatal(err)
			}
			use := func(profile string, clear bool) error {
				rows, err := UseWithPolicy(fx.home, profile, envregistry.CodexCLI, "", clear, Policy{})
				if err == nil {
					if len(rows) != 1 || !rows[0].OK {
						t.Fatalf("native selection did not complete: %+v", rows)
					}
				}
				return err
			}
			if action == "use-to" || action == "clear-scoped" {
				if err := SetCurrent(fx.home, other.profile); err != nil {
					t.Fatal(err)
				}
			}
			initial := fx.profile
			if action == "use-to" {
				initial = other.profile
			}
			if err := use(initial, false); err != nil {
				t.Fatalf("initial native selection: %v", err)
			}
			initialBytes, err := os.ReadFile(filepath.Join(fx.native[envregistry.CodexCLI], "AGENTS.md"))
			if err != nil {
				t.Fatal(err)
			}
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			write := func() error {
				switch action {
				case "use-from":
					return use(other.profile, false)
				case "use-to":
					return use(fx.profile, false)
				case "clear-scoped":
					return use("", true)
				case "current":
					return SetCurrent(fx.home, other.profile)
				case "scoped":
					return SetScoped(fx.home, "env:codex_cli", other.profile, false)
				default:
					t.Fatal("unknown action")
					return nil
				}
			}
			before, _ := captureFixtureState(t, fx)
			requireHomeLeaseRefusal(t, write(), lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("selection refusal changed preimage")
			}
			breakFixtureLease(t, m, lease)
			before, _ = captureFixtureState(t, fx)
			if err := write(); err != nil {
				t.Fatalf("unleased selection: %v", err)
			}
			after, _ = captureFixtureState(t, fx)
			if bytes.Equal(before, after) {
				t.Fatal("selection control made no actual change")
			}
			current, err := Current(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			scoped, err := ScopedCurrents(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			switch action {
			case "current":
				if current != other.profile {
					t.Fatalf("machine selection unchanged: %q", current)
				}
			case "clear-scoped":
				if current != other.profile || scoped["env:codex_cli"] != "" {
					t.Fatalf("scoped clear did not restore machine: %q %+v", current, scoped)
				}
			case "use-to":
				if scoped["env:codex_cli"] != fx.profile {
					t.Fatalf("new target not selected: %+v", scoped)
				}
			default:
				if scoped["env:codex_cli"] != other.profile {
					t.Fatalf("other target not selected: %+v", scoped)
				}
			}
			if strings.HasPrefix(action, "use-") || action == "clear-scoped" {
				actual, err := os.ReadFile(filepath.Join(fx.native[envregistry.CodexCLI], "AGENTS.md"))
				want := []byte("selection fixture")
				if action == "use-to" {
					want = []byte("hello")
				}
				if err != nil || bytes.Equal(actual, initialBytes) || !bytes.Contains(actual, want) {
					t.Fatal("native context did not switch to intended profile", err)
				}
			}
			checkDisjoint()
		})
	}
}

// W12: real missing credential-link repair refuses before writes under a
// matching lease, then relinks the exact native fixture target after break.
// An unowned regular-file conflict still refuses without changing state.
// The Codex tool-owned seed and marker inventory survive native seed drift;
// Muse seeds originate in its locked context, so no native-seed claim is made.
// A disjoint lease survives all controls. No real credential/Keychain is used.
func TestHomeLeaseCredentialLinkRepair(t *testing.T) {
	for _, tc := range []struct{ environment, link, target, seed, initial, changed string }{
		{envregistry.CodexCLI, "auth.json", "auth.json", "config.toml", "operator = true\n", "operator = false\n"},
		{envregistry.Muse, "config/muse/auth.json", "auth.json", "", "", ""},
	} {
		t.Run(tc.environment, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx := writeManagedFixture(t, "credential-fixture")
			seedLiveNativeCredentials(t, fx)
			if tc.environment == envregistry.Muse {
				fx.native[envregistry.Muse] = t.TempDir()
				if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], tc.target), []byte("synthetic-native\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.seed != "" {
				if err := os.WriteFile(filepath.Join(fx.native[tc.environment], tc.seed), []byte(tc.initial), 0600); err != nil {
					t.Fatal(err)
				}
			}
			req := fx.request(tc.environment)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			checkDisjoint := preserveDisjointHomeLease(t, fx)
			root := ManagedHomeDir(fx.home, fx.profile, tc.environment)
			link := filepath.Join(root, filepath.FromSlash(tc.link))
			target := filepath.Join(fx.native[tc.environment], filepath.FromSlash(tc.target))
			if got, err := os.Readlink(link); err != nil || got != target {
				t.Fatalf("initial credential target %q, err %v", got, err)
			}
			marker := readManagedMarker(t, fx, tc.environment)
			recorded := false
			if marker.Passthrough != nil {
				for _, row := range *marker.Passthrough {
					if row.Path == tc.link && row.Backend == "file" && (row.Strategy == envregistry.StrategyFileLink || (tc.environment == envregistry.CodexCLI && row.Strategy == envregistry.StrategyKeyringPreferred)) {
						recorded = true
					}
				}
			}
			if !recorded {
				t.Fatal("fixture does not have a recorded file-link")
			}
			if tc.seed != "" {
				seed, err := os.ReadFile(filepath.Join(root, tc.seed))
				if err != nil || string(seed) != tc.initial {
					t.Fatal("initial tool seed not provisioned", err)
				}
			}
			m, lease := leaseManagedFixture(t, fx, tc.environment)
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			if tc.seed != "" {
				if err := os.WriteFile(filepath.Join(fx.native[tc.environment], tc.seed), []byte(tc.changed), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := captureFixtureState(t, fx)
			_, err := Resolve(req)
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("leased credential repair changed preimage")
			}
			breakFixtureLease(t, m, lease)
			// Existing credential ownership remains independent of lease admission.
			if err := os.WriteFile(link, []byte("unowned-synthetic\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before, _ = captureFixtureState(t, fx)
			_, err = Resolve(req)
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
				t.Fatalf("unowned credential not refused: %v", err)
			}
			after, _ = captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("unowned credential conflict changed preimage")
			}
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatalf("unleased credential repair: %v", err)
			}
			if got, err := os.Readlink(link); err != nil || got != target {
				t.Fatalf("repaired credential target %q, err %v", got, err)
			}
			repaired := readManagedMarker(t, fx, tc.environment)
			if !reflect.DeepEqual(marker.Seeds, repaired.Seeds) || !reflect.DeepEqual(marker.CodexSeedRecord, repaired.CodexSeedRecord) {
				t.Fatal("credential repair changed seed inventory")
			}
			if repaired.Passthrough == nil || len(*marker.Passthrough) != len(*repaired.Passthrough) {
				t.Fatal("credential repair changed credential inventory")
			}
			for i, row := range *repaired.Passthrough {
				before := (*marker.Passthrough)[i]
				if before.Provenance != "provisioned" || row.Provenance != "repaired" {
					t.Fatal("credential marker did not record the actual publication transition")
				}
				before.Provenance, row.Provenance = "", ""
				if !reflect.DeepEqual(before, row) {
					t.Fatal("credential backend/target record changed")
				}
			}
			if tc.seed != "" {
				seed, err := os.ReadFile(filepath.Join(root, tc.seed))
				if err != nil || string(seed) != tc.initial {
					t.Fatal("credential repair refreshed a tool-owned seed", err)
				}
			}
			if _, err := Resolve(fx.request(tc.environment)); err != nil {
				t.Fatalf("repaired home not current: %v", err)
			}
			checkDisjoint()
		})
	}
}

// W06: actual install/import --use and automatic first activation refuse
// before publishing a new profile when they affect a leased old profile or
// a physical native-home alias. Nonactivating installs/imports and the same
// activation after break actually publish; an aliased target is then named
// directly to satisfy the independent native route gate. Disjoint protection stays.
// All import/native inputs are synthetic; no provider is launched.
func TestHomeLeaseInstallImportActivation(t *testing.T) {
	for _, operation := range []string{"install", "import"} {
		for _, activation := range []string{"use", "first-alias"} {
			t.Run(operation+"/"+activation, func(t *testing.T) {
				fx, m, lease := leasedManagedFixture(t)
				checkDisjoint := preserveDisjointHomeLease(t, fx)
				source := t.TempDir()
				writePackage(t, source, "activation-target", "1.0.0", "activation fixture\n")
				seams := pinImportSeams(t)
				for _, adapter := range Adapters {
					writeNativeFile(t, seams.native[adapter.ID], adapter.Target, "activation fixture\n")
				}
				write := func(name string, use bool) (Info, bool, error) {
					if operation == "import" {
						opts := seams.options(Policy{})
						opts.As, opts.Use = name, use
						info, activated, _, err := Import(fx.home, opts)
						return info, activated, err
					}
					info, activated, _, err := Install(fx.home, InstallOptions{Operand: source, As: name, Use: use})
					return info, activated, err
				}
				// A genuinely disjoint import/install must not require breaking
				// either existing reservation or switching the native scope.
				protected := captureTree(t, ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))
				info, activated, err := write("nonactivating", false)
				if err != nil || activated || info.Name != "nonactivating" {
					t.Fatalf("nonactivating control: name=%s activated=%t err=%v", info.Name, activated, err)
				}
				if !reflect.DeepEqual(protected, captureTree(t, ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))) {
					t.Fatal("disjoint install changed protected home")
				}
				rows, err := m.List(context.Background(), envregistry.CodexCLI, fx.profile)
				if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0], *lease) {
					t.Fatal("disjoint install changed target reservation", err)
				}
				checkDisjoint()
				if activation == "first-alias" {
					if err := os.Remove(CurrentFile(fx.home)); err != nil {
						t.Fatal(err)
					}
					alias := filepath.Join(t.TempDir(), "native-alias")
					if err := os.Symlink(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), alias); err != nil {
						t.Fatal(err)
					}
					t.Setenv("CODEX_HOME", alias)
				}
				before, _ := captureFixtureState(t, fx)
				_, _, err = write("activation-target", activation == "use")
				requireHomeLeaseRefusal(t, err, lease.LeaseID)
				after, _ := captureFixtureState(t, fx)
				if !bytes.Equal(before, after) {
					t.Fatal("activation refusal published partial profile/native state")
				}
				if _, err := os.Lstat(ProfileDir(fx.home, "activation-target")); !os.IsNotExist(err) {
					t.Fatal("refused profile was published", err)
				}
				breakFixtureLease(t, m, lease)
				if activation == "first-alias" {
					// The native writer independently refuses a symlinked home.
					// Its valid positive uses the same physical target directly.
					t.Setenv("CODEX_HOME", ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI))
				}
				info, activated, err = write("activation-target", activation == "use")
				if err != nil || !activated || info.Name != "activation-target" {
					t.Fatalf("activation control: name=%s activated=%t entries=%+v err=%v", info.Name, activated, info.Activation, err)
				}
				if len(info.Activation) != len(Adapters) {
					t.Fatalf("activation entry count %d", len(info.Activation))
				}
				for _, row := range info.Activation {
					if !row.OK {
						t.Fatalf("activation failed: %+v", row)
					}
				}
				current, err := Current(fx.home)
				if err != nil || current != "activation-target" {
					t.Fatalf("current=%q, err=%v", current, err)
				}
				if _, _, _, err := loadResolveInputs(fx.home, "activation-target"); err != nil {
					t.Fatal("actual installed lock/source missing", err)
				}
				for _, adapter := range Adapters {
					native, err := NativeHome(adapter)
					if err != nil {
						t.Fatal(err)
					}
					payload, err := os.ReadFile(filepath.Join(native, adapter.Target))
					if err != nil || !bytes.Contains(payload, []byte("activation fixture")) {
						t.Fatal("native adapter did not materialize intended context", adapter.ID, err)
					}
					marker, err := envmarker.Read(native)
					if err != nil || marker == nil || marker.Profile.Name != "activation-target" {
						t.Fatal("native marker does not match selection", adapter.ID, err)
					}
				}
				checkDisjoint()
			})
		}
	}
}

// W10: damage to actual manager-rendered documents cannot be repaired under
// the matching lease, even though their managed links are unchanged. The
// full rendered batch is restored after break while a disjoint lease stays.
// Claude's copied root and immutable referenced modules are outside this test;
// its real MCP document is included for both admitted forms. No provider runs.
func TestHomeLeaseRenderedDocuments(t *testing.T) {
	for _, tc := range []struct{ environment, form string }{
		{envregistry.CodexCLI, envregistry.FormMonolithic},
		{envregistry.ClaudeCode, envregistry.FormMonolithic},
		{envregistry.ClaudeCode, envregistry.FormReferenced},
		{envregistry.Muse, envregistry.FormMonolithic},
	} {
		t.Run(tc.environment+"/"+tc.form, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx := writeManagedFixture(t, "render-fixture")
			seedLiveNativeCredentials(t, fx)
			if tc.environment == envregistry.Muse {
				fx.native[envregistry.Muse] = t.TempDir()
				if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic-native\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			req := fx.request(tc.environment)
			req.Machine.Forms = map[string]string{tc.environment: tc.form}
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			checkDisjoint := preserveDisjointHomeLease(t, fx)
			m, err := HomeLeaseManager(req, nil)
			if err != nil {
				t.Fatal(err)
			}
			request := homelease.Request{Environment: tc.environment, Profile: fx.profile, LockSHA256: strings.TrimPrefix(fx.lockHash, "sha256:"), RequestID: "render-fixture", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "render-fixture"}}
			out, reason, err := m.Acquire(context.Background(), request)
			if err != nil || reason != "acquired" || !out.Changed || out.Lease == nil {
				t.Fatalf("render admission: %+v %s %v", out, reason, err)
			}
			check := req
			check.Repair = false
			before, _ := captureFixtureState(t, fx)
			if _, err := Resolve(check); err != nil {
				t.Fatal("current leased render read", err)
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("current render read wrote fixture")
			}
			originals := map[string][]byte{}
			root := ManagedHomeDir(fx.home, fx.profile, tc.environment)
			marker := readManagedMarker(t, fx, tc.environment)
			for _, surface := range marker.Surfaces {
				for _, rel := range surface.Paths {
					live := filepath.Join(root, filepath.FromSlash(rel))
					target, err := os.Readlink(live)
					if err != nil {
						continue
					}
					if !filepath.IsAbs(target) {
						target = filepath.Join(filepath.Dir(live), target)
					}
					// Select only the known rendered path for this surface, never
					// an immutable pin or credential/native target.
					target = filepath.Clean(target)
					if target != storeDocPath(fx.home, fx.profile, tc.environment, rel) {
						continue
					}
					payload, err := os.ReadFile(target)
					if err != nil {
						t.Fatal(err)
					}
					originals[target] = payload
					if err := os.WriteFile(target, []byte("rendered fixture drift\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(originals) == 0 {
				t.Fatal("fixture exercised no actual rendered document")
			}
			before, _ = captureFixtureState(t, fx)
			_, err = Resolve(req)
			requireHomeLeaseRefusal(t, err, out.Lease.LeaseID)
			after, _ = captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("leased rendered repair changed preimage")
			}
			breakFixtureLease(t, m, out.Lease)
			if _, err := Resolve(req); err != nil {
				t.Fatal("unleased rendered repair", err)
			}
			for path, want := range originals {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatal("rendered batch not restored", err)
				}
			}
			if _, err := Resolve(check); err != nil {
				t.Fatal("restored render not current", err)
			}
			t.Logf("qualified %d actual rendered documents", len(originals))
			checkDisjoint()
		})
	}
}

// W13/W17: actual interrupted apply and interrupted rollback leave durable
// custom migration journals. Replant their exact synthetic postimages around
// a real lease admission: public apply and borrowed recovery must refuse before
// any writes. After explicit break, real recovery restores prior link/marker,
// removes intent and admits a fresh replayable reservation. This fault plant
// does not claim that guarded production can create a journal under a lease,
// or that conservative whole-manager legacy recovery is disjoint-qualified.
func TestHomeLeaseInterruptedCustomMigration(t *testing.T) {
	for _, phase := range []string{"forward", "rollback"} {
		t.Run(phase, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, _ := captureFixture(t)
			root := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
			link := filepath.Join(root, "auth.json")
			priorTarget, err := os.Readlink(link)
			if err != nil {
				t.Fatal(err)
			}
			markerPath := filepath.Join(root, envmarker.Name)
			priorMarker, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}
			req := fx.migrateRequest()
			req.Profile, req.EnvID = fx.profile, envregistry.CodexCLI
			req.Machine.Isolation = map[string]map[string]string{fx.profile: {envregistry.CodexCLI: envregistry.IsolationIsolated}}
			plan, err := PlanMigration(req)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Ops()) != 1 || plan.Ops()[0].Kind != MigrateOpUnlink {
				t.Fatal("fixture lacks actual credential unlink plan")
			}
			req.Expect = plan.Hash
			rollbackStarted, rollbackFailures := false, 0
			req.InjectFault = func(point string, applied int) error {
				if phase == "forward" && point == MigrateFaultCrash && applied == 1 {
					return errors.New("fixture forward interruption")
				}
				if phase == "rollback" && point == MigrateFaultPublish {
					rollbackStarted = true
					return errors.New("fixture publication interruption")
				}
				return nil
			}
			req.symlink = func(target, path string) error {
				if rollbackStarted {
					rollbackFailures++
					return errors.New("fixture rollback interruption")
				}
				return os.Symlink(target, path)
			}
			_, err = ApplyMigration(req)
			if err == nil || (phase == "forward" && !strings.Contains(err.Error(), "interrupted")) || (phase == "rollback" && (!strings.Contains(err.Error(), "rollback incomplete") || rollbackFailures != 1)) {
				t.Fatalf("intended migration fault not reached: %v", err)
			}
			journal, err := readMigrationJournal(fx.home)
			if err != nil || journal == nil || journal.Complete || len(journal.Ops) != 1 || !journal.Ops[0].Done {
				t.Fatalf("real interruption journal absent: %+v %v", journal, err)
			}
			pending, err := os.ReadFile(migrationJournalPath(fx.home))
			if err != nil {
				t.Fatal(err)
			}
			pendingMarker, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(link); !os.IsNotExist(err) {
				t.Fatal("real interrupted unlink not exercised", err)
			}
			// Fault setup restores currentness only to obtain an actual lease;
			// no pending state is treated as safe by the admission implementation.
			if err := os.Remove(migrationJournalPath(fx.home)); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(priorTarget, link); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(markerPath, priorMarker, 0600); err != nil {
				t.Fatal(err)
			}
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(markerPath, pendingMarker, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(migrationJournalPath(fx.home), pending, 0600); err != nil {
				t.Fatal(err)
			}
			req.InjectFault, req.symlink = nil, nil
			before, _ := captureFixtureState(t, fx)
			_, err = ApplyMigration(req)
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			linkCalls := 0
			recover := func() (*MigrateRecovery, error) {
				op, err := beginOperationGuarded(fx.home, nil, []string{})
				if err != nil {
					return nil, err
				}
				defer func() { _ = op.close() }()
				return recoverMigrationJournal(fx.home, op, nil, func(target, path string) error { linkCalls++; return os.Symlink(target, path) }, os.Rename)
			}
			_, err = recover()
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if linkCalls != 0 || !bytes.Equal(before, after) {
				t.Fatal("leased custom recovery changed preimage or entered link syscall")
			}
			if err := os.Remove(migrationJournalPath(fx.home)); err != nil {
				t.Fatal(err)
			}
			breakFixtureLease(t, m, lease)
			if err := os.WriteFile(migrationJournalPath(fx.home), pending, 0600); err != nil {
				t.Fatal(err)
			}
			// Recovery inventory must still reject a foreign credential entry
			// after the independent lease gate is gone.
			if err := os.WriteFile(link, []byte("synthetic-unowned\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before, _ = captureFixtureState(t, fx)
			_, err = recover()
			if err == nil || !strings.Contains(err.Error(), "recovery found") || !strings.Contains(err.Error(), "regular file") {
				t.Fatalf("foreign recovery entry not refused: %v", err)
			}
			after, _ = captureFixtureState(t, fx)
			if linkCalls != 0 || !bytes.Equal(before, after) {
				t.Fatal("foreign inventory refusal changed preimage")
			}
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			recovered, err := recover()
			if err != nil || recovered == nil || recovered.Plan != plan.Hash || recovered.Done != 1 || recovered.Ops != 1 || linkCalls != 1 {
				t.Fatalf("actual recovery control: %+v calls=%d %v", recovered, linkCalls, err)
			}
			if got, err := os.Readlink(link); err != nil || got != priorTarget {
				t.Fatal("prior credential link not restored", err)
			}
			if got, err := os.ReadFile(markerPath); err != nil || !bytes.Equal(got, priorMarker) {
				t.Fatal("prior marker not restored", err)
			}
			if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
				t.Fatal("custom journal not removed", err)
			}
			if _, err := Resolve(fx.request(envregistry.CodexCLI)); err != nil {
				t.Fatal("recovered home not current", err)
			}
			fresh := homelease.Request{Environment: envregistry.CodexCLI, Profile: fx.profile, LockSHA256: strings.TrimPrefix(fx.lockHash, "sha256:"), RequestID: "after-recovery", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "after-recovery"}}
			admitted, reason, err := m.Acquire(context.Background(), fresh)
			if err != nil || reason != "acquired" || !admitted.Changed {
				t.Fatalf("recovery did not admit fresh acquire: %s %v", reason, err)
			}
			replayed, _, err := m.Acquire(context.Background(), fresh)
			if err != nil || replayed.Changed || replayed.Lease == nil || admitted.Lease == nil || replayed.Lease.LeaseID != admitted.Lease.LeaseID {
				t.Fatal("recovered acquire replay changed reservation", err)
			}
		})
	}
}

// W08/W18: standalone and borrowed real global additions refuse before
// publication under a matching lease. After break, actual Git-backed skill
// projection/declaration succeeds; a failed borrowed continuation restores
// mutable profile/native/declaration preimages and permits a real retry.
// Immutable admitted entries may remain. Legacy whole-manager disjoint
// availability is not claimed. The repository transport and homes are fixtures.
func TestHomeLeaseGlobalAddAndCompensation(t *testing.T) {
	for _, mode := range []string{"standalone", "borrowed", "compensate"} {
		t.Run(mode, func(t *testing.T) {
			fx, m, lease := leasedManagedFixture(t)
			globalRoot := filepath.Join(fx.home, "global")
			if err := os.MkdirAll(globalRoot, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := manifest.EnsureEmpty(globalRoot); err != nil {
				t.Fatal(err)
			}
			originalDeclaration, err := os.ReadFile(manifest.PathIn(globalRoot))
			if err != nil {
				t.Fatal(err)
			}
			ids := newGitIdentities(t)
			repo := gitRepo(t, map[string]string{"SKILL.md": "# Global fixture\n"}, "v1.0.0")
			commit := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
			operand := ids.serve(repo, "https://example.com/skills/global-added")
			addition := DirectSkill{Name: "global-added", Git: operand, Revision: commit}
			called := 0
			planted := errors.New("fixture legacy continuation failed")
			failContinuation := mode == "compensate"
			assertDeclaration := func() {
				t.Helper()
				decl, err := manifest.Load(filepath.Join(fx.home, "global"))
				if err != nil || decl == nil || len(decl.Skills) != 1 || decl.Skills[0].Name != addition.Name || decl.Skills[0].Ref.Kind != "revision" || decl.Skills[0].Ref.Value != commit {
					t.Fatalf("actual added declaration missing: %+v %v", decl, err)
				}
			}
			write := func() (Info, error) {
				if mode == "standalone" {
					return GlobalAdd(fx.home, addition, Policy{}, envregistry.DefaultMachineConfig(), fx.request(envregistry.CodexCLI).NativeHomeOf)
				}
				locks, err := managerlock.New(fx.home)
				if err != nil {
					return Info{}, err
				}
				lock, err := locks.AcquireHomeOnly(context.Background(), false)
				if err != nil {
					return Info{}, err
				}
				defer func() { _ = lock.Close() }()
				return GlobalSkillTransaction(fx.home, Policy{}, envregistry.DefaultMachineConfig(), &addition, fx.request(envregistry.CodexCLI).NativeHomeOf, lock, func() error {
					called++
					assertDeclaration()
					if failContinuation {
						return planted
					}
					return nil
				})
			}
			before, _ := captureFixtureState(t, fx)
			_, err = write()
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if called != 0 || !bytes.Equal(before, after) {
				t.Fatal("global-add refusal changed preimage or entered continuation")
			}
			breakFixtureLease(t, m, lease)
			if mode == "compensate" {
				mutable := func() map[string]map[string]string {
					v := map[string]map[string]string{"profiles": captureTree(t, ProfilesDir(fx.home)), "managed": captureTree(t, EnvRoot(fx.home))}
					for id, root := range fx.native {
						v[id] = captureTree(t, root)
					}
					return v
				}
				prior := mutable()
				_, err := write()
				var compensation *GlobalCompensationError
				if !errors.Is(err, planted) || errors.As(err, &compensation) || called != 1 {
					t.Fatalf("intended compensated failure missing: calls=%d %v", called, err)
				}
				if !reflect.DeepEqual(prior, mutable()) {
					t.Fatal("global compensation did not restore mutable profile/native preimages")
				}
				if got, err := os.ReadFile(manifest.PathIn(globalRoot)); err != nil || !bytes.Equal(got, originalDeclaration) {
					t.Fatal("global compensation did not restore original declaration", err)
				}
				current, err := Current(fx.home)
				if err != nil || current != fx.profile {
					t.Fatal("compensation changed selection", err)
				}
				failContinuation = false
			}
			info, err := write()
			if err != nil || info.Lock == nil {
				t.Fatal("unleased global addition failed", err)
			}
			member, ok := info.Lock.Find("skill", addition.Name)
			if !ok || member.Commit != commit {
				t.Fatal("real global addition did not extend lock")
			}
			wantCalls := 0
			if mode == "borrowed" {
				wantCalls = 1
			}
			if mode == "compensate" {
				wantCalls = 2
			}
			if called != wantCalls {
				t.Fatalf("continuation calls %d want %d", called, wantCalls)
			}
			if mode != "standalone" {
				assertDeclaration()
			}
			for _, adapter := range Adapters {
				native, err := NativeHome(adapter)
				if err != nil {
					t.Fatal(err)
				}
				payload, err := os.ReadFile(filepath.Join(native, inPlaceSkillsDir(adapter.ID), addition.Name, "SKILL.md"))
				if err != nil || string(payload) != "# Global fixture\n" {
					t.Fatal("actual native skill projection missing", adapter.ID, err)
				}
			}
			payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "skills", addition.Name, "SKILL.md"))
			if err != nil || string(payload) != "# Global fixture\n" {
				t.Fatal("actual managed skill projection missing", err)
			}
		})
	}
}
