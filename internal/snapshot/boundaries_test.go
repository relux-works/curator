package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

func writeSkill(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: review\ndescription: A skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func boundaryOutputs(project string) []string {
	return []string{filepath.Join(project, ".agents")}
}

func TestValidateBroadRootAllowed(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, "agents", "skills", "review"))
	physical, admitted, err := ValidateLocalPackage(project, "agents/skills/review", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err != nil {
		t.Fatalf("broad root with safe subdirectory: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(physical), "agents/skills/review") {
		t.Fatalf("physical = %s", physical)
	}
	if len(admitted) != 1 || admitted[0] != physical {
		t.Fatalf("admitted = %v, want [%s]", admitted, physical)
	}
}

func TestValidateManagedSourceRejected(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, ".agents", "skills", "review"))
	_, _, err := ValidateLocalPackage(project, ".agents/skills/review", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("managed source err = %v, want source_output_overlap", err)
	}
}

func TestValidateManagedMissingRefusedAsOverlapBeforeTraversal(t *testing.T) {
	// Isolate the pre-traversal output gate: a missing selector under a
	// managed root is refused as overlap, not as a missing member. The
	// output gate fires before any existence-dependent traversal, so
	// narrowing the first `else if pruned` to `directory == "."` falls
	// through to the missing-member error and this test fails — while
	// the direct managed-source test survives that mutant through the
	// post-resolution gate.
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err := ValidateLocalPackage(project, ".agents/absent", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("missing under managed err = %v, want source_output_overlap before traversal", err)
	}
}

func TestValidateManagedSpellingResolvingOutsideAdmitted(t *testing.T) {
	// Resolve-then-check semantics (spec section 1): a managed spelling
	// that resolves to authored bytes admits those bytes — containment
	// judges the physical package, not the spelling. Only authored bytes
	// enter; the managed tree itself is never traversed for capture.
	project := t.TempDir()
	good := filepath.Join(project, "agents", "skills", "good")
	writeSkill(t, good)
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(project, ".agents", "link")
	if err := os.Symlink(good, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	physical, admitted, err := ValidateLocalPackage(project, ".agents/link", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err != nil {
		t.Fatalf("managed spelling resolving to authored bytes: %v", err)
	}
	if len(admitted) != 1 || admitted[0] != physical {
		t.Fatalf("admitted = %v, want [%s]", admitted, physical)
	}
}

func TestPrepareLocalAcquisitionAdmitsSafeSubdirectory(t *testing.T) {
	// Path source "." with a safe selected subdirectory remains valid at
	// the production acquisition entry point.
	project := t.TempDir()
	pkg := filepath.Join(project, "agents", "skills", "review")
	writeSkill(t, pkg)
	if err := os.WriteFile(filepath.Join(pkg, "notes.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	acquisition, err := PrepareLocalAcquisition(project, "agents/skills/review", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	defer func() { _ = acquisition.Close() }()
	if len(acquisition.Admitted) != 1 || acquisition.Admitted[0] != acquisition.Physical {
		t.Fatalf("admitted = %v, want [%s]", acquisition.Admitted, acquisition.Physical)
	}
	if len(acquisition.Files) != 2 {
		t.Fatalf("files = %v, want SKILL.md and notes.md", acquisition.Files)
	}
	info, err := os.Stat(acquisition.Staging)
	if err != nil || !info.IsDir() {
		t.Fatalf("staging = %s: %v", acquisition.Staging, err)
	}
	if err := privatedir.Validate(acquisition.Staging); err != nil {
		t.Fatalf("staging is not private: %v", err)
	}
	staging := acquisition.Staging
	if err := acquisition.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Lstat(staging); !os.IsNotExist(err) {
		t.Fatalf("Close left staging behind: %v", err)
	}
}

func TestPrepareLocalAcquisitionAdmitsRootInputs(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, project)
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	acquisition, err := PrepareLocalAcquisition(project, ".", project, "s", boundaryOutputs(project), map[string][]string{"s": {"SKILL.md"}}, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("root package with root_inputs: %v", err)
	}
	defer func() { _ = acquisition.Close() }()
	if len(acquisition.Admitted) != 1 || len(acquisition.Files) != 1 {
		t.Fatalf("admitted = %v files = %v, want the SKILL.md selection", acquisition.Admitted, acquisition.Files)
	}
}

func TestPrepareLocalAcquisitionRefusals(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		project := t.TempDir()
		writeSkill(t, project)
		writeSkill(t, filepath.Join(project, ".agents", "skills", "review"))
		if err := os.MkdirAll(filepath.Join(project, "agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		return project
	}
	cases := []struct {
		name      string
		directory string
		alias     string
		inputs    map[string][]string
		known     map[string]bool
		class     string
	}{
		{"managed-source", ".agents/skills/review", "s", nil, map[string]bool{"s": true}, "source_output_overlap"},
		{"root-without-inputs", ".", "s", nil, map[string]bool{"s": true}, "source_output_overlap"},
		{"unknown-alias", "agents", "absent", nil, map[string]bool{"s": true}, "source_alias_unknown"},
		{"non-portable", "../escape", "s", nil, map[string]bool{"s": true}, "source_selection_invalid"},
		{"missing-member", "agents/absent", "s", nil, map[string]bool{"s": true}, "source_member_missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := setup(t)
			_, err := PrepareLocalAcquisition(project, tc.directory, project, tc.alias, boundaryOutputs(project), tc.inputs, tc.known, "")
			if err == nil || !strings.Contains(err.Error(), tc.class) {
				t.Fatalf("%s err = %v, want %s", tc.name, err, tc.class)
			}
		})
	}
}

func TestPrepareLocalAcquisitionValidatesBeforeStaging(t *testing.T) {
	// Order proof: with both an invalid package and an unusable staging
	// parent, the validation error — not a staging failure — must win,
	// proving overlap/root_inputs/boundary validation runs before any
	// staging side effect.
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, ".agents", "skills", "review"))
	bogusParent := filepath.Join(project, "no-such-parent")
	_, err := PrepareLocalAcquisition(project, ".agents/skills/review", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true}, bogusParent)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("err = %v, want the pre-staging source_output_overlap refusal", err)
	}
	if _, statErr := os.Lstat(bogusParent); !os.IsNotExist(statErr) {
		t.Fatalf("refused acquisition created staging state: %v", statErr)
	}
}

