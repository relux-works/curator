package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/skillspec"
	"gopkg.in/yaml.v3"
)

// ExpansionOptions supplies acquisition-owned Git trees and machine-owned
// output boundaries. GitRoots must contain one already acquired tree per alias;
// expansion never fetches, resolves refs, snapshots, or publishes anything.
// OutputRoots includes runtime stores, caches, snapshot stores and staging.
// Project managed paths and .git metadata are additionally excluded internally.
type ExpansionOptions struct {
	GitRoots    map[string]string
	OutputRoots []string
}

// Selection is one skill-level input to acquisition and dependency resolution.
// Index is the original zero-based Skillfile selection. Legacy declarations are
// preserved verbatim and have no Path or Spec. Path is an observation of source
// bytes, NOT an immutable snapshot or permission to install. Acquisition must
// freeze and revalidate both package bytes and this ordered membership first.
type Selection struct {
	Index     int
	Decl      Decl
	Directory string
	Path      string
	Spec      *skillspec.Spec
}

// Expand validates and expands a parsed draft manifest, in declaration order
// and UTF-8 folder byte order within collections. No partial result is returned
// on failure. The returned units retain dependencies in their individual Spec.
func Expand(m *Manifest, opts ExpansionOptions) ([]Selection, error) {
	if m == nil || m.SchemaVersion != 2 {
		return nil, fmt.Errorf("source_selection_invalid: expansion requires draft Skillfile schema 2")
	}
	project, err := filepath.Abs(filepath.Dir(m.Path))
	if err != nil {
		return nil, err
	}
	outputs := append([]string{}, opts.OutputRoots...)
	outputs = append(outputs, filepath.Join(project, ".agents"))
	for _, rel := range adapters.AgentPaths {
		outputs = append(outputs, filepath.Join(project, filepath.FromSlash(rel)))
	}
	for _, output := range outputs {
		if !filepath.IsAbs(output) {
			return nil, fmt.Errorf("source_output_overlap: output boundary must be absolute")
		}
		if err := checkOutputBoundary(output); err != nil {
			return nil, err
		}
	}
	var result []Selection
	seen := map[string]string{}
	appendMember := func(member Selection) error {
		// Portable installed identifiers are ASCII. Case folding also prevents a
		// lock produced on a sensitive host from aliasing destinations on another.
		key := strings.ToLower(member.Decl.Name)
		if previous, ok := seen[key]; ok {
			return fmt.Errorf("source_name_conflict: %s repeats destination %s", member.Decl.Name, previous)
		}
		seen[key] = member.Decl.Name
		result = append(result, member)
		return nil
	}
	for index, decl := range m.Skills {
		if decl.Selector == nil {
			if err := appendMember(Selection{Index: index, Decl: decl, Directory: "."}); err != nil {
				return nil, err
			}
			continue
		}
		s := decl.Selector
		source, ok := m.Sources[s.From]
		if !ok {
			return nil, fmt.Errorf("source_alias_unknown: %s", s.From)
		}
		// Reuse parser validation even for in-memory callers, before touching paths.
		entry := map[string]any{"from": s.From, "directory": s.Directory}
		if s.Collection {
			entry["include"] = memberValues(s.Include)
			entry["exclude"] = memberValues(s.Exclude)
		} else {
			entry["name"] = decl.Name
		}
		if _, err := parseSelector(entry, m.Sources, "selection"); err != nil {
			return nil, err
		}
		root := source.Path
		if root == "" {
			root = opts.GitRoots[s.From]
			if root == "" {
				return nil, fmt.Errorf("source_selection_invalid: alias %s requires an acquired Git tree", s.From)
			}
		} else if !filepath.IsAbs(root) {
			root = filepath.Join(project, root)
		}
		root, err = filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		root, err = filepath.EvalSymlinks(root)
		if err != nil {
			return nil, fmt.Errorf("source_member_invalid: alias %s: %w", s.From, err)
		}
		base, err := selectionDirectory(root, s.Directory, outputs)
		if err != nil {
			return nil, err
		}
		directories := []string{s.Directory}
		if s.Collection {
			folders, err := collectionFolders(root, base, s, outputs)
			if err != nil {
				return nil, err
			}
			directories = nil
			for _, folder := range folders {
				directories = append(directories, path.Join(s.Directory, folder))
			}
		}
		for _, directory := range directories {
			physical, err := selectionDirectory(root, directory, outputs)
			if err != nil {
				return nil, err
			}
			name, spec, err := selectedPackage(physical)
			if err != nil {
				return nil, fmt.Errorf("source_member_invalid: %s:%s: %w", s.From, directory, err)
			}
			if !s.Collection && name != decl.Name {
				return nil, fmt.Errorf("source_member_invalid: %s:%s: SKILL.md name %q differs from %q", s.From, directory, name, decl.Name)
			}
			individual := Decl{Name: name, Selector: &Selector{From: s.From, Directory: directory}}
			if err := appendMember(Selection{Index: index, Decl: individual, Directory: directory, Path: physical, Spec: spec}); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// checkOutputBoundary refuses an output root that lies under a non-directory
// on every platform. A single Stat is insufficient: on Windows a path below a
// regular file reports "not found" instead of ENOTDIR, so the boundary would
// be misread as an absent directory and admitted. Walk existing ancestors with
// Lstat; absent directories remain allowed, any other inspection failure stays
// a cannot-inspect refusal.
func checkOutputBoundary(output string) error {
	current := output
	for {
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				resolved, statErr := os.Stat(current)
				if statErr != nil {
					if os.IsNotExist(statErr) {
						return fmt.Errorf("source_output_overlap: output boundary %s lies under a non-directory %s", output, current)
					}
					return fmt.Errorf("source_output_overlap: cannot inspect output boundary: %w", statErr)
				}
				if !resolved.IsDir() {
					return fmt.Errorf("source_output_overlap: output boundary %s lies under a non-directory %s", output, current)
				}
				return nil
			}
			if !info.IsDir() {
				return fmt.Errorf("source_output_overlap: output boundary %s lies under a non-directory %s", output, current)
			}
			return nil
		}
		if os.IsNotExist(err) {
			parent := filepath.Dir(current)
			if parent == current {
				return nil
			}
			current = parent
			continue
		}
		return fmt.Errorf("source_output_overlap: cannot inspect output boundary: %w", err)
	}
}

func memberValues(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func collectionFolders(root, base string, s *Selector, outputs []string) ([]string, error) {
	selected := map[string]bool{}
	for _, name := range s.Include {
		if name == "*" {
			continue
		}
		// Explicit literals must exist and be directories before exclusion, even
		// when '*' would find the same member.
		if _, err := selectionDirectory(root, path.Join(s.Directory, name), outputs); err != nil {
			return nil, err
		}
		selected[name] = true
	}
	for _, name := range s.Include {
		if name != "*" {
			continue
		}
		entries, err := os.ReadDir(base)
		if err != nil {
			return nil, fmt.Errorf("source_member_invalid: collection %s: %w", s.Directory, err)
		}
		for _, entry := range entries {
			candidate := filepath.Join(base, entry.Name())
			if generatedPath(candidate, outputs) {
				continue
			}
			info, err := os.Stat(candidate)
			if err != nil {
				return nil, fmt.Errorf("source_member_invalid: %s: %w", entry.Name(), err)
			}
			if info.IsDir() {
				selected[entry.Name()] = true
			}
		}
	}
	for _, name := range s.Exclude {
		delete(selected, name)
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("source_selection_invalid: empty collection %s:%s", s.From, s.Directory)
	}
	for _, name := range names {
		if !identifiers.Valid(name) {
			return nil, fmt.Errorf("source_member_invalid: invalid discovered folder %q", name)
		}
	}
	return names, nil
}

func selectionDirectory(root, directory string, outputs []string) (string, error) {
	candidate := filepath.Join(root, filepath.FromSlash(directory))
	if generatedPath(candidate, outputs) {
		return "", fmt.Errorf("source_output_overlap: %s", directory)
	}
	physical, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		class := "source_member_invalid"
		if os.IsNotExist(err) {
			class = "source_member_missing"
		}
		return "", fmt.Errorf("%s: %s: %w", class, directory, err)
	}
	if !physicalWithin(root, physical) {
		return "", fmt.Errorf("source_selection_invalid: %s escapes source", directory)
	}
	if generatedPath(physical, outputs) {
		return "", fmt.Errorf("source_output_overlap: %s", directory)
	}
	info, err := os.Stat(physical)
	if err != nil {
		return "", fmt.Errorf("source_member_invalid: %s: %w", directory, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("source_member_invalid: %s is not a directory", directory)
	}
	return physical, nil
}

// SelectSkillPackage applies the Skillfile individual-selector directory
// grammar and containment checks to a dependency package in a pinned
// repository snapshot. The selected package must have a valid SKILL.md whose
// name matches expectedName. It returns the physical selected directory and
// its parsed skill manifest.
func SelectSkillPackage(root, directory, expectedName string) (string, *skillspec.Spec, error) {
	if !identifiers.ValidDirectory(directory) {
		return "", nil, fmt.Errorf("source_selection_invalid: directory must be a portable contained path or '.'")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", nil, fmt.Errorf("source_member_invalid: cannot resolve repository snapshot: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("source_member_missing: repository snapshot is missing: %w", err)
		}
		return "", nil, fmt.Errorf("source_member_invalid: cannot inspect repository snapshot: %w", err)
	}
	if err := rejectDirectorySymlinks(root, directory); err != nil {
		return "", nil, err
	}
	physical, err := selectionDirectory(root, directory, nil)
	if err != nil {
		return "", nil, err
	}
	name, spec, err := selectedPackage(physical)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("source_member_invalid: selected directory %q has no SKILL.md: %w", directory, err)
		}
		return "", nil, fmt.Errorf("source_member_invalid: %s: %w", directory, err)
	}
	if name != expectedName {
		return "", nil, fmt.Errorf("source_member_invalid: %s: SKILL.md name %q differs from dependency name %q", directory, name, expectedName)
	}
	return physical, spec, nil
}

