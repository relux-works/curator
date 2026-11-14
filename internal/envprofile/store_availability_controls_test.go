package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/pathboundary"
)

// A missing leaf is not a waiver for a missing/unsafe ancestor or a later
// member's structural violation. Actual Resolve must refuse before its repair
// seam with unchanged manager bytes/types/link targets. Owner failure uses the
// existing metadata seam; it is not a real foreign-owner filesystem test.
func TestStartupAbsentEntryDoesNotHideStructure(t *testing.T) {
	for _, plant := range []string{"missing-parent", "dangling-parent", "parent-owner", "later-hardlink"} {
		for _, level := range []int{3, 4} {
			t.Run(fmt.Sprintf("%s/level%d", plant, level), func(t *testing.T) {
				fx := storeBoundaryFixture(t, "current", true)
				req := fx.request("codex_cli")
				req.Repair = true
				req.Policy.StartupVerificationLevel = level
				entry := storeBoundaryEntry(t, fx, contextlock.KindContext)
				if err := os.RemoveAll(entry); err != nil {
					t.Fatal(err)
				}
				parent := filepath.Dir(entry)
				check := "regular_types"
				switch plant {
				case "missing-parent":
					if err := os.Rename(parent, parent+".saved"); err != nil {
						t.Fatal(err)
					}
				case "dangling-parent":
					if err := os.Rename(parent, parent+".saved"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink("missing-parent-target", parent); err != nil {
						t.Fatalf("symlink fixture unavailable: %v", err)
					}
					// A dangling link cannot prove link safety; both the named
					// route and plant must remain visible, never availability.
					check = "link_safety"
				case "parent-owner":
					check = "ownership"
					req.boundaryOwnerLookup = func(path string, info os.FileInfo) (pathboundary.OwnerIdentity, error) {
						if filepath.Clean(path) == filepath.Clean(parent) {
							return pathboundary.OwnerIdentity("different-owner"), nil
						}
						return pathboundary.DefaultOwnerLookup()(path, info)
					}
				case "later-hardlink":
					check = "single_link"
					target := filepath.Join(storeBoundaryEntry(t, fx, contextlock.KindMCP), "unused", "pin-only.txt")
					if err := os.Link(target, filepath.Join(t.TempDir(), "outside")); err != nil {
						t.Fatalf("hardlink fixture unavailable: %v", err)
					}
				}
				before := hashTreeForTest(t, fx.home)
				reached := false
				req.beforeRepairLock = func() { reached = true }
				result, err := Resolve(req)
				if err == nil || !strings.HasPrefix(err.Error(), "environment_store_untrusted:") || !strings.Contains(err.Error(), "failed "+check+" check:") || result != nil && len(result.Document) > 0 {
					t.Errorf("wrong pre-repair structural refusal: %v", err)
				}
				if reached {
					t.Error("structural violation reached repair-lock seam")
				}
				if hashTreeForTest(t, fx.home) != before {
					t.Error("structural refusal mutated manager state")
				}
			})
		}
	}
}

// A native Git skill retains existing SKILL.md package admission. Its actual
// provisioning and subsequent no-write startup use its recorded full roster.
// Removing its recorded SKILL.md must refuse repair at each level without a
// fragment or manager writes; restoring the same bytes restores the neighbor.
func TestStartupNativeSkillRecordedRoster(t *testing.T) {
	fx := writeManagedFixture(t, "example")
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	for i, member := range lock.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		source := "github.com/example/skill-native-only"
		commit := writeGitStoreFixture(t, fx.home, member.Kind, member.Name, source, map[string]string{
			"agent-skill.json": `{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}`,
			"SKILL.md":         "---\nname: myskill\ndescription: Native fixture skill.\n---\nNative fixture context.\n",
		})
		lock.Members[i].Source = source
		lock.Members[i].Commit = commit
	}
	lock.Sort()
	if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
		t.Fatal(err)
	}
	provisionStoreBoundaryHome(t, fx)
	skillPath := filepath.Join(storeBoundaryEntry(t, fx, contextlock.KindSkill), "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	for level := 1; level <= 4; level++ {
		req := fx.request("codex_cli")
		req.Policy.StartupVerificationLevel = level
		before := hashTreeForTest(t, fx.home)
		result, err := Resolve(req)
		if err != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("native recorded roster level%d: %v", level, err)
		}
		if hashTreeForTest(t, fx.home) != before {
			t.Fatalf("native recorded roster level%d wrote state", level)
		}
		if err := os.Remove(skillPath); err != nil {
			t.Fatal(err)
		}
		before = hashTreeForTest(t, fx.home)
		req.Repair = true
		result, err = Resolve(req)
		if err == nil || !strings.HasPrefix(err.Error(), "environment_store_untrusted:") || !strings.Contains(err.Error(), "failed regular_types check:") || !strings.Contains(err.Error(), skillPath) || result != nil && len(result.Document) != 0 {
			t.Fatalf("native missing SKILL.md level%d did not refuse the named payload: %v", level, err)
		}
		if hashTreeForTest(t, fx.home) != before {
			t.Fatalf("native missing SKILL.md level%d repair wrote state", level)
		}
		if err := os.WriteFile(skillPath, skillBytes, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
