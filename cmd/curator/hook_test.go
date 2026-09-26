package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/hookapproval"
	"github.com/relux-works/curator/internal/shell"
)

// These tests drive the production run() entry point for the closed approval
// surface of Manager profile §8.3 (curator hook approve | approvals |
// revoke), the §8.6 posture rows of curator status and curator env status,
// and the approval-to-hook end-to-end trust path.

// hookProject stages a project with one env file and returns the manager
// home, the config path, and the env file path. The config file itself is
// only needed by commands that load it; the hook commands resolve the home
// from the path alone.
func hookProject(t *testing.T, content string) (home, configPath, envPath string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home = filepath.Join(base, "home")
	configPath = filepath.Join(home, "config.json")
	envPath = filepath.Join(base, "project", ".agents", "env.sh")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return home, configPath, envPath
}

func TestHookApproveRecordsOperatorApproval(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	code, stdout, stderr := capture(t, configPath, "hook", "approve", envPath)
	if code != exitOK {
		t.Fatalf("hook approve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "approved") || !strings.Contains(stdout, envPath) {
		t.Fatalf("hook approve stdout:\n%s", stdout)
	}
	record, found, err := hookapproval.Lookup(home, envPath)
	if err != nil || !found {
		t.Fatalf("Lookup = %v, %v", found, err)
	}
	payload, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if record.SHA256 != hookapproval.Digest(payload) {
		t.Fatalf("recorded digest = %s, want the digest of the current bytes", record.SHA256)
	}
	if record.ApprovedBy != hookapproval.ApprovedByOperator {
		t.Fatalf("approved_by = %q, want operator", record.ApprovedBy)
	}
	if record.ApprovedAt.IsZero() {
		t.Fatal("approved_at is missing")
	}
	if record.Path != envPath {
		t.Fatalf("recorded path = %q, want %q", record.Path, envPath)
	}
}

func TestHookApproveFailsDistinctlyOnAbsentAndUnreadable(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "bytes\n")
	missing := filepath.Join(home, "project", ".agents", "env.sh")
	code, _, stderr := capture(t, configPath, "hook", "approve", missing)
	if code != exitFail {
		t.Fatalf("hook approve of an absent file = %d, want %d", code, exitFail)
	}
	if !strings.Contains(stderr, "absent") || !strings.Contains(stderr, "nothing recorded") {
		t.Fatalf("absent stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "unreadable") {
		t.Fatalf("absent verdict leaks the unreadable wording:\n%s", stderr)
	}
	// A directory is deterministically unreadable-as-bytes on every
	// platform and runner, including a privileged one.
	unreadable := filepath.Dir(envPath)
	if code, _, stderr := capture(t, configPath, "hook", "approve", unreadable); code != exitFail {
		t.Fatalf("hook approve of a directory = %d, want %d", code, exitFail)
	} else {
		if !strings.Contains(stderr, "unreadable") || !strings.Contains(stderr, "nothing recorded") {
			t.Fatalf("unreadable stderr:\n%s", stderr)
		}
		if strings.Contains(stderr, "absent") {
			t.Fatalf("unreadable verdict leaks the absent wording:\n%s", stderr)
		}
	}
	// A permission-stripped file is the realistic unreadable shape; it
	// proves the same verdict wherever the runner honors permission bits.
	stripped := filepath.Join(home, "stripped.sh")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stripped, []byte("bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(stripped, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(stripped, 0o600) }()
	if _, err := os.ReadFile(stripped); err != nil {
		if code, _, stderr := capture(t, configPath, "hook", "approve", stripped); code != exitFail ||
			!strings.Contains(stderr, "unreadable") || !strings.Contains(stderr, "nothing recorded") {
			t.Fatalf("hook approve of a stripped file = %d\nstderr:\n%s", code, stderr)
		}
	}
	// Neither failure records anything.
	if _, err := os.Stat(hookapproval.ApprovalsPath(home)); !os.IsNotExist(err) {
		t.Fatalf("a failed approval wrote state: %v", err)
	}
}

func TestHookApproveReRecordsAfterChange(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	first, found, err := hookapproval.Lookup(home, envPath)
	if err != nil || !found {
		t.Fatalf("Lookup = %v, %v", found, err)
	}
	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve after change = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	second, found, err := hookapproval.Lookup(home, envPath)
	if err != nil || !found {
		t.Fatalf("Lookup = %v, %v", found, err)
	}
	if second.SHA256 == first.SHA256 {
		t.Fatal("re-approval kept the stale digest")
	}
	payload, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if second.SHA256 != hookapproval.Digest(payload) || second.ApprovedBy != hookapproval.ApprovedByOperator {
		t.Fatalf("re-recorded = %+v", second)
	}
	records, err := hookapproval.List(home)
	if err != nil || len(records) != 1 {
		t.Fatalf("List = %v, %v; re-approval must not duplicate the record", records, err)
	}
}

func TestHookApproveIsIdempotentWhenRecordMatches(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	before, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("second hook approve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	} else if !strings.Contains(stdout, "approved") {
		t.Fatalf("second approve stdout:\n%s", stdout)
	}
	after, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("idempotent approval rewrote state:\n%s\nwas:\n%s", after, before)
	}
}

func TestHookApproveResolvesSymlinkedOperand(t *testing.T) {
	t.Parallel()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	configPath := filepath.Join(home, "config.json")
	realDir := filepath.Join(base, "real", ".agents")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	realPath := filepath.Join(realDir, "env.sh")
	if err := os.WriteFile(realPath, []byte("export CURATOR_PROJECT_ENV=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(filepath.Join(base, "real"), alias); err != nil {
		t.Skipf("this host cannot create the symlinked project directory the case needs: %v", err)
	}
	aliasPath := filepath.Join(alias, ".agents", "env.sh")
	if code, _, stderr := capture(t, configPath, "hook", "approve", aliasPath); code != exitOK {
		t.Fatalf("hook approve via alias = %d\nstderr:\n%s", code, stderr)
	}
	record, found, err := hookapproval.Lookup(home, realPath)
	if err != nil || !found {
		t.Fatalf("Lookup(real) = %v, %v", found, err)
	}
	resolvedReal, err := filepath.EvalSymlinks(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if record.Path != resolvedReal {
		t.Fatalf("stored path = %q, want %q", record.Path, resolvedReal)
	}
	records, err := hookapproval.List(home)
	if err != nil || len(records) != 1 {
		t.Fatalf("List = %v, %v; two spellings share one record", records, err)
	}
}

func TestHookApprovalsListsReadOnly(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	other := filepath.Join(filepath.Dir(filepath.Dir(envPath)), "env.ps1")
	if err := os.WriteFile(other, []byte("$env:X = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "hook", "approve", other); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	payload, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := hookapproval.Upsert(home, hookapproval.Record{
		Path: envPath, SHA256: hookapproval.Digest(payload),
		ApprovedBy: hookapproval.ApprovedByManager, ApprovedAt: time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, configPath, "hook", "approvals")
	if code != exitOK {
		t.Fatalf("hook approvals = %d\nstderr:\n%s", code, stderr)
	}
	for _, want := range []string{envPath, other, hookapproval.ApprovedByManager, hookapproval.ApprovedByOperator, "2026-09-16T00:00:00Z"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("approvals stdout lacks %q:\n%s", want, stdout)
		}
	}
	if strings.Index(stdout, envPath) > strings.Index(stdout, other) {
		t.Fatalf("approvals order is not stable (sorted by path):\n%s", stdout)
	}
	after, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("approvals mutated state:\n%s\nwas:\n%s", after, before)
	}
}

func TestHookApprovalsToleratesMalformedRecord(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	payload, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	body := envPath + "\t" + hookapproval.Digest(payload) + "\tmanager\t2026-09-16T00:00:00Z\n" +
		"malformed-line\n"
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hookapproval.ApprovalsPath(home), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, configPath, "hook", "approvals")
	if code != exitOK {
		t.Fatalf("hook approvals over a malformed record = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, envPath) || strings.Contains(stdout, "malformed-line") {
		t.Fatalf("approvals stdout:\n%s", stdout)
	}
	if !strings.Contains(stderr, "line 2") || !strings.Contains(stderr, "malformed") {
		t.Fatalf("approvals stderr does not report the malformed line:\n%s", stderr)
	}
	after, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != body {
		t.Fatalf("approvals repaired state:\n%s\nwas:\n%s", after, body)
	}
}

func TestHookApprovalsOnAbsentStateIsEmpty(t *testing.T) {
	t.Parallel()
	configPath := filepath.Join(t.TempDir(), "config.json")
	code, stdout, stderr := capture(t, configPath, "hook", "approvals")
	if code != exitOK || stdout != "" {
		t.Fatalf("hook approvals on absent state = %d %q\nstderr:\n%s", code, stdout, stderr)
	}
}

func TestHookRevokeRemovesRecord(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := capture(t, configPath, "hook", "revoke", envPath)
	if code != exitOK {
		t.Fatalf("hook revoke = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "revoked") || !strings.Contains(stdout, envPath) {
		t.Fatalf("hook revoke stdout:\n%s", stdout)
	}
	if _, found, err := hookapproval.Lookup(home, envPath); err != nil || found {
		t.Fatalf("Lookup after revoke = %v, %v", found, err)
	}
}

func TestHookRevokeAbsentLeavesStateByteIdentical(t *testing.T) {
	t.Parallel()
	home, configPath, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	before, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(filepath.Dir(home), "other", ".agents", "env.sh")
	code, stdout, stderr := capture(t, configPath, "hook", "revoke", other)
	if code != exitOK {
		t.Fatalf("hook revoke of a missing record = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	if !strings.Contains(stdout, "nothing to revoke") {
		t.Fatalf("hook revoke stdout does not report the no-op:\n%s", stdout)
	}
	after, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("revoke of a missing record rewrote state:\n%s\nwas:\n%s", after, before)
	}
	// Revoking from an absent state is the same benign no-op and creates
	// no state file.
	emptyConfig := filepath.Join(t.TempDir(), "config.json")
	if code, stdout, _ := capture(t, emptyConfig, "hook", "revoke", other); code != exitOK || !strings.Contains(stdout, "nothing to revoke") {
		t.Fatalf("hook revoke on absent state = %d %q", code, stdout)
	}
	if _, err := os.Stat(hookapproval.ApprovalsPath(filepath.Dir(emptyConfig))); !os.IsNotExist(err) {
		t.Fatalf("revoke on absent state created state: %v", err)
	}
}

func TestHookUsageErrors(t *testing.T) {
	t.Parallel()
	configPath := filepath.Join(t.TempDir(), "config.json")
	for _, args := range [][]string{
		{"hook"},
		{"hook", "bless"},
		{"hook", "approve"},
		{"hook", "approve", "a", "b"},
		{"hook", "approvals", "extra"},
		{"hook", "revoke"},
		{"hook", "revoke", "a", "b"},
	} {
		if code, _, _ := capture(t, configPath, args...); code != exitUsage {
			t.Fatalf("%v = %d, want usage %d", args, code, exitUsage)
		}
	}
}

// TestHookDiagnosticsAgreeWithShell pins the closed §8.4 spelling across
// the three surfaces that carry it (approval library, emitted hooks,
// posture rows) and the warn-first default of §8.5.
func TestHookDiagnosticsAgreeWithShell(t *testing.T) {
	t.Parallel()
	if hookapproval.DiagnosticEnvUnapproved != shell.DiagnosticEnvUnapproved ||
		hookapproval.DiagnosticEnvChanged != shell.DiagnosticEnvChanged {
		t.Fatalf("hookapproval diagnostics %q/%q disagree with shell %q/%q",
			hookapproval.DiagnosticEnvUnapproved, hookapproval.DiagnosticEnvChanged,
			shell.DiagnosticEnvUnapproved, shell.DiagnosticEnvChanged)
	}
	if shell.DefaultTrustProfile != shell.TrustProfileAWarning {
		t.Fatalf("shipped trust profile = %q, want A-warning", shell.DefaultTrustProfile)
	}
}

// hookStatusProject bootstraps a config and registers one empty project.
// The project is a git checkout with the managed gitignore entries, so the
// status read-only plan runs instead of skipping.
func hookStatusProject(t *testing.T) (configPath, project string) {
	t.Helper()
	configPath = filepath.Join(t.TempDir(), "config.json")
	if code := runCode(t, configPath, []string{"bootstrap", "--non-interactive", "--skills-root", t.TempDir()}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	project = t.TempDir()
	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runCode(t, configPath, []string{"project", "add", "app", project, "--agents", "codex_cli"}); code != exitOK {
		t.Fatalf("project add = %d", code)
	}
	return configPath, project
}

func writeHookEnvFile(t *testing.T, project, content string) string {
	t.Helper()
	envPath := filepath.Join(project, ".agents", "env.sh")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	// Posture rows name the canonical identity, so the expectation must
	// resolve the temporary directory's symlinked prefix too.
	resolved, err := filepath.EvalSymlinks(envPath)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// TestStatusReportsShellHookTrustPosture proves the §8.6 rows of curator
// status: unapproved files warn without failing --check, approved files
// report approved_by, and changed files fail --check as non-current.
func TestStatusReportsShellHookTrustPosture(t *testing.T) {
	t.Parallel()
	configPath, project := hookStatusProject(t)
	envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")

	code, stdout, _ := capture(t, configPath, "status", "app")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "shell-hook-trust: "+envPath) ||
		!strings.Contains(stdout, hookapproval.DiagnosticEnvUnapproved) ||
		!strings.Contains(stdout, "curator hook approve "+envPath) {
		t.Fatalf("status lacks the unapproved row:\n%s", stdout)
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check over an unapproved file = %d, want %d (warning rows never fail the check)", code, exitOK)
	}

	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, stdout, _ := capture(t, configPath, "status", "app"); code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	} else if !strings.Contains(stdout, "shell-hook-trust: "+envPath+": approved (approved_by=operator)") {
		t.Fatalf("status lacks the approved row:\n%s", stdout)
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check over an approved file = %d, want %d", code, exitOK)
	}

	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, _ := capture(t, configPath, "status", "app"); code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	} else if !strings.Contains(stdout, hookapproval.DiagnosticEnvChanged) ||
		!strings.Contains(stdout, "approved_by=operator") ||
		!strings.Contains(stdout, "curator hook approve "+envPath) {
		t.Fatalf("status lacks the changed row:\n%s", stdout)
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check over a changed file = %d, want %d", code, exitFail)
	}
}

// TestStatusJSONCarriesShellHookTrust proves the machine-readable posture:
// the document always carries one entry per known file with path, state,
// and — for recorded files — the diagnostic and approved_by value. An
// all-approved posture is still posture: the key stays present.
func TestStatusJSONCarriesShellHookTrust(t *testing.T) {
	t.Parallel()
	configPath, project := hookStatusProject(t)
	envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")
	decode := func(t *testing.T, stdout string) map[string]any {
		t.Helper()
		var decoded map[string]any
		if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
			t.Fatalf("status --json is not JSON: %v\n%s", err, stdout)
		}
		return decoded
	}

	// Unapproved: the key appears with the warning row, which carries no
	// approved_by (nothing is recorded).
	_, stdout, _ := capture(t, configPath, "status", "app", "--json")
	decoded := decode(t, stdout)
	rows, ok := decoded["shell_hook_trust"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("shell_hook_trust = %v, want the one unapproved file", decoded["shell_hook_trust"])
	}
	row, _ := rows[0].(map[string]any)
	if row["path"] != envPath || row["state"] != hookapproval.PostureUnapproved ||
		row["diagnostic"] != hookapproval.DiagnosticEnvUnapproved {
		t.Fatalf("shell_hook_trust row = %v", row)
	}
	if _, present := row["approved_by"]; present {
		t.Fatalf("unapproved row carries approved_by: %v", row)
	}
	if _, present := row["file"]; present {
		t.Fatalf("readable unapproved row carries a file qualifier: %v", row)
	}

	// Changed: approve, then modify; the key carries the changed row with
	// the recorded approved_by.
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, stdout, _ = capture(t, configPath, "status", "app", "--json")
	decoded = decode(t, stdout)
	rows, ok = decoded["shell_hook_trust"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("shell_hook_trust = %v, want the one changed file", decoded["shell_hook_trust"])
	}
	row, _ = rows[0].(map[string]any)
	if row["path"] != envPath || row["state"] != hookapproval.PostureChanged ||
		row["diagnostic"] != hookapproval.DiagnosticEnvChanged || row["approved_by"] != hookapproval.ApprovedByOperator {
		t.Fatalf("shell_hook_trust row = %v", row)
	}
	if _, present := row["file"]; present {
		t.Fatalf("readable changed row carries a file qualifier: %v", row)
	}

	// All approved: re-approve; the key stays present with the approved
	// row, which carries approved_by and no diagnostic.
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, _ := capture(t, configPath, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d", code)
	}
	decoded = decode(t, stdout)
	rows, ok = decoded["shell_hook_trust"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("shell_hook_trust = %v, want the one approved file", decoded["shell_hook_trust"])
	}
	row, _ = rows[0].(map[string]any)
	if row["path"] != envPath || row["state"] != hookapproval.PostureApproved ||
		row["approved_by"] != hookapproval.ApprovedByOperator {
		t.Fatalf("shell_hook_trust row = %v", row)
	}
	if _, present := row["diagnostic"]; present {
		t.Fatalf("approved row carries a diagnostic: %v", row)
	}
	if _, present := row["file"]; present {
		t.Fatalf("readable approved row carries a file qualifier: %v", row)
	}
}

// TestStatusReportsRecordedPathsBeyondTheCurrentProject proves the known
// set is every recorded path, not just the current project's env files.
func TestStatusReportsRecordedPathsBeyondTheCurrentProject(t *testing.T) {
	t.Parallel()
	configPath, project := hookStatusProject(t)
	_, _, foreign := hookProject(t, "export FOREIGN=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", foreign); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, stdout, _ := capture(t, configPath, "status", "app"); code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	} else if !strings.Contains(stdout, "shell-hook-trust: "+foreign+": approved (approved_by=operator)") {
		t.Fatalf("status lacks the foreign recorded row:\n%s", stdout)
	}
	// Mixed posture: an unapproved project file joins the approved foreign
	// record, and the machine-readable key lists each known file.
	envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")
	_, stdout, _ := capture(t, configPath, "status", "app", "--json")
	var decoded struct {
		ShellHookTrust []struct {
			Path  string `json:"path"`
			State string `json:"state"`
		} `json:"shell_hook_trust"`
	}
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("status --json is not JSON: %v\n%s", err, stdout)
	}
	if len(decoded.ShellHookTrust) != 2 {
		t.Fatalf("shell_hook_trust = %+v, want the foreign and project files", decoded.ShellHookTrust)
	}
	byPath := map[string]string{}
	for _, row := range decoded.ShellHookTrust {
		byPath[row.Path] = row.State
	}
	if byPath[foreign] != hookapproval.PostureApproved || byPath[envPath] != hookapproval.PostureUnapproved {
		t.Fatalf("shell_hook_trust rows = %v", byPath)
	}
}

