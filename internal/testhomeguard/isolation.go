package testhomeguard

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckIsolation verifies the neutral helper's boundary without implementing
// another isolation mechanism. Old helpers fail before m.Run. Until the shared
// helper owns provider/XDG overrides too, ambient overrides fail closed rather
// than remaining live; callers may explicitly clear them at the outer runner.
func CheckIsolation(previousHome string) error {
	home, err := os.UserHomeDir()
	parentVar := nativeParentEnvVar()
	if err != nil || !filepath.IsAbs(home) || home == previousHome || parentVar == "" || os.Getenv(parentVar) == "" {
		return fmt.Errorf("testhomeguard: neutral testgitenv HOME isolation required")
	}
	for _, name := range nativeAuxiliaryEnvVars() {
		if os.Getenv(name) != "" {
			return fmt.Errorf("testhomeguard: neutral testgitenv isolation leaves an ambient native-home override")
		}
	}
	return nil
}
