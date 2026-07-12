package gitignore

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestEnsureFailsWithoutIgnoreThenFixes(t *testing.T) {
	project := gitProject(t)
	entries := []string{".agents/", ".claude/skills/"}

	err := Ensure(project, entries, false)
	if err == nil || !strings.Contains(err.Error(), ".agents/") {
		t.Fatalf("err = %v, want missing entries", err)
	}

	if err := Ensure(project, entries, true); err != nil {
		t.Fatalf("fix failed: %v", err)
	}
	payload, _ := os.ReadFile(filepath.Join(project, ".gitignore"))
	text := string(payload)
	if !strings.Contains(text, BlockComment) || !strings.Contains(text, ".agents/") {
		t.Fatalf("gitignore content:\n%s", text)
	}

	// idempotent: no duplicate lines on a second fix
	if err := Ensure(project, entries, true); err != nil {
		t.Fatal(err)
	}
	payload, _ = os.ReadFile(filepath.Join(project, ".gitignore"))
	if strings.Count(string(payload), ".agents/") != 1 {
		t.Fatalf("duplicated entries:\n%s", payload)
	}
}

func TestEnsurePassesWhenAlreadyIgnored(t *testing.T) {
	project := gitProject(t)
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(project, []string{".agents/"}, false); err != nil {
		t.Fatalf("already-ignored entry reported missing: %v", err)
	}
}