// TestEnvStatusReportsShellHookTrustPosture proves the §8.6 rows of
// curator env status over recorded paths: approved rows report
// approved_by, changed files fail --check, and the JSON matrix carries
// the posture. The launch-directory project half is proven at the
// StatusOf level, where the launch directory is injectable.
func TestEnvStatusReportsShellHookTrustPosture(t *testing.T) {
	// No t.Parallel: profileHome sets process environment, exactly like
	// the other env matrix tests.
	// The §12 provider rows join the matrix: run and session are
	// always reported, and a missing row is non-current — so plant
	// stub providers, which warn outside the trust roots under
	// revision A but stay current, keeping the --check exit code
	// evidence of the trust posture alone.
	bin := t.TempDir()
	for _, name := range []string{"curator-run", "curator-session"} {
		full := filepath.Join(bin, name)
		if runtime.GOOS == "windows" {
			full += ".exe"
		}
		if err := os.WriteFile(full, []byte(""), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	for _, profile := range []string{"acme", "default"} {
		for _, env := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
			if code, _, stderr := runProfile(t, source, "env", "resolve", env, "--profile", profile, "--repair"); code != exitOK {
				t.Fatalf("repair %s %s stderr:\n%s", profile, env, stderr)
			}
		}
	}
	_, _, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")

	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check without approvals = %d\n%s", code, stdout)
	} else if strings.Contains(stdout, "shell-hook-trust:") {
		t.Fatalf("env status without approvals reports trust rows:\n%s", stdout)
	}

	// Approvals resolve the home from the config path, so the stub source
	// shares its home with the status matrix below.
	approveSource := stubConfigSource{path: source.path, cfg: source.cfg}
	if code, _, stderr := runProfile(t, approveSource, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status"); code != exitOK {
		t.Fatalf("env status = %d\n%s", code, stdout)
	} else if !strings.Contains(stdout, "shell-hook-trust: "+envPath+": approved (approved_by=operator)") {
		t.Fatalf("env status lacks the approved row:\n%s", stdout)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check over an approved file = %d, want %d\n%s", code, exitOK, stdout)
	}

	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status"); code != exitOK {
		t.Fatalf("env status = %d\n%s", code, stdout)
	} else if !strings.Contains(stdout, hookapproval.DiagnosticEnvChanged) {
		t.Fatalf("env status lacks the changed row:\n%s", stdout)
	}
	if code, _, _ := runProfile(t, source, "env", "status", "--check"); code != exitFail {
		t.Fatalf("env status --check over a changed file = %d, want %d", code, exitFail)
	}
	if code, jsonOut, _ := runProfile(t, source, "env", "status", "--json"); code != exitOK {
		t.Fatalf("env status --json = %d", code)
	} else {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
			t.Fatalf("env status --json is not JSON: %v", err)
		}
		rows, ok := decoded["shell_hook_trust"].([]any)
		if !ok || len(rows) != 1 {
			t.Fatalf("shell_hook_trust = %v, want the one changed file", decoded["shell_hook_trust"])
		}
		row, _ := rows[0].(map[string]any)
		if row["path"] != envPath || row["state"] != hookapproval.PostureChanged ||
			row["diagnostic"] != hookapproval.DiagnosticEnvChanged || row["approved_by"] != hookapproval.ApprovedByOperator {
			t.Fatalf("shell_hook_trust row = %v", row)
		}
	}
}

