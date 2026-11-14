package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The landed closed policy/ordered carriers survive typed decoding, while
// missing, misplaced and unknown fields fail before any fixture is returned.
func TestStoreVerificationLandedFields(t *testing.T) {
	for _, name := range []string{"policy", "ordered"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/store-verification-v2/" + name + "-cases.json")
			if err != nil {
				t.Fatal(err)
			}
			rows, err := decodeStoreVerificationCases(raw)
			if err != nil {
				t.Fatalf("landed positive neighbor: %v", err)
			}
			var target int
			for i, row := range rows {
				if name == "policy" && row.Policy != nil {
					target = i
					if row.Policy.Global == nil || *row.Policy.Global != 1 || len(row.Policy.Evidence) == 0 || row.Expected.WarningEmitted == nil {
						t.Fatal("typed policy fields lost")
					}
				}
				if name == "ordered" && len(row.Failures) != 0 {
					target = i
					if len(row.Failures) != 2 || row.Failures[0].Path == "" || row.Expected.OffendingPath == "" {
						t.Fatal("typed ordered failure fields lost")
					}
				}
			}
			mutations := []string{"unknown", "misplaced", "missing"}
			if name == "policy" {
				mutations = append(mutations, "warning-null", "fractional-level")
			}
			for _, mutation := range mutations {
				t.Run(mutation, func(t *testing.T) {
					var carrier []map[string]any
					if err := json.Unmarshal(raw, &carrier); err != nil {
						t.Fatal(err)
					}
					c := carrier[target]
					switch mutation {
					case "unknown":
						if name == "policy" {
							c["policy"].(map[string]any)["unknown"] = true
						} else {
							c["structural_failures"].([]any)[0].(map[string]any)["unknown"] = true
						}
					case "misplaced":
						c["condition"] = "intact"
					case "missing":
						if name == "policy" {
							delete(c, "policy")
						} else {
							delete(c, "structural_failures")
						}
					case "warning-null":
						c["expected"].(map[string]any)["warning_emitted"] = nil
					case "fractional-level":
						c["policy"].(map[string]any)["global_level"] = 1.5
					}
					plant, err := json.Marshal(carrier)
					if err != nil {
						t.Fatal(err)
					}
					decoded, err := decodeStoreVerificationCases(plant)
					if err == nil || decoded != nil || !strings.HasPrefix(err.Error(), storeVerificationUnsupported+":") {
						t.Fatalf("reader accepted landed-field plant: %s rows=%d err=%v", mutation, len(decoded), err)
					}
				})
			}
		})
	}
}

// Bypassing schema validation must fail at the missing-policy assertion, not
// compilation or setup. The same invocation keeps its valid policy neighbor.
func TestStoreVerificationLandedSchemaControl(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal/envprofile/store_verification_reader_test.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	before := "if err := schema.Validate(doc); err != nil {"
	if strings.Count(string(source), before) != 1 {
		t.Fatal("schema plant anchor must occur once")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "reader_test.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(source), before, "if err := schema.Validate(doc); false && err != nil {", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-p=1", "-parallel=1", "-overlay", path, "-count=1", "-timeout=3m", "-v", "-run", "^TestStoreVerificationLandedFields$/^policy$/^missing$", "./internal/envprofile")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestStoreVerificationLandedFields/policy/missing ") || !strings.Contains(string(output), "reader accepted landed-field plant: missing") {
		t.Fatalf("schema control missed intended assertion: %v\n%s", err, output)
	}
}
