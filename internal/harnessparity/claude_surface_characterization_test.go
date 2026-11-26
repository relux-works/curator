package harnessparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/mcpnative"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestClaudeSharedSurfaceCharacterization(t *testing.T) {
	t.Run("project skill path", func(t *testing.T) {
		if !adapters.KnownAgents()["claude_code"] {
			t.Fatal("registered harness was not recognized")
		}
		got := adapters.RequiredGitignoreEntries([]string{"claude_code"})
		want := []string{".agents/", ".claude/skills/"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ignore entries = %q, want %q", got, want)
		}
	})

	t.Run("MCP sources, inspection, and refusals", func(t *testing.T) {
		project, home := t.TempDir(), t.TempDir()
		writeClaudeFixture(t, project, ".mcp.json", `{"mcpServers":{"disabled":{"url":"https://project"},"project-only":{"url":"https://project"}}}`)
		writeClaudeFixture(t, project, ".claude/settings.json", `{"disabledMcpjsonServers":["disabled"]}`)
		writeClaudeFixture(t, home, ".claude.json", `{"mcpServers":{"home-only":{"url":"https://home"}}}`)
		env := mcp.Env{ProjectRoot: project, UserHome: home}

		configured, err := mcp.ConfiguredServers(env, "claude_code")
		if err != nil {
			t.Fatal(err)
		}
		if len(configured) != 2 || configured["project-only"] == nil || configured["home-only"] == nil {
			t.Fatalf("configured servers = %v, want project-only and home-only", configured)
		}
		if _, present := configured["disabled"]; present {
			t.Fatal("a disabled project server counted as configured")
		}
		projectOnly, err := mcp.ProjectOnlyServers(env, "claude_code")
		if err != nil {
			t.Fatal(err)
		}
		if !projectOnly["project-only"] || projectOnly["home-only"] || projectOnly["disabled"] {
			t.Fatalf("project-only servers = %v", projectOnly)
		}

		serverReq := func(name string) map[string]map[string]skillspec.McpServer {
			return map[string]map[string]skillspec.McpServer{
				"fixture": {name: {Name: name, Hint: "configure " + name, RequiredIn: "any"}},
			}
		}
		_, warnings, err := mcp.Verify(env, []string{"claude_code"}, serverReq("project-only"))
		if err != nil || len(warnings) != 1 || !strings.Contains(warnings[0], "project-level config") {
			t.Fatalf("project-only verification = warnings %q, error %v", warnings, err)
		}
		if _, _, err := mcp.Verify(env, []string{"claude_code"}, serverReq("disabled")); err == nil || !strings.Contains(err.Error(), "not configured") {
			t.Fatalf("disabled server refusal = %v", err)
		}

		if got := mcpnative.Path("claude_code"); got != contextmaterialize.MCPClaudePath {
			t.Fatalf("native path = %q, want %q", got, contextmaterialize.MCPClaudePath)
		}
		document := []byte(`{"setting":[1,2],"mcpServers":{"foreign":{"command":"fixture"}}}`)
		entries, err := mcpnative.Entries("claude_code", document)
		if err != nil {
			t.Fatal(err)
		}
		var entry map[string]any
		if err := json.Unmarshal(entries["foreign"], &entry); err != nil || entry["command"] != "fixture" {
			t.Fatalf("inspected entry = %s, error %v", entries["foreign"], err)
		}
		if _, _, err := mcpnative.Add("claude_code", document, contextmaterialize.MCPServer{Name: "foreign", Transport: "stdio", Command: "other"}); err == nil || !strings.Contains(err.Error(), mcpnative.Conflict) {
			t.Fatalf("existing entry was not refused: %v", err)
		}
		if _, err := mcpnative.Entries("claude_code", []byte(`{"mcpServers":[]}`)); err == nil || !strings.Contains(err.Error(), mcpnative.Unreadable) {
			t.Fatalf("malformed entry container was not refused: %v", err)
		}
	})
}

func writeClaudeFixture(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
