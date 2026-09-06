package contextmaterialize

import (
	"strings"
	"testing"
)

func testServers() []MCPServer {
	return []MCPServer{
		{Name: "figma-devmode", Transport: MCPTransportStdio, Command: "npx", Args: []string{"-y", "figma-developer-mcp", "--stdio"}, EnvNames: []string{"FIGMA_API_KEY"}},
		{Name: "docs-remote", Transport: MCPTransportHTTP, URL: "https://mcp.example.com/docs", EnvNames: []string{"DOCS_TOKEN", "DOCS_ORG"}, Environments: []string{"claude_code", "opencode"}},
		{Name: "codex-only", Transport: MCPTransportStdio, Command: "uvx", Environments: []string{"codex_cli"}},
	}
}

func setNames(set []MCPServer) []string {
	var names []string
	for _, server := range set {
		names = append(names, server.Name)
	}
	return names
}

func TestMCPSetSelectors(t *testing.T) {
	set, err := MCPSet(testServers(), "claude_code")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(setNames(set), ","), "docs-remote,figma-devmode"; got != want {
		t.Fatalf("claude set %q, want %q", got, want)
	}
	set, err = MCPSet(testServers(), "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(setNames(set), ","), "codex-only,figma-devmode"; got != want {
		t.Fatalf("codex set %q, want %q", got, want)
	}
	// A nil selector admits every adapter.
	set, err = MCPSet(testServers(), "pi")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(setNames(set), ","), "figma-devmode"; got != want {
		t.Fatalf("pi set %q, want %q", got, want)
	}
}

// TestMCPEnvNamesUnion narrows the sorted-union gate: the union is sorted
// over every member's names, so admitting an unsorted emission or dropping
// one member's names fails this test.
func TestMCPEnvNamesUnion(t *testing.T) {
	set, err := MCPSet(testServers(), "claude_code")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(MCPEnvNames(set), ","), "DOCS_ORG,DOCS_TOKEN,FIGMA_API_KEY"; got != want {
		t.Fatalf("env_names %q, want %q", got, want)
	}
}

func TestMCPFilePiWritesNothing(t *testing.T) {
	path, document, written, err := MCPFile("pi", testServers())
	if err != nil {
		t.Fatal(err)
	}
	if written || path != "" || document != nil {
		t.Fatal("pi declares no MCP file and no fragment section")
	}
}

func TestMCPFileEmptySetWritesNothing(t *testing.T) {
	for _, env := range []string{"claude_code", "codex_cli", "opencode"} {
		_, _, written, err := MCPFile(env, nil)
		if err != nil {
			t.Fatal(err)
		}
		if written {
			t.Fatalf("an empty %s set writes no file", env)
		}
	}
}

func TestMCPFilePaths(t *testing.T) {
	for env, want := range map[string]string{
		"claude_code": MCPClaudePath,
		"codex_cli":   MCPCodexPath,
		"opencode":    MCPOpenCodePath,
	} {
		path, _, written, err := MCPFile(env, testServers())
		if err != nil {
			t.Fatal(err)
		}
		if !written || path != want {
			t.Fatalf("%s file path %q written %v, want %q", env, path, written, want)
		}
	}
}

// TestMCPRejectsBadTransport narrows the transport gate: weakening it to
// admit an unknown transport must fail this test rather than render an
// undefined server shape.
func TestMCPRejectsBadTransport(t *testing.T) {
	servers := []MCPServer{{Name: "weird", Transport: "websocket", Command: "x"}}
	if _, err := MCPSet(servers, "claude_code"); err == nil {
		t.Fatal("an unknown transport must be rejected")
	}
}

func TestMCPRejectsNamelessCommand(t *testing.T) {
	servers := []MCPServer{{Name: "nocmd", Transport: MCPTransportStdio}}
	if _, err := MCPSet(servers, "claude_code"); err == nil {
		t.Fatal("a stdio server without a command must be rejected")
	}
	servers = []MCPServer{{Name: "nourl", Transport: MCPTransportHTTP}}
	if _, err := MCPSet(servers, "claude_code"); err == nil {
		t.Fatal("an http server without a url must be rejected")
	}
}

func TestMCPRejectsBadName(t *testing.T) {
	servers := []MCPServer{{Name: "../escape", Transport: MCPTransportHTTP, URL: "https://example.com/x"}}
	if _, err := MCPSet(servers, "claude_code"); err == nil {
		t.Fatal("a non-identifier server name must be rejected before it reaches a bare TOML key")
	}
}

func TestCodexTOMLExactness(t *testing.T) {
	servers := []MCPServer{
		{Name: "b-server", Transport: MCPTransportStdio, Command: "uvx"},
		{Name: "a-server", Transport: MCPTransportHTTP, URL: "https://example.com/hook"},
	}
	path, document, written, err := MCPFile("codex_cli", servers)
	if err != nil {
		t.Fatal(err)
	}
	if !written || path != "curator-mcp.config.toml" {
		t.Fatalf("codex layer path %q written %v", path, written)
	}
	want := "[mcp_servers.a-server]\nurl = \"https://example.com/hook\"\n[mcp_servers.b-server]\ncommand = \"uvx\"\nargs = []\n"
	if string(document) != want {
		t.Fatalf("codex TOML differs:\n got %q\nwant %q", document, want)
	}
}

// TestCodexTOMLEscaping narrows the quoting gate: a value carrying quotes,
// backslashes, or control bytes must stay inside one TOML basic string.
func TestCodexTOMLEscaping(t *testing.T) {
	servers := []MCPServer{{Name: "q", Transport: MCPTransportStdio, Command: "run", Args: []string{"a\"b", "c\\d", "e\nf"}}}
	_, document, _, err := MCPFile("codex_cli", servers)
	if err != nil {
		t.Fatal(err)
	}
	want := "[mcp_servers.q]\ncommand = \"run\"\nargs = [\"a\\\"b\", \"c\\\\d\", \"e\\nf\"]\n"
	if string(document) != want {
		t.Fatalf("escaped TOML differs:\n got %q\nwant %q", document, want)
	}
}

func TestClaudeMCPJSONShape(t *testing.T) {
	_, document, written, err := MCPFile("claude_code", testServers()[:2])
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("a non-empty claude set writes a file")
	}
	text := string(document)
	if !strings.HasSuffix(text, "}\n") {
		t.Fatalf("claude MCP file is not CCJ-1 plus one LF: %q", text)
	}
	if !strings.Contains(text, `"command":"npx"`) || !strings.Contains(text, `"type":"stdio"`) {
		t.Fatalf("claude stdio shape is wrong: %q", text)
	}
	if !strings.Contains(text, `"type":"http"`) {
		t.Fatalf("claude http shape is wrong: %q", text)
	}
	if strings.Contains(text, "FIGMA_API_KEY") {
		t.Fatalf("no value byte may enter the MCP file: %q", text)
	}
}

func TestOpenCodeMCPJSONShape(t *testing.T) {
	_, document, written, err := MCPFile("opencode", testServers())
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("a non-empty opencode set writes a file")
	}
	text := string(document)
	if !strings.Contains(text, `"command":["npx","-y","figma-developer-mcp","--stdio"],"type":"local"`) {
		t.Fatalf("opencode stdio shape is wrong: %q", text)
	}
	if !strings.Contains(text, `"type":"remote","url":"https://mcp.example.com/docs"`) {
		t.Fatalf("opencode http shape is wrong: %q", text)
	}
	if strings.Contains(text, "codex-only") {
		t.Fatalf("the codex-only server must not enter the opencode set: %q", text)
	}
}
