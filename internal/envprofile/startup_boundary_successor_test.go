package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
)

func provisionBoundaryFixture(t *testing.T) *managedFixture {
	t.Helper()
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	return fx
}

// The proposed all-member boundary must reject hardlinks before pin comparison
// or restoration. Every negative has an actual intact Resolve neighbor at the
// same level/mode; both preserve the complete fixture state. Resolve itself does
// not launch a child; actual child/no-child docking is a separate check.
func TestStartupNonSkillHardlinkBoundary(t *testing.T) {
	for _, kind := range []string{contextlock.KindContext, contextlock.KindMCP} {
		for _, mode := range []string{"observe", "repair", "dry-run"} {
			for level := 1; level <= 4; level++ {
				t.Run(fmt.Sprintf("%s/%s/level%d", kind, mode, level), func(t *testing.T) {
					fx := provisionBoundaryFixture(t)
					lock, _, err := readLock(fx.home, fx.profile)
					if err != nil {
						t.Fatal(err)
					}
					name, relative := "sample", "agent-context.json"
					if kind == contextlock.KindMCP {
						name, relative = "figma-devmode", "unused/pin-only.txt"
					}
					member, ok := lock.Find(kind, name)
					if !ok {
						t.Fatal("missing control member")
					}
					entry := memberEntryPath(fx.home, member)
					target := filepath.Join(entry, filepath.FromSlash(relative))
					req := fx.request("codex_cli")
					req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
					req.Policy.StartupVerificationLevel = level
					intact := hashTreeForTest(t, fx.home)
					result, err := Resolve(req)
					if err != nil || result == nil || len(result.Document) == 0 {
						t.Fatalf("intact positive refused: %v", err)
					}
					if got := hashTreeForTest(t, fx.home); got != intact {
						t.Fatal("intact positive wrote fixture state")
					}
					plant := filepath.Join(entry, "forbidden-hardlink")
					if err := os.Link(target, plant); err != nil {
						t.Fatalf("hardlink fixture unavailable: %v", err)
					}
					plantBefore, err := os.Lstat(plant)
					if err != nil {
						t.Fatal(err)
					}
					before := hashTreeForTest(t, fx.home)
					result, err = Resolve(req)
					var failure *storeEntryFailure
					named := err != nil && (strings.Contains(err.Error(), target) || strings.Contains(err.Error(), plant))
					if !errors.As(err, &failure) || failure.Check != "single_link" || !named || result != nil && len(result.Document) > 0 {
						t.Errorf("expected named single_link refusal, output=%v err=%v", result != nil && len(result.Document) > 0, err)
					}
					plantAfter, statErr := os.Lstat(plant)
					if statErr != nil || !os.SameFile(plantBefore, plantAfter) {
						t.Errorf("hardlink was erased or replaced: %v", statErr)
					}
					if got := hashTreeForTest(t, fx.home); got != before {
						t.Error("hardlink refusal changed store, lock, home or marker")
					}
				})
			}
		}
	}
}

// A marker link remains a direct structural refusal, including dry-run repair;
// a rebuild-plan diagnostic is forbidden even when the invocation starts no
// child. The intact neighbor and planted marker/target remain byte-identical.
func TestStartupMarkerBoundaryNoRebuildPlan(t *testing.T) {
	requireLinkCapability(t)
	for _, mode := range []string{"observe", "repair", "dry-run"} {
		for level := 1; level <= 4; level++ {
			t.Run(fmt.Sprintf("%s/level%d", mode, level), func(t *testing.T) {
				fx := provisionBoundaryFixture(t)
				req := fx.request("codex_cli")
				req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
				req.Policy.StartupVerificationLevel = level
				intact := hashTreeForTest(t, fx.home)
				result, err := Resolve(req)
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("intact marker positive refused: %v", err)
				}
				if got := hashTreeForTest(t, fx.home); got != intact {
					t.Fatal("intact marker positive wrote fixture state")
				}
				marker := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
				target := filepath.Join(filepath.Dir(marker), "marker-link-target")
				if err := os.Rename(marker, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Base(target), marker); err != nil {
					t.Fatal(err)
				}
				plantBefore, err := os.Lstat(marker)
				if err != nil {
					t.Fatal(err)
				}
				before := hashTreeForTest(t, fx.home)
				result, err = Resolve(req)
				if err == nil || !strings.HasPrefix(err.Error(), "environment_store_untrusted: environment marker failed link_safety check:") || !strings.Contains(err.Error(), marker) || result != nil && len(result.Document) > 0 {
					t.Errorf("expected direct marker boundary refusal, output=%v err=%v", result != nil && len(result.Document) > 0, err)
				}
				plantAfter, statErr := os.Lstat(marker)
				if statErr != nil || !os.SameFile(plantBefore, plantAfter) {
					t.Errorf("marker link was replaced: %v", statErr)
				}
				if got, readErr := os.Readlink(marker); readErr != nil || got != filepath.Base(target) {
					t.Errorf("marker link was retargeted: %v", readErr)
				}
				if got := hashTreeForTest(t, fx.home); got != before {
					t.Error("marker refusal changed store, lock, home or marker target")
				}
			})
		}
	}
}
