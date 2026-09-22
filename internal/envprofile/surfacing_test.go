package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
)

func gitLookPath() (string, error) { return exec.LookPath("git") }

func writeGitFile(t *testing.T, dir, name, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// Production entry points under test: Install, UpdateWithPolicy, Resolve,
// and StatusOf for the §2.3 surfacing rows, the §2.2 empty-allowlist
// warning, and the §10.3 S4 resolve bound.

// TestInstallSurfacesMCPDeclarations drives Install on a root requiring
// one http and one stdio declaration: Info.Surfacing carries the exact
// §2.3 rows in ascending package-name byte order — the stdio command,
// args, and env_names listed — and the empty allowlist warns.
func TestInstallSurfacesMCPDeclarations(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	alphaRepo := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "alpha", "version": "1.0.0",` +
			`"server": {"transport": "http", "url": "https://mcp.example.com/alpha", "env_names": []}}` + "\n",
	}, "v1.0.0")
	zuluRepo := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "zulu", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "uvx", "args": ["tool@1.0", "serve"], "env_names": ["ZULU_TOKEN"]}}` + "\n",
	}, "v1.0.0")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",` +
			`"requires": {"mcp": {` +
			`"zulu": {"git": "` + ids.serve(zuluRepo, "https://example.com/mcp-zulu") + `", "range": "*"},` +
			`"alpha": {"git": "` + ids.serve(alphaRepo, "https://example.com/mcp-alpha") + `", "range": "*"}}}}` + "\n",
	}, "v1.0.0")
	info, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/withmcp")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`mcp-declaration alpha 1.0.0 http command=- args=[] env_names=[]`,
		`mcp-declaration zulu 1.0.0 stdio command=uvx args=["tool@1.0","serve"] env_names=["ZULU_TOKEN"]`,
	}
	if !equalOrdered(info.Surfacing, want) {
		t.Fatalf("surfacing = %v, want %v", info.Surfacing, want)
	}
	assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
}

// TestInstallSurfacingSilentWithoutMCP narrows the empty-set edge: a
// root with no declaration packages surfaces no rows, and the allowlist
// warning still fires.
func TestInstallSurfacingSilentWithoutMCP(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := t.TempDir()
	writePackage(t, root, "acme", "1.0.0", "hello\n")
	info, _, _, err := Install(home, InstallOptions{Operand: root})
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Surfacing) != 0 {
		t.Fatalf("surfacing = %v, want no rows for an empty MCP set", info.Surfacing)
	}
	assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
}

