package envprofile

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/homelease"
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

// W03/13: a genuinely v1 home is admitted before enabling v2 writers. Both
// resolve migration and explicit migration refuse, then publish v2 identities
// after break. No credential payload is read by the lease guard.
func TestHomeLeaseIdentityMigration(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "resolve", true: "explicit"}[explicit], func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx := legacyIdentityFixture(t, envregistry.CodexCLI)
			hashing.EnableV2Writers = false
			m, lease := leaseManagedFixture(t, fx, envregistry.CodexCLI)
			hashing.EnableV2Writers = true
			write := func() error {
				if explicit {
					req := fx.migrateRequest()
					req.Profile, req.EnvID = fx.profile, envregistry.CodexCLI
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
