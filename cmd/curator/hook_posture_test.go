package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hookapproval"
)

// These tests drive the production run() entry point for the §8.6 posture
// corrections: a recorded-but-missing file keeps its row (R1), an
// unreadable candidate keeps its record and reports the failed read (R2),
// and an unreadable approval state surfaces in JSON and fails --check
// (R3). Absence and read failure are different facts on every surface.

// decodeTrustDoc decodes a status --json / env status --json document and
// returns its shell_hook_trust rows plus the raw document for warnings-key
// assertions.
func decodeTrustDoc(t *testing.T, stdout string) ([]map[string]any, map[string]any) {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("trust JSON is not an object: %v\n%s", err, stdout)
	}
	raw, present := decoded["shell_hook_trust"]
	if !present {
		t.Fatalf("trust JSON lacks shell_hook_trust:\n%s", stdout)
	}
	var rows []map[string]any
	if raw != nil {
		list, ok := raw.([]any)
		if !ok {
			t.Fatalf("shell_hook_trust = %T, want an array or null:\n%s", raw, stdout)
		}
		for _, entry := range list {
			row, ok := entry.(map[string]any)
			if !ok {
				t.Fatalf("shell_hook_trust row is not an object: %v", entry)
			}
			rows = append(rows, row)
		}
	}
	return rows, decoded
}

func trustWarningsOf(t *testing.T, decoded map[string]any) []any {
	t.Helper()
	raw, present := decoded["shell_hook_trust_warnings"]
	if !present {
		t.Fatalf("trust JSON lacks shell_hook_trust_warnings: %v", decoded)
	}
	warnings, ok := raw.([]any)
	if !ok {
		t.Fatalf("shell_hook_trust_warnings = %T, want an array", raw)
	}
	return warnings
}

// assertCheckTrustRow decodes a --check --json document and asserts its
// single posture row: an approved operator record qualified by file, with
// no diagnostic (no digest comparison ran). The document is decoded
// rather than substring-matched: JSON escapes path separators, so a raw
// Windows path never occurs verbatim in --json output.
func assertCheckTrustRow(t *testing.T, what, stdout, path, file string) {
	t.Helper()
	rows, _ := decodeTrustDoc(t, stdout)
	if len(rows) != 1 {
		t.Fatalf("%s shell_hook_trust = %v, want the one %s row", what, rows, file)
	}
	row := rows[0]
	if row["path"] != path || row["state"] != hookapproval.PostureApproved ||
		row["approved_by"] != hookapproval.ApprovedByOperator || row["file"] != file {
		t.Fatalf("%s shell_hook_trust row = %v", what, row)
	}
	if _, present := row["diagnostic"]; present {
		t.Fatalf("%s row claims a diagnostic without a digest comparison: %v", what, row)
	}
}

// TestStatusRecordedButMissingStaysInInventory proves R1 on curator
// status: deleting an approved env file keeps its posture row (path,
// approved_by, explicit missing qualifier — never a claimed digest
// comparison) in text and JSON, and --check treats it as non-current.
func TestStatusRecordedButMissingStaysInInventory(t *testing.T) {
	t.Parallel()
	configPath, project := hookStatusProject(t)
	envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check over an approved file = %d, want %d", code, exitOK)
	}
	if err := os.Remove(envPath); err != nil {
		t.Fatal(err)
	}

	code, stdout, _ := capture(t, configPath, "status", "app")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	for _, want := range []string{"shell-hook-trust: " + envPath, "approved_by=operator", "file is missing"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("status lacks %q:\n%s", want, stdout)
		}
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check over a recorded-but-missing file = %d, want %d", code, exitFail)
	}

	code, stdout, _ = capture(t, configPath, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\n%s", code, stdout)
	}
	rows, decoded := decodeTrustDoc(t, stdout)
	if len(rows) != 1 {
		t.Fatalf("shell_hook_trust = %v, want the one missing row", rows)
	}
	row := rows[0]
	if row["path"] != envPath || row["state"] != hookapproval.PostureApproved ||
		row["approved_by"] != hookapproval.ApprovedByOperator || row["file"] != hookapproval.PostureFileMissing {
		t.Fatalf("shell_hook_trust row = %v", row)
	}
	if _, present := row["diagnostic"]; present {
		t.Fatalf("missing row claims a diagnostic without a digest comparison: %v", row)
	}
	if _, present := decoded["shell_hook_trust_warnings"]; present {
		t.Fatalf("no warnings exist, yet the document carries shell_hook_trust_warnings: %v", decoded)
	}
	if code, stdout, _ := capture(t, configPath, "status", "app", "--check", "--json"); code != exitFail {
		t.Fatalf("status --check --json over a recorded-but-missing file = %d, want %d\n%s", code, exitFail, stdout)
	} else {
		assertCheckTrustRow(t, "status --check --json", stdout, envPath, hookapproval.PostureFileMissing)
	}
}