// TestUpdateSurfacesCandidateMCPSet drives UpdateWithPolicy onto a root
// whose new revision adds a declaration package: the update moves, the
// surfacing rows describe the candidate lock's MCP set, and the empty
// allowlist warns.
func TestUpdateSurfacesCandidateMCPSet(t *testing.T) {
	if _, err := gitLookPath(); err != nil {
		t.Skip("no git on PATH")
	}
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	toolRepo := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "npx", "args": ["-y", "tool"], "env_names": ["TOOL_TOKEN"]}}` + "\n",
	}, "v1.0.0")
	toolOperand := ids.serve(toolRepo, "https://example.com/mcp-tool")
	root := t.TempDir()
	writeGitFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.0"}`+"\n")
	gitRun(t, root, "init")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "one")
	gitRun(t, root, "tag", "v1.0.0")
	if _, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/grows")}); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.1",`+
		`"requires": {"mcp": {"tool": {"git": "`+toolOperand+`", "range": "*"}}}}`+"\n")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "two")
	gitRun(t, root, "tag", "v1.0.1")
	info, moved, err := UpdateWithPolicy(home, "grows", Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if !moved {
		t.Fatal("the update must move to the revision carrying the declaration")
	}
	want := []string{
		`mcp-declaration tool 1.0.0 stdio command=npx args=["-y","tool"] env_names=["TOOL_TOKEN"]`,
	}
	if !equalOrdered(info.Surfacing, want) {
		t.Fatalf("surfacing = %v, want %v", info.Surfacing, want)
	}
	assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
}

// TestAllowlistWarningOperations narrows the §2.2 warning gate across
// install, update, and status: an empty effective allowlist warns on
// every operation and a non-empty one stays silent.
func TestAllowlistWarningOperations(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := t.TempDir()
	writePackage(t, root, "acme", "1.0.0", "hello\n")
	nonEmpty := Policy{MCPAllowlist: []string{"https://example.com/m"}}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: nonEmpty})
	if err != nil {
		t.Fatal(err)
	}
	assertWarningAbsent(t, info.Warnings, DiagAllowlistEmpty)
	updated, moved, err := UpdateWithPolicy(home, "acme", nonEmpty)
	if err != nil {
		t.Fatal(err)
	}
	if moved {
		t.Fatal("a path root with no overlays must not move on update")
	}
	assertWarningAbsent(t, updated.Warnings, DiagAllowlistEmpty)
	fx := writeManagedFixture(t, "acme")
	req := statusRequest(fx)
	req.Policy = nonEmpty
	status, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	assertWarningAbsent(t, status.Warnings, DiagAllowlistEmpty)
}

func assertWarningAbsent(t *testing.T, warnings []string, diagnostic string) {
	t.Helper()
	for _, warning := range warnings {
		if strings.HasPrefix(warning, diagnostic+":") || warning == diagnostic {
			t.Fatalf("warnings %v carry a forbidden %s", warnings, diagnostic)
		}
	}
}

// TestResolvePassthroughKnobs drives the production Resolve across the
// S4 knob shapes: an absent knob passes unbounded with the unlisted
// warning, an explicit null passes unbounded silently, and a configured
// list bounds.
func TestResolvePassthroughKnobs(t *testing.T) {
	cases := []struct {
		name     string
		passable []string
		knobSet  bool
		envNames []string
		warning  string
	}{
		{"absent-warns", nil, false, []string{"FIGMA_API_KEY"}, "mcp_env_passthrough_unlisted"},
		{"null-silent", nil, true, []string{"FIGMA_API_KEY"}, ""},
		{"list-bounds-silently", []string{"FIGMA_API_KEY"}, true, []string{"FIGMA_API_KEY"}, ""},
		{"empty-list-bounds-all", []string{}, true, []string{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			req := fx.request("claude_code")
			req.Repair = true
			req.Machine.PassableEnvNames = tc.passable
			req.Machine.PassableEnvNamesSet = tc.knobSet
			result, err := Resolve(req)
			if err != nil {
				t.Fatal(err)
			}
			var fragment map[string]any
			if err := json.Unmarshal(result.Document, &fragment); err != nil {
				t.Fatalf("fragment is not JSON: %v", err)
			}
			mcp, ok := fragment["mcp"].(map[string]any)
			if !ok {
				t.Fatalf("fragment carries no mcp section: %v", fragment)
			}
			var names []string
			for _, raw := range mcp["env_names"].([]any) {
				names = append(names, raw.(string))
			}
			if !equalOrdered(names, tc.envNames) {
				t.Fatalf("fragment env_names = %v, want %v", names, tc.envNames)
			}
			if tc.warning == "" {
				assertWarningAbsent(t, result.Warnings, "mcp_env_passthrough_unlisted")
				assertWarningAbsent(t, result.Warnings, "mcp_env_passthrough_dropped")
				return
			}
			assertWarningPresent(t, result.Warnings, tc.warning)
			for _, warning := range result.Warnings {
				if !strings.HasPrefix(warning, tc.warning+":") {
					continue
				}
				if !strings.Contains(warning, "FIGMA_API_KEY") || !strings.Contains(warning, "passable_env_names") {
					t.Fatalf("warning %q names no variable or knob", warning)
				}
			}
		})
	}
}

// TestStatusS4Posture drives StatusOf over the managed fixture: the §12
// posture carries the active S4 profile with the effective
// passable_env_names, the allowlist-empty warning row when applicable,
// and the §2.3 rows for the current profile of the reported scope.
func TestStatusS4Posture(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := statusRequest(fx)
	machine := req.Machine
	machine.PassableEnvNames = []string{"FIGMA_API_KEY"}
	machine.PassableEnvNamesSet = true
	req.Machine = machine
	req.Policy = Policy{}
	status, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	if status.S4Profile != "s4-warn" {
		t.Fatalf("s4_profile = %q, want the shipped s4-warn", status.S4Profile)
	}
	if !equalOrdered(status.PassableEnvNames, []string{"FIGMA_API_KEY"}) {
		t.Fatalf("passable_env_names = %v", status.PassableEnvNames)
	}
	assertWarningPresent(t, status.Warnings, DiagAllowlistEmpty)
	want := `mcp-declaration figma-devmode 1.2.0 stdio command=npx args=["-y","figma-developer-mcp","--stdio"] env_names=["FIGMA_API_KEY"]`
	if len(status.MCPDeclarations) != 1 {
		t.Fatalf("declarations = %+v, want one scope", status.MCPDeclarations)
	}
	declarations := status.MCPDeclarations[0]
	if declarations.Scope != "machine" || declarations.Profile != "acme" {
		t.Fatalf("declarations scope = %+v", declarations)
	}
	if !equalOrdered(declarations.Rows, []string{want}) {
		t.Fatalf("rows = %v, want %v", declarations.Rows, []string{want})
	}
	// An absent knob postures as unbounded.
	req.Machine = statusRequest(fx).Machine
	absent, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	if absent.PassableEnvNames != nil {
		t.Fatalf("absent knob postures %v, want unbounded", absent.PassableEnvNames)
	}
}

// TestSurfacingUnreadableManifest narrows the §8.4 edge: a declaration
// whose manifest cannot be read is reported as unreadable — never as an
// empty set — and neither the builder nor status fails.
func TestSurfacingUnreadableManifest(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	entry := contextstore.EntryDir(fx.home, "mcp", "figma-devmode", strings.Repeat("b", 40))
	if err := os.Remove(filepath.Join(entry, "agent-mcp.json")); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(fx.home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	rows, unreadable := surfacingRows(fx.home, newGitManager(fx.home), lock)
	if len(rows) != 0 {
		t.Fatalf("rows = %v, want none for an unreadable declaration", rows)
	}
	if len(unreadable) != 1 || !strings.Contains(unreadable[0], "figma-devmode") {
		t.Fatalf("unreadable = %v, want the declaration named", unreadable)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, warning := range status.Warnings {
		found = found || strings.Contains(warning, "figma-devmode")
	}
	if !found {
		t.Fatalf("status warnings %v name no unreadable declaration", status.Warnings)
	}
}
