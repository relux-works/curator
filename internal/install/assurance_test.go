package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

type assuredTestProvider struct {
	identity       closureexec.ProviderIdentity
	toolchain      Toolchain
	builder        Builder
	negotiations   int
	toolchainReads int
	builderReads   int
	negotiateErr   error
	mutate         func(*closureexec.ProviderCapabilityReceipt)
	receipt        *closureexec.ProviderCapabilityReceipt
}

func (provider *assuredTestProvider) LosslessObservation() bool { return true }
func (provider *assuredTestProvider) EnforceAndObserve(context.Context, closureexec.ExecutionRequest) (closureexec.Audit, error) {
	return closureexec.Audit{}, errors.New("one-shot provider path is not the build-session path")
}
func (provider *assuredTestProvider) Identity() closureexec.ProviderIdentity {
	return provider.identity
}
func (provider *assuredTestProvider) Negotiate(_ context.Context, nonce string) (closureexec.ProviderCapabilityReceipt, error) {
	provider.negotiations++
	if provider.negotiateErr != nil {
		return closureexec.ProviderCapabilityReceipt{}, provider.negotiateErr
	}
	if provider.receipt == nil {
		now := time.Now()
		provider.receipt = &closureexec.ProviderCapabilityReceipt{
			Provider: provider.identity, Health: "healthy", Nonce: nonce,
			ObservedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute),
			Capabilities: verifiedTestCapabilities(),
		}
	}
	receipt := *provider.receipt
	receipt.Capabilities = append([]closureexec.CapabilityEvidence(nil), provider.receipt.Capabilities...)
	if provider.mutate != nil {
		provider.mutate(&receipt)
	}
	return receipt, nil
}
func (provider *assuredTestProvider) BuildToolchain() Toolchain {
	provider.toolchainReads++
	return provider.toolchain
}
func (provider *assuredTestProvider) BuildBuilder() Builder {
	provider.builderReads++
	return provider.builder
}

type assuredTestToolchain struct{ session BuildSession }

func (toolchain assuredTestToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	return toolchain.session.Target(), toolchain.session.Toolchain(), nil
}
func (toolchain assuredTestToolchain) Establish(context.Context) (BuildSession, error) {
	return toolchain.session, nil
}

type assuredProbeFailureToolchain struct {
	err    error
	probes int
}

func (toolchain *assuredProbeFailureToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	toolchain.probes++
	return buildmeta.Target{}, buildmeta.Toolchain{}, toolchain.err
}

func (*assuredProbeFailureToolchain) Establish(context.Context) (BuildSession, error) {
	return nil, errors.New("establish is not part of this probe test")
}

type assuredTestSession struct {
	target    buildmeta.Target
	toolchain buildmeta.Toolchain
}

func (session *assuredTestSession) Target() buildmeta.Target       { return session.target }
func (session *assuredTestSession) Toolchain() buildmeta.Toolchain { return session.toolchain }
func (*assuredTestSession) VerifyToolchain(context.Context) error  { return nil }
func (*assuredTestSession) Release() error                         { return nil }

type assuredTestBuilder struct {
	binding closureexec.AssuranceBinding
	starts  int
	root    string
}

func (builder *assuredTestBuilder) Stage(_ context.Context, request StageRequest) (StagedArtifact, error) {
	builder.starts++
	path := filepath.Join(builder.root, "tool")
	payload := []byte("verified artifact")
	if err := os.WriteFile(path, payload, 0o700); err != nil {
		return StagedArtifact{}, err
	}
	digest := sha256.Sum256(payload)
	artifact := buildmeta.Artifact{Path: "bin/tool", SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(payload))}
	session := request.Session.(*assuredTestSession)
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: request.Source.Identity(), BuildRoot: request.BuildRoot, Command: request.Command,
		SourceDir: request.SourceDir, Target: session.target, Toolchain: session.toolchain, Policy: buildmeta.FixedPolicy(),
	}
	providerReceipt := closuregraph.ID("sha256:" + strings.Repeat("e", 64))
	receipt, err := closureexec.NewVerifiedBuildSessionReceipt(builder.binding, input, artifact, providerReceipt)
	return StagedArtifact{Path: path, Metadata: artifact, ExecutionReceipt: receipt}, err
}

type assuredInspectSpy struct {
	calls int
	seen  closureexec.AssuranceBinding
}

func (spy *assuredInspectSpy) Inspect(expect buildcache.Expectation) buildcache.Result {
	spy.calls++
	spy.seen = expect.Assurance
	return buildcache.Result{Status: buildcache.Miss}
}

func TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch(t *testing.T) {
	session := &assuredTestSession{
		target:    buildmeta.Target{GOOS: "darwin", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}},
		toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath, GoVersion: "go version go1.25.5 darwin/arm64", ContentSHA256: "sha256:" + strings.Repeat("c", 64)},
	}
	builder := &assuredTestBuilder{root: t.TempDir()}
	provider := verifiedTestProvider(assuredTestToolchain{session: session}, builder)
	authority, err := NewBuildAuthority(t.Context(), verifiedTestConfig(provider.identity), provider)
	if err != nil {
		t.Fatal(err)
	}
	builder.binding = authority.Binding()

	spy := &assuredInspectSpy{}
	cache := &assuredCache{authority: authority, inner: spy}
	cache.Inspect(buildcache.Expectation{Input: assuredTestInput(session)})
	if spy.calls != 1 || spy.seen.AssuranceMode != closureexec.AssuranceVerified || spy.seen.CapabilityReceiptID == nil {
		t.Fatalf("cache did not receive exact verified authority: calls=%d binding=%+v", spy.calls, spy.seen)
	}

	wrappedToolchain, err := authority.toolchain(nil)
	if err != nil {
		t.Fatal(err)
	}
	wrappedSession, err := wrappedToolchain.Establish(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	wrappedBuilder, err := authority.builder(nil)
	if err != nil {
		t.Fatal(err)
	}
	token := assuredTestSource(t)
	defer func() { _ = token.Close() }()
	artifact, err := wrappedBuilder.Stage(t.Context(), StageRequest{
		Session: wrappedSession, Source: token, CommandObject: map[string]any{"type": "build", "driver": "go-v1", "source_dir": "src/cmd/tool"},
		BuildRoot: "src", SourceDir: "src/cmd/tool", Command: "tool",
	})
	if err != nil {
		t.Fatal(err)
	}
	if builder.starts != 1 || artifact.ExecutionReceipt.Binding.CapabilityReceiptID == nil ||
		*artifact.ExecutionReceipt.Binding.CapabilityReceiptID != *authority.Binding().CapabilityReceiptID {
		t.Fatalf("dispatch did not retain negotiated authority: starts=%d receipt=%+v", builder.starts, artifact.ExecutionReceipt)
	}
}

func TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint(t *testing.T) {
	innerErr := errors.New("inner trusted-toolchain probe failed")

	t.Run("authority refusal never reaches the inner probe", func(t *testing.T) {
		inner := &assuredProbeFailureToolchain{err: innerErr}
		toolchain := &assuredToolchain{authority: &BuildAuthority{}, inner: inner}
		_, _, err := toolchain.Probe(t.Context())
		if err == nil || !strings.Contains(err.Error(), "execution_mode_unknown") {
			t.Fatalf("probe error = %v, want the authority refusal", err)
		}
		if inner.probes != 0 {
			t.Fatalf("authority refusal reached the inner probe %d times", inner.probes)
		}
	})

	t.Run("valid portable authority preserves the inner failure", func(t *testing.T) {
		inner := &assuredProbeFailureToolchain{err: innerErr}
		toolchain := &assuredToolchain{authority: NewPortableBuildAuthority(), inner: inner}
		_, _, err := toolchain.Probe(t.Context())
		if !errors.Is(err, innerErr) {
			t.Fatalf("probe error = %v, want the inner failure", err)
		}
		if inner.probes != 1 {
			t.Fatalf("valid authority reached the inner probe %d times, want once", inner.probes)
		}
	})
}

func TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*assuredTestProvider)
		code   string
	}{
		{name: "unhealthy", code: "verified_provider_unavailable", mutate: func(provider *assuredTestProvider) { provider.negotiateErr = errors.New("unhealthy") }},
		{name: "capability drift", code: "verified_capabilities_unsatisfied", mutate: func(provider *assuredTestProvider) {
			provider.mutate = func(receipt *closureexec.ProviderCapabilityReceipt) { receipt.Capabilities[0].Status = "advisory" }
		}},
		{name: "capability claim inflation", code: "verified_capabilities_unsatisfied", mutate: func(provider *assuredTestProvider) {
			provider.mutate = func(receipt *closureexec.ProviderCapabilityReceipt) {
				receipt.Capabilities = append(receipt.Capabilities, closureexec.CapabilityEvidence{CapabilityID: "uncontracted-lossless-observation-v1", Status: "established"})
			}
		}},
		{name: "identity drift", code: "verified_provider_identity_invalid", mutate: func(provider *assuredTestProvider) { provider.identity.ProviderID = "drifted.provider" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			session := &assuredTestSession{target: testTarget(), toolchain: testToolchain()}
			builder := &assuredTestBuilder{root: t.TempDir()}
			provider := verifiedTestProvider(assuredTestToolchain{session: session}, builder)
			config := verifiedTestConfig(provider.identity)
			testCase.mutate(provider)
			authority, err := NewBuildAuthority(t.Context(), config, provider)
			if err == nil || authority != nil || !strings.Contains(err.Error(), testCase.code) {
				t.Fatalf("preflight = authority=%v error=%v, want %s", authority, err, testCase.code)
			}
			if provider.toolchainReads != 0 || provider.builderReads != 0 || builder.starts != 0 {
				t.Fatalf("failed preflight reached build capability: toolchain=%d builder=%d starts=%d", provider.toolchainReads, provider.builderReads, builder.starts)
			}
		})
	}
}

func TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*assuredTestProvider)
		code   string
	}{
		{name: "unhealthy", code: "verified_provider_unavailable", mutate: func(provider *assuredTestProvider) { provider.negotiateErr = errors.New("unhealthy") }},
		{name: "capability drift", code: "verified_capabilities_unsatisfied", mutate: func(provider *assuredTestProvider) {
			provider.mutate = func(receipt *closureexec.ProviderCapabilityReceipt) { receipt.Capabilities[0].Status = "advisory" }
		}},
		{name: "capability claim inflation", code: "verified_capabilities_unsatisfied", mutate: func(provider *assuredTestProvider) {
			provider.mutate = func(receipt *closureexec.ProviderCapabilityReceipt) {
				receipt.Capabilities = append(receipt.Capabilities, closureexec.CapabilityEvidence{CapabilityID: "uncontracted-lossless-observation-v1", Status: "established"})
			}
		}},
		{name: "identity drift", code: "verified_provider_identity_invalid", mutate: func(provider *assuredTestProvider) { provider.identity.ProviderID = "drifted.provider" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name+" before cache", func(t *testing.T) {
			session := &assuredTestSession{target: testTarget(), toolchain: testToolchain()}
			builder := &assuredTestBuilder{root: t.TempDir()}
			provider := verifiedTestProvider(assuredTestToolchain{session: session}, builder)
			authority, err := NewBuildAuthority(t.Context(), verifiedTestConfig(provider.identity), provider)
			if err != nil {
				t.Fatal(err)
			}
			testCase.mutate(provider)
			spy := &assuredInspectSpy{}
			result := (&assuredCache{authority: authority, inner: spy}).Inspect(buildcache.Expectation{Input: assuredTestInput(session)})
			if result.Status != buildcache.Corrupt || !strings.Contains(result.Reason, testCase.code) || spy.calls != 0 || builder.starts != 0 {
				t.Fatalf("cache recheck = result=%+v cache_calls=%d starts=%d", result, spy.calls, builder.starts)
			}
		})
		t.Run(testCase.name+" before dispatch", func(t *testing.T) {
			session := &assuredTestSession{target: testTarget(), toolchain: testToolchain()}
			builder := &assuredTestBuilder{root: t.TempDir()}
			provider := verifiedTestProvider(assuredTestToolchain{session: session}, builder)
			authority, err := NewBuildAuthority(t.Context(), verifiedTestConfig(provider.identity), provider)
			if err != nil {
				t.Fatal(err)
			}
			testCase.mutate(provider)
			toolchain, err := authority.toolchain(nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = toolchain.Establish(t.Context()); err == nil || !strings.Contains(err.Error(), testCase.code) || builder.starts != 0 {
				t.Fatalf("dispatch recheck = error=%v starts=%d", err, builder.starts)
			}
		})
	}
}

func verifiedTestProvider(toolchain Toolchain, builder Builder) *assuredTestProvider {
	return &assuredTestProvider{
		identity: closureexec.ProviderIdentity{
			Contract: closureexec.VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0",
			BinarySHA256: closuregraph.ID("sha256:" + strings.Repeat("b", 64)), TrustEvidence: "fixture-signature",
		},
		toolchain: toolchain, builder: builder,
	}
}

func verifiedTestConfig(identity closureexec.ProviderIdentity) closureexec.AssuranceConfig {
	return closureexec.AssuranceConfig{
		Mode: closureexec.AssuranceVerified, ProviderID: identity.ProviderID, ProviderVersion: identity.Version,
		ProviderBinarySHA256: identity.BinarySHA256, ProviderTrustEvidence: identity.TrustEvidence,
	}
}

func verifiedTestCapabilities() []closureexec.CapabilityEvidence {
	return []closureexec.CapabilityEvidence{
		{CapabilityID: "total-network-denial-v1", Status: "established"},
		{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
		{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
		{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
		{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
		{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
	}
}

func assuredTestSource(t *testing.T) *buildsource.Token {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	token, err := buildsource.Validate(root)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func assuredTestInput(session *assuredTestSession) buildmeta.Input {
	return buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("a", 64)},
		BuildRoot:   "src", Command: "tool", SourceDir: "src/cmd/tool",
		Target: session.target, Toolchain: session.toolchain, Policy: buildmeta.FixedPolicy(),
	}
}
