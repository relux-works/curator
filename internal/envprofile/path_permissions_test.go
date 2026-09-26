package envprofile

import (
	"os"
	"testing"
)

func makeEnvprofileTestPathUnreadable(t *testing.T, path string) func() {
	t.Helper()
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("this host cannot create a mode-000 file: %v", err)
	}
	if _, err := os.ReadFile(path); err == nil {
		_ = os.Chmod(path, 0o644)
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
	return func() {
		if err := os.Chmod(path, 0o644); err != nil {
			t.Errorf("restore test configuration permissions: %v", err)
		}
	}
}
