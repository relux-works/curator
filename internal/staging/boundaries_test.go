package staging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalizeResolvesSymlinksAndMissingSuffix(t *testing.T) {
	root := t.TempDir()
	actual := filepath.Join(root, "actual")
	if err := os.Mkdir(actual, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got, err := Canonicalize(link)
	if err != nil {
		t.Fatal(err)
	}
	want, err := Canonicalize(actual)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("symlink canonicalizes to %q, want %q", got, want)
	}
	missing := filepath.Join(link, "absent", "child")
	gotMissing, err := Canonicalize(missing)
	if err != nil {
		t.Fatal(err)
	}
	if gotMissing != filepath.Join(want, "absent", "child") {
		t.Fatalf("missing suffix canonicalizes to %q", gotMissing)
	}
	if _, err := Canonicalize(filepath.Join("relative", "path")); err == nil {
		t.Fatal("relative path canonicalized")
	}
}

func TestWithinCatchesSymlinkAliases(t *testing.T) {
	root := t.TempDir()
	managed := filepath.Join(root, ".agents")
	if err := os.MkdirAll(filepath.Join(managed, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	authored := filepath.Join(root, "agents", "skills", "review")
	if err := os.MkdirAll(authored, 0o755); err != nil {
		t.Fatal(err)
	}
	// A symlink inside the authored tree pointing into managed output is
	// within the output through physical identity, not string prefix.
	alias := filepath.Join(authored, "alias")
	if err := os.Symlink(filepath.Join(managed, "skills"), alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	within, err := Within(managed, alias)
	if err != nil || !within {
		t.Fatalf("symlink alias within = %v, %v; want true", within, err)
	}
	within, err = Within(managed, authored)
	if err != nil || within {
		t.Fatalf("authored tree within managed = %v, %v; want false", within, err)
	}
	within, err = Within(managed, filepath.Join(managed, "skills", "review"))
	if err != nil || !within {
		t.Fatalf("managed child within = %v, %v; want true", within, err)
	}
}

func TestIsOutputPathPrunesGitAndOutputs(t *testing.T) {
	root := t.TempDir()
	outputs := []string{filepath.Join(root, ".agents")}
	for _, path := range []string{
		filepath.Join(root, ".agents", "skills", "review"),
		filepath.Join(root, ".git", "objects"),
		filepath.Join(root, "skills", ".GIT", "x"),
	} {
		pruned, err := IsOutputPath(path, outputs)
		if err != nil || !pruned {
			t.Fatalf("%s pruned = %v, %v; want true", path, pruned, err)
		}
	}
	pruned, err := IsOutputPath(filepath.Join(root, "agents", "skills", "review"), outputs)
	if err != nil || pruned {
		t.Fatalf("authored path pruned = %v, %v; want false", pruned, err)
	}
}

func TestPlanRecheckRejectsRetargetSincePlanning(t *testing.T) {
	root := t.TempDir()
	managed := filepath.Join(root, "managed")
	other := filepath.Join(root, "other")
	if err := os.MkdirAll(managed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(managed, "entry")
	var plan Plan
	plan.Replace(ClassContext, "skill-a", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Recheck(snapshot, nil); err != nil {
		t.Fatalf("unchanged plan recheck: %v", err)
	}
	// Retarget the parent between planning and publication: the live path
	// now resolves elsewhere.
	if err := os.RemoveAll(managed); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, managed); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	err = plan.Recheck(snapshot, nil)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("retargeted plan err = %v, want source_output_overlap", err)
	}
}

func TestPlanRecheckRefusesAdmittedOverwriteBothDirections(t *testing.T) {
	root := t.TempDir()
	admittedDir := filepath.Join(root, "authored")
	if err := os.MkdirAll(admittedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	admittedFile := filepath.Join(admittedDir, "SKILL.md")
	if err := os.WriteFile(admittedFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Destination inside admitted input.
	var inside Plan
	inside.Replace(ClassContext, "a", filepath.Join(admittedDir, "entry"), filepath.Join(root, "staged-a"))
	snapshot, err := inside.Snapshot([]string{admittedDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := inside.Recheck(snapshot, []string{admittedDir}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("destination inside admitted err = %v, want source_output_overlap", err)
	}
	// Admitted input inside destination.
	destDir := filepath.Join(root, "dest")
	if err := os.MkdirAll(filepath.Join(destDir, "inner"), 0o755); err != nil {
		t.Fatal(err)
	}
	innerAdmitted := filepath.Join(destDir, "inner", "file")
	if err := os.WriteFile(innerAdmitted, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var outside Plan
	outside.Replace(ClassContext, "b", destDir, filepath.Join(root, "staged-b"))
	snapshot, err = outside.Snapshot([]string{innerAdmitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := outside.Recheck(snapshot, []string{innerAdmitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("admitted inside destination err = %v, want source_output_overlap", err)
	}
	// Disjoint destination stays valid.
	var okPlan Plan
	okPlan.Replace(ClassContext, "c", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged-c"))
	snapshot, err = okPlan.Snapshot([]string{admittedDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := okPlan.Recheck(snapshot, []string{admittedDir}); err != nil {
		t.Fatalf("disjoint destination recheck: %v", err)
	}
}

func TestPlanRecheckRefusesNewlyIntroducedLinks(t *testing.T) {
	root := t.TempDir()
	live := filepath.Join(root, "live", "file")
	if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(live, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Replace the live file with a link: publication must not follow it.
	if err := os.Remove(live); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "elsewhere"), live); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := plan.Recheck(snapshot, nil); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("newly introduced link err = %v, want source_output_overlap", err)
	}
}

func TestPlanRecheckRefusesEntryLinkIntoAdmitted(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(admitted), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(admitted, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, "managed", "link")
	if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(admitted, live); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var plan Plan
	plan.ReplaceEntry(ClassAdapterLedger, "entry", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Recheck(snapshot, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("entry link into admitted err = %v, want source_output_overlap", err)
	}
	if err := plan.Recheck(snapshot, []string{admitted}); !strings.Contains(err.Error(), "overwrites admitted input") {
		t.Fatalf("entry link into admitted err = %v, want the overlap gate, not the recording gate", err)
	}
}

func TestPlanRecheckRefusesSameSpellingParentReplacement(t *testing.T) {
	// The spelling-only counterexample: renaming the planned parent away
	// and creating a new directory at the same path must fail Recheck
	// through identity comparison, even though every spelling is unchanged.
	// Uses only rename and mkdir, so it runs on every lane.
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "skill-a", filepath.Join(parent, "out"), filepath.Join(root, "staged"))
	unchanged, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Recheck(unchanged, nil); err != nil {
		t.Fatalf("unchanged plan recheck: %v", err)
	}
	// The record under test is compared exactly once, after the swap —
	// mirroring the production publisher. An earlier comparison would
	// warm the identity cache on platforms that resolve it lazily and
	// mask a missing pin.
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, filepath.Join(root, "parent-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	err = plan.Recheck(snapshot, nil)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("same-spelling parent replacement err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "changed since planning") {
		t.Fatalf("same-spelling parent replacement err = %v, want the identity gate", err)
	}
}

func TestPlanRecheckRefusesVanishedAncestor(t *testing.T) {
	// Presence turning into absence is a changed boundary too: without
	// the identity record, a removed parent would resolve through the
	// missing-suffix walk to the same spelling and pass.
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "skill-a", filepath.Join(parent, "out"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(parent); err != nil {
		t.Fatal(err)
	}
	err = plan.Recheck(snapshot, nil)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("vanished ancestor err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "no longer exists") {
		t.Fatalf("vanished ancestor err = %v, want the absence gate", err)
	}
}

func TestPlanRecheckRefusesChangedAdmittedIdentity(t *testing.T) {
	// Admitted inputs are pinned at Snapshot: swapping the object at an
	// admitted spelling fails instead of comparing separation against the
	// replacement.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	unchanged, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Recheck(unchanged, []string{admitted}); err != nil {
		t.Fatalf("unchanged plan recheck: %v", err)
	}
	// The record under test is compared exactly once, after the swap —
	// mirroring the production publisher. An earlier comparison would
	// warm the identity cache on platforms that resolve it lazily and
	// mask a missing pin.
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(admitted, filepath.Join(root, "authored-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	err = plan.Recheck(snapshot, []string{admitted})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("swapped admitted input err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "admitted input") {
		t.Fatalf("swapped admitted input err = %v, want the admitted-identity gate", err)
	}
}

func TestPlanRecheckRefusesUnrecordedAdmitted(t *testing.T) {
	// An admitted input that Snapshot never pinned fails closed: Recheck
	// must not silently skip its identity check.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	err = plan.Recheck(snapshot, []string{admitted})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("unrecorded admitted err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "not recorded at planning") {
		t.Fatalf("unrecorded admitted err = %v, want the recording gate", err)
	}
}

func TestPlanRecheckWithoutSnapshotStillEnforcesSeparation(t *testing.T) {
	// The zero Snapshot records nothing, so retarget and identity checks
	// are skipped — but destination separation is still enforced.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var inside Plan
	inside.Replace(ClassContext, "a", filepath.Join(admitted, "entry"), filepath.Join(root, "staged"))
	if err := inside.Recheck(Snapshot{}, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("destination inside admitted err = %v, want source_output_overlap", err)
	}
	var disjoint Plan
	disjoint.Replace(ClassContext, "b", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	if err := disjoint.Recheck(Snapshot{}, []string{admitted}); err != nil {
		t.Fatalf("disjoint destination with zero snapshot: %v", err)
	}
}

func TestPlanRecheckCoversTargetsAddedAfterSnapshot(t *testing.T) {
	// The commit phase merges the consumer ledger after scope planning
	// took its snapshot. Such targets have no record: a disjoint one
	// passes, one inside admitted inputs is still refused by separation.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	plan.Replace(ClassConsumer, "consumer", filepath.Join(root, "consumers", "ledger"), filepath.Join(root, "staged-ledger"))
	if err := plan.Recheck(snapshot, []string{admitted}); err != nil {
		t.Fatalf("disjoint post-snapshot target: %v", err)
	}
	plan.Targets = plan.Targets[:1]
	plan.Replace(ClassConsumer, "consumer", filepath.Join(admitted, "ledger"), filepath.Join(root, "staged-ledger"))
	if err := plan.Recheck(snapshot, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("post-snapshot target inside admitted err = %v, want source_output_overlap", err)
	}
}

func TestPlanSnapshotRefusesMissingAdmitted(t *testing.T) {
	root := t.TempDir()
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	_, err := plan.Snapshot([]string{filepath.Join(root, "absent")})
	if err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("missing admitted err = %v, want source_member_missing", err)
	}
}

func TestRecheckOnePassesUnchangedTarget(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := RecheckOne(plan.Targets[0], snapshot, []string{admitted}); err != nil {
		t.Fatalf("unchanged target: %v", err)
	}
}

func TestRecheckOneRefusesSwappedParent(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	live := filepath.Join(parent, "entry")
	plan.Replace(ClassContext, "a", live, filepath.Join(root, "staged"))
	plan.Replace(ClassContext, "b", filepath.Join(root, "managed", "other"), filepath.Join(root, "staged-other"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	// Same-spelling replacement of the first target's parent only: the
	// per-write check of that target must refuse while the untouched
	// target still passes.
	if err := os.Rename(parent, filepath.Join(root, "parent-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := RecheckOne(plan.Targets[0], snapshot, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("swapped parent err = %v, want source_output_overlap", err)
	}
	if err := RecheckOne(plan.Targets[1], snapshot, []string{admitted}); err != nil {
		t.Fatalf("untouched target: %v", err)
	}
}

func TestRecheckOneRefusesDestinationInsideAdmitted(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(admitted, "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	if err := RecheckOne(plan.Targets[0], snapshot, []string{admitted}); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("destination inside admitted err = %v, want source_output_overlap", err)
	}
}

func TestRecheckOneMatchesRecheck(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	plan.ReplaceEntry(ClassContext, "b", filepath.Join(root, "managed", "link"), filepath.Join(root, "staged-link"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	// The single-target form decides each target exactly as the whole-plan
	// form does: both pass on unchanged boundaries.
	if err := plan.Recheck(snapshot, []string{admitted}); err != nil {
		t.Fatalf("Recheck: %v", err)
	}
	for _, target := range plan.Targets {
		if err := RecheckOne(target, snapshot, []string{admitted}); err != nil {
			t.Fatalf("RecheckOne(%s): %v", target.LivePath, err)
		}
	}
}

func TestWithinNeverFollowsCommittedMirrorLinks(t *testing.T) {
	// A staged-then-committed mirror entry carries a destination string
	// that did not resolve where os.Symlink created it, so on Windows the
	// published entry is a file link pointing at a directory. Containment
	// must be decided without following it: following fails the whole
	// check there (CreateFile Access denied) even though the destinations
	// are disjoint. This test builds that exact shape — a link created
	// against an absent target, resolved only after publication — and
	// pins the disjoint answer in both directions. It fails on Windows
	// before the Lstat fix and passes everywhere after; on unix the
	// follow succeeds, so the Windows gate lane is the enforcing runner.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	liveDir := filepath.Join(root, "project", ".claude", "skills")
	stageDir := filepath.Join(root, "staging", "links")
	for _, dir := range []string{admitted, liveDir, stageDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(root, "canonical", "skill-a")
	staged := filepath.Join(stageDir, "skill-a")
	// The target does not exist yet, so the created link cannot resolve
	// at creation — the production stageLink shape.
	if err := os.Symlink(target, staged); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(liveDir, "skill-a")
	if err := os.Rename(staged, live); err != nil {
		t.Fatal(err)
	}
	// Sanity: the committed entry is a link resolving outside admitted.
	resolved, err := Canonicalize(live)
	if err != nil {
		t.Fatal(err)
	}
	targetCanonical, err := Canonicalize(target)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != targetCanonical {
		t.Fatalf("committed link resolves to %q, want %q", resolved, targetCanonical)
	}
	// The committed entry is disjoint from admitted input: deciding that
	// must not fail following the entry, in either direction.
	within, err := Within(admitted, live)
	if err != nil || within {
		t.Fatalf("committed entry within admitted = %v, %v; want false, nil", within, err)
	}
	within, err = Within(live, admitted)
	if err != nil || within {
		t.Fatalf("admitted within committed entry = %v, %v; want false, nil", within, err)
	}
	// Positive control: the entry's own tree resolves inside its target.
	within, err = Within(target, filepath.Join(live, "SKILL.md"))
	if err != nil || !within {
		t.Fatalf("entry member within target = %v, %v; want true, nil", within, err)
	}
}

func TestWithinCatchesDanglingLinkCaseAlias(t *testing.T) {
	// A case-variant spelling of the same dangling link names one object:
	// Within must report containment through identity. A Stat-based walk
	// follows the link, fails NotExist on the absent target, and skips
	// the level — missing the alias; Lstat compares the link itself.
	// Case-insensitive volumes only; elsewhere the spellings name
	// different paths and the symlink-alias test above covers SameFile.
	root := t.TempDir()
	probe := filepath.Join(root, "caseprobe-lower")
	if err := os.WriteFile(probe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	lower, lowerErr := os.Stat(probe)
	upper, upperErr := os.Stat(filepath.Join(root, "CASEPROBE-LOWER"))
	if lowerErr != nil || upperErr != nil || !os.SameFile(lower, upper) {
		t.Skip("test filesystem is case-sensitive; dangling-link case alias needs a folding volume")
	}
	link := filepath.Join(root, "alias")
	if err := os.Symlink(filepath.Join(root, "absent-target"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	folded := filepath.Join(root, "ALIAS")
	within, err := Within(folded, link)
	if err != nil || !within {
		t.Fatalf("dangling link case alias within = %v, %v; want true, nil", within, err)
	}
	// Positive control: the dangling entry still resolves under its own
	// parent, which the string comparison decides without the walk.
	within, err = Within(root, link)
	if err != nil || !within {
		t.Fatalf("dangling link within parent = %v, %v; want true, nil", within, err)
	}
}

func TestPlanRecheckIdentityReadErrorIsNotOverlap(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	target := Target{LivePath: filepath.Join(parent, "out")}
	plan := Plan{Targets: []Target{target}}
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, parent+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("parent", parent); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	err = plan.Recheck(snapshot, nil)
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

func TestRecheckDurableOneMatchesRecheck(t *testing.T) {
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	live := filepath.Join(parent, "entry")
	plan.Replace(ClassContext, "a", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	if err := RecheckDurableOne(plan.Targets[0], durable); err != nil {
		t.Fatalf("unchanged durable target: %v", err)
	}
	if err := os.Rename(parent, filepath.Join(root, "parent-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	err = RecheckDurableOne(plan.Targets[0], durable)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("swapped parent durable err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "changed since planning") {
		t.Fatalf("swapped parent durable err = %v, want the identity gate", err)
	}
}

func TestRecheckDurableOneIdentityReadErrorIsNotOverlap(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	target := Target{LivePath: filepath.Join(parent, "out")}
	plan := Plan{Targets: []Target{target}}
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, parent+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("parent", parent); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	err = RecheckDurableOne(target, durable)
	if err == nil {
		t.Fatal("expected read failure")
	}
	if strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("durable identity read failure mislabeled as overlap: %v", err)
	}
	if !strings.Contains(err.Error(), "boundary_identity_unreadable") {
		t.Fatalf("durable identity read failure err = %v, want boundary_identity_unreadable", err)
	}
}

func TestDurableKeepsPlannedIdentityAfterParentReplacement(t *testing.T) {
	// Durable must serialize the identities Snapshot captured, not
	// whatever occupies each spelling at conversion time: the parent is
	// swapped between Snapshot and Durable, so a re-statting Durable
	// would persist the replacement and the durable recheck would pass.
	// Uses only rename and mkdir, so it runs on every lane.
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(parent, "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	aged := filepath.Join(root, "parent-old")
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	// The proof still names the original parent, so the durable recheck
	// refuses the replacement through the identity gate.
	err = RecheckDurableOne(plan.Targets[0], durable)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("converted-after-swap durable err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "changed since planning") {
		t.Fatalf("converted-after-swap durable err = %v, want the identity gate", err)
	}
	// Control: restoring the original parent makes the same proof pass,
	// proving the proof carries the planned identity and stays usable.
	replacement := filepath.Join(root, "parent-replacement")
	if err := os.Rename(parent, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, parent); err != nil {
		t.Fatal(err)
	}
	if err := RecheckDurableOne(plan.Targets[0], durable); err != nil {
		t.Fatalf("restored durable recheck: %v", err)
	}
}

func TestDurableKeepsPlannedAdmittedIdentityAfterSwap(t *testing.T) {
	// The admitted-input half of the same property: swapping an admitted
	// input between Snapshot and Durable must not substitute the
	// replacement into the proof. The destination stays disjoint, so only
	// the admitted-identity gate can refuse.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	aged := filepath.Join(root, "authored-old")
	if err := os.Rename(admitted, aged); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	err = RecheckDurableOne(plan.Targets[0], durable)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("converted-after-swap admitted err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "admitted input") {
		t.Fatalf("converted-after-swap admitted err = %v, want the admitted-identity gate", err)
	}
	// Control: restoring the original input makes the same proof pass.
	replacement := filepath.Join(root, "authored-replacement")
	if err := os.Rename(admitted, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, admitted); err != nil {
		t.Fatal(err)
	}
	if err := RecheckDurableOne(plan.Targets[0], durable); err != nil {
		t.Fatalf("restored admitted durable recheck: %v", err)
	}
}

func TestSnapshotCapturePinsAncestorBeforeHookSwap(t *testing.T) {
	// The capture window: a swap landing between the pin inspection and
	// the token derivation must not split the pin. The hook fires at the
	// exact point where a second pathname-based read would observe the
	// replacement; the pin must still name the pre-swap object, because
	// both proof forms derive from the single inspection before it.
	// Uses only rename and mkdir, so it runs on every lane.
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	canonicalParent, err := Canonicalize(parent)
	if err != nil {
		t.Fatal(err)
	}
	aged := filepath.Join(root, "parent-old")
	fired := false
	var hookErr error
	SetCaptureHook(func(canonical string) {
		if fired || hookErr != nil || canonical != canonicalParent {
			return
		}
		fired = true
		if err := os.Rename(parent, aged); err != nil {
			hookErr = err
			return
		}
		if err := os.Mkdir(parent, 0o755); err != nil {
			hookErr = err
		}
	})
	t.Cleanup(func() { SetCaptureHook(nil) })
	var plan Plan
	live := filepath.Join(parent, "entry")
	plan.Replace(ClassContext, "a", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot(nil)
	SetCaptureHook(nil)
	if err != nil {
		t.Fatal(err)
	}
	if hookErr != nil {
		t.Fatal(hookErr)
	}
	if !fired {
		t.Fatal("capture hook never fired for the planned parent")
	}
	// The pin names the original parent, not the replacement the hook
	// left at its spelling.
	pin, pinned := snapshot.Targets[live].Ancestors[canonicalParent]
	if !pinned {
		t.Fatalf("planned parent %s not pinned", canonicalParent)
	}
	original, err := FileIdentity(aged)
	if err != nil {
		t.Fatal(err)
	}
	if pin.Token != original {
		t.Fatalf("ancestor pin = %q, want pre-swap identity %q", pin.Token, original)
	}
	// The live spelling now holds the replacement, so the same-process
	// recheck refuses through the identity gate.
	err = RecheckOne(plan.Targets[0], snapshot, nil)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("captured-during-swap recheck err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "changed since planning") {
		t.Fatalf("captured-during-swap recheck err = %v, want the identity gate", err)
	}
	// Control: restoring the original parent makes the same pin pass,
	// proving the pin carries the planned identity and stays usable.
	replacement := filepath.Join(root, "parent-replacement")
	if err := os.Rename(parent, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, parent); err != nil {
		t.Fatal(err)
	}
	if err := RecheckOne(plan.Targets[0], snapshot, nil); err != nil {
		t.Fatalf("restored recheck: %v", err)
	}
}

func TestSnapshotCapturePinsAdmittedBeforeHookSwap(t *testing.T) {
	// The admitted-input half of the capture window: swapping an
	// admitted input from the capture hook must not substitute the
	// replacement into the pin. The destination stays disjoint, so only
	// the admitted-identity gate can refuse.
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	canonicalAdmitted, err := Canonicalize(admitted)
	if err != nil {
		t.Fatal(err)
	}
	aged := filepath.Join(root, "authored-old")
	fired := false
	var hookErr error
	SetCaptureHook(func(canonical string) {
		if fired || hookErr != nil || canonical != canonicalAdmitted {
			return
		}
		fired = true
		if err := os.Rename(admitted, aged); err != nil {
			hookErr = err
			return
		}
		if err := os.Mkdir(admitted, 0o755); err != nil {
			hookErr = err
		}
	})
	t.Cleanup(func() { SetCaptureHook(nil) })
	var plan Plan
	plan.Replace(ClassContext, "a", filepath.Join(root, "managed", "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	SetCaptureHook(nil)
	if err != nil {
		t.Fatal(err)
	}
	if hookErr != nil {
		t.Fatal(hookErr)
	}
	if !fired {
		t.Fatal("capture hook never fired for the admitted input")
	}
	pin, pinned := snapshot.Admitted[canonicalAdmitted]
	if !pinned {
		t.Fatalf("admitted input %s not pinned", canonicalAdmitted)
	}
	original, err := FileIdentity(aged)
	if err != nil {
		t.Fatal(err)
	}
	if pin.Token != original {
		t.Fatalf("admitted pin = %q, want pre-swap identity %q", pin.Token, original)
	}
	err = RecheckOne(plan.Targets[0], snapshot, []string{admitted})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("captured-during-swap admitted err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "admitted input") {
		t.Fatalf("captured-during-swap admitted err = %v, want the admitted-identity gate", err)
	}
	// Control: restoring the original input makes the same pin pass.
	replacement := filepath.Join(root, "authored-replacement")
	if err := os.Rename(admitted, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, admitted); err != nil {
		t.Fatal(err)
	}
	if err := RecheckOne(plan.Targets[0], snapshot, []string{admitted}); err != nil {
		t.Fatalf("restored admitted recheck: %v", err)
	}
}

func TestDurableRefusesPinWithoutCapturedToken(t *testing.T) {
	// Snapshot never produces an empty token (it fails closed at
	// planning), so an empty token means a hand-built record. Durable
	// fails closed as unreadable, never as a proven overlap, and never
	// persists an empty token the journal would have to reject later.
	withoutAncestor := Snapshot{
		Targets: map[string]TargetSnapshot{
			"/live": {Canonical: "/live", Ancestors: map[string]PinnedIdentity{"/": {Token: ""}}},
		},
	}
	_, err := withoutAncestor.Durable()
	if err == nil || !strings.Contains(err.Error(), "boundary_identity_unreadable") {
		t.Fatalf("missing ancestor token err = %v, want boundary_identity_unreadable", err)
	}
	if strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("missing ancestor token mislabeled as overlap: %v", err)
	}
	withoutAdmitted := Snapshot{
		Admitted: map[string]PinnedIdentity{"/authored": {Token: ""}},
	}
	_, err = withoutAdmitted.Durable()
	if err == nil || !strings.Contains(err.Error(), "boundary_identity_unreadable") {
		t.Fatalf("missing admitted token err = %v, want boundary_identity_unreadable", err)
	}
	if strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("missing admitted token mislabeled as overlap: %v", err)
	}
}
