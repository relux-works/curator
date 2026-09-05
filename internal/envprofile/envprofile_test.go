package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/envmarker"
)

// Production entry points under test: Install, List, Update, Remove, Use,
// Sync, Current, EnsureDefault.

func writePackage(t *testing.T, root, name, version, module string) {
	t.Helper()
	manifest := `{"schema_version": 1, "name": "` + name + `", "version": "` + version + `",` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	if err := os.MkdirAll(filepath.Join(root, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "agent-context.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "context", "a.md"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
}

func pinHomes(t *testing.T) map[string]string {
	t.Helper()
	base := t.TempDir()
	homes := map[string]string{
		"claude_code": filepath.Join(base, "claude"),
		"codex_cli":   filepath.Join(base, "codex"),
		"opencode":    filepath.Join(base, "xdg", "opencode"),
		"pi":          filepath.Join(base, "pi"),
	}
	t.Setenv("CLAUDE_CONFIG_DIR", homes["claude_code"])
	t.Setenv("CODEX_HOME", homes["codex_cli"])
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg"))
	t.Setenv("PI_CODING_AGENT_DIR", homes["pi"])
	return homes
}

// TestInstallPathAndList drives the production Install for a path operand:
// the profile resolves, audits, locks, stores, and lists with its hash.
func TestInstallPathAndList(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	info, activated, updated, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Fatal("fresh install must not report updated")
	}
	if info.Name != "acme" || !activated {
		t.Fatalf("install %+v activated=%v", info, activated)
	}
	profiles, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 { // acme plus the builtin default
		t.Fatalf("profiles %+v", profiles)
	}
	if profiles[0].Name != "acme" || len(profiles[0].LockHash) == 0 {
		t.Fatalf("profiles %+v", profiles)
	}
	current, err := Current(home)
	if err != nil {
		t.Fatal(err)
	}
	if current != "acme" {
		t.Fatalf("first install must activate, current=%q", current)
	}
}

// TestInstallRejectsSecretMember narrows the always-strict gate: a member
// carrying secret material fails installation. A mutant that skips the
// audit must fail this test.
func TestInstallRejectsSecretMember(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	writePackage(t, source, "evil", "1.0.0", "key AKIA1234567890ABCDEF here\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err == nil {
		t.Fatal("secret member must fail installation")
	} else if !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("error %v carries no %s", err, DiagSourceInvalid)
	}
	if _, err := readSource(home, "evil"); err == nil {
		t.Fatal("failed install must leave no profile record")
	}
}

// TestReinstallSameSourceIsAnUpdate checks re-installing an installed path
// with the same requirement re-resolves as an update.
func TestReinstallSameSourceIsAnUpdate(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, activated, updated, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	if !updated || activated {
		t.Fatalf("reinstall %+v activated=%v updated=%v", info, activated, updated)
	}
}

// TestNameTaken narrows the name gate: a different source under an
// installed name fails with profile_name_taken.
func TestNameTaken(t *testing.T) {
	home := t.TempDir()
	first, second := t.TempDir(), t.TempDir()
	writePackage(t, first, "acme", "1.0.0", "hello\n")
	writePackage(t, second, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := Install(home, InstallOptions{Operand: second, As: "acme"})
	if err == nil || !strings.Contains(err.Error(), DiagNameTaken) {
		t.Fatalf("error %v carries no %s", err, DiagNameTaken)
	}
}

// TestUseMaterializesLinkedHomes drives the production Use across the whole
// machine scope: the claude root context is a copied file, the others are
// store links, and every home carries a marker for the profile.
func TestUseMaterializesLinkedHomes(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	results, err := Use(home, "acme", "", "", false)
	if err != nil {
		t.Fatalf("use: %v (%+v)", err, results)
	}
	if len(results) != len(Adapters) {
		t.Fatalf("results %+v", results)
	}
	claude := filepath.Join(homes["claude_code"], "CLAUDE.md")
	payload, err := os.ReadFile(claude) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "curator-root-context-v2") || !strings.Contains(string(payload), "## Context: acme 1.0.0") {
		t.Fatalf("claude document:\n%s", payload)
	}
	codexLink := filepath.Join(homes["codex_cli"], "AGENTS.md")
	if info, err := os.Lstat(codexLink); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("codex AGENTS.md is not a link: %v", err)
	}
	for id, dir := range homes {
		marker, err := envmarker.Read(dir)
		if err != nil {
			t.Fatal(err)
		}
		if marker == nil || marker.Profile.Name != "acme" || marker.Mode != envmarker.ModeLinked {
			t.Fatalf("%s marker %+v", id, marker)
		}
	}
	current, _ := Current(home)
	if current != "acme" {
		t.Fatalf("current=%q", current)
	}
}

