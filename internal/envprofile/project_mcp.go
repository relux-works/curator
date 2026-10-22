package envprofile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/projectmcpstate"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
)

// ProjectMCPLockPath is the new portable outer carrier, independent of the
// unchanged Skillfile.lock.json and all profile/native ownership records.
func ProjectMCPLockPath(root string) string { return filepath.Join(root, "Skillfile.mcp-lock.json") }
func privateProjectMCPLock(home, root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(home, "mcp-project-locks", hex.EncodeToString(sum[:])+".json")
}

// PreflightProjectMCP validates the complete source set before any acquisition.
// Inline syntax cannot be activated by a purported approval inside install.
func PreflightProjectMCP(decls map[string]projectmcp.Declaration, policy Policy) error {
	for _, name := range projectmcp.Names(decls) {
		if err := decls[name].RequirePackage(); err != nil {
			return err
		}
	}
	if len(decls) == 0 {
		return nil
	}
	if err := policy.checkHardenedPassthrough(); err != nil {
		return err
	}
	if err := config.CheckSecurityPostureMCPAllowlist(policy.SecurityPosture, policy.MCPAllowlist, true); err != nil {
		return err
	}
	if err := config.CheckSecurityPostureSourceAllowlist(policy.SecurityPosture, policy.AllowedSources); err != nil {
		return err
	}
	for _, name := range projectmcp.Names(decls) {
		source, err := canonicalGit(decls[name].Package["git"].(string))
		if err != nil {
			return &projectmcp.Error{Code: projectmcp.Invalid, Field: "package.git"}
		}
		if !identity.Allowed(source, policy.AllowedSources) {
			return fmt.Errorf("%s: package source is outside allowed_sources", DiagSourceInvalid)
		}
		if len(policy.MCPAllowlist) > 0 && !identity.Allowed(source, policy.MCPAllowlist) {
			return fmt.Errorf("mcp_package_not_allowed: package source is outside MCP allowlist")
		}
	}
	return nil
}

