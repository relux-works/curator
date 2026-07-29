package skillcheck

import (
	"encoding/json"
	"fmt"
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

func TestValidateWarnsAboutBuildSourcePathInPromptContext(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		token     string
	}{
		{name: "posix", reference: "assets/build-tool/cmd/tool", token: "assets/build-tool/"},
		{name: "windows", reference: `assets\build-tool\cmd\tool`, token: `assets\build-tool\`},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			writeSkillFile(t, dir, "SKILL.md",
				"---\nname: skill\n---\n\nRun "+testCase.reference+".\n"+
					"Resolve .agents/bin/tool.cmd, then global/bin/tool, then command -v tool or Get-Command tool.\n")
			writeSkillFile(t, dir, "assets/build-tool/go.mod", "module example.com/tool\n")
			writeSkillFile(t, dir, "assets/build-tool/cmd/tool/main.go", "package main\nfunc main() {}\n")
			writeSkillFile(t, dir, "agent-skill.json", marshal(t, map[string]any{
				"schema_version": 6,
				"build_roots":    []string{"assets/build-tool"},
				"capabilities":   map[string]any{},
				"commands": map[string]any{
					"tool": map[string]any{
						"type": "build", "driver": "go-v1", "source_dir": "assets/build-tool/cmd/tool",
					},
				},
			}))
			issues := Validate(dir, "")
			if len(issues) != 1 {
				t.Fatalf("issues = %+v", issues)
			}
			issue := issues[0]
			wantMessage := fmt.Sprintf(
				"prompt-visible text references build-only path %q; Curator excludes that build root from installed skill context and runtime storage. Use the exported compiled command shim and keep manifest-relative build source paths source-checkout-only",
				testCase.token,
			)
			if issue.Severity != "warning" || issue.Code != "skill.build_root_in_prompt_context" ||
				issue.Path != "SKILL.md" || issue.Message != wantMessage {
				t.Fatalf("issue = %+v", issue)
			}
		})
	}
}

func TestValidateDoesNotScanMarkdownInsideExcludedBuildRoot(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "SKILL.md",
		"---\nname: skill\n---\n\nResolve .agents/bin/tool.cmd, then global/bin/tool, then command -v tool or Get-Command tool.\n")
	writeSkillFile(t, dir, "assets/build-tool/go.mod", "module example.com/tool\n")
	writeSkillFile(t, dir, "assets/build-tool/cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeSkillFile(t, dir, "assets/build-tool/docs/source.md", "Source: assets/build-tool/cmd/tool.\n")
	writeSkillFile(t, dir, "agent-skill.json", marshal(t, map[string]any{
		"schema_version": 6,
		"build_roots":    []string{"assets/build-tool"},
		"capabilities":   map[string]any{},
		"commands": map[string]any{
			"tool": map[string]any{
				"type": "build", "driver": "go-v1", "source_dir": "assets/build-tool/cmd/tool",
			},
		},
	}))
	if issues := Validate(dir, ""); len(issues) != 0 {
		t.Fatalf("excluded build-root Markdown produced warnings: %+v", issues)
	}
}

func TestValidateBuildCommandRequiresShellNeutralResolver(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "SKILL.md", "---\nname: skill\n---\n")
	writeSkillFile(t, dir, "build/go.mod", "module example.com/tool\n")
	writeSkillFile(t, dir, "build/cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeSkillFile(t, dir, "agent-skill.json", marshal(t, map[string]any{
		"schema_version": 6,
		"build_roots":    []string{"build"},
		"capabilities":   map[string]any{},
		"commands": map[string]any{
			"tool": map[string]any{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool"},
		},
	}))
	issues := Validate(dir, "")
	if len(issues) != 1 || issues[0].Code != "skill.command_resolution_contract_missing" ||
		!strings.Contains(issues[0].Message, "Windows .cmd shim suffix") {
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
