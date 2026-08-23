// Package moduleroots implements the schema-8 first-party module roots of
// Spec §4.2.3.
//
// The package states a claim and the manager checks it. Nothing here reads a
// replacement as an instruction: a declared directory is admitted only because
// it was validated against the immutable snapshot, and the effective replace
// set is read solely to check the declaration against it. `go.mod` is never
// parsed, and `Module.Replace.Dir`/`Module.Replace.GoMod` are never accepted as
// evidence that a path exists.
//
// The two halves sit on opposite sides of the fixed `go list`, exactly as the
// failure boundary of §4.2.3 requires: ValidateDeclaration runs before it, and
// EffectiveReplaceSet plus ValidateBijection run after it and before
// `go build`.
package moduleroots

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/relux-works/curator/internal/identifiers"
)

// The stable `phase: preflight` diagnostics of the manager profile. Callers
// branch on these codes; the detail text is for operators.
const (
	// DeclarationInvalid: a declared module directory is ".", is not a
	// portable relative path, is duplicated, is not a real link-free
	// directory in the snapshot, or has no go.mod directly inside it.
	DeclarationInvalid = "build_module_root_declaration_invalid"
	// ContainmentInvalid: a declared module directory equals, contains, or is
	// contained by another declared module directory, a declared build root,
	// or a runtime root, under exact or platform-path comparison.
	ContainmentInvalid = "build_module_root_containment_invalid"
	// DirectiveFormUnsupported: an effective replacement carries a version on
	// either side, is a module-to-module redirect, or has an unreadable or
	// unreconciled annotation shape.
	DirectiveFormUnsupported = "build_module_root_directive_form_unsupported"
	// DirectiveUndeclared: an effective replacement names a directory the
	// command does not declare.
	DirectiveUndeclared = "build_module_root_directive_undeclared"
	// DeclarationUnused: a declared module directory is named by no effective
	// replacement.
	DeclarationUnused = "build_module_root_declaration_unused"
)

// GoModName is the module file every declared directory must carry directly.
const GoModName = "go.mod"

// annotationPrefix and annotationSeparator are byte-exact per §4.2.3: only a
// line starting with exactly "# " and containing exactly " => " is a
// replacement annotation.
const (
	annotationPrefix    = "# "
	annotationSeparator = " => "
)

// Error is a module-root failure bound to a stable diagnostic code. Path is
// the manifest field path when the failure is a declaration failure, and empty
// for a build-graph failure that no manifest field owns.
type Error struct {
	DiagnosticCode string
	Path           string
	Detail         string
}

func (err *Error) Error() string {
	message := err.DiagnosticCode
	if err.Detail != "" {
		message += ": " + err.Detail
	}
	if err.Path != "" {
		return err.Path + ": " + message
	}
	return message
}

// Code returns the stable module-root diagnostic carried by err, or an empty
// string when err did not originate at a module-root boundary.
func Code(err error) string {
	var diagnostic *Error
	if errors.As(err, &diagnostic) {
		return diagnostic.DiagnosticCode
	}
	return ""
}

func fail(code, path, format string, args ...any) error {
	return &Error{DiagnosticCode: code, Path: path, Detail: fmt.Sprintf(format, args...)}
}

// ValidateDeclaration checks one build command's declared module directories
// against the immutable snapshot and against every declared build root and
// runtime root. All paths are snapshot-relative POSIX paths; field is the
// manifest path of the declaration, used for diagnostics only.
//
// It performs the whole "declaration and containment" half of §4.2.3, which
// MUST complete before the fixed `go list`.
func ValidateDeclaration(snapshot, field string, modules, buildRoots, runtimeRoots []string) error {
	seen := make(map[string]bool, len(modules))
	for _, module := range modules {
		if !identifiers.PortablePath(module) {
			return fail(DeclarationInvalid, field, "%q is not a portable relative path", module)
		}
		if seen[module] {
			return fail(DeclarationInvalid, field, "%q is declared more than once", module)
		}
		seen[module] = true
		if err := validateDirectory(snapshot, module, field); err != nil {
			return err
		}
	}
	return validateContainment(field, modules, buildRoots, runtimeRoots)
}

