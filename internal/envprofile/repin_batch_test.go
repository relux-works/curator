package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/envregistry"
)

// Claim: the batched CLI update path compensates a late takeover refusal on a repin exactly like the single-name path, and the same batch succeeds with explicit takeover.
func TestRepinBatchTakeoverRefusalPreservesState(t *testing.T) {
	f := newRepinFixture(t)
	native := claudeHome(t)
	f.homes["claude_code"] = native
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("unmanaged context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := f.snapshot(t)
	opts := UpdateOptions{RootRequirement: &Requirement{Revision: f.b}}
	var visitErr error
	var visitMoved, visited bool
	batchErr := UpdateProfilesWithOptions(f.home, []string{"repin"}, opts, func(_ string, _ Info, moved bool, err error) bool {
		visited, visitMoved, visitErr = true, moved, err
		return true
	})
	if batchErr != nil || !visited || visitErr == nil || visitMoved || !strings.Contains(visitErr.Error(), DiagUsePartial) {
		t.Fatalf("batch takeover refusal: batch=%v visited=%t moved=%t err=%v", batchErr, visited, visitMoved, visitErr)
	}
	if !reflect.DeepEqual(before, f.snapshot(t)) {
		after := f.snapshot(t)
		for group, entries := range before {
			for path, value := range entries {
				if after[group][path] != value {
					t.Errorf("late batch refusal changed %s/%s", group, path)
				}
			}
		}
		for group, entries := range after {
			for path, value := range entries {
				if before[group][path] != value {
					t.Errorf("late batch refusal added/changed %s/%s", group, path)
				}
			}
		}
		t.Fatal("late batch refusal changed mutable profile state")
	}
	opts.Policy.Takeover = true
	visited, visitMoved, visitErr = false, false, nil
	batchErr = UpdateProfilesWithOptions(f.home, []string{"repin"}, opts, func(_ string, _ Info, moved bool, err error) bool {
		visited, visitMoved, visitErr = true, moved, err
		return true
	})
	if batchErr != nil || !visited || visitErr != nil || !visitMoved {
		t.Fatalf("batch takeover control: batch=%v visited=%t moved=%t err=%v", batchErr, visited, visitMoved, visitErr)
	}
	rendered, err := os.ReadFile(filepath.Join(native, "CLAUDE.md"))
	if err != nil || !strings.Contains(string(rendered), "revision B") {
		t.Fatalf("batch takeover context: %s %v", rendered, err)
	}
}

// Claim: batch update and profile sync refuse a prepared v5 plan at entry, before any operation creates the default profile or reaches a per-profile visit.
func TestPreparedPlanProbeOnBatchUpdateAndSync(t *testing.T) {
	f := newPreparationFixture(t)
	if _, err := PrepareSkillPlan(f.home, "profile", f.hash, f.policy, func([]*closure.Node) error { return nil }); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, f.home, "profiles/current", "profile\n")
	lockBefore, err := os.ReadFile(lockPath(f.home, "profile"))
	if err != nil {
		t.Fatal(err)
	}
	defaultAbsent := func(stage string) {
		t.Helper()
		if _, err := os.Lstat(filepath.Join(f.home, "profiles", DefaultProfile)); !os.IsNotExist(err) {
			t.Fatalf("%s: pre-recovery probe did not run first; default profile was created (err=%v)", stage, err)
		}
		now, err := os.ReadFile(lockPath(f.home, "profile"))
		if err != nil || !bytes.Equal(lockBefore, now) {
			t.Fatalf("%s: v5 lock changed: %v", stage, err)
		}
	}
	// Control: main's managed-sync entry already refuses at its probe.
	if _, err := SyncManagedHomesWithPolicy(f.home, f.policy, envregistry.DefaultMachineConfig()); err == nil || !strings.Contains(err.Error(), "profile_skill_companion_required") {
		t.Fatalf("managed sync control: %v", err)
	}
	defaultAbsent("managed sync control")
	// The CLI sync entry must refuse the same way, before ensureDefault.
	if _, _, managedErr, _ := SyncHomesWithPolicy(f.home, f.policy, envregistry.DefaultMachineConfig()); managedErr == nil || !strings.Contains(managedErr.Error(), "profile_skill_companion_required") {
		t.Fatalf("sync probe: %v", managedErr)
	}
	defaultAbsent("sync")
	// The batched CLI update entry must refuse before any visit runs.
	visited := false
	batchErr := UpdateProfilesWithOptions(f.home, []string{"profile"}, UpdateOptions{Policy: f.policy}, func(_ string, _ Info, _ bool, _ error) bool {
		visited = true
		return true
	})
	if batchErr == nil || !strings.Contains(batchErr.Error(), "profile_skill_companion_required") || visited {
		t.Fatalf("batch update probe: batch=%v visited=%t", batchErr, visited)
	}
	defaultAbsent("batch update")
}
