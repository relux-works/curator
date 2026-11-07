package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/skillcheck"
)

func TestSkillCheckJSONCleanSkill(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "---\nname: clean-skill\n---\n\nA clean skill.\n")
	configPath := filepath.Join(t.TempDir(), "config.json")

	// Both flag positions are supported by the interspersed CLI parser.
	for _, args := range [][]string{
		{"skill", "check", dir, "--json"},
		{"skill", "check", "--json", dir},
	} {
		code, stdout, stderr := capture(t, configPath, args...)
		if code != exitOK || stdout != "[]\n" || stderr != "" {
			t.Errorf("clean JSON check = (%d, %q, %q), want (0, %q, %q)", code, stdout, stderr, "[]\n", "")
		}
	}
	code, stdout, stderr := capture(t, configPath, "skill", "check", dir)
	if code != exitOK || stdout != dir+": ok\n" || stderr != "" {
		t.Errorf("clean human check = (%d, %q, %q)", code, stdout, stderr)
	}
}

func TestSkillCheckJSONFindings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, skill, locale, severity, diagnostic string
		exit                                      int
	}{
		{name: "error", severity: "error", diagnostic: "skill.missing_skill_md", exit: exitFail},
		{name: "warning", skill: "---\nname: warning-skill\n---\n", locale: "ru", severity: "warning", diagnostic: "locale.selected_unavailable", exit: exitOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.skill != "" {
				writeFile(t, filepath.Join(dir, "SKILL.md"), tc.skill)
				writeFile(t, filepath.Join(dir, "locales", "metadata.json"), `{"locales":{"en":{"description":"English"}}}`)
				writeFile(t, filepath.Join(dir, ".skill_triggers", "en.md"), "- trigger\n")
			}
			code, stdout, stderr := capture(t, filepath.Join(t.TempDir(), "config.json"), "skill", "check", dir, "--json", "--locale", tc.locale)
			var issues []skillcheck.Issue
			if err := json.Unmarshal([]byte(stdout), &issues); err != nil {
				t.Fatalf("decode findings: %v; stdout = %q", err, stdout)
			}
			if code != tc.exit || stderr != "" || len(issues) != 1 {
				t.Fatalf("check = (%d, %q, %q), want exit %d and one issue", code, stdout, stderr, tc.exit)
			}
			if issues[0].Severity != tc.severity || issues[0].Code != tc.diagnostic || issues[0].Message == "" {
				t.Fatalf("issue = %+v", issues[0])
			}
		})
	}
}
