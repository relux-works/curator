package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

// Current-lock copied-home drift is independent of source pin drift. Read-only
// reports home staleness first; permissive repair cannot launch without verified
// inputs. Strict repair may rebuild Git inputs, then restore the home. Source-only
// and home-only neighbors retain their accepted decisions.
func TestStartupCombinedHomeAndSourceDrift(t *testing.T) {
	for level := 1; level <= 4; level++ {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			fx := storeBoundaryFixture(t, "current", true)
			req := fx.request(envregistry.ClaudeCode)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			_, entry := storeBoundaryRootMemberFor(fx)
			source := filepath.Join(entry, "context", "s.md")
			home := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md")
			sourceWant, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			homeWant, err := os.ReadFile(home)
			if err != nil {
				t.Fatal(err)
			}
			req.Policy.StartupVerificationLevel = level
			req.Repair = false
			if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("intact neighbor: %v", err)
			}
			if err := os.WriteFile(source, []byte("source drift\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(home, []byte("home drift\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"observe", "dry-run", "repair"} {
				t.Run(mode, func(t *testing.T) {
					req.Repair = mode != "observe"
					req.DryRun = mode == "dry-run"
					before := hashTreeForTest(t, fx.home)
					nodes := storeVerificationNodes(t, fx.home)
					result, err := Resolve(req)
					if mode == "repair" && level >= 3 {
						if err != nil || result == nil || len(result.Document) == 0 {
							t.Fatalf("verified combined repair refused: %v", err)
						}
						got, e := os.ReadFile(home)
						if e != nil || !bytes.Equal(got, homeWant) {
							t.Fatalf("combined repair did not restore home: %v", e)
						}
						got, e = os.ReadFile(source)
						if e != nil || !bytes.Equal(got, sourceWant) {
							t.Fatalf("combined repair did not restore verified source: %v", e)
						}
						return
					}
					if mode == "observe" && level == 1 {
						warning := ""
						if result != nil {
							for _, w := range result.Warnings {
								if strings.HasPrefix(w, DiagHomeStale+":") {
									warning = w
								}
							}
						}
						if err != nil || result == nil || len(result.Document) == 0 || !strings.Contains(warning, "--repair") {
							t.Fatalf("combined read-only warning lost: %v", err)
						}
					} else if err == nil || !strings.HasPrefix(err.Error(), DiagHomeStale+":") || !strings.Contains(err.Error(), "--repair") || result != nil && len(result.Document) != 0 {
						t.Fatalf("combined %s lost leading stale refusal or launched: %v", mode, err)
					}
					if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
						t.Fatal("combined refusal rewrote home or source")
					}
				})
			}
			// Restored source is a positive repair neighbor even when the home alone
			// remains drifted; source-only permissive behavior has its own full test.
			if err := os.WriteFile(source, sourceWant, 0600); err != nil {
				t.Fatal(err)
			}
			req.Repair = true
			req.DryRun = false
			result, err := Resolve(req)
			got, e := os.ReadFile(home)
			if err != nil || result == nil || len(result.Document) == 0 || e != nil || !bytes.Equal(got, homeWant) {
				t.Fatalf("verified home-only repair neighbor: %v %v", err, e)
			}
		})
	}
}

// Unavailable pinned Git input cannot be treated as verified for home repair.
// All levels keep the source/home/lock unchanged and emit no launch fragment.
func TestStartupCombinedRepairRequiresVerifiedInputs(t *testing.T) {
	for level := 1; level <= 4; level++ {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			fx := storeBoundaryFixture(t, "current", true)
			req := fx.request(envregistry.ClaudeCode)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			member, entry := storeBoundaryRootMemberFor(fx)
			if err := os.WriteFile(filepath.Join(entry, "context", "s.md"), []byte("source drift\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md"), []byte("home drift\n"), 0600); err != nil {
				t.Fatal(err)
			}
			object := filepath.Join(newGitManager(fx.home).repoDir(member.Source), ".git", "objects", member.Commit[:2], member.Commit[2:])
			if err := os.Chmod(object, 0600); err != nil {
				t.Fatalf("required unavailable-object plant: %v", err)
			}
			if err := os.Remove(object); err != nil {
				t.Fatal(err)
			}
			before := hashTreeForTest(t, fx.home)
			nodes := storeVerificationNodes(t, fx.home)
			req.Policy.StartupVerificationLevel = level
			result, err := Resolve(req)
			if err == nil || !strings.HasPrefix(err.Error(), DiagHomeStale+":") || !strings.Contains(err.Error(), "--repair") || result != nil && len(result.Document) != 0 {
				t.Fatalf("unverified combined repair lost stale refusal or launched: %v", err)
			}
			if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
				t.Fatal("unverified combined repair wrote state")
			}
		})
	}
}

