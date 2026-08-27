package godriver

// Reviewer probe for BUG-260825-11nmd5 / PR 40 (commit c9fe49c).
//
// Drop this file into internal/godriver/ and run:
//   go test ./internal/godriver/ -run TestReviewProbeGeneratorCarveOutHidesCgoImportDynamic -v -count=1
//
// Expected once the scanner is fixed: all three subtests PASS.
// Against c9fe49c: the middle subtest FAILS with code "" (the build succeeds).

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

// scanSourceDirectives stops at the FIRST 64 KiB window that matches ANY
// needle. A //go:generate hit in an earlier window therefore terminates the
// scan before //go:cgo_import_dynamic is ever read, and the vendored-generator
// carve-out then admits the package. Spec profiles/manager.md §2.3 is a
// containment predicate: an active non-standard GoFiles file is "scanned as
// exact bytes and rejected if it contains //go:cgo_import_dynamic".
func TestReviewProbeGeneratorCarveOutHidesCgoImportDynamic(t *testing.T) {
	pad := bytes.Repeat([]byte("// padding padding padding padding padding padding padding\n"), 2000) // > 64 KiB

	for _, testCase := range []struct {
		name    string
		payload []byte
	}{
		{
			name: "cgo directive alone (control)",
			payload: append(append([]byte("package board\n\n"), pad...),
				[]byte("//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.B.dylib\"\n")...),
		},
		{
			name: "generate directive in an earlier window than the cgo directive",
			payload: append(append([]byte("package board\n\n//go:generate go run ./gen\n"), pad...),
				[]byte("//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.B.dylib\"\n")...),
		},
		{
			name: "both directives in the same window (control)",
			payload: []byte("package board\n\n//go:generate go run ./gen\n" +
				"//go:cgo_import_dynamic libc_x x \"/usr/lib/libSystem.B.dylib\"\n"),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newModuleRootsFixture(t)
			fixture.modules = nil
			writeTestFile(t, filepath.Join(fixture.buildRoot, "vendor", "modules.txt"),
				[]byte("# example.test/vendored v1.0.0\n## explicit; go 1.23\nexample.test/vendored\n"), 0o644)
			vendored := fixture.replacedPackage()
			vendored.Module.Replace = nil
			writeTestFile(t, filepath.Join(vendored.Dir, "board.go"), testCase.payload, 0o644)
			fixture.start(stubScript{
				ListStdout: string(encodePackages(t, fixture.rootPackage(), vendored)),
				Artifact:   "audited third party executable",
			})
			_, err := Build(context.Background(), fixture.request(moduleRootLimits()))
			code := DiagnosticCode(err)
			t.Logf("diagnostic code = %q (err %v)", code, err)
			if code != "go_forbidden_compiler_directive" {
				t.Errorf("BYPASS: code = %q, want go_forbidden_compiler_directive", code)
			}
		})
	}
}
