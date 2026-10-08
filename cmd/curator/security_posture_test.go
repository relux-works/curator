package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
)

func decodeSecurityPostureRows(t testing.TB, document string) []map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(document), &payload); err != nil {
		t.Fatalf("status document is not a JSON object: %v\n%s", err, document)
	}
	values, ok := payload["security_posture_rows"].([]any)
	if !ok {
		t.Fatalf("security_posture_rows = %#v, want rows", payload["security_posture_rows"])
	}
	rows := make([]map[string]any, 0, len(values))
	for _, value := range values {
		row, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("posture row = %#v, want object", value)
		}
		rows = append(rows, row)
	}
	return rows
}

func requirePermissivePostureWarning(t testing.TB, stderr string) {
	t.Helper()
	if got := strings.Count(stderr, config.DiagSecurityPosturePermissive); got != 1 {
		t.Fatalf("permissive posture warning count = %d, want once per operation:\n%s", got, stderr)
	}
}

func TestSecurityPostureWarningAndEnvStatusRows(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`)
	code, stdout, stderr := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("env status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requirePermissivePostureWarning(t, stderr)
	for _, want := range []string{"security_posture", config.SecurityPostureMigrationHint} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("permissive warning misses %q:\n%s", want, stderr)
		}
	}
	rows := decodeSecurityPostureRows(t, stdout)
	if len(rows) != 14 || rows[0]["gate"] != "security_posture" ||
		rows[0]["value"] != config.SecurityPosturePermissive || rows[0]["source"] != "explicit" {
		t.Fatalf("env status posture rows = %#v", rows)
	}
}

func TestSecurityPostureWarningIsNotEmittedByEnforcedShim(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`)
	t.Setenv("CURATOR_CONFIG", source.path)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	previousStderr := os.Stderr
	os.Stderr = writer
	code := runEnforcedShim(filepath.Join(t.TempDir(), "curator"), nil)
	_ = writer.Close()
	os.Stderr = previousStderr
	if code != exitFail {
		t.Fatalf("enforced shim without a sidecar = %d, want %d", code, exitFail)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read enforced shim stderr: %v", err)
	}
	if strings.Contains(string(output), config.DiagSecurityPosturePermissive) {
		t.Fatalf("enforced shim leaked the permissive warning to its stderr: %s", output)
	}
}

func TestSecurityPostureHardenedContradictionFailsEnvStatusCheck(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"hardened","skills_root":"x","projects":{}}`)
	code, stdout, stderr := runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitFail {
		t.Fatalf("env status --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	var status struct {
		NonCurrent          bool             `json:"non_current"`
		SecurityPostureRows []map[string]any `json:"security_posture_rows"`
	}
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("env status --check JSON: %v\n%s", err, stdout)
	}
	if !status.NonCurrent {
		t.Fatalf("hardened empty source allowlist reports current: %s", stdout)
	}
	if len(status.SecurityPostureRows) != 14 || status.SecurityPostureRows[0]["value"] != "hardened" {
		t.Fatalf("hardened posture rows = %#v", status.SecurityPostureRows)
	}
}

func TestSecurityPostureSchema1EnvStatusHasNoPostureInventory(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":1,"skills_root":"x","projects":{}}`)
	code, stdout, stderr := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("schema-1 env status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requirePermissivePostureWarning(t, stderr)
	var status struct {
		SecurityPostureRows []any `json:"security_posture_rows"`
	}
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("schema-1 env status JSON: %v\n%s", err, stdout)
	}
	if status.SecurityPostureRows != nil {
		t.Fatalf("schema-1 env status posture rows = %#v, want null", status.SecurityPostureRows)
	}
}