func rejectDirectorySymlinks(root, directory string) error {
	if directory == "." {
		return nil
	}
	current := root
	for _, component := range strings.Split(directory, "/") {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				continue // selectionDirectory reports the required missing-path diagnostic.
			}
			return fmt.Errorf("source_member_invalid: cannot inspect directory %q: %w", directory, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		physical, err := filepath.EvalSymlinks(current)
		if err == nil && !physicalWithin(root, physical) {
			return fmt.Errorf("source_selection_invalid: directory %q escapes source", directory)
		}
		// Dependency directory selection requires a real, link-free folder.
		// Mention escape in the diagnostic as well: a dangling or otherwise
		// unresolvable link cannot establish containment and fails closed.
		return fmt.Errorf("source_selection_invalid: directory %q escapes source or contains a symlink", directory)
	}
	return nil
}

// SameFile ancestry uses the filesystem's actual case equivalence, including
// case-insensitive macOS volumes, rather than a GOOS-based string prefix.
func physicalWithin(root, candidate string) bool {
	rootInfo, err := os.Stat(root)
	if err != nil {
		return false
	}
	for current := candidate; ; current = filepath.Dir(current) {
		if info, err := os.Stat(current); err == nil && os.SameFile(rootInfo, info) {
			return true
		}
		if filepath.Dir(current) == current {
			return false
		}
	}
}

