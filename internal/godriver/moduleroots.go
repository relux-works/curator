package godriver

import (
	"bytes"
	"errors"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/moduleroots"
)

// This file carries the driver's half of Spec §4.2.3, first-party module
// roots. The declaration itself was already validated at parse time; the
// driver re-verifies it against the frozen snapshot before the fixed `go list`
// rather than trusting the caller, reads the effective replace set from
// `<build root>/vendor/modules.txt` after `go list` returns, and admits a
// replaced module exactly when the bijection mapped it onto a directory the
// driver itself proved.

// verifyModuleDeclaration re-runs the declaration and containment half of
// §4.2.3 against the frozen snapshot, before the fixed `go list`.
//
// The parser already ran this over the same snapshot. The driver runs it again
// because it is the component that starts Go: a caller that skipped or
// mis-plumbed the parse-time check must not be able to hand the compiler a
// directory nothing validated. Every declared directory the driver later
// admits a replacement onto is a directory this call proved.
//
// buildRootRel is the build root this command compiles; buildRoots is the
// skill's whole declared set, which §4.2.3 is written against. The two are
// unioned rather than substituted, so the command's own root is checked even
// when a caller supplies no set at all — the backstop can only get stricter,
// never weaker, than the single-root form it replaces.
func verifyModuleDeclaration(snapshot, buildRootRel string, buildRoots, modules, runtimeRoots []string) error {
	if len(modules) == 0 {
		return nil
	}
	declared := []string{buildRootRel}
	for _, root := range buildRoots {
		if root != buildRootRel {
			declared = append(declared, root)
		}
	}
	err := moduleroots.ValidateDeclaration(snapshot, "", modules, declared, runtimeRoots)
	return moduleRootDiagnostic(err)
}

// admitModuleRoots reads the effective replace set of the build root and
// checks it against the declared module directories, after `go list` has
// returned and before `go build`. It returns the set of module paths whose
// replacement the bijection admitted.
//
// `vendor/modules.txt` is the only surface §4.2.3 admits for this. The
// `Module.Replace` records of the `go list` stream are never the source of the
// set: those paths are derived lexically from `go.mod` text, Go does not stat
// them, and under `-mod=vendor` the stream reports them unchanged when the
// directory they name does not exist.
//
// The check runs even when the command declares nothing: §4.2.3 requires a
// command with an absent or empty `modules` list to have an *empty* effective
// replace set, so an undeclared replacement is rejected rather than ignored.
// An unused directive is materialized in `vendor/modules.txt` exactly like a
// used one, which is what stops one from hiding by going unused.
func admitModuleRoots(validation graphValidation) (map[string]bool, error) {
	payload, err := readVendorModules(validation)
	if err != nil {
		return nil, err
	}
	directives, err := moduleroots.EffectiveReplaceSet(payload)
	if err != nil {
		return nil, moduleRootDiagnostic(err)
	}
	if err := moduleroots.ValidateBijection(validation.BuildRootRel, validation.Modules, directives); err != nil {
		return nil, moduleRootDiagnostic(err)
	}
	// ValidateBijection proved every directive resolves onto exactly one
	// declared directory and that no declaration is left over, so admission is
	// a membership question from here on: is this result's module one the
	// bijection accounted for.
	admitted := make(map[string]bool, len(directives))
	for _, directive := range directives {
		admitted[directive.Module] = true
	}
	return admitted, nil
}