// TestStatusUnreadableCandidatesKeepRecord proves R2 on curator status: a
// recorded file that becomes unreadable keeps its approved_by and reports
// the failed read explicitly (never via the ordinary no-record warning
// path), and every unreadable file is non-current under --check.
func TestStatusUnreadableCandidatesKeepRecord(t *testing.T) {
	t.Parallel()
	t.Run("recorded", func(t *testing.T) {
		t.Parallel()
		configPath, project := hookStatusProject(t)
		envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")
		if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK {
			t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
		}
		// A directory is deterministically unreadable-as-bytes on every
		// platform and runner, including a privileged one.
		if err := os.Remove(envPath); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(envPath, 0o700); err != nil {
			t.Fatal(err)
		}

		code, stdout, _ := capture(t, configPath, "status", "app")
		if code != exitOK {
			t.Fatalf("status = %d\n%s", code, stdout)
		}
		for _, want := range []string{"shell-hook-trust: " + envPath, "approved_by=operator", "file is unreadable"} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("status lacks %q:\n%s", want, stdout)
			}
		}
		if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitFail {
			t.Fatalf("status --check over an unreadable recorded file = %d, want %d", code, exitFail)
		}

		_, stdout, _ = capture(t, configPath, "status", "app", "--json")
		rows, _ := decodeTrustDoc(t, stdout)
		if len(rows) != 1 {
			t.Fatalf("shell_hook_trust = %v, want the one unreadable row", rows)
		}
		row := rows[0]
		if row["path"] != envPath || row["state"] != hookapproval.PostureApproved ||
			row["approved_by"] != hookapproval.ApprovedByOperator || row["file"] != hookapproval.PostureFileUnreadable {
			t.Fatalf("shell_hook_trust row = %v", row)
		}
		if _, present := row["diagnostic"]; present {
			t.Fatalf("unreadable recorded row claims a diagnostic without a digest comparison: %v", row)
		}
		if code, stdout, _ := capture(t, configPath, "status", "app", "--check", "--json"); code != exitFail {
			t.Fatalf("status --check --json over an unreadable recorded file = %d, want %d\n%s", code, exitFail, stdout)
		} else {
			assertCheckTrustRow(t, "status --check --json", stdout, envPath, hookapproval.PostureFileUnreadable)
		}
	})
	t.Run("unrecorded", func(t *testing.T) {
		t.Parallel()
		configPath, project := hookStatusProject(t)
		// A never-approved env path that reads as a directory: unreadable
		// without a record is still an explicit failed read, not an
		// ordinary warning row.
		envPath := filepath.Join(project, ".agents", "env.sh")
		if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(envPath, 0o700); err != nil {
			t.Fatal(err)
		}
		resolved, err := filepath.EvalSymlinks(envPath)
		if err != nil {
			t.Fatal(err)
		}

		code, stdout, _ := capture(t, configPath, "status", "app")
		if code != exitOK {
			t.Fatalf("status = %d\n%s", code, stdout)
		}
		for _, want := range []string{"shell-hook-trust: " + resolved, hookapproval.DiagnosticEnvUnapproved, "file is unreadable"} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("status lacks %q:\n%s", want, stdout)
			}
		}
		if strings.Contains(stdout, "approved_by") {
			t.Fatalf("unrecorded row claims approved_by:\n%s", stdout)
		}
		if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitFail {
			t.Fatalf("status --check over an unreadable file = %d, want %d", code, exitFail)
		}

		_, stdout, _ = capture(t, configPath, "status", "app", "--json")
		rows, _ := decodeTrustDoc(t, stdout)
		if len(rows) != 1 {
			t.Fatalf("shell_hook_trust = %v, want the one unreadable row", rows)
		}
		row := rows[0]
		if row["path"] != resolved || row["state"] != hookapproval.PostureUnapproved ||
			row["diagnostic"] != hookapproval.DiagnosticEnvUnapproved || row["file"] != hookapproval.PostureFileUnreadable {
			t.Fatalf("shell_hook_trust row = %v", row)
		}
		if _, present := row["approved_by"]; present {
			t.Fatalf("unrecorded row carries approved_by: %v", row)
		}
	})
}

