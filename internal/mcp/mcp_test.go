package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newEnv(t *testing.T) Env {
	return Env{ProjectRoot: t.TempDir(), UserHome: t.TempDir()}
}

func req(name, requiredIn string) map[string]map[string]skillspec.McpServer {
	return map[string]map[string]skillspec.McpServer{
		"skill-x": {name: {Name: name, Hint: "connect " + name, RequiredIn: requiredIn}},
	}
}

func TestSurfacesPerAgent(t *testing.T) {
	env := newEnv(t)
	write(t, env.ProjectRoot, ".mcp.json", `{"mcpServers": {"alpha": {"command": "alpha-bin"}}}`)
	write(t, env.UserHome, ".claude.json", `{"mcpServers": {"beta": {"url": "https://x"}}}`)
	write(t, env.UserHome, ".codex/config.toml", "[mcp_servers.gamma]\ncommand = \"gamma-bin\"\n")
	write(t, env.ProjectRoot, "opencode.jsonc", `{
		// comment
		"mcp": {
			"delta": {"command": ["delta-bin", "--serve"]},
			"off": {"enabled": false, "command": ["off-bin"]}
		}
	}`)
	write(t, env.UserHome, ".codeium/windsurf/mcp_config.json", `{"mcpServers": {"epsilon": {}}}`)

	cases := []struct {
		agent string
		want  []string
	}{
		{"claude_code", []string{"alpha", "beta"}},
		{"codex_cli", []string{"gamma"}},
		{"opencode", []string{"delta"}}, // "off" excluded via enabled:false
		{"windsurf", []string{"epsilon"}},
		{"cursor", nil},
	}
	for _, tc := range cases {
		configured := ConfiguredServers(env, tc.agent)
		for _, name := range tc.want {
			if _, present := configured[name]; !present {
				t.Errorf("%s: missing %s (got %v)", tc.agent, name, configured)
			}
		}
		if tc.agent == "opencode" {
			if _, present := configured["off"]; present {
				t.Error("opencode enabled:false entry counted as configured")
			}
		}
	}
}

func TestClaudeDisabledServersDoNotCount(t *testing.T) {
	env := newEnv(t)
	write(t, env.ProjectRoot, ".mcp.json", `{"mcpServers": {"alpha": {}, "beta": {}}}`)
	write(t, env.ProjectRoot, ".claude/settings.json", `{"disabledMcpjsonServers": ["alpha"]}`)
	configured := ConfiguredServers(env, "claude_code")
	if _, present := configured["alpha"]; present {
		t.Fatal("disabled server counted as configured")
	}
	if _, present := configured["beta"]; !present {
		t.Fatal("enabled server missing")
	}
}

func TestVerifyAnySemantics(t *testing.T) {
	env := newEnv(t)
	write(t, env.UserHome, ".claude.json", `{"mcpServers": {"alpha": {"url": "https://remote"}}}`)
	agents := []string{"claude_code", "cursor"}

	findings, warnings, err := Verify(env, agents, req("alpha", "any"))
	if err != nil {
		t.Fatalf("any semantics must pass with one agent configured: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	found := findings["skill-x"].FoundIn["alpha"]
	if len(found) != 1 || found[0] != "claude_code" {
		t.Fatalf("found in: %v", found)
	}

	_, _, err = Verify(env, agents, req("ghost", "any"))
	if err == nil || !strings.Contains(err.Error(), "not configured in any target agent") || !strings.Contains(err.Error(), "connect ghost") {
		t.Fatalf("err = %v, want any-failure with hint", err)
	}
}

func TestVerifyAllSemantics(t *testing.T) {
	env := newEnv(t)
	write(t, env.UserHome, ".claude.json", `{"mcpServers": {"alpha": {"url": "https://remote"}}}`)
	agents := []string{"claude_code", "cursor"}
	_, _, err := Verify(env, agents, req("alpha", "all"))
	if err == nil || !strings.Contains(err.Error(), "not configured for agent(s): cursor") {
		t.Fatalf("err = %v, want all-failure naming cursor", err)
	}
	// unverifiable agents count as missing under all
	write(t, env.UserHome, ".cursor/mcp.json", `{"mcpServers": {"alpha": {}}}`)
	if _, _, err := Verify(env, agents, req("alpha", "all")); err != nil {
		t.Fatalf("all satisfied must pass: %v", err)
	}
}

func TestStaticStdioWarning(t *testing.T) {
	env := newEnv(t)
	write(t, env.UserHome, ".claude.json", `{"mcpServers": {"alpha": {"command": "definitely-missing-binary-xyz"}}}`)
	_, warnings, err := Verify(env, []string{"claude_code"}, req("alpha", "any"))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "not on PATH") {
		t.Fatalf("warnings: %v", warnings)
	}

	// a remote-shaped entry produces no warning
	env2 := newEnv(t)
	write(t, env2.UserHome, ".claude.json", `{"mcpServers": {"alpha": {"url": "https://remote"}}}`)
	_, warnings, err = Verify(env2, []string{"claude_code"}, req("alpha", "any"))
	if err != nil || len(warnings) != 0 {
		t.Fatalf("remote entry warned: %v %v", warnings, err)
	}
}

func TestProjectOnlyPendingWarning(t *testing.T) {
	env := newEnv(t)
	write(t, env.ProjectRoot, ".mcp.json", `{"mcpServers": {"alpha": {"url": "https://remote"}}}`)
	_, warnings, err := Verify(env, []string{"claude_code"}, req("alpha", "any"))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "project-level config") || !strings.Contains(joined, "pending until the checkout is trusted") {
		t.Fatalf("warnings: %v", warnings)
	}
}

func TestMalformedConfigCountsAsEmpty(t *testing.T) {
	env := newEnv(t)
	write(t, env.ProjectRoot, ".mcp.json", `{broken`)
	if servers := ConfiguredServers(env, "claude_code"); len(servers) != 0 {
		t.Fatalf("malformed config produced servers: %v", servers)
	}
}
