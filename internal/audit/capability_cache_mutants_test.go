package audit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Overlays retain the equality/validation gates while admitting one forbidden
// class. Only exit 1 and the named behavioral failure count as a killed mutant;
// setup failures, compilation errors and missing tests fail this harness.
func TestCapabilityCacheNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	type mutation struct{ name, file, before, after, pkg, run, fails, message string }
	equality := "if !reflect.DeepEqual(data.Capabilities, subject.Capabilities) {"
	var mutants []mutation
	for _, field := range []struct{ name, goField string }{
		{"network", "Network"}, {"exec", "Exec"}, {"filesystem-keyword", "Filesystem.Keyword"},
		{"filesystem-paths", "Filesystem.Paths"}, {"secrets", "Secrets"}, {"env-read", "EnvRead"}, {"prompt-scope", "PromptScope"},
	} {
		mutants = append(mutants, mutation{field.name, "internal/audit/audit.go", equality,
			"recordCaps := data.Capabilities\nrecordCaps." + field.goField + " = subject.Capabilities." + field.goField + "\nif !reflect.DeepEqual(recordCaps, subject.Capabilities) {",
			"./internal/audit", "^TestVerdictCacheBindsEffectiveCapabilities$", "TestVerdictCacheBindsEffectiveCapabilities/" + field.name, "changed capabilities reused the allowance"})
	}
	for _, direction := range []struct{ name, condition string }{
		{"admit-none-crossover", "len(subject.Capabilities.Network) != 0"},
		{"admit-declared-crossover", "len(subject.Capabilities.Network) == 0"},
	} {
		for _, suite := range []struct{ name, pkg, run, fails, message string }{
			{"gate", "./internal/audit", "^TestGateEffectiveCapabilityCacheCrossover$", "TestGateEffectiveCapabilityCacheCrossover/", "before Gate: hit=true want=false"},
			{"members", "./internal/envprofile", "^TestStrictAuditMemberCapabilityCacheIsolation$", "TestStrictAuditMemberCapabilityCacheIsolation/", "after another audit"},
		} {
			message := suite.message
			if direction.name == "admit-none-crossover" && suite.name == "members" {
				message = "audit reused skill allowance"
			}
			mutants = append(mutants, mutation{direction.name + "/" + suite.name, "internal/audit/audit.go", equality,
				"if !reflect.DeepEqual(data.Capabilities, subject.Capabilities) && " + direction.condition + " {", suite.pkg, suite.run, suite.fails, message})
		}
	}
	mutants = append(mutants,
		mutation{"reuse-ruleset-two", "internal/audit/audit.go", `RulesetVersion = "3"`, `RulesetVersion = "2"`, "./internal/audit", "^TestCapabilityCacheRulesetRetainsOldRecords$", "TestCapabilityCacheRulesetRetainsOldRecords", "want block, cache hit=false"},
		mutation{"accept-null-scope", "internal/capabilities/capabilities.go", `if rawScope, present := obj["prompt_scope"]; present {`, `if rawScope, present := obj["prompt_scope"]; present && rawScope != nil {`, "./internal/skillspec", "^TestCapabilitiesNullRejectedFromV3$", "TestCapabilitiesNullRejectedFromV3/v3/agent-skill.json/null-scope", "want refusal naming capabilities.prompt_scope"},
		mutation{"accept-null-schema-three", "internal/skillspec/parse.go", `if data["capabilities"] == nil {`, `if data["capabilities"] == nil && schema != 3 {`, "./internal/skillspec", "^TestCapabilitiesNullRejectedFromV3$", "TestCapabilitiesNullRejectedFromV3/v3/agent-skill.json/null ", "want refusal naming capabilities"},
	)
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
