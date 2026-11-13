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
	// Removing only localErr handling is not a violation: the earlier
	// structural scan independently refuses. Plant the shared snapshot
	// predicate instead so every redundant caller sees the intended defect.
	for _, plant := range []struct {
		name, file, before, after, run, failure, message, positive string
	}{
		{"floor", "gitsource.go", "StartupVerificationLevel: manifest.StartupVerificationLevel", "StartupVerificationLevel: nil", "TestInstallStartupFloorBinding$/^2$", "TestInstallStartupFloorBinding/2", "weak floor: <nil>", ""},
		{"generation", "composition.go", "} else if !equalStrings(initialGeneration, generation) {", "} else if false && !equalStrings(initialGeneration, generation) {", "TestCompositionStartupPolicyAndGeneration$/^mid-visit-drift$/^1$", "TestCompositionStartupPolicyAndGeneration/mid-visit-drift/1", "interrupted source did not receive the intended refusal: <nil>", ""},
		{"capture", "materialize_capture.go", "selected.strictStartupInput = true", "selected.strictStartupInput = false", "TestCaptureStartupPolicyRemainsStrict", "TestCaptureStartupPolicyRemainsStrict/true", "capture allowed mismatch: callback=true output=true err=<nil>", ""},
		{"repair", "startup_verification.go", "if !req.Repair || req.Policy.startupBaseline() < 3 {", "if !req.Repair {", "TestPermissiveResolvePreservesMismatchingEvidence$/^level1$/^repairtrue$", "TestPermissiveResolvePreservesMismatchingEvidence/level1/repairtrue", "startup changed evidence pin-only.txt: <nil>", ""},
		{"boundary-repair", "managed.go", "\tabsentStore, err := inspectStoreAvailability(req.Home, lock, req.boundaryOwnerLookup)\n\tif err != nil {\n\t\treturn nil, err\n\t}", "\tabsentStore, err := inspectStoreAvailability(req.Home, lock, req.boundaryOwnerLookup)\n\tif err != nil {\n\t\tif req.Repair {\n\t\t\tvar failure *storeEntryFailure\n\t\t\tif errors.As(err, &failure) { _ = rebuildUntrustedStoreEntry(&req, hash, failure) }\n\t\t}\n\t\treturn nil, err\n\t}", "(TestStartupBoundaryNeverBecomesWarning|TestStrictStartupRepairRestoresRegularBytes)", "TestStartupBoundaryNeverBecomesWarning/symlink/repair/level1", "fatal boundary reached repair-lock seam", "TestStrictStartupRepairRestoresRegularBytes"},
		{"boundary-snapshot-link", "../pathboundary/pathboundary.go", "return validateWithinWithRegularCheck(root, target, ownerLookup, defaultEntryInfo, singleLink)", "return validateWithinWithRegularCheck(root, target, ownerLookup, defaultEntryInfo, nil)", "(TestStartupNonSkillHardlinkBoundary|TestStrictStartupRepairRestoresRegularBytes)", "TestStartupNonSkillHardlinkBoundary/context/observe/level1", "expected named single_link refusal, output=true err=<nil>", "TestStrictStartupRepairRestoresRegularBytes"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			if plant.name == "capture" || plant.name == "repair" || strings.HasPrefix(plant.name, "boundary-") {
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
			cmd := exec.Command("go", "test", "-p=1", "-parallel=1", "-overlay", overlayPath, "-count=1", "-timeout=3m", "-v", "-run", "^"+plant.run+"$", "./internal/envprofile")
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.failure+" ") || !strings.Contains(string(output), plant.message) {
				t.Fatalf("control missed its intended behavioral failure: %v\n%s", err, output)
			}
			if plant.positive != "" {
				for _, level := range []string{"level3", "level4"} {
					if !strings.Contains(string(output), "--- PASS: "+plant.positive+"/"+level+" ") {
						t.Fatalf("control lost its valid same-pin repair neighbor: %s\n%s", level, output)
					}
				}
			}
		})
	}
}