// TestStatusUnreadableApprovalStateSurfaced proves R3 on curator status:
// an unreadable approval state surfaces in text and JSON (never dropped)
// and fails --check, while a truly absent state stays the normal
// unapproved-warning case.
func TestStatusUnreadableApprovalStateSurfaced(t *testing.T) {
	t.Parallel()
	configPath, project := hookStatusProject(t)
	envPath := writeHookEnvFile(t, project, "export CURATOR_PROJECT_ENV=1\n")
	home := filepath.Dir(configPath)
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check over an absent state = %d, want %d (ordinary warning case)", code, exitOK)
	}
	// A directory at the state file's own path is unreadable-as-bytes on
	// every platform (see TestScanRejectsUnreadableState).
	if err := os.MkdirAll(hookapproval.ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}

	code, stdout, _ := capture(t, configPath, "status", "app")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	for _, want := range []string{"cannot read", hookapproval.Filename, envPath} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("status hides the state read failure %q:\n%s", want, stdout)
		}
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check over an unreadable state = %d, want %d", code, exitFail)
	}

	code, stdout, _ = capture(t, configPath, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\n%s", code, stdout)
	}
	rows, decoded := decodeTrustDoc(t, stdout)
	if len(rows) != 1 || rows[0]["path"] != envPath {
		t.Fatalf("shell_hook_trust = %v, want the candidate reported unapproved rather than trusted", rows)
	}
	joined := strings.Join(func() []string {
		var out []string
		for _, warning := range trustWarningsOf(t, decoded) {
			text, _ := warning.(string)
			out = append(out, text)
		}
		return out
	}(), "\n")
	for _, want := range []string{"cannot read", hookapproval.Filename} {
		if !strings.Contains(joined, want) {
			t.Fatalf("shell_hook_trust_warnings hides %q: %v", want, joined)
		}
	}
	if code, stdout, _ := capture(t, configPath, "status", "app", "--check", "--json"); code != exitFail {
		t.Fatalf("status --check --json over an unreadable state = %d, want %d\n%s", code, exitFail, stdout)
	} else if !strings.Contains(stdout, "cannot read") {
		t.Fatalf("status --check --json drops the state read failure:\n%s", stdout)
	}

	// Paired control: removing the unreadable state restores the absent
	// case — an ordinary unapproved warning row, --check green, no
	// warnings key.
	if err := os.Remove(hookapproval.ApprovalsPath(home)); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check over an absent state = %d, want %d", code, exitOK)
	}
	code, stdout, _ = capture(t, configPath, "status", "app", "--check", "--json")
	if code != exitOK {
		t.Fatalf("status --check --json over an absent state = %d, want %d\n%s", code, exitOK, stdout)
	}
	rows, decoded = decodeTrustDoc(t, stdout)
	if len(rows) != 1 || rows[0]["state"] != hookapproval.PostureUnapproved || rows[0]["file"] != nil {
		t.Fatalf("shell_hook_trust = %v, want the ordinary unapproved row", rows)
	}
	if _, present := decoded["shell_hook_trust_warnings"]; present {
		t.Fatalf("absent state carries shell_hook_trust_warnings: %v", decoded)
	}
}

