package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/skillintegrity"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/startupverification"
)

func verificationCandidate(t *testing.T) (string, *contextresolve.Result) {
	t.Helper()
	home := gitFixtureHome(t)
	source := "github.com/example/verified-skill"
	commit := writeGitStoreFixture(t, home, "skill", "verified-skill", source, map[string]string{
		"SKILL.md":  "---\nname: verified-skill\ndescription: A valid skill.\n---\nContext.\n",
		"README.md": "Complete evidence includes this reference.\n",
	})
	contextSource := "github.com/example/verified-context"
	contextCommit := writeGitStoreFixture(t, home, "context", "sample", contextSource, map[string]string{"agent-context.json": `{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[]}}`})
	contextMember := contextlock.Member{Kind: "context", Name: "sample", Source: contextSource, Version: "1.0.0", Commit: contextCommit, RequiredBy: []string{}}
	member := contextlock.Member{Kind: "skill", Name: "verified-skill", Source: source, Commit: commit, RequiredBy: []string{"sample"}}
	return home, &contextresolve.Result{
		Lock:    &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Root: "sample", Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Members: []contextlock.Member{contextMember, member}},
		Members: map[string]contextresolve.Resolved{contextresolve.Key("context", "sample"): resolvedOf(contextMember), contextresolve.Key("skill", "verified-skill"): resolvedOf(member)},
	}
}

// Explicit adoption creates complete frozen evidence without writing a native
// declaration into the source snapshot. Failed admission leaves the candidate.
func TestPrepareStartupEvidence(t *testing.T) {
	for _, bad := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "forbidden-link"}[bad], func(t *testing.T) {
			home, result := verificationCandidate(t)
			before := result.Lock
			entry := memberEntryPath(home, before.Members[1])
			if bad {
				requireLinkCapability(t)
				if err := os.Symlink("SKILL.md", filepath.Join(entry, "late")); err != nil {
					t.Fatal(err)
				}
			}
			err := prepareStartupEvidence(home, newGitManager(home), result, Policy{})
			if bad {
				if !skillintegrity.IsCode(err, "skill_integrity_boundary_invalid") || result.Lock != before || result.LockHash != "" {
					t.Fatalf("refusal changed candidate: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Lock.SchemaVersion != 6 || len(result.Lock.SkillIntegrity) != 1 || result.LockHash == "" {
				t.Fatal("evidence not frozen")
			}
			if len(result.Lock.SkillIntegrity[0].Manifest.Files) != 2 {
				t.Fatal("reference was omitted")
			}
			if _, err := os.Lstat(filepath.Join(entry, "agent-skill.json")); !os.IsNotExist(err) {
				t.Fatal("source declaration was written")
			}
		})
	}
}

// A selected context floor is retained; a lower explicit floor refuses rather
// than being silently replaced by the machine baseline.
func TestPrepareStartupFloors(t *testing.T) {
	for _, floor := range []int{2, 3, 4} {
		home, result := verificationCandidate(t)
		selected := result.Members[contextresolve.Key("context", "sample")]
		selected.Package = &contextresolve.Package{StartupVerificationLevel: &floor}
		result.Members[contextresolve.Key("context", "sample")] = selected
		before := result.Lock
		err := prepareStartupEvidence(home, newGitManager(home), result, Policy{StartupVerificationLevel: 3})
		if floor == 2 {
			var diagnostic *startupverification.Diagnostic
			if !errors.As(err, &diagnostic) || diagnostic.Code != "startup_verification_weakening" || result.Lock != before {
				t.Fatal("weakening accepted or candidate changed")
			}
			continue
		}
		if err != nil || result.Lock.StartupVerificationLevel == nil || *result.Lock.StartupVerificationLevel != floor {
			t.Fatalf("floor %d: %v", floor, err)
		}
		if level, err := startupverification.EffectiveLevel(3, *result.Lock.StartupVerificationLevel); err != nil || level != floor {
			t.Fatalf("floor lost: %v", err)
		}
	}
}