func generatedPath(candidate string, outputs []string) bool {
	if physical, err := filepath.EvalSymlinks(candidate); err == nil {
		candidate = physical
	}
	for current := candidate; ; current = filepath.Dir(current) {
		if strings.EqualFold(filepath.Base(current), ".git") {
			return true
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	for _, output := range outputs {
		if physicalWithin(output, candidate) {
			return true
		}
	}
	return false
}

func selectedPackage(dir string) (string, *skillspec.Spec, error) {
	file := filepath.Join(dir, "SKILL.md")
	info, err := os.Lstat(file)
	if err != nil {
		return "", nil, err
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("SKILL.md must be a regular file")
	}
	data, err := os.ReadFile(file) // #nosec G304 -- contained selected package; fixed metadata filename
	if err != nil {
		return "", nil, err
	}
	lines := bytes.Split(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")), []byte("\n"))
	if len(lines) < 3 || string(lines[0]) != "---" {
		return "", nil, fmt.Errorf("missing YAML frontmatter")
	}
	end := 1
	for end < len(lines) && string(lines[end]) != "---" {
		end++
	}
	if end == len(lines) {
		return "", nil, fmt.Errorf("unterminated YAML frontmatter")
	}
	var metadata map[string]any
	if err := yaml.Unmarshal(bytes.Join(lines[1:end], []byte("\n")), &metadata); err != nil {
		return "", nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	name, ok := metadata["name"].(string)
	if !ok || !identifiers.Valid(name) {
		return "", nil, fmt.Errorf("invalid SKILL.md name")
	}
	description, ok := metadata["description"].(string)
	if !ok || strings.TrimSpace(description) == "" {
		return "", nil, fmt.Errorf("invalid SKILL.md description")
	}
	if raw, ok := metadata["triggers"]; ok {
		triggers, ok := raw.([]any)
		if !ok {
			return "", nil, fmt.Errorf("invalid triggers")
		}
		for _, raw := range triggers {
			value, ok := raw.(string)
			if !ok || strings.TrimSpace(value) == "" {
				return "", nil, fmt.Errorf("invalid trigger")
			}
		}
	}
	modern := false
	for _, file := range []string{skillspec.CanonicalManifestName, skillspec.LegacyManifestName} {
		present, err := regularMetadata(dir, file)
		if err != nil {
			return "", nil, err
		}
		modern = modern || present
	}
	if !modern {
		if _, err := regularMetadata(dir, skillspec.RuntimeFallbackName); err != nil {
			return "", nil, err
		}
	}
	spec, err := skillspec.Load(dir)
	if err != nil {
		return "", nil, err
	}
	// Schema 1 permits an optional identity extension. Later schemas reject
	// unknown identity fields in skillspec.Load; do not invent one for them.
	if spec.SourceFile == skillspec.CanonicalManifestName || spec.SourceFile == skillspec.LegacyManifestName {
		payload, err := os.ReadFile(filepath.Join(dir, spec.SourceFile)) // #nosec G304 -- filename chosen by skillspec.Load
		if err != nil {
			return "", nil, err
		}
		var object map[string]any
		if err := json.Unmarshal(payload, &object); err != nil {
			return "", nil, err
		}
		if declared, present := object["name"]; present && declared != name {
			return "", nil, fmt.Errorf("skill manifest name differs from SKILL.md")
		}
	}
	return name, spec, nil
}

// Metadata is read during expansion, before snapshot capture. Refuse links and
// special files here too, rather than reading outside the package or blocking
// on a FIFO. Only actual absence permits the ordinary manifest fallback.
func regularMetadata(root, relative string) (bool, error) {
	current := root
	parts := strings.Split(relative, "/")
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return false, fmt.Errorf("metadata path %s contains a symlink", relative)
		}
		if i == len(parts)-1 {
			if !info.Mode().IsRegular() {
				return false, fmt.Errorf("metadata %s must be a regular file", relative)
			}
		} else if !info.IsDir() {
			return false, fmt.Errorf("metadata parent %s is not a directory", relative)
		}
	}
	return true, nil
}