// validateDirectory proves the declared directory is a real, link-free
// directory strictly inside the snapshot that carries go.mod directly. Every
// package-controlled component is inspected with Lstat, so no link can
// redirect the check outside the snapshot.
func validateDirectory(snapshot, module, field string) error {
	current := snapshot
	for _, component := range strings.Split(module, "/") {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return fail(DeclarationInvalid, field, "declared module directory does not exist: %s", module)
			}
			return fail(DeclarationInvalid, field, "cannot inspect declared module directory %s: %v", module, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fail(DeclarationInvalid, field, "declared module directory must be link-free: %s", module)
		}
		if !info.IsDir() {
			return fail(DeclarationInvalid, field, "declared module directory must be a directory: %s", module)
		}
	}
	info, err := os.Lstat(filepath.Join(current, GoModName))
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fail(DeclarationInvalid, field, "declared module directory must contain a regular %s directly: %s", GoModName, module)
	}
	return nil
}

// validateContainment rejects a declared directory that equals, contains, or is
// contained by another declared directory, a declared build root, or a runtime
// root. The comparison runs twice: once on the exact protocol paths, and once
// under the platform path mapping of Spec §2, so two paths that differ only by
// case or by another platform folding are rejected even when only one of them
// exists in the snapshot.
func validateContainment(field string, modules, buildRoots, runtimeRoots []string) error {
	ordered := append([]string(nil), modules...)
	sort.Strings(ordered)
	for index, left := range ordered {
		for _, right := range ordered[index+1:] {
			if err := rejectOverlap(field, left, "declared module directory", right, "declared module directory"); err != nil {
				return err
			}
		}
	}
	for _, module := range modules {
		for _, root := range buildRoots {
			if err := rejectOverlap(field, module, "declared module directory", root, "build root"); err != nil {
				return err
			}
		}
		for _, root := range runtimeRoots {
			if err := rejectOverlap(field, module, "declared module directory", root, "runtime root"); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectOverlap(field, left, leftNoun, right, rightNoun string) error {
	if overlaps(left, right) {
		return fail(ContainmentInvalid, field, "%s %s overlaps %s %s", leftNoun, left, rightNoun, right)
	}
	if overlaps(platformKey(left), platformKey(right)) {
		return fail(ContainmentInvalid, field,
			"%s %s collides with %s %s under the platform path mapping", leftNoun, left, rightNoun, right)
	}
	return nil
}

func overlaps(left, right string) bool {
	return contains(left, right) || contains(right, left)
}

// contains reports whether rel is root or lies below it, comparing whole path
// components so "pkg/boarding" is not read as being below "pkg/board".
func contains(root, rel string) bool {
	rootParts := strings.Split(root, "/")
	relParts := strings.Split(rel, "/")
	if len(relParts) < len(rootParts) {
		return false
	}
	for index, part := range rootParts {
		if relParts[index] != part {
			return false
		}
	}
	return true
}

var caseFolder = cases.Fold()

// platformKey is the conservative model of the case-insensitive Windows and
// macOS filesystems this protocol supports: canonical decomposition around a
// full case fold. It is a comparison key only; no stored or hashed path is
// ever normalized.
func platformKey(value string) string {
	return norm.NFD.String(caseFolder.String(norm.NFD.String(value)))
}

// Directive is one effective replacement directive in directory form: a module
// path replaced by a path relative to the build root.
type Directive struct {
	Module string
	Target string
}

// EffectiveReplaceSet reads the effective replace set of a build root from the
// bytes of its vendor/modules.txt. That file is the only surface §4.2.3 admits
// for this: the paths in `go list`'s Module.Replace are derived lexically from
// go.mod text, Go does not stat them, and under -mod=vendor the stream reports
// them unchanged when the directory they name does not exist.
//
// Only annotations are read. Every admitted directive has exactly one token on
// each side; a two-token-left annotation is selection metadata and must be
// reconciled against its one-token-left directive, which is what enforces the
// no-version-on-the-left rule without parsing go.mod.
func EffectiveReplaceSet(modulesTxt []byte) ([]Directive, error) {
	type annotation struct {
		left  []string
		right []string
		line  string
	}
	var annotations []annotation
	for _, line := range strings.Split(string(modulesTxt), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if !strings.HasPrefix(line, annotationPrefix) || !strings.Contains(line, annotationSeparator) {
			continue
		}
		index := strings.Index(line, annotationSeparator)
		left := strings.Fields(strings.TrimPrefix(line[:index], annotationPrefix))
		right := strings.Fields(line[index+len(annotationSeparator):])
		if len(left) == 0 || len(left) > 2 || len(right) == 0 || len(right) > 2 {
			return nil, fail(DirectiveFormUnsupported, "", "unreadable replacement annotation: %s", line)
		}
		annotations = append(annotations, annotation{left: left, right: right, line: line})
	}

	directives := make([]Directive, 0, len(annotations))
	declared := make(map[string]bool, len(annotations))
	for _, item := range annotations {
		if len(item.left) != 1 {
			continue
		}
		if len(item.right) != 1 {
			return nil, fail(DirectiveFormUnsupported, "",
				"replacement is a module-to-module redirect, not a directory: %s", item.line)
		}
		key := item.left[0] + annotationSeparator + strings.Join(item.right, " ")
		if declared[key] {
			return nil, fail(DirectiveFormUnsupported, "", "duplicate replacement annotation: %s", item.line)
		}
		declared[key] = true
		directives = append(directives, Directive{Module: item.left[0], Target: item.right[0]})
	}
	// A two-token-left annotation with no exactly matching one-token-left
	// annotation is a versioned-left directive, which §4.2.3 rejects. The
	// match is on the whole annotation: same left module path, same right
	// replacement.
	for _, item := range annotations {
		if len(item.left) != 2 {
			continue
		}
		if !declared[item.left[0]+annotationSeparator+strings.Join(item.right, " ")] {
			return nil, fail(DirectiveFormUnsupported, "",
				"replacement carries a version on its left side: %s", item.line)
		}
	}
	seen := make(map[string]bool, len(directives))
	for _, directive := range directives {
		if seen[directive.Module] {
			return nil, fail(DirectiveFormUnsupported, "",
				"module %q carries more than one effective replacement", directive.Module)
		}
		seen[directive.Module] = true
	}
	return directives, nil
}

// ValidateBijection checks the one-to-one correspondence §4.2.3 requires
// between a command's declared module directories and its effective
// replacement directives. buildRoot and modules are snapshot-relative POSIX
// paths; each directive target is resolved against buildRoot.
//
// This is the entire use a manager makes of the replace records: they are
// checked against the declaration, never obeyed.
func ValidateBijection(buildRoot string, modules []string, directives []Directive) error {
	declared := make(map[string]bool, len(modules))
	for _, module := range modules {
		declared[module] = true
	}
	named := make(map[string]string, len(modules))
	for _, directive := range directives {
		target, err := resolveTarget(buildRoot, directive)
		if err != nil {
			return err
		}
		if !declared[target] {
			return fail(DirectiveUndeclared, "",
				"replacement of %q names undeclared directory %q", directive.Module, directive.Target)
		}
		// The declaration is consumed by exactly one directive. A second
		// directive resolving onto it names no distinct declaration of its
		// own, so it is undeclared for the same reason.
		if previous, taken := named[target]; taken {
			return fail(DirectiveUndeclared, "",
				"replacements of %q and %q both name declared directory %q", previous, directive.Module, target)
		}
		named[target] = directive.Module
	}
	ordered := append([]string(nil), modules...)
	sort.Strings(ordered)
	for _, module := range ordered {
		if _, used := named[module]; !used {
			return fail(DeclarationUnused, "", "declared module directory %q is named by no replacement", module)
		}
	}
	return nil
}

// resolveTarget maps one directive's right-hand token to a snapshot-relative
// path. Only a relative directory token is admitted; a target that leaves the
// snapshot names no declared directory and is reported as undeclared, exactly
// as a target inside the snapshot that was never declared is.
func resolveTarget(buildRoot string, directive Directive) (string, error) {
	target := directive.Target
	if target == "" || strings.HasPrefix(target, "/") || strings.Contains(target, `\`) ||
		(len(target) > 1 && target[1] == ':') {
		return "", fail(DirectiveFormUnsupported, "",
			"replacement of %q is not a relative directory path: %q", directive.Module, target)
	}
	resolved := path.Join(buildRoot, target)
	if resolved == "." || resolved == ".." || strings.HasPrefix(resolved, "../") {
		return "", fail(DirectiveUndeclared, "",
			"replacement of %q resolves outside the skill snapshot: %q", directive.Module, target)
	}
	return resolved, nil
}
