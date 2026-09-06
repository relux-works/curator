package contextmaterialize

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

func testLock() *contextlock.Lock {
	return &contextlock.Lock{
		Root: "companyA",
		Members: []contextlock.Member{
			{Name: "companyA", Kind: contextlock.KindContext, Version: "2.3.0", Commit: "0123456789abcdef0123456789abcdef01234567", Weight: 100},
		},
	}
}

func testPackages() map[string]Package {
	return map[string]Package{
		"companyA": {
			HasContext: true,
			Modules: []Module{
				{Module: contextpkg.Module{Path: "00-base.md"}, Bytes: []byte("# Base\n")},
				{Module: contextpkg.Module{Path: "20-claude.md", Environments: []string{"claude_code"}}, Bytes: []byte("# Claude\n")},
				{Module: contextpkg.Module{Path: "90-system.md", Class: "system"}, Bytes: []byte("Be terse.\n")},
			},
		},
	}
}

func TestReferencedClaude(t *testing.T) {
	lock := testLock()
	files, written, err := Referenced(lock, "sha256:aaaa", DefaultPrecedence, "claude_code", testPackages())
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("referenced claude_code is written")
	}
	root, ok := files["CLAUDE.md"]
	if !ok {
		t.Fatal("no CLAUDE.md in the file set")
	}
	text := string(root)
	if !strings.Contains(text, "curator-root-context-v2") {
		t.Fatal("root file carries no generation header")
	}
	if !strings.Contains(text, "## Context: companyA 2.3.0") {
		t.Fatal("root file carries no chapter part")
	}
	if !strings.Contains(text, "@.agent-context/modules/companyA/00-base.md\n") {
		t.Fatal("root file carries no @ reference part for the shared module")
	}
	if !strings.Contains(text, "@.agent-context/modules/companyA/20-claude.md\n") {
		t.Fatal("root file carries no @ reference part for the claude-only module")
	}
	if strings.Contains(text, "90-system.md") {
		t.Fatal("system modules must not enter the referenced root file")
	}
	if got := string(files[".agent-context/modules/companyA/00-base.md"]); got != "# Base\n" {
		t.Fatalf("module file bytes %q are not exact", got)
	}
	if _, ok := files[".agent-context/modules/companyA/90-system.md"]; ok {
		t.Fatal("system modules must not materialize as module files")
	}
	if got := SurfaceHash(files); len(got) != 7+64 {
		t.Fatalf("surface hash %q has the wrong shape", got)
	}
}

// TestReferencedRejectsCodex narrows the §7.2 form-support gate: weakening
// SupportsReferenced to admit codex_cli must fail this test, because the
// reference syntax for codex_cli is undefined and emitting monolithic bytes
// here would silently change the fallback contract.
func TestReferencedRejectsCodex(t *testing.T) {
	_, _, err := Referenced(testLock(), "sha256:aaaa", DefaultPrecedence, "codex_cli", testPackages())
	if err == nil || !strings.Contains(err.Error(), DiagFormUnsupported) {
		t.Fatalf("codex_cli referenced form must fail with %s, got %v", DiagFormUnsupported, err)
	}
}

func TestReferencedRejectsPi(t *testing.T) {
	_, _, err := Referenced(testLock(), "sha256:aaaa", DefaultPrecedence, "pi", testPackages())
	if err == nil || !strings.Contains(err.Error(), DiagFormUnsupported) {
		t.Fatalf("pi referenced form must fail with %s, got %v", DiagFormUnsupported, err)
	}
}

func TestReferencedOpenCode(t *testing.T) {
	lock := testLock()
	files, written, err := Referenced(lock, "sha256:aaaa", DefaultPrecedence, "opencode", testPackages())
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("referenced opencode is written")
	}
	root := string(files["AGENTS.md"])
	if strings.Contains(root, "@.agent-context") || strings.Contains(root, "## Context:") {
		t.Fatal("the opencode root file is the header part alone")
	}
	if !strings.Contains(root, "curator-root-context-v2") {
		t.Fatal("the opencode root file carries the generation header")
	}
	config := string(files["opencode.json"])
	if !strings.HasSuffix(config, "}\n") || strings.Contains(config, "\n\n") {
		t.Fatalf("opencode.json is not one CCJ-1 line plus LF: %q", config)
	}
	if !strings.Contains(config, `".agent-context/modules/companyA/00-base.md"`) {
		t.Fatalf("opencode.json carries no instructions entry: %q", config)
	}
	if strings.Contains(config, "20-claude.md") {
		t.Fatalf("the claude-only module must not enter the opencode set: %q", config)
	}
}

// TestReferencedOpenCodeZeroModules narrows the empty-instructions gate: an
// absent member would decode as null somewhere downstream, so the empty set
// must render as [] exactly.
func TestReferencedOpenCodeZeroModules(t *testing.T) {
	packages := map[string]Package{"companyA": {HasContext: true}}
	files, written, err := Referenced(testLock(), "sha256:aaaa", DefaultPrecedence, "opencode", packages)
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("a zero-module referenced home still writes the header and config")
	}
	if got := string(files["opencode.json"]); got != "{\"instructions\":[]}\n" {
		t.Fatalf("empty instructions render %q, want %q", got, "{\"instructions\":[]}\n")
	}
	if _, ok := files["AGENTS.md"]; !ok {
		t.Fatal("a zero-module referenced home still writes the header-only root file")
	}
}

func TestReferencedNoContext(t *testing.T) {
	packages := map[string]Package{"companyA": {HasContext: false}}
	_, written, err := Referenced(testLock(), "sha256:aaaa", DefaultPrecedence, "claude_code", packages)
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("a root without context declares no surface and writes nothing")
	}
}

// TestDetectPathCollision narrows the §5 platform-path gate: folding two
// protocol paths to one platform path must fail before any write, and the
// admission of exactly one folding pair must fail this test.
func TestDetectPathCollision(t *testing.T) {
	if err := DetectPathCollision([]string{"CLAUDE.md", ".agent-context/modules/a/00-base.md"}); err != nil {
		t.Fatalf("distinct paths must not collide: %v", err)
	}
	err := DetectPathCollision([]string{"CLAUDE.md", "claude.md"})
	if err == nil || !strings.Contains(err.Error(), DiagPathCollision) {
		t.Fatalf("folding paths must fail with %s, got %v", DiagPathCollision, err)
	}
}

func TestRootTarget(t *testing.T) {
	if got := RootTarget("claude_code"); got != "CLAUDE.md" {
		t.Fatalf("claude root target %q", got)
	}
	for _, env := range []string{"codex_cli", "opencode", "pi"} {
		if got := RootTarget(env); got != "AGENTS.md" {
			t.Fatalf("%s root target %q", env, got)
		}
	}
}
