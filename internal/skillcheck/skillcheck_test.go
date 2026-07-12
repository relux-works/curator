package skillcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateMissingSkillAndInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "csk-skill.json"), []byte(`{"schema_version":99}`), 0o644); err != nil {
		t.Fatal(err)
	}
	issues := Validate(dir, "")
	if len(issues) != 2 || issues[0].Code != "skill.missing_skill_md" || issues[1].Code != "skill.manifest_invalid" {
		t.Fatalf("issues = %+v", issues)
	}
	if !HasErrors(issues) || !strings.Contains(Format(issues[0]), "SKILL.md") {
		t.Fatalf("error helpers rejected issues: %+v", issues)
	}
}

func TestValidateLocaleWarning(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("SKILL.md", "---\nname: skill\n---\n")
	write("locales/metadata.json", `{"locales":{"en":{"description":"English"}}}`)
	write(".skill_triggers/en.md", "- trigger\n")
	issues := Validate(dir, "ru")
	if len(issues) != 1 || issues[0].Severity != "warning" || HasErrors(issues) {
		t.Fatalf("issues = %+v", issues)
	}
}
