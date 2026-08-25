package godriver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var forbiddenCompilerDirective = []byte("//go:cgo_import_dynamic")

var generatorDirective = []byte("//go:generate")

// Severity codes returned by scanSourceDirectives, ordered so the strongest
// rejection class wins when a file carries more than one directive.
const (
	directiveNone = iota
	directiveCgoImportDynamic
	directiveGenerate
)

type packageError struct {
	ImportStack []string
	Pos         string
	Err         string
}

type moduleJSON struct {
	Path       string
	Query      string
	Version    string
	Versions   []string
	Replace    *moduleJSON
	Time       *json.RawMessage
	Update     *moduleJSON
	Main       bool
	Indirect   bool
	Dir        string
	GoMod      string
	GoVersion  string
	Retracted  []string
	Deprecated string
	Error      *moduleError
}

type moduleError struct{ Err string }

type packageJSON struct {
	Dir           string
	ImportPath    string
	ImportComment string
	Name          string
	Doc           string
	Target        string
	Shlib         string
	Root          string
	ConflictDir   string
	ForTest       string
	Export        string
	BuildID       string
	Module        *moduleJSON
	Match         []string
	Goroot        bool
	Standard      bool
	DepOnly       bool
	BinaryOnly    bool
	Incomplete    bool
	Stale         bool
	StaleReason   string
	Error         *packageError
	DepsErrors    []*packageError

	GoFiles           []string
	CgoFiles          []string
	CompiledGoFiles   []string
	IgnoredGoFiles    []string
	IgnoredOtherFiles []string
	CFiles            []string
	CXXFiles          []string
	MFiles            []string
	HFiles            []string
	FFiles            []string
	SFiles            []string
	SwigFiles         []string
	SwigCXXFiles      []string
	SysoFiles         []string
	EmbedFiles        []string
	EmbedPatterns     []string
	TestGoFiles       []string
	XTestGoFiles      []string
	Imports           []string
	ImportMap         map[string]string
	Deps              []string
	TestImports       []string
	XTestImports      []string
}

type graphValidation struct {
	BuildRoot string
	SourceDir string
	GOROOT    string

	// Snapshot is the canonical frozen snapshot root. BuildRootRel and
	// Modules are protocol (slash-separated) paths relative to it: the
	// command's build root, and the schema-8 first-party module directories
	// that build root replaces (Spec §4.2.3). Modules is empty for every
	// schema-6 and schema-7 command, which is the single-module build root.
	Snapshot     string
	BuildRootRel string
	Modules      []string
}

func validatePackageGraph(payload []byte, validation graphValidation) error {
	if len(payload) == 0 {
		return diagnostic("go_list_incomplete", "go list returned an empty package stream")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	packages := make([]packageJSON, 0, 32)
	seen := make(map[string]bool)
	for {
		var item packageJSON
		err := decoder.Decode(&item)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return diagnosticErr("go_list_malformed", err, "go list returned an invalid or incomplete JSON stream")
		}
		if item.ImportPath == "" || seen[item.ImportPath] {
			return diagnostic("go_list_incomplete", "go list returned an empty or duplicate import path")
		}
		seen[item.ImportPath] = true
		packages = append(packages, item)
	}
	if len(packages) == 0 {
		return diagnostic("go_list_incomplete", "go list returned no package results")
	}

	rootIndexes := make([]int, 0, 2)
	for index := range packages {
		if !packages[index].DepOnly {
			rootIndexes = append(rootIndexes, index)
		}
	}
	if len(rootIndexes) == 0 {
		return diagnostic("build_package_not_main", "go list returned no root package")
	}
	if len(rootIndexes) != 1 {
		return diagnostic("build_package_ambiguous", "go list returned %d root packages", len(rootIndexes))
	}
	rootPackage := packages[rootIndexes[0]]
	if rootPackage.Name != "main" || rootPackage.Dir != validation.SourceDir {
		return diagnostic("build_package_not_main", "selected root is not the canonical main package")
	}

	// Spec §4.2.3 fixes the order: the effective replace set is read and the
	// bijection is checked once `go list` has returned, and only then does the
	// scan surface run with the declared directories included and the
	// audited-vendor allowances withheld from every replaced module.
	replaced, err := admitModuleRoots(validation)
	if err != nil {
		return err
	}

	hasVendoredModule := false
	for index := range packages {
		item := &packages[index]
		if item.Module != nil && !item.Module.Main {
			hasVendoredModule = true
		}
		if item.Incomplete || item.Error != nil || len(item.DepsErrors) != 0 {
			return diagnostic("go_list_incomplete", "package %q is incomplete or carries dependency errors", item.ImportPath)
		}
		if item.ForTest != "" {
			return diagnostic("go_test_input_forbidden", "go list selected a test package")
		}
		if err := validatePackageInputs(*item, validation, replaced); err != nil {
			return err
		}
	}
	if hasVendoredModule {
		modules := filepath.Join(validation.BuildRoot, "vendor", "modules.txt")
		if err := validateRegularAbsolute(modules, validation.BuildRoot, false); err != nil {
			return diagnosticErr("vendor_metadata_inconsistent", err, "vendored graph lacks a regular in-root vendor/modules.txt")
		}
	}
	return scanDeclaredModules(validation)
}

