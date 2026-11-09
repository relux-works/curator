//go:build unix

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func authEditFixture(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	policy := filepath.Join(dir, "system.json")
	if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
	path := filepath.Join(dir, "machine.json")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func authEditSelection() LauncherClaudeAuth {
	return LauncherClaudeAuth{Type: "oauth", Store: LauncherAuthStore{Type: "keychain", Service: "fixture-service", Account: "fixture-account"}}
}

func authEditRequest(t *testing.T, path string) ClaudeAuthEditRequest {
	t.Helper()
	p, err := PlanClaudeAuth(path, authEditSelection())
	if err != nil {
		t.Fatal(err)
	}
	return ClaudeAuthEditRequest{Path: path, Selection: authEditSelection(), ExpectSourceSHA256: p.SourceSHA256, ExpectResultSHA256: p.CandidateSHA256, ExpectPolicySHA256: p.PolicySHA256}
}

func authEditBackups(t *testing.T, path string) []string {
	t.Helper()
	files, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// Planning and applying a valid 2/7 carrier preserve unrelated settings and proxy;
// the exact source backup permits a digest-guarded, byte-exact restore.
// The Plan equivalence refusal is defense in depth: its current valid 2/7
// constructor cannot produce a differing non-auth configuration. These tests
// check preservation, not detection of deleting that unreachable Plan check.
func TestClaudeAuthEditRoundTrip(t *testing.T) {
	for _, text := range []string{
		`{"schema_version":2,"skills_root":"x","projects":{}}`,
		`{"schema_version":2,"skills_root":"x","projects":{},"preferred_locale":"en","security_posture":"permissive","environments":{"precedence":{"winner":"lower-weight","placement":"winner-first"}}}`,
		`{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"proxy":{"profile":"private"}}}`,
	} {
		t.Run(text, func(t *testing.T) {
			path := authEditFixture(t, text)
			p, err := PlanClaudeAuth(path, authEditSelection())
			if err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(filepath.Dir(path), "candidate.json")
			if err := WriteClaudeAuthCandidate(p, output); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(output)
			if !bytes.Equal(got, p.CandidateBytes()) {
				t.Fatal("candidate bytes differ")
			}
			info, _ := os.Stat(output)
			if info.Mode().Perm() != 0600 {
				t.Fatal("candidate is not private")
			}
			if err := WriteClaudeAuthCandidate(p, output); err == nil {
				t.Fatal("existing candidate overwritten")
			}
			before, _ := os.ReadFile(path)
			if string(before) != text || len(authEditBackups(t, path)) != 0 {
				t.Fatal("plan changed source or backups")
			}
			r, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
			if err != nil {
				t.Fatal(err)
			}
			if r.Status != "applied" || !r.ReadbackVerified || !r.Durable {
				t.Fatalf("bad applied result: %+v", r)
			}
			actual, _ := os.ReadFile(path)
			if !bytes.Equal(actual, got) {
				t.Fatal("apply did not write planned bytes")
			}
			backup, _ := os.ReadFile(r.BackupPath)
			if string(backup) != text {
				t.Fatal("backup is not exact source")
			}
			cfg, err := Load(path, nil)
			if err != nil || cfg.Schema != 7 || cfg.Extensions.ClaudeAuth.Store.Service != "fixture-service" {
				t.Fatal("ordinary loader cannot use result", err)
			}
			if strings.Contains(text, `"proxy"`) && cfg.Extensions.Proxy.Profile != "private" {
				t.Fatal("proxy lost")
			}
			restored, err := RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: r.BackupPath, ExpectCurrentSHA256: r.ResultSHA256, ExpectBackupSHA256: r.SourceSHA256})
			if err != nil || restored.Status != "restored" || !restored.ReadbackVerified {
				t.Fatal("restore failed", restored, err)
			}
			actual, _ = os.ReadFile(path)
			if string(actual) != text {
				t.Fatal("restore did not preserve original bytes")
			}
			if _, err := os.Stat(r.BackupPath); err != nil {
				t.Fatal("restore removed recovery copy")
			}
		})
	}
}

