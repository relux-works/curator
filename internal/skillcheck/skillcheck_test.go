package skillcheck

import (
	"encoding/json"
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

func TestValidateWarnsAboutRuntimePathInPromptContext(t *testing.T) {
	for _, runtimePath := range []string{"scripts/tool", `scripts\tool.cmd`} {
		t.Run(runtimePath, func(t *testing.T) {
			dir := t.TempDir()
			writeSkillFile(t, dir, "SKILL.md", "---\nname: skill\n---\n\nRun "+runtimePath+".\n")
			writeSkillFile(t, dir, "scripts/tool", "#!/bin/sh\n")
			writeSkillFile(t, dir, "csk-skill.json", marshal(t, map[string]any{
				"schema_version": 2,
				"runtime_roots":  []string{"scripts"},
				"commands": map[string]any{
					"tool": map[string]any{"type": "script", "unix_path": "scripts/tool"},
				},
			}))
			issues := Validate(dir, "")
			if len(issues) != 2 ||
				issues[0].Code != "skill.runtime_root_in_prompt_context" ||
				issues[1].Code != "skill.command_resolution_contract_missing" {
				t.Fatalf("issues = %+v", issues)
			}
		})
	}
}

func TestValidateWarnsWhenConsumerGuessesProviderRuntimePath(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "SKILL.md", "---\nname: consumer\n---\n\nRun scripts/tool.\n")
	writeSkillFile(t, dir, "csk-skill.json", marshal(t, map[string]any{
		"schema_version": 2,
		"commands":       map[string]any{},
		"dependencies": map[string]any{"commands": map[string]any{
			"tool": map[string]any{
				"type": "skill", "skill": "provider", "command": "tool",
			},
		}},
	}))
	issues := Validate(dir, "")
	if len(issues) != 2 ||
		issues[0].Code != "skill.provider_runtime_path_in_prompt_context" ||
		issues[1].Code != "skill.command_resolution_contract_missing" {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestValidateAcceptsCrossPlatformShellNeutralResolver(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "SKILL.md",
		"---\nname: skill\n---\n\n"+
			"Resolve .agents/bin/tool (tool.cmd on Windows), then global/bin/tool, "+
			"then command -v tool or Get-Command tool.\n")
	writeSkillFile(t, dir, "README.md", "Development source: scripts/tool.\n")
	writeSkillFile(t, dir, "scripts/tool", "#!/bin/sh\n")
	writeSkillFile(t, dir, "scripts/tool.cmd", "@echo off\r\n")
	writeSkillFile(t, dir, "csk-skill.json", marshal(t, map[string]any{
		"schema_version": 2,
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			"tool": map[string]any{
				"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool.cmd",
			},
		},
	}))
	if issues := Validate(dir, ""); len(issues) != 0 {
		t.Fatalf("shell-neutral skill warnings = %+v", issues)
	}
}

func writeSkillFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func marshal(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
