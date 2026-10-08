package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
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
	// Current writers publish schema 3 with hash_version 2 (Spec
	// environments §8.2); the record fields below are unchanged.
	if marker.Version != envmarker.VersionV3 || marker.HashVersion != 2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected one schema-3 credential record, got %+v", marker)
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
	if marker.Version != envmarker.VersionV3 || marker.HashVersion != 2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected a schema-3 pathless record: %+v", marker)
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
	source := writeMachineConfig(t, `{"schema_version": 2, "security_posture": "permissive", "skills_root": "x", "projects": {}}`)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "isolation", `{"acme": {"claude_code": "isolated"}}`); code != exitOK {
		t.Fatalf("set isolated mode = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	_, payload, marker := cliEnvMarker(t, source, "claude_code")
	if marker.Version != envmarker.VersionV3 || marker.HashVersion != 2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("expected one schema-3 credential record: %+v", marker)
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
	// The bytes-kept schema-1 repair is the frozen v1-lane behavior:
	// current writers MUST publish schema 3 (environments §8.2), and a
	// v1 marker over a v2 lock is a refused version mismatch. The
	// legacy lane stays pinned for the whole test — install,
	// provision, downgrade to genuine v1 bytes, repair — so the
	// downgrade never relabels v2 digests under a v1 schema.
	pinV1MarkerWriters(t)
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
	// Schemas 1 and 2 carry no hash_version (environments §8.2); the
	// published schema-3 value must not survive the downgrade.
	delete(object, "hash_version")
	// Preserve the historical A marker independently of the shipped seed rule.
	object["codex_seed_record"] = map[string]any{"revision": "A", "native_mcp_servers": []string{}}
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
	legacyMarker, err := envmarker.Parse(legacyBytes)
	if err != nil {
		t.Fatalf("schema-1 fixture must parse: %v", err)
	}
	if record := legacyMarker.CodexSeedRecord; record == nil || record.Revision != envregistry.CodexSeedRevisionA || record.NativeMCPServers == nil || len(record.NativeMCPServers) != 0 {
		t.Fatalf("schema-1 fixture lost the required empty revision-A seed record: %+v", record)
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

func TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome(t *testing.T) {
	// Same frozen v1 lane as TestEnvResolveKeepsSchema1BytesForMetadataOnly:
	// current writers MUST publish schema 3 (environments §8.2), so the
	// pre-rule schema-1 home under test must be provisioned, downgraded,
	// and repaired with the v1 writer to stay genuine.
	pinV1MarkerWriters(t)
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
	// Schemas 1 and 2 carry no hash_version (environments §8.2); the
	// published schema-3 value must not survive the downgrade.
	delete(object, "hash_version")
	delete(object, "codex_seed_record")
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
	legacyMarker, err := envmarker.Parse(legacyBytes)
	if err != nil {
		t.Fatalf("pre-rule schema-1 fixture must parse: %v", err)
	}
	if legacyMarker.CodexSeedRecord != nil {
		t.Fatalf("pre-rule schema-1 fixture unexpectedly has a Codex seed record: %+v", legacyMarker.CodexSeedRecord)
	}
	if err := os.WriteFile(markerPath, legacyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	managedConfig := filepath.Join(filepath.Dir(markerPath), "config.toml")
	legacyConfig := "model = \"gpt-5-codex\"\n\n[mcp_servers.legacy]\ncommand = \"server\"\nargs = []\n"
	if err := os.WriteFile(managedConfig, []byte(legacyConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(filepath.Dir(markerPath), "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("pre-rule schema-1 repair = %d\nstderr:\n%s", code, stderr)
	}
	after, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(legacyBytes) {
		t.Fatalf("metadata-only repair changed the pre-rule schema-1 marker:\n%s", after)
	}
	afterConfig, err := os.ReadFile(managedConfig)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterConfig) != legacyConfig {
		t.Fatalf("metadata-only repair changed the pre-rule Codex seed:\n%s", afterConfig)
	}
	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "mcp_seed_unstripped: this managed Codex home predates the seed record; re-provision it to apply seed revision B") {
		t.Fatalf("env status did not report the pre-rule Codex home:\n%s", stdout)
	}
}