// Real installed pins and frozen skill evidence drive every level. A definite
// content mismatch is warn-only at1 and refuses2/3/4 without repairing evidence.
func TestInstalledStartupEvidenceLevels(t *testing.T) {
	for _, mutate := range []bool{false, true} {
		t.Run(fmt.Sprint(mutate), func(t *testing.T) {
			home, result := verificationCandidate(t)
			if err := prepareStartupEvidence(home, newGitManager(home), result, Policy{}); err != nil {
				t.Fatal(err)
			}
			frozen, err := result.Lock.Canonical()
			if err != nil {
				t.Fatal(err)
			}
			if mutate {
				writeGitFile(t, memberEntryPath(home, result.Lock.Members[1]), "README.md", "changed reference")
			}
			checks, err := startupChecks(home, result.Lock)
			if err != nil {
				t.Fatal(err)
			}
			for level := 1; level <= 4; level++ {
				v, err := startupVerdict(Policy{StartupVerificationLevel: level}, result.Lock, checks)
				want := !mutate || level == 1
				if err != nil || v.Allow != want || mutate && (v.Verified || level == 1 && len(v.Warnings) == 0) {
					t.Fatalf("level %d: %+v %v", level, v, err)
				}
				if level == 4 && !mutate && (len(v.Warnings) != 1 || v.Warnings[0] != "verification_centre_unavailable") {
					t.Fatal("fallback warning missing")
				}
			}
			after, err := result.Lock.Canonical()
			if err != nil || !bytes.Equal(frozen, after) {
				t.Fatal("startup changed frozen evidence")
			}
		})
	}
}

// Legacy native profiles keep their original lock bytes and old pin algorithm.
// Manifestless legacy profiles remain unavailable until explicit adoption.
func TestLegacyStartupDoesNotSynthesize(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprint(native), func(t *testing.T) {
			home, result := verificationCandidate(t)
			if native {
				old := result.Lock.Members[1]
				commit := writeGitStoreFixture(t, home, "skill", old.Name, old.Source, map[string]string{"SKILL.md": "# Existing legacy skill\n", "agent-skill.json": `{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}`})
				result.Lock.Members[1].Commit = commit
			}
			before, err := result.Lock.Canonical()
			if err != nil {
				t.Fatal(err)
			}
			checks, err := startupChecks(home, result.Lock)
			if err != nil {
				t.Fatal(err)
			}
			for level := 1; level <= 4; level++ {
				v, err := startupVerdict(Policy{StartupVerificationLevel: level}, result.Lock, checks)
				if err != nil || v.Allow != (native || level < 3) {
					t.Fatalf("native %v level %d: %+v %v", native, level, v, err)
				}
				if native && level < 4 && len(v.Warnings) != 0 {
					t.Fatal("native legacy gained a warning")
				}
			}
			after, err := result.Lock.Canonical()
			if err != nil || !bytes.Equal(before, after) || result.Lock.SchemaVersion != 4 || result.Lock.SkillIntegrity != nil {
				t.Fatal("legacy evidence silently upgraded")
			}
		})
	}
}

