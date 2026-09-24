package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/scriptworker"
)

// schema7ScriptSkill creates a tagged schema-7 skill whose single script
// command is declared-only by construction: schema 7 cannot declare an
// execution policy. networkHosts selects the declared `network` value: nil
// declares none, otherwise the given host list.
func (e *env) schema7ScriptSkill(name string, networkHosts []string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	network := any("none")
	if networkHosts != nil {
		network = networkHosts
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 7,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": network, "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands": map[string]any{name + "-tool": map[string]any{
			"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool",
		}},
	}, "", "  ")
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// schema8DeclaredOnlySkillWithNetwork creates a tagged schema-8 skill whose
// single script command omits the execution policy. networkHosts selects
// the declared `network` value like schema7ScriptSkill.
func (e *env) schema8DeclaredOnlySkillWithNetwork(name string, networkHosts []string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	network := any("none")
	if networkHosts != nil {
		network = networkHosts
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": network, "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands": map[string]any{name + "-tool": map[string]any{
			"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool",
		}},
	}, "", "  ")
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// TestScriptAuditLabelsAtInstallEntry drives the four audit-label shapes
// through the production install: the validation messages (skill check)
// and the install-time audit gate both render the warning classes as
// warnings, never as errors, and every shape still installs — declared-only
// skills keep their ordinary shim, enforced skills their native launcher.
func TestScriptAuditLabelsAtInstallEntry(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name         string
		skill        string
		setup        func(e *env)
		wantLabels   []string
		wantLauncher bool
	}{
		{name: "schema7-script", skill: "auditlabel-schema7",
			setup:      func(e *env) { e.schema7ScriptSkill("auditlabel-schema7", nil) },
			wantLabels: []string{"script-command-declared-only"}},
		{name: "schema8-declared-only-script", skill: "auditlabel-declared",
			setup:      func(e *env) { e.schema8DeclaredOnlySkillWithNetwork("auditlabel-declared", nil) },
			wantLabels: []string{"script-command-declared-only"}},
		{name: "schema8-enforced-script", skill: "auditlabel-enforced",
			setup:        func(e *env) { e.schema8ScriptSkill("auditlabel-enforced", true) },
			wantLauncher: true},
		{name: "schema8-enforced-unfiltered-network", skill: "auditlabel-unfiltered",
			setup: func(e *env) {
				e.schema8EnforcedSkillWithCaps("auditlabel-unfiltered", map[string]any{
					"network": []string{"audit-label.example.com"},
				})
			},
			wantLabels:   []string{"script-command-unfiltered-declared-network"},
			wantLauncher: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			testCase.setup(e)
			e.declare(testCase.skill)
			e.cfg.Audit.Enabled = true
			e.cfg.Audit.Mode = "advisory"

			result := e.install(Options{})
			if result.Status != "ok" {
				t.Fatalf("a labelled skill failed to install: %+v", result)
			}
			messages := strings.Join(result.Messages, "\n")
			errors := strings.Join(result.Errors, "\n")
			for _, label := range testCase.wantLabels {
				if !strings.Contains(messages, label) {
					t.Fatalf("install messages do not carry %q: %s", label, messages)
				}
				// Both install-time surfaces render the class: the
				// validation report and the audit gate.
				if !strings.Contains(messages, "warning: "+label) {
					t.Fatalf("no validation warning carries %q: %s", label, messages)
				}
				if !strings.Contains(messages, "audit warning:") || !strings.Contains(messages, label) {
					t.Fatalf("no audit-gate warning carries %q: %s", label, messages)
				}
				if strings.Contains(errors, label) {
					t.Fatalf("the label escaped into install errors: %s", errors)
				}
			}
			if len(testCase.wantLabels) == 0 {
				for _, label := range []string{"script-command-declared-only", "script-command-unfiltered-declared-network"} {
					if strings.Contains(messages, label) {
						t.Fatalf("an enforced script with no declared hosts was labelled %q: %s", label, messages)
					}
				}
			}
			// The negative rows: enforced commands never render
			// declared-only, and the unfiltered row renders nothing else.
			if testCase.wantLauncher {
				if strings.Contains(messages, "script-command-declared-only") {
					t.Fatalf("an enforced install rendered declared-only: %s", messages)
				}
			}
			if testCase.name == "schema8-enforced-unfiltered-network" &&
				strings.Count(messages, "script-command-unfiltered-declared-network") == 0 {
				t.Fatalf("the unfiltered row rendered no network label: %s", messages)
			}
			launcher := testCase.skill + "-tool"
			if testCase.wantLauncher {
				launcher = nativeLauncherName(launcher)
			} else {
				launcher = shimName(launcher)
			}
			if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", launcher)); err != nil {
				t.Fatalf("no launcher was published: %v", err)
			}
			if testCase.wantLauncher {
				if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", launcher) + scriptworker.ShimSidecarSuffix); err != nil {
					t.Fatalf("the enforced shape published no sidecar: %v", err)
				}
			} else if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", launcher) + scriptworker.ShimSidecarSuffix); err == nil {
				t.Fatal("a declared-only shape published a native launcher sidecar")
			}
		})
	}
}