func validatePackageInputs(item packageJSON, validation graphValidation, replaced map[string]bool) error {
	trustedStandard := item.Standard && item.Goroot
	root := validation.BuildRoot
	// A result whose module carries a replacement is first-party code the
	// package declared, so §4.2.3 withholds from it every allowance a profile
	// grants to audited third-party vendored code.
	firstParty := item.Module != nil && item.Module.Replace != nil
	if trustedStandard {
		root = validation.GOROOT
		if item.Module != nil {
			return diagnostic("go_standard_input_escape", "standard package %q unexpectedly has module metadata", item.ImportPath)
		}
		if item.Root != validation.GOROOT && (item.Root != "" || !strings.HasPrefix(item.ImportPath, "vendor/")) {
			return diagnostic("go_standard_input_escape", "standard package %q has an unexpected Root", item.ImportPath)
		}
	} else if err := validateModule(item, validation, replaced); err != nil {
		return err
	}
	if !trustedStandard && item.Root != "" {
		if err := validateDirectory(item.Root, validation.BuildRoot, false); err != nil {
			return diagnosticErr("go_source_input_escape", err, "package %q Root escapes the build root", item.ImportPath)
		}
	}
	if err := validateDirectory(item.Dir, root, trustedStandard); err != nil {
		return packagePathError(item, "directory", err)
	}
	if len(item.SysoFiles) != 0 {
		return diagnostic("go_syso_forbidden", "package %q contains SysoFiles", item.ImportPath)
	}
	if !trustedStandard {
		nativeFields := []struct {
			name  string
			files []string
		}{
			{"CgoFiles", item.CgoFiles}, {"CFiles", item.CFiles}, {"CXXFiles", item.CXXFiles}, {"MFiles", item.MFiles},
			{"HFiles", item.HFiles}, {"FFiles", item.FFiles}, {"SwigFiles", item.SwigFiles}, {"SwigCXXFiles", item.SwigCXXFiles},
		}
		for _, field := range nativeFields {
			if len(field.files) != 0 {
				if field.name == "CgoFiles" {
					return diagnostic("cgo_required", "package %q contains active cgo input", item.ImportPath)
				}
				return diagnostic("go_native_input_forbidden", "package %q contains %s", item.ImportPath, field.name)
			}
		}
		if len(item.SFiles) != 0 {
			vendorRoot := filepath.Join(validation.BuildRoot, "vendor")
			if firstParty || !strictlyBelow(item.Dir, vendorRoot) {
				return diagnostic("go_assembly_forbidden", "package %q contains non-standard assembly", item.ImportPath)
			}
			for _, name := range item.SFiles {
				if _, err := validateRegularInput(item.Dir, name, validation.BuildRoot, false); err != nil {
					return diagnosticErr("go_assembly_forbidden", err, "package %q contains an invalid assembly input", item.ImportPath)
				}
			}
		}
	}

	active := append([]string(nil), item.GoFiles...)
	active = append(active, item.CompiledGoFiles...)
	if trustedStandard {
		active = append(active, item.CgoFiles...)
		active = append(active, item.CFiles...)
		active = append(active, item.CXXFiles...)
		active = append(active, item.MFiles...)
		active = append(active, item.HFiles...)
		active = append(active, item.FFiles...)
		active = append(active, item.SFiles...)
		active = append(active, item.SwigFiles...)
		active = append(active, item.SwigCXXFiles...)
	}
	for _, name := range active {
		path, err := validateRegularInput(item.Dir, name, root, trustedStandard)
		if err != nil {
			return packagePathError(item, "source", err)
		}
		if !trustedStandard && isListed(name, item.GoFiles) {
			matched, readErr := scanSourceDirectives(path)
			if readErr != nil {
				return diagnosticErr("go_source_unreadable", readErr, "cannot read active Go file in %q", item.ImportPath)
			}
			if matched == directiveCgoImportDynamic && (firstParty ||
				item.ImportPath != "golang.org/x/sys" && !strings.HasPrefix(item.ImportPath, "golang.org/x/sys/")) {
				return diagnostic("go_forbidden_compiler_directive", "package %q contains //go:cgo_import_dynamic", item.ImportPath)
			}
			// §2.3 words //go:generate as inert: managers never run generators
			// and `go build -mod=vendor` does not execute them, so its presence
			// in an already materialized vendor tree does not fail preflight.
			// The build root and a replaced module are code the package itself
			// declares, so both stay held to the unexceptioned rule.
			if matched == directiveGenerate && (firstParty || !strictlyBelow(item.Dir, filepath.Join(validation.BuildRoot, "vendor"))) {
				return diagnostic("go_generator_forbidden", "package %q contains an active generator directive", item.ImportPath)
			}
		}
	}
	for _, name := range item.EmbedFiles {
		if _, err := validateRegularInput(item.Dir, name, root, trustedStandard); err != nil {
			code := "go_embed_input_escape"
			if trustedStandard {
				code = "go_standard_input_escape"
			}
			return diagnosticErr(code, err, "package %q has an escaped or invalid embed input", item.ImportPath)
		}
	}
	if !trustedStandard {
		pgo := filepath.Join(item.Dir, "default.pgo")
		if info, err := os.Lstat(pgo); err == nil && info.Mode().IsRegular() {
			return diagnostic("go_pgo_forbidden", "package %q contains default.pgo", item.ImportPath)
		} else if err != nil && !os.IsNotExist(err) {
			return diagnosticErr("go_pgo_forbidden", err, "cannot safely inspect default.pgo for %q", item.ImportPath)
		}
	}
	return nil
}

