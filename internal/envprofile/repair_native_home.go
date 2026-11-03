package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
)

// A repair may seed from a native home only when its physical location is
// outside managed environment ownership. This checks paths, never credentials.
func (req *ResolveRequest) validateRepairNativeHome(adapter envregistry.Adapter) error {
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return err
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
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		info, statErr := os.Lstat(path)
		if statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symbolic link location %s is unresolved", path)
		}
		if statErr != nil && !os.IsNotExist(statErr) {
			return "", statErr
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		suffix = append(suffix, filepath.Base(path))
		path = parent
	}
}
