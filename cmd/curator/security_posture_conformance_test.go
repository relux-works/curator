package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
)

type cliSecurityPostureVector struct {
	Name             string                          `json:"name"`
	RolloutRevision  string                          `json:"rollout_revision"`
	Machine          map[string]any                  `json:"machine"`
	System           map[string]any                  `json:"system"`
	ShippedRevisions config.SecurityPostureRevisions `json:"shipped_revisions"`
	Operation        struct {
		Kind                         string   `json:"kind"`
		MCPDeclarationsPresent       bool     `json:"mcp_declarations_present"`
		UnreachableTrustedRegistries []string `json:"unreachable_trusted_registries"`
		ArtifactsWithoutEvidence     []string `json:"artifacts_without_evidence"`
	} `json:"operation"`
	Expected struct {
		Outcome           string                         `json:"outcome"`
		Diagnostics       []cliSecurityPostureDiagnostic `json:"diagnostics"`
		CuratorStatusRows []config.SecurityPostureRow    `json:"curator_status_rows"`
		EnvStatusRows     json.RawMessage                `json:"env_status_rows"`
		Effective         map[string]any                 `json:"effective"`
	} `json:"expected"`
}

type cliSecurityPostureDiagnostic struct {
	Code          string `json:"code"`
	Severity      string `json:"severity"`
	Count         *int   `json:"count,omitempty"`
	NamesKnob     bool   `json:"names_knob,omitempty"`
	MigrationHint bool   `json:"migration_hint,omitempty"`
}

