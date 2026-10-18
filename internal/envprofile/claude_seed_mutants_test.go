package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Overlay plants must fail the named Resolve behavioral test with exit 1 and the intended diagnostic; setup/build failures do not count as kills.
func TestClaudeSeedNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal", "envprofile", "managed.go")
	payload, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	for _, mutant := range []struct{ name, before, after, run, fails, message string }{
		{
			"whole-object-ccj", "payload, err = json.Marshal(object)", "payload, err = protocoljsonMarshal(object)",
			"^TestClaudeRepairPreservesNumberText$/^referenced$/^new-launch$",
			"TestClaudeRepairPreservesNumberText/referenced/new-launch", "CCJ-1 numbers must be shortest-form base-10 integers",
		},
		{
			"normalize-numbers", "decoder.UseNumber()", "// decoder.UseNumber() retained as a token; behavior deliberately normalizes numbers.",
			"^TestClaudeRepairPreservesNumberText$/^referenced$/^new-launch$",
			"TestClaudeRepairPreservesNumberText/referenced/new-launch", "repair changed unowned values or number texts",
		},
		{
			"admit-untrusted-approved-entry",
			"satisfied := trusted && (form != envregistry.FormReferenced || approved)",
			"satisfied := (trusted || approved) && (form != envregistry.FormReferenced || approved)",
			"^TestClaudeRepairPreservesNumberText$/^referenced$/^missing-trust$",
			"TestClaudeRepairPreservesNumberText/referenced/missing-trust", "repair changed unowned values or number texts",
		},
		{
			"admit-trusted-unapproved-entry",
			"satisfied := trusted && (form != envregistry.FormReferenced || approved)",
			"satisfied := trusted && (form != envregistry.FormReferenced || approved || trusted)",
			"^TestClaudeRepairPreservesNumberText$/^referenced$/^missing-approval$",
			"TestClaudeRepairPreservesNumberText/referenced/missing-approval", "has no external-includes approval",
		},
		{
			"rewrite-satisfied-referenced-entry", "if present && satisfied {", "if present && satisfied && form == envregistry.FormMonolithic {",
			"^TestClaudeRepairSatisfiedIsNoOp$/^referenced$",
			"TestClaudeRepairSatisfiedIsNoOp/referenced", "satisfied .claude.json was rewritten",
		},
		{
			"admit-null-object", "if err := decoder.Decode(&object); err != nil || object == nil {",
			"decodeErr := decoder.Decode(&object)\nif decodeErr == nil && object == nil { object = map[string]any{} }\nif decodeErr != nil || object == nil {",
			"^TestClaudeRepairRefusesUnsafeState$/^null$",
			"TestClaudeRepairRefusesUnsafeState/null", "want environment_seed_unreadable refusal",
		},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
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
