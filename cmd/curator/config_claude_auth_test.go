//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// The real CLI uses the selected source, produces a private candidate without
// opening the absent token file, applies exact bytes, and restores its backup.
func TestConfigClaudeAuthRoundTrip(t *testing.T) {
	dir := t.TempDir()
	policy := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
	path := filepath.Join(dir, "override.json")
	t.Setenv("CURATOR_CONFIG", path)
	original := []byte(`{"schema_version":2,"skills_root":"x","projects":{},"preferred_locale":"en"}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(dir, "absent-token")
	candidate := filepath.Join(dir, "candidate.json")
	invoke := func(args ...string) string {
		t.Helper()
		var out, stderr bytes.Buffer
		code := run(append([]string{"config", "claude-auth"}, args...), fileConfigSource(config.UserPath()), &out, &stderr)
		if code != exitOK || stderr.Len() != 0 {
			t.Fatalf("CLI failed: %d %s", code, stderr.String())
		}
		return out.String()
	}
	var plan struct {
		config.ClaudeAuthEditPlan
		SelectionSource string `json:"selection_source"`
	}
	if err := json.Unmarshal([]byte(invoke("plan", "--file", store, "--out", candidate)), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.ConfigPath != path || plan.SelectionSource != "CURATOR_CONFIG" || plan.SchemaFrom != 2 || plan.SchemaTo != 7 {
		t.Fatalf("wrong provenance: %+v", plan)
	}
	if _, err := os.Stat(store); !os.IsNotExist(err) {
		t.Fatal("token file touched")
	}
	var applied config.ClaudeAuthEditResult
	if err := json.Unmarshal([]byte(invoke("apply", "--file", store, "--expect-source-sha256", plan.SourceSHA256, "--expect-result-sha256", plan.CandidateSHA256, "--expect-policy-sha256", plan.PolicySHA256)), &applied); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want, _ := os.ReadFile(candidate)
	if applied.Status != "applied" || !bytes.Equal(got, want) {
		t.Fatal("did not apply exact candidate")
	}
	invoke("restore", "--backup", applied.BackupPath, "--expect-current-sha256", applied.ResultSHA256)
	got, _ = os.ReadFile(path)
	if !bytes.Equal(got, original) {
		t.Fatal("restore changed original bytes")
	}
}

// Invalid CLI forms reject before touching source/store and never reflect
// supplied hostile values, including unknown/repeated or mixed store flags.
func TestConfigClaudeAuthUsagePrivacy(t *testing.T) {
	for _, args := range [][]string{
		{"plan", "--token", "HOSTILE_SECRET_VALUE"},
		{"plan", "--file", "/absent", "--keychain-service", "HOSTILE_SECRET_VALUE"},
		{"plan", "--file", "/one", "--file", "/two", "--out", "unused"},
		{"restore", "--file", "HOSTILE_SECRET_VALUE"},
		{"apply", "operand-HOSTILE_SECRET_VALUE"},
	} {
		var out, stderr bytes.Buffer
		code := run(append([]string{"config", "claude-auth"}, args...), fileConfigSource("absent"), &out, &stderr)
		if code != exitUsage || out.Len() != 0 || strings.Contains(stderr.String(), "HOSTILE_SECRET") {
			t.Fatalf("bad refusal %d %q %q", code, out.String(), stderr.String())
		}
	}
}

// Invalid source metadata must not be copied into a candidate/backup or
// reflected in diagnostics. This exercises the real CLI write boundary.
func TestConfigClaudeAuthSourcePrivacy(t *testing.T) {
	dir := t.TempDir()
	policy := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
	for _, raw := range []string{
		`{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"claude_auth":{"type":"oauth","token":"HOSTILE_SECRET_VALUE","store":{"type":"file","path":"/absent"}}}}`,
		`{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"s","account":"a","path":"HOSTILE_SECRET_VALUE"}}}}`,
	} {
		path := filepath.Join(t.TempDir(), "machine.json")
		candidate := path + ".candidate"
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		var out, stderr bytes.Buffer
		code := run([]string{"config", "claude-auth", "plan", "--file", "/absent-token", "--out", candidate}, fileConfigSource(path), &out, &stderr)
		if code != exitFail || out.Len() != 0 || strings.Contains(stderr.String(), "HOSTILE_SECRET") {
			t.Fatal("source disclosed or accepted")
		}
		got, _ := os.ReadFile(path)
		files, _ := filepath.Glob(path + ".bak-*")
		if !bytes.Equal(got, []byte(raw)) || len(files) != 0 {
			t.Fatal("refusal changed state")
		}
		if _, err := os.Lstat(candidate); !os.IsNotExist(err) {
			t.Fatal("refusal wrote candidate")
		}
	}
}

type claudeAuthBrokenOutput struct{}

func (claudeAuthBrokenOutput) Write([]byte) (int, error) {
	return 0, errors.New("injected output failure")
}

// A failed stdout after successful rename reports applied on stderr and
// returns failure without claiming preservation or deleting recovery bytes.
func TestConfigClaudeAuthAppliedReportingFailure(t *testing.T) {
	dir := t.TempDir()
	policy := filepath.Join(dir, "policy.json")
	path := filepath.Join(dir, "machine.json")
	if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
	original := []byte(`{"schema_version":2,"skills_root":"x","projects":{}}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	selection := config.LauncherClaudeAuth{Type: "oauth", Store: config.LauncherAuthStore{Type: "file", Path: "/absent-token"}}
	plan, err := config.PlanClaudeAuth(path, selection)
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	code := run([]string{"config", "claude-auth", "apply", "--file", "/absent-token", "--expect-source-sha256", plan.SourceSHA256, "--expect-result-sha256", plan.CandidateSHA256, "--expect-policy-sha256", plan.PolicySHA256}, fileConfigSource(path), claudeAuthBrokenOutput{}, &stderr)
	if code != exitFail || !strings.Contains(stderr.String(), "applied (result reporting failed)") {
		t.Fatal("committed state misreported", stderr.String())
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, plan.CandidateBytes()) {
		t.Fatal("rename was not committed")
	}
	files, _ := filepath.Glob(path + ".bak-*")
	if len(files) != 1 {
		t.Fatal("backup missing")
	}
	saved, _ := os.ReadFile(files[0])
	if !bytes.Equal(saved, original) {
		t.Fatal("backup changed")
	}
}