// provisionedEnvMatrix provisions every profile × environment home so the
// matrix is current and an `env status --check` exit code is evidence of
// the trust posture alone.
func provisionedEnvMatrix(t *testing.T) (stubConfigSource, string) {
	t.Helper()
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
	source, home := profileHome(t)
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
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check without approvals = %d\n%s", code, stdout)
	}
	return source, home
}

// TestEnvStatusMissingAndUnreadableKeepRecord proves R1+R2 on curator env
// status with an otherwise-current matrix: a recorded file that is
// deleted, then restored, then made unreadable keeps its row and
// approved_by throughout, reports the failed read explicitly in text and
// JSON, and fails --check exactly while the bytes are unavailable.
func TestEnvStatusMissingAndUnreadableKeepRecord(t *testing.T) {
	// No t.Parallel: profileHome sets process environment, exactly like
	// the other env matrix tests.
	source, _ := provisionedEnvMatrix(t)
	approveSource := stubConfigSource{path: source.path, cfg: source.cfg}
	_, _, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := runProfile(t, approveSource, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check over an approved file = %d, want %d\n%s", code, exitOK, stdout)
	}

	assertRow := func(t *testing.T, file, human string) {
		t.Helper()
		code, stdout, _ := runProfile(t, source, "env", "status")
		if code != exitOK {
			t.Fatalf("env status = %d\n%s", code, stdout)
		}
		for _, want := range []string{"shell-hook-trust: " + envPath, "approved_by=operator", human} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("env status lacks %q:\n%s", want, stdout)
			}
		}
		if code, _, _ := runProfile(t, source, "env", "status", "--check"); code != exitFail {
			t.Fatalf("env status --check with file %s = %d, want %d", file, code, exitFail)
		}
		_, stdout, _ = runProfile(t, source, "env", "status", "--json")
		rows, _ := decodeTrustDoc(t, stdout)
		if len(rows) != 1 {
			t.Fatalf("shell_hook_trust = %v, want the one %s row", rows, file)
		}
		row := rows[0]
		if row["path"] != envPath || row["state"] != hookapproval.PostureApproved ||
			row["approved_by"] != hookapproval.ApprovedByOperator || row["file"] != file {
			t.Fatalf("shell_hook_trust row = %v", row)
		}
		if _, present := row["diagnostic"]; present {
			t.Fatalf("%s row claims a diagnostic without a digest comparison: %v", file, row)
		}
		if code, stdout, _ := runProfile(t, source, "env", "status", "--check", "--json"); code != exitFail {
			t.Fatalf("env status --check --json with file %s = %d, want %d\n%s", file, code, exitFail, stdout)
		} else {
			assertCheckTrustRow(t, "env status --check --json", stdout, envPath, file)
		}
	}

	// R1: the approved file disappears; its row stays.
	if err := os.Remove(envPath); err != nil {
		t.Fatal(err)
	}
	assertRow(t, hookapproval.PostureFileMissing, "file is missing")

	// Restoring the identical bytes re-approves by digest: --check green.
	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check after restore = %d, want %d\n%s", code, exitOK, stdout)
	}

	// R2: the approved file becomes unreadable-as-bytes; its record stays.
	if err := os.Remove(envPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(envPath, 0o700); err != nil {
		t.Fatal(err)
	}
	assertRow(t, hookapproval.PostureFileUnreadable, "file is unreadable")
}

