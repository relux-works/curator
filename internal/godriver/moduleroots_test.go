package godriver

import (
	"context"
	"encoding/json"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildsource"
)

// moduleRootsFixture is a frozen multi-module snapshot: one build root whose
// vendor tree carries a first-party module, and the declared directory that
// module replaces.
//
//	snapshot/
//	  pkg/board/{go.mod,board.go}
//	  scripts/run.sh                      (a runtime root)
//	  build/{go.mod,cmd/tool/main.go}
//	  build/vendor/modules.txt
//	  build/vendor/example.test/board/board.go
func newModuleRootsFixture(t *testing.T) *workerFixture {
	t.Helper()
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	writeTestFile(t, filepath.Join(snapshot, "build", "go.mod"), []byte("module example.test/build\n\ngo 1.23\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, "build", "cmd", "tool", "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, "pkg", "board", "go.mod"), []byte("module example.test/board\n\ngo 1.23\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, "pkg", "board", "board.go"), []byte("package board\n\nfunc Name() string { return \"board\" }\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, "scripts", "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	writeTestFile(t, filepath.Join(snapshot, "build", "vendor", "example.test", "board", "board.go"),
		[]byte("package board\n\nfunc Name() string { return \"board\" }\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, "build", "vendor", "modules.txt"), []byte(strings.Join([]string{
		"# example.test/board v0.0.0 => ../pkg/board",
		"## explicit; go 1.23",
		"example.test/board",
		"# example.test/board => ../pkg/board",
		"",
	}, "\n")), 0o644)
	snapshot = mustPhysical(t, snapshot)
	return &workerFixture{
		t: t, root: snapshot,
		buildRoot:    filepath.Join(snapshot, "build"),
		sourceDir:    filepath.Join(snapshot, "build", "cmd", "tool"),
		buildRootRel: "build",
		sourceDirRel: "build/cmd/tool",
		modules:      []string{"pkg/board"},
		runtimeRoots: []string{"scripts"},
	}
}

// replacedPackage is the vendored result of the declared first-party module,
// shaped exactly as `go list -mod=vendor -deps -json` reports one: the package
// directory is inside the build root's vendor tree, the module carries a
// Replace record, and Module.Dir/Module.GoMod are absent while
// Module.Replace.Dir/GoMod point outside the build root.
func (fixture *workerFixture) replacedPackage() packageJSON {
	declared := filepath.Join(fixture.root, "pkg", "board")
	return packageJSON{
		Dir:        filepath.Join(fixture.buildRoot, "vendor", "example.test", "board"),
		ImportPath: "example.test/board",
		Name:       "board",
		DepOnly:    true,
		GoFiles:    []string{"board.go"},
		Module: &moduleJSON{
			Path: "example.test/board", Version: "v0.0.0", GoVersion: "1.23",
			Replace: &moduleJSON{
				Path: "../pkg/board", Dir: declared,
				GoMod: filepath.Join(declared, "go.mod"), GoVersion: "1.23",
			},
		},
	}
}

func moduleRootLimits() ResourceLimits {
	return ResourceLimits{
		Timeout: 30 * time.Second, OutputBytes: 64 * 1024, ArtifactBytes: 1024,
		DiskBytes: 8 * 1024 * 1024, MemoryBytes: 64 * 1024 * 1024, Processes: 8,
	}
}

// TestModuleRootsBuildAMultiModuleVendoredSnapshot is the positive case of
// Spec §4.2.3: a build root that replaces a declared first-party module builds,
// the declared directory is validated against the snapshot, and the compiled
// bytes still come from the vendor tree.
func TestModuleRootsBuildAMultiModuleVendoredSnapshot(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage())),
		Artifact:   "multi module executable",
	})
	result, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if err != nil {
		t.Fatalf("declared module roots rejected a valid multi-module build: %v", err)
	}
	if result.Artifact.Metadata.Size != int64(len("multi module executable")) {
		t.Fatalf("artifact = %+v", result.Artifact.Metadata)
	}
	if calls := fixture.sourceAwareCalls(); len(calls) != 2 {
		t.Fatalf("source-aware calls = %d, want exactly one list and one build", len(calls))
	}
}

