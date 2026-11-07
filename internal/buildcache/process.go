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

// A native Linux snapshot additionally records inode identities and whether an
// foreign-user image was omitted. Such omissions are safe only for
// independently protected, owner-only executable candidates. Test overrides
// and other platforms retain their existing path-only contract.
type executableSnapshot struct {
	paths           []string
	images          []os.FileInfo
	foreignExcluded bool
}

func (snapshot executableSnapshot) retains(info os.FileInfo) bool {
	if info == nil {
		return true
	}
	if snapshot.foreignExcluded && (info.Mode().Perm()&0o077 != 0 || info.Mode()&(os.ModeSetuid|os.ModeSetgid) != 0) {
		return true
	}
	for _, image := range snapshot.images {
		if os.SameFile(info, image) {
			return true
		}
	}
	return false
}
