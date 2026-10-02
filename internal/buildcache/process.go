package buildcache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Keep both the reported path and its canonical spelling (notably /var and
// /private/var on macOS). A disappeared executable still protects its reported
// build; every other resolution failure is uncertainty, never proof of absence.
func normalizeExecutablePaths(paths []string) ([]string, error) {
	var normalized []string
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return nil, fmt.Errorf("process image path is not absolute: %q", path)
		}
		normalized = append(normalized, filepath.Clean(path))
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("resolve process image %q: %w", path, err)
		}
		if err == nil {
			normalized = append(normalized, resolved)
		}
	}
	return normalized, nil
}

func buildInUse(path, canonical string, executables []string) bool {
	for _, executable := range executables {
		for _, dir := range []string{path, canonical} {
			if runtime.GOOS == "windows" {
				dir, executable = strings.ToLower(dir), strings.ToLower(executable)
			}
			// A separator boundary prevents a sibling with a shared prefix
			// from retaining the wrong build. Any file below the build counts,
			// regardless of command name or daemon registration.
			if strings.HasPrefix(executable, filepath.Clean(dir)+string(filepath.Separator)) {
				return true
			}
		}
	}
	return false
}
