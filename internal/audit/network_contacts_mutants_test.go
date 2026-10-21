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
		{"full-url-percent-path", "internal/audit/audit.go", `\\\\/?#`, `\\\\`, "./internal/audit", "^TestNetworkContactNormalizationBoundaries$/^percent-path-real$", "TestNetworkContactNormalizationBoundaries/percent-path-real", "URL classification: []"},
		{"contains-test", "internal/audit/audit.go", `switch host[lastDot+1:] {`, `if strings.Contains(host, ".test") { return host, true }; switch host[lastDot+1:] {`, "./internal/audit", "^TestNetworkContactVectors$/^lookalike-nested-test$", "TestNetworkContactVectors/lookalike-nested-test", "Gate pass=false"},
		{"exempt-example-com", "internal/audit/audit.go", `switch host[lastDot+1:] {`, `if host == "example.com" { return host, true }; switch host[lastDot+1:] {`, "./internal/audit", "^TestNetworkContactVectors$/^documentation-com$", "TestNetworkContactVectors/documentation-com", "Gate pass=false"},
		{"exempt-bare-test", "internal/audit/audit.go", `if lastDot <= 0 {`, `if host == "test" { return host, true }; if lastDot <= 0 {`, "./internal/audit", "^TestNetworkContactVectors$/^bare-test$", "TestNetworkContactVectors/bare-test", "Gate pass=false"},
		{"reserved-clears-file", "internal/audit/audit.go", `if exempt || declaredHost(declaredHosts, host) {`, `if exempt { return nil }; if exempt || declaredHost(declaredHosts, host) {`, "./internal/audit", "^TestNetworkContactVectors$/^mixed-reserved-first$", "TestNetworkContactVectors/mixed-reserved-first", "Gate pass=false"},
		{"reuse-ruleset-three", "internal/audit/audit.go", `RulesetVersion = "5"`, `RulesetVersion = "3"`, "./internal/audit", "^TestNetworkContactCacheInvalidatesRulesetThree$", "TestNetworkContactCacheInvalidatesRulesetThree/", "stale ruleset reused"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			original := filepath.Join(root, filepath.FromSlash(mutant.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0600); err != nil {
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
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails) || !strings.Contains(string(output), mutant.message) {
				t.Fatalf("mutant survived or lacked intended behavioral failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
