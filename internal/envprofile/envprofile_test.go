package envprofile

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
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
	pinHomes(t)
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
	pinHomes(t)
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
	pinHomes(t)
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

// TestInstallGitResolvesNetworkIdentity checks the git pipeline hermetically:
// a local repository served under a fake network identity (git insteadOf)
// resolves its tags, extracts from the object database, and locks the
// commit under the canonical source identity. A file:// operand is rejected
// (see TestFileOperandIsRefused) and must never reach this path.
func TestInstallGitResolvesNetworkIdentity(t *testing.T) {
	pinHomes(t)
	repo := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "groot", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"context/a.md": "git module\n",
	}, "v1.0.0")
	ids := newGitIdentities(t)
	operand := ids.serve(repo, "https://example.com/groot")
	home := t.TempDir()
	info, _, _, err := Install(home, InstallOptions{Operand: operand})
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
	if member.Source != "example.com/groot" {
		t.Fatalf("lock source %q, want the canonical identity", member.Source)
	}
	if err := info.Lock.Validate(); err != nil {
		t.Fatalf("lock invalid: %v", err)
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
	ids := newGitIdentities(t)
	mcpRepo := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "definitely-absent-command-xyz", "args": []}}` + "\n",
	}, "v1.0.0")
	mcpOperand := ids.serve(mcpRepo, "https://example.com/mcp-tool")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",` +
			`"requires": {"mcp": {"tool": {"git": "` + mcpOperand + `", "range": "*"}}}}` + "\n",
	}, "v1.0.0")
	info, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/withmcp")})
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

