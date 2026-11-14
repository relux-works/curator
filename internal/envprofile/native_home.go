package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

const nativeHomeFallbackNotice = "Managed caller home is not a native origin; using the operator HOME default. This does not recover the prior native selector."

func nativeManagerHome() string { return filepath.Dir(config.UserPath()) }

// nativeHomeAt excludes the active manager's environments root before choosing
// a seed or credential origin. The fallback is a current selection, never a
// reconstruction of an earlier selector. No environment variable is changed.
func nativeHomeAt(adapter Adapter, managerHome string) (string, bool, error) {
	return nativeHomeSelection(adapter, managerHome, os.Getenv(adapter.EnvVar))
}

func nativeHomeRefusal(path string, cause error) error {
	return fmt.Errorf("%s: native home %s is not an external native origin: %v; run repair with the intended external native home", envregistry.DiagCredentialConflict, path, cause)
}

func nativeHomeSelection(adapter Adapter, managerHome, value string) (string, bool, error) {
	fallback := false
	if value != "" {
		if _, err := nativeComparablePath(value); err != nil {
			return "", false, nativeHomeRefusal(value, err)
		}
		candidate := value
		if adapter.ID == "opencode" || adapter.ID == envregistry.Muse {
			candidate = filepath.Join(value, adapter.ID)
		}
		managed, err := nativePathIsManaged(candidate, managerHome)
		if err != nil {
			return "", false, nativeHomeRefusal(candidate, err)
		}
		if !managed {
			if adapter.ID == "opencode" || adapter.ID == envregistry.Muse {
				return filepath.Join(value, adapter.ID), false, nil
			}
			return value, false, nil
		}
		fallback = true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
	}
	if _, err := nativeComparablePath(home); err != nil {
		return "", false, nativeHomeRefusal(home, err)
	}
	native := filepath.Join(home, adapter.DefaultDir)
	if adapter.ID == "opencode" || adapter.ID == envregistry.Muse {
		native = filepath.Join(home, ".config", adapter.ID)
	}
	managed, err := nativePathIsManaged(native, managerHome)
	if err != nil {
		return "", false, nativeHomeRefusal(native, err)
	}
	if managed {
		return "", false, fmt.Errorf("%s: native home %s is inside managed environments; run repair with the intended external native home", envregistry.DiagCredentialConflict, native)
	}
	return native, fallback, nil
}

func nativePathIsManaged(path, managerHome string) (bool, error) {
	return nativePathManagedBoundary(path, managerHome, true)
}

// XDG seed parents are containers, not linked origins. They may contain the
// manager home; every selected child still passes the full overlap check.
func nativePathInsideManaged(path, managerHome string) (bool, error) {
	return nativePathManagedBoundary(path, managerHome, false)
}

func nativePathManagedBoundary(path, managerHome string, refuseAncestor bool) (bool, error) {
	candidate, err := nativeComparablePath(path)
	if err != nil {
		return false, err
	}
	root, err := nativeComparablePath(EnvRoot(managerHome))
	if err != nil {
		return false, err
	}
	rootState, err := stateread.Stat(root)
	if err != nil {
		return false, err
	}
	if refuseAncestor {
		contains, err := nativePathContains(candidate, root, rootState.Kind == stateread.KindPresent)
		if err != nil {
			return false, err
		}
		if contains {
			return false, fmt.Errorf("native origin contains managed environments")
		}
	}
	if rootState.Kind == stateread.KindPresent {
		// EvalSymlinks preserves component case on some filesystems. Compare
		// actual ancestors to the managed root instead of case-folding names:
		// case-distinct directories must remain distinct on filesystems that
		// support them. Missing candidate tails are inspected without creation.
		for current := candidate; ; current = filepath.Dir(current) {
			state, err := stateread.Stat(current)
			if err != nil {
				return false, err
			}
			if state.Kind == stateread.KindPresent && os.SameFile(rootState.Info, state.Info) {
				return true, nil
			}
			if filepath.Dir(current) == current {
				return false, nil
			}
		}
	}
	// No managed root exists yet. Preserve lexical containment of its
	// unresolved tail after canonicalizing the existing ancestors above.
	_, err = filepath.Rel(root, candidate)
	if err != nil {
		// Different Windows volumes cannot contain each other.
		if !strings.EqualFold(filepath.VolumeName(root), filepath.VolumeName(candidate)) {
			return false, nil
		}
		return false, err
	}
	inside := nativeLexicalPrefix(root, candidate)
	if !inside {
		foldedOverlap := nativeLexicalPrefix(strings.ToLower(root), strings.ToLower(candidate)) || (refuseAncestor && nativeLexicalPrefix(strings.ToLower(candidate), strings.ToLower(root)))
		if foldedOverlap {
			return false, fmt.Errorf("ambiguous case identity below absent managed root")
		}
	}

	return inside, nil
}

// Missing native directories are valid first-provisioning inputs. Resolve the
// nearest existing ancestor, preserving the absent tail, while refusing broken
// links, loops and inspection failures instead of interpreting them as absence.
func nativeComparablePath(path string) (string, error) {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == rune(filepath.Separator) }) {
		if part == ".." {
			return "", fmt.Errorf("parent traversal is not a native selector")
		}
	}
	current, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var tail []string
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			parts := append([]string{resolved}, tail...)
			return filepath.Join(parts...), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		state, inspectErr := stateread.Lstat(current)
		if inspectErr != nil {
			return "", inspectErr
		}
		if state.Kind == stateread.KindPresent {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		tail = append([]string{filepath.Base(current)}, tail...)
		current = parent
	}
}

// Check physical ancestors when the root exists, and unresolved component
// boundaries when it does not. Equality is managed, not a strict ancestor.
func nativePathContains(candidate, root string, rootPresent bool) (bool, error) {
	state, err := stateread.Stat(candidate)
	if err != nil {
		return false, err
	}
	if state.Kind == stateread.KindPresent {
		for current := filepath.Dir(root); ; current = filepath.Dir(current) {
			parent, err := stateread.Stat(current)
			if err != nil {
				return false, err
			}
			if parent.Kind == stateread.KindPresent && os.SameFile(state.Info, parent.Info) {
				return true, nil
			}
			if filepath.Dir(current) == current {
				break
			}
		}
	}
	if rootPresent {
		return false, nil
	}
	rel, err := filepath.Rel(candidate, root)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// Keep component case explicit even on Windows, where filepath.Rel folds it.
// Volume spelling is platform-defined and does not establish component identity.
func nativeLexicalPrefix(root, candidate string) bool {
	normalize := func(path string) string {
		volume := filepath.VolumeName(path)
		return strings.ToUpper(volume) + path[len(volume):]
	}
	root, candidate = normalize(root), normalize(candidate)
	if root == candidate {
		return true
	}
	if !strings.HasSuffix(root, string(filepath.Separator)) {
		root += string(filepath.Separator)
	}
	return strings.HasPrefix(candidate, root)
}