// scanSourceDirectives reports the highest-severity directive an active Go file
// contains: directiveCgoImportDynamic, directiveGenerate, or directiveNone.
//
// Severity, not first hit, decides the result. //go:generate is exempt inside a
// materialized vendor tree while //go:cgo_import_dynamic is not, so stopping the
// scan at a //go:generate would let a cgo directive in any later window ride in
// unread and turn the carve-out into a bypass. Only the cgo directive, which
// nothing weaker can override, ends the scan early; a generate hit is recorded
// and the file is still read to EOF.
func scanSourceDirectives(path string) (int, error) {
	matched := directiveNone
	err := scanFileWindows(path, len(forbiddenCompilerDirective)-1, func(window []byte) bool {
		if bytes.Contains(window, forbiddenCompilerDirective) {
			matched = directiveCgoImportDynamic
			return true
		}
		if matched == directiveNone && bytes.Contains(window, generatorDirective) {
			matched = directiveGenerate
		}
		return false
	})
	if err != nil {
		return directiveNone, err
	}
	return matched, nil
}

func validateModule(item packageJSON, validation graphValidation, replaced map[string]bool) error {
	buildRoot := validation.BuildRoot
	if item.Module == nil || item.Module.Path == "" || item.Module.Error != nil {
		return diagnostic("vendor_metadata_inconsistent", "non-standard package %q has missing or failed module metadata", item.ImportPath)
	}
	module := item.Module
	if module.Main {
		if module.Replace != nil {
			return diagnostic("vendor_metadata_inconsistent", "package %q resolves through a replaced main module", item.ImportPath)
		}
		if module.Dir != buildRoot || module.GoMod != filepath.Join(buildRoot, "go.mod") {
			return diagnostic("nested_build_module", "package %q resolves through an escaped or nested main module", item.ImportPath)
		}
		if err := validateRegularAbsolute(module.GoMod, buildRoot, false); err != nil {
			return diagnosticErr("build_module_missing", err, "main module go.mod is invalid")
		}
		return nil
	}
	vendorRoot := filepath.Join(buildRoot, "vendor")
	if module.Replace != nil {
		// A replacement is admitted only because admitModuleRoots proved that
		// vendor/modules.txt materializes exactly one directive for this
		// module and that it names a declared directory the driver validated
		// against the snapshot. The stream's own Module.Replace.Dir and
		// Module.Replace.GoMod are never read: §4.2.3 forbids treating them as
		// evidence that any path exists. A replacement the effective set does
		// not carry means `go list` and vendor/modules.txt disagree.
		if !replaced[module.Path] {
			return diagnostic("vendor_metadata_inconsistent",
				"package %q carries a replacement that vendor/modules.txt does not materialize", item.ImportPath)
		}
		// The compiled bytes still come from below R/vendor. A replaced module
		// is not versioned there, which is the one rule §4.2.3 relaxes.
		if !strictlyBelow(item.Dir, vendorRoot) {
			return diagnostic("vendor_dependency_missing", "package %q is not resolved from the checked-in vendor tree", item.ImportPath)
		}
	} else if module.Version == "" || !strictlyBelow(item.Dir, vendorRoot) {
		return diagnostic("vendor_dependency_missing", "package %q is not resolved from the checked-in vendor tree", item.ImportPath)
	}
	for _, candidate := range []string{module.Dir, module.GoMod} {
		if candidate != "" {
			if err := validateContainedPath(candidate, vendorRoot, false); err != nil {
				return diagnosticErr("go_module_input_escape", err, "vendored module metadata escapes the build root")
			}
		}
	}
	return nil
}

