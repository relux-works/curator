package whitelist

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
)

// FileSelection adds relative file globs to the built-in skill context roots.
// Includes widen that set and excludes narrow it. A nil selection preserves
// the legacy context walk exactly.
type FileSelection struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

var errInvalidFileSelection = errors.New("skill files selection invalid")

// ValidateFileSelection checks the portable, relative glob grammar used by
// profile skill declarations. Globs use POSIX separators and path.Match
// metacharacters; a directory match applies to its descendants.
func ValidateFileSelection(selection *FileSelection) error {
	if selection == nil {
		return nil
	}
	for _, list := range [][]string{selection.Include, selection.Exclude} {
		seen := map[string]bool{}
		for _, pattern := range list {
			if err := validatePattern(pattern); err != nil {
				return err
			}
			if seen[pattern] {
				return fmt.Errorf("%w: repeated pattern", errInvalidFileSelection)
			}
			seen[pattern] = true
		}
	}
	for _, include := range selection.Include {
		for _, exclude := range selection.Exclude {
			if include == exclude {
				return fmt.Errorf("%w: include and exclude patterns conflict", errInvalidFileSelection)
			}
		}
	}
	for _, pattern := range selection.Exclude {
		for _, protected := range []string{"SKILL.md", "agent-skill.json", "csk-skill.json", "agents/runtime.json"} {
			if matchesAnyPrefix([]string{pattern}, protected) {
				return fmt.Errorf("%w: required skill file cannot be excluded", errInvalidFileSelection)
			}
		}
	}
	return nil
}

// ValidateFileSet checks a canonical effective context list suitable for a
// portable profile lock.
func ValidateFileSet(files []string) error {
	if files == nil {
		return fmt.Errorf("%w: missing effective file set", errInvalidFileSelection)
	}
	for i, file := range files {
		if !identifiers.PortablePath(file) || i > 0 && files[i-1] >= file {
			return fmt.Errorf("%w: invalid or noncanonical effective file set", errInvalidFileSelection)
		}
	}
	index := sort.SearchStrings(files, "SKILL.md")
	if index >= len(files) || files[index] != "SKILL.md" {
		return fmt.Errorf("%w: SKILL.md is required context", errInvalidFileSelection)
	}
	return nil
}

