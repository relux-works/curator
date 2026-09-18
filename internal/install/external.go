package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/staging"
)

// ExternalDeps are operator-owned external-repository boundaries. Package
// data cannot select the Git executable, credential broker, signer, audit
// policy, protected store, or acquisition implementation.
type ExternalDeps struct {
	GitTool       buildrepo.GitTool
	Limits        buildrepo.Limits
	StoreRoot     string
	SigningPolicy string
	Acquire       func(context.Context, ExternalSource) (*buildrepo.Snapshot, error)
	Audit         func(context.Context, buildrepo.AuditSubject) error
	AuditWarnings func(context.Context, buildrepo.AuditSubject) ([]string, error)
	// BuildSSH is the operator's SSH credential selection for this run: the
	// explicit flags/env values and the configured build_ssh scopes they take
	// precedence over.
	BuildSSH BuildSSHSelection
	// BuildHTTPS is the operator's captured HTTPS selection for this run.
	BuildHTTPS BuildHTTPSSelection
	// DraftTransportResolution opts this run's external-repository lane into
	// bounded transport resolution (repository-transport-v1 §2): with a
	// present machine policy, acquisition runs the resolved executor over the
	// policy's endpoint plan. The zero value keeps the legacy lane for every
	// fetch, without consulting any policy file.
	DraftTransportResolution bool
	// DraftPolicyPath names the machine source-policy.json consulted when
	// DraftTransportResolution is set. "" resolves to the default path beside
	// the manager configuration.
	DraftPolicyPath string
	// DraftProvidersPath names the machine source-providers.json consulted
	// when DraftTransportResolution is set and a planned attempt names a
	// policy provider. "" resolves to the default path beside the manager
	// configuration. An absent file configures no providers; a
	// present-but-invalid file fails the fetch before any traffic.
	DraftProvidersPath string
	// DraftProviderReader reads provider HTTPS secrets through the
	// operator's trusted broker. nil resolves no named HTTPS material:
	// every non-anonymous named HTTPS provider is unavailable.
	DraftProviderReader buildrepo.ProviderSecretReader
	// DraftTransportTrace receives the resolved lane's sanitized
	// per-endpoint provenance records (canonical identity, listed URL,
	// resolved host and port, alias and mirror_of when used). It is
	// machine-private: the caller stores records in operation
	// diagnostics, separately from portable identity. nil records
	// nothing.
	DraftTransportTrace buildrepo.TransportTrace
}

// ExternalSource is the exact declared/effective source passed to an injected
// acquisition boundary. It contains no credential or signing material.
// GitURL is the declared fetchable repository URL; Declared.Repository is the
// repository's configured name and never a fetch endpoint.
type ExternalSource struct {
	Skill, Repository string
	GitURL            string
	Declared          buildrepo.DeclaredState
	Effective         buildrepo.EffectiveState
	Substitution      *devsub.BuildRepositorySubstitution
}

type plannedExternal struct {
	node       *closure.Node
	command    skillspec.Command
	repository skillspec.BuildRepository
	declared   buildrepo.DeclaredState
	effective  buildrepo.EffectiveState
	sub        *devsub.BuildRepositorySubstitution
	// pkg is the frozen package identity of the declaring draft member, or
	// nil on the legacy lane; it selects the receipt-3 wrapper and namespace.
	pkg    *buildmeta.Package
	result buildrepo.PipelineResult
}

type externalPlan struct {
	scope, projectIdentity string
	rows                   []plannedExternal
	deps                   ExternalDeps
	toolchain              buildrepo.ToolchainIdentity
	targetIdentity         buildmeta.Target
	toolchainInput         buildmeta.Toolchain
	authority              *BuildAuthority
	// credentials holds the operator SSH selection of every repository that
	// is actually fetched over SSH. A repository absent from it needs none.
	credentials map[buildSSHKey]buildrepo.OperatorSSHCredentials
	// httpsCredentials carries resolved per-repository material for the
	// manager credential broker. The fetch wiring consumes it in a later layer.
	httpsCredentials map[buildHTTPSKey]BuildHTTPSCredentials
	// messages report where each selection came from. Populated on a dry run,
	// where the provenance is the whole point of the report.
	messages []string
}