func TestSecurityPostureCuratorStatusRows(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	project, home := legacyProject(t)
	configPath := filepath.Join(home, "config.json")
	data := map[string]any{
		"schema_version":   2,
		"security_posture": "permissive",
		"skills_root":      filepath.Join(filepath.Dir(home), "skills"),
		"projects": map[string]any{
			"app": map[string]any{"path": project, "agents": []any{"codex_cli"}},
		},
	}
	contents, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runProfile(t, stubConfigSource{path: configPath, cfg: cfg}, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("curator status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	rows := decodeSecurityPostureRows(t, stdout)
	if len(rows) != 14 || rows[0]["gate"] != "security_posture" || rows[0]["value"] != "permissive" {
		t.Fatalf("curator status posture rows = %#v", rows)
	}
	requirePermissivePostureWarning(t, stderr)
}

func TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"hardened","allowed_sources":[],"skills_root":"x","projects":{}}`)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	code, stdout, stderr := runProfile(t, source, "profile", "install", pkg)
	if code != exitFail || !strings.Contains(stderr, config.DiagSourceAllowlistEmpty) {
		t.Fatalf("hardened profile install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Lstat(envprofile.ProfileDir(source.cfg.Home(), "acme")); !os.IsNotExist(err) {
		t.Fatalf("refused install published profile state: %v", err)
	}
}

func TestSecurityPostureProfileInstallRefusesMCPDeclarations(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	requireGit(t)
	// Keep source-signature enforcement explicit so this test reaches the
	// independent hardened MCP contradiction it is intended to exercise.
	configSource := writeMachineConfig(t, `{"schema_version":2,"security_posture":"hardened",`+
		`"allowed_sources":["example.com"],"skills_root":"x","projects":{},`+
		`"environments":{"mcp_package_allowlist":[],"require_source_signers":false}}`)
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init")
	commitGitRepo(t, tool, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version":1,"name":"withmcp","version":"1.0.0",`+
		`"requires":{"mcp":{"tool":{"git":"https://example.com/mcp-tool","range":"*"}}}}`+"\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{"https://example.com/mcp-tool": tool, "https://example.com/withmcp": root})
	code, stdout, stderr := runProfile(t, configSource, "profile", "install", "https://example.com/withmcp")
	if code != exitFail || !strings.Contains(stderr, config.DiagMCPAllowlistEmpty) {
		t.Fatalf("hardened MCP profile install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Lstat(envprofile.ProfileDir(configSource.cfg.Home(), "withmcp")); !os.IsNotExist(err) {
		t.Fatalf("refused MCP install published profile state: %v", err)
	}
}

func TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	requireGit(t)
	vector := cliSecurityPostureVector{
		Machine: map[string]any{
			"allowed_sources": []any{"https://example.com/skills"},
			"environments": map[string]any{
				"mcp_package_allowlist": []any{"https://example.com/mcp"},
			},
		},
	}
	vector.Operation.MCPDeclarationsPresent = true
	root := securityPostureVectorPackage(t, stubConfigSource{}, vector)
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","allowed_sources":["example.com/skills","example.com/mcp"],`+
		`"skills_root":"x","projects":{},"environments":{"mcp_package_allowlist":["example.com/mcp"]}}`)
	if code, stdout, stderr := runProfile(t, source, "profile", "install", root); code != exitOK {
		t.Fatalf("permissive MCP profile install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lockedConfig := map[string]any{
		"schema_version":   2,
		"security_posture": config.SecurityPostureHardened,
		"allowed_sources":  []string{"example.com/skills", "example.com/mcp"},
		"skills_root":      source.cfg.SkillsRoot,
		"projects":         map[string]any{},
		"environments": map[string]any{
			"mcp_package_allowlist": []string{},
		},
	}
	writePostureVectorFile(t, filepath.Dir(source.path), filepath.Base(source.path), lockedConfig)
	source = reloadSource(t, source)
	code, stdout, stderr := runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitFail {
		t.Fatalf("hardened empty MCP allowlist status --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	var status struct {
		NonCurrent      bool                          `json:"non_current"`
		Warnings        []string                      `json:"warnings"`
		MCPDeclarations []envprofile.DeclarationScope `json:"mcp_declarations"`
	}
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("env status --check JSON: %v\n%s", err, stdout)
	}
	if !status.NonCurrent {
		t.Fatalf("hardened MCP contradiction reports current: %s", stdout)
	}
	declarations := false
	for _, scope := range status.MCPDeclarations {
		declarations = declarations || len(scope.Rows) > 0
	}
	if !declarations {
		t.Fatalf("fixture status contains no MCP declaration rows: %s", stdout)
	}
	if !strings.Contains(strings.Join(status.Warnings, "\n"), config.DiagMCPAllowlistEmpty) {
		t.Fatalf("hardened empty MCP allowlist status lacks its diagnostic: %+v", status.Warnings)
	}
}

func TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"hardened",`+
		`"allowed_sources":["example.com"],"skills_root":"x","projects":{},`+
		`"environments":{"mcp_package_allowlist":[]}}`)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	code, stdout, stderr := runProfile(t, source, "profile", "install", pkg)
	if code != exitOK {
		t.Fatalf("hardened declaration-free install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, config.DiagMCPAllowlistEmpty) {
		t.Fatalf("hardened declaration-free install lacks the non-refusing empty-allowlist warning:\n%s", stderr)
	}
	if strings.Contains(stderr, config.DiagSecurityPosturePermissive) {
		t.Fatalf("hardened install received the permissive warning:\n%s", stderr)
	}
}

func TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"hardened",`+
		`"allowed_sources":["example.com"],"skills_root":"x","projects":{},`+
		`"environments":{"mcp_package_allowlist":["https://example.com/mcp"],"passable_env_names":null}}`)
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair")
	if code != exitFail || !strings.Contains(stderr, config.DiagPassableEnvUnbounded) {
		t.Fatalf("hardened resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("refused resolve emitted a fragment: %q", stdout)
	}
	if _, err := os.Lstat(filepath.Join(source.cfg.Home(), "environments")); !os.IsNotExist(err) {
		t.Fatalf("refused resolve wrote an environment home: %v", err)
	}
	code, stdout, stderr = runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitFail {
		t.Fatalf("hardened env status --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	var status struct {
		NonCurrent bool `json:"non_current"`
	}
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("hardened env status --check JSON: %v\n%s", err, stdout)
	}
	if !status.NonCurrent {
		t.Fatalf("hardened explicit unbounded passthrough reports current: %s", stdout)
	}
}

// Drive run() -> config.Load -> the production operation gates with an
// omitted posture. No test helper supplies a hardened posture on their behalf.
func TestSecurityPostureRevisionBDefaultRefusesEmptySources(t *testing.T) {
	for _, allowlist := range []struct{ name, member string }{
		{"absent", ""}, {"explicit-empty", `,"allowed_sources":[]`},
	} {
		for _, command := range []string{"install", "update", "upgrade", "profile-install"} {
			t.Run(allowlist.name+"/"+command, func(t *testing.T) {
				t.Setenv("CURATOR_SYSTEM_CONFIG", "")
				source := writeMachineConfig(t, `{"schema_version":2,"skills_root":"x","projects":{}`+allowlist.member+`}`)
				args := []string{command, "app"}
				if command == "profile-install" {
					pkg := t.TempDir()
					writeContextPackage(t, pkg, "posture", "1.0.0", "posture test\n")
					args = []string{"profile", "install", pkg}
				}
				code, stdout, stderr := runProfile(t, source, args...)
				if code != exitFail || !strings.Contains(stderr, config.DiagSourceAllowlistEmpty) {
					t.Fatalf("%s = %d, want source refusal\nstdout:\n%s\nstderr:\n%s", command, code, stdout, stderr)
				}
				if strings.Contains(stderr, config.DiagSecurityPosturePermissive) {
					t.Fatalf("default hardened operation emitted permissive warning: %s", stderr)
				}
				if _, err := os.Lstat(filepath.Join(source.cfg.Home(), "profiles")); !os.IsNotExist(err) {
					t.Fatalf("refusal published profile state: %v", err)
				}
			})
		}
	}
}

func TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{},"allowed_sources":[]}`)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "posture", "1.0.0", "permissive profile\n")
	code, stdout, stderr := runProfile(t, source, "profile", "install", pkg)
	if code != exitOK {
		t.Fatalf("explicit permissive install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requirePermissivePostureWarning(t, stderr)
	if _, err := os.Lstat(envprofile.ProfileDir(source.cfg.Home(), "posture")); err != nil {
		t.Fatalf("permissive install did not publish profile: %v", err)
	}
}

func TestSecurityPostureRevisionBDefaultEnvStatus(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	source := writeMachineConfig(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
	for _, check := range []bool{false, true} {
		args := []string{"env", "status", "--json"}
		wantCode := exitOK
		if check {
			args = append(args, "--check")
			wantCode = exitFail
		}
		code, stdout, stderr := runProfile(t, source, args...)
		if code != wantCode {
			t.Fatalf("env status check=%t = %d, want %d: %s", check, code, wantCode, stderr)
		}
		rows := decodeSecurityPostureRows(t, stdout)
		if len(rows) != 14 || rows[0]["value"] != "hardened" || rows[0]["source"] != "profile" {
			t.Fatalf("default posture rows = %#v", rows)
		}
		if strings.Contains(stderr, config.DiagSecurityPosturePermissive) {
			t.Fatalf("hardened status warned permissive: %s", stderr)
		}
	}
}