// TestModuleRootsRejectARequestWhoseDeclarationTheDriverCannotProve pins the
// driver's own re-verification of the declaration half. The parser already ran
// it; the driver is the component that starts Go, so a request whose
// declaration does not hold against the frozen snapshot must never reach
// `go list`.
func TestModuleRootsRejectARequestWhoseDeclarationTheDriverCannotProve(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		modules []string
		code    string
	}{
		{name: "absent directory", modules: []string{"pkg/missing"}, code: "build_module_root_declaration_invalid"},
		{name: "not a portable path", modules: []string{"../outside"}, code: "build_module_root_declaration_invalid"},
		{name: "dot", modules: []string{"."}, code: "build_module_root_declaration_invalid"},
		{name: "no go.mod inside", modules: []string{"scripts"}, code: "build_module_root_declaration_invalid"},
		{name: "duplicated", modules: []string{"pkg/board", "pkg/board"}, code: "build_module_root_declaration_invalid"},
		{name: "inside the build root", modules: []string{"build"}, code: "build_module_root_containment_invalid"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newModuleRootsFixture(t)
			fixture.modules = testCase.modules
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage())),
				Artifact:   "unreachable",
			})
			_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
			if code := DiagnosticCode(err); code != testCase.code {
				t.Fatalf("code = %q, want %s (error %v)", code, testCase.code, err)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 0 {
				t.Fatalf("go list ran %d source-aware commands, want the declaration half to fail before it", len(calls))
			}
		})
	}
}

// TestModuleRootsRejectAContainmentCollisionWithARuntimeRoot proves the driver
// re-runs the whole containment comparison, including the runtime roots it
// carries for no other purpose.
func TestModuleRootsRejectAContainmentCollisionWithARuntimeRoot(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	fixture.runtimeRoots = []string{"pkg"}
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "unreachable"})
	_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "build_module_root_containment_invalid" {
		t.Fatalf("code = %q, want build_module_root_containment_invalid (error %v)", code, err)
	}
	if calls := fixture.sourceAwareCalls(); len(calls) != 0 {
		t.Fatalf("go list ran %d source-aware commands, want the declaration half to fail before it", len(calls))
	}
}