// e2eBash resolves one POSIX interpreter for the approval end-to-end
// test, including Git Bash on Windows. The skip names the absent
// interpreter (host-capability), never the platform alone.
func e2eBash(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"bash", "sh"} {
		if executable, err := exec.LookPath(name); err == nil {
			return executable
		}
	}
	if runtime.GOOS == "windows" {
		for _, candidate := range []string{
			`C:\Program Files\Git\bin\bash.exe`,
			`C:\Program Files\Git\usr\bin\bash.exe`,
			`C:\Program Files\Git\bin\sh.exe`,
			`C:\Program Files\Git\usr\bin\sh.exe`,
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
		t.Skip("bash is unavailable (Git Bash absent on this runner)")
	}
	t.Skip("bash is unavailable (no POSIX shell on this runner)")
	return ""
}

// TestHookApproveEndToEndWithGeneratedHook drives the real CLI approval
// commands and the real generated hook: approving silences activation,
// changing bytes warns, re-approving silences again, and revoking warns
// again — all under the shipped A-warning profile, which keeps sourcing.
func TestHookApproveEndToEndWithGeneratedHook(t *testing.T) {
	t.Parallel()
	bash := e2eBash(t)
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	configPath := filepath.Join(home, "config.json")
	project := filepath.Join(base, "project")
	envPath := filepath.Join(project, ".agents", "env.sh")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte("export CURATOR_E2E_MARKER=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	hook, err := shell.HookWithProfile("bash", false, shell.DefaultTrustProfile)
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(base, "hook")
	if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
		t.Fatal(err)
	}
	activate := func(t *testing.T) (stdout, stderr string) {
		t.Helper()
		script := "cd \"$PROJ\"\n. \"$HOOK\"\nprintf 'sourced=%s\\n' \"${CURATOR_E2E_MARKER:-no}\"\n"
		baseName := strings.ToLower(filepath.Base(bash))
		baseName = strings.TrimSuffix(baseName, ".exe")
		args := []string{"--noprofile", "--norc", "-c", script}
		if baseName == "sh" || baseName == "dash" {
			args = []string{"-c", script}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, bash, args...)
		command.Env = append(os.Environ(),
			"PROJ="+project, "HOOK="+hookPath,
			"CURATOR_CONFIG="+configPath)
		var out, errOut strings.Builder
		command.Stdout = &out
		command.Stderr = &errOut
		if err := command.Run(); err != nil {
			t.Fatalf("activation: %v\nstdout:\n%s\nstderr:\n%s", err, out.String(), errOut.String())
		}
		return out.String(), strings.ReplaceAll(errOut.String(), "\r\n", "\n")
	}
	check := func(t *testing.T, wantSourced, wantDiagnostic string, wantWarnings int) {
		t.Helper()
		stdout, stderr := activate(t)
		if !strings.Contains(stdout, "sourced="+wantSourced) {
			t.Fatalf("stdout lacks sourced=%s:\nstdout:\n%s\nstderr:\n%s", wantSourced, stdout, stderr)
		}
		if got := strings.Count(stderr, "curator hook approve "); got != wantWarnings {
			t.Fatalf("warning count = %d, want %d:\n%s", got, wantWarnings, stderr)
		}
		if wantDiagnostic == "" {
			if strings.Contains(stderr, "shell_hook_env_") {
				t.Fatalf("trusted activation warned:\n%s", stderr)
			}
			return
		}
		if !strings.Contains(stderr, wantDiagnostic) {
			t.Fatalf("stderr lacks diagnostic %q:\n%s", wantDiagnostic, stderr)
		}
		if runtime.GOOS != "windows" && !strings.Contains(stderr, "curator hook approve "+envPath) {
			t.Fatalf("warning does not name the approval command:\n%s", stderr)
		}
	}

	check(t, "1", hookapproval.DiagnosticEnvUnapproved, 1)
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	check(t, "1", "", 0)
	if code, stdout, _ := capture(t, configPath, "hook", "approvals"); code != exitOK || !strings.Contains(stdout, envPath) {
		t.Fatalf("hook approvals = %d:\n%s", code, stdout)
	}
	if err := os.WriteFile(envPath, []byte("export CURATOR_E2E_MARKER=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	check(t, "2", hookapproval.DiagnosticEnvChanged, 1)
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve after change = %d\nstderr:\n%s", code, stderr)
	}
	check(t, "2", "", 0)
	if code, _, stderr := capture(t, configPath, "hook", "revoke", envPath); code != exitOK {
		t.Fatalf("hook revoke = %d\nstderr:\n%s", code, stderr)
	}
	check(t, "2", hookapproval.DiagnosticEnvUnapproved, 1)
	if code, stdout, _ := capture(t, configPath, "hook", "approvals"); code != exitOK || stdout != "" {
		t.Fatalf("hook approvals after revoke = %d %q", code, stdout)
	}
}
