package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapAndAddProject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "ru", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := Bootstrap(path, "/other", "", nil, false); err == nil {
		t.Fatal("bootstrap must preserve an existing config without --force")
	}
	project := t.TempDir()
	if err := AddProject(path, "app", project, []string{"claude_code"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PreferredLocale != "ru" || cfg.Projects["app"].Path != project {
		t.Fatalf("config = %+v", cfg)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("config mode = %o, want private", info.Mode().Perm())
	}
}
