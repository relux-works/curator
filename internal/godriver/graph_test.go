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

// TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows pins that the scan
// resolves by severity, not by which directive it reaches first. //go:generate
// is exempt inside a materialized vendor tree and //go:cgo_import_dynamic never
// is, so a scan that stopped at a generate hit would leave a cgo directive in
// any later window unread and hand the carve-out a bypass.
func TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows(t *testing.T) {
	// Each block is wider than one 64 KiB read, so a needle placed after it
	// lands in a strictly later window than one placed before it.
	pad := bytes.Repeat([]byte("// padding padding padding padding padding padding padding\n"), 2000)
	cgo := []byte("//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.B.dylib\"\n")
	generate := []byte("//go:generate go run ./gen\n")
	join := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }

	for _, testCase := range []struct {
		name    string
		payload []byte
		want    int
	}{
		{name: "generate before cgo", payload: join(generate, pad, cgo), want: directiveCgoImportDynamic},
		{name: "cgo before generate", payload: join(cgo, pad, generate), want: directiveCgoImportDynamic},
		{name: "generate only", payload: join(generate, pad, pad), want: directiveGenerate},
		{name: "generate in a later window", payload: join(pad, generate), want: directiveGenerate},
		{name: "neither", payload: join(pad, pad), want: directiveNone},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main.go")
			writeTestFile(t, path, append([]byte("package board\n"), testCase.payload...), 0o644)
			matched, err := scanSourceDirectives(path)
			if err != nil {
				t.Fatalf("scan: %v", err)
			}
			if matched != testCase.want {
				t.Fatalf("scan = %d, want %d", matched, testCase.want)
			}
		})
	}
}
