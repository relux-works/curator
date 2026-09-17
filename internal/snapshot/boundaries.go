package snapshot

// Draft local-source boundary enforcement (revision 1).
//
// This file implements protocol/skillfile-sources.md section 2 for local
// acquisition: canonicalize physical paths with symlink and actual
// filesystem case-equivalence rules, distinguish authored inputs from
// managed outputs, validate operator root_inputs, and reject unsafe overlap
// before traversal. Publication recheck lives in internal/staging; capture
// and hashing belong to the snapshot-capture task.
//
// A broad alias path "." stays legal: only the selected package and its
// effective inputs are judged, so path "." with directory
// "agents/skills/review" remains valid while ".agents/skills/review" fails.
//
// Revision 2 (ports, mirrors, aliases) is a separate future leaf. Root-input
// validation is version-agnostic: a v2 policy carries the same root_inputs
// shape, and callers pass its entries through unchanged.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/staging"
)

// LocalAcquisition is a validated local package ready for capture: the
// boundary checks ran before any traversal, the admitted regular files
// are enumerated, and Staging is the operation-private directory the
// capture leaf fills with the frozen copy. The caller owns Staging and
// must call Close. Byte capture, hashing, race detection and store
// publication belong to the capture leaf (TASK-260910-16k7xy), which
// consumes this entry point through Capture; Outputs and RootEntries
// carry the exact admission inputs Capture re-enumerates after the copy
// so a concurrent membership change fails instead of publishing a mix.
type LocalAcquisition struct {
	// Physical is the canonical package directory.
	Physical string
	// Admitted holds the canonical admitted roots: the full package tree
	// for ordinary packages, the validated root-input selections for a
	// project-root package.
	Admitted []string
	// Files holds the admitted regular files in deterministic order.
	Files []string
	// Staging is the private directory reserved for the frozen copy.
	Staging string
	// Outputs holds the managed-output roots used at admission. Capture
	// reuses them for its post-copy re-enumeration.
	Outputs []string
	// RootEntries holds the source-relative root-input selections for a
	// project-root package, or nil for the full-tree admission of an
	// ordinary package. The nil/non-nil distinction matches
	// EnumerateInputs: nil admits the whole tree.
	RootEntries []string
}

// Close removes the reserved staging directory. It is nil-safe and
// reports removal failures.
func (acquisition *LocalAcquisition) Close() error {
	if acquisition == nil || acquisition.Staging == "" {
		return nil
	}
	staging := acquisition.Staging
	acquisition.Staging = ""
	return os.RemoveAll(staging)
}

// PrepareLocalAcquisition is the production local-source acquisition
// entry point: it validates overlap, root_inputs and physical boundaries
// before traversal, enumerates the admitted files, and reserves private
// staging for the frozen copy. Parameters match ValidateLocalPackage;
// stagingParent holds the private directory ("" selects the default
// temporary root). A refusal returns before creating any staging or
// traversing any package bytes.
func PrepareLocalAcquisition(sourceRoot, directory, projectRoot, alias string, outputs []string, rootInputs map[string][]string, knownAliases map[string]bool, stagingParent string) (*LocalAcquisition, error) {
	physical, admitted, err := ValidateLocalPackage(sourceRoot, directory, projectRoot, alias, outputs, rootInputs, knownAliases)
	if err != nil {
		return nil, err
	}
	projectCanonical, err := staging.Canonicalize(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("boundary_identity_unreadable: project root: %v", err)
	}
	var files []string
	var rootEntries []string
	same, err := sameFile(physical, projectCanonical)
	if err != nil {
		return nil, fmt.Errorf("boundary_identity_unreadable: project root: %v", err)
	}
	if same {
		rootEntries = append([]string(nil), rootInputs[alias]...)
		files, err = EnumerateInputs(physical, outputs, rootInputs[alias])
	} else {
		files, err = EnumerateInputs(physical, outputs, nil)
	}
	if err != nil {
		return nil, err
	}
	reserved, err := privatedir.TempStaging(stagingParent, "curator-local-")
	if err != nil {
		return nil, err
	}
	return &LocalAcquisition{Physical: physical, Admitted: admitted, Files: files, Staging: reserved, Outputs: append([]string(nil), outputs...), RootEntries: rootEntries}, nil
}