// TestModuleRootsRejectAContainmentCollisionWithASiblingBuildRoot proves the
// driver's containment backstop covers EVERY declared build root and not only
// the one this command compiles, which is what §4.2.3 actually says. Both
// directions are asserted from one fixture: the identical declaration builds
// when `pkg` is not a declared build root and is refused before `go list` when
// it is, so the test cannot pass by rejecting everything.
func TestModuleRootsRejectAContainmentCollisionWithASiblingBuildRoot(t *testing.T) {
	permitted := newModuleRootsFixture(t)
	permitted.buildRootsRel = []string{"build"}
	permitted.start(stubScript{
		ListStdout: string(encodePackages(t, permitted.rootPackage(), permitted.replacedPackage())),
		Artifact:   "multi module executable",
	})
	if _, err := Build(context.Background(), permitted.request(moduleRootLimits())); err != nil {
		t.Fatalf("a declaration overlapping no declared build root was rejected: %v", err)
	}

	// `pkg` contains the declared module directory `pkg/board`, and the command
	// still compiles `build`, so only the wider declared set can catch this.
	refused := newModuleRootsFixture(t)
	refused.buildRootsRel = []string{"build", "pkg"}
	refused.start(stubScript{
		ListStdout: string(encodePackages(t, refused.rootPackage(), refused.replacedPackage())),
		Artifact:   "unreachable",
	})
	_, err := Build(context.Background(), refused.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "build_module_root_containment_invalid" {
		t.Fatalf("code = %q, want build_module_root_containment_invalid (error %v)", code, err)
	}
	if calls := refused.sourceAwareCalls(); len(calls) != 0 {
		t.Fatalf("go list ran %d source-aware commands, want the declaration half to fail before it", len(calls))
	}
}

// TestModuleRootsCheckTheBijectionAfterGoListAndBeforeGoBuild pins the second
// half of the failure boundary: every one of these rejections happens with the
// `go list` stream in hand and no `go build` started.
func TestModuleRootsCheckTheBijectionAfterGoListAndBeforeGoBuild(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		modules     []string
		annotations []string
		code        string
	}{
		{
			name: "declared directory named by no replacement", modules: []string{"pkg/board"},
			annotations: []string{}, code: "build_module_root_declaration_unused",
		},
		{
			name: "replacement naming an undeclared directory", modules: []string{},
			annotations: []string{"# example.test/board => ../pkg/board"}, code: "build_module_root_directive_undeclared",
		},
		{
			name: "replacement escaping the snapshot", modules: []string{},
			annotations: []string{"# example.test/escape => ../../../outside"}, code: "build_module_root_directive_undeclared",
		},
		{
			name: "module to module redirect", modules: []string{},
			annotations: []string{"# example.test/board => example.test/fork v1.2.3"},
			code:        "build_module_root_directive_form_unsupported",
		},
		{
			name: "versioned left side", modules: []string{"pkg/board"},
			annotations: []string{"# example.test/board v1.2.3 => ../pkg/board"},
			code:        "build_module_root_directive_form_unsupported",
		},
		{
			name: "two replacements onto one declaration", modules: []string{"pkg/board"},
			annotations: []string{"# example.test/board => ../pkg/board", "# example.test/fork => ../pkg/board"},
			code:        "build_module_root_directive_undeclared",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newModuleRootsFixture(t)
			fixture.modules = testCase.modules
			writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"),
				[]byte(strings.Join(testCase.annotations, "\n")+"\n"), 0o644)
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage())),
				Artifact:   "unreachable",
			})
			_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
			if code := DiagnosticCode(err); code != testCase.code {
				t.Fatalf("code = %q, want %s (error %v)", code, testCase.code, err)
			}
			calls := fixture.sourceAwareCalls()
			if len(calls) != 1 || calls[0].Argv[0] != "list" {
				t.Fatalf("source-aware calls = %+v, want exactly the fixed go list", calls)
			}
		})
	}
}

// TestReplacementsAreRejectedWithoutADeclaration is the rule §4.2.3 states for
// an absent or empty modules list: the effective replace set must be empty, so
// a schema-6 or schema-7 command keeps its single-module build root and an
// unused directive cannot hide by going unused.
func TestReplacementsAreRejectedWithoutADeclaration(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	fixture.modules = nil
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage())),
		Artifact:   "unreachable",
	})
	_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "build_module_root_directive_undeclared" {
		t.Fatalf("code = %q, want build_module_root_directive_undeclared (error %v)", code, err)
	}
}

// TestDeclaredModuleRootsRequireVendorMetadata: the effective replace set is
// defined by the bytes of vendor/modules.txt, so its absence under a declared
// list is inconsistent vendor metadata rather than an empty replace set.
func TestDeclaredModuleRootsRequireVendorMetadata(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	if err := os.Remove(filepath.Join(fixture.buildRoot, "vendor", "modules.txt")); err != nil {
		t.Fatal(err)
	}
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "unreachable"})
	_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "vendor_metadata_inconsistent" {
		t.Fatalf("code = %q, want vendor_metadata_inconsistent (error %v)", code, err)
	}
}

// TestVendorMetadataMustBeARegularFile pins the other half of that rule: the
// replace set is read from the BYTES of vendor/modules.txt, so anything
// standing in that path which is not a plain file is inconsistent metadata,
// never an empty set. `os.Lstat` does not follow the final component and
// `IsRegular` is false for every non-regular mode bit, so one predicate
// decides every shape.
func TestVendorMetadataMustBeARegularFile(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	path := filepath.Join(fixture.buildRoot, "vendor", "modules.txt")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "unreachable"})
	_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "vendor_metadata_inconsistent" {
		t.Fatalf("code = %q, want vendor_metadata_inconsistent (error %v)", code, err)
	}
	if calls := fixture.sourceAwareCalls(); len(calls) != 1 || calls[0].Argv[0] != "list" {
		t.Fatalf("source-aware calls = %+v, want exactly one list and no build", calls)
	}
}

