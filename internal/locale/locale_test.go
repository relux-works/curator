package locale

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lay(t *testing.T, root string, files map[string]string) string {
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
	return root
}

const metadata = `{"locales": {"ru": {"description": "Русское описание",
	"display_name": "Имя", "short_description": "Кратко", "default_prompt": "Промпт"}}}`

func localizedSkill(t *testing.T) string {
	return lay(t, t.TempDir(), map[string]string{
		"SKILL.md":              "---\nname: skill-a\ndescription: english\n---\n\n# body\n",
		"locales/metadata.json": metadata,
		".skill_triggers/ru.md": "- триггер один\n```\n- not a trigger\n```\n- 'триггер два'\n",
		"agents/openai.yaml":    "interface:\n  display_name: \"en\"\n",
	})
}

func TestNoLocaleSelectedMeansNoChecks(t *testing.T) {
	snapshot := lay(t, t.TempDir(), map[string]string{
		"SKILL.md":              "x",
		"locales/metadata.json": `{"broken`, // would fail if analyzed
	})
	analysis := Analyze(snapshot, "")
	if analysis.Failed() || analysis.LocaleToRender != "" {
		t.Fatalf("no selected locale must skip checks: %+v", analysis)
	}
}

func TestNoMetadataInstallsUnderAnyLocale(t *testing.T) {
	snapshot := lay(t, t.TempDir(), map[string]string{"SKILL.md": "x"})
	analysis := Analyze(snapshot, "ru")
	if analysis.Failed() || len(analysis.Issues) != 0 {
		t.Fatalf("skill without localization must pass: %+v", analysis)
	}
}

func TestConsistentLocaleRenders(t *testing.T) {
	snapshot := localizedSkill(t)
	installed := lay(t, t.TempDir(), map[string]string{
		"SKILL.md":           "---\nname: skill-a\ndescription: english\n---\n\n# body\n",
		"agents/openai.yaml": "interface:\n  display_name: \"en\"\n",
	})
	issues, err := Render(snapshot, installed, "ru")
	if err != nil || len(issues) != 0 {
		t.Fatalf("render: %v %v", issues, err)
	}
	skill, _ := os.ReadFile(filepath.Join(installed, "SKILL.md"))
	text := string(skill)
	if !strings.Contains(text, "name: skill-a") {
		t.Fatalf("name lost:\n%s", text)
	}
	if !strings.Contains(text, `description: "Русское описание"`) {
		t.Fatalf("description not rendered:\n%s", text)
	}
	if !strings.Contains(text, `- "триггер один"`) || !strings.Contains(text, `- "триггер два"`) {
		t.Fatalf("triggers not rendered:\n%s", text)
	}
	if strings.Contains(text, "not a trigger") {
		t.Fatalf("code-fenced line leaked into triggers:\n%s", text)
	}
	if !strings.Contains(text, "# body") {
		t.Fatalf("body lost:\n%s", text)
	}
	openai, _ := os.ReadFile(filepath.Join(installed, "agents", "openai.yaml"))
	if !strings.Contains(string(openai), `display_name: "Имя"`) {
		t.Fatalf("openai.yaml not rendered:\n%s", openai)
	}
}

func TestUnavailableLocaleFallsBackWithWarning(t *testing.T) {
	snapshot := localizedSkill(t)
	installed := lay(t, t.TempDir(), map[string]string{"SKILL.md": "---\nname: skill-a\ndescription: english\n---\n"})
	issues, err := Render(snapshot, installed, "hy")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Severity != "warning" || issues[0].Code != "locale.selected_unavailable" {
		t.Fatalf("issues: %+v", issues)
	}
	skill, _ := os.ReadFile(filepath.Join(installed, "SKILL.md"))
	if !strings.Contains(string(skill), "description: english") {
		t.Fatalf("source SKILL.md must stay unrendered:\n%s", skill)
	}
}

func TestInconsistentCatalogsFail(t *testing.T) {
	// metadata without a matching trigger catalog
	snapshot := lay(t, t.TempDir(), map[string]string{
		"SKILL.md":              "x",
		"locales/metadata.json": metadata,
		".skill_triggers/en.md": "- t",
	})
	analysis := Analyze(snapshot, "ru")
	if !analysis.Failed() || analysis.Issues[0].Code != "locale.no_consistent_catalog" {
		t.Fatalf("analysis: %+v", analysis)
	}

	// triggers without metadata
	snapshot = lay(t, t.TempDir(), map[string]string{
		"SKILL.md":              "x",
		".skill_triggers/ru.md": "- t",
	})
	analysis = Analyze(snapshot, "ru")
	if !analysis.Failed() || analysis.Issues[0].Code != "locale.metadata_missing" {
		t.Fatalf("analysis: %+v", analysis)
	}
}