// ValidateLocalPackage resolves and admits one selected local package before
// traversal or copying.
//
// sourceRoot is the already-joined absolute source root for alias (the
// caller resolves relative Skillfile paths first). directory is the selector
// directory ("." or a portable contained path). projectRoot is the absolute
// declaring project root. outputs lists absolute managed-output roots:
// project .agents and adapter directories, runtime stores, build caches,
// transaction staging and the snapshot store. rootInputs maps source aliases
// to operator-owned root-input entries; knownAliases holds every declared
// source alias for unknown-alias detection (nil disables the check).
//
// It returns the physical package directory and the admitted input roots
// (absolute canonical paths) the publication recheck must protect: the full
// package tree for ordinary packages, the validated root-input selections
// for a project-root package.
func ValidateLocalPackage(sourceRoot, directory, projectRoot, alias string, outputs []string, rootInputs map[string][]string, knownAliases map[string]bool) (string, []string, error) {
	if knownAliases != nil && !knownAliases[alias] {
		return "", nil, fmt.Errorf("source_alias_unknown: %s", alias)
	}
	if directory != "." && (!identifiers.PortablePath(directory) || strings.ContainsAny(directory, "*?[]")) {
		return "", nil, fmt.Errorf("source_selection_invalid: directory must be a portable contained path or '.'")
	}
	if !filepath.IsAbs(sourceRoot) {
		return "", nil, fmt.Errorf("source_selection_invalid: source root must be absolute")
	}
	sourceCanonical, err := staging.Canonicalize(sourceRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("source_member_missing: alias %s: %v", alias, err)
		}
		return "", nil, fmt.Errorf("source_member_invalid: alias %s: %v", alias, err)
	}
	if info, err := os.Stat(sourceCanonical); err != nil || !info.IsDir() { // #nosec G304 -- canonical source root under validation
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("source_member_missing: alias %s: %v", alias, err)
		}
		return "", nil, fmt.Errorf("source_member_invalid: alias %s is not a directory", alias)
	}
	candidate := filepath.Join(sourceCanonical, filepath.FromSlash(directory))
	if pruned, err := staging.IsOutputPath(candidate, outputs); err != nil {
		return "", nil, fmt.Errorf("boundary_identity_unreadable: %s: %v", directory, err)
	} else if pruned {
		return "", nil, fmt.Errorf("source_output_overlap: %s", directory)
	}
	physical, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		class := "source_member_invalid"
		if os.IsNotExist(err) {
			class = "source_member_missing"
		}
		return "", nil, fmt.Errorf("%s: %s: %v", class, directory, err)
	}
	physical = filepath.Clean(physical)
	within, err := staging.Within(sourceCanonical, physical)
	if err != nil {
		return "", nil, fmt.Errorf("boundary_identity_unreadable: %s: %v", directory, err)
	}
	if !within {
		return "", nil, fmt.Errorf("source_selection_invalid: %s escapes source", directory)
	}
	if pruned, err := staging.IsOutputPath(physical, outputs); err != nil {
		return "", nil, fmt.Errorf("boundary_identity_unreadable: %s: %v", directory, err)
	} else if pruned {
		return "", nil, fmt.Errorf("source_output_overlap: %s", directory)
	}
	if info, err := os.Stat(physical); err != nil || !info.IsDir() { // #nosec G304 -- resolved package under validation
		return "", nil, fmt.Errorf("source_member_invalid: %s is not a directory", directory)
	}
	projectCanonical, err := staging.Canonicalize(projectRoot)
	if err != nil {
		return "", nil, fmt.Errorf("boundary_identity_unreadable: project root: %v", err)
	}
	// A project-root package must prove separation through explicit
	// operator root_inputs; anything else would infer author intent from
	// a recursively copied root.
	same, err := sameFile(physical, projectCanonical)
	if err != nil {
		return "", nil, fmt.Errorf("boundary_identity_unreadable: project root: %v", err)
	}
	if same {
		entries, present := rootInputs[alias]
		if !present || len(entries) == 0 {
			return "", nil, fmt.Errorf("source_output_overlap: root package requires root_inputs for %s", alias)
		}
		admitted, err := ValidateRootInputs(alias, entries, physical, outputs, knownAliases)
		if err != nil {
			return "", nil, err
		}
		return physical, admitted, nil
	}
	// Pruning an explicitly declared runtime/build input is an error, not
	// permission to admit an incomplete package.
	if err := checkDeclaredInputsDisjoint(physical, outputs); err != nil {
		return "", nil, err
	}
	return physical, []string{physical}, nil
}