// TestALinkStandingInForVendorMetadataNeverReachesTheDriver records WHY the
// regular-file predicate above needs no separate symlink term: the frozen
// build source is validated link-free before a session exists, so a link at
// `vendor/modules.txt` is refused a whole layer earlier and the driver's own
// check can never observe one. Asserting the earlier boundary is what keeps
// that reasoning honest -- if the snapshot ever started tolerating the link,
// this fails rather than the removed term silently mattering again.
func TestALinkStandingInForVendorMetadataNeverReachesTheDriver(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	path := filepath.Join(fixture.buildRoot, "vendor", "modules.txt")
	target := filepath.Join(t.TempDir(), "modules.txt")
	writeTestFile(t, target, []byte("# example.test/board => ../pkg/board\n"), 0o644)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		// Reported, not skipped: a skip would add a new class to the ledger for
		// a rule the directory case already covers on every runner.
		t.Logf("this host cannot create a symlink, so the snapshot boundary is unexercised here: %v", err)
		return
	}
	if _, err := buildsource.Validate(fixture.root); err == nil {
		t.Fatal("the frozen build source accepted a link at vendor/modules.txt")
	} else if !strings.Contains(err.Error(), "link forbidden") {
		t.Fatalf("build source error = %v, want a link refusal naming the path", err)
	}
}

// TestAReplacedResultOutsideTheAdmittedSetIsRejected is the driver's
// fail-closed backstop: `go list` reporting a replacement that
// vendor/modules.txt does not materialize means the two disagree, and no
// package result may be admitted on the strength of its own Replace record.
func TestAReplacedResultOutsideTheAdmittedSetIsRejected(t *testing.T) {
	fixture := newModuleRootsFixture(t)
	rogue := fixture.replacedPackage()
	rogue.ImportPath = "example.test/rogue"
	rogue.Dir = filepath.Join(fixture.buildRoot, "vendor", "example.test", "rogue")
	rogue.Module.Path = "example.test/rogue"
	writeTestFile(t, filepath.Join(rogue.Dir, "board.go"), []byte("package board\n"), 0o644)
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage(), rogue)),
		Artifact:   "unreachable",
	})
	_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
	if code := DiagnosticCode(err); code != "vendor_metadata_inconsistent" {
		t.Fatalf("code = %q, want vendor_metadata_inconsistent (error %v)", code, err)
	}
}

