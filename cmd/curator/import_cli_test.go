package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests drive the production run() entry point for the stage (c)
// rows: profile import, profile compose of a path source, and the
// takeover flags.

// pinOperatorHome points the operator home at a temporary directory so the
// opencode global skills surface resolves below test control on every
// platform.
func pinOperatorHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// TestProfileImportRow drives `profile import` through run(): with empty
// native homes the lossless import installs and activates the imported
// profile, and the list reports the imported marker.
func TestProfileImportRow(t *testing.T) {
	source, _ := profileHome(t)
	pinOperatorHome(t)
	if code, stdout, stderr := runProfile(t, source, "profile", "import"); code != exitOK {
		t.Fatalf("import = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	} else if !strings.Contains(stdout, "installed and activated profile imported") {
		t.Fatalf("stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "list")
	if code != exitOK {
		t.Fatalf("list = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "imported") {
		t.Fatalf("list stdout:\n%s", stdout)
	}
}

// TestProfileImportLossyRow drives `profile import` over an unrecoverable
// skills entry: the row stops with environment_import_lossy and proceeds
// only under --allow-lossy, re-reporting the list as warnings.
func TestProfileImportLossyRow(t *testing.T) {
	source, _ := profileHome(t)
	pinOperatorHome(t)
	skills := filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "skills", "mystery")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "SKILL.md"), []byte("# mystery\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "profile", "import"); code != exitFail {
		t.Fatalf("import = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	} else if !strings.Contains(stderr, "environment_import_lossy") {
		t.Fatalf("stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "import", "--allow-lossy")
	if code != exitOK {
		t.Fatalf("import --allow-lossy = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	} else if !strings.Contains(stderr, "environment_import_lossy") {
		t.Fatalf("consent must re-report the loss list as warnings:\n%s", stderr)
	}
}

// TestProfileImportNameTakenRow drives `profile import --as` with a taken
// name: the row stops with profile_import_name_taken before any write.
func TestProfileImportNameTakenRow(t *testing.T) {
	source, _ := profileHome(t)
	pinOperatorHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "legacy", "1.0.0", "legacy\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg, "--as", "imported"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "import"); code != exitFail {
		t.Fatalf("import = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	} else if !strings.Contains(stderr, "profile_import_name_taken") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestProfileComposeAddPathRow drives `profile compose add` of a path
// source: the bare source is accepted, the list shows the path form, and
// a requirement flag with a path source is a usage error.
func TestProfileComposeAddPathRow(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	overlay := t.TempDir()
	if code, _, stderr := runProfile(t, source, "profile", "compose", "acme", "add", overlay); code != exitOK {
		t.Fatalf("compose add = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	code, stdout, stderr := runProfile(t, source, "profile", "compose", "acme", "list")
	if code != exitOK {
		t.Fatalf("compose list = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, overlay) || !strings.Contains(stdout, "path") {
		t.Fatalf("list stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if code, _, _ := runProfile(t, source, "profile", "compose", "acme", "add", overlay, "--range", "*"); code != exitUsage {
		t.Fatalf("compose add of a path with --range = %d, want %d", code, exitUsage)
	}
}

// TestResolveTakeoverRequiresRepair drives `env resolve --takeover`
// without --repair: the flag applies only with --repair, so the row is a
// usage error.
func TestResolveTakeoverRequiresRepair(t *testing.T) {
	source, _ := profileHome(t)
	code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--takeover")
	if code != exitUsage {
		t.Fatalf("resolve --takeover = %d, want %d\nstderr:\n%s", code, exitUsage, stderr)
	}
	if !strings.Contains(stderr, "--takeover applies only with --repair") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestSyncTakeoverFlagParses drives `profile sync --takeover` through
// run(): the flag is accepted on the sync row.
func TestSyncTakeoverFlagParses(t *testing.T) {
	source, _ := profileHome(t)
	if code, _, stderr := runProfile(t, source, "profile", "sync", "--takeover"); code != exitOK {
		t.Fatalf("sync --takeover = %d\nstderr:\n%s", code, stderr)
	}
}
