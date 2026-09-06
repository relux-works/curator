package envfragment

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func testFragment() *Fragment {
	return &Fragment{
		Environment: "claude_code",
		Profile:     "companyA",
		LockSHA256:  "0305581b4f24d74ce271f9809d24f77b794b5ef9ad004bcc5c398fa2ea2e54ab",
		Winner:      "higher-weight",
		Placement:   "winner-last",
		Env:         map[string]string{"CLAUDE_CONFIG_DIR": "/manager/environments/companyA/claude_code"},
		SystemPrompt: &SystemPrompt{
			Path: "/manager/environments/companyA/claude_code/.agent-context/system-prompt.md",
			Channels: []envregistry.Channel{
				{Kind: "flag", Semantics: "append", Flag: "--append-system-prompt-file", Argument: "path"},
				{Kind: "flag", Semantics: "replace", Flag: "--system-prompt-file", Argument: "path"},
			},
		},
		MCP: &MCP{
			Path:     "/manager/environments/companyA/claude_code/.agent-context/mcp/claude_code.json",
			EnvNames: []string{"FIGMA_API_KEY"},
			Channels: []envregistry.Channel{
				{Kind: "flag", Flag: "--mcp-config", Argument: "path", With: []string{"--strict-mcp-config"}},
			},
		},
	}
}

// TestFragmentJSONCanonical narrows the canonical-form gate: --format json
// is the CCJ-1 bytes plus exactly one LF, so weakening canonicalization
// (key order, spacing) fails this byte-exact test.
func TestFragmentJSONCanonical(t *testing.T) {
	document, err := testFragment().JSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(document)
	if !strings.HasSuffix(text, "}\n") || strings.HasSuffix(text, "\n\n") {
		t.Fatalf("fragment json is not CCJ-1 plus one LF: %q", text)
	}
	want := `{"env":{"CLAUDE_CONFIG_DIR":"/manager/environments/companyA/claude_code"},"environment":"claude_code","fragment":"launch-env-fragment-v1","mcp":{"channels":[{"argument":"path","flag":"--mcp-config","kind":"flag","with":["--strict-mcp-config"]}],"env_names":["FIGMA_API_KEY"],"path":"/manager/environments/companyA/claude_code/.agent-context/mcp/claude_code.json"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"0305581b4f24d74ce271f9809d24f77b794b5ef9ad004bcc5c398fa2ea2e54ab","name":"companyA"},"system_prompt":{"channels":[{"argument":"path","flag":"--append-system-prompt-file","kind":"flag","semantics":"append"},{"argument":"path","flag":"--system-prompt-file","kind":"flag","semantics":"replace"}],"path":"/manager/environments/companyA/claude_code/.agent-context/system-prompt.md"}}` + "\n"
	if text != want {
		t.Fatalf("fragment bytes differ:\n got %q\nwant %q", text, want)
	}
}

func TestFragmentCodexNameChannel(t *testing.T) {
	fragment := &Fragment{
		Environment: "codex_cli",
		Profile:     "companyA",
		LockSHA256:  strings.Repeat("a", 64),
		Winner:      "higher-weight",
		Placement:   "winner-last",
		Env:         map[string]string{"CODEX_HOME": "/manager/environments/companyA/codex_cli"},
		MCP: &MCP{
			Path:     "/manager/environments/companyA/codex_cli/curator-mcp.config.toml",
			EnvNames: []string{},
			Channels: []envregistry.Channel{
				{Kind: "flag", Flag: "-p", Argument: "name", Name: "curator-mcp"},
			},
		},
	}
	document, err := fragment.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(document), `{"argument":"name","flag":"-p","kind":"flag","name":"curator-mcp"}`) {
		t.Fatalf("codex mcp channel descriptor is wrong: %q", document)
	}
}

func TestFragmentEnvAndShellFormats(t *testing.T) {
	adapter, err := envregistry.ByID("claude_code")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(testFragment().EnvFormat(adapter)); got != "CLAUDE_CONFIG_DIR=/manager/environments/companyA/claude_code\n" {
		t.Fatalf("env format %q", got)
	}
	if got := string(testFragment().ShellFormat(adapter)); got != "export CLAUDE_CONFIG_DIR='/manager/environments/companyA/claude_code'\n" {
		t.Fatalf("shell format %q", got)
	}
	fragment := testFragment()
	fragment.Env["CLAUDE_CONFIG_DIR"] = "/manager/it's/quoted"
	if got := string(fragment.ShellFormat(adapter)); !strings.Contains(got, `export CLAUDE_CONFIG_DIR='/manager/it'\''s/quoted'`) {
		t.Fatalf("shell quoting %q", got)
	}
}