// credentialsFor returns the operator SSH selection of one planned repository.
func (p externalPlan) credentialsFor(row plannedExternal) buildrepo.OperatorSSHCredentials {
	return p.credentials[buildSSHKey{skill: row.node.Name, command: row.command.Name}]
}

func (p externalPlan) httpsCredentialsFor(row plannedExternal) BuildHTTPSCredentials {
	return p.httpsCredentials[buildHTTPSKeyFor(row)]
}

type stagedExternal struct {
	root    string
	entries map[string]map[string]externalEntry
}

type externalEntry struct {
	planned      plannedExternal
	result       buildrepo.PipelineResult
	artifactPath string
	receiptPath  string
	record       marker.Build
	existing     bool
}

func (deps ExternalDeps) resolved(home string) ExternalDeps {
	if deps.StoreRoot == "" {
		deps.StoreRoot = filepath.Join(home, "external-build-cache")
	}
	if deps.Limits == (buildrepo.Limits{}) {
		deps.Limits = buildrepo.DefaultLimits()
	}
	return deps
}

func planExternalBuilds(ctx context.Context, scope, projectIdentity, home string, nodes []*closure.Node, substitutions map[string]map[string]devsub.BuildRepositorySubstitution, packages map[string]*buildmeta.Package, toolchain Toolchain, deps ExternalDeps, authority *BuildAuthority, dryRun bool) (externalPlan, error) {
	if authority == nil {
		return externalPlan{}, fmt.Errorf("build assurance authority is absent")
	}
	plan := externalPlan{scope: scope, projectIdentity: projectIdentity, deps: deps.resolved(home), authority: authority}
	items := externalCommands(nodes)
	if len(items) == 0 {
		return plan, nil
	}
	target, tool, err := toolchain.Probe(ctx)
	if err != nil {
		return plan, err
	}
	plan.toolchain = externalToolchain(target, tool)
	plan.targetIdentity, plan.toolchainInput = target, tool
	store := &buildrepo.DiskProtectedStore{Root: plan.deps.StoreRoot}
	for _, item := range items {
		repository := item.node.Spec.BuildRepositories[item.command.Repository]
		declared := declaredRepository(repository)
		sub := repositorySubstitution(substitutions, item.node.Name, repository.Name)
		effective, err := effectiveRepository(projectIdentity, repository, sub)
		if err != nil {
			return plan, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
		}
		plan.rows = append(plan.rows, plannedExternal{node: item.node, command: item.command, repository: repository, declared: declared, effective: effective, sub: sub, pkg: packages[item.node.Name]})
	}
	// Credentials are selected for the whole run before the first repository is
	// reached, so a closure holding one unselected private repository fails
	// closed naming every one of them instead of part way through the network.
	credentials, provenance, err := resolveBuildSSH(plan.deps.BuildSSH, plan.rows)
	if err != nil {
		return plan, err
	}
	plan.credentials = credentials
	httpsCredentials, httpsProvenance, err := resolveBuildHTTPS(ctx, plan.deps.BuildHTTPS, plan.rows)
	if err != nil {
		return plan, err
	}
	plan.httpsCredentials = httpsCredentials
	if dryRun {
		plan.messages = append(provenance, httpsProvenance...)
	}
	for index, row := range plan.rows {
		source := ExternalSource{Skill: row.node.Name, Repository: row.repository.Name, GitURL: row.repository.Git, Declared: row.declared, Effective: row.effective, Substitution: row.sub}
		request, err := externalPipelineRequest(plan.deps, source, plan.credentialsFor(row), plan.httpsCredentialsFor(row), row.command, row.pkg, store, identityOnlyExternalGo{identity: plan.toolchain, target: plan.targetIdentity, toolchain: plan.toolchainInput}, buildrepo.OperationDryRun, plan.authority)
		if err != nil {
			return plan, err
		}
		result, err := buildrepo.RunPipeline(ctx, request)
		if err != nil {
			return plan, fmt.Errorf("%s.%s: %w", row.node.Name, row.command.Name, err)
		}
		plan.rows[index].result = result
	}
	return plan, nil
}

