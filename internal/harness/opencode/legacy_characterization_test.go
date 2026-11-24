package opencode_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestLegacyOpenCodeSkillDiscoveryCharacterization(t *testing.T) {
	if !adapters.NativeDiscoveryAgents["opencode"] {
		t.Fatal("OpenCode must use native skill discovery")
	}
	if _, mirrored := adapters.AgentPaths["opencode"]; mirrored {
		t.Fatal("OpenCode must not receive a project-level skill mirror")
	}
	if got, want := adapters.RequiredGitignoreEntries([]string{"opencode"}), []string{".agents/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("required ignores = %v, want %v", got, want)
	}

	project := t.TempDir()
	projectSkills := filepath.Join(project, ".agents", "skills")
	writeOpenCodeFixture(t, filepath.Join(projectSkills, "skill-a", "README.md"), "project skill")
	projectMirror, err := adapters.StageProject(t.TempDir(), project, []string{"opencode"},
		[]adapters.Group{{Root: projectSkills, Skills: []string{"skill-a"}}}, "copy")
	if err != nil {
		t.Fatal(err)
	}
	if got := projectMirror.Plan().Targets; len(got) != 0 {
		t.Fatalf("project-level native-discovery targets = %+v, want none", got)
	}

	curatorHome, userHome := t.TempDir(), t.TempDir()
	globalSkills := filepath.Join(curatorHome, "global", "skills")
	writeOpenCodeFixture(t, filepath.Join(globalSkills, "skill-a", "README.md"), "global skill")
	globalMirror, err := adapters.StageGlobal(t.TempDir(), curatorHome, userHome,
		[]string{"opencode"}, []string{"skill-a"}, "copy", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantTarget := filepath.Join(userHome, filepath.FromSlash(adapters.NativeDiscoveryHomePath), "skill-a")
	found := false
	for _, target := range globalMirror.Plan().Targets {
		if target.LivePath == wantTarget {
			found = true
		}
	}
	if !found {
		t.Fatalf("global native-discovery mirror has no target %q: %+v", wantTarget, globalMirror.Plan().Targets)
	}
}

func TestLegacyOpenCodeMCPInspectionCharacterization(t *testing.T) {
	env := mcp.Env{ProjectRoot: t.TempDir(), UserHome: t.TempDir()}
	writeOpenCodeFixture(t, filepath.Join(env.ProjectRoot, "opencode.jsonc"), `{
		// OpenCode's project MCP source is JSONC.
		"mcp": {
			"project-only": {"url": "https://example.invalid/mcp"},
			"shared": {"url": "https://shared.invalid/mcp"},
			"disabled": {"enabled": false, "url": "https://disabled.invalid/mcp"}
		}
	}`)
	writeOpenCodeFixture(t, filepath.Join(env.UserHome, ".config", "opencode", "opencode.json"), `{
		"mcp": {
			"home-only": {"url": "https://home.invalid/mcp"},
			"shared": {"url": "https://home-shared.invalid/mcp"}
		}
	}`)

	servers, err := mcp.ConfiguredServers(env, "opencode")
	if err != nil {
		t.Fatalf("read configured OpenCode MCP servers: %v", err)
	}
	for _, name := range []string{"project-only", "home-only", "shared"} {
		if _, found := servers[name]; !found {
			t.Errorf("configured OpenCode MCP server %q missing from %v", name, servers)
		}
	}
	if _, found := servers["disabled"]; found {
		t.Fatalf("disabled OpenCode MCP server counted as configured: %v", servers)
	}
	projectOnly, err := mcp.ProjectOnlyServers(env, "opencode")
	if err != nil {
		t.Fatalf("read project-only OpenCode MCP servers: %v", err)
	}
	if got, want := projectOnly, map[string]bool{"project-only": true}; !reflect.DeepEqual(got, want) {
		t.Fatalf("project-only OpenCode MCP servers = %v, want %v", got, want)
	}

	requirements := map[string]map[string]skillspec.McpServer{
		"skill-a": {"project-only": {Name: "project-only", Hint: "configure the project server", RequiredIn: "all"}},
	}
	findings, warnings, err := mcp.Verify(env, []string{"opencode"}, requirements)
	if err != nil {
		t.Fatalf("configured project MCP requirement refused: %v", err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "pending until the checkout is trusted") {
		t.Fatalf("project-only MCP warning = %v, want the existing checkout-trust warning", warnings)
	}
	if got, want := findings["skill-a"].FoundIn["project-only"], []string{"opencode"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("MCP inspection result = %v, want %v", got, want)
	}

	missing := map[string]map[string]skillspec.McpServer{
		"skill-a": {"missing": {Name: "missing", Hint: "configure the server", RequiredIn: "all"}},
	}
	if _, _, err := mcp.Verify(env, []string{"opencode"}, missing); err == nil {
		t.Fatal("missing required MCP server was accepted")
	}
}

func TestLegacyOpenCodeRefusalCharacterization(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.OpenCode)
	if err != nil {
		t.Fatal(err)
	}
	if adapter.EnvVar != "XDG_CONFIG_HOME" || !adapter.ParentVar || adapter.RootTarget != "AGENTS.md" {
		t.Fatalf("OpenCode home contract = %+v", adapter)
	}
	if adapter.MCP == nil || adapter.MCP.Kind != envregistry.KindVariable || adapter.MCP.Variable != "OPENCODE_CONFIG" {
		t.Fatalf("OpenCode MCP channel = %+v", adapter.MCP)
	}

	if _, err := adapter.ResolveForm("unsupported"); err == nil || !strings.Contains(err.Error(), envregistry.DiagFormUnsupported) {
		t.Fatalf("unsupported form refusal = %v, want %s", err, envregistry.DiagFormUnsupported)
	}
	if _, err := adapter.ResolveIsolation(envregistry.IsolationIsolated, true); err == nil || !strings.Contains(err.Error(), envregistry.DiagIsolatedUnsupported) {
		t.Fatalf("isolated credential route refusal = %v, want %s", err, envregistry.DiagIsolatedUnsupported)
	}
}

func writeOpenCodeFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