// Invalid schema/grammar/source metadata refuses before creating any backup;
// diagnostic text must not reflect hostile supplied keys or values.
func TestClaudeAuthEditPreflightRefusals(t *testing.T) {
	for _, text := range []string{
		`{"schema_version":1,"skills_root":"x","projects":{}}`, `{"schema_version":3,"skills_root":"x","projects":{}}`,
		`{"schema_version":5,"skills_root":"x","projects":{}}`, `{"schema_version":6,"skills_root":"x","projects":{}}`,
		`{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":null}}`,
		`{"schema_version":8.0,"skills_root":"x","projects":{}}`, `{"schema_version":8e0,"skills_root":"x","projects":{}}`, `{"schema_version":2,"schema_version":7,"skills_root":"x","projects":{}}`,
		`{"schema_version":2,"skills_root":"x","projects":{},"HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}`,
		`{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"claude_auth":{"type":"oauth","token":"HOSTILE_SECRET_VALUE","store":{"type":"keychain","service":"s","account":"a"}}}}`,
	} {
		t.Run(text, func(t *testing.T) {
			path := authEditFixture(t, text)
			_, err := PlanClaudeAuth(path, authEditSelection())
			if err == nil || strings.Contains(err.Error(), "HOSTILE_SECRET") {
				t.Fatal("invalid configuration accepted or disclosed", err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != text || len(authEditBackups(t, path)) != 0 {
				t.Fatal("refusal changed state")
			}
		})
	}
	for _, kind := range []string{"missing", "directory", "symlink", "mode0640", "system-source", "mixed-store", "relative-file"} {
		t.Run(kind, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			selection := authEditSelection()
			switch kind {
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				target := path + ".target"
				if err := os.Rename(path, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			case "mode0640":
				if err := os.Chmod(path, 0640); err != nil {
					t.Fatal(err)
				}
			case "system-source":
				t.Setenv("CURATOR_SYSTEM_CONFIG", path)
			case "mixed-store":
				selection.Store.Path = "/forbidden"
			case "relative-file":
				selection.Store = LauncherAuthStore{Type: "file", Path: "relative"}
			}
			if _, err := PlanClaudeAuth(path, selection); err == nil {
				t.Fatal("forbidden input accepted")
			}
			if len(authEditBackups(t, path)) != 0 {
				t.Fatal("preflight created backup")
			}
		})
	}
}

// A stale source/result/policy is an exact refusal before backup. A real write
// interleaved after validation is caught by the final recheck; the later rename
// gap is explicitly not an atomic CAS against noncooperating writers.
func TestClaudeAuthEditStaleControls(t *testing.T) {
	for _, kind := range []string{"source", "result", "policy-removed", "policy-changed", "interleaved-source"} {
		t.Run(kind, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			policy := filepath.Join(filepath.Dir(path), "system.json")
			if kind == "policy-changed" {
				if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[],"preferred_locale":"en"}`), 0600); err != nil {
					t.Fatal(err)
				}
				t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
			}
			req := authEditRequest(t, path)
			ops := claudeAuthEditOps{}
			switch kind {
			case "source":
				req.ExpectSourceSHA256 = strings.Repeat("0", 64)
			case "result":
				req.ExpectResultSHA256 = strings.Repeat("0", 64)
			case "policy-removed":
				if err := os.Remove(SystemCandidatePath()); err != nil {
					t.Fatal(err)
				}
			case "policy-changed":
				if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[],"preferred_locale":"fr"}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "interleaved-source":
				ops.beforeRecheck = func() error {
					return os.WriteFile(path, []byte(`{"schema_version":2,"skills_root":"changed","projects":{}}`), 0600)
				}
			}
			before, _ := os.ReadFile(path)
			_, err := applyClaudeAuth(context.Background(), req, ops)
			if !errors.Is(err, ErrClaudeAuthStale) {
				t.Fatalf("wanted stale refusal, got %v", err)
			}
			got, _ := os.ReadFile(path)
			if kind == "interleaved-source" {
				if !strings.Contains(string(got), "changed") {
					t.Fatal("interleaved edit overwritten")
				}
			} else if !bytes.Equal(before, got) {
				t.Fatal("stale refusal changed source")
			}
			if len(authEditBackups(t, path)) != 0 {
				t.Fatal("stale preflight created backup")
			}
		})
	}
}

// Post-backup failures retain exact recovery bytes without replacing source;
// once rename succeeds, a later failure must still report applied.
func TestClaudeAuthEditPhaseOutcomes(t *testing.T) {
	for _, kind := range []string{"interrupt-after-backup", "rename-fails", "after-rename-fails"} {
		t.Run(kind, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			before, _ := os.ReadFile(path)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ops := claudeAuthEditOps{}
			switch kind {
			case "interrupt-after-backup":
				ops.afterBackup = func() error { cancel(); return nil }
			case "rename-fails":
				ops.rename = func(string, string) error { return errors.New("injected rename failure") }
			case "after-rename-fails":
				ops.syncDirectory = func(string) error { return errors.New("injected directory sync failure") }
			}
			r, err := applyClaudeAuth(ctx, authEditRequest(t, path), ops)
			if err == nil || r == nil {
				t.Fatal("failure disappeared", r, err)
			}
			backup, _ := os.ReadFile(r.BackupPath)
			if !bytes.Equal(backup, before) {
				t.Fatal("recovery backup lost or changed")
			}
			got, _ := os.ReadFile(path)
			if kind == "after-rename-fails" {
				if r.Status != "applied" || bytes.Equal(got, before) || !r.ReadbackVerified {
					t.Fatal("committed outcome misreported", r)
				}
			} else {
				if r.Status != "not_applied_backup_retained" || !bytes.Equal(got, before) {
					t.Fatal("source changed before rename", r)
				}
			}
		})
	}
}

// Restore binds both the current and backup digests and cannot overwrite a
// changed source or use an unrelated, malformed, or unsafe recovery file.
func TestClaudeAuthRestoreRefusals(t *testing.T) {
	for _, kind := range []string{"current", "backup-digest", "backup-mode", "interleaved-source"} {
		t.Run(kind, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			r, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
			if err != nil {
				t.Fatal(err)
			}
			req := ClaudeAuthRestoreRequest{Path: path, BackupPath: r.BackupPath, ExpectCurrentSHA256: r.ResultSHA256, ExpectBackupSHA256: r.SourceSHA256}
			ops := claudeAuthEditOps{}
			switch kind {
			case "current":
				req.ExpectCurrentSHA256 = strings.Repeat("0", 64)
			case "backup-digest":
				req.ExpectBackupSHA256 = strings.Repeat("0", 64)
			case "backup-mode":
				if err := os.Chmod(r.BackupPath, 0640); err != nil {
					t.Fatal(err)
				}
			case "interleaved-source":
				ops.beforeRecheck = func() error {
					return os.WriteFile(path, []byte(`{"schema_version":7,"skills_root":"changed","projects":{}}`), 0600)
				}
			}
			before, _ := os.ReadFile(path)
			backup, _ := os.ReadFile(r.BackupPath)
			if _, err := restoreClaudeAuth(context.Background(), req, ops); err == nil {
				t.Fatal("unsafe restore accepted")
			}
			got, _ := os.ReadFile(path)
			if kind == "interleaved-source" {
				if !strings.Contains(string(got), "changed") {
					t.Fatal("newer edit overwritten")
				}
			} else if !bytes.Equal(before, got) {
				t.Fatal("restore refusal changed source")
			}
			gotBackup, _ := os.ReadFile(r.BackupPath)
			if !bytes.Equal(backup, gotBackup) {
				t.Fatal("restore changed recovery copy")
			}
		})
	}
}

// A real last-moment writer cannot be overwritten by the final observed
// freshness check. Apply retains its backup; restore creates no extra backup.
// This does not cover the explicitly documented recheck-to-rename race.
func TestClaudeAuthFinalInterleaving(t *testing.T) {
	for _, operation := range []string{"apply", "restore"} {
		t.Run(operation, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			original, _ := os.ReadFile(path)
			var result *ClaudeAuthEditResult
			var err error
			interleaved := []byte(`{"schema_version":7,"skills_root":"concurrent","projects":{}}`)
			ops := claudeAuthEditOps{beforeFinalRecheck: func() error { return os.WriteFile(path, interleaved, 0600) }}
			if operation == "apply" {
				result, err = applyClaudeAuth(context.Background(), authEditRequest(t, path), ops)
			} else {
				applied, e := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
				if e != nil {
					t.Fatal(e)
				}
				result, err = restoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: applied.BackupPath, ExpectCurrentSHA256: applied.ResultSHA256}, ops)
			}
			if !errors.Is(err, ErrClaudeAuthStale) {
				t.Fatalf("expected exact stale refusal, got %v", err)
			}
			got, _ := os.ReadFile(path)
			if !bytes.Equal(got, interleaved) {
				t.Fatal("concurrent source overwritten")
			}
			backups := authEditBackups(t, path)
			if len(backups) != 1 {
				t.Fatal("recovery backup count changed")
			}
			saved, _ := os.ReadFile(backups[0])
			if !bytes.Equal(saved, original) {
				t.Fatal("recovery backup changed")
			}
			if result == nil || result.Status == "applied" || result.Status == "restored" {
				t.Fatal("refusal reported committed")
			}
			temps, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".claude-auth-*.tmp"))
			if len(temps) != 0 {
				t.Fatal("temporary candidate leaked")
			}
		})
	}
}

// Restore validates the backup again at the final replacement boundary, so a
// replaced or edited saved file cannot silently substitute recovery bytes.
func TestClaudeAuthRestoreFinalBackupChange(t *testing.T) {
	path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
	applied, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	_, err = restoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: applied.BackupPath, ExpectCurrentSHA256: applied.ResultSHA256}, claudeAuthEditOps{beforeFinalRecheck: func() error {
		return os.WriteFile(applied.BackupPath, []byte(`{"schema_version":2,"skills_root":"different","projects":{}}`), 0600)
	}})
	if !errors.Is(err, ErrClaudeAuthStale) {
		t.Fatal("changed backup accepted", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, before) {
		t.Fatal("failed restore changed current file")
	}
}

// The policy token distinguishes absence, creation, identical-byte replacement
// and a different absent path; no real system policy is read by this test.
func TestClaudeAuthPolicyIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	absent, err := claudeAuthPolicyAt(path)
	if err != nil {
		t.Fatal(err)
	}
	other, err := claudeAuthPolicyAt(path + "-other")
	if err != nil || absent == other {
		t.Fatal("absent policy path not bound")
	}
	raw := []byte(`{"schema_version":2,"locked":[]}`)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	present, err := claudeAuthPolicyAt(path)
	if err != nil || present == absent {
		t.Fatal("creation not bound")
	}
	if err := os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	replaced, err := claudeAuthPolicyAt(path)
	if err != nil || replaced == present {
		t.Fatal("identical-byte replacement identity lost")
	}
}

// The normal system overlay still enforces its locked posture while the
// candidate preserves the machine declaration rather than serializing policy.
func TestClaudeAuthEditSystemOverlayPreserved(t *testing.T) {
	path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{},"security_posture":"permissive"}`)
	if err := os.WriteFile(SystemCandidatePath(), []byte(`{"schema_version":2,"locked":["security_posture"],"security_posture":"hardened"}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := PlanClaudeAuth(path, authEditSelection())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(p.CandidateBytes(), []byte(`"security_posture": "permissive"`)) {
		t.Fatal("candidate serialized overlay over machine declaration")
	}
	r, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
	if err != nil || r.Status != "applied" {
		t.Fatal(r, err)
	}
	cfg, err := Load(path, nil)
	if err != nil || cfg.SecurityPosture != SecurityPostureHardened || !cfg.Locked["security_posture"] {
		t.Fatal("effective system protection lost", err)
	}
}

// A multiply linked source cannot acquire backup/write authority; cancellation
// before any write also preserves the exact source and empty backup set.
func TestClaudeAuthEditPrewriteState(t *testing.T) {
	for _, kind := range []string{"hardlink", "cancelled"} {
		t.Run(kind, func(t *testing.T) {
			path := authEditFixture(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			req := authEditRequest(t, path)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "hardlink" {
				if err := os.Link(path, path+".alias"); err != nil {
					t.Fatal(err)
				}
			} else {
				cancel()
			}
			before, _ := os.ReadFile(path)
			_, err := ApplyClaudeAuth(ctx, req)
			if kind == "hardlink" && !errors.Is(err, ErrClaudeAuthEditRefused) {
				t.Fatal("hardlink admission", err)
			}
			if kind == "cancelled" && !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation lost", err)
			}
			got, _ := os.ReadFile(path)
			if !bytes.Equal(got, before) || len(authEditBackups(t, path)) != 0 {
				t.Fatal("prewrite refusal changed state")
			}
		})
	}
}

// Even with a freshly observed current digest, restore cannot discard a later
// unrelated extension change. An auth-only change remains a valid recovery.
func TestClaudeAuthRestorePreservesLaterExtensions(t *testing.T) {
	for _, kind := range []string{"added-proxy", "changed-proxy", "removed-proxy", "auth-only"} {
		t.Run(kind, func(t *testing.T) {
			text := `{"schema_version":7,"skills_root":"x","projects":{}}`
			if kind == "changed-proxy" || kind == "removed-proxy" {
				text = `{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"proxy":{"profile":"original"}}}`
			}
			path := authEditFixture(t, text)
			applied, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
			if err != nil {
				t.Fatal(err)
			}
			current, err := readClaudeAuthFile(path)
			if err != nil {
				t.Fatal(err)
			}
			object, err := claudeAuthObject(current.raw)
			if err != nil {
				t.Fatal(err)
			}
			ext := object["extensions"].(map[string]any)
			switch kind {
			case "added-proxy", "changed-proxy":
				ext["proxy"] = map[string]any{"profile": "later"}
			case "removed-proxy":
				delete(ext, "proxy")
			case "auth-only":
				ext["claude_auth"] = map[string]any{"type": "oauth", "store": map[string]any{"type": "keychain", "service": "later", "account": "fixture"}}
			}
			raw, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			result, err := RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: applied.BackupPath, ExpectCurrentSHA256: claudeAuthSHA(raw)})
			got, _ := os.ReadFile(path)
			if kind == "auth-only" {
				if err != nil || result == nil || result.Status != "restored" || string(got) != text {
					t.Fatal("auth-only recovery refused", result, err)
				}
				return
			}
			if !errors.Is(err, ErrClaudeAuthEditRefused) || result != nil {
				t.Fatalf("later unrelated extension not refused: result=%+v err=%v", result, err)
			}
			if !bytes.Equal(got, raw) {
				t.Fatal("restore discarded later unrelated extension")
			}
			backups := authEditBackups(t, path)
			saved, readErr := os.ReadFile(applied.BackupPath)
			if len(backups) != 1 || readErr != nil || string(saved) != text {
				t.Fatal("refused restore changed recovery set")
			}
		})
	}
}

// A verified published backup remains reported after directory-sync failure,
// including when its later metadata/readback is unavailable. No source rename
// occurs and the existing recovery copy is retained with durability unknown.
func TestClaudeAuthBackupSyncFailureReported(t *testing.T) {
	for _, unreadable := range []bool{false, true} {
		t.Run(map[bool]string{false: "sync-only", true: "sync-and-unreadable"}[unreadable], func(t *testing.T) {
			text := `{"schema_version":2,"skills_root":"x","projects":{}}`
			path := authEditFixture(t, text)
			published := ""
			ops := claudeAuthEditOps{syncBackupDirectory: func(backup string) error {
				published = backup
				if unreadable {
					if err := os.Chmod(backup, 0000); err != nil {
						t.Fatal(err)
					}
				}
				return errors.New("injected backup directory sync failure")
			}}
			result, err := applyClaudeAuth(context.Background(), authEditRequest(t, path), ops)
			if !errors.Is(err, ErrClaudeAuthEditIO) || result == nil || result.BackupPath != published ||
				result.Status != "not_applied_backup_retained" || result.BackupSHA256 != claudeAuthSHA([]byte(text)) || result.Durable || result.ReadbackVerified {
				t.Fatalf("published backup not reported truthfully: result=%+v err=%v", result, err)
			}
			if unreadable {
				if _, readErr := readClaudeAuthFile(published); !errors.Is(readErr, ErrClaudeAuthEditRefused) {
					t.Fatal("readback-failure control not exercised", readErr)
				}
				if err := os.Chmod(published, 0600); err != nil {
					t.Fatal(err)
				}
			}
			current, _ := os.ReadFile(path)
			saved, readErr := os.ReadFile(published)
			if string(current) != text || readErr != nil || string(saved) != text || len(authEditBackups(t, path)) != 1 {
				t.Fatal("backup sync failure changed source or lost exact recovery copy")
			}
		})
	}
}

// Invalid external policy topology refuses with no token; the adjacent identity
// test covers valid absence and a regular file. No real system input is read.
func TestClaudeAuthPolicyReadRefusals(t *testing.T) {
	for _, kind := range []string{"directory", "non-directory-parent"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "policy.json")
			if kind == "directory" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(path, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(path, "policy.json")
			}
			token, err := claudeAuthPolicyAt(path)
			if !errors.Is(err, ErrClaudeAuthEditRefused) || token != "" {
				t.Fatalf("invalid policy topology returned token=%q, error=%v", token, err)
			}
		})
	}
}