func stageExternalBuilds(ctx context.Context, plan externalPlan, toolchain Toolchain, builder Builder, private *privateRoot) (stagedExternal, error) {
	staged := stagedExternal{entries: map[string]map[string]externalEntry{}}
	if len(plan.rows) == 0 {
		return staged, nil
	}
	privateDir, err := private.dir("external-cache-")
	if err != nil {
		return staged, err
	}
	// The staging store root is created by the protected store itself, not by
	// the temporary-directory helper: only the store's private creation yields
	// a directory the store later proves as its own (a non-inheritable,
	// owner-private DACL on Windows; 0700 on unix).
	root := filepath.Join(privateDir, "store")
	staged.root = root
	session, err := toolchain.Establish(ctx)
	if err != nil {
		return staged, err
	}
	defer func() { _ = session.Release() }()
	adapter := externalGoAdapter{session: session, builder: builder}
	store := &buildrepo.DiskProtectedStore{Root: root}
	if err := store.PrepareNamespaces(); err != nil {
		return stagedExternal{}, err
	}
	for _, row := range plan.rows {
		if row.result.State == "cache-hit" {
			if err := plan.authority.revalidate(ctx); err != nil {
				return stagedExternal{}, err
			}
			entryRoot := filepath.Join(plan.deps.StoreRoot, buildrepo.ArtifactsDir(row.result.ReceiptSchemaVersion), strings.TrimPrefix(row.result.CacheKey, "sha256:"))
			artifactPath := filepath.Join(entryRoot, "artifact")
			artifact, readErr := os.ReadFile(artifactPath) // #nosec G304 -- manager-derived protected cache path from a validated cache key.
			if readErr != nil {
				return stagedExternal{}, readErr
			}
			artifactHash := sha256.Sum256(artifact)
			receiptHash := sha256.Sum256(row.result.Receipt)
			record := externalMarkerBuild(row, row.result, adapter.Identity().GOOS, "sha256:"+hex.EncodeToString(receiptHash[:]), "sha256:"+hex.EncodeToString(artifactHash[:]))
			if staged.entries[row.node.Name] == nil {
				staged.entries[row.node.Name] = map[string]externalEntry{}
			}
			staged.entries[row.node.Name][row.command.Name] = externalEntry{planned: row, result: row.result, artifactPath: artifactPath, receiptPath: filepath.Join(entryRoot, "receipt.json"), record: record, existing: true}
			continue
		}
		// A verified final hit is copied into operation-private staging so the
		// transaction and shim path are identical for hits and misses.
		request, err := externalPipelineRequest(plan.deps, ExternalSource{Skill: row.node.Name, Repository: row.repository.Name, GitURL: row.repository.Git, Declared: row.declared, Effective: row.effective, Substitution: row.sub}, plan.credentialsFor(row), plan.httpsCredentialsFor(row), row.command, row.pkg, store, adapter, buildrepo.OperationInstall, plan.authority)
		if err != nil {
			return stagedExternal{}, err
		}
		result, err := buildrepo.RunPipeline(ctx, request)
		if err != nil {
			return stagedExternal{}, fmt.Errorf("%s.%s: %w", row.node.Name, row.command.Name, err)
		}
		entryRoot := filepath.Join(root, buildrepo.ArtifactsDir(result.ReceiptSchemaVersion), strings.TrimPrefix(result.CacheKey, "sha256:"))
		artifactPath := filepath.Join(entryRoot, "artifact")
		receiptPath := filepath.Join(entryRoot, "receipt.json")
		artifact, err := os.ReadFile(artifactPath) // #nosec G304 -- pipeline result supplies the manager-staged artifact path.
		if err != nil {
			return stagedExternal{}, err
		}
		receipt, err := os.ReadFile(receiptPath) // #nosec G304 -- pipeline result supplies the manager-staged canonical receipt path.
		if err != nil {
			return stagedExternal{}, err
		}
		artifactHash := sha256.Sum256(artifact)
		receiptHash := sha256.Sum256(receipt)
		record := externalMarkerBuild(row, result, adapter.Identity().GOOS, "sha256:"+hex.EncodeToString(receiptHash[:]), "sha256:"+hex.EncodeToString(artifactHash[:]))
		if staged.entries[row.node.Name] == nil {
			staged.entries[row.node.Name] = map[string]externalEntry{}
		}
		staged.entries[row.node.Name][row.command.Name] = externalEntry{planned: row, result: result, artifactPath: artifactPath, receiptPath: receiptPath, record: record}
	}
	if err := session.VerifyToolchain(ctx); err != nil {
		return stagedExternal{}, err
	}
	if err := staged.prepareFinalNamespaces(plan.deps.StoreRoot); err != nil {
		return stagedExternal{}, err
	}
	return staged, nil
}

