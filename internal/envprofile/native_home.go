package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/stateread"
)

const nativeHomeFallbackNotice = "Managed caller home is not a native origin; using the operator HOME default. This does not recover the prior native selector."

func nativeManagerHome() string { return filepath.Dir(config.UserPath()) }

// nativeHomeAt excludes the active manager's environments root before choosing
// a seed or credential origin. The fallback is a current selection, never a
// reconstruction of an earlier selector. No environment variable is changed.
func nativeHomeAt(adapter Adapter, managerHome string) (string, bool, error) {
	value := os.Getenv(adapter.EnvVar)
	fallback := false
	if value != "" {
		candidate := value
		if adapter.ID == "opencode" {
			candidate = filepath.Join(value, "opencode")
		}
		managed, err := nativePathIsManaged(candidate, managerHome)
		if err != nil {
			return "", false, fmt.Errorf("cannot establish native home for %s: %w", adapter.ID, err)
		}
		if !managed {
			if adapter.ID == "opencode" {
				return filepath.Join(value, "opencode"), false, nil
			}
			return value, false, nil
		}
		fallback = true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fallback, fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
	}
	native := filepath.Join(home, adapter.DefaultDir)
	if adapter.ID == "opencode" {
		native = filepath.Join(home, ".config", "opencode")
	}
	managed, err := nativePathIsManaged(native, managerHome)
	if err != nil {
		return "", fallback, fmt.Errorf("cannot establish native default for %s: %w", adapter.ID, err)
	}
	if managed {
		return "", fallback, fmt.Errorf("native HOME default for %s is inside the managed environments root", adapter.ID)
	}
	return native, fallback, nil
}

func nativePathIsManaged(path, managerHome string) (bool, error) {
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
	// An absent root still has an existing physical ancestor. Rebase both
	// unresolved tails on that identity so aliases of the manager directory
	// cannot change containment before the environments directory is created.
	anchor := filepath.Dir(root)
	for {
		state, err := stateread.Stat(anchor)
		if err != nil {
			return false, err
		}
		if state.Kind == stateread.KindPresent {
			for current := candidate; ; current = filepath.Dir(current) {
				candidateState, err := stateread.Stat(current)
				if err != nil {
					return false, err
				}
				if candidateState.Kind == stateread.KindPresent && os.SameFile(state.Info, candidateState.Info) {
					rootTail, err := filepath.Rel(anchor, root)
					if err != nil {
						return false, err
					}
					candidateTail, err := filepath.Rel(current, candidate)
					if err != nil {
						return false, err
					}
					return nativeMissingTailContains(current, rootTail, candidateTail)
				}
				if filepath.Dir(current) == current {
					return false, nil
				}
			}
		}
		parent := filepath.Dir(anchor)
		if parent == anchor {
			return false, fmt.Errorf("managed root has no observable physical ancestor")
		}
		anchor = parent
	}
}

func nativeMissingTailContains(anchor, rootTail, candidateTail string) (bool, error) {
	rel, err := filepath.Rel(rootTail, candidateTail)
	if err != nil {
		return false, err
	}
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true, nil
	}
	rootParts := strings.Split(rootTail, string(filepath.Separator))
	candidateParts := strings.Split(candidateTail, string(filepath.Separator))
	if len(candidateParts) >= len(rootParts) && strings.EqualFold(rootTail, filepath.Join(candidateParts[:len(rootParts)]...)) {
		// Existing differently spelled children prove distinct identities when
		// the root spelling is absent. With both spellings absent, no identity
		// establishes filesystem case semantics: refuse instead of guessing.
		state, err := stateread.Stat(filepath.Join(anchor, candidateParts[0]))
		if err != nil {
			return false, err
		}
		if state.Kind == stateread.KindAbsent {
			return false, fmt.Errorf("cannot establish case identity of absent managed-root tails")
		}
	}
	return false, nil
}

// Missing native directories are valid first-provisioning inputs. Resolve the
// nearest existing ancestor, preserving the absent tail, while refusing broken
// links, loops and inspection failures instead of interpreting them as absence.
func nativeComparablePath(path string) (string, error) {
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