// TestSecurityPostureVectorsThroughCLI drives the current posture vectors
// through run(), then checks production status and operation results. Only
// historical revision-A implicit defaults are bounded; they are not counted
// as driven by substituting explicit permissive configurations.
func TestSecurityPostureVectorsThroughCLI(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "security-posture.json")) // #nosec G304 -- explicit pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []cliSecurityPostureVector `json:"cases"`
	}
	if err := json.Unmarshal(payload, &corpus); err != nil {
		t.Fatal(err)
	}
	tally := conformancecoverage.RunOutcomes(t, "security-posture/vectors", corpus.Cases,
		func(tc cliSecurityPostureVector) string { return tc.Name },
		func(caseT *testing.T, tc cliSecurityPostureVector) conformancecoverage.Observation {
			if reason := cliSecurityPostureVectorBound(tc.Name); reason != "" {
				return conformancecoverage.Observation{BoundReason: reason}
			}
			source := securityPostureCLIFixture(caseT, tc)

			curatorCode, curatorOut, curatorErr := runProfile(caseT, source, "status", "app", "--json")
			if curatorCode != exitOK {
				caseT.Fatalf("curator status = %d\nstdout:\n%s\nstderr:\n%s", curatorCode, curatorOut, curatorErr)
			}
			var curatorStatus struct {
				SecurityPostureRows []config.SecurityPostureRow `json:"security_posture_rows"`
			}
			if err := json.Unmarshal([]byte(curatorOut), &curatorStatus); err != nil {
				caseT.Fatalf("decode curator status: %v\n%s", err, curatorOut)
			}
			assertPostureJSONEqual(caseT, "curator status rows", cliExpectedSecurityPostureRows(tc.Expected.CuratorStatusRows, source.cfg), curatorStatus.SecurityPostureRows)

			envCode, envOut, envErr := runProfile(caseT, source, "env", "status", "--json")
			if envCode != exitOK {
				caseT.Fatalf("env status = %d\nstdout:\n%s\nstderr:\n%s", envCode, envOut, envErr)
			}
			var envStatus struct {
				SecurityPostureRows []config.SecurityPostureRow `json:"security_posture_rows"`
				PassableEnvNames    []string                    `json:"passable_env_names"`
			}
			if err := json.Unmarshal([]byte(envOut), &envStatus); err != nil {
				caseT.Fatalf("decode env status: %v\n%s", err, envOut)
			}
			var wantEnvRows any
			if err := json.Unmarshal(tc.Expected.EnvStatusRows, &wantEnvRows); err != nil {
				caseT.Fatalf("decode expected env status rows: %v", err)
			}
			if rows, ok := wantEnvRows.([]any); ok {
				wantEnvRows = cliExpectedSecurityPostureRows(decodeCLISecurityPostureRows(caseT, rows), source.cfg)
			}
			assertPostureJSONEqual(caseT, "env status rows", wantEnvRows, envStatus.SecurityPostureRows)
			if want, ok := tc.Expected.Effective["passable_env_names"]; ok {
				if !cliVectorSetsPassableEnvNames(tc.Machine) {
					if securityPostureRevisions().EnvPassthrough == string(envfragment.S4Enforce) {
						want = []string{}
					} else {
						want = nil
					}
				}
				assertPostureJSONEqual(caseT, "effective passable_env_names", want, envStatus.PassableEnvNames)
			}

			operationCode, operationOut, operationErr := runSecurityPostureVectorOperation(caseT, source, tc)
			if (tc.Expected.Outcome == "refused" || tc.Expected.Outcome == "non-current") && operationCode != exitFail {
				caseT.Fatalf("operation exit = %d, want refusal (%d)\nstdout:\n%s\nstderr:\n%s", operationCode, exitFail, operationOut, operationErr)
			}
			if tc.Expected.Outcome != "refused" && tc.Expected.Outcome != "non-current" && operationCode != exitOK {
				caseT.Fatalf("operation exit = %d, want %q (%d)\nstdout:\n%s\nstderr:\n%s", operationCode, tc.Expected.Outcome, exitOK, operationOut, operationErr)
			}
			diagnosticOutput := operationOut + operationErr
			if source.cfg.SecurityPostureHardened() && strings.Contains(diagnosticOutput, config.DiagSecurityPosturePermissive) {
				caseT.Fatalf("hardened operation emitted a permissive warning:\n%s", diagnosticOutput)
			}
			if tc.Expected.Outcome == "non-current" {
				var status struct {
					NonCurrent      bool `json:"non_current"`
					MCPDeclarations []struct {
						Rows []json.RawMessage `json:"rows"`
					} `json:"mcp_declarations"`
				}
				if err := json.Unmarshal([]byte(operationOut), &status); err != nil || !status.NonCurrent {
					caseT.Fatalf("status-check did not report non-current: %v\n%s", err, operationOut)
				}
				if tc.Operation.MCPDeclarationsPresent {
					rows := 0
					for _, scope := range status.MCPDeclarations {
						rows += len(scope.Rows)
					}
					if rows == 0 {
						caseT.Fatal("status-check vector did not reproduce its MCP declarations")
					}
				}
			}
			for _, diagnostic := range tc.Expected.Diagnostics {
				if got := strings.Count(diagnosticOutput, diagnostic.Code); got != 1 {
					caseT.Fatalf("diagnostic %q occurred %d times, want once in production output:\n%s", diagnostic.Code, got, diagnosticOutput)
				}
				if diagnostic.Code == "registry_unreachable_during_install" {
					severityLine := diagnostic.Severity + ": GATE NOTICE " + diagnostic.Code
					if !strings.Contains(operationErr, severityLine) || strings.Contains(operationOut, diagnostic.Code) {
						caseT.Fatalf("registry gate diagnostic must be a %s on stderr only:\nstdout:\n%s\nstderr:\n%s", diagnostic.Severity, operationOut, operationErr)
					}
					for _, registryURL := range tc.Operation.UnreachableTrustedRegistries {
						if !strings.Contains(operationErr, registryURL) {
							caseT.Errorf("registry gate diagnostic omits trusted registry %q:\n%s", registryURL, operationErr)
						}
					}
					for _, artifact := range tc.Operation.ArtifactsWithoutEvidence {
						name, _, _ := strings.Cut(artifact, "@")
						if !strings.Contains(operationErr, name+"@") {
							caseT.Errorf("registry gate diagnostic omits artifact %q:\n%s", artifact, operationErr)
						}
					}
					if got, want := strings.Count(operationErr, "@"), len(tc.Operation.ArtifactsWithoutEvidence); got != want {
						caseT.Errorf("registry gate diagnostic names %d artifacts, want %d:\n%s", got, want, operationErr)
					}
				}
			}
			return conformancecoverage.Observation{}
		})
	if tally.Driven != 13 || tally.Bound != 4 || tally.KnownGap != 0 || tally.Skipped != 0 || tally.Total() != 17 {
		t.Fatalf("unexpected posture coverage: %+v", tally)
	}
}