// The Keychain CLI form encodes references only; the private candidate remains
// consumable by the ordinary launcher projection without credential lookup.
func TestConfigClaudeAuthKeychainReferences(t *testing.T) {
	dir := t.TempDir()
	policy := filepath.Join(dir, "system.json")
	path := filepath.Join(dir, "machine.json")
	candidate := filepath.Join(dir, "candidate.json")
	if err := os.WriteFile(policy, []byte(`{"schema_version":2,"locked":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
	if err := os.WriteFile(path, []byte(`{"schema_version":2,"skills_root":"x","projects":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	code := run([]string{"config", "claude-auth", "plan", "--keychain-service", "fixture-service", "--keychain-account", "fixture-account", "--out", candidate}, fileConfigSource(path), &out, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatal("keychain plan refused", code, stderr.String())
	}
	var plan config.ClaudeAuthEditPlan
	if err := json.Unmarshal(out.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	code = run([]string{"config", "claude-auth", "apply", "--keychain-service", "fixture-service", "--keychain-account", "fixture-account", "--expect-source-sha256", plan.SourceSHA256, "--expect-result-sha256", plan.CandidateSHA256, "--expect-policy-sha256", plan.PolicySHA256}, fileConfigSource(path), &out, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatal("keychain apply refused", code, stderr.String())
	}
	cfg, err := config.Load(path, nil)
	if err != nil || cfg.LauncherProjection().Extensions.ClaudeAuth.Store.Service != "fixture-service" || cfg.Extensions.ClaudeAuth.Store.Account != "fixture-account" {
		t.Fatal("keychain references lost", err)
	}
}
