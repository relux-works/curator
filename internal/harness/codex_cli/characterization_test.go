package codex_cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/harness/codex_cli"
	"github.com/relux-works/curator/internal/mcpnative"
)

func TestCodexSharedHomeSkillAndPolicyFacts(t *testing.T) {
	if codex_cli.ID != "codex_cli" {
		t.Fatalf("canonical ID = %q, want codex_cli", codex_cli.ID)
	}
	if got := adapters.AgentPaths[string(codex_cli.ID)]; got != ".codex/skills" {
		t.Fatalf("project skill path = %q, want .codex/skills", got)
	}
	if !adapters.KnownAgents()[string(codex_cli.ID)] {
		t.Fatal("Codex CLI is missing from the known skill-mirror agents")
	}
	ignored := adapters.RequiredGitignoreEntries([]string{string(codex_cli.ID)})
	if !containsString(ignored, ".codex/skills/") {
		t.Fatalf("required generated-path ignores = %v, want .codex/skills/", ignored)
	}

	var home envprofile.Adapter
	found := false
	for _, adapter := range envprofile.Adapters {
		if adapter.ID == string(codex_cli.ID) {
			home, found = adapter, true
			break
		}
	}
	if !found || home.EnvVar != "CODEX_HOME" || home.DefaultDir != ".codex" || home.Target != "AGENTS.md" {
		t.Fatalf("legacy home row = %+v, found=%t", home, found)
	}

	row, err := envregistry.ByID(string(codex_cli.ID))
	if err != nil {
		t.Fatal(err)
	}
	if row.RootTarget != "AGENTS.md" || row.SkillsDir != "skills" || row.DefaultForm != envregistry.FormMonolithic {
		t.Fatalf("legacy root/skills/instruction facts = %+v", row)
	}
	if len(row.SystemPrompt) != 1 || row.SystemPrompt[0].Kind != envregistry.KindConfigKey || row.SystemPrompt[0].Key != "model_instructions_file" || row.SystemPrompt[0].Semantics != envregistry.SemanticsReplace {
		t.Fatalf("Codex instruction channel = %+v", row.SystemPrompt)
	}
	if row.MCP == nil || row.MCP.Kind != envregistry.KindFlag || row.MCP.Flag != "-p" || row.MCP.Argument != envregistry.ArgumentName || row.MCP.Name != "curator-mcp" {
		t.Fatalf("Codex MCP channel = %+v", row.MCP)
	}
	credentials := row.Passthrough["default"]
	if len(credentials) != 1 || credentials[0].Path != "auth.json" || credentials[0].Strategy != envregistry.StrategyKeyringPreferred || credentials[0].FileLinkTarget != "auth.json" {
		t.Fatalf("Codex credential route = %+v", credentials)
	}
	if len(row.Seeds) != 1 || row.Seeds[0] != "config.toml" || row.CodexSeedRevision != envregistry.CodexSeedRevisionC {
		t.Fatalf("Codex seed facts = seeds:%v revision:%q", row.Seeds, row.CodexSeedRevision)
	}
}

func TestCodexMCPInspectionAndRefusalFacts(t *testing.T) {
	if got := mcpnative.Path(string(codex_cli.ID)); got != "curator-mcp.config.toml" {
		t.Fatalf("Codex MCP source path = %q", got)
	}
	native := []byte("# foreign settings stay byte-identical\n[mcp_servers.foreign]\ncommand = \"fixture\"\nargs = [\"--stdio\"]\n")
	entries, err := mcpnative.Entries(string(codex_cli.ID), native)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("[mcp_servers.foreign]\ncommand = \"fixture\"\nargs = [\"--stdio\"]")
	if !bytes.Equal(entries["foreign"], want) {
		t.Fatalf("inspected Codex MCP entry = %q, want %q", entries["foreign"], want)
	}
	if _, err := mcpnative.Entries(string(codex_cli.ID), []byte("[mcp_servers.broken\ncommand = \"fixture\"\n")); err == nil || !strings.Contains(err.Error(), mcpnative.Unreadable) {
		t.Fatalf("malformed Codex MCP inspection error = %v, want %s", err, mcpnative.Unreadable)
	}

	home := t.TempDir()
	before, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	result, err := envprofile.ProfileMCP(envprofile.ProfileMCPRequest{
		Home: home, Profile: "sample", Scope: "project", Mode: "status",
		Environments: []string{string(codex_cli.ID)},
	})
	if err == nil || result == nil || result.Code != "mcp_contribution_policy_refused" || result.Published {
		t.Fatalf("project-scope MCP refusal = result:%+v err:%v", result, err)
	}
	after, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("refused MCP inspection changed temporary home: before=%v after=%v", before, after)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