func TestValidateSymlinkManagedAndEscape(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, "skills", "good"))
	managed := filepath.Join(project, ".agents", "skills", "output")
	writeSkill(t, managed)
	link := filepath.Join(project, "skills", "link")
	if err := os.Symlink(managed, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, _, err := ValidateLocalPackage(project, "skills/link", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("symlink to managed err = %v, want source_output_overlap", err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	writeSkill(t, filepath.Join(outside, "pkg"))
	if err := os.Symlink(filepath.Join(outside, "pkg"), link); err != nil {
		t.Fatal(err)
	}
	_, _, err = ValidateLocalPackage(project, "skills/link", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
		t.Fatalf("escaping symlink err = %v, want source_selection_invalid", err)
	}
}

func TestValidateCaseAliasUsesFilesystemIdentity(t *testing.T) {
	// Prove the gate follows SameFile identity rather than string prefixes:
	// a symlink alias is refused through the same Within path a
	// case-variant spelling takes on an insensitive volume.
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, "skills", "good"))
	managed := filepath.Join(project, ".agents")
	if err := os.MkdirAll(managed, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(project, "skills", "alias")
	if err := os.Symlink(managed, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, _, err := ValidateLocalPackage(project, "skills/alias/output", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil {
		t.Fatalf("filesystem alias through a link was admitted")
	}
	if !strings.Contains(err.Error(), "source_output_overlap") && !strings.Contains(err.Error(), "source_member") {
		t.Fatalf("alias err = %v, want overlap or member refusal", err)
	}
	// On a truly case-insensitive volume, exercise the .AGENTS spelling
	// from the semantic corpus directly.
	probe := filepath.Join(project, "caseprobe-lower")
	if err := os.WriteFile(probe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	lower, lowerErr := os.Stat(probe)
	upper, upperErr := os.Stat(filepath.Join(project, "CASEPROBE-LOWER"))
	if lowerErr != nil || upperErr != nil {
		t.Skip("test filesystem is case-sensitive; symlink alias above covers the SameFile path")
	}
	same := false
	func() {
		defer func() { _ = recover() }()
		same = os.SameFile(lower, upper)
	}()
	if !same {
		t.Skip("test filesystem is case-sensitive; symlink alias above covers the SameFile path")
	}
	writeSkill(t, filepath.Join(project, ".agents", "skills", "review"))
	_, _, err = ValidateLocalPackage(project, ".AGENTS/skills/review", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("case alias err = %v, want source_output_overlap", err)
	}
}

func TestValidateRootRequiresInputs(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, project)
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err := ValidateLocalPackage(project, ".", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("root without inputs err = %v, want source_output_overlap", err)
	}
	physical, admitted, err := ValidateLocalPackage(project, ".", project, "s", boundaryOutputs(project), map[string][]string{"s": {"SKILL.md"}}, map[string]bool{"s": true})
	if err != nil {
		t.Fatalf("root with inputs: %v", err)
	}
	if physical == "" || len(admitted) != 1 {
		t.Fatalf("root admitted = %v", admitted)
	}
}

func TestValidateRootInputsRefusals(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		project := t.TempDir()
		writeSkill(t, project)
		if err := os.WriteFile(filepath.Join(project, "agent-skill.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(project, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(project, "scripts", "run.js"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		return project
	}
	t.Run("unknown-alias", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("absent", []string{"SKILL.md"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_alias_unknown") {
			t.Fatalf("err = %v, want source_alias_unknown", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"SKILL.md", "SKILL.md"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("err = %v, want source_selection_invalid", err)
		}
	})
	t.Run("overlap", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"scripts", "scripts/run.js"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("err = %v, want source_selection_invalid", err)
		}
	})
	t.Run("non-portable", func(t *testing.T) {
		project := setup(t)
		for _, entry := range []string{"../escape", "a\\b", ".", ""} {
			if _, err := ValidateRootInputs("s", []string{entry}, project, boundaryOutputs(project), map[string]bool{"s": true}); err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
				t.Fatalf("%q err = %v, want source_selection_invalid", entry, err)
			}
		}
	})
	t.Run("symlink", func(t *testing.T) {
		project := setup(t)
		if err := os.Symlink(filepath.Join(project, "scripts"), filepath.Join(project, "linked")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		_, err := ValidateRootInputs("s", []string{"linked"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("err = %v, want source_selection_invalid", err)
		}
	})
	t.Run("output-overlap", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"SKILL.md", ".agents"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
			t.Fatalf("err = %v, want source_output_overlap", err)
		}
	})
	t.Run("missing", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"SKILL.md", "absent"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_member_missing") {
			t.Fatalf("err = %v, want source_member_missing", err)
		}
	})
	t.Run("missing-skill-coverage", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"scripts"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
			t.Fatalf("err = %v, want source_output_overlap", err)
		}
	})
	t.Run("missing-manifest-coverage", func(t *testing.T) {
		project := setup(t)
		_, err := ValidateRootInputs("s", []string{"SKILL.md", "scripts"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
			t.Fatalf("err = %v, want source_output_overlap", err)
		}
	})
	t.Run("declared-root-coverage", func(t *testing.T) {
		project := setup(t)
		if err := os.MkdirAll(filepath.Join(project, "runtime"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(project, "agent-skill.json"), []byte(`{"schema_version":2,"runtime_roots":["runtime"]}`), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := ValidateRootInputs("s", []string{"SKILL.md", "agent-skill.json", "scripts"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
			t.Fatalf("err = %v, want source_output_overlap", err)
		}
		admitted, err := ValidateRootInputs("s", []string{"SKILL.md", "agent-skill.json", "scripts", "runtime"}, project, boundaryOutputs(project), map[string]bool{"s": true})
		if err != nil {
			t.Fatalf("valid root inputs: %v", err)
		}
		if len(admitted) != 4 {
			t.Fatalf("admitted = %v", admitted)
		}
	})
}

func TestEnumerateInputsPrunesAndRejects(t *testing.T) {
	project := t.TempDir()
	pkg := filepath.Join(project, "agents", "skills", "review")
	writeSkill(t, pkg)
	if err := os.WriteFile(filepath.Join(pkg, "notes.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pruned subtrees are skipped silently.
	if err := os.MkdirAll(filepath.Join(pkg, ".git", "objects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, ".git", "objects", "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputs := []string{filepath.Join(pkg, "output")}
	if err := os.MkdirAll(outputs[0], 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputs[0], "gen.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := EnumerateInputs(pkg, outputs, nil)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(files, "\n")
	if !strings.Contains(joined, "SKILL.md") || !strings.Contains(joined, "notes.md") {
		t.Fatalf("admitted files = %v", files)
	}
	if strings.Contains(joined, ".git") || strings.Contains(joined, "output") {
		t.Fatalf("pruned content admitted: %v", files)
	}
	// Links in admitted inputs are rejected.
	if err := os.Symlink(filepath.Join(pkg, "notes.md"), filepath.Join(pkg, "alias.md")); err == nil {
		if _, err := EnumerateInputs(pkg, outputs, nil); err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
			t.Fatalf("link err = %v, want source_member_invalid", err)
		}
		_ = os.Remove(filepath.Join(pkg, "alias.md"))
	}
	// A pruned package root is an error, not an empty admission.
	if _, err := EnumerateInputs(filepath.Join(pkg, "output"), outputs, nil); err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("pruned root err = %v, want source_output_overlap", err)
	}
}

func TestEnumerateRootEntriesSelectFileOrTree(t *testing.T) {
	project := t.TempDir()
	writeSkill(t, project)
	if err := os.MkdirAll(filepath.Join(project, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "scripts", "a.js"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "scripts", "b.js"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := EnumerateInputs(project, boundaryOutputs(project), []string{"SKILL.md", "scripts"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("root selection files = %v, want 3", files)
	}
}

func TestValidateLocalPackageIdentityReadErrorIsNotOverlap(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	writeSkill(t, filepath.Join(project, "agents", "skills", "review"))
	outputs := []string{filepath.Join(root, "managed")}
	if err := os.MkdirAll(filepath.Join(root, "managed"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ValidateLocalPackage(project, "agents/skills/review", project, "s", outputs, nil, map[string]bool{"s": true}); err != nil {
		t.Fatalf("disjoint package: %v", err)
	}
	parent := filepath.Join(root, "managed")
	if err := os.Rename(parent, parent+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("managed", parent); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, _, err := ValidateLocalPackage(project, "agents/skills/review", project, "s", outputs, nil, map[string]bool{"s": true})
	if err == nil {
		t.Fatal("expected read failure")
	}
	if strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("identity read failure mislabeled as overlap: %v", err)
	}
	if !strings.Contains(err.Error(), "boundary_identity_unreadable") {
		t.Fatalf("identity read failure err = %v, want boundary_identity_unreadable", err)
	}
}
