package envfragment

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func testFragment() *Fragment {
	return &Fragment{
		Environment: "claude_code",
		Profile:     "companyA",
		LockSHA256:  "0305581b4f24d74ce271f9809d24f77b794b5ef9ad004bcc5c398fa2ea2e54ab",
		Permissions: Permissions{Mode: "yolo", Source: "profile"},
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
	want := `{"env":{"CLAUDE_CONFIG_DIR":"/manager/environments/companyA/claude_code"},"environment":"claude_code","fragment":"launch-env-fragment-v2","mcp":{"channels":[{"argument":"path","flag":"--mcp-config","kind":"flag","with":["--strict-mcp-config"]}],"env_names":["FIGMA_API_KEY"],"path":"/manager/environments/companyA/claude_code/.agent-context/mcp/claude_code.json"},"permissions":{"locked":false,"mode":"yolo","source":"profile"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"0305581b4f24d74ce271f9809d24f77b794b5ef9ad004bcc5c398fa2ea2e54ab","name":"companyA"},"system_prompt":{"channels":[{"argument":"path","flag":"--append-system-prompt-file","kind":"flag","semantics":"append"},{"argument":"path","flag":"--system-prompt-file","kind":"flag","semantics":"replace"}],"path":"/manager/environments/companyA/claude_code/.agent-context/system-prompt.md"}}` + "\n"
	if text != want {
		t.Fatalf("fragment bytes differ:\n got %q\nwant %q", text, want)
	}
}

func TestFragmentCodexNameChannel(t *testing.T) {
	fragment := &Fragment{
		Environment: "codex_cli",
		Profile:     "companyA",
		LockSHA256:  strings.Repeat("a", 64),
		Permissions: Permissions{Mode: "native", Source: "default"},
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

// TestFragmentEnvCarriesOnlyEnvObject narrows the §10.1 gate: --format
// env|shell render exactly the fragment's env object, never a
// variable-kind channel value, so the three formats stay renderings of
// one object.
func TestFragmentEnvCarriesOnlyEnvObject(t *testing.T) {
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
	if got := string(fragment.EnvFormat(adapter)); got != "XDG_CONFIG_HOME=/manager/environments/companyA/opencode\n" {
		t.Fatalf("env format carries more than the env object: %q", got)
	}
	if got := string(fragment.ShellFormat(adapter)); got != "export XDG_CONFIG_HOME='/manager/environments/companyA/opencode'\n" {
		t.Fatalf("shell format carries more than the env object: %q", got)
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
	// Platform-absolute fixture root: CheckBoundary is built on
	// path/filepath, so POSIX-rooted literals are not absolute on Windows.
	// Every value below is derived from this root with filepath.Join.
	base := t.TempDir()
	root := filepath.Join(base, "environments")
	newFragment := func(envValue string) *Fragment {
		fragment := testFragment()
		fragment.Env = map[string]string{"CLAUDE_CONFIG_DIR": envValue}
		fragment.SystemPrompt.Path = filepath.Join(root, "companyA", "claude_code", ".agent-context", "system-prompt.md")
		fragment.MCP.Path = filepath.Join(root, "companyA", "claude_code", ".agent-context", "mcp", "claude_code.json")
		return fragment
	}
	good := newFragment(filepath.Join(root, "companyA", "claude_code"))
	if err := CheckBoundary(adapter, root, good); err != nil {
		t.Fatalf("a conforming fragment must pass the boundary: %v", err)
	}
	badName := newFragment(filepath.Join(root, "companyA", "claude_code"))
	badName.Env["EVIL"] = filepath.Join(root, "x")
	if err := CheckBoundary(adapter, root, badName); err == nil {
		t.Fatal("a profile-derived variable name must fail the boundary")
	}
	relative := newFragment(filepath.Join("environments", "companyA", "claude_code"))
	if err := CheckBoundary(adapter, root, relative); err == nil {
		t.Fatal("a relative value must fail the boundary")
	}
	traversal := newFragment(root + string(filepath.Separator) + ".." + string(filepath.Separator) + "escape")
	if err := CheckBoundary(adapter, root, traversal); err == nil {
		t.Fatal("a .. traversal must fail the boundary even when it resolves inside")
	}
	outside := newFragment(filepath.Join(base, "evil"))
	if err := CheckBoundary(adapter, root, outside); err == nil {
		t.Fatal("a value outside the environments root must fail the boundary")
	}
	sibling := newFragment(filepath.Join(base, "profiles", "acme", "rendered", "claude_code", "AGENTS.md"))
	if err := CheckBoundary(adapter, root, sibling); err == nil {
		t.Fatal("a value below the environments root parent but outside the root must fail the boundary")
	}
	outsideMCP := newFragment(filepath.Join(root, "companyA", "claude_code"))
	outsideMCP.MCP.Path = filepath.Join(base, "evil.json")
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

// TestResolvePassthroughProfiles drives the §10.3 S4 matrix: s4-warn
// keeps unbounded behaviour for an absent knob with the unlisted warning,
// s4-enforce drops unlisted names with the dropped warning, an explicit
// null stays unbounded and silent under both, and reserved names stay
// excluded everywhere.
func TestResolvePassthroughProfiles(t *testing.T) {
	cases := []struct {
		name       string
		requested  []string
		passable   []string
		knobSet    bool
		profile    S4Profile
		passed     []string
		dropped    []string
		diagnostic string
	}{
		{"warn-absent-passes-with-warning",
			[]string{"FIGMA_API_KEY"}, nil, false, S4Warn,
			[]string{"FIGMA_API_KEY"}, []string{}, DiagPassthroughUnlisted},
		{"enforce-absent-drops-all",
			[]string{"FIGMA_API_KEY"}, nil, false, S4Enforce,
			[]string{}, []string{"FIGMA_API_KEY"}, DiagPassthroughDropped},
		{"warn-null-unbounded-silent",
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, nil, true, S4Warn,
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, []string{}, ""},
		{"enforce-null-unbounded-silent",
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, nil, true, S4Enforce,
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, []string{}, ""},
		{"warn-list-bounds-silently",
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, []string{"FIGMA_API_KEY"}, true, S4Warn,
			[]string{"FIGMA_API_KEY"}, []string{"GITHUB_TOKEN"}, ""},
		{"enforce-list-drops-with-diagnostic",
			[]string{"FIGMA_API_KEY", "GITHUB_TOKEN"}, []string{"FIGMA_API_KEY"}, true, S4Enforce,
			[]string{"FIGMA_API_KEY"}, []string{"GITHUB_TOKEN"}, DiagPassthroughDropped},
		{"enforce-empty-list-drops-all",
			[]string{"FIGMA_API_KEY"}, []string{}, true, S4Enforce,
			[]string{}, []string{"FIGMA_API_KEY"}, DiagPassthroughDropped},
		{"warn-absent-nothing-passed-silent",
			nil, nil, false, S4Warn, []string{}, []string{}, ""},
		{"warn-absent-reserved-excluded",
			[]string{"FIGMA_API_KEY", "PATH", "LD_PRELOAD"}, nil, false, S4Warn,
			[]string{"FIGMA_API_KEY"}, []string{}, DiagPassthroughUnlisted},
		{"enforce-list-reserved-excluded-silently",
			[]string{"FIGMA_API_KEY", "PATH"}, []string{}, true, S4Enforce,
			[]string{}, []string{"FIGMA_API_KEY"}, DiagPassthroughDropped},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict := ResolvePassthrough(tc.requested, tc.passable, tc.knobSet, tc.profile)
			if strings.Join(verdict.Passed, ",") != strings.Join(tc.passed, ",") {
				t.Fatalf("passed = %v, want %v", verdict.Passed, tc.passed)
			}
			if strings.Join(verdict.Dropped, ",") != strings.Join(tc.dropped, ",") {
				t.Fatalf("dropped = %v, want %v", verdict.Dropped, tc.dropped)
			}
			if tc.diagnostic == "" {
				if verdict.Warning != "" {
					t.Fatalf("warning = %q, want silent", verdict.Warning)
				}
				return
			}
			if !strings.HasPrefix(verdict.Warning, tc.diagnostic+":") {
				t.Fatalf("warning = %q, want the %s diagnostic", verdict.Warning, tc.diagnostic)
			}
			named := tc.dropped
			if tc.diagnostic == DiagPassthroughUnlisted {
				named = tc.passed
			}
			for _, name := range named {
				if !strings.Contains(verdict.Warning, name) {
					t.Fatalf("warning %q names no %q", verdict.Warning, name)
				}
			}
			if tc.diagnostic == DiagPassthroughUnlisted {
				if !strings.Contains(verdict.Warning, "passable_env_names") {
					t.Fatalf("warning %q names no knob", verdict.Warning)
				}
				if !strings.Contains(verdict.Warning, MigrationHint) {
					t.Fatalf("warning %q carries no migration hint", verdict.Warning)
				}
			}
		})
	}
}

// TestEffectivePassable pins the profile default behind the knob
// presence bit: absent is unbounded under s4-warn and empty under
// s4-enforce, explicit null is unbounded under both.
func TestEffectivePassable(t *testing.T) {
	if got := EffectivePassable(nil, false, S4Warn); got != nil {
		t.Fatalf("s4-warn absent = %v, want unbounded", got)
	}
	if got := EffectivePassable(nil, false, S4Enforce); got == nil || len(got) != 0 {
		t.Fatalf("s4-enforce absent = %v, want empty", got)
	}
	if got := EffectivePassable(nil, true, S4Warn); got != nil {
		t.Fatalf("s4-warn null = %v, want unbounded", got)
	}
	if got := EffectivePassable(nil, true, S4Enforce); got != nil {
		t.Fatalf("s4-enforce null = %v, want unbounded", got)
	}
	if got := EffectivePassable([]string{"A_B"}, true, S4Warn); len(got) != 1 || got[0] != "A_B" {
		t.Fatalf("s4-warn list = %v", got)
	}
}

// TestActiveProfileIsWarn pins the shipped default: the warning release
// goes out first, and the flip is a later release.
func TestActiveProfileIsWarn(t *testing.T) {
	if ActiveS4Profile != S4Warn {
		t.Fatalf("ActiveS4Profile = %q, want s4-warn first", ActiveS4Profile)
	}
}
