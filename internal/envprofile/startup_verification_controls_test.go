package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each known violation must reach its intended behavioral assertion. A source
// plant that did not apply, build/setup failure, crash or missing test cannot
// count as a detected violation. Subtests run every plant after an earlier red.
func TestStartupVerificationControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, plant := range []struct {
		name, file, before, after, run, failure, message string
	}{
		{"floor", "gitsource.go", "StartupVerificationLevel: manifest.StartupVerificationLevel", "StartupVerificationLevel: nil", "TestInstallStartupFloorBinding$/^2$", "TestInstallStartupFloorBinding/2", "weak floor: <nil>"},
		{"generation", "composition.go", "} else if !equalStrings(initialGeneration, generation) {", "} else if false && !equalStrings(initialGeneration, generation) {", "TestCompositionStartupPolicyAndGeneration$/^mid-visit-drift$/^1$", "TestCompositionStartupPolicyAndGeneration/mid-visit-drift/1", "interrupted source did not receive the intended refusal: <nil>"},
		{"capture", "materialize_capture.go", "selected.strictStartupInput = true", "selected.strictStartupInput = false", "TestCaptureStartupPolicyRemainsStrict", "TestCaptureStartupPolicyRemainsStrict/true", "capture allowed mismatch: callback=true output=true err=<nil>"},
		{"repair", "startup_verification.go", "if !req.Repair || req.Policy.startupBaseline() < 3 {", "if !req.Repair {", "TestPermissiveResolvePreservesMismatchingEvidence$/^level1$/^repairtrue$", "TestPermissiveResolvePreservesMismatchingEvidence/level1/repairtrue", "startup changed evidence pin-only.txt: <nil>"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			if plant.name == "capture" || plant.name == "repair" {
				requireLinkCapability(t)
			}
			original := filepath.Join(root, "internal", "envprofile", plant.file)
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if strings.Count(source, plant.before) != 1 {
				t.Fatal("source plant anchor must occur exactly once")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
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
			cmd := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=3m", "-v", "-run", "^"+plant.run+"$", "./internal/envprofile")
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.failure+" ") || !strings.Contains(string(output), plant.message) {
				t.Fatalf("control missed its intended behavioral failure: %v\n%s", err, output)
			}
		})
	}
}