func TestFragmentOpenCodeVariableOrder(t *testing.T) {
	adapter, err := envregistry.ByID("opencode")
	if err != nil {
		t.Fatal(err)
	}
	fragment := &Fragment{
		Environment: "opencode",
		Profile:     "companyA",
		LockSHA256:  strings.Repeat("b", 64),
		Winner:      "higher-weight",
		Placement:   "winner-last",
		Env:         map[string]string{"XDG_CONFIG_HOME": "/manager/environments/companyA/opencode"},
		MCP: &MCP{
			Path:     "/manager/environments/companyA/opencode/opencode/.agent-context/mcp/opencode.json",
			EnvNames: []string{},
			Channels: []envregistry.Channel{{Kind: "variable", Variable: "OPENCODE_CONFIG"}},
		},
	}
	got := string(fragment.EnvFormat(adapter))
	want := "XDG_CONFIG_HOME=/manager/environments/companyA/opencode\nOPENCODE_CONFIG=/manager/environments/companyA/opencode/opencode/.agent-context/mcp/opencode.json\n"
	if got != want {
		t.Fatalf("opencode variable order differs:\n got %q\nwant %q", got, want)
	}
}

// TestCheckBoundary narrows the §10.3 gate: a profile-derived variable
// name, a relative value, a .. traversal, and a value outside the
// environments root must each fail the build, and weakening any one check
// fails this test.
func TestCheckBoundary(t *testing.T) {
	adapter, err := envregistry.ByID("claude_code")
	if err != nil {
		t.Fatal(err)
	}
	root := "/manager/environments"
	good := testFragment()
	good.Env = map[string]string{"CLAUDE_CONFIG_DIR": "/manager/environments/companyA/claude_code"}
	if err := CheckBoundary(adapter, root, good); err != nil {
		t.Fatalf("a conforming fragment must pass the boundary: %v", err)
	}
	badName := testFragment()
	badName.Env = map[string]string{"CLAUDE_CONFIG_DIR": "/manager/environments/companyA/claude_code", "EVIL": "/manager/environments/x"}
	if err := CheckBoundary(adapter, root, badName); err == nil {
		t.Fatal("a profile-derived variable name must fail the boundary")
	}
	relative := testFragment()
	relative.Env = map[string]string{"CLAUDE_CONFIG_DIR": "environments/companyA/claude_code"}
	if err := CheckBoundary(adapter, root, relative); err == nil {
		t.Fatal("a relative value must fail the boundary")
	}
	traversal := testFragment()
	traversal.Env = map[string]string{"CLAUDE_CONFIG_DIR": "/manager/environments/../escape"}
	if err := CheckBoundary(adapter, root, traversal); err == nil {
		t.Fatal("a .. traversal must fail the boundary even when it resolves inside")
	}
	outside := testFragment()
	outside.Env = map[string]string{"CLAUDE_CONFIG_DIR": "/tmp/evil"}
	if err := CheckBoundary(adapter, root, outside); err == nil {
		t.Fatal("a value outside the environments root must fail the boundary")
	}
	outsideMCP := testFragment()
	outsideMCP.MCP.Path = "/tmp/evil.json"
	if err := CheckBoundary(adapter, root, outsideMCP); err == nil {
		t.Fatal("an mcp path outside the environments root must fail the boundary")
	}
}

// TestBoundEnvNames narrows the double bound: reserved names drop even
// when passable is unbounded, and the passable list intersects the rest.
func TestBoundEnvNames(t *testing.T) {
	got := BoundEnvNames([]string{"FIGMA_API_KEY", "PATH", "HOME"}, nil)
	if len(got) != 1 || got[0] != "FIGMA_API_KEY" {
		t.Fatalf("reserved names must drop under unbounded passable: %v", got)
	}
	got = BoundEnvNames([]string{"AAA", "BBB"}, []string{"BBB"})
	if len(got) != 1 || got[0] != "BBB" {
		t.Fatalf("passable must intersect: %v", got)
	}
	got = BoundEnvNames([]string{"AAA"}, []string{})
	if len(got) != 0 {
		t.Fatalf("an empty passable list bounds everything out: %v", got)
	}
}
