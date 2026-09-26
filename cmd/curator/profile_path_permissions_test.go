package main

import (
	"os"
	"testing"
)

func makeProfileTestPathUnreadable(t *testing.T, path string, directory bool) func() {
	t.Helper()
	kind := "file"
	mode := os.FileMode(0o644)
	if directory {
		kind = "directory"
		mode = 0o755
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("this host cannot create a mode-000 %s: %v", kind, err)
	}
	readable := func() bool {
		if directory {
			_, err := os.ReadDir(path)
			return err == nil
		}
		_, err := os.ReadFile(path)
		return err == nil
	}
	if readable() {
		_ = os.Chmod(path, mode)
		t.Skipf("this environment can read a mode-000 %s; unreadability is untestable here", kind)
	}
	return func() {
		if err := os.Chmod(path, mode); err != nil {
			t.Errorf("restore %s permissions: %v", path, err)
		}
	}
}
