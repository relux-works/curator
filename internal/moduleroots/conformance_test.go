package moduleroots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModuleRootVectors runs the published module-roots family against the two
// halves of Spec §4.2.3 that this package owns, in the order the vector file
// itself declares: declaration and containment before the fixed `go list`,
// then directive form and the bijection before `go build`.
//
// A vector that fails before `go list` must be rejected without the effective
// replace set ever being read; a vector that fails before `go build` must be
// admitted by the declaration half and rejected by the replace half. Asserting
// both halves separately is what proves the failure boundary, not just the
// diagnostic.
func TestModuleRootVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "module-roots.json"))
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		EvaluationOrder []string `json:"evaluation_order"`
		Cases           []struct {
			Name        string `json:"name"`
			Declaration struct {
				BuildRoot    string   `json:"build_root"`
				BuildRoots   []string `json:"build_roots"`
				Modules      []string `json:"modules"`
				RuntimeRoots []string `json:"runtime_roots"`
			} `json:"declaration"`
			Snapshot struct {
				Directories []string          `json:"directories"`
				GoModFiles  []string          `json:"go_mod_files"`
				LinkPaths   []json.RawMessage `json:"link_paths"`
			} `json:"snapshot"`
			VendorModuleAnnotations []string `json:"vendor_module_annotations"`
			ExpectedError           string   `json:"expected_error"`
			FailsBefore             string   `json:"fails_before"`
			BuildPermitted          bool     `json:"build_permitted"`
			GoListStarted           bool     `json:"go_list_started"`
			GoBuildStarted          bool     `json:"go_build_started"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(payload, &suite); err != nil {
		t.Fatal(err)
	}
	if len(suite.Cases) == 0 {
		t.Fatal("the module-roots vector family published no cases")
	}
	for _, testCase := range suite.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			if len(testCase.Snapshot.LinkPaths) != 0 {
				// Silently ignoring a link fixture would report a green run for
				// a rule this test never exercised.
				t.Fatalf("vector %q declares link_paths, which this test does not materialise", testCase.Name)
			}
			snapshotRoot := t.TempDir()
			for _, directory := range testCase.Snapshot.Directories {
				if err := os.MkdirAll(filepath.Join(snapshotRoot, filepath.FromSlash(directory)), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range testCase.Snapshot.GoModFiles {
				path := filepath.Join(snapshotRoot, filepath.FromSlash(file))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("module fixture\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			declaration := testCase.Declaration
			declarationErr := ValidateDeclaration(snapshotRoot, "commands.tool.modules",
				declaration.Modules, declaration.BuildRoots, declaration.RuntimeRoots)
			if testCase.FailsBefore == "go-list" {
				if testCase.GoListStarted {
					t.Fatalf("vector %q fails before go list yet records it as started", testCase.Name)
				}
				wantCode(t, declarationErr, testCase.ExpectedError)
				return
			}
			// Every remaining vector reaches `go list`, so the declaration half
			// must have admitted it.
			wantCode(t, declarationErr, "")

			directives, replaceErr := EffectiveReplaceSet(
				[]byte(strings.Join(testCase.VendorModuleAnnotations, "\n")))
			if replaceErr == nil {
				replaceErr = ValidateBijection(declaration.BuildRoot, declaration.Modules, directives)
			}
			if testCase.FailsBefore == "go-build" {
				if testCase.GoBuildStarted || testCase.BuildPermitted {
					t.Fatalf("vector %q fails before go build yet records a started or permitted build", testCase.Name)
				}
				wantCode(t, replaceErr, testCase.ExpectedError)
				return
			}
			if testCase.FailsBefore != "" || testCase.ExpectedError != "" || !testCase.BuildPermitted {
				t.Fatalf("vector %q has a failure boundary this test does not model: before=%q error=%q",
					testCase.Name, testCase.FailsBefore, testCase.ExpectedError)
			}
			wantCode(t, replaceErr, "")
		})
	}
}
