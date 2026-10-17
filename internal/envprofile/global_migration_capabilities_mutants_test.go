package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay changes behavior without editing the candidate. Only exit 1
// with the named behavioral failure counts as a kill; a bad anchor, compile
// error or missing test is a harness failure. The first plant restricts valid
// admission; the others admit a forbidden case while retaining nearby gates.
func TestGlobalSkillCapabilitiesNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal", "envprofile", "envprofile.go")
	payload, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	for _, mutant := range []struct{ name, before, after, run, fails, message string }{
		{
			"implicit-none", "declared = spec.Capabilities", "declared = capabilities.ImplicitNone()",
			"^TestMigratedGlobalSkillNetworkCapabilities$",
			"TestMigratedGlobalSkillNetworkCapabilities/exact-host-declared", "covering declaration refused",
		},
		{
			"malformed-network-fallback",
			"if err != nil {\n\t\t\treturn nil, fmt.Errorf(\"audit blocked: %v\", err)\n\t\t}\n\t\t// Environments",
			"if err != nil {\nif !strings.Contains(err.Error(), \"capabilities.network\") { return nil, fmt.Errorf(\"audit blocked: %v\", err) }; spec = &skillspec.Spec{Capabilities: capabilities.ImplicitNone()}\n}\n\t\t// Environments",
			"^TestMigratedGlobalSkillManifestErrorsBlock$",
			"TestMigratedGlobalSkillManifestErrorsBlock/malformed-network", "want manifest refusal",
		},
		{
			"undeclared-neighbour-admitted", "declared = spec.Capabilities",
			"declared = spec.Capabilities\nif len(declared.Network) == 1 && declared.Network[0] == \"downloads.example.com\" { declared.Network = append(declared.Network, \"uploads.example.com\") }",
			"^TestMigratedGlobalSkillDeclarationKeepsOtherGates$",
			"TestMigratedGlobalSkillDeclarationKeepsOtherGates/neighbouring-host", "want refusal carrying",
		},
		{
			"context-authorized-by-skill-manifest", "if resolved.Kind == contextlock.KindSkill {",
			"if resolved.Kind == contextlock.KindSkill || resolved.Kind == contextlock.KindContext {",
			"^TestInstalledMemberCapabilitiesStayKindScoped$",
			"TestInstalledMemberCapabilitiesStayKindScoped/context-network-not-authorized", "want network refusal",
		},
		{
			"unstated-verdict-reused",
			"packageIdentity, _ = audit.SkillPackage(resolved.Source, \"\", resolved.Name, resolved.Commit, resolved.Directory)",
			"packageIdentity, _ = audit.SkillPackage(resolved.Source, \"\", resolved.Name, resolved.Commit, resolved.Directory)\npackageIdentity = audit.SourcePackage{}",
			"^TestMigratedGlobalSkillIgnoresPreManifestVerdictCache$",
			"TestMigratedGlobalSkillIgnoresPreManifestVerdictCache/declared-skill-old-refusal", "migration verdict lost package identity",
		},
		{
			"legacy-schema-bypasses-network-finding", "if spec.SchemaVersion >= 3 {", "if spec.SchemaVersion >= 0 {",
			"^TestMigratedGlobalSkillNetworkCapabilities$",
			"TestMigratedGlobalSkillNetworkCapabilities/no-manifest", "hid the network finding behind a pin",
		},
	} {
		t.Run(mutant.name, func(t *testing.T) {
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
			command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", mutant.run, "./internal/envprofile")
			command.Dir = root
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails+" ") || !strings.Contains(string(output), mutant.message) {
				t.Fatalf("mutant survived or lacked the intended behavioral failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
