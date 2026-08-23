package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/skillspec"
)

// schema8ScriptSkill creates a tagged schema-8 skill whose single script
// command is enforced or declared-only. It is the install-side twin of the
// suite's valid-script-worker-enforced schema case.
func (e *env) schema8ScriptSkill(name string, enforced bool) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	command := map[string]any{
		"type":      "script",
		"unix_path": "scripts/" + name + "-tool",
		"win_path":  "scripts/" + name + "-tool",
	}
	if enforced {
		command["execution_policy"] = skillspec.ScriptExecutionPolicy
		command["interpreter"] = "python3-v1"
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": "none", "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands":      map[string]any{name + "-tool": command},
	}, "", "  ")
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// An enforced command must not reach a shim. Before the policy existed the
// manifest was rejected outright by the supported-schema gate; admitting
// schema 8 removed that accident, so the refusal has to be deliberate.
func TestEnforcedScriptCommandIsRefusedAtInstall(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.schema8ScriptSkill("enforced-skill", true)
	e.declare("enforced-skill")

	result := e.install(Options{})
	if result.Status == "ok" {
		t.Fatalf("an enforced script command installed: %+v", result)
	}
	reported := strings.Join(append(result.Errors, result.Messages...), "\n")
	if !strings.Contains(reported, scriptpolicy.PolicyUnsupported) {
		t.Fatalf("install did not report %s: %s", scriptpolicy.PolicyUnsupported, reported)
	}
	// The forbidden outcome is the shim, not just the exit status: a partial
	// install that still published the launcher would run the package code
	// uncontained even though the run reported a failure.
	shim := filepath.Join(e.project, ".agents", "bin", shimName("enforced-skill-tool"))
	if _, err := os.Lstat(shim); err == nil {
		t.Fatal("a shim was published for an enforced script command")
	}
}

// The control: schema 8 by itself changes nothing for a command that did not
// opt in, so a declared-only script command installs exactly as before.
func TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.schema8ScriptSkill("plain-skill", false)
	e.declare("plain-skill")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a declared-only schema-8 skill failed to install: %+v", result)
	}
	shim := filepath.Join(e.project, ".agents", "bin", shimName("plain-skill-tool"))
	if _, err := os.Lstat(shim); err != nil {
		t.Fatalf("shim missing for a declared-only schema-8 command: %v", err)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "plain-skill", "SKILL.md")); err != nil {
		t.Fatalf("context missing: %v", err)
	}
}

// The shim writer is the last layer that could turn an enforced command into
// an uncontained launcher, so it refuses on its own rather than trusting the
// preflight that normally runs before it.
func TestActiveScriptCommandsRefusesEnforcedCommand(t *testing.T) {
	t.Parallel()
	node := &closure.Node{
		Name: "skill", Spec: &skillspec.Spec{Commands: map[string]skillspec.Command{
			"plain": {Name: "plain", Type: "script", UnixPath: "scripts/plain"},
			"enforced": {
				Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
				ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
			},
		}},
		Edges: []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
	}
	_, err := activeScriptCommands(node, node.ActiveCommands())
	if err == nil {
		t.Fatal("the shim writer accepted an enforced script command")
	}
	if scriptpolicy.Code(err) != scriptpolicy.PolicyUnsupported {
		t.Fatalf("Code = %q, want %q", scriptpolicy.Code(err), scriptpolicy.PolicyUnsupported)
	}
	if !strings.Contains(err.Error(), "skill.enforced") {
		t.Fatalf("refusal does not name the command: %v", err)
	}

	// An enforced command that no edge activates is not staged at all, so the
	// remaining declared-only command still reaches the writer normally.
	node.Edges = []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "runtime", Commands: []string{"plain"}}}
	commands, err := activeScriptCommands(node, node.ActiveCommands())
	if err != nil {
		t.Fatalf("declared-only command refused: %v", err)
	}
	if len(commands) != 1 || commands[0].Name != "plain" {
		t.Fatalf("commands = %+v", commands)
	}
}