// ResolveProjectMCP is an explicit acquisition attempt. It uses the existing
// Git/version/signer resolver and strict package audit, then freezes snapshots.
// It does not publish a lock, ownership record or native configuration.
func ResolveProjectMCP(home string, decls map[string]projectmcp.Declaration, policy Policy) (entries []projectmcpstate.Entry, warnings []string, resultErr error) {
	if err := PreflightProjectMCP(decls, policy); err != nil {
		return nil, nil, err
	}
	entries = []projectmcpstate.Entry{}
	if len(decls) == 0 {
		return entries, nil, nil
	}
	manager := newGitManager(home).withPolicy(policy)
	manager.readOnlyAudit = true
	requirements := []contextresolve.Requirement{}
	for _, name := range projectmcp.Names(decls) {
		selector := decls[name].Package
		git := selector["git"].(string)
		source, _ := canonicalGit(git)
		manager.recordRaw(source, git)
		requirements = append(requirements, contextresolve.Requirement{Kind: contextlock.KindMCP, Name: name, Source: source, Range: textMember(selector, "range"), Tag: textMember(selector, "tag"), Revision: textMember(selector, "revision"), Directory: textMember(selector, "directory")})
	}
	// All package sources have passed both allowlists before the first fetch.
	fetched := map[string]bool{}
	for _, req := range requirements {
		if !fetched[req.Source] {
			if err := manager.fetch(req.Source); err != nil {
				return nil, nil, err
			}
			fetched[req.Source] = true
		}
	}
	resolved, err := contextresolve.ResolveMCP(projectMCPSource{manager}, requirements, contextresolve.Input{MCPAllowlist: policy.MCPAllowlist, SourceSigners: policy.SourceSigners, RequireSourceSigners: policy.RequireSourceSigners})
	if err != nil {
		return nil, nil, err
	}
	for _, name := range projectmcp.Names(decls) {
		member := resolved[name]
		stage, err := os.MkdirTemp("", "curator-project-mcp-*")
		if err != nil {
			return nil, nil, err
		}
		defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(stage)) }()
		if err := gitops.Extract(manager.repoDir(member.Source), member.Commit, stage); err != nil {
			return nil, nil, err
		}
		entry, memberWarnings, err := projectMCPEntry(home, manager, stage, member, decls[name], policy)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, entry)
		warnings = append(warnings, memberWarnings...)
	}
	// Candidate validation is complete. Protected immutable snapshots may now be
	// cached; failed lock publication still leaves previous consumers untouched.
	for _, name := range projectmcp.Names(decls) {
		member := resolved[name]
		store, err := manager.ensureEntry(home, member)
		if err != nil {
			return nil, nil, err
		}
		if err := pathboundary.ValidateWithin(contextstore.Root(home), store); err != nil {
			return nil, nil, err
		}
		actual, expected, err := storeEntryPinHashes(home, contextlock.Member{Kind: member.Kind, Name: member.Name, Source: member.Source, Commit: member.Commit}, store, member.HashVersion)
		if err != nil {
			return nil, nil, err
		}
		if actual != expected {
			return nil, nil, fmt.Errorf("source_snapshot_changed: MCP store differs from committed bytes")
		}
	}
	if warning := AllowlistEmptyWarning(policy.MCPAllowlist); warning != "" {
		warnings = append(warnings, warning)
	}
	return entries, warnings, nil
}
func textMember(obj map[string]any, key string) string { value, _ := obj[key].(string); return value }
func projectMCPEntry(home string, manager *gitManager, entry string, member contextresolve.Resolved, decl projectmcp.Declaration, policy Policy) (projectmcpstate.Entry, []string, error) {
	root, err := selectedPackageRoot(entry, member.Directory)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	manifest, err := projectmcp.LoadNamedPackage(root, member.Name)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	if manifest.Version != member.Version {
		return projectmcpstate.Entry{}, nil, fmt.Errorf("context_version_mismatch: MCP package version differs")
	}
	payload, err := os.ReadFile(filepath.Join(root, contextpkg.MCPManifestName))
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	normalized, err := projectmcp.NormalizeServer(object["server"])
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	report, err := contextaudit.DetectAtVersion(root, pinOf(member), member.HashVersion, policy.SecretWaivers)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	if finding, blocked := report.FirstBlocking(); blocked {
		return projectmcpstate.Entry{}, nil, fmt.Errorf("%s: MCP package contains a blocking finding", finding.Class)
	}
	requested := manifest.Server.EnvNames
	verdict := envfragment.ResolvePassthrough(requested, policy.PassableEnvNames, policy.PassableEnvNamesSet, envfragment.S4Enforce)
	if len(verdict.Dropped) > 0 {
		return projectmcpstate.Entry{}, nil, fmt.Errorf("mcp_contribution_policy_refused: declared environment names exceed machine ceiling")
	}
	warnings, err := strictAuditMember(home, manager, member, entry, policy)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	hash, err := hashing.ContentSHA256WithVersion(root, nil, member.HashVersion)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	serverHash, err := projectmcp.Digest(normalized)
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	declHash, err := projectmcp.Digest(decl.Model())
	if err != nil {
		return projectmcpstate.Entry{}, nil, err
	}
	var directory *string
	if member.Directory != "" {
		directory = &member.Directory
	}
	return projectmcpstate.Entry{Name: member.Name, Declaration: decl.Model(), DeclarationSHA256: declHash, Resolved: projectmcpstate.Resolved{Source: member.Source, Directory: directory, Commit: member.Commit, Version: member.Version, ContentSHA256: hash, Server: normalized, ServerSHA256: serverHash}}, warnings, nil
}

