package crossconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/transaction"
)

// Boundary semantic rows (protocol/skillfile-sources.md §1): every case
// drives the production admission entry snapshot.ValidateLocalPackage
// (or install.Project for the write-time retarget), never the corpus
// labels.

func init() {
	registerDraftSemantic("broad-root", driveBroadRoot)
	registerDraftSemantic("managed-source", driveManagedSource)
	registerDraftSemantic("symlink-managed", driveSymlinkManaged)
	registerDraftSemantic("case-alias", driveCaseAlias)
	registerDraftSemantic("write-boundary-retarget", driveWriteBoundaryRetarget)
	registerDraftSemantic("root-no-inputs", driveRootNoInputs)
	registerDraftSemantic("selector-escape", driveSelectorEscape)
}

func draftBoundaryOutputs(project string) []string {
	return []string{filepath.Join(project, ".agents")}
}

func driveBroadRoot(t *testing.T, c draftSemanticCase) {
	project := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "agents", "skills", "review"), "review")
	physical, admitted, err := snapshot.ValidateLocalPackage(project, "agents/skills/review", project, "s", draftBoundaryOutputs(project), nil, map[string]bool{"s": true})
	if err != nil {
		t.Fatalf("broad root with safe subdirectory: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(physical), "agents/skills/review") {
		t.Fatalf("physical = %s", physical)
	}
	if len(admitted) != 1 || admitted[0] != physical {
		t.Fatalf("admitted = %v, want [%s]", admitted, physical)
	}
	if c.Expected != "allow" {
		t.Fatalf("expected = %q, want allow", c.Expected)
	}
}

func driveManagedSource(t *testing.T, _ draftSemanticCase) {
	project := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, ".agents", "skills", "review"), "review")
	_, _, err := snapshot.ValidateLocalPackage(project, ".agents/skills/review", project, "s", draftBoundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("err = %v, want source_output_overlap", err)
	}
}

func driveSymlinkManaged(t *testing.T, _ draftSemanticCase) {
	project := t.TempDir()
	target := filepath.Join(project, ".agents", "skills", "review")
	writeDraftSkill(t, target, "review")
	link := filepath.Join(project, "authored", "review")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, _, err := snapshot.ValidateLocalPackage(project, "authored/review", project, "s", draftBoundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("err = %v, want source_output_overlap", err)
	}
}

func driveCaseAlias(t *testing.T, _ draftSemanticCase) {
	project := t.TempDir()
	lower := filepath.Join(project, ".agents", "skills", "review")
	writeDraftSkill(t, lower, "review")
	upper := filepath.Join(project, ".AGENTS", "skills", "review")
	if err := os.MkdirAll(upper, 0o755); err != nil {
		t.Fatal(err)
	}
	// The alias is observable only where the filesystem folds it.
	sentinel := filepath.Join(lower, ".alias-probe")
	if err := os.WriteFile(sentinel, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(upper, ".alias-probe")); err != nil {
		t.Skip("case-alias needs a case-insensitive filesystem: host case sensitivity differs")
	}
	_, _, err := snapshot.ValidateLocalPackage(project, ".AGENTS/skills/review", project, "s", draftBoundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("err = %v, want source_output_overlap", err)
	}
}

func driveRootNoInputs(t *testing.T, _ draftSemanticCase) {
	project := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review")
	_, _, err := snapshot.ValidateLocalPackage(project, ".", project, "s", draftBoundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("err = %v, want source_output_overlap", err)
	}
}

func driveSelectorEscape(t *testing.T, _ draftSemanticCase) {
	base := t.TempDir()
	sourceRoot := filepath.Join(base, "skills")
	outside := filepath.Join(base, "outside", "review")
	writeDraftSkill(t, outside, "review")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(sourceRoot, "review")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, _, err := snapshot.ValidateLocalPackage(sourceRoot, "review", base, "s", draftBoundaryOutputs(base), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
		t.Fatalf("err = %v, want source_selection_invalid", err)
	}
}

// driveWriteBoundaryRetarget proves the per-write guard refuses a
// publication output retargeted after planning: the planned output
// (project/.agents) is swapped for a copy between the durable
// preparation and the commit, and install.Project must fail with
// source_output_overlap and roll every target back.
func driveWriteBoundaryRetarget(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	if result := draftInstall(t, project, home, install.Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	refreshDraftLocked(t, project, home, payload)
	before := treeDigest(t, project) + treeDigest(t, home)
	skillsDir := filepath.Join(project, ".agents", "skills")
	stagedAway := skillsDir + ".orig"
	fired := false
	hooks := transaction.Hooks{Observe: func(event transaction.Event) {
		if event.Point != transaction.PointPrepared || fired {
			return
		}
		fired = true
		if err := os.Rename(skillsDir, stagedAway); err != nil {
			panic(err)
		}
		copyTree(t, stagedAway, skillsDir)
	}}
	result := draftInstall(t, project, home, install.Options{Commit: install.CommitDeps{Hooks: hooks, MaxRestarts: 1}})
	if !fired {
		t.Fatal("retarget hook never fired")
	}
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_output_overlap") {
		t.Fatalf("install = %+v, want the source_output_overlap refusal", result)
	}
	// Undo the test's own retarget, sweep the journal sidecars the swap
	// moved aside and back, then prove the failed run changed nothing.
	if err := os.RemoveAll(skillsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(stagedAway, skillsDir); err != nil {
		t.Fatal(err)
	}
	sweepTxnSidecars(t, skillsDir)
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("rollback did not restore the prior state")
	}
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		from := filepath.Join(src, entry.Name())
		to := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyTree(t, from, to)
			continue
		}
		payload, err := os.ReadFile(from)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func sweepTxnSidecars(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".curator-txn-") {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
}