// The real launcher's unconditional repair flag cannot rewrite a mismatch at
// permissive levels. Level1 emits warnings; level2 refuses the same contradiction.
func TestPermissiveResolvePreservesMismatchingEvidence(t *testing.T) {
	requireLinkCapability(t)
	for _, level := range []int{1, 2} {
		for _, repair := range []bool{false, true} {
			t.Run(fmt.Sprintf("level%d/repair%v", level, repair), func(t *testing.T) {
				fx := writeManagedFixture(t, "sample")
				seedLiveNativeCredentials(t, fx)
				provision := fx.request("codex_cli")
				provision.Repair = true
				if _, err := Resolve(provision); err != nil {
					t.Fatal(err)
				}
				lock, _, err := readLock(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
				if !ok {
					t.Fatal("missing control member")
				}
				changed := filepath.Join(memberEntryPath(fx.home, member), "unused", "pin-only.txt")
				edited := []byte("modified verification evidence\n")
				if err := os.WriteFile(changed, edited, 0o644); err != nil {
					t.Fatal(err)
				}
				markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
				beforeMarker, err := os.ReadFile(markerPath)
				if err != nil {
					t.Fatal(err)
				}
				beforeLock, err := os.ReadFile(lockPath(fx.home, fx.profile))
				if err != nil {
					t.Fatal(err)
				}
				req := fx.request("codex_cli")
				req.Repair = repair
				req.Policy.StartupVerificationLevel = level
				result, err := Resolve(req)
				if level == 1 {
					if err != nil || result == nil || len(result.Document) == 0 || len(result.Warnings) == 0 {
						t.Fatalf("warning launch: err=%v output=%v", err, result != nil)
					}
				} else if err == nil || result != nil && len(result.Document) != 0 {
					t.Fatalf("mismatch not refused: err=%v output=%v", err, result != nil)
				}
				for path, want := range map[string][]byte{changed: edited, markerPath: beforeMarker, lockPath(fx.home, fx.profile): beforeLock} {
					got, e := os.ReadFile(path)
					if e != nil || !bytes.Equal(want, got) {
						t.Fatalf("startup changed evidence %s: %v", filepath.Base(path), e)
					}
				}
			})
		}
	}
}

// Composition uses every profile check before visiting. Level1 may warn about
// an initial contradiction, but a source change during visiting always refuses.
func TestCompositionStartupPolicyAndGeneration(t *testing.T) {
	for _, mode := range []string{"valid", "initial-drift", "mid-visit-drift"} {
		for _, level := range []int{1, 2, 3, 4} {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				fx := writeManagedFixture(t, "sample")
				lock, hash, err := readLock(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
				if !ok {
					t.Fatal("missing member")
				}
				target := filepath.Join(memberEntryPath(fx.home, member), "unused", "pin-only.txt")
				if mode == "initial-drift" {
					if err := os.WriteFile(target, []byte("initial change"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				visits := 0
				warnings, err := VisitCompositionProfileWithPolicy(fx.home, fx.profile, strings.TrimPrefix(hash, "sha256:"), Policy{StartupVerificationLevel: level}, func(_ *contextlock.Lock, _ contextlock.Member, _ string) error {
					visits++
					if mode == "mid-visit-drift" && visits == 1 {
						return os.WriteFile(target, []byte("changed during composition"), 0o644)
					}
					return nil
				})
				if mode == "mid-visit-drift" {
					if err == nil || level == 1 && !strings.Contains(err.Error(), "source generation changed") {
						t.Fatalf("interrupted source did not receive the intended refusal: %v", err)
					}
					return
				}
				allowed := mode == "valid" || level == 1
				if allowed && (err != nil || visits != len(lock.Members)) || !allowed && (err == nil || visits != 0) {
					t.Fatalf("visits%d warnings%v: %v", visits, warnings, err)
				}
				if mode == "initial-drift" && level == 1 && len(warnings) == 0 {
					t.Fatal("missing mismatch warning")
				}
			})
		}
	}
}

// A payload link is a fatal boundary at every level, including repair and
// dry-run repair. Refusal preserves the link, lock and existing home marker;
// ordinary same-pin byte restoration is covered separately below and at docking.
func TestStartupBoundaryNeverBecomesWarning(t *testing.T) {
	requireLinkCapability(t)
	for _, kind := range []string{"symlink", "hardlink"} {
		for _, mode := range []string{"observe", "repair", "dry-run"} {
			for level := 1; level <= 4; level++ {
				t.Run(fmt.Sprintf("%s/%s/level%d", kind, mode, level), func(t *testing.T) {
					fx := writeManagedFixture(t, "sample")
					seedLiveNativeCredentials(t, fx)
					provision := fx.request("codex_cli")
					provision.Repair = true
					if _, err := Resolve(provision); err != nil {
						t.Fatal(err)
					}
					lock, _, err := readLock(fx.home, fx.profile)
					if err != nil {
						t.Fatal(err)
					}
					member, _ := lock.Find(contextlock.KindSkill, "myskill")
					plant := filepath.Join(memberEntryPath(fx.home, member), "forbidden")
					if kind == "symlink" {
						if err := os.Symlink("SKILL.md", plant); err != nil {
							t.Fatal(err)
						}
					} else if err := os.Link(filepath.Join(memberEntryPath(fx.home, member), "SKILL.md"), plant); err != nil {
						t.Fatal(err)
					}
					plantBefore, err := os.Lstat(plant)
					if err != nil {
						t.Fatal(err)
					}
					paths := []string{lockPath(fx.home, fx.profile), filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name), filepath.Join(memberEntryPath(fx.home, member), "SKILL.md")}
					before := map[string][]byte{}
					for _, path := range paths {
						data, err := os.ReadFile(path)
						if err != nil {
							t.Fatal(err)
						}
						before[path] = data
					}
					req := fx.request("codex_cli")
					req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
					req.Policy.StartupVerificationLevel = level
					repairReached := false
					req.beforeRepairLock = func() { repairReached = true }
					result, err := Resolve(req)
					if repairReached {
						t.Fatal("fatal boundary reached repair-lock seam")
					}
					var failure *storeEntryFailure
					rightClass := kind == "symlink" && errors.As(err, &failure) && failure.Check == "link_safety" || kind == "hardlink" && skillintegrity.IsCode(err, "skill_integrity_boundary_invalid")
					namesPath := err != nil && (strings.Contains(err.Error(), plant) || kind == "hardlink" && strings.Contains(err.Error(), filepath.Join(memberEntryPath(fx.home, member), "SKILL.md")))
					if !rightClass || !namesPath || result != nil && len(result.Document) > 0 {
						t.Fatalf("output=%v err=%v; expected named fatal payload-link refusal", result != nil, err)
					}
					plantAfter, err := os.Lstat(plant)
					if err != nil || !os.SameFile(plantBefore, plantAfter) {
						t.Fatalf("refusal erased or replaced the link: %v", err)
					}
					if kind == "symlink" {
						if target, err := os.Readlink(plant); err != nil || target != "SKILL.md" {
							t.Fatalf("refusal retargeted the link: %v", err)
						}
					}
					for path, want := range before {
						got, err := os.ReadFile(path)
						if err != nil || !bytes.Equal(got, want) {
							t.Fatalf("refusal changed %s: %v", filepath.Base(path), err)
						}
					}
				})
			}
		}
	}
}

// Strict repair can restore ordinary regular-file bytes from the same Git pin
// after structural checks; it preserves the profile lock and current marker.
func TestStrictStartupRepairRestoresRegularBytes(t *testing.T) {
	requireLinkCapability(t)
	for _, level := range []int{3, 4} {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			fx := writeManagedFixture(t, "sample")
			seedLiveNativeCredentials(t, fx)
			req := fx.request("codex_cli")
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
			if !ok {
				t.Fatal("missing pinned control member")
			}
			path := filepath.Join(memberEntryPath(fx.home, member), "unused", "pin-only.txt")
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			before := map[string][]byte{}
			for _, p := range []string{lockPath(fx.home, fx.profile), filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)} {
				data, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				before[p] = data
			}
			if err := os.WriteFile(path, []byte("changed ordinary bytes\n"), 0600); err != nil {
				t.Fatal(err)
			}
			req.Policy.StartupVerificationLevel = level
			result, err := Resolve(req)
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("same-pin repair refused: %v", err)
			}
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("same-pin bytes not restored: %v", err)
			}
			if failure := validateNamedStorePins(fx.home, lock); failure != nil {
				t.Fatalf("repair changed the pin: %v", failure)
			}
			for p, want := range before {
				got, err := os.ReadFile(p)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatalf("repair changed %s: %v", filepath.Base(p), err)
				}
			}
		})
	}
}