// ValidateRootInputs validates the operator-owned root-input allowlist for
// one source alias whose package is the project root. Each entry is
// source-relative and portable, link-free and disjoint from outputs; a file
// selects itself, a directory its recursive contents. Unknown aliases,
// duplicates and overlaps fail. Every entry must exist; the set must cover
// SKILL.md, the effective manifest and every declared runtime and build
// root. It returns the admitted absolute canonical paths.
//
// This validates admission configuration, not a package-provided hook:
// changing entries requires explicit refresh by the caller, never silent
// reinterpretation.
func ValidateRootInputs(alias string, entries []string, sourceRoot string, outputs []string, knownAliases map[string]bool) ([]string, error) {
	if knownAliases != nil && !knownAliases[alias] {
		return nil, fmt.Errorf("source_alias_unknown: %s", alias)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("source_output_overlap: root package requires root_inputs for %s", alias)
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry] {
			return nil, fmt.Errorf("source_selection_invalid: duplicate root input %q", entry)
		}
		seen[entry] = true
	}
	sourceCanonical, err := staging.Canonicalize(sourceRoot)
	if err != nil {
		return nil, fmt.Errorf("boundary_identity_unreadable: source root: %v", err)
	}
	type validatedRoot struct {
		entry    string
		physical string
	}
	validated := make([]validatedRoot, 0, len(entries))
	for _, entry := range entries {
		if !identifiers.PortablePath(entry) {
			return nil, fmt.Errorf("source_selection_invalid: root input %q must be a portable source-relative path", entry)
		}
		if err := checkLinkFree(sourceCanonical, entry); err != nil {
			return nil, err
		}
		joined := filepath.Join(sourceCanonical, filepath.FromSlash(entry))
		physical, err := staging.Canonicalize(joined)
		if err != nil {
			return nil, fmt.Errorf("source_member_missing: root input %q: %v", entry, err)
		}
		info, err := os.Lstat(physical) // #nosec G304 -- validated root-input entry
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("source_member_missing: root input %q: %v", entry, err)
			}
			return nil, fmt.Errorf("source_member_invalid: root input %q: %v", entry, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source_selection_invalid: root input %q contains a symlink", entry)
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			return nil, fmt.Errorf("source_member_invalid: root input %q is not a regular file or directory", entry)
		}
		if staging.ContainsGit(physical) {
			return nil, fmt.Errorf("source_output_overlap: root input %q overlaps managed output", entry)
		}
		for _, output := range outputs {
			within, err := staging.Within(output, physical)
			if err != nil {
				return nil, fmt.Errorf("boundary_identity_unreadable: root input %q: %v", entry, err)
			}
			reverse, err := staging.Within(physical, output)
			if err != nil {
				return nil, fmt.Errorf("boundary_identity_unreadable: root input %q: %v", entry, err)
			}
			if within || reverse {
				return nil, fmt.Errorf("source_output_overlap: root input %q overlaps managed output", entry)
			}
		}
		validated = append(validated, validatedRoot{entry: entry, physical: physical})
	}
	for i, first := range validated {
		for _, second := range validated[i+1:] {
			forward, err := staging.Within(first.physical, second.physical)
			if err != nil {
				return nil, fmt.Errorf("boundary_identity_unreadable: root input %q: %v", second.entry, err)
			}
			reverse, err := staging.Within(second.physical, first.physical)
			if err != nil {
				return nil, fmt.Errorf("boundary_identity_unreadable: root input %q: %v", first.entry, err)
			}
			if forward || reverse {
				return nil, fmt.Errorf("source_selection_invalid: root inputs %q and %q overlap", first.entry, second.entry)
			}
		}
	}
	physicals := make([]string, 0, len(validated))
	for _, root := range validated {
		physicals = append(physicals, root.physical)
	}
	if err := checkRootCoverage(sourceCanonical, physicals); err != nil {
		return nil, err
	}
	return physicals, nil
}

// EnumerateInputs lists admitted regular files for capture: the full
// package tree after mandatory pruning, or exactly the root-input
// selections for a project-root package. Managed and generated output
// subtrees, .git metadata and the snapshot/staging trees in outputs are
// pruned deterministically. Links and special files in admitted inputs are
// rejected. Pruning the package root itself or a root-input entry is an
// error, never silent admission of an incomplete set.
func EnumerateInputs(packageRoot string, outputs []string, rootEntries []string) ([]string, error) {
	packageCanonical, err := staging.Canonicalize(packageRoot)
	if err != nil {
		return nil, fmt.Errorf("source_member_invalid: package: %v", err)
	}
	roots := []string{packageCanonical}
	if rootEntries != nil {
		roots = nil
		for _, entry := range rootEntries {
			joined := filepath.Join(packageCanonical, filepath.FromSlash(entry))
			physical, err := staging.Canonicalize(joined)
			if err != nil {
				return nil, fmt.Errorf("source_member_missing: %q: %v", entry, err)
			}
			roots = append(roots, physical)
		}
	}
	var files []string
	for _, root := range roots {
		info, err := os.Lstat(root) // #nosec G304 -- admitted enumeration root
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("source_member_missing: %s: %v", root, err)
			}
			return nil, fmt.Errorf("source_member_invalid: %s: %v", root, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source_member_invalid: %s is a link", root)
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			return nil, fmt.Errorf("source_member_invalid: %s is not a regular file or directory", root)
		}
		if pruned, err := staging.IsOutputPath(root, outputs); err != nil {
			return nil, fmt.Errorf("boundary_identity_unreadable: %s: %v", root, err)
		} else if pruned {
			return nil, fmt.Errorf("source_output_overlap: admitted input %s overlaps managed output", root)
		}
		if info.Mode().IsRegular() {
			files = append(files, root)
			continue
		}
		walked, err := walkPruned(root, outputs)
		if err != nil {
			return nil, err
		}
		files = append(files, walked...)
	}
	sort.Strings(files)
	return files, nil
}