// ContextFiles returns the canonical effective files from a skill package.
// With no selection it delegates to the legacy walker and changes no matching
// or safety behavior. A selection can add paths from the skill directory,
// while runtime/build and fixed developer exclusions remain in force.
func ContextFiles(snapshot string, includeScripts bool, excludeRoots []string, selection *FileSelection) ([]string, error) {
	if err := ValidateFileSelection(selection); err != nil {
		return nil, err
	}
	files := map[string]struct{}{}
	err := WalkContext(snapshot, includeScripts, excludeRoots, func(_, posix string) error {
		files[posix] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if selection == nil {
		return sortedFiles(files), nil
	}

	root, err := filepath.EvalSymlinks(snapshot)
	if err != nil {
		return nil, err
	}
	matchedInclude := false
	err = filepath.WalkDir(snapshot, func(pathname string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if pathname == snapshot {
			return nil
		}
		rel, err := filepath.Rel(snapshot, pathname)
		if err != nil {
			return err
		}
		posix := canonicalContextPath(snapshot, filepath.ToSlash(rel))
		if entry.IsDir() && (PathExcluded(posix, excludeRoots) || SelectionExcludesDirectory(selection, posix)) {
			deleteFileTree(files, posix)
			return filepath.SkipDir
		}
		if PathExcluded(posix, excludeRoots) || SelectionExcludesFile(selection, posix) {
			deleteFileTree(files, posix)
			return nil
		}
		if !entry.IsDir() && matchesAnyPrefix(selection.Include, posix) {
			files[posix] = struct{}{}
			matchedInclude = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(selection.Include) > 0 && !matchedInclude {
		return nil, fmt.Errorf("profile_skill_files_no_match: no included file exists")
	}
	result := sortedFiles(files)
	if err := ValidateFileSet(result); err != nil {
		return nil, err
	}
	for _, file := range result {
		if err := validateSelectedFile(root, snapshot, file, excludeRoots, selection); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// CopyContextFiles copies exactly the pinned effective file set. It is used
// after planning so hashing and publication consume the same path list.
func CopyContextFiles(snapshot, destination string, files []string, excludeRoots []string, selection *FileSelection) error {
	if err := ValidateFileSelection(selection); err != nil {
		return err
	}
	if err := ValidateFileSet(files); err != nil {
		return err
	}
	root, err := filepath.EvalSymlinks(snapshot)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	for _, posix := range files {
		if err := validateSelectedFile(root, snapshot, posix, excludeRoots, selection); err != nil {
			return err
		}
		if err := copyFile(filepath.Join(snapshot, filepath.FromSlash(posix)), filepath.Join(destination, filepath.FromSlash(posix))); err != nil {
			return err
		}
	}
	return nil
}

func validatePattern(pattern string) error {
	if pattern == "" || len(pattern) > 4096 || !utf8.ValidString(pattern) || path.IsAbs(pattern) || strings.ContainsAny(pattern, "\\:") || strings.IndexFunc(pattern, unicode.IsControl) >= 0 {
		return fmt.Errorf("%w: pattern must be a portable relative glob", errInvalidFileSelection)
	}
	for _, segment := range strings.Split(pattern, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("%w: pattern escapes or aliases the skill directory", errInvalidFileSelection)
		}
	}
	if _, err := path.Match(pattern, ""); err != nil {
		return fmt.Errorf("%w: malformed glob", errInvalidFileSelection)
	}
	return nil
}

func validRelativeFile(file string) bool {
	if file == "" || !utf8.ValidString(file) || path.IsAbs(file) || strings.ContainsAny(file, "\\\x00:") || strings.IndexFunc(file, unicode.IsControl) >= 0 {
		return false
	}
	for _, segment := range strings.Split(file, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func matchesPattern(pattern, file string) bool {
	matched, err := path.Match(pattern, file)
	return err == nil && matched
}

func matchesAnyPrefix(patterns []string, file string) bool {
	for candidate := file; candidate != "."; candidate = path.Dir(candidate) {
		for _, pattern := range patterns {
			if matchesPattern(pattern, candidate) {
				return true
			}
		}
	}
	return false
}

// SelectionExcludesFile reports whether an exclusion removes a relative file
// path, including an exclusion of one of its parent directories.
func SelectionExcludesFile(selection *FileSelection, file string) bool {
	return selection != nil && matchesAnyPrefix(selection.Exclude, file)
}

// SelectionExcludesDirectory reports whether the selection excludes this
// directory and all of its descendants.
func SelectionExcludesDirectory(selection *FileSelection, directory string) bool {
	if selection == nil {
		return false
	}
	if SelectionExcludesFile(selection, directory) {
		return true
	}
	for _, pattern := range selection.Exclude {
		if pattern == directory+"/**" {
			return true
		}
	}
	return false
}

func deleteFileTree(files map[string]struct{}, directory string) {
	for file := range files {
		if file == directory || strings.HasPrefix(file, directory+"/") {
			delete(files, file)
		}
	}
}

func sortedFiles(files map[string]struct{}) []string {
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}

func canonicalContextPath(snapshot, posix string) string {
	root, rest, nested := strings.Cut(posix, "/")
	for _, known := range append(append([]string(nil), IncludeRoots...), "scripts") {
		if !strings.EqualFold(root, known) {
			continue
		}
		actualInfo, actualErr := os.Stat(filepath.Join(snapshot, root))
		canonicalInfo, canonicalErr := os.Stat(filepath.Join(snapshot, known))
		if actualErr == nil && canonicalErr == nil && os.SameFile(actualInfo, canonicalInfo) {
			root = known
			break
		}
	}
	if nested {
		return root + "/" + rest
	}
	return root
}

// CanonicalContextPath applies the portable root spelling used by the shared
// context walker to one relative path.
func CanonicalContextPath(snapshot, posix string) string {
	return canonicalContextPath(snapshot, posix)
}

func validateSelectedFile(root, snapshot, posix string, excludeRoots []string, selection *FileSelection) error {
	if !validRelativeFile(posix) {
		return fmt.Errorf("%w: selected path escapes the skill directory", errInvalidFileSelection)
	}
	selected := filepath.Join(snapshot, filepath.FromSlash(posix))
	resolved, err := filepath.EvalSymlinks(selected)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("%w: selected path escapes the skill directory", errInvalidFileSelection)
	}
	resolvedPath := canonicalContextPath(snapshot, filepath.ToSlash(rel))
	if PathExcluded(resolvedPath, excludeRoots) || selection != nil && matchesAnyPrefix(selection.Exclude, resolvedPath) {
		return fmt.Errorf("%w: selected path resolves into an excluded path", errInvalidFileSelection)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%w: selected path is not a regular file", errInvalidFileSelection)
	}
	return nil
}