// PublishProjectMCPLock commits the portable lock, its protected authority copy,
// and unchanged source-lock/bindings records in one existing journal transaction.
func PublishProjectMCPLock(home, root string, lock *projectmcpstate.Lock, other map[string][]byte) (resultErr error) {
	if err := lock.Validate(); err != nil {
		return err
	}
	payload, err := projectmcpstate.Canonical(lock)
	if err != nil {
		return err
	}
	private := privateProjectMCPLock(home, root)
	records := map[string][]byte{}
	for path, data := range other {
		records[path] = data
	}
	records[ProjectMCPLockPath(root)] = payload
	records[private] = payload
	validate := func() error {
		for path := range records {
			boundary := root
			if within, err := filepath.Rel(home, path); err == nil && within != ".." && !strings.HasPrefix(within, ".."+string(os.PathSeparator)) {
				boundary = home
			}
			if err := validateMCPRoute(boundary, path); err != nil {
				return err
			}
		}
		current, err := stateread.ReadFile(filepath.Join(root, "Skillfile.json"))
		if err != nil {
			return err
		}
		hash, err := sourcelock.ManifestDigest(current.Bytes)
		if err != nil {
			return err
		}
		if hash != lock.ManifestSHA256 {
			return fmt.Errorf("source_lock_stale: Skillfile changed before MCP lock publication")
		}
		return nil
	}
	if err := validate(); err != nil {
		return err
	}
	targets := staging.Plan{}
	for path := range records {
		targets.Targets = append(targets.Targets, staging.Target{LivePath: path, Kind: staging.KindBytes})
	}
	snapshot, err := targets.Snapshot([]string{filepath.Join(root, "Skillfile.json")})
	if err != nil {
		return err
	}
	durable, err := snapshot.Durable()
	if err != nil {
		return err
	}
	boundaries := &staging.Boundaries{Snapshot: snapshot, Durable: durable, Admitted: []string{filepath.Join(root, "Skillfile.json")}}
	if err := privatedir.MakeAll(filepath.Dir(private)); err != nil {
		return err
	}
	if err := privatedir.Validate(filepath.Dir(private)); err != nil {
		return err
	}
	op, err := beginOperationValidated(home, validate)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, op.close()) }()
	return op.publishProtected(records, boundaries, validate)
}

// ReplayProjectMCP consumes only locally authorized exact locked input. Preview
// never fetches, resolves a moving ref, writes a cache or creates a directory.
func ReplayProjectMCP(home, root, project, manifestHash string, decls map[string]projectmcp.Declaration, policy Policy) (*projectmcpstate.Lock, []string, error) {
	return replayProjectMCP(home, root, project, manifestHash, decls, policy, false)
}

// RestoreProjectMCP permits an actual install to reconstruct a proven missing
// snapshot from the exact locked commit under the current acquisition policy.
// It never resolves a moving selector, publishes consumer records or starts a
// runtime. Preview/status use ReplayProjectMCP instead.
func RestoreProjectMCP(home, root, project, manifestHash string, decls map[string]projectmcp.Declaration, policy Policy) (*projectmcpstate.Lock, []string, error) {
	return replayProjectMCP(home, root, project, manifestHash, decls, policy, true)
}