// readVendorModules returns the bytes of the build root's vendor/modules.txt,
// or nothing when the build root carries no vendor metadata at all.
//
// A command that declares module roots cannot be validated without that file:
// the effective replace set is defined by its bytes, so its absence is
// inconsistent vendor metadata rather than an empty replace set.
func readVendorModules(validation graphValidation) ([]byte, error) {
	path := filepath.Join(validation.BuildRoot, "vendor", "modules.txt")
	info, err := os.Lstat(path)
	switch {
	// Lstat does not follow the final component, and IsRegular is false for
	// every non-regular mode bit including ModeSymlink, so this one predicate
	// already rejects a link standing where the metadata must be.
	case err == nil && info.Mode().IsRegular():
		payload, readErr := os.ReadFile(path) // #nosec G304 -- path is derived from the validated build root
		if readErr != nil {
			return nil, diagnosticErr("vendor_metadata_inconsistent", readErr, "cannot read vendor/modules.txt")
		}
		return payload, nil
	case err == nil:
		return nil, diagnostic("vendor_metadata_inconsistent", "vendor/modules.txt is not a regular file")
	case os.IsNotExist(err):
		if len(validation.Modules) != 0 {
			return nil, diagnostic("vendor_metadata_inconsistent",
				"declared module roots require a regular vendor/modules.txt below the build root")
		}
		return nil, nil
	default:
		return nil, diagnosticErr("vendor_metadata_inconsistent", err, "cannot inspect vendor/modules.txt")
	}
}

// moduleRootDiagnostic re-labels a module-root failure as a driver diagnostic
// without moving its stable code or its protocol detail.
func moduleRootDiagnostic(err error) error {
	if err == nil {
		return nil
	}
	var failure *moduleroots.Error
	if !errors.As(err, &failure) {
		return err
	}
	detail := failure.Detail
	if failure.Path != "" {
		detail = failure.Path + ": " + detail
	}
	return &Diagnostic{Code: failure.DiagnosticCode, Detail: detail, Err: err}
}

// nativeDeclaredInput classifies one file of a declared module directory by
// the `go list` field it would occupy, so the directory is held to the same
// emptiness rules §4.2 states for the main module. The extension set is the
// one the Go build system itself recognizes.
var nativeDeclaredInput = map[string]struct{ field, code string }{
	".c":       {"CFiles", "go_native_input_forbidden"},
	".cc":      {"CXXFiles", "go_native_input_forbidden"},
	".cpp":     {"CXXFiles", "go_native_input_forbidden"},
	".cxx":     {"CXXFiles", "go_native_input_forbidden"},
	".m":       {"MFiles", "go_native_input_forbidden"},
	".h":       {"HFiles", "go_native_input_forbidden"},
	".hh":      {"HFiles", "go_native_input_forbidden"},
	".hpp":     {"HFiles", "go_native_input_forbidden"},
	".hxx":     {"HFiles", "go_native_input_forbidden"},
	".f":       {"FFiles", "go_native_input_forbidden"},
	".F":       {"FFiles", "go_native_input_forbidden"},
	".for":     {"FFiles", "go_native_input_forbidden"},
	".f90":     {"FFiles", "go_native_input_forbidden"},
	".swig":    {"SwigFiles", "go_native_input_forbidden"},
	".swigcxx": {"SwigCXXFiles", "go_native_input_forbidden"},
	".s":       {"SFiles", "go_assembly_forbidden"},
	".S":       {"SFiles", "go_assembly_forbidden"},
	".sx":      {"SFiles", "go_assembly_forbidden"},
	".syso":    {"SysoFiles", "go_syso_forbidden"},
}

// scanDeclaredModules extends the directive, cgo, and assembly scan surface of
// §4.2 over every declared module directory, which §4.2.3 requires in the
// declared directory as well as in the vendor copy the compiler reads.
//
// The vendor copy is covered by the `go list` stream, which reports its real
// per-package fields. The declared directory is not in that stream — it takes
// no part in `-mod=vendor` resolution — and the fixed argument vectors admit
// exactly one `go list`, so the directory cannot be classified by the
// toolchain. It is therefore classified conservatively from the tree itself:
// every native-input extension is rejected wherever it appears, and every
// non-test Go file is rejected if it imports "C" or carries the exact
// `//go:cgo_import_dynamic` bytes. That is a superset of what any single build
// configuration would compile, which is the fail-closed direction; §4.2.3
// withholds every audited-vendor allowance here on purpose, and first-party
// code in the package's own repository can simply not use these constructs.
//
// Directories the Go build system never compiles from are skipped, because a
// file there is not an input under any configuration: `testdata`, names
// beginning with "." or "_", and a nested `vendor` tree, which §4.2.3 states
// takes no part in resolution at the build root and which the manager must not
// read as a dependency source.
func scanDeclaredModules(validation graphValidation) error {
	for _, module := range validation.Modules {
		directory := filepath.Join(validation.Snapshot, filepath.FromSlash(module))
		if err := scanDeclaredModule(directory, module); err != nil {
			return err
		}
	}
	return nil
}

