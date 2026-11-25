package adapters

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
)

func makeSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".csk-install.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeLedgerFor(t *testing.T, adapterRoot string, names ...string) {
	t.Helper()
	entries := map[string]bool{}
	for _, name := range names {
		entries[name] = true
	}
	payload, err := ledgerPayload(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(adapterRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adapterRoot, LedgerName), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// find returns the staged target for one live path.
func find(plan staging.Plan, livePath string) (staging.Target, bool) {
	for _, target := range plan.Targets {
		if target.LivePath == livePath {
			return target, true
		}
	}
	return staging.Target{}, false
}

func TestStageProjectJournalsEveryMirrorAndTheLedger(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	makeSkill(t, canonical, "skill-b")

	mirror, err := StageProject(stageRoot, project, []string{"claude_code", "cursor"},
		[]Group{{Root: canonical, Skills: []string{"skill-a", "skill-b"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	plan := mirror.Plan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("staged adapter plan is invalid: %v", err)
	}

	// Nothing may exist live: staging derives desired state only.
	for _, rel := range []string{".claude/skills", ".cursor/rules"} {
		if _, err := os.Lstat(filepath.Join(project, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("staging touched the live adapter root %s: %v", rel, err)
		}
	}

	for _, rel := range []string{
		".claude/skills/skill-a", ".claude/skills/skill-b",
		".cursor/rules/skill-a", ".cursor/rules/skill-b",
	} {
		live := filepath.Join(project, filepath.FromSlash(rel))
		target, found := find(plan, live)
		if !found {
			t.Fatalf("mirror entry %s is not a journaled target", rel)
		}
		if target.Class != staging.ClassAdapterLedger {
			t.Fatalf("mirror entry %s committed in class %q", rel, target.Class)
		}
		if _, err := os.Stat(filepath.Join(target.StagedPath, "SKILL.md")); err != nil {
			t.Fatalf("staged copy of %s is incomplete: %v", rel, err)
		}
	}
	for _, rel := range []string{".claude/skills", ".cursor/rules"} {
		live := filepath.Join(project, filepath.FromSlash(rel), LedgerName)
		if _, found := find(plan, live); !found {
			t.Fatalf("ownership ledger of %s is not a journaled target", rel)
		}
	}
}

func TestStageProjectRefusesAnUnreadableAdapterLedger(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	adapterRoot := filepath.Join(project, ".claude", "skills")
	writeLedgerFor(t, adapterRoot, "stale-skill")
	ledgerPath := filepath.Join(adapterRoot, LedgerName)
	original, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(ledgerPath, 0o000); err != nil {
		t.Skipf("this host cannot create mode-000 manager state: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(ledgerPath, 0o644) })
	if _, err := os.ReadFile(ledgerPath); err == nil {
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
	if _, err := StageProject(stageRoot, project, []string{"claude_code"}, nil, "copy"); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("StageProject with unreadable adapter ledger error = %v, want typed unreadable refusal", err)
	}
	if err := os.Chmod(ledgerPath, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(ledgerPath)
	if err != nil || string(got) != string(original) {
		t.Fatalf("unreadable adapter ledger changed = (%q, %v), want original bytes", got, err)
	}
}

func TestStaleManagedEntriesBecomeJournaledRemovals(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	adapterRoot := filepath.Join(project, ".claude", "skills")
	writeLedgerFor(t, adapterRoot, "skill-a", "skill-gone")
	stale := filepath.Join(adapterRoot, "skill-gone")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stale, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	target, found := find(mirror.Plan(), stale)
	if !found {
		t.Fatal("a stale managed entry produced no journaled removal")
	}
	if target.Class != staging.ClassRemoval || !target.Removal() {
		t.Fatalf("stale entry target = %+v, want an absent desired state in the removal class", target)
	}
	if target.Kind != staging.KindEntry {
		t.Fatalf("stale entry kind = %q, want the entry kind so a link is restorable", target.Kind)
	}
	if _, err := os.Stat(filepath.Join(stale, "SKILL.md")); err != nil {
		t.Fatalf("staging removed the stale entry itself: %v", err)
	}
}

func TestStaleSpecialFileIsReportedRatherThanRemoved(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	adapterRoot := filepath.Join(project, ".claude", "skills")
	writeLedgerFor(t, adapterRoot, "skill-a", "skill-fifo")
	fifo := filepath.Join(adapterRoot, "skill-fifo")
	if err := makeFIFO(fifo); err != nil {
		t.Skipf("named pipes unavailable: %v", err)
	}

	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	if _, found := find(mirror.Plan(), fifo); found {
		t.Fatal("a special file was journaled for removal")
	}
	if len(mirror.Messages) == 0 || !strings.Contains(strings.Join(mirror.Messages, " "), "skill-fifo") {
		t.Fatalf("the untouched special file was not reported: %v", mirror.Messages)
	}
}

func TestSymlinkModeStagesRelativeLinkTargets(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "symlink")
	if err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	live := filepath.Join(project, ".claude", "skills", "skill-a")
	target, found := find(mirror.Plan(), live)
	if !found {
		t.Fatal("the symlink mirror entry is not a journaled target")
	}
	if target.Kind != staging.KindEntry {
		t.Fatalf("symlink mirror kind = %q, want the entry kind", target.Kind)
	}
	destination, err := os.Readlink(target.StagedPath)
	if err != nil {
		t.Fatalf("the staged mirror is not a symlink: %v", err)
	}
	if filepath.IsAbs(destination) {
		t.Fatalf("symlink must be relative: %s", destination)
	}
	// The destination is expressed from the live entry's own directory, so it
	// resolves once the journal renames the staged link into place.
	if resolved := filepath.Join(filepath.Dir(live), destination); resolved != filepath.Join(canonical, "skill-a") {
		t.Fatalf("staged link resolves to %s, want %s", resolved, filepath.Join(canonical, "skill-a"))
	}
}

func TestSymlinkModeStagesNothingWhenTheLiveLinkIsAlreadyCorrect(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	adapterRoot := filepath.Join(project, ".claude", "skills")
	writeLedgerFor(t, adapterRoot, "skill-a")
	live := filepath.Join(adapterRoot, "skill-a")
	destination := linkDestination(filepath.Join(canonical, "skill-a"), live)
	if err := os.Symlink(destination, live); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "symlink")
	if err != nil {
		t.Fatal(err)
	}
	if _, found := find(mirror.Plan(), live); found {
		t.Fatal("an already-correct link was journaled for replacement")
	}
	if _, found := find(mirror.Plan(), filepath.Join(adapterRoot, LedgerName)); !found {
		t.Fatal("the ownership ledger stopped being a target")
	}
}

func TestModeTransitionReplacesTheDirectoryEntryItself(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	adapterRoot := filepath.Join(project, ".claude", "skills")
	writeLedgerFor(t, adapterRoot, "skill-a")
	live := filepath.Join(adapterRoot, "skill-a")
	if err := os.Symlink(linkDestination(filepath.Join(canonical, "skill-a"), live), live); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	mirror, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	target, found := find(mirror.Plan(), live)
	if !found {
		t.Fatal("switching a live link to copy mode produced no target")
	}
	if target.Kind != staging.KindEntry {
		t.Fatalf("copy over a live link has kind %q, want the entry kind so the link is the preimage", target.Kind)
	}
	if _, err := os.Stat(filepath.Join(target.StagedPath, "SKILL.md")); err != nil {
		t.Fatalf("the staged replacement is not a copied tree: %v", err)
	}
}

func TestUnmanagedConflictRefused(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	// a hand-placed directory without a marker in the adapter root
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
		t.Fatalf("err = %v, want unmanaged conflict", err)
	}
	if _, statErr := os.Stat(filepath.Join(foreign, "mine.md")); statErr != nil {
		t.Fatal("foreign content was destroyed")
	}
}

func TestMarkerDirectoryIsAdoptable(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	// a copied directory carrying an install marker counts as ours
	adopted := filepath.Join(project, ".claude", "skills", "skill-a")
	if err := os.MkdirAll(adopted, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adopted, ".csk-install.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := StageProject(stageRoot, project, []string{"claude_code"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy"); err != nil {
		t.Fatalf("marker directory must be adoptable: %v", err)
	}
}

func TestNativeDiscoveryAgentsGetNoProjectMirror(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	mirror, err := StageProject(stageRoot, project, []string{"opencode", "windsurf"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	if len(mirror.Plan().Targets) != 0 {
		t.Fatalf("native-discovery agents produced project targets: %+v", mirror.Plan().Targets)
	}
}

func TestWindsurfNativeDiscoveryLegacyContract(t *testing.T) {
	if !KnownAgents()["windsurf"] {
		t.Fatal("native-discovery agent was not recognized")
	}
	if unknown := UnknownAgents([]string{"windsurf"}); len(unknown) != 0 {
		t.Fatalf("native-discovery agent was rejected: %v", unknown)
	}
	if entries := RequiredGitignoreEntries([]string{"windsurf"}); len(entries) != 1 || entries[0] != ".agents/" {
		t.Fatalf("native store ignore entries: %v", entries)
	}

	project := t.TempDir()
	projectStage := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	projectMirror, err := StageProject(projectStage, project, []string{"windsurf"},
		[]Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	if len(projectMirror.Plan().Targets) != 0 {
		t.Fatalf("native discovery produced a project mirror: %+v", projectMirror.Plan().Targets)
	}

	home := t.TempDir()
	userHome := t.TempDir()
	globalStage := t.TempDir()
	globalCanonical := filepath.Join(home, "global", "skills")
	makeSkill(t, globalCanonical, "skill-a")
	globalMirror, err := StageGlobal(globalStage, home, userHome, []string{"windsurf"}, []string{"skill-a"}, "copy", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := find(globalMirror.Plan(), filepath.Join(userHome, ".agents", "skills", "skill-a")); !found {
		t.Fatal("global native-discovery mirror is missing")
	}
}

func TestRegisteredSkillRoutingFallbackAndPathExpansion(t *testing.T) {
	if _, ported, err := registeredSkills("cursor", harnessapi.ProjectScope); err != nil || !ported {
		t.Fatalf("registered Cursor adapter did not select its registry declaration: ported=%t err=%v", ported, err)
	}
	if _, ported, err := registeredSkills("claude_code", harnessapi.ProjectScope); err != nil || ported {
		t.Fatalf("unported adapter did not select the explicit legacy fallback: ported=%t err=%v", ported, err)
	}
	if _, ported, err := registeredSkills("unknown", harnessapi.ProjectScope); err != nil || ported {
		t.Fatalf("unknown adapter changed the legacy empty result: ported=%t err=%v", ported, err)
	}
	if _, _, err := registeredSkills("cursor", "unsupported"); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("registry context error was masked: %v", err)
	}

	var relative []harnessvalue.RelativePath
	for _, name := range []string{".agents/skills", "shared/skills"} {
		path, err := harnessvalue.NewRelativePath(name)
		if err != nil {
			t.Fatal(err)
		}
		relative = append(relative, path)
	}
	paths, err := harnessapi.NewNamedPaths(relative)
	if err != nil {
		t.Fatal(err)
	}
	roots := map[string]string{}
	addSkillRoots(roots, "adapter", t.TempDir(), paths)
	if len(roots) != 2 || roots["adapter"] == "" || roots["adapter#1"] == "" {
		t.Fatalf("named skill roots were not preserved: %v", roots)
	}
}

func TestStageGlobalCoversTheNativeDiscoveryMirrorOnce(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(home, "global", "skills")
	makeSkill(t, canonical, "skill-g")

	mirror, err := StageGlobal(stageRoot, home, userHome,
		[]string{"claude_code", "opencode", "windsurf"}, []string{"skill-g"}, "copy", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := mirror.Plan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("staged global adapter plan is invalid: %v", err)
	}
	for _, rel := range []string{".claude/skills/skill-g", NativeDiscoveryHomePath + "/skill-g"} {
		if _, found := find(plan, filepath.Join(userHome, filepath.FromSlash(rel))); !found {
			t.Fatalf("global mirror %s is not a journaled target", rel)
		}
	}
}

func TestStageGlobalMirrorsCursorAtItsExistingHomePath(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(home, "global", "skills")
	makeSkill(t, canonical, "skill-c")

	mirror, err := StageGlobal(stageRoot, home, userHome,
		[]string{"cursor"}, []string{"skill-c"}, "copy", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := mirror.Plan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("staged global adapter plan is invalid: %v", err)
	}
	root := filepath.Join(userHome, ".cursor", "rules")
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("staging touched the live Cursor root: %v", err)
	}
	entry, found := find(plan, filepath.Join(root, "skill-c"))
	if !found {
		t.Fatal("Cursor home skill directory is not a journaled target")
	}
	if _, err := os.Stat(filepath.Join(entry.StagedPath, "SKILL.md")); err != nil {
		t.Fatalf("staged Cursor skill is incomplete: %v", err)
	}
	if _, found := find(plan, filepath.Join(root, LedgerName)); !found {
		t.Fatal("Cursor home ownership ledger is not a journaled target")
	}
}

func TestGitignoreEntriesAndUnknownAgents(t *testing.T) {
	entries := RequiredGitignoreEntries([]string{"claude_code", "opencode", "bogus"})
	joined := strings.Join(entries, " ")
	if !strings.Contains(joined, ".agents/") || !strings.Contains(joined, ".claude/skills/") {
		t.Fatalf("entries: %v", entries)
	}
	if strings.Contains(joined, "bogus") {
		t.Fatalf("unknown agent produced an entry: %v", entries)
	}
	unknown := UnknownAgents([]string{"claude_code", "bogus", "bogus"})
	if len(unknown) != 1 || unknown[0] != "bogus" {
		t.Fatalf("unknown: %v", unknown)
	}
}

// Gemini's project and user skill projections retain their existing native
// destination while the path facts move behind the adapter contract.
func TestGeminiSkillProjectionDestinations(t *testing.T) {
	project := t.TempDir()
	stageRoot := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-g")

	projectMirror, err := StageProject(stageRoot, project, []string{"gemini"},
		[]Group{{Root: canonical, Skills: []string{"skill-g"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	projectTarget := filepath.Join(project, ".gemini", "skills", "skill-g")
	if _, found := find(projectMirror.Plan(), projectTarget); !found {
		t.Fatalf("project mirror missing existing Gemini destination %s", projectTarget)
	}
	entries := strings.Join(RequiredGitignoreEntries([]string{"gemini"}), "\n")
	if !strings.Contains(entries, ".gemini/skills/") {
		t.Fatalf("generated-path ignore entry missing Gemini destination: %q", entries)
	}

	home := t.TempDir()
	userHome := t.TempDir()
	globalStage := t.TempDir()
	globalCanonical := filepath.Join(home, "global", "skills")
	makeSkill(t, globalCanonical, "skill-g")
	globalMirror, err := StageGlobal(globalStage, home, userHome, []string{"gemini"},
		[]string{"skill-g"}, "copy", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	userTarget := filepath.Join(userHome, ".gemini", "skills", "skill-g")
	if _, found := find(globalMirror.Plan(), userTarget); !found {
		t.Fatalf("user mirror missing existing Gemini destination %s", userTarget)
	}
}

func TestGeminiRemainsARecognizedSkillTarget(t *testing.T) {
	if unknown := UnknownAgents([]string{"gemini"}); len(unknown) != 0 {
		t.Fatalf("registered skill target rejected as unknown: %v", unknown)
	}
}
