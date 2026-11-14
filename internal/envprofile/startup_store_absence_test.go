package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// Each level/mode has an intact no-write control. Whole Git-entry absence
// permits only strict repair (or a strict dry-run plan); a missing named
// payload inside a present entry remains a direct structural refusal.
// Snapshots cover bytes, entry types and link targets, not all OS metadata.
func TestStartupStoreAbsenceClassification(t *testing.T) {
	for _, member := range []string{"context", "mcp", "skill"} {
		for _, damage := range []string{"entry", "payload"} {
			for _, mode := range []string{"observe", "repair", "dry-run"} {
				for level := 1; level <= 4; level++ {
					t.Run(fmt.Sprintf("%s/%s/%s/level%d", member, damage, mode, level), func(t *testing.T) {
						fx := storeBoundaryFixture(t, "current", true)
						req := fx.request("codex_cli")
						req.Policy.StartupVerificationLevel = level
						req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
						before := hashTreeForTest(t, fx.home)
						positive, err := Resolve(req)
						if err != nil || positive == nil || len(positive.Document) == 0 || hashTreeForTest(t, fx.home) != before {
							t.Fatalf("intact no-write neighbor: %v", err)
						}
						lockBefore, readErr := os.ReadFile(lockPath(fx.home, fx.profile))
						if readErr != nil {
							t.Fatal(readErr)
						}
						entry := storeBoundaryEntry(t, fx, member)
						target := entry
						if damage == "entry" {
							if err := os.RemoveAll(entry); err != nil {
								t.Fatal(err)
							}
						} else {
							rel := map[string]string{"context": "context/a.md", "mcp": "agent-mcp.json", "skill": "SKILL.md"}[member]
							target = filepath.Join(entry, filepath.FromSlash(rel))
							if err := os.Remove(target); err != nil {
								t.Fatal(err)
							}
						}
						before = hashTreeForTest(t, fx.home)
						result, err := Resolve(req)
						if damage == "entry" && mode == "repair" && level >= 3 {
							if err != nil || result == nil || len(result.Document) == 0 {
								t.Fatalf("same-pin Git recovery: %v", err)
							}
							lock, _, readErr := contextlock.Read(lockPath(fx.home, fx.profile))
							if readErr != nil {
								t.Fatal(readErr)
							}
							if failure := validateNamedStorePins(fx.home, lock); failure != nil {
								t.Fatalf("recreated entry not at original pin: %v", failure)
							}
							if hash, readErr := os.ReadFile(lockPath(fx.home, fx.profile)); readErr != nil || !bytes.Equal(hash, lockBefore) {
								t.Fatalf("lock unavailable: %v", readErr)
							}
							return
						}
						want := "environment_store_untrusted:"
						if damage == "entry" && mode == "dry-run" && level >= 3 {
							want = "would-rebuild-untrusted-store:"
						}
						if err == nil || !strings.HasPrefix(err.Error(), want) || !strings.Contains(err.Error(), "regular_types") || !strings.Contains(err.Error(), target) || result != nil && len(result.Document) > 0 {
							t.Errorf("wrong named absence refusal/plan: fragment=%v err=%v", result != nil && len(result.Document) > 0, err)
						}
						if hashTreeForTest(t, fx.home) != before {
							t.Error("absence refusal/plan wrote manager state")
						}
					})
				}
			}
		}
	}
}