// TestUseReplacesWithBackup checks versioned backups: switching profiles
// backs up the replaced file into .agent-environment-backup/1/.
func TestUseReplacesWithBackup(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	first, second := t.TempDir(), t.TempDir()
	writePackage(t, first, "one", "1.0.0", "one\n")
	writePackage(t, second, "two", "1.0.0", "two\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(home, "one", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: second}); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(home, "two", "", "", false); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(homes["claude_code"], ".agent-environment-backup", "1", "CLAUDE.md")
	payload, err := os.ReadFile(backup) // #nosec G304 -- test home
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if !strings.Contains(string(payload), "## Context: one 1.0.0") {
		t.Fatalf("backup holds:\n%s", payload)
	}
}

// TestUsePartialRecordsNothing narrows the transaction gate: an unmanaged
// file in one home fails that entry with
// environment_surface_unmanaged_conflict, the other homes still switch, and
// the recorded current is unchanged with profile_use_partial.
func TestUsePartialRecordsNothing(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source, As: "acme"}); err != nil {
		t.Fatal(err)
	}
	// A first use with no marker anywhere succeeds and records current.
	if _, err := Use(home, "acme", "", "", false); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	writePackage(t, other, "other", "1.0.0", "other\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: other, As: "other"}); err != nil {
		t.Fatal(err)
	}
	// Simulate an unmanaged operator file: drop the codex marker but keep
	// a hand-written AGENTS.md behind.
	if err := os.Remove(filepath.Join(homes["codex_cli"], ".agent-environment.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(homes["codex_cli"], "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homes["codex_cli"], "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results, err := Use(home, "other", "", "", false)
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("partial use must fail as %s, got %v (%+v)", DiagUsePartial, err, results)
	}
	seenConflict := false
	oks := 0
	for _, result := range results {
		if result.OK {
			oks++
		}
		if result.Adapter == "codex_cli" && strings.Contains(result.Detail, DiagUnmanagedConflict) {
			seenConflict = true
		}
	}
	if !seenConflict || oks != len(results)-1 {
		t.Fatalf("results %+v", results)
	}
	if current, _ := Current(home); current != "acme" {
		t.Fatalf("partial use must not record current, got %q", current)
	}
}

// TestRemoveRefusesCurrent narrows the removal gate: a profile current in
// any scope fails with profile_in_use.
func TestRemoveRefusesCurrent(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	if err := Remove(home, "acme", false); err == nil || !strings.Contains(err.Error(), DiagInUse) {
		t.Fatalf("error %v carries no %s", err, DiagInUse)
	}
}

// TestRemovePurgeCleansHomes checks --purge removes the recorded surfaces,
// markers, and backups of the profile's homes.
func TestRemovePurgeCleansHomes(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	first, second := t.TempDir(), t.TempDir()
	writePackage(t, first, "one", "1.0.0", "one\n")
	writePackage(t, second, "two", "1.0.0", "two\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: second}); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(home, "two", "", "", false); err != nil {
		t.Fatal(err)
	}
	// "one" is current nowhere: purge removes its record. Its homes were
	// overwritten by "two", so purging "one" must not touch them.
	if err := Remove(home, "one", true); err != nil {
		t.Fatal(err)
	}
	if _, err := readSource(home, "one"); err == nil {
		t.Fatal("purged profile record remains")
	}
	marker, err := envmarker.Read(homes["claude_code"])
	if err != nil || marker == nil || marker.Profile.Name != "two" {
		t.Fatalf("purge of one must not touch two's marker: %+v %v", marker, err)
	}
}