func replayProjectMCP(home, root, project, manifestHash string, decls map[string]projectmcp.Declaration, policy Policy, restore bool) (resultLock *projectmcpstate.Lock, resultWarnings []string, resultErr error) {
	var observedLock *projectmcpstate.Lock
	defer func() {
		if resultErr != nil && resultLock == nil {
			resultLock = observedLock
		}
	}()
	preflightErr := PreflightProjectMCP(decls, policy)
	for _, decl := range decls {
		if decl.Server != nil {
			return nil, nil, preflightErr
		}
	}
	private := privateProjectMCPLock(home, root)
	if err := validateMCPRoute(home, private); err != nil {
		return nil, nil, fmt.Errorf("source_snapshot_unavailable: protected MCP lock route is unavailable: %w", err)
	}
	if err := privatedir.Validate(filepath.Dir(private)); err != nil {
		return nil, nil, fmt.Errorf("source_snapshot_unavailable: protected MCP lock is unavailable")
	}
	if err := privatedir.ValidateFile(private); err != nil {
		return nil, nil, fmt.Errorf("source_snapshot_unavailable: protected MCP lock is unavailable")
	}
	authority, err := stateread.ReadRegularFile(private)
	if err != nil {
		return nil, nil, err
	}
	if err := validateMCPRoute(root, ProjectMCPLockPath(root)); err != nil {
		return nil, nil, err
	}
	portable, err := stateread.ReadRegularFile(ProjectMCPLockPath(root))
	if err != nil {
		return nil, nil, err
	}
	if authority.Kind == stateread.KindAbsent || portable.Kind == stateread.KindAbsent {
		return nil, nil, fmt.Errorf("source_snapshot_unavailable: MCP lock is absent")
	}
	if !bytes.Equal(authority.Bytes, portable.Bytes) {
		return nil, nil, fmt.Errorf("source_lock_stale: MCP lock lacks current local authority")
	}
	lock, err := projectmcpstate.DecodeLock(authority.Bytes)
	if err != nil {
		return nil, nil, err
	}
	if err := lock.CheckManifest(project, manifestHash); err != nil {
		return nil, nil, err
	}
	observedLock = lock
	if preflightErr != nil {
		return nil, nil, preflightErr
	}
	if err := validateMCPRoute(root, sourcelock.PathIn(root)); err != nil {
		return nil, nil, err
	}
	sourceFile, err := stateread.ReadRegularFile(sourcelock.PathIn(root))
	if err != nil {
		return nil, nil, err
	}
	if sourceFile.Kind == stateread.KindAbsent {
		return nil, nil, fmt.Errorf("source_snapshot_unavailable: skills lock is absent")
	}
	enclosed, err := protocoljson.MarshalCanonical(lock.SkillsLock)
	if err != nil {
		return nil, nil, err
	}
	if !bytes.Equal(sourceFile.Bytes, enclosed) {
		return nil, nil, fmt.Errorf("source_lock_stale: enclosed skills lock differs from current lock")
	}
	if len(lock.Entries) != len(decls) {
		return nil, nil, fmt.Errorf("source_lock_stale: MCP declaration membership differs")
	}
	manager := newGitManager(home).withPolicy(policy)
	manager.readOnlyAudit = true
	warnings := []string{}
	for _, entry := range lock.Entries {
		decl, present := decls[entry.Name]
		if !present {
			return nil, nil, fmt.Errorf("source_lock_stale: MCP declaration is absent")
		}
		digest, err := projectmcp.Digest(decl.Model())
		if err != nil {
			return nil, nil, err
		}
		if digest != entry.DeclarationSHA256 {
			return nil, nil, fmt.Errorf("source_lock_stale: MCP declaration changed")
		}
		directory := ""
		if entry.Resolved.Directory != nil {
			directory = *entry.Resolved.Directory
		}
		member := contextresolve.Resolved{Kind: contextlock.KindMCP, Name: entry.Name, Source: entry.Resolved.Source, Directory: directory, Commit: entry.Resolved.Commit, Version: entry.Resolved.Version, HashVersion: hashing.WriteVersion()}
		manager.recordRaw(member.Source, textMember(decl.Package, "git"))
		store := manager.entryPath(home, member)
		metadata, err := stateread.Lstat(store)
		if err != nil {
			return nil, nil, err
		}
		if metadata.Kind == stateread.KindAbsent && restore {
			if err := restoreProjectMCPSnapshot(home, manager, member, decl, entry, policy); err != nil {
				return nil, nil, err
			}
		}
		if err := pathboundary.ValidateWithin(contextstore.Root(home), store); err != nil {
			return nil, nil, fmt.Errorf("source_snapshot_unavailable: MCP snapshot is unavailable: %w", err)
		}
		actual, expected, err := storeEntryPinHashes(home, contextlock.Member{Kind: member.Kind, Name: member.Name, Source: member.Source, Commit: member.Commit}, store, member.HashVersion)
		if err != nil {
			return nil, nil, err
		}
		if actual != expected {
			return nil, nil, fmt.Errorf("source_snapshot_changed: MCP snapshot differs from commit")
		}
		if err := contextresolve.VerifyMCPPin(localMCPSource{manager}, contextresolve.Requirement{Name: member.Name, Source: member.Source, Revision: member.Commit, Tag: replayMCPTag(decl, member.Version)}, policy.SourceSigners, policy.RequireSourceSigners); err != nil {
			return nil, nil, err
		}
		observed, memberWarnings, err := projectMCPEntry(home, manager, store, member, decl, policy)
		if err != nil {
			return nil, nil, err
		}
		if observed.Resolved.ContentSHA256 != entry.Resolved.ContentSHA256 || observed.Resolved.ServerSHA256 != entry.Resolved.ServerSHA256 {
			return nil, nil, fmt.Errorf("source_snapshot_changed: MCP content/server identity differs")
		}
		warnings = append(warnings, memberWarnings...)
	}
	return lock, warnings, nil
}

