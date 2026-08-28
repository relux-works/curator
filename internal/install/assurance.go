package install

import (
	"context"
	"fmt"
	"reflect"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
)

const buildSessionProviderCapability = "curator-build-session-provider-v1"

// VerifiedBuildSessionProvider is the additive production-build capability of
// a separately installed host provider. It composes with the existing
// authenticated Go manager/worker lifecycle without widening the preserved
// one-shot EnforceObserveProvider contract.
type VerifiedBuildSessionProvider interface {
	closureexec.VerifiedProvider
	BuildToolchain() Toolchain
	BuildBuilder() Builder
}

// BuildAuthority is one opaque, preflighted operation authority. The same
// pointer is required by cache inspection, compiler dispatch, receipt
// validation, and publication.
type BuildAuthority struct {
	binding   closureexec.AssuranceBinding
	provider  VerifiedBuildSessionProvider
	operation *closureexec.AssuredOperation
}

// NewBuildAuthority performs the mode/provider negotiation that must dominate
// every production cache lookup and process start.
func NewBuildAuthority(ctx context.Context, config closureexec.AssuranceConfig, provider VerifiedBuildSessionProvider) (*BuildAuthority, error) {
	if config.Mode == "" || config.Mode == closureexec.AssurancePortable {
		binding, err := closureexec.PreflightAssurance(ctx, config, provider)
		if err != nil {
			return nil, err
		}
		return &BuildAuthority{binding: binding}, nil
	}
	executor, err := closureexec.NewAssuredExecutor(config, nil, provider, "production-build-session-v1")
	if err != nil {
		return nil, err
	}
	operation, err := executor.Preflight(ctx)
	if err != nil {
		return nil, err
	}
	if provider.BuildToolchain() == nil || provider.BuildBuilder() == nil {
		return nil, fmt.Errorf("verified_provider_unavailable: provider lacks %s", buildSessionProviderCapability)
	}
	return &BuildAuthority{binding: operation.Binding(), provider: provider, operation: operation}, nil
}

// NewPortableBuildAuthority returns the exact manager-owned default. It starts
// no process and exists for direct library callers that omit CLI configuration.
func NewPortableBuildAuthority() *BuildAuthority {
	return &BuildAuthority{binding: closureexec.PortableAssuranceBinding()}
}

// Binding returns a detached copy of the authority's exact assurance record.
func (authority *BuildAuthority) Binding() closureexec.AssuranceBinding {
	if authority == nil {
		return closureexec.AssuranceBinding{}
	}
	binding := authority.binding
	binding.ActualCapabilities = append([]closureexec.CapabilityEvidence(nil), binding.ActualCapabilities...)
	return binding
}

func (authority *BuildAuthority) toolchain(fallback Toolchain) (Toolchain, error) {
	if authority == nil {
		return nil, fmt.Errorf("build assurance authority is absent")
	}
	selected := fallback
	if authority.binding.AssuranceMode == closureexec.AssuranceVerified {
		selected = authority.provider.BuildToolchain()
	}
	if selected == nil {
		return nil, fmt.Errorf("build assurance selected no toolchain session")
	}
	return &assuredToolchain{authority: authority, inner: selected}, nil
}

func (authority *BuildAuthority) revalidate(ctx context.Context) error {
	if authority == nil {
		return fmt.Errorf("build assurance authority is absent")
	}
	if authority.binding.AssuranceMode == closureexec.AssuranceVerified {
		return authority.operation.Revalidate(ctx)
	}
	return authority.binding.Validate()
}

func (authority *BuildAuthority) builder(fallback Builder) (Builder, error) {
	if authority == nil {
		return nil, fmt.Errorf("build assurance authority is absent")
	}
	selected := fallback
	if authority.binding.AssuranceMode == closureexec.AssuranceVerified {
		selected = authority.provider.BuildBuilder()
	}
	if selected == nil {
		return nil, fmt.Errorf("build assurance selected no builder")
	}
	return &assuredBuilder{authority: authority, inner: selected}, nil
}

type assuredToolchain struct {
	authority *BuildAuthority
	inner     Toolchain
}

func (toolchain *assuredToolchain) Probe(ctx context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	if err := toolchain.authority.revalidate(ctx); err != nil {
		return buildmeta.Target{}, buildmeta.Toolchain{}, err
	}
	return toolchain.inner.Probe(ctx)
}