// TestAuditedVendorAllowancesAreWithheldFromAReplacedModule: §4.2.3 says no
// allowance a profile grants to audited third-party vendored code extends to a
// module carrying a replacement. Each case is accepted for an ordinary
// vendored module and rejected for the replaced one, so the test proves the
// allowance exists and that it is withheld, not merely that something failed.
func TestAuditedVendorAllowancesAreWithheldFromAReplacedModule(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		mark func(t *testing.T, item *packageJSON)
	}{
		{
			name: "pure Go assembly in a vendored package", code: "go_assembly_forbidden",
			mark: func(t *testing.T, item *packageJSON) {
				t.Helper()
				writeTestFile(t, filepath.Join(item.Dir, "mask.s"), []byte("// pure Go assembly\n"), 0o644)
				item.SFiles = []string{"mask.s"}
			},
		},
		{
			// The motivating case: bubbletea reaches clipperhouse/displaywidth,
			// whose gen.go carries a bare //go:generate that no released
			// version drops. The directive is inert under -mod=vendor, so an
			// audited vendored dependency may carry it and a replaced module
			// may not.
			name: "//go:generate in a vendored package", code: "go_generator_forbidden",
			mark: func(t *testing.T, item *packageJSON) {
				t.Helper()
				writeTestFile(t, filepath.Join(item.Dir, "board.go"),
					[]byte("package board\n\n//go:generate go run -C internal/gen .\n"), 0o644)
			},
		},
		{
			name: "cgo_import_dynamic under the golang.org/x/sys allowlist", code: "go_forbidden_compiler_directive",
			mark: func(t *testing.T, item *packageJSON) {
				t.Helper()
				writeTestFile(t, filepath.Join(item.Dir, "board.go"),
					[]byte("package board\n\n//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.dylib\"\n"), 0o644)
				item.ImportPath = "golang.org/x/sys"
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Run("allowed for an audited third-party module", func(t *testing.T) {
				fixture := newModuleRootsFixture(t)
				fixture.modules = nil
				writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"),
					[]byte("# example.test/vendored v1.0.0\n## explicit; go 1.23\nexample.test/vendored\n"), 0o644)
				vendored := fixture.replacedPackage()
				vendored.Module.Replace = nil
				testCase.mark(t, &vendored)
				fixture.start(stubScript{
					ListStdout: string(encodePackages(t, fixture.rootPackage(), vendored)),
					Artifact:   "audited third party executable",
				})
				if _, err := Build(context.Background(), fixture.request(moduleRootLimits())); err != nil {
					t.Fatalf("the audited third-party allowance did not apply: %v", err)
				}
			})
			t.Run("withheld from a replaced module", func(t *testing.T) {
				fixture := newModuleRootsFixture(t)
				replaced := fixture.replacedPackage()
				testCase.mark(t, &replaced)
				fixture.start(stubScript{
					ListStdout: string(encodePackages(t, fixture.rootPackage(), replaced)),
					Artifact:   "unreachable",
				})
				_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
				if code := DiagnosticCode(err); code != testCase.code {
					t.Fatalf("code = %q, want %s (error %v)", code, testCase.code, err)
				}
			})
		})
	}
}

// TestDeclaredModuleDirectoriesJoinTheScanSurface pins the scan §4.2.3 extends
// over the declared directory itself, which the `go list` stream never reports
// because that copy takes no part in `-mod=vendor` resolution.
func TestDeclaredModuleDirectoriesJoinTheScanSurface(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		path    string
		payload string
		code    string
	}{
		{name: "assembly", path: "mask.s", payload: "// assembly\n", code: "go_assembly_forbidden"},
		{name: "host object", path: "prebuilt.syso", payload: "\x00", code: "go_syso_forbidden"},
		{name: "C source", path: "shim.c", payload: "int shim(void) { return 0; }\n", code: "go_native_input_forbidden"},
		{name: "C header", path: "shim.h", payload: "int shim(void);\n", code: "go_native_input_forbidden"},
		{name: "C++ source", path: "shim.cpp", payload: "int shim() { return 0; }\n", code: "go_native_input_forbidden"},
		{name: "Objective-C source", path: "shim.m", payload: "int shim(void) { return 0; }\n", code: "go_native_input_forbidden"},
		{name: "Fortran source", path: "shim.f90", payload: "end\n", code: "go_native_input_forbidden"},
		{name: "SWIG interface", path: "shim.swig", payload: "%module shim\n", code: "go_native_input_forbidden"},
		{
			name: "cgo", path: "cgo.go", code: "cgo_required",
			payload: "package board\n\n// #include <stdio.h>\nimport \"C\"\n",
		},
		{
			name: "cgo below a subdirectory", path: "codec/cgo.go", code: "cgo_required",
			payload: "package codec\n\nimport \"C\"\n",
		},
		{
			name: "forbidden compiler directive", path: "dynamic.go", code: "go_forbidden_compiler_directive",
			payload: "package board\n\n//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.dylib\"\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newModuleRootsFixture(t)
			writeTestFile(t, filepath.Join(fixture.root, "pkg", "board", filepath.FromSlash(testCase.path)),
				[]byte(testCase.payload), 0o644)
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage())),
				Artifact:   "unreachable",
			})
			_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
			if code := DiagnosticCode(err); code != testCase.code {
				t.Fatalf("code = %q, want %s (error %v)", code, testCase.code, err)
			}
		})
	}
}