// prepareFinalNamespaces creates, through the protected store's private
// creation, every final-root parent the transaction will rename staged
// entries into: the root, the snapshot namespace and the artifact namespace of
// each receipt schema version being published. The commit's generic directory
// scaffolding must never create a protected parent, because a parent it
// creates is not provably private and every later lookup refuses it. Nothing
// is prepared when no entry is published.
func (staged stagedExternal) prepareFinalNamespaces(finalRoot string) error {
	versions := map[int]bool{}
	for _, commands := range staged.entries {
		for _, entry := range commands {
			if !entry.existing {
				versions[entry.result.ReceiptSchemaVersion] = true
			}
		}
	}
	if len(versions) == 0 {
		return nil
	}
	list := make([]int, 0, len(versions))
	for version := range versions {
		list = append(list, version)
	}
	sort.Ints(list)
	return (&buildrepo.DiskProtectedStore{Root: finalRoot}).PrepareNamespaces(list...)
}

// externalAdoption is one protected entry the transaction publishes into the
// final store and the commit must adopt through the store before any lookup.
type externalAdoption struct {
	root                 string
	receiptSchemaVersion int
	key                  string
	snapshot             bool
}

func (adoption externalAdoption) adopt() error {
	store := &buildrepo.DiskProtectedStore{Root: adoption.root}
	if adoption.snapshot {
		return store.AdoptSnapshot(adoption.key)
	}
	return store.AdoptArtifact(adoption.receiptSchemaVersion, adoption.key)
}

// adoptions lists, in commit order, every entry transactionPlan publishes into
// the final root: the artifact entry of each newly built command and each
// snapshot published with it. The transaction recreates these objects as fresh
// files (a mode-only copy, then a rename), which keeps the unix proof but not
// the Windows one; runCommit adopts each entry through the protected store
// immediately after the journal commit so both arms are exact hits afterwards.
func (staged stagedExternal) adoptions(finalRoot string) []externalAdoption {
	var list []externalAdoption
	snapshots := map[string]bool{}
	for _, commands := range staged.entries {
		for _, entry := range commands {
			if entry.existing {
				continue
			}
			list = append(list, externalAdoption{root: finalRoot, receiptSchemaVersion: entry.result.ReceiptSchemaVersion, key: entry.result.CacheKey})
			if entry.result.SnapshotKey != "" && !snapshots[entry.result.SnapshotKey] {
				snapshots[entry.result.SnapshotKey] = true
				list = append(list, externalAdoption{root: finalRoot, key: entry.result.SnapshotKey, snapshot: true})
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].snapshot != list[j].snapshot {
			return !list[i].snapshot
		}
		return list[i].key < list[j].key
	})
	return list
}