func walkPruned(root string, outputs []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("source_member_invalid: %s: %w", path, err)
		}
		if path != root {
			if pruned, pruneErr := staging.IsOutputPath(path, outputs); pruneErr != nil {
				return fmt.Errorf("boundary_identity_unreadable: %s: %v", path, pruneErr)
			} else if pruned {
				if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source_member_invalid: %s is a link", path)
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source_member_invalid: %s is not a regular file", path)
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func checkLinkFree(sourceRoot, entry string) error {
	current := sourceRoot
	for _, part := range strings.Split(entry, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current) // #nosec G304 -- root-input component under validation
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("source_member_missing: root input %q: %v", entry, err)
			}
			return fmt.Errorf("source_member_invalid: root input %q: %v", entry, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source_selection_invalid: root input %q contains a symlink", entry)
		}
	}
	return nil
}

func checkRootCoverage(sourceRoot string, admitted []string) error {
	covered := func(relative string) (bool, error) {
		target := filepath.Join(sourceRoot, filepath.FromSlash(relative))
		targetCanonical, err := staging.Canonicalize(target)
		if err != nil {
			return false, fmt.Errorf("boundary_identity_unreadable: %s: %v", relative, err)
		}
		for _, root := range admitted {
			within, err := staging.Within(root, targetCanonical)
			if err != nil {
				return false, fmt.Errorf("boundary_identity_unreadable: %s: %v", relative, err)
			}
			if within {
				return true, nil
			}
		}
		return false, nil
	}
	skillFile := filepath.Join(sourceRoot, "SKILL.md")
	if info, err := os.Lstat(skillFile); err != nil || !info.Mode().IsRegular() { // #nosec G304 -- required root input under validation
		if os.IsNotExist(err) {
			return fmt.Errorf("source_member_missing: root inputs must include SKILL.md")
		}
		return fmt.Errorf("source_member_invalid: root SKILL.md: %v", err)
	} else if ok, err := covered("SKILL.md"); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("source_output_overlap: root inputs must include SKILL.md")
	}
	manifest := skillspec.ManifestSourcePath(sourceRoot)
	if manifest != "" {
		if ok, err := covered(manifest); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("source_output_overlap: root inputs must include %s", manifest)
		}
	}
	spec, err := skillspec.Load(sourceRoot)
	if err != nil {
		return fmt.Errorf("source_member_invalid: root manifest: %v", err)
	}
	for _, root := range append(append([]string{}, spec.RuntimeRoots...), spec.BuildRoots...) {
		joined := filepath.Join(sourceRoot, filepath.FromSlash(root))
		if _, err := os.Lstat(joined); os.IsNotExist(err) { // #nosec G304 -- declared root under validation
			return fmt.Errorf("source_member_missing: declared root %q is missing", root)
		}
		if ok, err := covered(root); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("source_output_overlap: root inputs must include declared root %q", root)
		}
	}
	return nil
}

func checkDeclaredInputsDisjoint(packageRoot string, outputs []string) error {
	spec, err := skillspec.Load(packageRoot)
	if err != nil {
		// An unreadable manifest is the member validator's report, not a
		// boundary pass: fail with the member class, never admit blindly.
		return fmt.Errorf("source_member_invalid: manifest: %v", err)
	}
	for _, root := range append(append([]string{}, spec.RuntimeRoots...), spec.BuildRoots...) {
		joined := filepath.Join(packageRoot, filepath.FromSlash(root))
		physical, err := staging.Canonicalize(joined)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: declared root %q: %v", root, err)
		}
		info, err := os.Lstat(physical) // #nosec G304 -- declared input under validation
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("boundary_identity_unreadable: declared root %q: %v", root, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			continue
		}
		if pruned, err := staging.IsOutputPath(physical, outputs); err != nil {
			return fmt.Errorf("boundary_identity_unreadable: declared root %q: %v", root, err)
		} else if pruned {
			return fmt.Errorf("source_output_overlap: declared root %q overlaps managed output", root)
		}
	}
	return nil
}

func sameFile(first, second string) (bool, error) {
	firstInfo, err := os.Stat(first) // #nosec G304 -- boundary identity comparison
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	secondInfo, err := os.Stat(second) // #nosec G304 -- boundary identity comparison
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return os.SameFile(firstInfo, secondInfo), nil
}
