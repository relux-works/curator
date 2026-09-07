package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envmarker"
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
// source through run(): the requirement form is required only for a git
// source (manager-config-v2 $defs/overlay, environments §1 and §12.1), so
// the bare path source is accepted, the written row parses back as a path
// declaration, and the list reports it in the `path` form column — the
// branch that was dead while every overlay needed a form.
func TestProfileComposeAddPathRow(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	overlay := t.TempDir()
	if code, stdout, stderr := runProfile(t, source, "profile", "compose", "acme", "add", overlay); code != exitOK {
		t.Fatalf("bare compose add = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	source = reloadSource(t, source)
	decls := source.cfg.Env.Overlays["acme"]
	if len(decls) != 1 || decls[0].Source != overlay {
		t.Fatalf("overlays: %+v", source.cfg.Env.Overlays)
	}
	if decls[0].Range != "" || decls[0].Tag != "" || decls[0].Revision != "" || decls[0].Directory != "" {
		t.Fatalf("path overlay carries a requirement form: %+v", decls[0])
	}
	code, stdout, stderr := runProfile(t, source, "profile", "compose", "acme", "list")
	if code != exitOK {
		t.Fatalf("compose list = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, overlay) || !strings.Contains(stdout, "\tpath\t") {
		t.Fatalf("list stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

// TestProfileComposeAddRefusesAFormOnAPathSource drives the other arm
// through run(): section 1 makes a range, tag, revision, or directory on a
// path declaration profile_source_invalid, so the row refuses it rather
// than writing a declaration no resolution can accept. A git source is
// driven beside it in both directions, so a mutant that merely deletes the
// kind switch admits a form-free git row and fails here.
func TestProfileComposeAddRefusesAFormOnAPathSource(t *testing.T) {
	overlay := t.TempDir()
	revision := strings.Repeat("ab", 20)
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"path with range", []string{overlay, "--range", "^1.2"}, exitUsage},
		{"path with tag", []string{overlay, "--tag", "v1.0.0"}, exitUsage},
		{"path with revision", []string{overlay, "--revision", revision}, exitUsage},
		{"path with directory", []string{overlay, "--directory", "sub"}, exitUsage},
		{"relative path with tag", []string{"packages/team-context", "--tag", "v1.0.0"}, exitUsage},
		{"relative path bare", []string{"packages/team-context"}, exitOK},
		{"git bare", []string{"https://example.com/org/pkg"}, exitUsage},
		{"git two forms", []string{"https://example.com/org/pkg", "--range", "^1", "--tag", "v1"}, exitUsage},
		{"git with range", []string{"https://example.com/org/pkg", "--range", "^1.2"}, exitOK},
		{"scp git bare", []string{"git@example.com:org/pkg"}, exitUsage},
		{"scp git with tag", []string{"git@example.com:org/pkg", "--tag", "v1.0.0"}, exitOK},
		{"unknown scheme", []string{"svn://example.com/org/pkg", "--range", "^1"}, exitUsage},
		{"bare drive letter", []string{"D:"}, exitUsage},
		{"scp host grammar", []string{"git@my_host:org/pkg", "--range", "^1"}, exitUsage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
			args := append([]string{"profile", "compose", "acme", "add"}, tc.args...)
			code, stdout, stderr := runProfile(t, source, args...)
			if code != tc.want {
				t.Fatalf("compose add %v = %d, want %d\nstdout:\n%s\nstderr:\n%s", tc.args, code, tc.want, stdout, stderr)
			}
			if tc.want != exitOK {
				return
			}
			// An accepted row must parse back: the reader and this row
			// share one discriminator, so the kind cannot change between
			// writing the declaration and reading it.
			reloadSource(t, source)
		})
	}
}

// TestPathOverlayFromMachineConfigJoinsTheClosure is the §6 promise driven
// end to end through run(): a path overlay declared in machine
// configuration joins the closure beside the root, carries its declared
// weight, and is recorded as an overlay member in the environment marker.
// Before the landed rule no operator could declare this row at all.
func TestPathOverlayFromMachineConfigJoinsTheClosure(t *testing.T) {
	pinOperatorHome(t)
	root := t.TempDir()
	writeContextPackage(t, root, "acme", "1.0.0", "root\n")
	overlay := t.TempDir()
	writeContextPackage(t, overlay, "personal", "0.3.0", "overlay\n")
	encoded, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {},
		"environments": {"overlays": {"acme": [{"source": `+string(encoded)+`, "weight": 250}]}}}`)
	if code, stdout, stderr := runProfile(t, source, "profile", "install", root, "--use"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	marker, err := envmarker.Read(os.Getenv("CLAUDE_CONFIG_DIR"))
	if err != nil || marker == nil {
		t.Fatalf("marker %+v %v", marker, err)
	}
	var found *envmarker.Member
	for index := range marker.Members {
		if marker.Members[index].Name == "personal" {
			found = &marker.Members[index]
		}
	}
	if found == nil {
		t.Fatalf("marker members %+v carry no overlay", marker.Members)
	}
	if !found.Overlay {
		t.Fatalf("overlay member %+v is not flagged overlay", *found)
	}
	if found.Weight != 250 {
		t.Fatalf("overlay weight %d, want the declared 250", found.Weight)
	}
	if found.StateSHA256 == "" || found.Commit != "" {
		t.Fatalf("path overlay member %+v carries no state pin", *found)
	}
	payload, err := os.ReadFile(filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "## Context: personal 0.3.0") {
		t.Fatalf("document:\n%s", payload)
	}
}

// TestOverlayFromMachineConfigIsRefusedByKind drives the refusals through
// run() from the same surface: a path overlay carrying a requirement form
// and a source that is neither kind are both refused before the load
// succeeds, so no such declaration can reach resolution.
func TestOverlayFromMachineConfigIsRefusedByKind(t *testing.T) {
	overlay := t.TempDir()
	encoded, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	revision := strings.Repeat("ab", 20)
	cases := []struct {
		name string
		row  string
	}{
		{"path with revision", `{"source": ` + string(encoded) + `, "revision": "` + revision + `"}`},
		{"path with range", `{"source": ` + string(encoded) + `, "range": "^1.2"}`},
		{"path with directory", `{"source": ` + string(encoded) + `, "directory": "sub"}`},
		{"relative path with tag", `{"source": "packages/team-context", "tag": "v1.0.0"}`},
		{"git with no form", `{"source": "https://example.com/org/pkg"}`},
		{"unknown scheme", `{"source": "svn://example.com/org/pkg", "range": "^1"}`},
		{"file url", `{"source": "file:///tmp/pkg"}`},
		{"bare drive letter", `{"source": "C:"}`},
		{"scp host grammar", `{"source": "git@my_host:org/pkg", "range": "^1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, _ := profileHome(t)
			path := filepath.Join(t.TempDir(), "config.json")
			text := `{"schema_version": 2, "skills_root": "x", "projects": {},
				"environments": {"overlays": {"acme": [` + tc.row + `]}}}`
			if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, loadErr := config.Load(path, nil)
			if loadErr == nil {
				t.Fatalf("config.Load accepted %s", tc.row)
			}
			// The refusal reaches the operator through run(): the config
			// source hands run() exactly the reader's error, so every row
			// stops there rather than resolving a declaration no kind
			// admits.
			source.path = path
			source.cfg = cfg
			source.err = loadErr
			code, _, stderr := runProfile(t, source, "profile", "list")
			if code == exitOK {
				t.Fatalf("run() accepted %s\nstderr:\n%s", tc.row, stderr)
			}
			if !strings.Contains(stderr, "environments.overlays") {
				t.Fatalf("run() stderr does not name the overlay row:\n%s", stderr)
			}
		})
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
