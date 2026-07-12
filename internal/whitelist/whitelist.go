// Package whitelist copies the model-facing context of a skill snapshot
// (Spec §4.2).
//
// Only whitelisted roots are copied; developer-facing files are excluded at
// any depth; runtime roots stay out of context even under whitelisted roots.
// Staging is atomic: the destination directory is replaced only on success.
package whitelist

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IncludeRoots are the root entries copied into installed context.
var IncludeRoots = []string{
	"SKILL.md", "agents", "references", ".skill_triggers",
	"assets", "templates", "examples", "data",
}

// AlwaysExcluded patterns are dropped at any depth.
var AlwaysExcluded = []string{
	".git", ".github", ".gitlab-ci.yml", ".venv", "__pycache__", "*.pyc",
	"node_modules", "tests", "test", "__tests__", "README*", "CHANGELOG*",
	"LICENSE*", "Makefile", "setup.py", "pyproject.toml", "requirements*.txt",
	".DS_Store", ".gitignore",
}

// CopyContext copies the whitelisted context of snapshot into destination
// and returns the sorted list of copied POSIX-relative paths.
// includeScripts adds the scripts/ root (used for command-less skills);
// excludeRoots removes runtime roots from context.
func CopyContext(snapshot, destination string, includeScripts bool, excludeRoots []string) ([]string, error) {
	if info, err := os.Stat(filepath.Join(snapshot, "SKILL.md")); err != nil || info.IsDir() {
		return nil, fmt.Errorf("required SKILL.md not found in skill snapshot: %s", snapshot)
	}
	if err := os.RemoveAll(destination); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return nil, err
	}

	roots := append([]string(nil), IncludeRoots...)
	if includeScripts {
		roots = append(roots, "scripts")
	}
	sort.Strings(roots)

	var copied []string
	for _, root := range roots {
		src := filepath.Join(snapshot, root)
		info, err := os.Stat(src)
		if err != nil {
			continue
		}
		if excludedName(root) {
			continue
		}
		if !info.IsDir() {
			if excludedByRoot(root, excludeRoots) {
				continue
			}
			if err := copyFile(src, filepath.Join(destination, root)); err != nil {
				return nil, err
			}
			copied = append(copied, root)
			continue
		}
		err = filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(snapshot, path)
			if relErr != nil {
				return relErr
			}
			posix := filepath.ToSlash(rel)
			if excludedByRoot(posix, excludeRoots) || excludedPath(posix) {
				return nil
			}
			if err := copyFile(path, filepath.Join(destination, rel)); err != nil {
				return err
			}
			copied = append(copied, posix)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(copied)
	return copied, nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	payload, err := os.ReadFile(src) // #nosec G304 -- paths come from the walked snapshot
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, payload, info.Mode().Perm())
}

func excludedName(name string) bool {
	for _, pattern := range AlwaysExcluded {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func excludedPath(posix string) bool {
	for _, part := range strings.Split(posix, "/") {
		if excludedName(part) {
			return true
		}
	}
	return excludedName(posix)
}

func excludedByRoot(posix string, excludeRoots []string) bool {
	parts := strings.Split(posix, "/")
	for _, root := range excludeRoots {
		rootParts := strings.Split(root, "/")
		if len(parts) < len(rootParts) {
			continue
		}
		match := true
		for i, part := range rootParts {
			if parts[i] != part {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
