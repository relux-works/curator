//go:build windows

package gitops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Qualified by TestWindowsGitMetadataLimit against actual clone and reopen
// operations, not by assuming that Win32's nominal MAX_PATH is usable.
const windowsGitMetadataLimit = 259

func checkCloneDestination(destination string) error {
	root, err := expandedWindowsClonePath(destination)
	if err != nil {
		return fmt.Errorf("resolve Windows Git metadata root: %w", err)
	}
	return checkMetadataPathLength(root, windowsGitMetadataLimit)
}

// Expand existing 8.3 ancestors before appending nonexistent components. Git's
// current directory can expand a short alias, so its spelling is not a budget.
// This performs no mkdir, clone, permission repair or path redirection.
func expandedWindowsClonePath(destination string) (string, error) {
	absolute, err := filepath.Abs(destination)
	if err != nil {
		return "", err
	}
	ancestor := absolute
	var missing []string
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !os.IsNotExist(err) && !errors.Is(err, windows.ERROR_FILENAME_EXCED_RANGE) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", fmt.Errorf("no existing clone ancestor")
		}
		missing = append(missing, filepath.Base(ancestor))
		ancestor = parent
	}
	// Charge both spellings: a long junction name cannot borrow its short
	// target's budget, and a short junction cannot hide a long target.
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	root, err := expandedWindowsAncestor(ancestor)
	if err != nil {
		return "", err
	}
	resolved, err = expandedWindowsAncestor(resolved)
	if err != nil {
		return "", err
	}
	if utf16PathLength(resolved) > utf16PathLength(root) {
		root = resolved
	}
	for i := len(missing) - 1; i >= 0; i-- {
		root = filepath.Join(root, missing[i])
	}
	return root, nil
}

func expandedWindowsAncestor(ancestor string) (string, error) {
	p, err := windows.UTF16PtrFromString(ancestor)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetLongPathName(p, &buf[0], uint32(len(buf)))
	if err != nil {
		return "", err
	}
	if n == 0 || int(n) >= len(buf) {
		return "", fmt.Errorf("expanded clone ancestor exceeds path buffer")
	}
	root := windows.UTF16ToString(buf[:n])
	if strings.HasPrefix(root, `\\?\UNC\`) {
		root = `\\` + root[len(`\\?\UNC\`):]
	} else {
		root = strings.TrimPrefix(root, `\\?\`)
	}
	return root, nil
}
