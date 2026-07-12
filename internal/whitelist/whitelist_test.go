package whitelist

import (
	"os"
	"path/filepath"
	"testing"
)

func lay(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCopyContextWhitelist(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{
		"SKILL.md":                 "skill",
		"references/details.md":    "ref",
		"agents/openai.yaml":       "meta",
		".skill_triggers/en.md":    "- trigger",
		"assets/logo.txt":          "logo",
		"scripts/tool":             "runtime",
		"tests/x_test.py":          "test",
		"README.md":                "readme",
		"Makefile":                 "make",
		"references/tests/hid.md":  "nested tests dir excluded",
		"references/.DS_Store":     "noise",
		"data/rows.json":           "data",
		"pyproject.toml":           "build",
		"references/CHANGELOG.md":  "log",
		"examples/basic/run.md":    "ex",
		"templates/tpl.md":         "tpl",
		".github/workflows/ci.yml": "ci",
	})
	dest := filepath.Join(t.TempDir(), "ctx")
	files, err := CopyContext(snapshot, dest, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"SKILL.md": true, "references/details.md": true, "agents/openai.yaml": true,
		".skill_triggers/en.md": true, "assets/logo.txt": true, "data/rows.json": true,
		"examples/basic/run.md": true, "templates/tpl.md": true,
	}
	got := map[string]bool{}
	for _, file := range files {
		got[file] = true
	}
	for file := range want {
		if !got[file] {
			t.Errorf("missing from context: %s", file)
		}
	}
	for _, banned := range []string{"scripts/tool", "tests/x_test.py", "README.md", "Makefile",
		"references/tests/hid.md", "references/.DS_Store", "pyproject.toml", "references/CHANGELOG.md"} {
		if got[banned] {
			t.Errorf("must be excluded: %s", banned)
		}
	}
}

func TestScriptsIncludedForCommandlessSkill(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{"SKILL.md": "s", "scripts/helper.py": "h"})
	dest := filepath.Join(t.TempDir(), "ctx")
	files, err := CopyContext(snapshot, dest, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range files {
		if file == "scripts/helper.py" {
			found = true
		}
	}
	if !found {
		t.Fatalf("scripts must be included for command-less skills: %v", files)
	}
}

func TestRuntimeRootsExcludedFromContext(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{
		"SKILL.md":       "s",
		"agents/tool.py": "runtime under a whitelisted root",
		"agents/info.md": "context",
	})
	dest := filepath.Join(t.TempDir(), "ctx")
	files, err := CopyContext(snapshot, dest, false, []string{"agents"})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file == "agents/tool.py" || file == "agents/info.md" {
			t.Fatalf("runtime root leaked into context: %v", files)
		}
	}
}

func TestMissingSkillMdFails(t *testing.T) {
	if _, err := CopyContext(t.TempDir(), filepath.Join(t.TempDir(), "ctx"), false, nil); err == nil {
		t.Fatal("missing SKILL.md must fail")
	}
}

func TestDestinationReplaced(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{"SKILL.md": "new"})
	dest := t.TempDir()
	lay(t, dest, map[string]string{"stale.md": "old"})
	if _, err := CopyContext(snapshot, dest, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "stale.md")); err == nil {
		t.Fatal("stale content survived the copy")
	}
}
