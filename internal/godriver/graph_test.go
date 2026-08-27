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

// TestPackageGraphExemptsGeneratorDirectiveOnlyBelowTheVendorTree pins the
// decision 0005 relaxation and its bound. A materialized vendor tree may carry a
// bare //go:generate that no released upstream version drops (bubbletea reaches
// clipperhouse/displaywidth through charmbracelet/x/ansi), and the directive is
// inert because the manager never runs generators. The carve-out is bounded to
// that tree: the build root's own code is still rejected, a first-party
// main-module package parked below the vendor directory is still rejected, and
// the directive is not a hole for //go:cgo_import_dynamic.
func TestPackageGraphExemptsGeneratorDirectiveOnlyBelowTheVendorTree(t *testing.T) {
	const generator = "//go:generate stringer -type=Value\n"
	fixture := newSnapshotFixture(t)
	vendorDir := filepath.Join(fixture.buildRoot, "vendor", "example.test", "dep", "value")
	writeTestFile(t, filepath.Join(vendorDir, "value.go"), []byte("package value\n"+generator+"const V = 1\n"), 0o644)
	writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"),
		[]byte("# example.test/dep v1.0.0\n## explicit; go 1.23\nexample.test/dep/value\n"), 0o644)
	goroot := mustPhysical(t, t.TempDir())

	root := fixture.rootPackage()
	vendored := packageJSON{
		Dir: vendorDir, ImportPath: "example.test/dep/value", Name: "value", DepOnly: true,
		GoFiles: []string{"value.go"},
		Module:  &moduleJSON{Path: "example.test/dep", Version: "v1.0.0", GoVersion: "1.23"},
	}
	validation := graphValidation{BuildRoot: fixture.buildRoot, SourceDir: fixture.sourceDir, GOROOT: goroot}
	if err := validatePackageGraph(encodePackages(t, vendored, root), validation); err != nil {
		t.Fatalf("vendored generator directive must not fail preflight: %v", err)
	}

	// A first-party package that sits below the vendor directory: the module is
	// the main module, so the path prefix alone must not buy the exemption.
	parkedDir := filepath.Join(fixture.buildRoot, "vendor", "example.test", "build", "internal", "gen")
	writeTestFile(t, filepath.Join(parkedDir, "gen.go"), []byte("package gen\n"+generator+"const G = 1\n"), 0o644)
	parked := packageJSON{
		Dir: parkedDir, ImportPath: "example.test/build/internal/gen", Name: "gen", DepOnly: true,
		GoFiles: []string{"gen.go"}, Module: root.Module,
	}

	// Each case restores both scanned files first, so no case can pass because an
	// earlier one left a rejectable byte behind.
	cleanRoot := []byte("package main\nfunc main() {}\n")
	cleanVendored := []byte("package value\n" + generator + "const V = 1\n")
	for _, testCase := range []struct {
		name, code string
		items      []packageJSON
		setup      func()
	}{
		{
			name: "build root package", code: "go_generator_forbidden",
			items: []packageJSON{vendored, root},
			setup: func() {
				writeTestFile(t, filepath.Join(fixture.sourceDir, "main.go"),
					[]byte("package main\n"+generator+"func main() {}\n"), 0o644)
			},
		},
		{
			name: "first-party package below the vendor tree", code: "go_generator_forbidden",
			items: []packageJSON{vendored, parked, root},
		},
		{
			name: "vendored cgo_import_dynamic", code: "go_forbidden_compiler_directive",
			items: []packageJSON{vendored, root},
			setup: func() {
				writeTestFile(t, filepath.Join(vendorDir, "value.go"),
					[]byte("package value\n//go:cgo_import_dynamic x y z\nconst V = 1\n"), 0o644)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			writeTestFile(t, filepath.Join(fixture.sourceDir, "main.go"), cleanRoot, 0o644)
			writeTestFile(t, filepath.Join(vendorDir, "value.go"), cleanVendored, 0o644)
			if testCase.setup != nil {
				testCase.setup()
			}
			err := validatePackageGraph(encodePackages(t, testCase.items...), validation)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
		})
	}
}
