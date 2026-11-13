package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// A safe ordinary-byte repair succeeds at3/4. A structural object planted at
// the existing repair-lock seam must instead preserve the exact plant and
// complete post-state without a fragment. This bounds the observed seam; it
// does not claim exclusion of arbitrary external filesystem writers.
func TestStartupRepairRejectsLateStorePlant(t *testing.T) {
	for _, plant := range []string{"none", "symlink", "hardlink"} {
		for _, level := range []int{3, 4} {
			t.Run(fmt.Sprintf("%s/level%d", plant, level), func(t *testing.T) {
				fx := storeBoundaryFixture(t, "current", true)
				entry := storeBoundaryEntry(t, fx, contextlock.KindMCP)
				target := filepath.Join(entry, "unused", "pin-only.txt")
				if err := os.WriteFile(target, []byte("changed regular bytes\n"), 0600); err != nil {
					t.Fatal(err)
				}
				req := fx.request("codex_cli")
				req.Repair = true
				req.Policy.StartupVerificationLevel = level
				outside := t.TempDir()
				var before, outsideBefore, linkTarget string
				var observed os.FileInfo
				called := 0
				req.beforeRepairLock = func() {
					called++
					if called != 1 {
						t.Fatal("repair seam repeated")
					}
					switch plant {
					case "none":
					case "symlink":
						if err := os.Rename(target, target+".saved"); err != nil {
							t.Fatal(err)
						}
						linkTarget = filepath.Base(target) + ".saved"
						if err := os.Symlink(linkTarget, target); err != nil {
							t.Fatalf("late symlink fixture unavailable: %v", err)
						}
					case "hardlink":
						if err := os.Link(target, filepath.Join(outside, "plant")); err != nil {
							t.Fatalf("late hardlink fixture unavailable: %v", err)
						}
					default:
						t.Fatal("unknown plant")
					}
					var err error
					observed, err = os.Lstat(target)
					if err != nil {
						t.Fatal(err)
					}
					before = hashTreeForTest(t, fx.home)
					outsideBefore = hashTreeForTest(t, outside)
				}
				result, err := Resolve(req)
				if called != 1 || observed == nil {
					t.Fatal("late plant seam was not applied; not a successful control")
				}
				if plant == "none" {
					if err != nil || result == nil || len(result.Document) == 0 {
						t.Fatalf("safe byte-repair neighbor refused: %v", err)
					}
					got, readErr := os.ReadFile(target)
					if readErr != nil || string(got) != "not surfaced\n" {
						t.Fatalf("safe bytes not restored: %q %v", got, readErr)
					}
					return
				}
				check := "link_safety"
				if plant == "hardlink" {
					check = "single_link"
				}
				if err == nil || !strings.HasPrefix(err.Error(), "environment_store_untrusted:") || !strings.Contains(err.Error(), "failed "+check+" check:") || !strings.Contains(err.Error(), target) || result != nil && len(result.Document) > 0 {
					t.Errorf("late structural plant was not directly refused: fragment=%v err=%v", result != nil && len(result.Document) > 0, err)
				}
				after, statErr := os.Lstat(target)
				if statErr != nil || !os.SameFile(observed, after) {
					t.Errorf("late plant was erased/replaced: %v", statErr)
				}
				if plant == "symlink" {
					if got, readErr := os.Readlink(target); readErr != nil || got != linkTarget {
						t.Errorf("late link changed: %q %v", got, readErr)
					}
				}
				if hashTreeForTest(t, fx.home) != before || hashTreeForTest(t, outside) != outsideBefore {
					t.Error("late structural refusal changed store, lock, home, marker or outside plant")
				}
			})
		}
	}
}