// TestTheDeclaredScanSurfaceStopsWhereGoWouldNeverCompile keeps the
// conservative scan from rejecting files that are an input under no build
// configuration at all.
func TestTheDeclaredScanSurfaceStopsWhereGoWouldNeverCompile(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		path    string
		payload string
	}{
		{name: "testdata", path: "testdata/shim.c", payload: "int shim(void) { return 0; }\n"},
		{name: "nested vendor tree", path: "vendor/example.test/dep/mask.s", payload: "// audited assembly\n"},
		{name: "dot directory", path: ".build/shim.c", payload: "int shim(void) { return 0; }\n"},
		{name: "underscore directory", path: "_scratch/shim.c", payload: "int shim(void) { return 0; }\n"},
		{name: "test file", path: "board_test.go", payload: "package board\n\nimport \"C\"\n"},
		{
			name: "the cgo import spelled inside a string", path: "doc.go",
			payload: "package board\n\nconst example = `import \"C\"`\n",
		},
		{name: "an unrelated data file", path: "board.json", payload: "{}\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newModuleRootsFixture(t)
			writeTestFile(t, filepath.Join(fixture.root, "pkg", "board", filepath.FromSlash(testCase.path)),
				[]byte(testCase.payload), 0o644)
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage(), fixture.replacedPackage())),
				Artifact:   "multi module executable",
			})
			if _, err := Build(context.Background(), fixture.request(moduleRootLimits())); err != nil {
				t.Fatalf("the declared-directory scan rejected a file Go would never compile: %v", err)
			}
		})
	}
}

// TestTheDeclaredModulesSurfaceIsClosed keeps `modules` under the same rule as
// every other package-controlled build-command field: it must be exactly the
// list the manager validated.
func TestTheDeclaredModulesSurfaceIsClosed(t *testing.T) {
	base := BuildRequest{SourceDir: "build/cmd/tool", Modules: []string{"pkg/board"}}
	base.CommandObject = BuildCommand{
		"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool", "modules": []string{"pkg/board"},
	}
	if err := validatePackageCommandSurface(base); err != nil {
		t.Fatalf("the declared schema-8 surface was rejected: %v", err)
	}
	decoded := base
	decoded.CommandObject = BuildCommand{
		"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool", "modules": []any{"pkg/board"},
	}
	if err := validatePackageCommandSurface(decoded); err != nil {
		t.Fatalf("a decoded manifest modules list was rejected: %v", err)
	}
	for _, testCase := range []struct {
		name     string
		modules  []string
		declared any
	}{
		{name: "widened", modules: []string{"pkg/board"}, declared: []string{"pkg/board", "pkg/extra"}},
		{name: "reordered", modules: []string{"pkg/board", "pkg/extra"}, declared: []string{"pkg/extra", "pkg/board"}},
		{name: "substituted", modules: []string{"pkg/board"}, declared: []string{"pkg/extra"}},
		{name: "declared where nothing was validated", modules: nil, declared: []string{"pkg/board"}},
		{name: "omitted where a list was validated", modules: []string{"pkg/board"}, declared: nil},
		{name: "not a list", modules: nil, declared: "pkg/board"},
		{name: "not a list of strings", modules: nil, declared: []any{1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := BuildRequest{SourceDir: "build/cmd/tool", Modules: testCase.modules}
			request.CommandObject = BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool"}
			if testCase.declared != nil {
				request.CommandObject["modules"] = testCase.declared
			}
			if code := DiagnosticCode(validatePackageCommandSurface(request)); code != CodePackageInfluenceForbidden {
				t.Fatalf("code = %q, want %s", code, CodePackageInfluenceForbidden)
			}
		})
	}
}

