package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
)

func installCLIEnvProfile(t *testing.T, source stubConfigSource) {
	t.Helper()
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
}

func cliEnvMarker(t *testing.T, source stubConfigSource, env string) (string, []byte, *envmarker.Marker) {
	t.Helper()
	path := filepath.Join(source.cfg.Home(), "environments", "acme", env, envmarker.Name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := envmarker.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	return path, payload, marker
}

func requireCLICodexRecord(t *testing.T, marker *envmarker.Marker, provenance, backend string) {
	t.Helper()
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected one schema-2 credential record, got %+v", marker)
	}
	record := (*marker.Passthrough)[0]
	if record.Path != "auth.json" || record.Isolation != "shared" || record.Strategy != "keyring-preferred" ||
		record.SourceRole != "native" || record.Backend != backend || record.BackendVersion != "0.153.2" || record.Provenance != provenance {
		t.Fatalf("Codex credential record = %+v", record)
	}
}

// TestEnvResolveCredentialRecordProvisionAndRepair drives both mutation
// origins through the CLI production entry point.
func TestEnvResolveCredentialRecordProvisionAndRepair(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair provision = %d\nstderr:\n%s", code, stderr)
	}
	markerPath, _, marker := cliEnvMarker(t, source, "codex_cli")
	requireCLICodexRecord(t, marker, "provisioned", "file")
	if err := os.Remove(filepath.Join(filepath.Dir(markerPath), "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair repair = %d\nstderr:\n%s", code, stderr)
	}
	_, _, marker = cliEnvMarker(t, source, "codex_cli")
	requireCLICodexRecord(t, marker, "repaired", "file")
}

// TestEnvResolveCredentialRecordPathlessKeyring proves a keyring backend is
// recorded as ambient without a path through the CLI production entry point.
func TestEnvResolveCredentialRecordPathlessKeyring(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	if err := os.WriteFile(filepath.Join(os.Getenv("CODEX_HOME"), "config.toml"), []byte("cli_auth_credentials_store = \"keyring\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	_, payload, marker := cliEnvMarker(t, source, "codex_cli")
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected a schema-2 pathless record: %+v", marker)
	}
	record := (*marker.Passthrough)[0]
	if record.Path != "" || record.Isolation != "shared" || record.Strategy != "keyring-preferred" || record.SourceRole != "native" || record.Backend != "ambient" || record.BackendVersion != "0.153.2" || record.Provenance != "provisioned" {
		t.Fatalf("keyring credential record = %+v", record)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	entries, _ := object["passthrough"].([]any)
	fields, _ := entries[0].(map[string]any)
	if _, exists := fields["path"]; exists {
		t.Fatalf("the keyring record omits path: %s", payload)
	}
}

// TestEnvResolveCredentialRecordIsolatedKeychain drives the other supported
// linkless passthrough strategy through the CLI production entry point.
func TestEnvResolveCredentialRecordIsolatedKeychain(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Claude's per-home Keychain strategy is supported on macOS")
	}
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "isolation", `{"acme": {"claude_code": "isolated"}}`); code != exitOK {
		t.Fatalf("set isolated mode = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	_, payload, marker := cliEnvMarker(t, source, "claude_code")
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected one schema-2 credential record: %+v", marker)
	}
	record := (*marker.Passthrough)[0]
	if record.Path != "" || record.Isolation != "isolated" || record.Strategy != "per-home-keychain" ||
		record.SourceRole != "managed" || record.Backend != "keychain" || record.BackendVersion != "2.1.261" || record.Provenance != "provisioned" {
		t.Fatalf("isolated Claude credential record = %+v", record)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	entries, _ := object["passthrough"].([]any)
	fields, _ := entries[0].(map[string]any)
	if _, exists := fields["path"]; exists {
		t.Fatalf("the per-home Keychain record omits path: %s", payload)
	}
}

// TestEnvResolveKeepsSchema1BytesForMetadataOnly proves an explicit repair
// does not upgrade a schema-1 marker when its legacy state is unchanged.
func TestEnvResolveKeepsSchema1BytesForMetadataOnly(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair provision = %d\nstderr:\n%s", code, stderr)
	}
	markerPath, payload, _ := cliEnvMarker(t, source, "codex_cli")
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	object["version"] = float64(envmarker.VersionV1)
	modern, _ := object["passthrough"].([]any)
	legacy := make([]any, 0, len(modern))
	for _, value := range modern {
		fields, _ := value.(map[string]any)
		path, _ := fields["path"].(string)
		strategy, _ := fields["strategy"].(string)
		if strategy == "keyring-preferred" && fields["backend"] == "file" {
			strategy = "file-link"
		}
		legacy = append(legacy, map[string]any{"path": path, "strategy": strategy})
	}
	object["passthrough"] = legacy
	legacyBytes, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	legacyBytes = append(legacyBytes, '\n')
	if _, err := envmarker.Parse(legacyBytes); err != nil {
		t.Fatalf("schema-1 fixture must parse: %v", err)
	}
	if err := os.WriteFile(markerPath, legacyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(filepath.Dir(markerPath), "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("schema-1 repair = %d\nstderr:\n%s", code, stderr)
	}
	after, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(legacyBytes) {
		t.Fatalf("metadata-only upgrade changed schema-1 marker bytes:\n%s", after)
	}
}
