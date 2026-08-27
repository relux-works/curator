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
}

// ExternalSource is the exact declared/effective source passed to an injected
// acquisition boundary. It contains no credential or signing material.
type ExternalSource struct {
	Skill, Repository string
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
	result     buildrepo.PipelineResult
}

type externalPlan struct {
	scope, projectIdentity string
	rows                   []plannedExternal
	deps                   ExternalDeps
	toolchain              buildrepo.ToolchainIdentity
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

func planExternalBuilds(ctx context.Context, scope, projectIdentity, home string, nodes []*closure.Node, substitutions map[string]map[string]devsub.BuildRepositorySubstitution, toolchain Toolchain, deps ExternalDeps, dryRun bool) (externalPlan, error) {
	plan := externalPlan{scope: scope, projectIdentity: projectIdentity, deps: deps.resolved(home)}
	items := externalCommands(nodes)
	if len(items) == 0 {
		return plan, nil
	}
	target, tool, err := toolchain.Probe(ctx)
	if err != nil {
		return plan, err
	}
	plan.toolchain = externalToolchain(target, tool)
	store := &buildrepo.DiskProtectedStore{Root: plan.deps.StoreRoot}
	for _, item := range items {
		repository := item.node.Spec.BuildRepositories[item.command.Repository]
		declared := declaredRepository(repository)
		sub := repositorySubstitution(substitutions, item.node.Name, repository.Name)
		effective, err := effectiveRepository(projectIdentity, repository, sub)
		if err != nil {
			return plan, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
		}
		plan.rows = append(plan.rows, plannedExternal{node: item.node, command: item.command, repository: repository, declared: declared, effective: effective, sub: sub})
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
		source := ExternalSource{Skill: row.node.Name, Repository: row.repository.Name, Declared: row.declared, Effective: row.effective, Substitution: row.sub}
		request, err := externalPipelineRequest(plan.deps, source, plan.credentialsFor(row), plan.httpsCredentialsFor(row), row.command, store, identityOnlyExternalGo{identity: plan.toolchain}, buildrepo.OperationDryRun)
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
	root, err := private.dir("external-cache-")
	if err != nil {
		return staged, err
	}
	staged.root = root
	session, err := toolchain.Establish(ctx)
	if err != nil {
		return staged, err
	}
	defer func() { _ = session.Release() }()
	adapter := externalGoAdapter{session: session, builder: builder}
	store := &buildrepo.DiskProtectedStore{Root: root}
	for _, row := range plan.rows {
		if row.result.State == "cache-hit" {
			entryRoot := filepath.Join(plan.deps.StoreRoot, "artifacts", strings.TrimPrefix(row.result.CacheKey, "sha256:"))
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
		request, err := externalPipelineRequest(plan.deps, ExternalSource{Skill: row.node.Name, Repository: row.repository.Name, Declared: row.declared, Effective: row.effective, Substitution: row.sub}, plan.credentialsFor(row), plan.httpsCredentialsFor(row), row.command, store, adapter, buildrepo.OperationInstall)
		if err != nil {
			return stagedExternal{}, err
		}
		result, err := buildrepo.RunPipeline(ctx, request)
		if err != nil {
			return stagedExternal{}, fmt.Errorf("%s.%s: %w", row.node.Name, row.command.Name, err)
		}
		entryRoot := filepath.Join(root, "artifacts", strings.TrimPrefix(result.CacheKey, "sha256:"))
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
	return staged, nil
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
			plan.Replace("05-external-cache", skill+"/"+command, filepath.Join(finalRoot, "artifacts", key), filepath.Dir(entry.artifactPath))
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

func externalPipelineRequest(deps ExternalDeps, source ExternalSource, credentials buildrepo.OperatorSSHCredentials, httpsCredentials BuildHTTPSCredentials, command skillspec.Command, store buildrepo.ProtectedStore, goSession buildrepo.GoSession, operation buildrepo.Operation) (buildrepo.PipelineRequest, error) {
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
			git := selected.Declared.Repository
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
			return buildrepo.AcquireNetwork(ctx, buildrepo.NetworkRequest{Source: buildrepo.Source{Git: git, Transport: transport, Identity: identity}, Lock: buildrepo.LockedCommit{ObjectFormat: selected.Effective.ObjectFormat, Hex: commit}, Tag: tag, RefKind: refKind, RefValue: refValue, Tool: tool, Limits: deps.Limits})
		}
	}
	return buildrepo.PipelineRequest{Operation: operation, Command: command.Name, Target: command.Target, Declared: source.Declared, Effective: source.Effective, Acquire: func(ctx context.Context) (*buildrepo.Snapshot, error) { return acquire(ctx, source) }, Audit: deps.Audit, AuditWarnings: deps.AuditWarnings, Store: store, Go: goSession, SigningPolicy: deps.SigningPolicy}, nil
}

func externalGitTool(tool buildrepo.GitTool, source ExternalSource, sshCredentials buildrepo.OperatorSSHCredentials, httpsCredentials BuildHTTPSCredentials) buildrepo.GitTool {
	tool.SSHCredentials = sshCredentials
	if httpsCredentials.Selected() {
		tool.HTTPSCredentials = buildrepo.NewHTTPSCredentials(
			buildHTTPSHost(source.Effective.Identity), httpsCredentials.Username, httpsCredentials.Secret())
	}
	return tool
}

type identityOnlyExternalGo struct{ identity buildrepo.ToolchainIdentity }

func (g identityOnlyExternalGo) Identity() buildrepo.ToolchainIdentity { return g.identity }
func (identityOnlyExternalGo) Compile(context.Context, buildrepo.CompileRequest) ([]byte, error) {
	return nil, fmt.Errorf("read-only external plan attempted compilation")
}

type externalGoAdapter struct {
	session BuildSession
	builder Builder
}

func (g externalGoAdapter) Identity() buildrepo.ToolchainIdentity {
	return externalToolchain(g.session.Target(), g.session.Toolchain())
}
func (g externalGoAdapter) Compile(ctx context.Context, request buildrepo.CompileRequest) ([]byte, error) {
	token, err := buildsource.Validate(request.Root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = token.Close() }()
	result, err := g.builder.Stage(ctx, StageRequest{Session: g.session, Source: token, CommandObject: map[string]any{"type": "build", "driver": "go-v1", "source_dir": request.SourceDir}, BuildRoot: ".", SourceDir: request.SourceDir, Command: request.Command})
	if err != nil {
		return nil, err
	}
	return os.ReadFile(result.Path)
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
	record := marker.Build{Driver: "go-repository-v1", ReceiptSchemaVersion: 2, ExecutionPolicy: buildmeta.ExecutionPolicy, Repository: row.repository.Name, DeclaredIdentity: &marker.RepositoryIdentity{Kind: "network-git", Value: row.declared.Identity}, DeclaredLockedCommit: &marker.RepositoryCommit{ObjectFormat: row.declared.ObjectFormat, Hex: row.declared.Commit}, DeclaredTag: row.declared.Tag, EffectiveIdentity: &marker.RepositoryIdentity{Kind: row.effective.IdentityKind, Value: row.effective.Identity}, ObjectFormat: row.effective.ObjectFormat, Commit: row.effective.Commit, Substituted: row.effective.Substituted, BuildSource: &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: result.BuildSource}, DescriptorTarget: row.command.Target, CacheKey: buildmeta.CacheKey(result.CacheKey), ReceiptSHA256: buildmeta.ReceiptHash(receiptHash), ArtifactSHA256: artifactHash}
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
