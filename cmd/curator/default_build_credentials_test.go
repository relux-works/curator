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
	// GIT_SSH is deliberately not set in the host process; the production
	// selector has no environment lookup for it.
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

// A project git first on PATH cannot supply even a discovery probe. The fixed
// system-selected Git and manager helpers remain the valid positive control.
func TestProductionGitToolRejectsProjectPATH(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(bin, "git")
	if err := os.WriteFile(fake, []byte("foreign program"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("CURATOR_BUILD_GIT_EXECUTABLE", "")
	tool := productionGitTool()
	if tool.Executable == "" || tool.Executable == fake || tool.Fingerprint == "" {
		t.Fatalf("fixed search failed: %+v", tool)
	}
	t.Setenv("CURATOR_BUILD_GIT_EXECUTABLE", fake)
	if tool := productionGitTool(); tool.Executable != "" {
		t.Fatalf("project git admitted: %+v", tool)
	}
	if got, _ := os.ReadFile(fake); string(got) != "foreign program" {
		t.Fatal("foreign bytes changed")
	}
}
