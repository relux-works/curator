package audit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay keeps the classifier gate but admits a forbidden host/cache
// class. Only exit 1 with the named behavioral assertion counts as a kill.
func TestNetworkContactNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	type mutation struct{ name, file, before, after, pkg, run, fails, message string }
	mutants := []mutation{
		{"markdown-scripts-always-visible", "internal/audit/audit.go", `includeScripts = len(spec.Commands) == 0`, `includeScripts = true`, "./internal/audit", "^TestMarkdownScopeScriptsAndManifest$/^exports$", "TestMarkdownScopeScriptsAndManifest/exports", "excluded Markdown refused"},
		{"markdown-scripts-never-visible", "internal/audit/audit.go", `includeScripts = len(spec.Commands) == 0`, `includeScripts = false`, "./internal/audit", "^TestMarkdownScopeScriptsAndManifest$/^commandless$", "TestMarkdownScopeScriptsAndManifest/commandless", "visible Markdown verdict"},

		{"markdown-everywhere", "internal/audit/audit.go", `if markdown && (kind == contextlock.KindSkill || contextModules) {`, `if false && markdown && (kind == contextlock.KindSkill || contextModules) {`, "./internal/audit", "^TestMarkdownScopeRepositoryNoise$", "TestMarkdownScopeRepositoryNoise", "excluded Markdown refused"},
		{"markdown-drop-references", "internal/audit/audit.go", `return scan(path, posix) // Agent-facing Markdown.`, `if strings.HasPrefix(posix, "references/") { return nil }; return scan(path, posix) // Agent-facing Markdown.`, "./internal/audit", "^TestMarkdownScopeVisibleContacts$/^references/guide.md$", "TestMarkdownScopeVisibleContacts/references/guide.md", "visible Markdown verdict"},
		{"markdown-case-sensitive-entry", "internal/audit/audit.go", "", "", "./internal/audit", "^TestMarkdownScopeCaseAliases$", "TestMarkdownScopeCaseAliases/", "visible Markdown verdict"},
		{"markdown-reuse-ruleset-six", "internal/audit/audit.go", `RulesetVersion = "` + RulesetVersion + `"`, `RulesetVersion = "6"`, "./internal/audit", "^TestMarkdownScopeInvalidatesRulesetSixAllowance$", "TestMarkdownScopeInvalidatesRulesetSixAllowance/", "visible Markdown verdict"},
		{"markdown-reuse-ruleset-five", "internal/audit/audit.go", `RulesetVersion = "` + RulesetVersion + `"`, `RulesetVersion = "5"`, "./internal/audit", "^TestMarkdownScopeInvalidatesRulesetFive$", "TestMarkdownScopeInvalidatesRulesetFive/", "stale Markdown refusal reused"},

		{"markdown-drop-context-modules", "internal/audit/audit.go", `for _, module := range manifest.Modules {`, `for _, module := range manifest.Modules[:0] {`, "./internal/envprofile", "^TestProfileInstallMarkdownContextCLI$", "TestProfileInstallMarkdownContextCLI/undeclared", "CLI context not specifically refused"},
		{"markdown-reuse-ruleset-seven", "internal/audit/audit.go", `RulesetVersion = "` + RulesetVersion + `"`, `RulesetVersion = "7"`, "./internal/audit", "^TestMarkdownScopeInvalidatesRulesetSevenAllowance$", "TestMarkdownScopeInvalidatesRulesetSevenAllowance/", "old ruleset-7 record changed"},
		{"markdown-kind-cache-borrow", "internal/audit/audit.go", `if data.Kind != subject.Kind {`, `if false && data.Kind != subject.Kind {`, "./internal/audit", "^TestMarkdownScopeCacheBindsKind$", "TestMarkdownScopeCacheBindsKind/", "kind cache borrowed allowance"},

		{"markdown-ignore-missing-context-module", "internal/audit/audit.go", `Severity: SeverityCritical, File: posix, Evidence: "cannot read declared context module"`, `Severity: SeverityInfo, File: posix, Evidence: "cannot read declared context module"`, "./internal/audit", "^TestMarkdownScopeContextMissingModule$", "TestMarkdownScopeContextMissingModule", "missing module verdict"},

		{"full-url-percent-path", "internal/audit/audit.go", `\\\\/?#`, `\\\\`, "./internal/audit", "^TestNetworkContactNormalizationBoundaries$/^percent-path-real$", "TestNetworkContactNormalizationBoundaries/percent-path-real", "URL classification: []"},
		{"contains-test", "internal/audit/audit.go", `switch host[lastDot+1:] {`, `if strings.Contains(host, ".test") { return host, true }; switch host[lastDot+1:] {`, "./internal/audit", "^TestNetworkContactVectors$/^lookalike-nested-test$", "TestNetworkContactVectors/lookalike-nested-test", "Gate pass=false"},
		{"exempt-example-com", "internal/audit/audit.go", `switch host[lastDot+1:] {`, `if host == "example.com" { return host, true }; switch host[lastDot+1:] {`, "./internal/audit", "^TestNetworkContactVectors$/^documentation-com$", "TestNetworkContactVectors/documentation-com", "Gate pass=false"},
		{"exempt-bare-test", "internal/audit/audit.go", `if lastDot <= 0 {`, `if host == "test" { return host, true }; if lastDot <= 0 {`, "./internal/audit", "^TestNetworkContactVectors$/^bare-test$", "TestNetworkContactVectors/bare-test", "Gate pass=false"},
		{"reserved-clears-file", "internal/audit/audit.go", `if exempt || declaredHost(declaredHosts, host) {`, `if exempt { return nil }; if exempt || declaredHost(declaredHosts, host) {`, "./internal/audit", "^TestNetworkContactVectors$/^mixed-reserved-first$", "TestNetworkContactVectors/mixed-reserved-first", "Gate pass=false"},
		{"reuse-ruleset-three", "internal/audit/audit.go", `RulesetVersion = "` + RulesetVersion + `"`, `RulesetVersion = "3"`, "./internal/audit", "^TestNetworkContactCacheInvalidatesRulesetThree$", "TestNetworkContactCacheInvalidatesRulesetThree/", "stale ruleset reused"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Contains(mutant.name, "case-sensitive-entry") || strings.Contains(mutant.name, "ruleset-six") {
				if !caseInsensitiveSnapshot(t, t.TempDir()) {
					t.Skip("alias control requires a case-insensitive filesystem")
				}
			}
			original := filepath.Join(root, filepath.FromSlash(mutant.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if mutant.name == "markdown-case-sensitive-entry" {
				source = rulesetSixSource(t, source)
			} else {
				if strings.Count(source, mutant.before) != 1 {
					t.Fatal("mutation anchor is not unique")
				}
				source = strings.Replace(source, mutant.before, mutant.after, 1)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", mutant.run, mutant.pkg)
			command.Dir = root
			if mutant.name == "markdown-drop-context-modules" {
				// The CLI regression builds a child binary; carry the same source overlay into it.
				command.Env = append(os.Environ(), "GOFLAGS=-overlay="+overlayPath)
			}
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails) || !strings.Contains(string(output), mutant.message) {
				t.Fatalf("mutant survived or lacked intended behavioral failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