func (toolchain *assuredToolchain) Establish(ctx context.Context) (BuildSession, error) {
	if err := toolchain.authority.revalidate(ctx); err != nil {
		return nil, err
	}
	session, err := toolchain.inner.Establish(ctx)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("build assurance provider returned no session")
	}
	return &assuredSession{BuildSession: session, authority: toolchain.authority}, nil
}

type assuredSession struct {
	BuildSession
	authority *BuildAuthority
}

func (session *assuredSession) VerifyToolchain(ctx context.Context) error {
	if err := session.authority.revalidate(ctx); err != nil {
		return err
	}
	return session.BuildSession.VerifyToolchain(ctx)
}

type assuredBuilder struct {
	authority *BuildAuthority
	inner     Builder
}

func (builder *assuredBuilder) Stage(ctx context.Context, request StageRequest) (StagedArtifact, error) {
	if err := builder.authority.revalidate(ctx); err != nil {
		return StagedArtifact{}, err
	}
	session, ok := request.Session.(*assuredSession)
	if !ok || session.authority != builder.authority {
		return StagedArtifact{}, fmt.Errorf("assurance_evidence_mismatch: compiler session authority differs from the operation")
	}
	request.Session = session.BuildSession
	artifact, err := builder.inner.Stage(ctx, request)
	if err != nil {
		return StagedArtifact{}, err
	}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: request.Source.Identity(), BuildRoot: request.BuildRoot,
		Command: request.Command, SourceDir: request.SourceDir,
		Target: session.Target(), Toolchain: session.Toolchain(), Policy: buildmeta.FixedPolicy(),
	}
	if err := artifact.ExecutionReceipt.ValidateFor(builder.authority.binding, input, artifact.Metadata); err != nil {
		return StagedArtifact{}, err
	}
	return artifact, nil
}

type assuredCache struct {
	authority *BuildAuthority
	inner     CacheInspector
}

func (cache *assuredCache) Inspect(expect buildcache.Expectation) buildcache.Result {
	if err := cache.authority.revalidate(context.Background()); err != nil {
		return buildcache.Result{Status: buildcache.Corrupt, Reason: err.Error()}
	}
	if !zeroAssurance(expect.Assurance) && !reflect.DeepEqual(expect.Assurance, cache.authority.binding) {
		return buildcache.Result{Status: buildcache.Corrupt, Reason: "assurance_evidence_mismatch: cache expectation authority differs from the operation"}
	}
	expect.Assurance = cache.authority.Binding()
	return cache.inner.Inspect(expect)
}

type assuredPublisher struct {
	*assuredCache
	inner CachePublisher
}

func (publisher *assuredPublisher) Publish(publication buildcache.Publication, lock buildcache.HomeLock) (buildcache.PublicationResult, error) {
	if err := publisher.authority.revalidate(context.Background()); err != nil {
		return buildcache.PublicationResult{}, err
	}
	if !zeroAssurance(publication.Assurance) && !reflect.DeepEqual(publication.Assurance, publisher.authority.binding) {
		return buildcache.PublicationResult{}, fmt.Errorf("assurance_evidence_mismatch: cache publication authority differs from the operation")
	}
	publication.Assurance = publisher.authority.Binding()
	return publisher.inner.Publish(publication, lock)
}

func (publisher *assuredPublisher) Revert(key buildmeta.CacheKey, published buildcache.PublicationResult, lock buildcache.HomeLock) error {
	return publisher.inner.Revert(key, published, lock)
}

func bindCommitPublisher(authority *BuildAuthority, commit CommitDeps) (CommitDeps, error) {
	if authority == nil || commit.Publisher == nil {
		return CommitDeps{}, fmt.Errorf("build assurance cannot bind an absent commit publisher")
	}
	cache := &assuredCache{authority: authority, inner: commit.Publisher}
	commit.Publisher = &assuredPublisher{assuredCache: cache, inner: commit.Publisher}
	return commit, nil
}

func zeroAssurance(binding closureexec.AssuranceBinding) bool {
	return binding.AssuranceMode == "" && binding.PolicyID == "" && binding.ExecutionPolicyID == "" &&
		binding.ProviderContract == nil && binding.Provider == nil && binding.CapabilityReceiptID == nil && len(binding.ActualCapabilities) == 0
}

func assuredBuildKey(input buildmeta.Input, authority *BuildAuthority) (buildmeta.CacheKey, error) {
	if authority == nil {
		return "", fmt.Errorf("build assurance authority is absent")
	}
	id, err := (closureexec.AssuredBuildCacheInput{BuildInput: input, Binding: authority.Binding()}).ID()
	return buildmeta.CacheKey(id), err
}