// TestAnEmptyDeclarationIsTheSchemaSixSurface: an absent and an empty modules
// list are one declaration, so neither spelling changes the closed surface a
// schema-6 or schema-7 command presents.
func TestAnEmptyDeclarationIsTheSchemaSixSurface(t *testing.T) {
	for _, declared := range []any{nil, []string{}, []any{}} {
		request := BuildRequest{SourceDir: "build/cmd/tool"}
		request.CommandObject = BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool"}
		if declared != nil {
			request.CommandObject["modules"] = declared
		}
		if err := validatePackageCommandSurface(request); err != nil {
			t.Fatalf("modules=%#v rejected: %v", declared, err)
		}
	}
}

// TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched compiles the multi-module
// vendored fixture with the real toolchain, which is the only way to prove the
// declared-root path against Go's own vendor consistency check rather than
// against a canned stream.
func TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched(t *testing.T) {
	if os.Getenv("CURATOR_REAL_GO_BUILD_TEST") != "1" {
		t.Skip("set CURATOR_REAL_GO_BUILD_TEST=1 for the bounded native go-v1 integration")
	}
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	if err := os.CopyFS(snapshot, os.DirFS("testdata/realmodules")); err != nil {
		t.Fatal(err)
	}
	snapshot = mustPhysical(t, snapshot)
	token, err := buildsource.Validate(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = token.Close() }()

	config := ConfigFromEnvironment(t.TempDir(), snapshot)
	config.CuratorGo = filepath.Join(build.Default.GOROOT, "bin", platformGoName)
	session, err := Establish(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	request := BuildRequest{
		Session: session, Source: token,
		CommandObject: BuildCommand{
			"type": "build", "driver": "go-v1", "source_dir": "tools/cli",
			"modules": []string{"pkg/board", "pkg/remoteconfig"},
		},
		BuildRoot: "tools/cli", SourceDir: "tools/cli", Command: "golden-tool",
		Modules: []string{"pkg/board", "pkg/remoteconfig"}, RuntimeRoots: []string{"scripts"},
		Limits: ResourceLimits{
			Timeout: 90 * time.Second, OutputBytes: 2 * 1024 * 1024, ArtifactBytes: 64 * 1024 * 1024,
			FileBytes: defaultFileLimit, DiskBytes: defaultDiskLimit, MemoryBytes: defaultMemoryLimit,
			Processes: defaultProcessLimit,
		},
	}
	result, err := Build(context.Background(), request)
	if err != nil {
		t.Fatalf("the real toolchain rejected the declared multi-module build: %v", err)
	}
	if result.Artifact.Metadata.Size <= 0 || result.Artifact.Metadata.SHA256 == "" {
		t.Fatalf("artifact = %+v", result.Artifact)
	}

	// The same snapshot without the declaration must be refused: the effective
	// replace set is not empty, so the schema-6 single-module rule denies it.
	undeclared := request
	undeclared.Modules = nil
	undeclared.CommandObject = BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "tools/cli"}
	if _, err := Build(context.Background(), undeclared); DiagnosticCode(err) != "build_module_root_directive_undeclared" {
		t.Fatalf("undeclared build = %v, want build_module_root_directive_undeclared", err)
	}
	// Deliberately do not execute result.Artifact.StagedPath.
}

// moduleRootVectorCase is one published module-roots vector, read as the
// driver consumes it: a snapshot to materialize, a declaration to present, the
// annotations the build root's vendor/modules.txt carries, and the exact point
// in the fixed process the operation must stop at.
type moduleRootVectorCase struct {
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
}