// Provisioning at1 carries its unavailable-evidence warning. A subsequent
// strict repair cannot bypass the missing-evidence refusal or synthesize a lock.
func TestRepairPreservesLegacyEvidenceDecision(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	lock.SchemaVersion = 4
	lock.HashVersion = 2
	lock.Precedence = contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}
	for i, member := range lock.Members {
		if member.Kind == contextlock.KindSkill {
			lock.Members[i].Source = "github.com/example/manifestless-legacy"
			lock.Members[i].Commit = writeGitStoreFixture(t, fx.home, "skill", member.Name, lock.Members[i].Source, map[string]string{"SKILL.md": "# Legacy context\n"})
			if skillspec.ManifestSourcePath(memberEntryPath(fx.home, lock.Members[i])) != "" {
				t.Fatal("manifest-absence plant did not apply")
			}
		}
	}
	if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	req.Policy.StartupVerificationLevel = 1
	result, err := Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 || len(result.Warnings) == 0 {
		t.Fatalf("permissive provisioning lost warning: has_result=%t err=%v", result != nil, err)
	}
	before, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
	beforeMarker, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	req.Policy.StartupVerificationLevel = 3
	result, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), "startup_verification_unavailable") || result != nil && len(result.Document) > 0 {
		t.Fatalf("strict repair bypass: has_result=%t err=%v", result != nil, err)
	}
	after, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("repair synthesized legacy evidence")
	}
	afterMarker, err := os.ReadFile(markerPath)
	if err != nil || !bytes.Equal(beforeMarker, afterMarker) {
		t.Fatal("refused repair changed marker")
	}
}