// TestEnvStatusUnreadableApprovalStateSurfaced proves R3 on curator env
// status with an otherwise-current matrix: an unreadable approval state
// surfaces in text and JSON (never dropped) and fails --check, while a
// truly absent state stays the normal warning case.
func TestEnvStatusUnreadableApprovalStateSurfaced(t *testing.T) {
	// No t.Parallel: profileHome sets process environment, exactly like
	// the other env matrix tests.
	source, home := provisionedEnvMatrix(t)
	approveSource := stubConfigSource{path: source.path, cfg: source.cfg}
	_, _, envPath := hookProject(t, "export CURATOR_PROJECT_ENV=1\n")
	if code, _, stderr := runProfile(t, approveSource, "hook", "approve", envPath); code != exitOK {
		t.Fatalf("hook approve = %d\nstderr:\n%s", code, stderr)
	}
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check over an approved file = %d, want %d\n%s", code, exitOK, stdout)
	}
	// A directory at the state file's own path is unreadable-as-bytes on
	// every platform (see TestScanRejectsUnreadableState).
	if err := os.Remove(hookapproval.ApprovalsPath(home)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(hookapproval.ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}

	code, stdout, _ := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\n%s", code, stdout)
	}
	for _, want := range []string{"cannot read", hookapproval.Filename} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("env status hides the state read failure %q:\n%s", want, stdout)
		}
	}
	if code, _, _ := runProfile(t, source, "env", "status", "--check"); code != exitFail {
		t.Fatalf("env status --check over an unreadable state = %d, want %d", code, exitFail)
	}

	code, stdout, _ = runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("env status --json = %d\n%s", code, stdout)
	}
	_, decoded := decodeTrustDoc(t, stdout)
	joined := strings.Join(func() []string {
		var out []string
		for _, warning := range trustWarningsOf(t, decoded) {
			text, _ := warning.(string)
			out = append(out, text)
		}
		return out
	}(), "\n")
	for _, want := range []string{"cannot read", hookapproval.Filename} {
		if !strings.Contains(joined, want) {
			t.Fatalf("shell_hook_trust_warnings hides %q: %v", want, joined)
		}
	}
	if code, stdout, _ := runProfile(t, source, "env", "status", "--check", "--json"); code != exitFail {
		t.Fatalf("env status --check --json over an unreadable state = %d, want %d\n%s", code, exitFail, stdout)
	} else if !strings.Contains(stdout, "cannot read") {
		t.Fatalf("env status --check --json drops the state read failure:\n%s", stdout)
	}

	// Paired control: removing the unreadable state restores the absent
	// case — --check green, no warnings key.
	if err := os.Remove(hookapproval.ApprovalsPath(home)); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := runProfile(t, source, "env", "status", "--check"); code != exitOK {
		t.Fatalf("env status --check over an absent state = %d, want %d", code, exitOK)
	}
	code, stdout, _ = runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitOK {
		t.Fatalf("env status --check --json over an absent state = %d, want %d\n%s", code, exitOK, stdout)
	}
	rows, decoded := decodeTrustDoc(t, stdout)
	if len(rows) != 0 {
		t.Fatalf("shell_hook_trust = %v, want no rows without records or project candidates", rows)
	}
	if _, present := decoded["shell_hook_trust_warnings"]; present {
		t.Fatalf("absent state carries shell_hook_trust_warnings:\n%s", stdout)
	}
	if strings.Contains(stdout, "cannot read") {
		t.Fatalf("absent state reports a read failure:\n%s", stdout)
	}
}

// TestHookApprovalsUnreadableStateNamesReadFailure proves the approvals
// listing names a state read failure as such (never as unapproved) and
// fails without printing rows.
func TestHookApprovalsUnreadableStateNamesReadFailure(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	configPath := filepath.Join(home, "config.json")
	if err := os.MkdirAll(hookapproval.ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, configPath, "hook", "approvals")
	if code != exitFail {
		t.Fatalf("hook approvals over an unreadable state = %d, want %d", code, exitFail)
	}
	if !strings.Contains(stderr, "cannot list approvals") || !strings.Contains(stderr, "read state") {
		t.Fatalf("approvals stderr does not name the read failure:\n%s", stderr)
	}
	if strings.Contains(stderr, "unapproved") || stdout != "" {
		t.Fatalf("approvals over an unreadable state = %q %q", stdout, stderr)
	}
}
