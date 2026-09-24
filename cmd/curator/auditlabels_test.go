package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/manifest"
)

// auditLabelProject lays out a project declaring one single-command skill
// shaped like an audit-label vector case, and returns the config path that
// points the CLI at it. enforced selects the `script-worker-v1` policy with
// the closed `python3-v1` interpreter; network selects the declared
// `network` hosts (nil declares none).
func auditLabelProject(t *testing.T, schema int, enforced bool, network []string) string {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	skillRepo := filepath.Join(skillsRoot, "label-skill")
	if err := os.MkdirAll(skillRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, skillRepo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(skillRepo, "SKILL.md"),
		[]byte("---\nname: label-skill\ndescription: d\n---\n# Label\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillRepo, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillRepo, "scripts", "tool"),
		[]byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command := map[string]any{"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool"}
	if enforced {
		command["execution_policy"] = "script-worker-v1"
		command["interpreter"] = "python3-v1"
	}
	declaredNetwork := any("none")
	if network != nil {
		declaredNetwork = network
	}
	skillManifest, err := json.Marshal(map[string]any{
		"schema_version": schema,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": declaredNetwork, "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands":      map[string]any{"tool": command},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillRepo, "agent-skill.json"), skillManifest, 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, skillRepo, "add", ".")
	runGit(t, skillRepo, "commit", "-qm", "initial skill")
	runGit(t, skillRepo, "tag", "v1")

	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"),
		[]byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"label-skill","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	return configPath
}

// auditLabelMixedProject lays out a project declaring one schema-8 skill
// with two script commands: a declared-only `tool` and an enforced
// `guarded` with no declared network hosts.
func auditLabelMixedProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	skillRepo := filepath.Join(skillsRoot, "label-skill")
	if err := os.MkdirAll(skillRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, skillRepo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(skillRepo, "SKILL.md"),
		[]byte("---\nname: label-skill\ndescription: d\n---\n# Label\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillRepo, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"tool", "guarded"} {
		if err := os.WriteFile(filepath.Join(skillRepo, "scripts", script),
			[]byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	skillManifest, err := json.Marshal(map[string]any{
		"schema_version": 8,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": "none", "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands": map[string]any{
			"tool":    map[string]any{"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool"},
			"guarded": map[string]any{"type": "script", "unix_path": "scripts/guarded", "win_path": "scripts/guarded", "execution_policy": "script-worker-v1", "interpreter": "python3-v1"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillRepo, "agent-skill.json"), skillManifest, 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, skillRepo, "add", ".")
	runGit(t, skillRepo, "commit", "-qm", "initial skill")
	runGit(t, skillRepo, "tag", "v1")

	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"),
		[]byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"label-skill","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	return configPath
}

// TestCLIAuditEmitsScriptLabels drives the four audit-label shapes through
// the production `curator audit` entry. Labelled skills warn with the exact
// vector text and exit zero; the enforced skill with no declared hosts
// audits clean. Every row also carries its per-command record entry: the
// exact `script-worker-v1` identity or the explicit `(none)` absence.
func TestCLIAuditEmitsScriptLabels(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		schema     int
		enforced   bool
		network    []string
		wantLabel  string
		wantPolicy string
	}{
		{"schema7-script", 7, false, nil, "script-command-declared-only", "(none)"},
		{"schema8-declared-only-script", 8, false, nil, "script-command-declared-only", "(none)"},
		{"schema8-enforced-script", 8, true, nil, "", "script-worker-v1"},
		{"schema8-enforced-unfiltered-network", 8, true, []string{"audit-label.example.com"},
			"script-command-unfiltered-declared-network", "script-worker-v1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			configPath := auditLabelProject(t, testCase.schema, testCase.enforced, testCase.network)
			code, stdout, stderr := capture(t, configPath, "audit", "app")
			if code != exitOK {
				t.Fatalf("audit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
			}
			if stderr != "" {
				t.Fatalf("audit wrote errors for a labelled skill:\n%s", stderr)
			}
			wantRecord := "app: audit info: label-skill: command 'tool' execution_policy=" + testCase.wantPolicy
			if !strings.Contains(stdout, wantRecord) {
				t.Fatalf("audit output does not carry the record entry %q:\n%s", wantRecord, stdout)
			}
			if testCase.wantLabel == "" {
				want := wantRecord + "\napp: audit clean\n"
				if stdout != want {
					t.Fatalf("an enforced skill with no declared hosts audited:\n%s\nwant exactly:\n%s", stdout, want)
				}
				for _, label := range []string{"script-command-declared-only", "script-command-unfiltered-declared-network"} {
					if strings.Contains(stdout, label) {
						t.Fatalf("a warning-free audit rendered %q:\n%s", label, stdout)
					}
				}
				return
			}
			if !strings.Contains(stdout, testCase.wantLabel) {
				t.Fatalf("audit output does not carry %q:\n%s", testCase.wantLabel, stdout)
			}
			if !strings.Contains(stdout, "audit warning:") {
				t.Fatalf("the label is not an audit warning:\n%s", stdout)
			}
			if testCase.enforced && strings.Contains(stdout, "script-command-declared-only") {
				t.Fatalf("an enforced audit rendered declared-only:\n%s", stdout)
			}
		})
	}
}

// TestCLIAuditMixedSkillRecordsBothPolicies drives a two-command skill —
// one declared-only and one enforced — through `curator audit`. Both
// record entries render with the exact identity or the explicit absence,
// while only the declared-only command warns.
func TestCLIAuditMixedSkillRecordsBothPolicies(t *testing.T) {
	t.Parallel()
	configPath := auditLabelMixedProject(t)
	code, stdout, stderr := capture(t, configPath, "audit", "app")
	if code != exitOK {
		t.Fatalf("audit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("audit wrote errors:\n%s", stderr)
	}
	for _, want := range []string{
		"app: audit info: label-skill: command 'guarded' execution_policy=script-worker-v1",
		"app: audit info: label-skill: command 'tool' execution_policy=(none)",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("audit output does not carry the record entry %q:\n%s", want, stdout)
		}
	}
	if !strings.Contains(stdout, "script-command-declared-only") {
		t.Fatalf("the declared-only command did not warn:\n%s", stdout)
	}
	if strings.Contains(stdout, "script-command-unfiltered-declared-network") {
		t.Fatalf("the enforced no-host command earned the network label:\n%s", stdout)
	}
	if strings.Count(stdout, "audit warning:") != 1 {
		t.Fatalf("a mixed skill must warn exactly once:\n%s", stdout)
	}
}

// TestCLIAuditJSONCarriesScriptLabels proves the machine-readable audit
// output renders the classes as warnings, never errors.
func TestCLIAuditJSONCarriesScriptLabels(t *testing.T) {
	t.Parallel()
	configPath := auditLabelProject(t, 8, true, []string{"audit-label.example.com"})
	code, stdout, stderr := capture(t, configPath, "audit", "app", "--json")
	if code != exitOK {
		t.Fatalf("audit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("audit wrote errors:\n%s", stderr)
	}
	var outputs []struct {
		Scope    string   `json:"scope"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
		Policies []struct {
			Skill   string `json:"skill"`
			Command string `json:"command"`
			Policy  string `json:"policy"`
		} `json:"script_policies"`
	}
	if err := json.Unmarshal([]byte(stdout), &outputs); err != nil {
		t.Fatalf("audit --json is not machine-readable: %v\n%s", err, stdout)
	}
	if len(outputs) != 1 || outputs[0].Scope != "app" {
		t.Fatalf("audit --json scopes = %+v", outputs)
	}
	if len(outputs[0].Errors) != 0 {
		t.Fatalf("audit --json errors = %v", outputs[0].Errors)
	}
	if !strings.Contains(strings.Join(outputs[0].Warnings, "\n"), "script-command-unfiltered-declared-network") {
		t.Fatalf("audit --json warnings = %v", outputs[0].Warnings)
	}
	if len(outputs[0].Policies) != 1 {
		t.Fatalf("audit --json script_policies = %+v, want the enforced entry", outputs[0].Policies)
	}
	entry := outputs[0].Policies[0]
	if entry.Skill != "label-skill" || entry.Command != "tool" || entry.Policy != "script-worker-v1" {
		t.Fatalf("audit --json script_policies = %+v, want the exact enforced identity", outputs[0].Policies)
	}
}

// TestCLIAuditJSONRecordsExplicitAbsence proves the machine-readable
// audit output carries the explicit policy absence for a declared-only
// command alongside its warning.
func TestCLIAuditJSONRecordsExplicitAbsence(t *testing.T) {
	t.Parallel()
	configPath := auditLabelProject(t, 8, false, nil)
	code, stdout, stderr := capture(t, configPath, "audit", "app", "--json")
	if code != exitOK {
		t.Fatalf("audit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("audit wrote errors:\n%s", stderr)
	}
	var outputs []struct {
		Scope    string   `json:"scope"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
		Policies []struct {
			Skill   string `json:"skill"`
			Command string `json:"command"`
			Policy  string `json:"policy"`
		} `json:"script_policies"`
	}
	if err := json.Unmarshal([]byte(stdout), &outputs); err != nil {
		t.Fatalf("audit --json is not machine-readable: %v\n%s", err, stdout)
	}
	if len(outputs) != 1 {
		t.Fatalf("audit --json scopes = %+v", outputs)
	}
	if !strings.Contains(strings.Join(outputs[0].Warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("audit --json warnings = %v", outputs[0].Warnings)
	}
	if len(outputs[0].Policies) != 1 {
		t.Fatalf("audit --json script_policies = %+v, want the declared-only entry", outputs[0].Policies)
	}
	entry := outputs[0].Policies[0]
	if entry.Skill != "label-skill" || entry.Command != "tool" || entry.Policy != "" {
		t.Fatalf("audit --json script_policies = %+v, want the explicit absence", outputs[0].Policies)
	}
}