func validateDirectory(path, root string, allowToolchainLinks bool) error {
	if err := validateContainedPath(path, root, allowToolchainLinks); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err == nil && allowToolchainLinks && info.Mode()&fs.ModeSymlink != 0 {
		info, err = os.Stat(path)
	}
	if err != nil || !info.IsDir() || (!allowToolchainLinks && info.Mode()&fs.ModeSymlink != 0) {
		return fmt.Errorf("not a real directory: %w", err)
	}
	return nil
}

func validateRegularInput(directory, name, root string, allowToolchainLinks bool) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) {
		return "", errors.New("empty or NUL-containing input name")
	}
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(directory, filepath.FromSlash(name))
	}
	if err := validateRegularAbsolute(path, root, allowToolchainLinks); err != nil {
		return "", err
	}
	return path, nil
}

func validateRegularAbsolute(path, root string, allowToolchainLinks bool) error {
	if err := validateContainedPath(path, root, allowToolchainLinks); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if allowToolchainLinks && info.Mode()&fs.ModeSymlink != 0 {
		resolved, resolveErr := filepath.EvalSymlinks(path)
		if resolveErr != nil || !isWithin(resolved, root) {
			return fmt.Errorf("unsafe toolchain link: %w", resolveErr)
		}
		info, err = os.Stat(path)
	}
	if err != nil || !info.Mode().IsRegular() || (!allowToolchainLinks && info.Mode()&fs.ModeSymlink != 0) {
		return fmt.Errorf("not a regular input: %w", err)
	}
	return nil
}

func validateContainedPath(path, root string, allowToolchainLinks bool) error {
	if !filepath.IsAbs(path) || path != root && !isWithin(path, root) {
		return errors.New("path is outside its trusted root")
	}
	clean := filepath.Clean(path)
	if clean != path {
		return errors.New("path is not canonical")
	}
	if allowToolchainLinks {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil || !isWithin(resolved, root) {
			return fmt.Errorf("resolved path escapes trusted root: %w", err)
		}
	}
	return nil
}

func packagePathError(item packageJSON, kind string, err error) error {
	code := "go_source_input_escape"
	if item.Standard && item.Goroot {
		code = "go_standard_input_escape"
	}
	return diagnosticErr(code, err, "package %q has an invalid %s input", item.ImportPath, kind)
}

func isListed(name string, values []string) bool {
	for _, value := range values {
		if name == value {
			return true
		}
	}
	return false
}
