package claudememory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/stateread"
)

// AncestorFiles names the three Claude instruction forms in every strict
// ancestor of projectRoot. It includes lexical and canonical spellings, but
// never a path within either project-root spelling or the managed root file.
// launchDir preserves ancestor spellings when an alias enters a nested project
// directory without providing a lexical spelling of the containing Git root.
// Only path metadata is inspected; instruction contents are never read.
func AncestorFiles(projectRoot, launchDir, managedRoot string) ([]string, error) {
	if !filepath.IsAbs(projectRoot) || !filepath.IsAbs(launchDir) || !filepath.IsAbs(managedRoot) {
		return nil, fmt.Errorf("claude memory roots must be absolute")
	}
	projectRoot = filepath.Clean(projectRoot)
	canonicalRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve Claude project root: %w", err)
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("claude project root must be an existing directory")
	}
	canonicalLaunch, err := filepath.EvalSymlinks(launchDir)
	if err != nil || !withinMemoryRoot(canonicalRoot, canonicalLaunch) {
		return nil, fmt.Errorf("claude launch directory must be within the project root")
	}
	if launchInfo, err := os.Stat(canonicalLaunch); err != nil || !launchInfo.IsDir() {
		return nil, fmt.Errorf("claude launch directory must be an existing directory")
	}
	canonicalManaged, err := canonicalMemoryPath(managedRoot)
	if err != nil {
		return nil, err
	}
	roots := []string{projectRoot, canonicalRoot}
	seen := map[string]bool{}
	var files []string
	for _, root := range append(append([]string{}, roots...), filepath.Clean(launchDir)) {
		for dir := filepath.Dir(root); dir != root; dir = filepath.Dir(dir) {
			for _, name := range []string{"CLAUDE.md", filepath.Join(".claude", "CLAUDE.md"), "CLAUDE.local.md"} {
				lexical := filepath.Join(dir, name)
				canonical, err := canonicalMemoryPath(lexical)
				if err != nil {
					return nil, err
				}
				// An ancestor link into the project or managed policy is still
				// allowed policy; excluding its alias could hide that policy.
				protected := canonical == canonicalManaged || lexical == filepath.Clean(managedRoot)
				for _, candidate := range []string{lexical, canonical} {
					for _, project := range roots {
						protected = protected || withinMemoryRoot(project, candidate)
					}
				}
				if protected {
					continue
				}
				for _, candidate := range []string{lexical, canonical} {
					if !seen[candidate] {
						seen[candidate] = true
						files = append(files, candidate)
					}
				}
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return files, nil
}

func withinMemoryRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		return true
	}
	// EvalSymlinks alone need not normalize case or Unicode spellings on a
	// case-insensitive filesystem. Metadata identity protects those aliases too.
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return false
	}
	for candidate := path; ; candidate = filepath.Dir(candidate) {
		if info, err := os.Stat(candidate); err == nil && os.SameFile(rootInfo, info) {
			return true
		}
		if filepath.Dir(candidate) == candidate {
			return false
		}
	}
}

// MissingAncestorFiles names paths without their required exact literal entry.
// A user glob may also cover a path; this is policy reconciliation evidence,
// not an assertion that the vendor has loaded that file.
func MissingAncestorFiles(payload []byte, paths []string) ([]string, error) {
	_, excludes, err := exclusionSettings(payload)
	if err != nil {
		return nil, err
	}
	present := map[string]bool{}
	for _, pattern := range excludes {
		present[pattern] = true
	}
	var missing []string
	for _, path := range paths {
		if !present[literalPattern(filepath.ToSlash(path))] {
			missing = append(missing, path)
		}
	}
	return missing, nil
}

// ValidateNativeTargets preserves the native exception without letting an
// alias turn it into an exclusion of AGENTS, the composed root, or an entire
// project. Such a collision refuses; it is never repaired by changing links.
func ValidateNativeTargets(nativeFile, projectRoot, managedRoot string) error {
	native, err := canonicalMemoryPath(nativeFile)
	if err != nil {
		return err
	}
	rules, err := canonicalMemoryPath(filepath.Join(filepath.Dir(nativeFile), "rules"))
	if err != nil {
		return err
	}
	managed, err := canonicalMemoryPath(managedRoot)
	if err != nil {
		return err
	}
	project, err := canonicalMemoryPath(projectRoot)
	if err != nil {
		return err
	}
	if filepath.Base(native) == "AGENTS.md" || native == managed || withinMemoryRoot(rules, managed) || withinMemoryRoot(rules, project) {
		return fmt.Errorf("operator-native Claude memory aliases protected project or managed policy")
	}
	return nil
}

// Missing instruction files are still excluded: a later-created loader must
// not bypass the policy. Resolve the nearest existing parent without reading
// file contents. An existing unresolved link refuses instead of being guessed.
func canonicalMemoryPath(path string) (string, error) {
	path = filepath.Clean(path)
	prefix := path
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(prefix)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve Claude memory path: %w", err)
		}
		metadata, statErr := stateread.Lstat(prefix)
		if statErr != nil {
			return "", fmt.Errorf("inspect Claude memory path: %w", statErr)
		}
		if metadata.Kind != stateread.KindAbsent {
			return "", fmt.Errorf("claude memory path contains an unresolved link or unreadable entry")
		}
		parent := filepath.Dir(prefix)
		if parent == prefix {
			return "", fmt.Errorf("claude memory path has no resolvable parent")
		}
		missing = append(missing, filepath.Base(prefix))
		prefix = parent
	}
}