// The hardened outage refusal is independent of the per-artifact registry
// policy override: §7.1 makes unreachability itself a blocking gate notice.
func TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	tc := cliSecurityPostureVector{
		Name: "hardened-registry-outage-with-advisory-policy",
		Machine: map[string]any{
			"schema_version":   2,
			"security_posture": "hardened",
			"allowed_sources":  []any{"https://github.com/example/skills"},
			"audit":            map[string]any{"registry_policy": "advisory"},
		},
	}
	tc.Operation.Kind = "install"
	tc.Operation.UnreachableTrustedRegistries = []string{"https://registry.example"}
	tc.Operation.ArtifactsWithoutEvidence = []string{"golden-skill@0123456789abcdef0123456789abcdef01234567"}
	source := securityPostureCLIFixture(t, tc)
	source.cfg.AuditRegistries = postureVectorUnreachableRegistries(tc.Operation.UnreachableTrustedRegistries)
	prepareRegistryOutageProject(t, source, tc.Operation.ArtifactsWithoutEvidence)
	var code int
	var stdout, stderr string
	withRegistryUnavailable(t, tc.Operation.UnreachableTrustedRegistries, func() {
		code, stdout, stderr = runProfile(t, source, "upgrade", "app")
	})
	if code != exitFail {
		t.Fatalf("hardened upgrade with advisory registry policy = %d, want refusal\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if got := strings.Count(stderr, "registry_unreachable_during_install"); got != 1 || !strings.Contains(stderr, "error: GATE NOTICE") {
		t.Fatalf("hardened gate diagnostic count/severity = %d\nstdout:\n%s\nstderr:\n%s", got, stdout, stderr)
	}
	if !strings.Contains(stderr, "https://registry.example") || !strings.Contains(stderr, "golden-skill@") || strings.Contains(stdout, "registry_unreachable_during_install") {
		t.Fatalf("hardened gate diagnostic does not identify the outage on stderr\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestPermissiveRegistryOutageWarnsWithArtifactIdentity(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	tc := cliSecurityPostureVector{
		Name: "permissive-registry-outage",
		Machine: map[string]any{
			"schema_version":   2,
			"security_posture": "permissive",
			"allowed_sources":  []any{"https://github.com/example/skills"},
			"audit":            map[string]any{"registry_policy": "advisory"},
		},
	}
	tc.Operation.Kind = "install"
	tc.Operation.UnreachableTrustedRegistries = []string{"https://registry.example"}
	tc.Operation.ArtifactsWithoutEvidence = []string{"golden-skill@0123456789abcdef0123456789abcdef01234567"}
	source := securityPostureCLIFixture(t, tc)
	source.cfg.AuditRegistries = postureVectorUnreachableRegistries(tc.Operation.UnreachableTrustedRegistries)
	prepareRegistryOutageProject(t, source, tc.Operation.ArtifactsWithoutEvidence)
	var code int
	var stdout, stderr string
	withRegistryUnavailable(t, tc.Operation.UnreachableTrustedRegistries, func() {
		code, stdout, stderr = runProfile(t, source, "install", "app")
	})
	if code != exitOK {
		t.Fatalf("permissive install with unreachable registry = %d, want proceed\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if got := strings.Count(stderr, "registry_unreachable_during_install"); got != 1 || !strings.Contains(stderr, "warning: GATE NOTICE") {
		t.Fatalf("permissive gate diagnostic count/severity = %d\nstdout:\n%s\nstderr:\n%s", got, stdout, stderr)
	}
	if !strings.Contains(stderr, "https://registry.example") || !strings.Contains(stderr, "golden-skill@") || strings.Contains(stdout, "registry_unreachable_during_install") {
		t.Fatalf("permissive gate diagnostic does not identify the outage on stderr\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if got := strings.Count(stderr, config.DiagSecurityPosturePermissive); got != 1 {
		t.Fatalf("permissive posture warning count = %d, want once\n%s", got, stderr)
	}
}

func cliVectorSetsPassableEnvNames(machine map[string]any) bool {
	environments, _ := machine["environments"].(map[string]any)
	_, configured := environments["passable_env_names"]
	return configured
}

// CLI status reports the revisions compiled into this manager. The config
// vector driver separately verifies the revisions supplied by each vector;
// here replace only those unrelated shipped rows so the production entry
// assertion can focus on the posture policy under test.
func cliExpectedSecurityPostureRows(rows []config.SecurityPostureRow, cfg *config.Config) []config.SecurityPostureRow {
	actual := securityPostureRevisions()
	shipped := map[string]any{
		"hook-trust":           actual.HookTrust,
		"env-passthrough":      actual.EnvPassthrough,
		"provider-trust-roots": actual.ProviderTrustRoots,
		"update-confirmation":  actual.UpdateConfirmation,
		"codex-seed":           actual.CodexSeed,
	}
	expected := append([]config.SecurityPostureRow(nil), rows...)
	for index := range expected {
		if expected[index].Gate == "source-allowlist" && cfg != nil {
			expected[index].Value = len(cfg.AllowedSources)
		}
		if expected[index].Source != "shipped" {
			continue
		}
		if value, ok := shipped[expected[index].Gate]; ok {
			expected[index].Value = value
		}
	}
	return expected
}

func decodeCLISecurityPostureRows(t testing.TB, rows []any) []config.SecurityPostureRow {
	t.Helper()
	payload, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []config.SecurityPostureRow
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func cliSecurityPostureVectorBound(name string) string {
	switch name {
	case "revision-A-default-permissive-status", "revision-A-permissive-warning-once-install", "locked-value-beats-explicit", "unreachable-registry-permissive-warns":
		return "historical revision-A implicit permissive default; this manager ships revision B (explicit permissive compatibility is tested separately)"
	default:
		return ""
	}
}

func securityPostureCLIFixture(t *testing.T, tc cliSecurityPostureVector) stubConfigSource {
	t.Helper()
	configSource, _ := profileHome(t)
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	project, home := legacyProject(t)

	machine := clonePostureVectorMap(tc.Machine)
	signingKey := ""
	machine["skills_root"] = filepath.Join(filepath.Dir(home), "skills")
	machine["projects"] = map[string]any{"app": map[string]any{
		"path": project, "agents": []any{"codex_cli"},
	}}
	if raw, ok := machine["allowed_sources"]; ok {
		machine["allowed_sources"] = canonicalPostureSources(raw)
	}
	if raw, ok := machine["environments"].(map[string]any); ok {
		environments := clonePostureVectorMap(raw)
		if allowlist, ok := environments["mcp_package_allowlist"]; ok {
			environments["mcp_package_allowlist"] = canonicalPostureSources(allowlist)
		}
		machine["environments"] = environments
	}
	if tc.Operation.Kind == "profile-install" && tc.Operation.MCPDeclarationsPresent {
		// The production resolver applies the broad source allowlist to every
		// git member in the closure, then applies the MCP allowlist as an
		// additional bound. Make this CLI fixture valid for both checks.
		allowed := postureVectorStrings(machine["allowed_sources"])
		seen := make(map[string]bool, len(allowed))
		for _, source := range allowed {
			seen[source] = true
		}
		environments, _ := machine["environments"].(map[string]any)
		if environments == nil {
			environments = map[string]any{}
		}
		for _, source := range postureVectorStrings(environments["mcp_package_allowlist"]) {
			if !seen[source] {
				allowed = append(allowed, source)
				seen[source] = true
			}
		}
		machine["allowed_sources"] = allowed
		if required, _ := tc.Expected.Effective["require_source_signers"].(bool); required {
			privateKey, publicKey := generateSSHKey(t)
			signingKey = privateKey
			sourceSigners := make(map[string]any, len(allowed))
			for _, source := range allowed {
				sourceSigners[source] = []any{map[string]any{
					"type": "ssh", "key": publicKey + " operator@example",
				}}
			}
			// Empty MCP policy still needs a real signed dependency so the
			// resolver reaches the posture gate after signature verification.
			if len(postureVectorStrings(environments["mcp_package_allowlist"])) == 0 && len(allowed) > 0 {
				sourceSigners[allowed[0]+"/mcp"] = sourceSigners[allowed[0]]
			}
			environments["source_signers"] = sourceSigners
			machine["environments"] = environments
		}
	}
	system := clonePostureVectorMap(tc.System)
	if _, ok := system["schema_version"]; !ok {
		system["schema_version"] = 2
	}
	machinePath := writePostureVectorFile(t, home, "config.json", machine)
	systemPath := writePostureVectorFile(t, t.TempDir(), "system.json", system)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(machinePath, nil)
	if err != nil {
		t.Fatalf("load vector config: %v", err)
	}
	configSource.path = machinePath
	configSource.cfg = cfg
	configSource.signingKey = signingKey
	return configSource
}

func postureVectorUnreachableRegistries(urls []string) []config.Registry {
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	publicKey := "ed25519:" + base64.StdEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	registries := make([]config.Registry, 0, len(urls))
	for index, registryURL := range urls {
		registries = append(registries, config.Registry{
			Name: fmt.Sprintf("trusted-%d", index+1), URL: registryURL,
			PublicKeys: []string{publicKey}, Enabled: true,
		})
	}
	return registries
}

// prepareRegistryOutageProject maps each pinned vector artifact name to a
// deterministic local skill repository. The vector's illustrative commit is
// replaced by the real fixture commit so the production resolver can acquire
// and install the package rather than testing a fabricated closure.
func prepareRegistryOutageProject(t *testing.T, source stubConfigSource, artifacts []string) {
	t.Helper()
	if len(artifacts) != 1 {
		t.Fatalf("pinned registry outage fixture currently expects one network artifact, got %d", len(artifacts))
	}
	if len(source.cfg.AllowedSources) == 0 {
		t.Fatal("registry outage fixture requires the vector's allowed source")
	}
	identity := source.cfg.AllowedSources[0]
	gitURL := identity
	if !strings.Contains(gitURL, "://") {
		gitURL = "https://" + gitURL
	}
	declarations := make([]map[string]any, 0, len(artifacts))
	repositories := make(map[string]string, len(artifacts))
	for _, artifact := range artifacts {
		name, _, ok := strings.Cut(artifact, "@")
		if !ok || name == "" {
			t.Fatalf("registry vector artifact %q has no name@commit form", artifact)
		}
		root := t.TempDir()
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("---\nname: "+name+"\n---\n# Registry outage fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit(t, root, "init", "-q", "-b", "main")
		runGit(t, root, "add", ".")
		runGit(t, root, "commit", "-qm", "registry outage fixture")
		runGit(t, root, "tag", "v1")
		declarations = append(declarations, map[string]any{"name": name, "git": gitURL, "tag": "v1"})
		repositories[gitURL] = root
	}
	serveGitRepos(t, repositories)
	project := source.cfg.Projects["app"].Path
	manifestValue := map[string]any{
		"schema_version": 1,
		"agents":         []string{"codex_cli"},
		"skills":         declarations,
	}
	payload, err := json.Marshal(manifestValue)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// withRegistryUnavailable stops the real HTTP registry client at its dial
// boundary, keeping the production entry path deterministic and offline.
func withRegistryUnavailable(t *testing.T, registryURLs []string, run func()) {
	t.Helper()
	hosts := make(map[string]bool, len(registryURLs))
	for _, registryURL := range registryURLs {
		parsed, err := url.Parse(registryURL)
		if err != nil {
			t.Fatalf("parse registry URL %q: %v", registryURL, err)
		}
		hosts[parsed.Hostname()] = true
	}
	previous := http.DefaultTransport
	baseTransport, ok := previous.(*http.Transport)
	if !ok {
		t.Fatalf("default HTTP transport has type %T, want *http.Transport", previous)
	}
	transport := baseTransport.Clone()
	dial := transport.DialContext
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(address)
		if err == nil && hosts[host] {
			return nil, errors.New("test fixture: trusted registry is unreachable")
		}
		return dial(ctx, network, address)
	}
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
	run()
}

func runSecurityPostureVectorOperation(t *testing.T, source stubConfigSource, tc cliSecurityPostureVector) (int, string, string) {
	t.Helper()
	switch tc.Operation.Kind {
	case "status":
		return runProfile(t, source, "status", "app", "--json")
	case "status-check":
		if tc.Operation.MCPDeclarationsPresent {
			preparePostureStatusMCPProfile(t, source)
		}
		return runProfile(t, source, "env", "status", "--check", "--json")
	case "install":
		if len(tc.Operation.ArtifactsWithoutEvidence) > 0 {
			source.cfg.AuditRegistries = postureVectorUnreachableRegistries(tc.Operation.UnreachableTrustedRegistries)
			prepareRegistryOutageProject(t, source, tc.Operation.ArtifactsWithoutEvidence)
			var code int
			var stdout, stderr string
			withRegistryUnavailable(t, tc.Operation.UnreachableTrustedRegistries, func() {
				code, stdout, stderr = runProfile(t, source, "install", "app")
			})
			return code, stdout, stderr
		}
		return runProfile(t, source, "install", "app")
	case "profile-install":
		installSource := securityPostureVectorPackage(t, source, tc)
		return runProfile(t, source, "profile", "install", installSource)
	default:
		t.Fatalf("unbounded vector has unsupported production operation %q", tc.Operation.Kind)
		return exitFail, "", ""
	}
}

// Install an MCP-bearing profile through the production CLI under explicit
// permissive policy, then restore the vector's B configuration. This models
// an existing profile encountering hardened contradictions after the flip.
func preparePostureStatusMCPProfile(t *testing.T, source stubConfigSource) {
	t.Helper()
	original, err := os.ReadFile(source.path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.WriteFile(source.path, original, 0o600); err != nil {
			t.Error(err)
		}
	}()
	fixture := cliSecurityPostureVector{Machine: map[string]any{
		"allowed_sources": []any{"https://example.com/posture-root", "https://example.com/posture-mcp"},
		"environments":    map[string]any{"mcp_package_allowlist": []any{"https://example.com/posture-mcp"}},
	}}
	fixture.Operation.MCPDeclarationsPresent = true
	setup := map[string]any{
		"schema_version": 2, "security_posture": "permissive",
		"skills_root": source.cfg.SkillsRoot, "projects": map[string]any{},
		"allowed_sources": []string{"example.com/posture-root", "example.com/posture-mcp"},
		"environments":    map[string]any{"mcp_package_allowlist": []string{"example.com/posture-mcp"}},
	}
	writePostureVectorFile(t, filepath.Dir(source.path), filepath.Base(source.path), setup)
	setupSource := reloadSource(t, source)
	pkg := securityPostureVectorPackage(t, setupSource, fixture)
	code, stdout, stderr := runProfile(t, setupSource, "profile", "install", pkg)
	if code != exitOK {
		t.Fatalf("seed MCP profile = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

func securityPostureVectorPackage(t *testing.T, source stubConfigSource, tc cliSecurityPostureVector) string {
	t.Helper()
	if !tc.Operation.MCPDeclarationsPresent {
		root := t.TempDir()
		writeContextPackage(t, root, "posture", "1.0.0", "posture test module\n")
		return root
	}
	requireGit(t)
	allowed := postureVectorStrings(tc.Machine["allowed_sources"])
	environment, _ := tc.Machine["environments"].(map[string]any)
	mcpAllowed := postureVectorStrings(environment["mcp_package_allowlist"])
	if len(allowed) == 0 {
		t.Fatalf("MCP vector fixture needs source and MCP allowlists: sources=%v mcp=%v", allowed, mcpAllowed)
	}
	rootIdentity, mcpIdentity := allowed[0], allowed[0]+"/mcp"
	if len(mcpAllowed) > 0 {
		mcpIdentity = mcpAllowed[0]
	}
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init", "-q", "-b", "main")
	commitPostureGitRepo(t, tool, "v1.0.0", source.signingKey)
	root := t.TempDir()
	manifest := `{"schema_version":1,"name":"posture-root","version":"1.0.0",` +
		`"requires":{"mcp":{"tool":{"git":"` + mcpIdentity + `","range":"*"}}}}` + "\n"
	writeGitRepoFile(t, root, "agent-context.json", manifest)
	runGitRepo(t, root, "init", "-q", "-b", "main")
	commitPostureGitRepo(t, root, "v1.0.0", source.signingKey)
	serveGitRepos(t, map[string]string{rootIdentity: root, mcpIdentity: tool})
	return rootIdentity
}

func commitPostureGitRepo(t *testing.T, repo, tag, signingKey string) {
	t.Helper()
	runGitRepo(t, repo, "add", ".")
	if signingKey == "" {
		runGitRepo(t, repo, "commit", "-m", tag)
		runGitRepo(t, repo, "tag", tag)
		return
	}
	runGitRepo(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+signingKey, "commit", "-S", "-m", tag)
	runGitRepo(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+signingKey, "tag", "-s", "-m", tag, tag)
}

func postureVectorStrings(raw any) []string {
	switch values := raw.(type) {
	case []string:
		return append([]string(nil), values...)
	case []any:
		result := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

// CLI policy consumes canonical host/path identities, while vector fixtures
// use transport URLs for convenient source examples. Keep the tested source
// identity the same while removing the transport used only for cloning.
func canonicalPostureSources(raw any) []string {
	sources := postureVectorStrings(raw)
	canonical := make([]string, len(sources))
	for index, source := range sources {
		canonical[index] = identity.Canonical(source)
		if canonical[index] == "" {
			canonical[index] = source
		}
	}
	return canonical
}

func clonePostureVectorMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input)+2)
	for key, value := range input {
		result[key] = value
	}
	return result
}

func writePostureVectorFile(t *testing.T, dir, name string, value any) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertPostureJSONEqual(t testing.TB, label string, want, got any) {
	t.Helper()
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal expected %s: %v", label, err)
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal actual %s: %v", label, err)
	}
	var wantValue, gotValue any
	if err := json.Unmarshal(wantJSON, &wantValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(gotJSON, &gotValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantValue, gotValue) {
		var pretty bytes.Buffer
		_ = json.Indent(&pretty, gotJSON, "", "  ")
		t.Fatalf("%s mismatch:\n got: %s\nwant: %s", label, pretty.String(), wantJSON)
	}
}