// projectMCPSource changes only this lane's server validator while keeping the
// shared source selection, version and signer machinery.
type projectMCPSource struct{ *gitManager }

func (source projectMCPSource) Manifest(kind, name, identity, directory, commit string) (*contextresolve.Package, error) {
	stage, err := os.MkdirTemp("", "curator-project-mcp-manifest-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	if err := gitops.Extract(source.repoDir(identity), commit, stage); err != nil {
		return nil, err
	}
	root, err := selectedPackageRoot(stage, directory)
	if err != nil {
		return nil, err
	}
	manifest, err := projectmcp.LoadNamedPackage(root, name)
	if err != nil {
		return nil, err
	}
	return &contextresolve.Package{Version: manifest.Version}, nil
}

// Signatures may be on an exact tag or the commit. The cached tag is usable
// only if it still peels to the locked commit; it never selects replay bytes.
func replayMCPTag(decl projectmcp.Declaration, version string) string {
	if tag := textMember(decl.Package, "tag"); tag != "" {
		return tag
	}
	if textMember(decl.Package, "range") != "" {
		return "v" + version
	}
	return ""
}

type localMCPSource struct{ *gitManager }

func (source localMCPSource) VerifyCandidate(kind, name, identity, tag, commit string, allowed []contextresolve.Signer) (contextresolve.CandidateSignatures, error) {
	if tag != "" {
		out, err := exec.Command("git", "-C", source.repoDir(identity), "rev-parse", "--verify", "refs/tags/"+tag+"^{commit}").Output()
		if err != nil || strings.TrimSpace(string(out)) != commit {
			tag = ""
		}
	}
	evidence, found, err := source.VerifyCandidateLocal(identity, tag, commit, allowed)
	if !found && err == nil {
		err = fmt.Errorf("source_snapshot_unavailable: cached signature input is absent")
	}
	return evidence, err
}

// Validate existing route components without following an absent component or
// a link. New directories may be created only below the checked boundary.
func validateMCPRoute(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("boundary_identity_unreadable: MCP target escapes boundary")
	}
	current := root
	for _, part := range strings.Split(relative, string(os.PathSeparator)) {
		next := filepath.Join(current, part)
		_, err := os.Lstat(next)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return err
		}
		current = next
	}
	return pathboundary.ValidateRouteWithOwner(root, current, pathboundary.DefaultOwnerLookup())
}

func restoreProjectMCPSnapshot(home string, manager *gitManager, member contextresolve.Resolved, decl projectmcp.Declaration, locked projectmcpstate.Entry, policy Policy) (resultErr error) {
	repo, err := manager.ensureRepo(member.Source)
	if err != nil {
		return err
	}
	manager.missingPins = map[string]bool{}
	if err := manager.acquireMissingPin(repo, member.Source, member.Commit); err != nil {
		return fmt.Errorf("source_snapshot_unavailable: exact MCP commit is unavailable: %w", err)
	}
	if err := contextresolve.VerifyMCPPin(localMCPSource{manager}, contextresolve.Requirement{Name: member.Name, Source: member.Source, Revision: member.Commit, Tag: replayMCPTag(decl, member.Version)}, policy.SourceSigners, policy.RequireSourceSigners); err != nil {
		return err
	}
	stage, err := os.MkdirTemp("", "curator-project-mcp-replay-*")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(stage)) }()
	if err := gitops.Extract(repo, member.Commit, stage); err != nil {
		return err
	}
	observed, _, err := projectMCPEntry(home, manager, stage, member, decl, policy)
	if err != nil {
		return err
	}
	if observed.Resolved.ContentSHA256 != locked.Resolved.ContentSHA256 || observed.Resolved.ServerSHA256 != locked.Resolved.ServerSHA256 {
		return fmt.Errorf("source_snapshot_changed: reconstructed MCP identity differs")
	}
	_, err = manager.ensureEntry(home, member)
	return err
}