// TestModuleRootVectorsDriveTheWholeBuild runs the published module-roots
// family through Build itself, not through the validators underneath it.
//
// internal/moduleroots already asserts each vector against the half of
// §4.2.3 that owns it. This asserts the thing a manager is actually judged on:
// that the driver reaches exactly the process state the vector records.
// `go_list_started` and `go_build_started` are read off the stub launcher's
// own call log, so a rejection that arrives one fixed command too late fails
// here even when it carries the right diagnostic.
func TestModuleRootVectorsDriveTheWholeBuild(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "module-roots.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no module-roots vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Cases []moduleRootVectorCase `json:"cases"`
	}
	if err := json.Unmarshal(payload, &suite); err != nil {
		t.Fatal(err)
	}
	if len(suite.Cases) == 0 {
		t.Fatal("the module-roots vector family published no cases")
	}
	for _, testCase := range suite.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			fixture := newVectorFixture(t, testCase)
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage())),
				Artifact:   "vector executable",
			})
			_, buildErr := Build(context.Background(), fixture.request(moduleRootLimits()))

			listed, built := false, false
			for _, call := range fixture.sourceAwareCalls() {
				switch call.Argv[0] {
				case "list":
					listed = true
				case "build":
					built = true
				}
			}
			if listed != testCase.GoListStarted || built != testCase.GoBuildStarted {
				t.Fatalf("go list started = %t (want %t), go build started = %t (want %t); error %v",
					listed, testCase.GoListStarted, built, testCase.GoBuildStarted, buildErr)
			}
			if testCase.BuildPermitted {
				if buildErr != nil {
					t.Fatalf("a permitted build was rejected: %v", buildErr)
				}
				return
			}
			if code := DiagnosticCode(buildErr); code != testCase.ExpectedError {
				t.Fatalf("code = %q, want %s (error %v)", code, testCase.ExpectedError, buildErr)
			}
		})
	}
}

// newVectorFixture materializes one vector's snapshot and presents its
// declaration to the driver. The build root doubles as the command's source
// directory, which §4.2 permits, so the vector needs to describe no package
// layout of its own.
//
// Both `build_root` and `build_roots` are presented: §4.2.3's containment rule
// names every declared build root, and today every published case declares the
// one it compiles. Reading the field rather than reconstructing it is what
// makes a future case that declares a second root actually exercise the rule
// instead of passing on the single-root form.
func newVectorFixture(t *testing.T, testCase moduleRootVectorCase) *workerFixture {
	t.Helper()
	if len(testCase.Snapshot.LinkPaths) != 0 {
		// Passing over a link fixture would report a green run for a rule this
		// test never exercised.
		t.Fatalf("vector %q declares link_paths, which this test does not materialise", testCase.Name)
	}
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	for _, directory := range testCase.Snapshot.Directories {
		if err := os.MkdirAll(filepath.Join(snapshot, filepath.FromSlash(directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range testCase.Snapshot.GoModFiles {
		writeTestFile(t, filepath.Join(snapshot, filepath.FromSlash(file)), []byte("module fixture\n\ngo 1.23\n"), 0o644)
	}
	buildRootRel := testCase.Declaration.BuildRoot
	if buildRootRel == "" {
		t.Fatalf("vector %q declares no build_root", testCase.Name)
	}
	if !isListed(buildRootRel, testCase.Declaration.BuildRoots) {
		// A vector whose compiled root is not in its own declared set would
		// describe a manifest the parser rejects, so materialising it would
		// prove nothing about §4.2.3.
		t.Fatalf("vector %q declares build_root %q outside its build_roots %v",
			testCase.Name, buildRootRel, testCase.Declaration.BuildRoots)
	}
	writeTestFile(t, filepath.Join(snapshot, filepath.FromSlash(buildRootRel), "main.go"),
		[]byte("package main\n\nfunc main() {}\n"), 0o644)
	writeTestFile(t, filepath.Join(snapshot, filepath.FromSlash(buildRootRel), "vendor", "modules.txt"),
		[]byte(strings.Join(testCase.VendorModuleAnnotations, "\n")+"\n"), 0o644)
	snapshot = mustPhysical(t, snapshot)
	return &workerFixture{
		t: t, root: snapshot,
		buildRoot:     filepath.Join(snapshot, filepath.FromSlash(buildRootRel)),
		sourceDir:     filepath.Join(snapshot, filepath.FromSlash(buildRootRel)),
		buildRootRel:  buildRootRel,
		sourceDirRel:  buildRootRel,
		buildRootsRel: testCase.Declaration.BuildRoots,
		modules:       testCase.Declaration.Modules,
		runtimeRoots:  testCase.Declaration.RuntimeRoots,
	}
}
