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
// source: every overlay carries exactly one requirement form (§12.1), so
// the bare source is a usage error while a path source with an exact form
// is accepted and listed.
func TestProfileComposeAddPathRow(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	overlay := t.TempDir()
	if code, _, _ := runProfile(t, source, "profile", "compose", "acme", "add", overlay); code != exitUsage {
		t.Fatalf("bare compose add = %d, want %d", code, exitUsage)
	}
	revision := strings.Repeat("ab", 20)
	if code, _, stderr := runProfile(t, source, "profile", "compose", "acme", "add", overlay, "--revision", revision); code != exitOK {
		t.Fatalf("compose add = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	code, stdout, stderr := runProfile(t, source, "profile", "compose", "acme", "list")
	if code != exitOK {
		t.Fatalf("compose list = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, overlay) || !strings.Contains(stdout, "revision="+revision) {
		t.Fatalf("list stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

// TestProfileUseTakeoverRow drives `profile use` over an unmanaged native
// file: without the flag the row fails rather than overwrite; with
// --takeover it converges with the replace notice and a backup.
func TestProfileUseTakeoverRow(t *testing.T) {
	source, _ := profileHome(t)
	pinOperatorHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "managed\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	fresh := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", fresh)
	if err := os.WriteFile(filepath.Join(fresh, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "acme"); code != exitFail {
		t.Fatalf("use = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	} else if !strings.Contains(stderr, "environment_surface_unmanaged_conflict") {
		t.Fatalf("stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "use", "acme", "--takeover")
	if code != exitOK {
		t.Fatalf("use --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	} else if !strings.Contains(stderr, "notice:") || !strings.Contains(stderr, ".agent-environment-backup") {
		t.Fatalf("takeover must report the replace notice:\n%s", stderr)
	}
	backup, err := os.ReadFile(filepath.Join(fresh, ".agent-environment-backup", "1", "CLAUDE.md"))
	if err != nil || string(backup) != "mine\n" {
		t.Fatalf("backup = %q, err = %v", backup, err)
	}
}

// TestProfileInstallReinstallHonoursUseAndTakeover drives the §9.5
// stop-and-retry through run(): a `profile install --use` stopped by an
// unmanaged native file publishes the lock but leaves the machine current
// unchanged, and the documented retry `profile install <same path> --as
// <same name> --use --takeover` must take over with backup and activate —
// never report success for no work. All four --use/--takeover combinations
// run on a non-current path root with an unmanaged blocker, plus the
// current-root reinstall with an edited source. A mutant that drops the
// reinstall activation for the already-installed (identical-lock) case must
// fail this test.
func TestProfileInstallReinstallHonoursUseAndTakeover(t *testing.T) {
	setupBlockedReinstall := func(t *testing.T) (stubConfigSource, string, string) {
		t.Helper()
		source, _ := profileHome(t)
		pinOperatorHome(t)
		tk := t.TempDir()
		writeContextPackage(t, tk, "tk", "1.0.0", "managed tk\n")
		if code, _, stderr := runProfile(t, source, "profile", "install", tk, "--as", "tk"); code != exitOK {
			t.Fatalf("install tk = %d\nstderr:\n%s", code, stderr)
		}
		other := t.TempDir()
		writeContextPackage(t, other, "other", "1.0.0", "managed other\n")
		if code, _, stderr := runProfile(t, source, "profile", "install", other, "--as", "other", "--use"); code != exitOK {
			t.Fatalf("install other --use = %d\nstderr:\n%s", code, stderr)
		}
		// An unmanaged native file no marker records, in a home the
		// machine scope will write when it switches to tk.
		fresh := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", fresh)
		if err := os.WriteFile(filepath.Join(fresh, "CLAUDE.md"), []byte("OPERATOR HAND-WRITTEN CONTEXT\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return source, tk, fresh
	}
	currentOf := func(t *testing.T, source stubConfigSource) string {
		t.Helper()
		code, stdout, stderr := runProfile(t, source, "profile", "list")
		if code != exitOK {
			t.Fatalf("list = %d\nstderr:\n%s", code, stderr)
		}
		for _, line := range strings.Split(stdout, "\n") {
			fields := strings.Split(line, "\t")
			if len(fields) > 0 && strings.Contains(line, "current") {
				return fields[0]
			}
		}
		return "<none>"
	}

	t.Run("use without takeover attempts the switch and fails loudly", func(t *testing.T) {
		source, tk, fresh := setupBlockedReinstall(t)
		code, _, stderr := runProfile(t, source, "profile", "install", tk, "--as", "tk", "--use")
		if code != exitFail {
			t.Fatalf("install --use = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
		}
		if !strings.Contains(stderr, "environment_surface_unmanaged_conflict") {
			t.Fatalf("a --use reinstall that meets unmanaged state must name the conflict:\n%s", stderr)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want the switch refused and other still current", got)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a refused switch must write no backup")
		}
	})

	t.Run("takeover without use switches nothing", func(t *testing.T) {
		source, tk, fresh := setupBlockedReinstall(t)
		code, stdout, stderr := runProfile(t, source, "profile", "install", tk, "--as", "tk", "--takeover")
		if code != exitOK {
			t.Fatalf("install --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stdout, "updated profile tk") {
			t.Fatalf("a same-source reinstall reports an update:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want other: --takeover without --use activates nothing", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
	})

	t.Run("neither flag switches nothing", func(t *testing.T) {
		source, tk, _ := setupBlockedReinstall(t)
		code, stdout, stderr := runProfile(t, source, "profile", "install", tk, "--as", "tk")
		if code != exitOK {
			t.Fatalf("reinstall = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stdout, "updated profile tk") {
			t.Fatalf("a same-source reinstall reports an update:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want other", got)
		}
	})

	t.Run("use with takeover takes over and activates", func(t *testing.T) {
		source, tk, fresh := setupBlockedReinstall(t)
		code, stdout, stderr := runProfile(t, source, "profile", "install", tk, "--as", "tk", "--use", "--takeover")
		if code != exitOK {
			t.Fatalf("install --use --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stderr, "notice:") || !strings.Contains(stderr, ".agent-environment-backup") {
			t.Fatalf("the retry must report the replace notice:\n%s", stderr)
		}
		if got := currentOf(t, source); got != "tk" {
			t.Fatalf("current=%q, want the retry to activate tk", got)
		}
		backup, err := os.ReadFile(filepath.Join(fresh, ".agent-environment-backup", "1", "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(backup) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("backup = %q, err = %v: the takeover must back up before the first write", backup, err)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || !strings.Contains(string(payload), "curator-root-context-v2") {
			t.Fatalf("CLAUDE.md was not materialized from the store:\n%s", payload)
		}
	})

	t.Run("current root reinstall refreshes the pin and stays current", func(t *testing.T) {
		source, _ := profileHome(t)
		pinOperatorHome(t)
		src := t.TempDir()
		writeContextPackage(t, src, "tk", "1.0.0", "original\n")
		if code, _, stderr := runProfile(t, source, "profile", "install", src, "--as", "tk"); code != exitOK {
			t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
		}
		if got := currentOf(t, source); got != "tk" {
			t.Fatalf("current=%q, want the first install to activate tk", got)
		}
		if err := os.WriteFile(filepath.Join(src, "context", "a.md"), []byte("EDITED AFTER INSTALL\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := runProfile(t, source, "profile", "install", src, "--as", "tk", "--use", "--takeover")
		if code != exitOK {
			t.Fatalf("reinstall --use --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stdout, "updated profile tk") {
			t.Fatalf("a same-source reinstall reports an update:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "tk" {
			t.Fatalf("current=%q, want tk still current", got)
		}
		payload, err := os.ReadFile(filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || !strings.Contains(string(payload), "EDITED AFTER INSTALL") {
			t.Fatalf("the reinstall must re-materialize the edited bytes:\n%s", payload)
		}
	})
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
