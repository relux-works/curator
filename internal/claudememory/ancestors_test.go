package claudememory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every strict ancestor contributes all three literal filenames, including
// missing loaders. Project files, AGENTS forms and managed policy stay allowed.
func TestAncestorFilesCoverageAndBoundary(t *testing.T) {
	base := t.TempDir()
	project := filepath.Join(base, "ancestor[1]", "project")
	if err := os.MkdirAll(project, 0o700); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(base, "CLAUDE.md")
	files, err := AncestorFiles(project, project, managed)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, file := range files {
		if seen[file] || withinMemoryRoot(project, file) || file == managed || strings.Contains(file, "AGENTS.md") {
			t.Fatalf("duplicate or protected exclusion: %s", file)
		}
		seen[file] = true
	}
	for _, name := range []string{"CLAUDE.md", filepath.Join(".claude", "CLAUDE.md"), "CLAUDE.local.md"} {
		if !seen[filepath.Join(filepath.Dir(project), name)] {
			t.Fatalf("missing literal ancestor form %s", name)
		}
		volumeRoot := filepath.VolumeName(project) + string(filepath.Separator)
		if !seen[filepath.Join(volumeRoot, name)] {
			t.Fatalf("filesystem-root form missing: %s", name)
		}
	}
}

// A symlinked project retains both ancestor spellings. An ancestor loader
// pointing into project policy is protected, rather than hidden through its alias.
func TestAncestorFilesCanonicalAliases(t *testing.T) {
	base := t.TempDir()
	physical := filepath.Join(base, "physical")
	project := filepath.Join(physical, "project")
	if err := os.MkdirAll(project, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(physical, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	allowed := filepath.Join(project, "CLAUDE.md")
	if err := os.WriteFile(allowed, []byte("project policy"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(allowed, filepath.Join(physical, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	files, err := AncestorFiles(filepath.Join(alias, "project"), filepath.Join(alias, "project"), filepath.Join(base, "managed", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, file := range files {
		seen[file] = true
	}
	canonicalPhysical, err := filepath.EvalSymlinks(physical)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{alias, canonicalPhysical} {
		if !seen[filepath.Join(dir, "CLAUDE.local.md")] {
			t.Fatalf("ancestor alias omitted: %s", dir)
		}
		if seen[filepath.Join(dir, "CLAUDE.md")] {
			t.Fatalf("project policy alias excluded: %s", dir)
		}
	}
	if seen[allowed] {
		t.Fatal("canonical project policy excluded")
	}
}

// Invalid roots and unresolved existing links fail without returning a partial
// exclusion list; nearby valid and absent-loader paths are exercised above.
func TestAncestorFilesRefusesUnresolvedPaths(t *testing.T) {
	base := t.TempDir()
	project := filepath.Join(base, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(base, "managed", "CLAUDE.md")
	for _, root := range []string{"relative", filepath.Join(base, "absent")} {
		if files, err := AncestorFiles(root, root, managed); err == nil || files != nil {
			t.Fatal("invalid project root produced exclusions")
		}
	}
	if files, err := AncestorFiles(project, project, "relative"); err == nil || files != nil {
		t.Fatal("relative managed root admitted")
	}
	if err := os.Symlink(filepath.Join(base, "missing"), filepath.Join(base, "CLAUDE.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if files, err := AncestorFiles(project, project, managed); err == nil || files != nil {
		t.Fatal("unresolved existing loader link produced a partial policy")
	}
}

// The exact-entry checker detects the current-main policy's missing ancestor,
// then accepts the repaired policy. This is a bounded checker control, not a
// substitute for witnessing the native loader on the current-main candidate.
func TestMissingAncestorCheckerControl(t *testing.T) {
	base := t.TempDir()
	native := filepath.Join(base, "operator", ".claude", "CLAUDE.md")
	ancestor := filepath.Join(base, "outer[1]", "CLAUDE.md")
	current, err := Merge(nil, native)
	if err != nil {
		t.Fatal(err)
	}
	missing, err := MissingAncestorFiles(current, []string{ancestor})
	if err != nil || len(missing) != 1 || missing[0] != ancestor {
		t.Fatalf("known missing literal ancestor not detected: %v, %v", missing, err)
	}
	repaired, _, err := Reconcile(current, nil, native, []string{ancestor})
	if err != nil {
		t.Fatal(err)
	}
	if missing, err := MissingAncestorFiles(repaired, []string{ancestor}); err != nil || len(missing) != 0 {
		t.Fatalf("valid literal ancestor rejected: %v, %v", missing, err)
	}
}

// An alias directly into a nested checkout directory must not lose its
// lexical outside-project ancestors or hide the nested project's own files.
func TestAncestorFilesNestedLaunchAlias(t *testing.T) {
	base := t.TempDir()
	project := filepath.Join(base, "physical", "project")
	nested := filepath.Join(project, "nested")
	outside := filepath.Join(base, "lexical")
	for _, dir := range []string{nested, outside} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	alias := filepath.Join(outside, "entry")
	if err := os.Symlink(nested, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	files, err := AncestorFiles(project, alias, filepath.Join(base, "managed", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, file := range files {
		seen[file] = true
		if withinMemoryRoot(project, file) {
			t.Fatalf("project file excluded through launch alias: %s", file)
		}
	}
	if !seen[filepath.Join(outside, "CLAUDE.md")] || !seen[filepath.Join(base, "physical", "CLAUDE.local.md")] {
		t.Fatal("lexical or canonical ancestor chain omitted")
	}
}

// The filesystem root has no strict ancestors; the path policy cannot turn a
// root project into an exclusion of every instruction file below it.
func TestAncestorFilesFilesystemRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.VolumeName(base) + string(filepath.Separator)
	files, err := AncestorFiles(root, root, filepath.Join(base, "managed", "CLAUDE.md"))
	if err != nil || len(files) != 0 {
		t.Fatalf("root project has ancestor exclusions: %v, %v", files, err)
	}
}

// The native exception is valid for a HOME-rooted project, but an alias to
// AGENTS, the composed root, or a whole project must refuse without link edits.
func TestNativeTargetAliasBoundaries(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "operator")
	managed := filepath.Join(base, "managed", "CLAUDE.md")
	native := filepath.Join(home, ".claude", "CLAUDE.md")
	for _, dir := range []string{filepath.Dir(native), filepath.Dir(managed)} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateNativeTargets(native, home, managed); err != nil {
		t.Fatalf("ordinary HOME native exception refused: %v", err)
	}
	agents := filepath.Join(home, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("project instructions"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(agents, native); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := ValidateNativeTargets(native, home, managed); err == nil {
		t.Fatal("native alias would exclude AGENTS")
	}
	if err := os.Remove(native); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managed, []byte("managed instructions"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(managed, native); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNativeTargets(native, home, managed); err == nil {
		t.Fatal("native alias would exclude composed policy")
	}
	if err := os.Remove(native); err != nil {
		t.Fatal(err)
	}
	rules := filepath.Join(filepath.Dir(native), "rules")
	if err := os.Symlink(home, rules); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNativeTargets(native, home, managed); err == nil {
		t.Fatal("rules alias would exclude the entire project")
	}
	if target, err := os.Readlink(rules); err != nil || target != home {
		t.Fatal("refusal changed operator link")
	}
}
