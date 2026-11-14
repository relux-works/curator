package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// A repair may seed from a native home only when its physical location is
// outside managed environment ownership. This checks paths, never credentials.
func (req *ResolveRequest) validateRepairNativeHome(adapter envregistry.Adapter) error {
	if adapter.ID == envregistry.OpenCode {
		if _, err := req.externalOperatorXDG(); err != nil {
			return err
		}
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		// Windows private-root creation requires an exact directory entry,
		// rather than a root link. Preserve that admission before publication.
		state, err := stateread.Lstat(native)
		if err != nil {
			return fmt.Errorf("%s: native home cannot be inspected: %w", envregistry.DiagCredentialConflict, err)
		}
		switch state.Kind {
		case stateread.KindAbsent:
		case stateread.KindPresent:
			if state.Info == nil || !state.Info.IsDir() {
				return fmt.Errorf("%s: native home must be a direct directory on Windows", envregistry.DiagCredentialConflict)
			}
		default:
			return fmt.Errorf("%s: native home has unusable read state %q", envregistry.DiagCredentialConflict, state.Kind)
		}
	}
	managedPath, inspectErr := nativePathIsManaged(native, req.Home)
	if inspectErr != nil {
		return nativeHomeRefusal(native, inspectErr)
	}
	if managedPath {
		return nativeHomeRefusal(native, fmt.Errorf("inside managed environments"))
	}
	native, err = repairDirectoryLocation(native)
	if err != nil {
		return fmt.Errorf("%s: native home location cannot be established: %w", envregistry.DiagCredentialConflict, err)
	}
	managed, err := repairDirectoryLocation(EnvRoot(req.Home))
	if err != nil {
		return fmt.Errorf("%s: managed home location cannot be established: %w", envregistry.DiagCredentialConflict, err)
	}
	relative, err := filepath.Rel(managed, native)
	if err == nil && (relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return nil
	}
	if err != nil && filepath.VolumeName(managed) != filepath.VolumeName(native) {
		return nil
	}
	return fmt.Errorf("%s: native home %s is inside managed environments; run repair with the intended external native home", envregistry.DiagCredentialConflict, native)
}

// Missing ordinary directories are allowed. Resolve the existing ancestor
// before appending the missing suffix; a broken or unreadable symlink is
// unknown, not a missing ordinary directory.
func repairDirectoryLocation(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			state, statErr := stateread.Stat(resolved)
			if statErr != nil {
				return "", statErr
			}
			if state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.IsDir() {
				return "", fmt.Errorf("directory location %s is not a readable directory", path)
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		state, statErr := stateread.Lstat(path)
		if statErr != nil {
			return "", statErr
		}
		if state.Kind != stateread.KindAbsent {
			return "", fmt.Errorf("directory location %s is unresolved or changed", path)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		suffix = append(suffix, filepath.Base(path))
		path = parent
	}
}