// linesWith selects the message lines containing every given part.
func linesWith(messages []string, parts ...string) []string {
	var selected []string
outer:
	for _, message := range messages {
		for _, part := range parts {
			if !strings.Contains(message, part) {
				continue outer
			}
		}
		selected = append(selected, message)
	}
	return selected
}

// TestScriptAuditLabelsAtGlobalInstallEntry drives the four audit-label
// shapes through the machine-wide install with its DEFAULT production
// audit gate (no AuditGate override), so the global.go Subject wiring is
// exercised exactly as shipped. Each label renders on its own audit-gate
// line carrying the exact vector text, distinct from the skillcheck
// validation line, and every shape still installs its legacy launcher.
func TestScriptAuditLabelsAtGlobalInstallEntry(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name         string
		skill        string
		setup        func(e *env)
		wantLabels   []string
		wantLauncher bool
	}{
		{name: "schema7-script", skill: "gauditlabel-schema7",
			setup:      func(e *env) { e.schema7ScriptSkill("gauditlabel-schema7", nil) },
			wantLabels: []string{"script-command-declared-only"}},
		{name: "schema8-declared-only-script", skill: "gauditlabel-declared",
			setup:      func(e *env) { e.schema8DeclaredOnlySkillWithNetwork("gauditlabel-declared", nil) },
			wantLabels: []string{"script-command-declared-only"}},
		{name: "schema8-enforced-script", skill: "gauditlabel-enforced",
			setup:        func(e *env) { e.schema8ScriptSkill("gauditlabel-enforced", true) },
			wantLauncher: true},
		{name: "schema8-enforced-unfiltered-network", skill: "gauditlabel-unfiltered",
			setup: func(e *env) {
				e.schema8EnforcedSkillWithCaps("gauditlabel-unfiltered", map[string]any{
					"network": []string{"audit-label.example.com"},
				})
			},
			wantLabels:   []string{"script-command-unfiltered-declared-network"},
			wantLauncher: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			testCase.setup(e)
			if _, err := GlobalInit(e.home); err != nil {
				t.Fatal(err)
			}
			if err := manifestAddGlobal(e, testCase.skill); err != nil {
				t.Fatal(err)
			}
			e.cfg.Audit.Enabled = true
			e.cfg.Audit.Mode = "advisory"

			result := Global(e.cfg, t.TempDir(), Options{Platform: installPlatform()})
			if result.Status != "ok" {
				t.Fatalf("a labelled skill failed the global install: %+v", result)
			}
			errors := strings.Join(result.Errors, "\n")
			for _, label := range testCase.wantLabels {
				// The audit-gate warning carries the exact label on
				// the same message line. A global wiring that drops
				// the parsed commands still validates (the
				// skillcheck line below survives) but renders no
				// such line, so this assertion kills it.
				auditLines := linesWith(result.Messages, "audit warning:", label)
				if len(auditLines) == 0 {
					t.Fatalf("no global audit-gate warning carries %q: %s",
						label, strings.Join(result.Messages, "\n"))
				}
				checkLines := linesWith(result.Messages, "warning: "+label)
				if len(checkLines) == 0 {
					t.Fatalf("no global validation warning carries %q: %s",
						label, strings.Join(result.Messages, "\n"))
				}
				for _, auditLine := range auditLines {
					for _, checkLine := range checkLines {
						if auditLine == checkLine {
							t.Fatalf("the audit and validation warnings share a line: %q", auditLine)
						}
					}
				}
				if strings.Contains(errors, label) {
					t.Fatalf("the label escaped into global install errors: %s", errors)
				}
			}
			if len(testCase.wantLabels) == 0 {
				messages := strings.Join(result.Messages, "\n")
				for _, label := range []string{"script-command-declared-only", "script-command-unfiltered-declared-network"} {
					if strings.Contains(messages, label) {
						t.Fatalf("an enforced script with no declared hosts was labelled %q: %s", label, messages)
					}
				}
			}
			if testCase.wantLauncher {
				if messages := strings.Join(result.Messages, "\n"); strings.Contains(messages, "script-command-declared-only") {
					t.Fatalf("an enforced global install rendered declared-only: %s", messages)
				}
			}
			launcher := testCase.skill + "-tool"
			if testCase.wantLauncher {
				launcher = nativeLauncherName(launcher)
			} else {
				launcher = shimName(launcher)
			}
			globalBin := filepath.Join(e.home, "global", "bin", launcher)
			if _, err := os.Lstat(globalBin); err != nil {
				t.Fatalf("no global launcher was published: %v", err)
			}
			if testCase.wantLauncher {
				if _, err := os.Lstat(globalBin + scriptworker.ShimSidecarSuffix); err != nil {
					t.Fatalf("the enforced shape published no global sidecar: %v", err)
				}
			} else if _, err := os.Lstat(globalBin + scriptworker.ShimSidecarSuffix); err == nil {
				t.Fatal("a declared-only shape published a global native launcher sidecar")
			}
		})
	}
}

// TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly is the
// global negative row for the network class: a declared-only command
// that declares network hosts installs unchanged and renders only the
// declared-only warning, on both the audit-gate and validation lines.
func TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.schema8DeclaredOnlySkillWithNetwork("gauditlabel-hosts", []string{"audit-label.example.com"})
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "gauditlabel-hosts"); err != nil {
		t.Fatal(err)
	}
	e.cfg.Audit.Enabled = true
	e.cfg.Audit.Mode = "advisory"

	result := Global(e.cfg, t.TempDir(), Options{Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("a declared-only skill failed the global install: %+v", result)
	}
	messages := strings.Join(result.Messages, "\n")
	if len(linesWith(result.Messages, "audit warning:", "script-command-declared-only")) == 0 {
		t.Fatalf("no global audit-gate warning carries declared-only: %s", messages)
	}
	if len(linesWith(result.Messages, "warning: script-command-declared-only")) == 0 {
		t.Fatalf("no global validation warning carries declared-only: %s", messages)
	}
	if strings.Contains(messages, "script-command-unfiltered-declared-network") {
		t.Fatalf("a declared-only global install earned the network label: %s", messages)
	}
	if _, err := os.Lstat(filepath.Join(e.home, "global", "bin", shimName("gauditlabel-hosts-tool"))); err != nil {
		t.Fatalf("global shim missing for a declared-only command: %v", err)
	}
}

// TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly is the install
// negative row for the network class: a declared-only command that
// declares network hosts installs unchanged and renders only the
// declared-only warning.
func TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.schema8DeclaredOnlySkillWithNetwork("auditlabel-hosts", []string{"audit-label.example.com"})
	e.declare("auditlabel-hosts")
	e.cfg.Audit.Enabled = true
	e.cfg.Audit.Mode = "advisory"

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a declared-only skill failed to install: %+v", result)
	}
	messages := strings.Join(result.Messages, "\n")
	if !strings.Contains(messages, "script-command-declared-only") {
		t.Fatalf("install messages do not carry declared-only: %s", messages)
	}
	if strings.Contains(messages, "script-command-unfiltered-declared-network") {
		t.Fatalf("a declared-only install earned the network label: %s", messages)
	}
	if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", shimName("auditlabel-hosts-tool"))); err != nil {
		t.Fatalf("shim missing for a declared-only command: %v", err)
	}
}
