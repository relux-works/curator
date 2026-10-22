package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: these bounded source plants make the two required narrowing mutants
// pass a forbidden operation, and the corresponding regression witness fails
// for its intended assertion. A setup/compile failure is not a caught mutant.
// Opt in explicitly; the ordinary suite keeps only production behavior tests.
func TestAdmissionNarrowingMutants(t *testing.T) {
	if os.Getenv("CURATOR_PROJECT_ADMISSION_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, test, old, new, witness string }{
		{"root-only", "admission.go", "TestClosureSnapshotsRejectTamper", "for _, m := range lock.Members {", "for _, m := range lock.Members {\n if m.Name != lock.Root { continue }", "want project_context_drift"},
		{"cross-checkout-hash", "admission.go", "TestSameHashOtherCheckoutRefuses", "a, ok := op.state.Approvals[id+\":\"+c.Item.ID]", "a, ok := op.state.Approvals[id+\":\"+c.Item.ID]\n if !ok { for _, other := range op.state.Approvals { if other.Record.ItemSHA256 == c.ItemSHA256 { a,ok=other,true; break } } }", "want project_item_unapproved"},
		{"root-path-only", "registration.go", "TestRegistrationBindingAndNestedCWD", "identity(b) != r.Record.RootIdentity", "b.Repository.Root != r.Record.ProjectRoot", "want project_registration_mismatch"},
		{"unknown-fields", "types.go", "TestClosedDeclarationReader", "d.DisallowUnknownFields()", "// unknown members incorrectly accepted", "invalid request accepted"},
		{"stored-git-exclusion", "store.go", "TestProtectedSnapshotRejectsGitMetadata", "captureTree(r, \"files\", false)", "captureTree(r, \"files\", true)", "want project_context_drift"},
		{"hardlink-read", "files.go", "TestHardlinkCandidateRefuses", " || !singleLink(before)", "", "want project_context_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(cwd, tc.file)
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), tc.old) != 1 {
				t.Fatalf("plant anchor count differs: %s", tc.name)
			}
			mutant := strings.Replace(string(source), tc.old, tc.new, 1)
			if tc.name == "cross-checkout-hash" {
				old := " || a.Record.CheckoutID != r.Record.CheckoutID"
				if strings.Count(mutant, old) != 1 {
					t.Fatal("checkout guard anchor differs")
				}
				mutant = strings.Replace(mutant, old, "", 1)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "admission.go")
			if err := os.WriteFile(replacement, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			b, _ := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err := os.WriteFile(overlay, b, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-overlay", overlay, "-count=1", "-run", "^"+tc.test+"$", ".")
			cmd.Dir = cwd
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "GIT_DIR=") && !strings.HasPrefix(e, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "--- FAIL: "+tc.test) || !strings.Contains(string(out), tc.witness) || strings.Contains(string(out), "[build failed]") {
				t.Fatalf("mutant did not hit intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended forbidden-operation witness", tc.name)
		})
	}
}
