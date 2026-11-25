package whitelist

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeSelectionFile(t *testing.T, root, rel string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(rel+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func selectionFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, file := range []string{
		"SKILL.md",
		"README.md",
		"tests/example.md",
		"references/base.md",
		"references/private/nested/secret.md",
		"references/runtime/token.txt",
		"data/kept.json",
		"data/build-only.json",
		"docs/declared.json",
		"scripts/guide.md",
		"scripts/run.sh",
	} {
		writeSelectionFile(t, root, file)
	}
	return root
}

func TestContextFilesPreservesLegacyGoldenAndAppliesSelection(t *testing.T) {
	root := selectionFixture(t)
	excludedRoots := []string{"references/runtime", "data/build-only.json"}

	legacy, err := ContextFiles(root, false, excludedRoots, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantLegacy := []string{"SKILL.md", "data/kept.json", "references/base.md", "references/private/nested/secret.md"}
	if !reflect.DeepEqual(legacy, wantLegacy) {
		t.Fatalf("legacy effective files = %v, want golden %v", legacy, wantLegacy)
	}
	commandless, err := ContextFiles(root, true, excludedRoots, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantCommandless := append(append([]string(nil), wantLegacy...), "scripts/guide.md", "scripts/run.sh")
	if !reflect.DeepEqual(commandless, wantCommandless) {
		t.Fatalf("commandless legacy effective files = %v, want golden %v", commandless, wantCommandless)
	}

	selected, err := ContextFiles(root, false, excludedRoots, &FileSelection{
		Include: []string{"docs/declared.json"},
		Exclude: []string{"references/private"},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantSelected := []string{"SKILL.md", "data/kept.json", "docs/declared.json", "references/base.md"}
	if !reflect.DeepEqual(selected, wantSelected) {
		t.Fatalf("selected effective files = %v, want %v", selected, wantSelected)
	}
	selectedScript, err := ContextFiles(root, false, excludedRoots, &FileSelection{Include: []string{"scripts/guide.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if !containsFile(selectedScript, "scripts/guide.md") {
		t.Fatalf("explicit script include was not selected: %v", selectedScript)
	}
}

func containsFile(files []string, target string) bool {
	for _, file := range files {
		if file == target {
			return true
		}
	}
	return false
}

func TestContextFilesKeepsHardExclusionsWithExplicitIncludes(t *testing.T) {
	root := selectionFixture(t)
	selected, err := ContextFiles(root, false, []string{"references/runtime", "data/build-only.json"}, &FileSelection{
		Include: []string{"README.md", "tests", "references/runtime", "data/build-only.json", "docs/declared.json"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"SKILL.md", "data/kept.json", "docs/declared.json", "references/base.md", "references/private/nested/secret.md"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("explicit include bypassed a hard exclusion: got %v, want %v", selected, want)
	}
}

func TestContextFilesRejectsUnmatchedInclude(t *testing.T) {
	root := selectionFixture(t)
	_, err := ContextFiles(root, false, nil, &FileSelection{Include: []string{"missing/**"}})
	if err == nil || !strings.Contains(err.Error(), "profile_skill_files_no_match") {
		t.Fatalf("unmatched include did not have a stable refusal: %v", err)
	}
}

func TestContextFilesRejectsIncludeSymlinkOutsideSkill(t *testing.T) {
	root := selectionFixture(t)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "extras", "outside.md")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := ContextFiles(root, false, nil, &FileSelection{Include: []string{"extras/outside.md"}}); err == nil {
		t.Fatal("include followed a symlink outside the skill directory")
	}
}

func TestContextFilesRejectsSymlinkAliasesIntoExcludedPaths(t *testing.T) {
	root := selectionFixture(t)
	for _, test := range []struct {
		name         string
		target       string
		excludeRoots []string
		selection    *FileSelection
	}{
		{name: "developer exclusion", target: "README.md", selection: &FileSelection{Include: []string{"docs/readme.md"}}},
		{name: "runtime exclusion", target: "references/runtime/token.txt", excludeRoots: []string{"references/runtime"}, selection: &FileSelection{Include: []string{"docs/runtime-token.txt"}}},
		{name: "declared file exclusion", target: "references/private/nested/secret.md", selection: &FileSelection{Include: []string{"docs/private-secret.md"}, Exclude: []string{"references/private"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			link := filepath.Join(root, "docs", filepath.Base(test.selection.Include[0]))
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(root, filepath.FromSlash(test.target)), link); err != nil {
				t.Skipf("symlink creation unavailable: %v", err)
			}
			if _, err := ContextFiles(root, false, test.excludeRoots, test.selection); err == nil {
				t.Fatal("symlink alias bypassed an excluded target")
			}
		})
	}
}

func TestValidateFileSelectionRejectsUnsafeOrInvalidGlobs(t *testing.T) {
	for _, selection := range []*FileSelection{
		{Include: []string{"../outside.md"}},
		{Include: []string{"/outside.md"}},
		{Include: []string{"references/[broken"}},
		{Exclude: []string{"SKILL.md"}},
		{Exclude: []string{"agent-skill.json"}},
		{Include: []string{"references/guide.md"}, Exclude: []string{"references/guide.md"}},
		{Exclude: []string{"*.md"}},
	} {
		if err := ValidateFileSelection(selection); err == nil {
			t.Errorf("unsafe selection accepted: %+v", selection)
		}
	}
}