func (staged stagedExternal) transactionPlan(finalRoot string) staging.Plan {
	var plan staging.Plan
	snapshots := map[string]bool{}
	for skill, commands := range staged.entries {
		for command, entry := range commands {
			if entry.existing {
				continue
			}
			key := strings.TrimPrefix(entry.result.CacheKey, "sha256:")
			plan.Replace("05-external-cache", skill+"/"+command, filepath.Join(finalRoot, buildrepo.ArtifactsDir(entry.result.ReceiptSchemaVersion), key), filepath.Dir(entry.artifactPath))
			if entry.result.SnapshotKey != "" {
				snapshotKey := strings.TrimPrefix(entry.result.SnapshotKey, "sha256:")
				if !snapshots[snapshotKey] {
					plan.Replace("05-external-snapshot", snapshotKey, filepath.Join(finalRoot, "snapshots", snapshotKey), filepath.Join(staged.root, "snapshots", snapshotKey))
					snapshots[snapshotKey] = true
				}
			}
		}
	}
	return plan
}

func externalPipelineRequest(deps ExternalDeps, source ExternalSource, credentials buildrepo.OperatorSSHCredentials, httpsCredentials BuildHTTPSCredentials, command skillspec.Command, pkg *buildmeta.Package, store buildrepo.ProtectedStore, goSession buildrepo.GoSession, operation buildrepo.Operation, authority *BuildAuthority) (buildrepo.PipelineRequest, error) {
	if deps.Audit == nil && deps.AuditWarnings == nil {
		return buildrepo.PipelineRequest{}, fmt.Errorf("build_repository_audit_blocked: independent external repository audit is not configured")
	}
	// The tool is bound per repository so each fetch offers exactly the
	// credentials selected for the identity it is about to reach, and nothing
	// selected for a different host in the same closure.
	tool := externalGitTool(deps.GitTool, source, credentials, httpsCredentials)
	acquire := deps.Acquire
	if acquire == nil {
		acquire = func(ctx context.Context, selected ExternalSource) (*buildrepo.Snapshot, error) {
			if selected.Substitution != nil && selected.Substitution.Path != "" {
				return buildrepo.AdmitLocal(ctx, buildrepo.LocalRequest{Path: selected.Substitution.Path, Tool: tool, Limits: deps.Limits})
			}
			git := selected.GitURL
			transport := selected.Declared.Transport
			identity := selected.Declared.Identity
			commit := selected.Declared.Commit
			tag := selected.Declared.Tag
			refKind, refValue := "", ""
			if selected.Substitution != nil {
				git, transport, identity = selected.Substitution.Git, selected.Substitution.Transport, selected.Substitution.Identity
				commit, tag = selected.Effective.Commit, ""
				refKind, refValue = selected.Substitution.RefKind, selected.Substitution.RefValue
			}
			return deps.acquireDraftNetwork(ctx, tool, git, transport, identity, buildrepo.LockedCommit{ObjectFormat: selected.Effective.ObjectFormat, Hex: commit}, tag, refKind, refValue)
		}
	}
	if authority == nil {
		return buildrepo.PipelineRequest{}, fmt.Errorf("build assurance authority is absent")
	}
	return buildrepo.PipelineRequest{Operation: operation, Command: command.Name, Target: command.Target, Declared: source.Declared, Effective: source.Effective, Acquire: func(ctx context.Context) (*buildrepo.Snapshot, error) { return acquire(ctx, source) }, Audit: deps.Audit, AuditWarnings: deps.AuditWarnings, Store: store, Go: goSession, SigningPolicy: deps.SigningPolicy, Assurance: authority.Binding(), AssuranceCheck: authority.revalidate, Package: pkg}, nil
}

func externalGitTool(tool buildrepo.GitTool, source ExternalSource, sshCredentials buildrepo.OperatorSSHCredentials, httpsCredentials BuildHTTPSCredentials) buildrepo.GitTool {
	tool.SSHCredentials = sshCredentials
	if httpsCredentials.Selected() {
		tool.HTTPSCredentials = buildrepo.NewHTTPSCredentials(
			buildHTTPSHost(source.Effective.Identity), httpsCredentials.Username, httpsCredentials.Secret())
	}
	return tool
}

type identityOnlyExternalGo struct {
	identity  buildrepo.ToolchainIdentity
	target    buildmeta.Target
	toolchain buildmeta.Toolchain
}