func scanDeclaredModule(directory, module string) error {
	walkErr := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != directory && skippedDeclaredDirectory(entry.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		return classifyDeclaredInput(path, module)
	})
	if walkErr == nil {
		return nil
	}
	if DiagnosticCode(walkErr) != "" {
		return walkErr
	}
	return diagnosticErr("go_source_unreadable", walkErr, "cannot scan declared module directory %q", module)
}

func skippedDeclaredDirectory(name string) bool {
	return name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func classifyDeclaredInput(path, module string) error {
	extension := filepath.Ext(path)
	if native, forbidden := nativeDeclaredInput[extension]; forbidden {
		return diagnostic(native.code, "declared module directory %q contains %s: %s",
			module, native.field, filepath.ToSlash(filepath.Base(path)))
	}
	if extension != ".go" || strings.HasSuffix(filepath.Base(path), "_test.go") {
		return nil
	}
	if imports, err := importsCgo(path); err != nil {
		return diagnosticErr("go_source_unreadable", err, "cannot read Go file in declared module directory %q", module)
	} else if imports {
		return diagnostic("cgo_required", "declared module directory %q contains active cgo input: %s",
			module, filepath.Base(path))
	}
	found, err := fileContainsBytes(path, forbiddenCompilerDirective)
	if err != nil {
		return diagnosticErr("go_source_unreadable", err, "cannot read Go file in declared module directory %q", module)
	}
	if found {
		return diagnostic("go_forbidden_compiler_directive",
			"declared module directory %q contains //go:cgo_import_dynamic: %s", module, filepath.Base(path))
	}
	return nil
}

// importsCgo reports whether a Go file imports "C", which is exactly what makes
// it a CgoFile. The import block is parsed rather than byte-matched so the
// bytes `import "C"` inside a string or comment cannot reject a build that has
// no cgo in it. A file that does not parse is not a Go input in any build, so
// it cannot introduce cgo and is left to the byte scan.
func importsCgo(path string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
			return false, err
		}
		return false, nil
	}
	for _, item := range file.Imports {
		if item.Path != nil && item.Path.Value == `"C"` {
			return true, nil
		}
	}
	return false, nil
}

// fileContainsBytes streams path looking for one exact byte sequence.
func fileContainsBytes(path string, needle []byte) (bool, error) {
	found := false
	err := scanFileWindows(path, len(needle)-1, func(window []byte) bool {
		if bytes.Contains(window, needle) {
			found = true
			return true
		}
		return false
	})
	return found, err
}

// scanFileWindows streams path in fixed reads, prefixing each with the last
// overlap bytes of the previous one so a needle straddling a read boundary is
// still seen whole. visit returns true to stop early.
func scanFileWindows(path string, overlap int, visit func(window []byte) bool) error {
	file, err := os.Open(path) // #nosec G304 -- path was constrained to the frozen snapshot
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if overlap < 0 {
		overlap = 0
	}
	buffer := make([]byte, 64*1024)
	carry := make([]byte, 0, overlap)
	for {
		read, readErr := file.Read(buffer)
		if read > 0 {
			window := append(carry, buffer[:read]...)
			if visit(window) {
				return nil
			}
			keep := overlap
			if len(window) < keep {
				keep = len(window)
			}
			carry = append(carry[:0], window[len(window)-keep:]...)
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}
