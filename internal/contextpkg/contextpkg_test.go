package contextpkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production entry points under test: ParseManifest, ParseMCP, LoadManifest,
// ValidateModules, ValidateModuleBytes, ReservedEnvName.

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestValidManifestParses checks the happy path, including root-only
// weights and one module with a selector.
func TestValidManifestParses(t *testing.T) {
	manifest, err := ParseManifest([]byte(`{
		"schema_version": 1, "name": "acme", "version": "1.2.0",
		"weights": {"other": 7},
		"context": {"modules": [
			{"path": "a.md"},
			{"path": "b.md", "environments": ["claude_code"], "class": "system"}
		]},
		"requires": {"contexts": {"other": {"git": "https://example.com/other", "range": "^1.0.0", "weight": 7}}}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "acme" || manifest.Version != "1.2.0" || !manifest.HasContext {
		t.Fatalf("manifest %+v", manifest)
	}
	if manifest.Weights["other"] != 7 {
		t.Fatalf("weights %+v", manifest.Weights)
	}
	if len(manifest.Modules) != 2 || manifest.Modules[1].Class != "system" {
		t.Fatalf("modules %+v", manifest.Modules)
	}
	if manifest.Contexts["other"].Range != "^1.0.0" || manifest.Contexts["other"].Form() != "range" {
		t.Fatalf("requirement %+v", manifest.Contexts["other"])
	}
}

// TestUnknownFieldIsRejected narrows the strict-reader gate: unknown fields
// anywhere must fail with context_manifest_invalid. A mutant that tolerates
// one extra key must fail this test.
func TestUnknownFieldIsRejected(t *testing.T) {
	payloads := []string{
		`{"schema_version": 1, "name": "a", "version": "1.0.0", "surprise": 1}`,
		`{"schema_version": 1, "name": "a", "version": "1.0.0", "context": {"modules": [{"path": "a.md", "bogus": true}]}}`,
		`{"schema_version": 1, "name": "a", "version": "1.0.0", "requires": {"contexts": {"x": {"git": "u", "range": "*", "bogus": 1}}}}`,
	}
	for _, payload := range payloads {
		_, err := ParseManifest([]byte(payload))
		if err == nil {
			t.Fatalf("payload %s must not parse", payload)
		}
		diag, ok := err.(*Diagnostic)
		if !ok || diag.Code != DiagManifestInvalid {
			t.Fatalf("payload %s error %v carries no %s", payload, err, DiagManifestInvalid)
		}
	}
}

// TestRequirementNeedsExactlyOneForm narrows the form gate: zero or two of
// range/tag/revision must fail.
func TestRequirementNeedsExactlyOneForm(t *testing.T) {
	for _, requires := range []string{
		`{"contexts": {"x": {"git": "u"}}}`,
		`{"contexts": {"x": {"git": "u", "range": "*", "tag": "v1.0.0"}}}`,
	} {
		_, err := ParseManifest([]byte(`{"schema_version": 1, "name": "a", "version": "1.0.0", "requires": ` + requires + `}`))
		if err == nil {
			t.Fatalf("requires %s must not parse", requires)
		}
	}
}

// TestBadRangeInManifestIsRejected checks the manifest surfaces an
// unparsable range with the manifest diagnostic.
func TestBadRangeInManifestIsRejected(t *testing.T) {
	_, err := ParseManifest([]byte(`{"schema_version": 1, "name": "a", "version": "1.0.0",
		"requires": {"contexts": {"x": {"git": "u", "range": "1.0.0 - 2.0.0"}}}}`))
	if err == nil || !strings.Contains(err.Error(), DiagManifestInvalid) {
		t.Fatalf("hyphen range must fail as %s, got %v", DiagManifestInvalid, err)
	}
}

// TestValidMCPParses checks both transports.
func TestValidMCPParses(t *testing.T) {
	stdio, err := ParseMCP([]byte(`{"schema_version": 1, "name": "m", "version": "1.0.0",
		"server": {"transport": "stdio", "command": "serve", "args": ["--x"], "env_names": ["ACME_TOKEN"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if stdio.Server.Command != "serve" || len(stdio.Server.Args) != 1 {
		t.Fatalf("server %+v", stdio.Server)
	}
	http, err := ParseMCP([]byte(`{"schema_version": 1, "name": "m", "version": "1.0.0",
		"server": {"transport": "http", "url": "https://example.com/mcp"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if http.Server.URL != "https://example.com/mcp" {
		t.Fatalf("server %+v", http.Server)
	}
}

// TestMCPGate narrows the agent-mcp.json gate on its three named refusals:
// a pathed command, a non-https URL, and a manager-reserved env name. A
// mutant admitting exactly one of them must fail this test.
func TestMCPGate(t *testing.T) {
	payloads := []string{
		`{"schema_version": 1, "name": "m", "version": "1.0.0", "server": {"transport": "stdio", "command": "./serve", "args": []}}`,
		`{"schema_version": 1, "name": "m", "version": "1.0.0", "server": {"transport": "http", "url": "http://example.com/mcp"}}`,
		`{"schema_version": 1, "name": "m", "version": "1.0.0", "server": {"transport": "stdio", "command": "serve", "args": [], "env_names": ["PATH"]}}`,
	}
	for _, payload := range payloads {
		_, err := ParseMCP([]byte(payload))
		if err == nil {
			t.Fatalf("payload %s must not parse", payload)
		}
		diag, ok := err.(*Diagnostic)
		if !ok || diag.Code != DiagMCPInvalid {
			t.Fatalf("payload %s error %v carries no %s", payload, err, DiagMCPInvalid)
		}
	}
	if !ReservedEnvName("LD_PRELOAD") || ReservedEnvName("ACME_TOKEN") {
		t.Fatal("reserved-name classification is wrong")
	}
}

// TestModuleBytesRules pins the §3 byte rules: LF-only, valid UTF-8,
// exactly one trailing LF.
func TestModuleBytesRules(t *testing.T) {
	if err := ValidateModuleBytes([]byte("hello\n")); err != nil {
		t.Fatalf("valid module rejected: %v", err)
	}
	for _, payload := range []string{"", "no-trailing-lf", "crlf\r\n", "two\n\n", "nul\x00\n"} {
		_ = payload
	}
	if err := ValidateModuleBytes([]byte("no-trailing-lf")); err == nil {
		t.Fatal("missing trailing LF must fail")
	}
	if err := ValidateModuleBytes([]byte("crlf\r\n")); err == nil {
		t.Fatal("CRLF must fail")
	}
	if err := ValidateModuleBytes([]byte("two\n\n")); err == nil {
		t.Fatal("double trailing LF must fail")
	}
	if err := ValidateModuleBytes([]byte{0xff, 0xfe, '\n'}); err == nil {
		t.Fatal("invalid UTF-8 must fail")
	}
}

// TestMissingModuleIsReported checks LoadManifest plus ValidateModules over
// a real package root: a declared but absent module is profile_module_missing.
func TestMissingModuleIsReported(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ManifestName), `{"schema_version": 1, "name": "a",
		"version": "1.0.0", "context": {"modules": [{"path": "gone.md"}]}}`)
	writeFile(t, filepath.Join(root, ContextDir, ".keep"), "x\n")
	manifest, err := LoadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ValidateModules(root, manifest, nil)
	if err == nil || !strings.Contains(err.Error(), DiagModuleMissing) {
		t.Fatalf("absent module must fail as %s, got %v", DiagModuleMissing, err)
	}
}

// TestUnknownEnvironmentIsAWarning checks the selector discipline: an
// unregistered environment warns with profile_selector_unknown_environment
// and selects nothing, rather than failing.
func TestUnknownEnvironmentIsAWarning(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ManifestName), `{"schema_version": 1, "name": "a",
		"version": "1.0.0", "context": {"modules": [{"path": "a.md", "environments": ["nope"]}]}}`)
	writeFile(t, filepath.Join(root, ContextDir, "a.md"), "hello\n")
	manifest, err := LoadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	warnings, err := ValidateModules(root, manifest, map[string]bool{"claude_code": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], DiagSelectorUnknownEnvironment) {
		t.Fatalf("warnings %v", warnings)
	}
	if manifest.Modules[0].Applies("claude_code") || !manifest.Modules[0].Applies("nope") {
		t.Fatal("selector applies() is wrong")
	}
}
