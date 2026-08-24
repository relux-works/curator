package godriver

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectiveScanFindsExactTokenAcrossReadBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.go")
	payload := append(bytes.Repeat([]byte{'x'}, 64*1024-5), forbiddenCompilerDirective...)
	writeTestFile(t, path, payload, 0o644)
	matched, err := scanSourceDirectives(path)
	if err != nil || matched != 1 {
		t.Fatalf("scan = %d, %v", matched, err)
	}
}

func TestPackageGraphValidatesStandardMainModuleAndVendoredTransitiveInputs(t *testing.T) {
	fixture := newSnapshotFixture(t)
	vendorDir := filepath.Join(fixture.buildRoot, "vendor", "example.test", "dep", "value")
	writeTestFile(t, filepath.Join(vendorDir, "value.go"), []byte("package value\nconst V = 1\n"), 0o644)
	writeTestFile(t, filepath.Join(vendorDir, "data.txt"), []byte("data"), 0o644)
	writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"), []byte("# example.test/dep v1.0.0\n## explicit; go 1.23\nexample.test/dep/value\n"), 0o644)
	writeTestFile(t, filepath.Join(fixture.sourceDir, "message.txt"), []byte("message"), 0o644)
	goroot := mustPhysical(t, t.TempDir())
	standardDir := filepath.Join(goroot, "src", "fmt")
	writeTestFile(t, filepath.Join(standardDir, "print.go"), []byte("package fmt\n"), 0o644)

	standard := packageJSON{Dir: standardDir, ImportPath: "fmt", Name: "fmt", Root: goroot, Standard: true, Goroot: true, DepOnly: true, GoFiles: []string{"print.go"}}
	vendored := packageJSON{
		Dir: vendorDir, ImportPath: "example.test/dep/value", Name: "value", DepOnly: true,
		GoFiles: []string{"value.go"}, EmbedFiles: []string{"data.txt"}, Module: &moduleJSON{Path: "example.test/dep", Version: "v1.0.0", GoVersion: "1.23"},
	}
	root := fixture.rootPackage()
	root.EmbedFiles = []string{"message.txt"}
	validation := graphValidation{BuildRoot: fixture.buildRoot, SourceDir: fixture.sourceDir, GOROOT: goroot}
	if err := validatePackageGraph(encodePackages(t, standard, vendored, root), validation); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name, code string
		mutate     func(*packageJSON, *packageJSON)
	}{
		{name: "standard escape", code: "go_standard_input_escape", mutate: func(std, _ *packageJSON) { std.GoFiles = []string{"../../../../outside.go"} }},
		{name: "transitive syso", code: "go_syso_forbidden", mutate: func(_, dep *packageJSON) { dep.SysoFiles = []string{"host.syso"} }},
		{name: "transitive native", code: "go_native_input_forbidden", mutate: func(_, dep *packageJSON) { dep.CXXFiles = []string{"host.cc"} }},
		{name: "transitive assembly", code: "go_assembly_forbidden", mutate: func(_, dep *packageJSON) { dep.SFiles = []string{"host.s"} }},
		{name: "transitive embed escape", code: "go_embed_input_escape", mutate: func(_, dep *packageJSON) { dep.EmbedFiles = []string{"../../../../../../outside"} }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdCopy, depCopy := standard, vendored
			testCase.mutate(&stdCopy, &depCopy)
			err := validatePackageGraph(encodePackages(t, stdCopy, depCopy, root), validation)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
		})
	}
}

func TestPackageGraphRejectsMalformedMissingAndDuplicateResults(t *testing.T) {
	fixture := newSnapshotFixture(t)
	validation := graphValidation{BuildRoot: fixture.buildRoot, SourceDir: fixture.sourceDir, GOROOT: t.TempDir()}
	root := fixture.rootPackage()
	for _, testCase := range []struct {
		name, code string
		payload    []byte
	}{
		{name: "empty", code: "go_list_incomplete", payload: nil},
		{name: "malformed", code: "go_list_malformed", payload: []byte(`{"ImportPath":`)},
		{name: "duplicate", code: "go_list_incomplete", payload: encodePackages(t, root, root)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if code := DiagnosticCode(validatePackageGraph(testCase.payload, validation)); code != testCase.code {
				t.Fatalf("code = %s, want %s", code, testCase.code)
			}
		})
	}
}

func TestVerifyArtifactRejectsHardLinks(t *testing.T) {
	base := t.TempDir()
	stage := filepath.Join(base, "stage")
	artifact := filepath.Join(stage, "bin", "tool")
	writeTestFile(t, artifact, []byte("artifact"), 0o600)
	if err := os.Link(artifact, filepath.Join(base, "outside-link")); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}
	_, err := verifyArtifact(stage, artifact, "bin/tool", 1024)
	if code := DiagnosticCode(err); code != "artifact_link" {
		t.Fatalf("error = %v", err)
	}
}

