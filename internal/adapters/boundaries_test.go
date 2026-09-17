package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/staging"
)

func TestProjectOutputRootsCoverManagedTrees(t *testing.T) {
	project := t.TempDir()
	roots := ProjectOutputRoots(project)
	joined := strings.Join(roots, "\n")
	for _, want := range []string{".agents", ".claude/skills", ".codex/skills", ".cursor/rules", ".gemini/skills"} {
		if !strings.Contains(joined, filepath.FromSlash(want)) {
			t.Fatalf("roots = %v, want %s", roots, want)
		}
	}
	if strings.Contains(joined, filepath.Join("agents", "skills")) && !strings.Contains(joined, ".agents") {
		t.Fatal("authored agents tree listed as output")
	}
}

func TestValidateDestinationsRefusesAdmittedOverwrite(t *testing.T) {
	project := t.TempDir()
	admitted := filepath.Join(project, "agents", "skills", "review")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(admitted, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A destination inside admitted inputs fails.
	var inside staging.Plan
	inside.Replace(staging.ClassAdapterLedger, "entry", filepath.Join(admitted, "entry"), filepath.Join(project, "staged"))
	if err := ValidateDestinations(inside, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("inside err = %v, want source_output_overlap", err)
	}
	// An admitted input inside a destination fails.
	dest := filepath.Join(project, ".agents", "skills")
	inner := filepath.Join(dest, "inner", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(inner), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inner, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var outside staging.Plan
	outside.Replace(staging.ClassAdapterLedger, "entry", dest, filepath.Join(project, "staged"))
	if err := ValidateDestinations(outside, []string{inner}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("reverse err = %v, want source_output_overlap", err)
	}
	// A disjoint managed destination stays valid.
	var okPlan staging.Plan
	okPlan.Replace(staging.ClassAdapterLedger, "entry", filepath.Join(project, ".claude", "skills", "review"), filepath.Join(project, "staged"))
	if err := ValidateDestinations(okPlan, []string{admitted}); err != nil {
		t.Fatalf("disjoint destination: %v", err)
	}
}

func TestValidateDestinationsFollowsLinksAndParents(t *testing.T) {
	project := t.TempDir()
	admitted := filepath.Join(project, "agents", "skills", "review")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	// A destination whose parent is a link into admitted inputs fails.
	linkParent := filepath.Join(project, "linked")
	if err := os.Symlink(admitted, linkParent); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var plan staging.Plan
	plan.Replace(staging.ClassAdapterLedger, "entry", filepath.Join(linkParent, "entry"), filepath.Join(project, "staged"))
	if err := ValidateDestinations(plan, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("linked parent err = %v, want source_output_overlap", err)
	}
	// A live link destination pointing into admitted inputs fails.
	live := filepath.Join(project, ".claude", "skills", "review")
	if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(admitted, "SKILL.md"), live); err != nil {
		t.Fatal(err)
	}
	var linkPlan staging.Plan
	linkPlan.ReplaceEntry(staging.ClassAdapterLedger, "entry", live, filepath.Join(project, "staged"))
	if err := ValidateDestinations(linkPlan, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("link destination err = %v, want source_output_overlap", err)
	}
}

func TestValidateDestinationsRefusesEntryInsideAdmitted(t *testing.T) {
	// Only the entry-location arm can refuse this shape: the live entry
	// sits inside admitted inputs while resolving outside them, so the
	// fully-resolved comparison passes and the managed entry's own
	// location must still fail.
	project := t.TempDir()
	admitted := filepath.Join(project, "admitted")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(project, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(admitted, "entry")
	if err := os.Symlink(outside, live); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var plan staging.Plan
	plan.ReplaceEntry(staging.ClassAdapterLedger, "entry", live, filepath.Join(project, "staged"))
	if err := ValidateDestinations(plan, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("entry inside admitted err = %v, want source_output_overlap", err)
	}
}

func TestUnmanagedTakeoverRefusedBesideAdmittedCheck(t *testing.T) {
	// The root-input allowlist authorizes reading source inputs; it never
	// authorizes overwriting an unmanaged live destination.
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	admitted := filepath.Join(project, "agents", "skills", "skill-a")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(project, ".claude", "skills", "skill-a")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, "mine.md"), []byte("hands off"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err == nil || !strings.Contains(err.Error(), "not managed") {
		t.Fatalf("unmanaged err = %v, want unmanaged conflict", err)
	}
	// The admitted-input check alone would pass for this disjoint
	// destination; unmanaged protection is the refusing gate.
	var plan staging.Plan
	plan.Replace(staging.ClassAdapterLedger, "entry", foreign, filepath.Join(stageRoot, "entry"))
	if err := ValidateDestinations(plan, []string{admitted}); err != nil {
		t.Fatalf("disjoint admitted check should pass, keeping unmanaged as the refusing gate: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(foreign, "mine.md")); statErr != nil {
		t.Fatal("foreign content was destroyed")
	}
}

func TestStageProjectRefusesMirrorOverAdmittedInput(t *testing.T) {
	// The production planner enforces destination separation itself: a
	// group whose admitted inputs cover the planned mirror is refused at
	// StageProject, not by a helper the caller must remember.
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	adapterRoot := filepath.Join(project, ".claude", "skills")
	_, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}, Admitted: []string{adapterRoot}}}, "copy")
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("StageProject over admitted err = %v, want source_output_overlap", err)
	}
	if _, statErr := os.Lstat(filepath.Join(adapterRoot, "skill-a")); !os.IsNotExist(statErr) {
		t.Fatalf("refused planning touched the live destination: %v", statErr)
	}
}

func TestStageProjectAdmitsDisjointAdmittedInputs(t *testing.T) {
	// Pinned admitted inputs disjoint from every mirror plan unchanged:
	// the gate refuses overlap, not the presence of the pin.
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	admitted := filepath.Join(project, "agents", "skills", "skill-a")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}, Admitted: []string{admitted}}}, "copy")
	if err != nil {
		t.Fatalf("StageProject with disjoint admitted: %v", err)
	}
	if len(mirror.Plan().Targets) == 0 {
		t.Fatal("StageProject planned no targets")
	}
}

func TestCasingAliasHelperFlagsCaseVariants(t *testing.T) {
	parent := t.TempDir()
	if err := os.Mkdir(filepath.Join(parent, "Skill-A"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !casingAliasUnmanaged(filepath.Join(parent, "skill-a")) {
		t.Fatal("case-variant spelling was not flagged")
	}
	if casingAliasUnmanaged(filepath.Join(parent, "Skill-A")) {
		t.Fatal("exact spelling was flagged as an alias")
	}
	if casingAliasUnmanaged(filepath.Join(parent, "absent")) {
		t.Fatal("absent name was flagged as an alias")
	}
}

func TestValidateDestinationsIdentityReadErrorIsNotOverlap(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan staging.Plan
	plan.Replace(staging.ClassContext, "a", filepath.Join(parent, "out"), filepath.Join(root, "staged"))
	if err := ValidateDestinations(plan, []string{admitted}); err != nil {
		t.Fatalf("disjoint destinations: %v", err)
	}
	if err := os.Rename(parent, parent+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("parent", parent); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	err := ValidateDestinations(plan, []string{admitted})
	if err == nil {
		t.Fatal("expected read failure")
	}
	if strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("identity read failure mislabeled as overlap: %v", err)
	}
	if !strings.Contains(err.Error(), "boundary_identity_unreadable") {
		t.Fatalf("identity read failure err = %v, want boundary_identity_unreadable", err)
	}
}