func (g identityOnlyExternalGo) Identity() buildrepo.ToolchainIdentity { return g.identity }
func (g identityOnlyExternalGo) BuildInput(request buildrepo.CompileRequest) (buildmeta.Input, error) {
	return externalBuildInput(request, g.target, g.toolchain)
}
func (identityOnlyExternalGo) Compile(context.Context, buildrepo.CompileRequest) (buildrepo.CompileResult, error) {
	return buildrepo.CompileResult{}, fmt.Errorf("read-only external plan attempted compilation")
}

type externalGoAdapter struct {
	session BuildSession
	builder Builder
}

func (g externalGoAdapter) Identity() buildrepo.ToolchainIdentity {
	return externalToolchain(g.session.Target(), g.session.Toolchain())
}
func (g externalGoAdapter) BuildInput(request buildrepo.CompileRequest) (buildmeta.Input, error) {
	return externalBuildInput(request, g.session.Target(), g.session.Toolchain())
}
func (g externalGoAdapter) Compile(ctx context.Context, request buildrepo.CompileRequest) (buildrepo.CompileResult, error) {
	token, err := buildsource.Validate(request.Root)
	if err != nil {
		return buildrepo.CompileResult{}, err
	}
	defer func() { _ = token.Close() }()
	// Thread the exact receipt-3 wrapper digest so the staged execution
	// receipt binds sha256(CCJ-1(receipt.input)), never the compiler view.
	result, err := g.builder.Stage(ctx, StageRequest{Session: g.session, Source: token, CommandObject: map[string]any{"type": "build", "driver": "go-v1", "source_dir": request.SourceDir}, BuildRoot: "build", SourceDir: request.SourceDir, Command: request.Command, Package: request.Package, ExpectedBuildInputSHA256: request.ExpectedDigest})
	if err != nil {
		return buildrepo.CompileResult{}, err
	}
	artifact, err := os.ReadFile(result.Path)
	if err != nil {
		return buildrepo.CompileResult{}, err
	}
	return buildrepo.CompileResult{Artifact: artifact, ExecutionReceipt: result.ExecutionReceipt}, nil
}

func externalBuildInput(request buildrepo.CompileRequest, target buildmeta.Target, toolchain buildmeta.Toolchain) (buildmeta.Input, error) {
	token, err := buildsource.Validate(request.Root)
	if err != nil {
		return buildmeta.Input{}, err
	}
	defer func() { _ = token.Close() }()
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Package: request.Package, Driver: buildmeta.DriverGoV1,
		BuildSource: token.Identity(), BuildRoot: "build", Command: request.Command, SourceDir: request.SourceDir,
		Target: target, Toolchain: toolchain, Policy: buildmeta.FixedPolicy(),
	}
	if err := input.Validate(); err != nil {
		return buildmeta.Input{}, err
	}
	return input, nil
}

func externalToolchain(target buildmeta.Target, tool buildmeta.Toolchain) buildrepo.ToolchainIdentity {
	return buildrepo.ToolchainIdentity{ContentSHA256: tool.ContentSHA256, GoVersion: tool.GoVersion, GoRelpath: tool.GoRelpath, GOOS: target.GOOS, GOARCH: target.GOARCH, Tuning: target.Tuning}
}

func declaredRepository(repository skillspec.BuildRepository) buildrepo.DeclaredState {
	return buildrepo.DeclaredState{Repository: repository.Name, Identity: repository.Identity, Transport: repository.Transport, ObjectFormat: repository.LockedCommit.ObjectFormat, Commit: repository.LockedCommit.Hex, Tag: repository.Tag}
}