// TestScopedUseAndClear checks narrowed switches: a scoped use records a
// scoped current without moving the machine current, and --clear drops the
// record and re-materializes the scope from the machine default.
func TestScopedUseAndClear(t *testing.T) {
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
	if _, err := Use(home, "two", "codex_cli", "", false); err != nil {
		t.Fatal(err)
	}
	if current, _ := Current(home); current != "one" {
		t.Fatalf("scoped use moved the machine current to %q", current)
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		t.Fatal(err)
	}
	if scoped["env:codex_cli"] != "two" {
		t.Fatalf("scoped %+v", scoped)
	}
	marker, err := envmarker.Read(homes["codex_cli"])
	if err != nil || marker == nil || marker.Profile.Name != "two" {
		t.Fatalf("codex marker %+v %v", marker, err)
	}
	if _, err := Use(home, "", "codex_cli", "", true); err != nil {
		t.Fatal(err)
	}
	scoped, err = ScopedCurrents(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(scoped) != 0 {
		t.Fatalf("clear left scoped %+v", scoped)
	}
	marker, err = envmarker.Read(homes["codex_cli"])
	if err != nil || marker == nil || marker.Profile.Name != "one" {
		t.Fatalf("codex marker after clear %+v %v", marker, err)
	}
}

// TestListFailsOnCorruptRecord narrows the listing gate: a present but
// unreadable install record fails the listing rather than silently dropping
// the profile. Reserved state without a record (scoped/) is skipped.
func TestListFailsOnCorruptRecord(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(home, "acme", "codex_cli", "", false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ProfileDir(home, "acme"), "source.json"), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := List(home); err == nil {
		t.Fatal("corrupt record must fail the listing")
	}
}

// directorySecret is a secret-shaped module payload assembled at runtime so
// no secret-shaped literal appears in the source; the detector still sees
// the joined bytes.
func directorySecret() string { return "key " + "AK" + "IA1234567890ABCDEF" + " here\n" }

// TestInstallDirectoryRejectsSecretUnderSubdirectory narrows the audit-scope
// gate (review F1): a --directory install scans the package root below the
// directory, not the snapshot root. Production entry point: Install with
// Directory set. A mutant that restores the snapshot-root audit path must
// fail this test while TestInstallRejectsSecretMember still passes.
func TestInstallDirectoryRejectsSecretUnderSubdirectory(t *testing.T) {
	repo := gitRepo(t, map[string]string{
		"sub/agent-context.json": `{"schema_version": 1, "name": "acme", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"sub/context/a.md": directorySecret(),
	}, "v1.0.0")
	ids := newGitIdentities(t)
	home := t.TempDir()
	if _, _, _, err := Install(home, InstallOptions{Operand: ids.serve(repo, "https://example.com/acme"), Directory: "sub"}); err == nil {
		t.Fatal("directory-addressed secret member must fail installation")
	} else if !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("error %v carries no %s", err, DiagSourceInvalid)
	}
	if _, err := readSource(home, "acme"); err == nil {
		t.Fatal("failed install must leave no profile record")
	}
}

// TestInstallTransitiveDirectoryRejectsSecret narrows the same gate for the
// transitive shape (review F1): a requires.contexts[].directory member with
// a secret fails installation. Production entry point: Install of a clean
// root whose dependency is directory-addressed.
func TestInstallTransitiveDirectoryRejectsSecret(t *testing.T) {
	dep := gitRepo(t, map[string]string{
		"sub/agent-context.json": `{"schema_version": 1, "name": "dep", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"sub/context/a.md": directorySecret(),
	}, "v1.0.0")
	ids := newGitIdentities(t)
	depOperand := ids.serve(dep, "https://example.com/dep")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "clean", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"contexts": {"dep": {"git": "` + depOperand + `", "range": "*", "directory": "sub"}}}}` + "\n",
		"context/a.md": "clean\n",
	}, "v1.0.0")
	home := t.TempDir()
	if _, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/clean")}); err == nil {
		t.Fatal("transitive directory-addressed secret member must fail installation")
	} else if !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("error %v carries no %s", err, DiagSourceInvalid)
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

// TestDefaultProfileMaterializesOnFreshHome drives the four reviewer-named
// production paths on an isolated home (review F2): profile list,
// profile use default, profile sync, and profile use --clear --env <id>.
// All four must work on a fresh manager home: the default lock pins a real
// store entry, and the contextless root materializes markers alone.
func TestDefaultProfileMaterializesOnFreshHome(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	profiles, err := List(home)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(profiles) != 1 || profiles[0].Name != DefaultProfile || profiles[0].Source.Kind != KindLocal {
		t.Fatalf("profiles %+v", profiles)
	}
	if results, err := Use(home, DefaultProfile, "", "", false); err != nil {
		t.Fatalf("use default: %v (%+v)", err, results)
	}
	if results, err := Sync(home); err != nil {
		t.Fatalf("sync: %v (%+v)", err, results)
	}
	if results, err := Use(home, "", "claude_code", "", true); err != nil {
		t.Fatalf("use --clear --env claude_code: %v (%+v)", err, results)
	}
	for id, dir := range homes {
		marker, err := envmarker.Read(dir)
		if err != nil {
			t.Fatal(err)
		}
		if marker == nil || marker.Profile.Name != DefaultProfile || marker.Mode != envmarker.ModeLinked {
			t.Fatalf("%s marker %+v", id, marker)
		}
	}
	// The contextless default root declares no root-context surface: no
	// adapter home gains a root-context file.
	if _, err := os.Stat(filepath.Join(homes["claude_code"], "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("default must write no root-context file, stat err %v", err)
	}
	// Migration is idempotent: the lock stands where the first call left it.
	before, beforeHash, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefault(home); err != nil {
		t.Fatal(err)
	}
	after, afterHash, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	if beforeHash != afterHash || len(before.Members) != len(after.Members) {
		t.Fatalf("second migration moved the lock: %s -> %s", beforeHash, afterHash)
	}
}

// TestEnsureDefaultMigratesGlobalSkills checks the §9.4 migration carries
// the machine's global git-pinned skills into the default lock as skill
// lock members with stored, audited pins. Production entry point:
// EnsureDefault over a home whose global scope declares one tag-pinned
// skill served from a local git remote.
func TestEnsureDefaultMigratesGlobalSkills(t *testing.T) {
	skill := gitRepo(t, map[string]string{"README.md": "skill\n"}, "v1.0.0")
	ids := newGitIdentities(t)
	skillOperand := ids.serve(skill, "https://example.com/skills/hello")
	home := t.TempDir()
	pinHomes(t)
	if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillfile := `{"schema_version": 1, "skills": [{"name": "sk", "git": "` + skillOperand + `", "tag": "v1.0.0"}]}` + "\n"
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.json"), []byte(skillfile), 0o644); err != nil {
		t.Fatal(err)
	}
	// Explicit empty policy: this test covers the migration mechanics, not
	// the machine-policy loader (see the F10 loader tests for that).
	if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Validate(); err != nil {
		t.Fatalf("migrated lock invalid: %v", err)
	}
	if len(lock.Members) != 2 {
		t.Fatalf("migrated lock %+v", lock.Members)
	}
	var member *contextlock.Member
	for index := range lock.Members {
		if lock.Members[index].Kind == contextlock.KindSkill {
			member = &lock.Members[index]
		}
	}
	if member == nil || member.Name != "sk" || member.Commit == "" || member.Source != "example.com/skills/hello" {
		t.Fatalf("migrated skill member %+v", lock.Members)
	}
	if !contextstore.Exists(home, contextlock.KindSkill, "sk", member.Commit) {
		t.Fatal("migrated skill pins a store entry that does not exist")
	}
	root, ok := lock.RootMember()
	if !ok || !contextstore.Exists(home, root.Kind, root.Name, root.StateHash) {
		t.Fatalf("default root pins a store entry that does not exist: %+v", lock.Members)
	}
}

// TestMutationLockContention narrows the serialization gate (review F3): a
// profile mutation that cannot acquire the manager-home mutation lock fails
// with environment_lock_unavailable instead of racing the holder.
// Production entry point: Install.
func TestMutationLockContention(t *testing.T) {
	home := t.TempDir()
	old := lockTimeout
	lockTimeout = 50 * time.Millisecond
	defer func() { lockTimeout = old }()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err == nil {
		t.Fatal("a mutation under a held lock must not proceed")
	} else if !strings.Contains(err.Error(), DiagLockUnavailable) {
		t.Fatalf("error %v carries no %s", err, DiagLockUnavailable)
	}
}

// TestIncompleteJournalRecoversOnEntry checks the recovery gate (review
// F3): a prepared-but-uncommitted record journal left by a crashed
// operation completes when the next operation begins. Production entry
// point: List, which recovers journals on lock entry. A mutant that drops
// the recovery must fail this test.
func TestIncompleteJournalRecoversOnEntry(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	if err := os.MkdirAll(ProfilesDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	staged, err := os.CreateTemp("", "curator-profile-recover-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(staged.Name()) }()
	if _, err := staged.WriteString(DefaultProfile + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := staged.Close(); err != nil {
		t.Fatal(err)
	}
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := transaction.New(home)
	if err != nil {
		_ = held.Close()
		t.Fatal(err)
	}
	preimage, err := transaction.DigestTarget(transaction.KindBytes, CurrentFile(home))
	if err != nil {
		_ = held.Close()
		t.Fatal(err)
	}
	plan := transaction.Plan{
		TransactionID:   newTransactionID(),
		ProjectIdentity: "profiles",
		Targets: []transaction.Target{{
			Class: "profile", Identifier: CurrentFile(home), Kind: transaction.KindBytes,
			LivePath: CurrentFile(home), StagedSource: staged.Name(), PreimageDigest: preimage,
		}},
	}
	if _, err := engine.Prepare(held, plan); err != nil {
		_ = held.Close()
		t.Fatal(err)
	}
	// Crash before commit: the journal stands, the pointer is unwritten.
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := List(home); err != nil {
		t.Fatalf("list after a crashed journal: %v", err)
	}
	if current, err := Current(home); err != nil || current != DefaultProfile {
		t.Fatalf("recovered current %q, err %v", current, err)
	}
}

// TestOperationsLeaveNoJournalResidue checks the journal gate (review F3):
// successful profile operations commit and clean their journals, and the
// published records carry the exact bytes the operation resolved.
// Production entry points: Install, Use.
func TestOperationsLeaveNoJournalResidue(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Use(home, "acme", "", "", false); err != nil {
		t.Fatal(err)
	}
	journals := filepath.Join(home, "state", "transactions", "v1")
	if entries, err := os.ReadDir(journals); err == nil && len(entries) != 0 {
		t.Fatalf("journal residue %+v", entries)
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	_, hash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if hash != info.LockHash {
		t.Fatalf("lock hash %s, want %s", hash, info.LockHash)
	}
	if current, err := Current(home); err != nil || current != "acme" {
		t.Fatalf("current %q, err %v", current, err)
	}
}