// TestInertGeneratorDirectiveIsExemptOnlyInsideTheVendorTree pins the profile's
// `//go:generate` rule: the driver runs a fixed `go list`/`go build` vector and
// never `go generate`, so the comment is inert and its presence in vendored
// `GoFiles` does not fail preflight. The exemption is scoped by location alone,
// so every case below that is not inside the vendor tree must still be refused
// by validatePackageGraph, the preflight the build calls before `go build`.
func TestInertGeneratorDirectiveIsExemptOnlyInsideTheVendorTree(t *testing.T) {
	// Real vendored dependencies carry the directive unguarded --
	// github.com/clipperhouse/displaywidth reaches every bubbletea consumer
	// through charmbracelet/x/ansi with a bare `//go:generate` in gen.go --
	// and a package cannot reasonably fork them.
	const generator = "package value\nconst V = 1\n\n//go:generate go run ./internal/gen\n"

	t.Run("audited vendored dependency is admitted", func(t *testing.T) {
		validation, vendored, root := generatorGraphFixture(t)
		writeTestFile(t, filepath.Join(vendored.Dir, "value.go"), []byte(generator), 0o644)
		if err := validatePackageGraph(encodePackages(t, vendored, root), validation); err != nil {
			t.Fatalf("an inert generator comment in a vendored GoFile failed preflight: %v", err)
		}
	})

	t.Run("main package in the build root is refused", func(t *testing.T) {
		validation, vendored, root := generatorGraphFixture(t)
		writeTestFile(t, filepath.Join(root.Dir, "main.go"),
			[]byte("package main\n\n//go:generate go run ./internal/gen\nfunc main() {}\n"), 0o644)
		err := validatePackageGraph(encodePackages(t, vendored, root), validation)
		if code := DiagnosticCode(err); code != "go_generator_forbidden" {
			t.Fatalf("code = %q, want go_generator_forbidden (error %v)", code, err)
		}
	})

	t.Run("first-party package below the build root is refused", func(t *testing.T) {
		validation, vendored, root := generatorGraphFixture(t)
		firstParty := filepath.Join(validation.BuildRoot, "internal", "value")
		writeTestFile(t, filepath.Join(firstParty, "value.go"), []byte(generator), 0o644)
		own := packageJSON{
			Dir: firstParty, ImportPath: "example.test/build/internal/value", Name: "value", DepOnly: true,
			GoFiles: []string{"value.go"}, Module: root.Module,
		}
		err := validatePackageGraph(encodePackages(t, vendored, own, root), validation)
		if code := DiagnosticCode(err); code != "go_generator_forbidden" {
			t.Fatalf("code = %q, want go_generator_forbidden (error %v)", code, err)
		}
	})

	// The exemption covers exactly one directive. `//go:cgo_import_dynamic` is
	// scoped by an import-path allowlist instead, so the same vendored package
	// must still be refused for it.
	t.Run("the exemption does not extend to cgo_import_dynamic", func(t *testing.T) {
		validation, vendored, root := generatorGraphFixture(t)
		writeTestFile(t, filepath.Join(vendored.Dir, "value.go"),
			[]byte("package value\n\n//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.dylib\"\n"), 0o644)
		err := validatePackageGraph(encodePackages(t, vendored, root), validation)
		if code := DiagnosticCode(err); code != "go_forbidden_compiler_directive" {
			t.Fatalf("code = %q, want go_forbidden_compiler_directive (error %v)", code, err)
		}
	})
}

// generatorGraphFixture returns a snapshot whose single vendored dependency is
// ready for a directive to be written into it.
func generatorGraphFixture(t *testing.T) (graphValidation, packageJSON, packageJSON) {
	t.Helper()
	fixture := newSnapshotFixture(t)
	vendorDir := filepath.Join(fixture.buildRoot, "vendor", "example.test", "dep", "value")
	writeTestFile(t, filepath.Join(vendorDir, "value.go"), []byte("package value\nconst V = 1\n"), 0o644)
	writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"),
		[]byte("# example.test/dep v1.0.0\n## explicit; go 1.23\nexample.test/dep/value\n"), 0o644)
	vendored := packageJSON{
		Dir: vendorDir, ImportPath: "example.test/dep/value", Name: "value", DepOnly: true,
		GoFiles: []string{"value.go"},
		Module:  &moduleJSON{Path: "example.test/dep", Version: "v1.0.0", GoVersion: "1.23"},
	}
	validation := graphValidation{
		BuildRoot: fixture.buildRoot, SourceDir: fixture.sourceDir,
		GOROOT: mustPhysical(t, t.TempDir()),
	}
	return validation, vendored, fixture.rootPackage()
}