// A permissive provisioning request cannot trigger the old writer's implicit
// lock-identity migration. Existing lock4 provisioning is the nearby valid case.
func TestPermissiveProvisionDoesNotMigrateLock(t *testing.T) {
	requireLinkCapability(t)
	old := hashing.EnableV2Writers
	t.Cleanup(func() { hashing.EnableV2Writers = old })
	hashing.EnableV2Writers = false
	fx := writeManagedFixture(t, "legacy")
	seedLiveNativeCredentials(t, fx)
	hashing.EnableV2Writers = true
	before, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil || !repairWouldChangeLockHash(lock) {
		t.Fatalf("migration plant: %v", err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	req.Policy.StartupVerificationLevel = 1
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagRepairNeedsLockMigration) || result != nil && len(result.Document) > 0 {
		t.Fatalf("implicit migration: has_result=%v err=%v", result != nil, err)
	}
	after, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("permissive provisioning changed legacy lock")
	}
	if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)); !os.IsNotExist(err) {
		t.Fatal("refusal published a marker")
	}
}

// Real manifest3 path installation carries the declared floor through package
// resolution into the published lock; a weaker declaration publishes no profile.
// Real package parsing and installation retain an explicit floor, refuse a
// weakening before lock publication, and preserve the lock-v6 carrier.
func TestInstallStartupFloorBinding(t *testing.T) {
	pinOperatorHome(t, t.TempDir())
	_ = pinHomes(t)
	for _, floor := range []int{2, 3, 4} {
		t.Run(fmt.Sprint(floor), func(t *testing.T) {
			home := gitFixtureHome(t)
			if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
				t.Fatal(err)
			}
			source := t.TempDir()
			writeManifestPackage(t, source, fmt.Sprintf(`{"schema_version":3,"name":"floor-profile","version":"1.0.0","startup_verification":{"level":%d},"context":{"modules":[{"path":"guide.md"}]}}`, floor), map[string]string{"guide.md": "Profile instructions.\n"})
			info, _, _, err := Install(home, InstallOptions{Operand: source, Policy: Policy{StartupVerificationLevel: 3}})
			if floor == 2 {
				var diagnostic *startupverification.Diagnostic
				if !errors.As(err, &diagnostic) || diagnostic.Code != "startup_verification_weakening" {
					t.Fatalf("weak floor: %v", err)
				}
				if _, err := os.Lstat(lockPath(home, "floor-profile")); !os.IsNotExist(err) {
					t.Fatal("refused floor published a lock")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			published, _, err := readLock(home, "floor-profile")
			if err != nil || info.Lock.StartupVerificationLevel == nil || published.StartupVerificationLevel == nil || *published.StartupVerificationLevel != floor || published.SchemaVersion != 6 {
				t.Fatalf("published floor %d: %v", floor, err)
			}
		})
	}
}

// A selected overlay contributes its floor through real resolution; neither a
// weaker nor a stronger non-root declaration may be lost at the package seam.
func TestInstallStartupOverlayFloorBinding(t *testing.T) {
	pinOperatorHome(t, t.TempDir())
	_ = pinHomes(t)
	for _, floor := range []int{2, 4} {
		t.Run(fmt.Sprint(floor), func(t *testing.T) {
			home := gitFixtureHome(t)
			if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
				t.Fatal(err)
			}
			root, overlay := t.TempDir(), t.TempDir()
			writeManifestPackage(t, root, `{"schema_version":3,"name":"root-floor","version":"1.0.0","context":{"modules":[{"path":"guide.md"}]}}`, map[string]string{"guide.md": "Root.\n"})
			writeManifestPackage(t, overlay, fmt.Sprintf(`{"schema_version":3,"name":"overlay-floor","version":"1.0.0","startup_verification":{"level":%d},"context":{"modules":[{"path":"guide.md"}]}}`, floor), map[string]string{"guide.md": "Overlay.\n"})
			policy := Policy{StartupVerificationLevel: 3, OverlaysAllowed: true, OverlayDefaultWeight: 1000, Overlays: map[string][]OverlaySpec{"root-floor": {{Source: overlay}}}}
			info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
			if floor == 2 {
				var diagnostic *startupverification.Diagnostic
				if !errors.As(err, &diagnostic) || diagnostic.Code != "startup_verification_weakening" {
					t.Fatalf("overlay weakening: %v", err)
				}
				if _, err := os.Lstat(lockPath(home, "root-floor")); !os.IsNotExist(err) {
					t.Fatal("refused overlay published a lock")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if info.Lock.StartupVerificationLevel == nil || *info.Lock.StartupVerificationLevel != 4 || len(info.Lock.Members) != 2 {
				t.Fatal("overlay floor was not retained")
			}
		})
	}
}

// Evidence capture remains strict at level1: an intact source reaches the
// callback, while a definite source contradiction refuses before any output.
func TestCaptureStartupPolicyRemainsStrict(t *testing.T) {
	requireLinkCapability(t)
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			fx, req := captureFixture(t)
			req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
				return envregistry.DefaultMachineConfig(), Policy{StartupVerificationLevel: 1}, nil
			}
			if changed {
				lock, _, err := readLock(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
				if !ok {
					t.Fatal("missing MCP")
				}
				if err := os.WriteFile(filepath.Join(memberEntryPath(fx.home, member), "unused", "pin-only.txt"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			req.Capture = func(*MaterializeCapture) ([]byte, error) { called = true; return []byte("captured"), nil }
			result, err := CaptureMaterialize(req)
			if changed {
				if err == nil || !strings.Contains(err.Error(), "startup_verification_failed") || result != nil || called {
					t.Fatalf("capture allowed mismatch: callback=%v output=%v err=%v", called, result != nil, err)
				}
			} else if err != nil || result == nil || !called {
				t.Fatalf("valid capture: %v", err)
			}
		})
	}
}

// The legacy context-store contract admits regular Unix filenames beyond the
// skill manifest's portable-path grammar. New skill checks must preserve those
// admitted context bytes while the existing context link boundary still refuses.
func TestStartupPreservesContextPathContract(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this host cannot create ordinary colon filenames: this legacy filename witness requires Unix")
	}
	home, source := gitFixtureHome(t), t.TempDir()
	writeManifestPackage(t, source, `{"schema_version":1,"name":"context-path","version":"1.0.0","context":{"modules":[{"path":"guide.md"}]}}`, map[string]string{"guide.md": "Context.\n"})
	if err := os.WriteFile(filepath.Join(source, "notes:local.txt"), []byte("Admitted context reference."), 0600); err != nil {
		t.Fatal(err)
	}
	entry, pin, err := contextstore.EnsureState(home, "context", "context-path", source)
	if err != nil {
		t.Fatal(err)
	}
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Root: "context-path", Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Members: []contextlock.Member{{Kind: contextlock.KindContext, Name: "context-path", Version: "1.0.0", RequiredBy: []string{}, StateHash: strings.TrimPrefix(pin, "sha256:")}}}
	if _, err := lock.Canonical(); err != nil {
		t.Fatal(err)
	}
	if failure := validateNamedStoreBoundaries(home, lock, nil); failure != nil {
		t.Fatal(failure)
	}
	checks, err := startupChecks(home, lock)
	if err != nil {
		t.Fatalf("admitted context path narrowed by skill checks: %v", err)
	}
	for _, level := range []int{1, 2, 3, 4} {
		verdict, err := startupVerdict(Policy{StartupVerificationLevel: level}, lock, checks)
		if err != nil || !verdict.Allow {
			t.Fatalf("admitted context level%d: %v", level, err)
		}
	}
	requireLinkCapability(t)
	if err := os.Symlink("notes:local.txt", filepath.Join(entry, "forbidden")); err != nil {
		t.Fatal(err)
	}
	if failure := validateNamedStoreBoundaries(home, lock, nil); failure == nil || failure.Check != "link_safety" {
		t.Fatalf("context link gate changed: %v", failure)
	}
}
