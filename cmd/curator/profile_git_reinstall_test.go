package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestProfileGitReinstallHonoursUseAndTakeover drives the §9.5
// stop-and-retry for a git root through run(): a same-source reinstall of
// an already-installed git root honours --use and --takeover exactly as a
// first install does — it re-resolves as an update and then runs the
// install row's activation — and is reported as an update. All four
// --use/--takeover combinations run on a git root that is not current with
// an unmanaged blocker in the machine scope, and again on a git root that
// is current: eight rows. The git root is served under a fake canonical
// network identity through a GIT_CONFIG_GLOBAL insteadOf rewrite (a file://
// operand is refused by canonicalGit and must never reach this path). A
// mutant that drops the activation call on the git reinstall branch
// restores the silent no-op — exit 0 saying "updated profile" with no
// switch — and must fail the --use rows below.
func TestProfileGitReinstallHonoursUseAndTakeover(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	// One fixture repository for every row: the manager clones per
	// manager home, so rows stay isolated while resolving identically.
	repo := t.TempDir()
	writeContextPackage(t, repo, "groot", "1.0.0", "git module\n")
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init")
	git("add", ".")
	git("commit", "-m", "one")
	git("tag", "v1.0.0")
	gitconfig := filepath.Join(t.TempDir(), "gitconfig")
	rewrite := "[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = https://example.com/groot\n"
	if err := os.WriteFile(gitconfig, []byte(rewrite), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	const operand = "https://example.com/groot"

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
	// setupBlockedReinstall installs a path profile and the git root so
	// that the git root ends current or not as asked, then plants an
	// unmanaged native file no marker records in a home the machine
	// scope writes when it switches.
	setupBlockedReinstall := func(t *testing.T, gitCurrent bool) (stubConfigSource, string) {
		t.Helper()
		source, _ := profileHome(t)
		pinOperatorHome(t)
		other := t.TempDir()
		writeContextPackage(t, other, "other", "1.0.0", "managed other\n")
		if gitCurrent {
			if code, _, stderr := runProfile(t, source, "profile", "install", operand); code != exitOK {
				t.Fatalf("install groot = %d\nstderr:\n%s", code, stderr)
			}
			if code, _, stderr := runProfile(t, source, "profile", "install", other); code != exitOK {
				t.Fatalf("install other = %d\nstderr:\n%s", code, stderr)
			}
		} else {
			if code, _, stderr := runProfile(t, source, "profile", "install", other); code != exitOK {
				t.Fatalf("install other = %d\nstderr:\n%s", code, stderr)
			}
			if code, _, stderr := runProfile(t, source, "profile", "install", operand); code != exitOK {
				t.Fatalf("install groot = %d\nstderr:\n%s", code, stderr)
			}
		}
		if got, want := currentOf(t, source), map[bool]string{true: "groot", false: "other"}[gitCurrent]; got != want {
			t.Fatalf("current=%q, want %q before the reinstall", got, want)
		}
		fresh := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", fresh)
		if err := os.WriteFile(filepath.Join(fresh, "CLAUDE.md"), []byte("OPERATOR HAND-WRITTEN CONTEXT\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return source, fresh
	}
	updatedLine := func(t *testing.T, stdout string) {
		t.Helper()
		lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
		last := lines[len(lines)-1]
		if !strings.HasPrefix(last, "updated profile groot (lock ") || !strings.HasSuffix(last, ")") {
			t.Fatalf("operator line %q, want the reinstall reported as an update", last)
		}
	}

	t.Run("not-current/neither-flag switches nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, false)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand)
		if code != exitOK {
			t.Fatalf("reinstall = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if strings.Contains(stdout, "switched") {
			t.Fatalf("a reinstall without --use must not switch:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want other", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall that switches nothing must write no backup")
		}
	})

	t.Run("not-current/use without takeover attempts the switch and fails loudly", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, false)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--use")
		if code != exitFail {
			t.Fatalf("install --use = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
		}
		updatedLine(t, stdout)
		if !strings.Contains(stderr, "environment_surface_unmanaged_conflict") {
			t.Fatalf("a --use reinstall that meets unmanaged state must name the conflict:\n%s", stderr)
		}
		if !strings.Contains(stderr, "profile_use_partial") {
			t.Fatalf("a --use reinstall that meets unmanaged state must report the partial switch:\n%s", stderr)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want the switch refused and other still current", got)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a refused switch must write no backup")
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
	})

	t.Run("not-current/takeover without use switches nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, false)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--takeover")
		if code != exitOK {
			t.Fatalf("install --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if strings.Contains(stdout, "switched") {
			t.Fatalf("a reinstall without --use must not switch:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "other" {
			t.Fatalf("current=%q, want other: --takeover without --use activates nothing", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall that switches nothing must write no backup")
		}
	})

	t.Run("not-current/use with takeover takes over and activates", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, false)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--use", "--takeover")
		if code != exitOK {
			t.Fatalf("install --use --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if !strings.Contains(stdout, "claude_code: switched") {
			t.Fatalf("the retry must report the switch:\n%s", stdout)
		}
		if !strings.Contains(stderr, "notice:") || !strings.Contains(stderr, ".agent-environment-backup") {
			t.Fatalf("the retry must report the replace notice:\n%s", stderr)
		}
		if got := currentOf(t, source); got != "groot" {
			t.Fatalf("current=%q, want the retry to activate groot", got)
		}
		backup, err := os.ReadFile(filepath.Join(fresh, ".agent-environment-backup", "1", "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(backup) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("backup = %q, err = %v: the takeover must back up before the first write", backup, err)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || !strings.Contains(string(payload), "## Context: groot 1.0.0") {
			t.Fatalf("CLAUDE.md was not materialized from the git root:\n%s", payload)
		}
	})

	t.Run("current/neither-flag switches nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, true)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand)
		if code != exitOK {
			t.Fatalf("reinstall = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if got := currentOf(t, source); got != "groot" {
			t.Fatalf("current=%q, want groot still current", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall of the current root must write no backup")
		}
	})

	t.Run("current/use without takeover stays current and writes nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, true)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--use")
		if code != exitOK {
			t.Fatalf("reinstall --use = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if strings.Contains(stdout, "switched") {
			t.Fatalf("a reinstall of the current root must not re-switch:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "groot" {
			t.Fatalf("current=%q, want groot still current", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall of the current root must write no backup")
		}
	})

	t.Run("current/takeover without use stays current and writes nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, true)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--takeover")
		if code != exitOK {
			t.Fatalf("reinstall --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if got := currentOf(t, source); got != "groot" {
			t.Fatalf("current=%q, want groot still current", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall of the current root must write no backup")
		}
	})

	t.Run("current/use with takeover stays current and writes nothing", func(t *testing.T) {
		source, fresh := setupBlockedReinstall(t, true)
		code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--use", "--takeover")
		if code != exitOK {
			t.Fatalf("reinstall --use --takeover = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		updatedLine(t, stdout)
		if strings.Contains(stdout, "switched") {
			t.Fatalf("a reinstall of the current root must not re-switch:\n%s", stdout)
		}
		if got := currentOf(t, source); got != "groot" {
			t.Fatalf("current=%q, want groot still current", got)
		}
		payload, err := os.ReadFile(filepath.Join(fresh, "CLAUDE.md")) // #nosec G304 -- test home
		if err != nil || string(payload) != "OPERATOR HAND-WRITTEN CONTEXT\n" {
			t.Fatalf("unmanaged bytes = %q, want them untouched", payload)
		}
		if _, err := os.Stat(filepath.Join(fresh, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("a reinstall of the current root must write no backup")
		}
	})
}
