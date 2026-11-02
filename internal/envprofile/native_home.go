package envprofile

import (
	"fmt"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/stateread"
	"os"
	"path/filepath"
)

// NativeHome resolves the adapter's native default home. The mechanism
// variable names the home, except for opencode where it names the XDG
// parent and the tool reads the opencode child (environments §7.1). The
// four defaults below the user home are implementation choices the
// specification's mechanism implies but does not spell; every test pins
// homes through the variables instead.
func NativeHome(adapter Adapter) (string, error) {
	if adapter.ID == "opencode" {
		if value := os.Getenv(adapter.EnvVar); value != "" {
			return filepath.Join(value, "opencode"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
		}
		return filepath.Join(home, ".config", "opencode"), nil
	}
	if value := os.Getenv(adapter.EnvVar); value != "" {
		return value, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
	}
	return filepath.Join(home, adapter.DefaultDir), nil
}

// Native surface custody is independent of credential-origin selection. An
// ambient selector into managed custody refuses instead of becoming an in-place
// write or borrowing that home's record. No HOME/selector substitution occurs.
func resolveNativeHome(manager string, adapter Adapter, resolver func(string) (string, error)) (string, error) {
	var native string
	var err error
	if resolver != nil {
		native, err = resolver(adapter.ID)
	} else {
		native, err = NativeHome(adapter)
	}
	if err != nil {
		return "", err
	}
	if err := envorigin.ValidateExternal(manager, native); err != nil {
		return "", err
	}
	return native, nil
}

// nativeSurfacePath is a read-only planning join for a proven absent root.
// Existing roots retain the normal no-follow child checks. No directory is
// created here; exact parent metadata and physical boundary pins belong to the
// complete unit journal. Errors and broken links are never absence.
func nativeSurfacePath(manager, root, rel string) (string, error) {
	parts, err := managedPathParts(rel)
	if err != nil {
		return "", err
	}
	if err := envorigin.ValidateExternal(manager, root); err != nil {
		return "", err
	}
	state, err := stateread.Lstat(root)
	if err != nil {
		return "", err
	}
	if state.Kind == stateread.KindAbsent {
		return filepath.Join(append([]string{root}, parts...)...), nil
	}
	return managedPath(root, rel, false)
}