// Each changed decision has an assertion-specific narrowing control, with the
// child's intact neighbor preserved. A compile/setup failure never qualifies.
func TestStartupCombinedDriftControls(t *testing.T) {
	for _, p := range []struct{ name, file, before, after, test, witness string }{
		{"skill-content", "home_owned_drift.go", "if err := verifySkillProjection(req.Home, skillOf{root: target, hash: identity, version: version}); err != nil {", "if err := verifySkillProjection(req.Home, skillOf{root: target, hash: identity, version: version}); false && err != nil {", "TestStartupSourceOnlySkillDriftKeepsExistingHome", "changed original projection launched"},
		{"skill-receipt", "home_owned_drift.go", "return err != nil || actual != surface.ContentSHA256", "return false && (err != nil || actual != surface.ContentSHA256)", "TestStartupSourceOnlySkillDriftKeepsExistingHome", "retargeted skill receipt launched"},
		{"receipt", "home_owned_drift.go", "if complete {", "if false && complete {", "TestStartupCombinedHomeAndSourceDrift/level1/repair", "combined repair lost leading stale refusal or launched"},
		{"read-only", "managed.go", "|| ownedHomeDrift)", ")", "TestStartupCombinedHomeAndSourceDrift/level2/observe", "combined observe lost leading stale refusal or launched"},
		{"permissive-repair", "managed.go", "&& !ownedHomeDrift {", "{", "TestStartupCombinedHomeAndSourceDrift/level1/repair", "combined repair lost leading stale refusal or launched"},
		{"level2-repair", "managed.go", "|| ownedHomeDrift)", ")", "TestStartupCombinedHomeAndSourceDrift/level2/repair", "combined repair lost leading stale refusal or launched"},
		{"level2-dry-run", "managed.go", "|| ownedHomeDrift)", ")", "TestStartupCombinedHomeAndSourceDrift/level2/dry-run", "combined dry-run lost leading stale refusal or launched"},
		{"strict-dry-run", "home_owned_drift.go", "if ownedHomeDrift {", "if false && ownedHomeDrift {", "TestStartupCombinedHomeAndSourceDrift/level3/dry-run", "combined dry-run lost leading stale refusal or launched"},
		{"strict-rebuild", "managed.go", "if failure := validateNamedStorePinsForRepair(pinRequest, lock); failure != nil {", "if failure := validateNamedStorePinsForRepair(pinRequest, lock); false && failure != nil {", "TestStartupCombinedHomeAndSourceDrift/level3/repair", "verified combined repair refused"},
		{"unavailable", "home_owned_drift.go", "if ownedHomeDrift {", "if false && ownedHomeDrift {", "TestStartupCombinedRepairRequiresVerifiedInputs/level3", "unverified combined repair lost stale refusal or launched"},
	} {
		t.Run(p.name, func(t *testing.T) { runObservationControl(t, p.file, p.before, p.after, p.test, p.witness) })
	}
}

// A changed skill source must not turn its intact original projection link
// into home-owned drift. A genuinely retargeted link is the negative neighbor.
func TestStartupSourceOnlySkillDriftKeepsExistingHome(t *testing.T) {
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(contextlock.KindSkill, "myskill")
	if !ok {
		t.Fatal("skill missing")
	}
	path := filepath.Join(memberEntryPath(fx.home, member), "SKILL.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(body, []byte("\nsource-only change\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	before := hashTreeForTest(t, fx.home)
	nodes := storeVerificationNodes(t, fx.home)
	req.Policy.StartupVerificationLevel = 1
	result, err := Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("source-only skill neighbor refused: %v", err)
	}
	if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
		t.Fatal("source-only skill neighbor wrote state")
	}
	adapter, err := envregistry.ByID(envregistry.CodexCLI)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), adapter.SkillsDir, "myskill")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	// Retarget to another fully valid projection, so only marker receipt
	// binding can reject it; a missing-target failure cannot satisfy this case.
	alternate := []byte("# alternate projection\n")
	digest, err := hashing.ContentSHA256Files(map[string][]byte{"SKILL.md": alternate}, lock.ContentHashVersion())
	if err != nil {
		t.Fatal(err)
	}
	wrong := filepath.Join(filepath.Dir(target), hashing.Normalize(digest))
	if err := privatedir.MakeAll(wrong); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wrong, "SKILL.md"), alternate, 0600); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(wrong); err != nil {
		t.Fatal(err)
	}
	if err := verifySkillProjection(fx.home, skillOf{root: wrong, hash: digest, version: lock.ContentHashVersion()}); err != nil {
		t.Fatalf("alternate projection setup: %v", err)
	}
	if err := os.Symlink(wrong, link); err != nil {
		t.Fatal(err)
	}
	before = hashTreeForTest(t, fx.home)
	nodes = storeVerificationNodes(t, fx.home)
	result, err = Resolve(req)
	if err == nil || !strings.HasPrefix(err.Error(), DiagHomeStale+":") || result != nil && len(result.Document) != 0 {
		t.Fatalf("retargeted skill receipt launched: %v", err)
	}
	if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
		t.Fatal("retargeted skill refusal wrote state")
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	result, err = Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("restored skill link neighbor refused: %v", err)
	}
	// The original target's name alone is not a receipt of its current bytes.
	projected := filepath.Join(target, "SKILL.md")
	original, err := os.ReadFile(projected)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projected, []byte("projection drift\n"), 0600); err != nil {
		t.Fatal(err)
	}
	before = hashTreeForTest(t, fx.home)
	nodes = storeVerificationNodes(t, fx.home)
	result, err = Resolve(req)
	if err == nil || !strings.HasPrefix(err.Error(), DiagHomeStale+":") || result != nil && len(result.Document) != 0 {
		t.Fatalf("changed original projection launched: %v", err)
	}
	if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
		t.Fatal("projection refusal wrote state")
	}
	if err := os.WriteFile(projected, original, 0600); err != nil {
		t.Fatal(err)
	}
	result, err = Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("restored projection neighbor refused: %v", err)
	}
}