func effectiveRepository(projectIdentity string, repository skillspec.BuildRepository, sub *devsub.BuildRepositorySubstitution) (buildrepo.EffectiveState, error) {
	if sub == nil {
		return buildrepo.EffectiveState{IdentityKind: "network-git", Identity: repository.Identity, Transport: repository.Transport, ObjectFormat: repository.LockedCommit.ObjectFormat, Commit: repository.LockedCommit.Hex}, nil
	}
	if sub.Path != "" {
		identity, err := buildrepo.LocalIdentity(projectIdentity, sub.Selector)
		if err != nil {
			return buildrepo.EffectiveState{}, err
		}
		return buildrepo.EffectiveState{IdentityKind: "operator-local-git", Identity: identity, ObjectFormat: repository.LockedCommit.ObjectFormat, Commit: repository.LockedCommit.Hex, Substituted: true, Substitution: &buildrepo.SubstitutionState{Type: "local-path"}}, nil
	}
	commit := repository.LockedCommit.Hex
	if sub.RefKind == "revision" {
		commit = sub.RefValue
		if len(commit) == 64 {
			repository.LockedCommit.ObjectFormat = "sha256"
		} else {
			repository.LockedCommit.ObjectFormat = "sha1"
		}
	}
	return buildrepo.EffectiveState{IdentityKind: "network-git", Identity: sub.Identity, Transport: sub.Transport, ObjectFormat: repository.LockedCommit.ObjectFormat, Commit: commit, Substituted: true, Substitution: &buildrepo.SubstitutionState{Type: "network-git", RefKind: sub.RefKind, RefValue: sub.RefValue}}, nil
}

func externalMarkerBuild(row plannedExternal, result buildrepo.PipelineResult, goos, receiptHash, artifactHash string) marker.Build {
	effective := result.Subject.Effective
	record := marker.Build{Driver: "go-repository-v1", ReceiptSchemaVersion: result.ReceiptSchemaVersion, ExecutionPolicy: buildmeta.ExecutionPolicy, Repository: row.repository.Name, DeclaredIdentity: &marker.RepositoryIdentity{Kind: "network-git", Value: row.declared.Identity}, DeclaredLockedCommit: &marker.RepositoryCommit{ObjectFormat: row.declared.ObjectFormat, Hex: row.declared.Commit}, DeclaredTag: row.declared.Tag, EffectiveIdentity: &marker.RepositoryIdentity{Kind: row.effective.IdentityKind, Value: row.effective.Identity}, ObjectFormat: row.effective.ObjectFormat, Commit: row.effective.Commit, Substituted: row.effective.Substituted, BuildSource: &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: result.BuildSource}, DescriptorTarget: row.command.Target, CacheKey: buildmeta.CacheKey(result.CacheKey), ReceiptSHA256: buildmeta.ReceiptHash(receiptHash), ArtifactSHA256: artifactHash}
	record.EffectiveIdentity = &marker.RepositoryIdentity{Kind: effective.IdentityKind, Value: effective.Identity}
	record.ObjectFormat, record.Commit, record.Substituted = effective.ObjectFormat, effective.Commit, effective.Substituted
	record.ArtifactPath, _ = buildmeta.ArtifactPath(row.command.Name, goos)
	if effective.Substitution != nil {
		record.Substitution = &marker.RepositorySubstitute{Type: effective.Substitution.Type}
		if effective.Substitution.RefKind != "" {
			record.Substitution.Ref = &marker.RepositoryRef{Kind: effective.Substitution.RefKind, Value: effective.Substitution.RefValue}
		}
	}
	return record
}

func externalCommands(nodes []*closure.Node) []plannedCommand {
	var rows []plannedCommand
	for _, node := range nodes {
		active := node.ActiveCommands()
		for _, name := range node.ActiveCommandNames() {
			command := node.Spec.Commands[name]
			if command.Type == "build" && command.Driver == "go-repository-v1" && active[name] {
				rows = append(rows, plannedCommand{node: node, command: command})
			}
		}
	}
	return rows
}

func repositorySubstitution(values map[string]map[string]devsub.BuildRepositorySubstitution, skill, repository string) *devsub.BuildRepositorySubstitution {
	value, ok := values[skill][repository]
	if !ok {
		return nil
	}
	return &value
}

func sortedBuildSubstitutionSkills(values map[string]map[string]devsub.BuildRepositorySubstitution) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func sortedBuildSubstitutionRepositories(values map[string]devsub.BuildRepositorySubstitution) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
