package main

import (
	"os"
	"path/filepath"
	"testing"
)

// An inherited GIT_SSH never selects the production wrapper; Curator's own
// executable supplies both manager helpers. No SSH process is started here.
func TestProductionGitToolIgnoresInheritedSSHWrapper(t *testing.T) {
	inherited := filepath.Join(t.TempDir(), "foreign-wrapper")
	if err := os.WriteFile(inherited, []byte("foreign"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_SSH", inherited)
	tool := productionGitTool()
	manager, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	manager = admittedOperatorFile(manager)
	if tool.Executable == "" || manager == "" || tool.SSHWrapper != manager || tool.AskPass != manager {
		t.Fatalf("manager helpers not bound: %+v", tool)
	}
}