// TestUpdatePathMovesLock checks a path profile re-resolves against its
// directory: edited modules move the lock.
func TestUpdatePathMovesLock(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	before, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	after, moved, err := Update(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if !moved || after.LockHash == before.LockHash {
		t.Fatalf("update must move the lock: %q -> %q", before.LockHash, after.LockHash)
	}
}

// TestInstallGitFromFileRemote checks the git pipeline hermetically: a
// local repository installed through a file:// URL resolves its tags,
// extracts from the object database, and locks the commit.
func TestInstallGitFromFileRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	repo := t.TempDir()
	writePackage(t, repo, "groot", "1.0.0", "git module\n")
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "one")
	run("tag", "v1.0.0")
	home := t.TempDir()
	info, _, _, err := Install(home, InstallOptions{Operand: "file://" + repo})
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "groot" || info.Source.Kind != KindGit {
		t.Fatalf("install %+v", info)
	}
	member, ok := info.Lock.RootMember()
	if !ok || member.Commit == "" {
		t.Fatalf("lock %+v", info.Lock)
	}
}

// TestInstallSurfacesSystemModuleWarning checks the always-warn class: a
// system module installs fine and is reported in warnings.
func TestInstallSurfacesSystemModuleWarning(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	manifest := `{"schema_version": 1, "name": "sys", "version": "1.0.0",` +
		`"context": {"modules": [{"path": "s.md", "class": "system"}]}}`
	if err := os.MkdirAll(filepath.Join(source, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "agent-context.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "s.md"), []byte("system\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, contextaudit.ClassSystemModulePresent) {
			found = true
		}
	}
	if !found {
		t.Fatalf("warnings %+v", info.Warnings)
	}
}

func gitRepo(t *testing.T, files map[string]string, tag string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	repo := t.TempDir()
	for path, content := range files {
		full := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "one")
	run("tag", tag)
	return repo
}

// TestInstallSurfacesUnresolvedMCPCommand checks the mcp_command_unresolved
// warning: a stdio server whose command is absent from PATH installs with a
// warning, never a failure.
func TestInstallSurfacesUnresolvedMCPCommand(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	mcp := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "definitely-absent-command-xyz", "args": []}}` + "\n",
	}, "v1.0.0")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",` +
			`"requires": {"mcp": {"tool": {"git": "file://` + mcp + `", "range": "*"}}}}` + "\n",
	}, "v1.0.0")
	info, _, _, err := Install(home, InstallOptions{Operand: "file://" + root})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, "mcp_command_unresolved") {
			found = true
		}
	}
	if !found {
		t.Fatalf("warnings %+v", info.Warnings)
	}
}

// TestUpdateDefaultIsBlocked narrows the migration gate: the builtin local
// profile never moves.
func TestUpdateDefaultIsBlocked(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	if _, _, err := Update(home, DefaultProfile); err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
		t.Fatalf("error %v carries no %s", err, DiagUpdateBlocked)
	}
}

// TestEnsureDefaultCreatesLocalProfile checks the §9.4 migration: the first
// use of the profile surface creates the builtin local default umbrella.
func TestEnsureDefaultCreatesLocalProfile(t *testing.T) {
	home := t.TempDir()
	if err := EnsureDefault(home); err != nil {
		t.Fatal(err)
	}
	source, err := readSource(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	if source.Kind != KindLocal {
		t.Fatalf("default kind %q", source.Kind)
	}
	lock, _, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Validate(); err != nil {
		t.Fatalf("default lock invalid: %v", err)
	}
}
