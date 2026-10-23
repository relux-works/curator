package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: these bounded source plants make admission regressions
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
		{"root-path-only", "registration.go", "TestRootReplacementBindingStillRequired", "identity(b) != r.Record.RootIdentity", "b.Repository.Root != r.Record.ProjectRoot", "replacement reused native root identity"},
		{"reserved-packaging", "types.go", "TestClosedDeclarationReader", "if _, present := raw.Items[i][\"packaging\"]; present {", "if false {", "invalid request accepted"},
		{"skill-activation-closure", "skill.go", "TestWholeSkillUnsupportedClosureRefuses", "if len(spec.Requirements) != 0 || len(spec.McpServers) != 0 || len(spec.BuildRepositories) != 0 {", "if false {", "unsupported closure was admitted"},
		{"unknown-fields", "types.go", "TestClosedDeclarationReader", "if err := exactMembers(b, reflect.TypeOf(v)); err != nil {", "if err := exactMembers(b, reflect.TypeOf(v)); false && err != nil {", "invalid request accepted"},
		{"stored-git-exclusion", "store.go", "TestProtectedSnapshotRejectsGitMetadata", "captureTree(r, \"files\", false)", "captureTree(r, \"files\", true)", "want project_context_drift"},
		{"approved-requiredness", "admission.go", "TestRequiredApprovalSurvivesPreview", "if a.Candidate != nil {", "for _, c := range op.state.Reviews[id].Review.Items { if c.Item.ID == itemID { return &c, nil } }; if a.Candidate != nil {", "two-preview required removal accepted"},
		{"required-admission", "admission.go", "TestAdmissionRetainsRequiredAuthority", "if !found {", "if false && !found {", "want project_item_unapproved"},
		{"legacy-skill-exact-case", "skill.go", "TestLegacySkillClosedMembers", "if schema != 1 {", "if schema != 0 {", "want project_context_invalid"},
		{"exact-case", "types.go", "TestExactCaseProjectRecords", "if err := exactMembers(b, reflect.TypeOf(v)); err != nil {", "if err := exactMembers(b, reflect.TypeOf(v)); false && err != nil {", "case alias accepted"},
		{"skill-semantic-owners", "skill.go", "TestSkillSemanticReview", "for _, name := range names {", "for _, name := range names[:0] {", "skill semantic review omitted command owners"},
		{"optional-unsupported", "admission.go", "TestOptionalUnsupportedSkillClosures", "if !req.Required && isCode(err, \"project_context_unsupported\") {", "if false && !req.Required && isCode(err, \"project_context_unsupported\") {", "optional unsupported closure aborted"},
		{"root-access", "registration.go", "TestExecuteOnlyRootRefuses", "st.Mode().Perm()&0500 != 0500", "st.Mode().Perm()&0500 == 0", "want project_registration_mismatch"},
		{"root-controls", "registration.go", "TestRootPathGrammar", "if r < 0x20 || r == 0x7f {", "if false && (r < 0x20 || r == 0x7f) {", "want project_registration_mismatch"},
		{"preview-baseline", "admission.go", "TestPreviewUsesApprovedBaseline", "if c, err := op.approvedCandidate(id, next.Item.ID); err != nil {", "candidate := func() (*Candidate,error) { for _, c := range op.state.Reviews[id].Review.Items { if c.Item.ID == next.Item.ID { return &c,nil } }; return nil,nil }; if c, err := candidate(); err != nil {", "preview lost approved file delta"},
		{"canonical-diagnostics", "diagnostic.go", "TestCanonicalProjectDiagnostics", "if err == nil {", "if err != nil { return err }; if err == nil {", "want project_context_unreadable"},
		{"root-lapse", "registration.go", "TestRootAuthorityRequiresExplicitRecovery", "slot.Lapsed = true", "slot.Lapsed = false", "restored root revived old authority"},
		{"root-lapse-slots", "registration.go", "TestRootLapseSlotsAndShadow", "slot.Active && slot.Record.ProjectRoot == r.Record.ProjectRoot", "slot.Active && id == r.Record.CheckoutID", "sibling slot revived root authority"},
		{"lapse-capacity", "lapses.go", "TestRootLapseAtCapacity", "return writeAtomic(op.root, \"lapses.json\", b)", "if len(lapses.CheckoutIDs) != 0 { return op.save() }; return writeAtomic(op.root, \"lapses.json\", b)", "capacity failure revived old authority"},
		{"lapse-reread", "store.go", "TestRootLapseAtCapacity", "if err := op.loadLapses(r); err != nil {", "if err := func() error { return nil }(); err != nil {", "capacity failure revived old authority"},
		{"lapse-identity", "lapses.go", "TestRetainedLapseRecordRefusesMalformed", "if !ok || !idRE.MatchString(id) || id <= previous {", "if false && (!ok || !idRE.MatchString(id) || id <= previous) {", "want project_context_drift"},
		{"lapse-canonical", "lapses.go", "TestRetainedLapseRecordRefusesMalformed", "string(can) != string(b)", "len(can) < 0", "want project_context_drift"},
		{"lapse-missing", "lapses.go", "TestRetainedLapseRecordRefusesMalformed", "if op.state.Version == 1 {", "if op.state.Version != 0 {", "want project_context_drift"},
		{"native-authentication", "context.go", "TestNativeAuthenticationNeverPublished", "if nativeAuthenticationPath(key) {", "if false && nativeAuthenticationPath(key) {", "native authentication candidate was published"},
		{"native-authentication-alias", "context.go", "TestNativeAuthenticationNeverPublished", "nativeAuthenticationPath(key)", "nativeAuthenticationPath(e.Path)", "native authentication candidate was published"},
		{"retained-native-authentication", "store.go", "TestRetainedNativeAuthenticationRefuses", "if err := validateInventory(c); err != nil {", "if err := validateInventory(c); false && err != nil {", "retained native authentication snapshot was read"},
		{"skill-before-transition", "admission.go", "TestSkillTransitionSemantics", "prior != nil && prior.Item.Type == \"skill\"", "prior != nil && prior.Item.Type == \"skill\" && c.Item.Type == \"skill\"", "skill transition omitted semantic delta"},
		{"skill-before-removal", "admission.go", "TestSkillTransitionSemantics", "change.Semantics = &SemanticChange{Before: before}", "change.Semantics = &SemanticChange{}; _ = before", "skill transition lost approved semantics"},
		{"whole-skill-document", "skill.go", "TestWholeSkillDocumentShape", "_, spec, err := skillmanifest.ReadSkillPackage(snapshot)", "_, spec, err := skillmanifest.ReadSkillPackage(snapshot); if err != nil { spec, err = skillspec.Load(snapshot) }", "invalid whole-skill document was admitted"},
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
			if tc.name == "root-access" {
				for old, replacement := range map[string]string{"f, err := os.Open(canonical)": "if false { f, err := os.Open(canonical)", "fs, err := filesystemIdentity(st)": "}; fs, err := filesystemIdentity(st)"} {
					if strings.Count(mutant, old) != 1 {
						t.Fatal("root access secondary anchor differs")
					}
					mutant = strings.Replace(mutant, old, replacement, 1)
				}
			}
			if tc.name == "unknown-fields" {
				mutant = strings.Replace(mutant, "d.DisallowUnknownFields()", "// unknown members accepted", 1)
			}
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
			cmd := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-overlay", overlay, "-count=1", "-run", "^"+tc.test+"$", ".")
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
