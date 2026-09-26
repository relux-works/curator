package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/stateread"
)

// readManifestDocument reads one Skillfile. Schema 2 source declarations
// are part of the default project reader path.
func readProjectManifestDocument(root string) (*manifest.Manifest, string, []byte, error) {
	path := manifest.PathIn(root)
	current, err := readDocument(path)
	if err != nil {
		return nil, "", nil, err
	}
	if !current.exists {
		return nil, current.generation, nil, nil
	}
	parsed, err := manifest.ParseBytes(current.payload, path)
	if err != nil {
		return nil, "", nil, err
	}
	return parsed, current.generation, current.payload, nil
}

// DraftBindingsPath is the deterministic machine-private bindings path for
// one project. Bindings never enter the project tree: they record
// canonical source locations and root inputs beside the manager
// configuration. The CLI resolve path and frozen consumption share this
// derivation so both ends of an explicit attempt agree.
func DraftBindingsPath(home, projectRoot string) string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(home, "source-bindings", hex.EncodeToString(sum[:])[:16]+".json")
}

// draftFrozenNodes consumes the pinned lock for one project install.
// When the local snapshot store is empty, it re-materializes each missing
// member from its declared source and admits it only if both the package
// identity and projected content hash still match the lock. It never
// replaces the lock or resolves a declared tag or branch.
func draftFrozenNodes(cfg *config.Config, projectRoot string, payload []byte) ([]*closure.Node, map[string]string, *sourcelock.Lock, error) {
	home := cfg.Home()
	lockPath := sourcelock.PathIn(projectRoot)
	lock, err := sourcelock.Read(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, fmt.Errorf("source_snapshot_unavailable: no Skillfile lock at %s; run explicit resolve first", lockPath)
		}
		return nil, nil, nil, err
	}
	if err := lock.CheckStale(payload); err != nil {
		return nil, nil, nil, err
	}
	opts, sources, projectManifest, err := draftFrozenInput(home, projectRoot, cfg.SkillsRoot, payload, lock)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := replayMissingDraftSnapshots(cfg, projectRoot, lock, opts, sources, projectManifest); err != nil {
		return nil, nil, nil, err
	}
	nodes, frozen, err := closure.LoadDraftFrozenNodes(home, lock, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	return nodes, frozen, lock, nil
}

type draftReplaySource struct {
	Alias      string
	Directory  string
	Source     manifest.Source
	RootInputs []string
	HasBinding bool
}

var errLockedNetworkRepositoryMissing = errors.New("locked network repository checkout is absent")

// draftFrozenInput maps every locked Git member to the local repository
// that proves its bytes and every locked root member to its accepted
// declared identity. Draft selector roots recover their declaring alias
// from the lock member's selection index (skillfile-sources §3 binds each
// root member to its declaring selection) and resolve the proving
// repository through the machine bindings written beside the lock by the
// explicit attempt when present; a fresh machine can replay from the source
// declaration. Unchanged legacy named entries (§1 retains them) recover
// their own indexed declaration (name/source/git/tag/branch/revision) and
// resolve through the SkillsRoot checkout the legacy lane snapshotted them
// from; transitive members flowed through the legacy lane, so their
// repository is the SkillsRoot checkout the lane persists. A stale
// bindings generation fails stale.
func draftFrozenInput(home, projectRoot, skillsRoot string, payload []byte, lock *sourcelock.Lock) (closure.FrozenOptions, map[string]draftReplaySource, *manifest.Manifest, error) {
	opts := closure.FrozenOptions{GitRepos: map[string]string{}, Refs: map[string]closure.FrozenRef{}, GitKeys: map[string]string{}}
	sources := map[string]draftReplaySource{}
	projectManifest, err := manifest.ParseBytes(payload, manifest.PathIn(projectRoot))
	if err != nil {
		return opts, sources, nil, err
	}
	bindings, err := readOptionalFreshBindings(home, projectRoot, lock)
	if err != nil {
		return opts, sources, nil, err
	}
	for _, member := range lock.Members {
		pkg := member.Package
		ref := closure.FrozenRef{Source: member.Name}
		if member.Selection != nil {
			index := *member.Selection
			if index < 0 || index >= len(projectManifest.Skills) {
				return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
			}
			decl := projectManifest.Skills[index]
			if decl.Selector != nil {
				alias := decl.Selector.From
				source, ok := projectManifest.Sources[alias]
				if !ok {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				replay := draftReplaySource{Alias: alias, Directory: member.Directory, Source: source}
				if bindings != nil {
					if binding, ok := bindings.Sources[alias]; ok {
						replay.RootInputs = append([]string(nil), binding.RootInputs...)
						replay.HasBinding = true
						if source.Path == "" && binding.Location != "" {
							opts.GitRepos[member.Name] = binding.Location
						}
					}
				}
				sources[member.Name] = replay
				// Draft selector root (§3). Local draft roots declare no
				// ref and keep Source only; Git roots recover the alias.
				if pkg.Kind == sourcelock.KindLocalSnapshot {
					if source.Path == "" {
						return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring path source", member.Name)
					}
				} else if pkg.Kind == sourcelock.KindNetworkGit {
					_, source, err := declaringGitSource(projectManifest, member)
					if err != nil {
						return opts, sources, nil, err
					}
					ref.Kind, ref.Ref, ref.Git = source.Ref.Kind, source.Ref.Value, source.Git
				} else if pkg.IsGit() {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
			} else {
				// Unchanged legacy named entry (§1): the indexed declaration
				// is the root itself (name/source/git/tag/branch/revision),
				// never an alias table entry.
				if decl.Name != member.Name {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if pkg.Kind == sourcelock.KindLocalSnapshot {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if !pkg.IsGit() {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				ref.Source = decl.Source
				if ref.Source == "" {
					ref.Source = member.Name
				}
				ref.Git = decl.Git
				ref.Kind = decl.Ref.Kind
				ref.Ref = decl.Ref.Value
				if ref.Kind == "" || ref.Ref == "" {
					return opts, sources, nil, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if skillsRoot == "" {
					return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository", member.Name)
				}
				// Legacy roots were snapshotted under the legacy source key
				// (decl.Source), not the canonical repository: authenticate
				// the SkillsRoot checkout under that key.
				cacheKey := decl.Source
				if cacheKey == "" {
					cacheKey = member.Name
				}
				opts.GitKeys[member.Name] = cacheKey
				opts.GitRepos[member.Name] = filepath.Join(skillsRoot, filepath.FromSlash(cacheKey))
			}
			opts.Refs[member.Name] = ref
			continue
		}
		if pkg.IsGit() {
			// Transitive members carry no alias: their bytes live in the
			// SkillsRoot checkout the legacy lane resolved them from. Members
			// selected from one schema-9 repository share that checkout when
			// this machine has one. On a fresh machine, the source is recovered
			// from the locked requirer's frozen manifest below.
			if pkg.Kind == sourcelock.KindNetworkGit {
				if skillsRoot == "" {
					continue
				}
				if member.Directory != "." {
					location, err := lockedNetworkRepository(skillsRoot, lock, pkg.Repository)
					if err != nil {
						if errors.Is(err, errLockedNetworkRepositoryMissing) {
							continue
						}
						return opts, sources, nil, err
					}
					opts.GitRepos[member.Name] = location
					continue
				}
				location := filepath.Join(skillsRoot, filepath.FromSlash(member.Name))
				metadata, err := stateread.Lstat(location)
				if err != nil {
					return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: cannot inspect repository checkout for %s: %w", member.Name, err)
				}
				if metadata.Kind == stateread.KindAbsent {
					continue
				}
				if metadata.Kind != stateread.KindPresent {
					return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: cannot inspect repository checkout for %s: %w",
						member.Name, stateread.UnusableError(location, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
				}
				if err := gitops.EnsureRepo(location); err != nil {
					return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: repository checkout for %s is unusable: %w",
						member.Name, stateread.UnusableError(location, err))
				}
				opts.GitRepos[member.Name] = location
				continue
			}
			if skillsRoot == "" {
				return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository; capture it with an explicit attempt", member.Name)
			}
			opts.GitRepos[member.Name] = filepath.Join(skillsRoot, filepath.FromSlash(pkg.Source))
		}
	}
	return opts, sources, projectManifest, nil
}

func lockedNetworkRepository(skillsRoot string, lock *sourcelock.Lock, repository string) (string, error) {
	for _, candidate := range lock.Members {
		if candidate.Package.Kind != sourcelock.KindNetworkGit || candidate.Package.Repository != repository {
			continue
		}
		location := filepath.Join(skillsRoot, filepath.FromSlash(candidate.Name))
		metadata, err := stateread.Lstat(location)
		if err != nil {
			return "", fmt.Errorf("source_snapshot_unavailable: cannot inspect repository checkout for %s: %w", candidate.Name, err)
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		if metadata.Kind != stateread.KindPresent {
			return "", fmt.Errorf("source_snapshot_unavailable: cannot inspect repository checkout for %s: %w",
				candidate.Name, stateread.UnusableError(location, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
		}
		if err := gitops.EnsureRepo(location); err != nil {
			return "", fmt.Errorf("source_snapshot_unavailable: repository checkout for %s is unusable: %w",
				candidate.Name, stateread.UnusableError(location, err))
		}
		return location, nil
	}
	return "", fmt.Errorf("source_snapshot_unavailable: no checkout for locked repository %s: %w", repository, errLockedNetworkRepositoryMissing)
}

// declaringGitSource recovers the declaring Git source alias of one locked
// root member from its selection index: skillfile-sources §3 binds each
// root member to its declaring selection, so the alias is the From of the
// manifest selection the lock index names. Aliases that share one
// canonical repository identity must not collapse to the first alias:
// the index names the exact selection the member resolved from. Anything
// else — an index outside the manifest, a legacy declaration without a
// selector, an unknown alias, or a local path source — fails closed.
func declaringGitSource(projectManifest *manifest.Manifest, member sourcelock.Member) (string, manifest.Source, error) {
	none := manifest.Source{}
	invalid := func() (string, manifest.Source, error) {
		return "", none, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
	}
	if member.Selection == nil {
		return invalid()
	}
	index := *member.Selection
	if index < 0 || index >= len(projectManifest.Skills) {
		return invalid()
	}
	decl := projectManifest.Skills[index]
	if decl.Selector == nil {
		return invalid()
	}
	alias := decl.Selector.From
	source, ok := projectManifest.Sources[alias]
	if !ok || source.Path != "" {
		return invalid()
	}
	return alias, source, nil
}

// readOptionalFreshBindings loads machine bindings when this machine has
// them. A missing bindings file is expected on a fresh machine; a present but
// unreadable, malformed, or stale file still fails closed.
func readOptionalFreshBindings(home, projectRoot string, lock *sourcelock.Lock) (*sourcelock.Bindings, error) {
	path := DraftBindingsPath(home, projectRoot)
	bindings, err := sourcelock.ReadBindings(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := bindings.CheckFresh(lock); err != nil {
		return nil, err
	}
	return bindings, nil
}

// replayMissingDraftSnapshots restores missing caches from the declared source.
// Existing but changed cache entries remain refusals and are never overwritten.
func replayMissingDraftSnapshots(cfg *config.Config, projectRoot string, lock *sourcelock.Lock, opts closure.FrozenOptions, sources map[string]draftReplaySource, projectManifest *manifest.Manifest) error {
	var roots []sourcelock.Member
	var pending []sourcelock.Member
	for _, member := range lock.Members {
		if member.Selection != nil || opts.Refs[member.Name].Git != "" || opts.Refs[member.Name].Kind != "" {
			roots = append(roots, member)
		} else {
			pending = append(pending, member)
		}
	}
	for _, member := range roots {
		if err := replayMissingDraftSnapshot(cfg, projectRoot, member, opts, sources[member.Name], projectManifest); err != nil {
			return err
		}
	}

	// Requirers can be read only after their own committed snapshots have been
	// replayed. Discover dependencies from those frozen manifests, then repeat
	// so a transitive skill may itself declare another exact repository source.
	for len(pending) > 0 {
		declared, err := declaredDependencyReplaySources(cfg, lock, opts)
		if err != nil {
			return err
		}
		next := make([]sourcelock.Member, 0, len(pending))
		progress := false
		for _, member := range pending {
			source, hasSource := declared[member.Name]
			_, hasCheckout := opts.GitRepos[member.Name]
			if !hasSource && !hasCheckout {
				next = append(next, member)
				continue
			}
			if err := replayMissingDraftSnapshot(cfg, projectRoot, member, opts, source, projectManifest); err != nil {
				return err
			}
			progress = true
		}
		if !progress && len(next) > 0 {
			// Let the regular refusal path provide its stable diagnostic for a
			// member with neither a frozen declaration nor a local checkout.
			return replayMissingDraftSnapshot(cfg, projectRoot, next[0], opts, draftReplaySource{}, projectManifest)
		}
		pending = next
	}
	return nil
}

func replayMissingDraftSnapshot(cfg *config.Config, projectRoot string, member sourcelock.Member, opts closure.FrozenOptions, source draftReplaySource, projectManifest *manifest.Manifest) error {
	pkg := member.Package
	switch pkg.Kind {
	case sourcelock.KindLocalSnapshot:
		if _, err := snapshot.OpenLocal(cfg.Home(), pkg.Snapshot); err == nil {
			return nil
		} else if !strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
			return err
		}
		if source.Source.Path == "" {
			return fmt.Errorf("source_snapshot_unavailable: no declared path source can replay %s", member.Name)
		}
		return replayLocalDraftSnapshot(cfg, projectRoot, member, source, projectManifest)
	case sourcelock.KindNetworkGit, sourcelock.KindConfiguredGit:
		key := lockedGitCacheKey(member)
		if override := opts.GitKeys[member.Name]; override != "" {
			key = override
		}
		repo := opts.GitRepos[member.Name]
		if _, err := snapshot.AuthenticateGit(cfg.Home(), key, repo, pkg.Commit.Hex); err == nil {
			return nil
		} else if !strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
			return err
		}
		if pkg.Kind == sourcelock.KindNetworkGit && source.Source.Identity != "" && source.Source.Identity != pkg.Repository {
			return fmt.Errorf("source_snapshot_changed: declared repository identity for %s differs from the lock", member.Name)
		}
		return replayGitDraftSnapshot(cfg, projectRoot, member, key, repo, source, opts.GitRepos)
	default:
		return fmt.Errorf("source_selection_invalid: unknown package kind %q", pkg.Kind)
	}
}

func declaredDependencyReplaySources(cfg *config.Config, lock *sourcelock.Lock, opts closure.FrozenOptions) (map[string]draftReplaySource, error) {
	declared := map[string]draftReplaySource{}
	var policy *config.SourcePolicy
	policyLoaded := false
	members := append([]sourcelock.Member(nil), lock.Members...)
	sort.Slice(members, func(i, j int) bool { return members[i].Name < members[j].Name })
	for _, consumer := range members {
		if consumer.Package.Kind == sourcelock.KindLocalSnapshot && consumer.Selection == nil {
			continue
		}
		if consumer.Package.Kind != sourcelock.KindLocalSnapshot && !consumer.Package.IsGit() {
			continue
		}
		if consumer.Selection == nil && (opts.Refs[consumer.Name].Git != "" || opts.Refs[consumer.Name].Kind != "") {
			continue // a schema-1 or unchanged legacy root declaration
		}
		if consumer.Package.IsGit() && opts.GitRepos[consumer.Name] == "" {
			continue
		}
		tree, err := lockedMemberSkillTree(cfg.Home(), consumer, opts)
		if err != nil {
			if strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
				continue
			}
			return nil, err
		}
		spec, err := skillspec.Load(tree)
		if err != nil {
			return nil, fmt.Errorf("source_snapshot_changed: locked requirer %s no longer has a valid manifest: %v", consumer.Name, err)
		}
		requirementNames := make([]string, 0, len(spec.Requirements))
		for name := range spec.Requirements {
			requirementNames = append(requirementNames, name)
		}
		sort.Strings(requirementNames)
		for _, requirementName := range requirementNames {
			requirement := spec.Requirements[requirementName]
			target, ok := lock.Find(requirement.Name)
			if !ok || target.Selection != nil || opts.Refs[target.Name].Git != "" || opts.Refs[target.Name].Kind != "" {
				continue
			}
			if target.Package.Kind != sourcelock.KindNetworkGit {
				continue
			}
			directory := requirement.Directory
			if directory == "" {
				directory = "."
			}
			if directory != target.Package.Directory {
				return nil, fmt.Errorf("source_snapshot_changed: locked dependency %s selects %q, but %s declares %q", target.Name, target.Package.Directory, consumer.Name, directory)
			}
			if !policyLoaded {
				policy, err = loadReplaySourcePolicy(cfg)
				if err != nil {
					return nil, err
				}
				policyLoaded = true
			}
			resolution, err := config.ResolveRepositoryEndpoints(policy, requirement.Git, "")
			if err != nil {
				return nil, err
			}
			if resolution.Identity != target.Package.Repository {
				return nil, fmt.Errorf("source_snapshot_changed: locked dependency %s repository differs from %s manifest", target.Name, consumer.Name)
			}
			if _, exists := declared[target.Name]; exists {
				continue // stable first consumer wins, matching closure identity recovery
			}
			declared[target.Name] = draftReplaySource{
				Alias:     consumer.Name + ":" + requirementName,
				Directory: directory,
				Source: manifest.Source{
					Git:      requirement.Git,
					Identity: resolution.Identity,
					Ref:      manifest.Ref{Kind: requirement.RefKind, Value: requirement.RefValue},
				},
			}
		}
	}
	return declared, nil
}

func lockedMemberSkillTree(home string, member sourcelock.Member, opts closure.FrozenOptions) (string, error) {
	pkg := member.Package
	if pkg.Kind == sourcelock.KindLocalSnapshot {
		tree, err := snapshot.OpenLocal(home, pkg.Snapshot)
		if err != nil {
			return "", err
		}
		spec, err := skillspec.Load(tree)
		if err != nil {
			return "", fmt.Errorf("source_snapshot_changed: locked package %s no longer has a valid manifest: %v", member.Name, err)
		}
		content, err := closure.ContentHashFor(tree, spec)
		if err != nil {
			return "", err
		}
		if content != member.ContentSHA256 {
			return "", fmt.Errorf("source_snapshot_changed: locked package %s no longer matches its context digest", member.Name)
		}
		return tree, nil
	}
	key := lockedGitCacheKey(member)
	if override := opts.GitKeys[member.Name]; override != "" {
		key = override
	}
	repo := opts.GitRepos[member.Name]
	if repo == "" {
		return "", fmt.Errorf("source_snapshot_unavailable: no locked repository is available for %s", member.Name)
	}
	if _, err := snapshot.AuthenticateGit(home, key, repo, pkg.Commit.Hex); err != nil {
		return "", err
	}
	if err := verifyLockedGitMember(repo, member); err != nil {
		return "", err
	}
	tree := snapshot.Dir(home, key, pkg.Commit.Hex)
	if member.Directory != "." {
		tree = filepath.Join(tree, filepath.FromSlash(member.Directory))
	}
	return tree, nil
}

func replayLocalDraftSnapshot(cfg *config.Config, projectRoot string, member sourcelock.Member, source draftReplaySource, projectManifest *manifest.Manifest) error {
	projectAbs, err := filepath.Abs(projectRoot)
	if err != nil {
		return err
	}
	sourceRoot := source.Source.Path
	if !filepath.IsAbs(sourceRoot) {
		sourceRoot = filepath.Join(projectAbs, sourceRoot)
	}
	rootInputs := source.RootInputs
	if !source.HasBinding {
		policy, err := loadReplaySourcePolicy(cfg)
		if err != nil {
			return err
		}
		if policy != nil {
			rootInputs = policy.RootInputs[source.Alias]
		}
	}
	knownAliases := make(map[string]bool, len(projectManifest.Sources))
	for alias := range projectManifest.Sources {
		knownAliases[alias] = true
	}
	outputs := []string{filepath.Join(projectAbs, ".agents"), snapshot.LocalStoreDir(cfg.Home())}
	for _, path := range adapters.AgentPaths {
		outputs = append(outputs, filepath.Join(projectAbs, filepath.FromSlash(path)))
	}
	acquisition, err := snapshot.PrepareLocalAcquisition(sourceRoot, source.Directory, projectAbs, source.Alias, outputs, map[string][]string{source.Alias: rootInputs}, knownAliases, os.TempDir())
	if err != nil {
		if strings.Contains(err.Error(), "source_member_missing") || strings.Contains(err.Error(), "permission denied") || errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("source_snapshot_unavailable: declared path source for %s cannot be reached", member.Name)
		}
		return err
	}
	defer func() { _ = acquisition.Close() }()
	inventory, err := snapshot.Capture(acquisition)
	if err != nil {
		return err
	}
	actualPackage, err := sourcelock.LocalPackage(inventory.Snapshot)
	if err != nil {
		return err
	}
	if actualPackage != member.Package {
		return fmt.Errorf("source_snapshot_changed: declared path source for %s no longer matches its locked package identity", member.Name)
	}
	spec, err := skillspec.Load(acquisition.Staging)
	if err != nil {
		return fmt.Errorf("source_snapshot_changed: declared path source for %s no longer contains a valid package: %v", member.Name, err)
	}
	content, err := closure.ContentHashFor(acquisition.Staging, spec)
	if err != nil {
		return err
	}
	if content != member.ContentSHA256 {
		return fmt.Errorf("source_snapshot_changed: declared path source for %s no longer matches its locked content", member.Name)
	}
	_, err = snapshot.PublishLocal(cfg.Home(), acquisition.Staging, inventory)
	return err
}

func replayGitDraftSnapshot(cfg *config.Config, projectRoot string, member sourcelock.Member, key, repo string, source draftReplaySource, gitRepos map[string]string) error {
	commit := member.Package.Commit.Hex
	if repo == "" {
		if member.Package.Kind != sourcelock.KindNetworkGit || (source.Source.Git == "" && source.Source.Repository == "") {
			return fmt.Errorf("source_snapshot_unavailable: no local repository can replay %s", member.Name)
		}
		var err error
		repo, err = fetchLockedCommitFromDeclaredSource(cfg, projectRoot, source, member)
		if err != nil {
			return err
		}
	}
	if err := verifyLockedGitMember(repo, member); err != nil {
		if !strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
			return err
		}
		// A fresh binding checkout may not yet contain the locked commit.
		// Fetch the full object ID alone; never fetch tags or branches.
		fetchErr := gitops.FetchCommitIsolated(repo, commit)
		if fetchErr == nil {
			err = verifyLockedGitMember(repo, member)
			if err != nil {
				return err
			}
		} else if member.Package.Kind == sourcelock.KindNetworkGit && (source.Source.Git != "" || source.Source.Repository != "") {
			freshRepo, sourceErr := fetchLockedCommitFromDeclaredSource(cfg, projectRoot, source, member)
			if sourceErr != nil {
				return sourceErr
			}
			repo = freshRepo
			err = verifyLockedGitMember(repo, member)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot provide the locked revision", member.Name)
		}
	}
	gitRepos[member.Name] = repo
	if _, err := snapshot.Get(cfg.Home(), key, repo, commit); err != nil {
		if strings.Contains(err.Error(), "source_snapshot_changed") {
			return err
		}
		if errors.Is(err, snapshot.ErrDestinationConflict) {
			return fmt.Errorf("source_snapshot_changed: stored snapshot for %s at %s does not match its locked commit", member.Name, commit)
		}
		return fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot provide the locked revision", member.Name)
	}
	if _, err := snapshot.AuthenticateGit(cfg.Home(), key, repo, commit); err != nil {
		return err
	}
	return nil
}

func verifyLockedGitMember(repo string, member sourcelock.Member) error {
	objectFormat, err := gitops.RepositoryObjectFormat(repo)
	if err != nil {
		return fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot verify its object format", member.Name)
	}
	if objectFormat != member.Package.Commit.ObjectFormat {
		return fmt.Errorf("source_snapshot_changed: locked Git object format for %s is %q, repository uses %q", member.Name, member.Package.Commit.ObjectFormat, objectFormat)
	}
	staging, err := os.MkdirTemp("", "curator-locked-git-replay-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(staging) }()
	if err := gitops.Extract(repo, member.Package.Commit.Hex, staging); err != nil {
		return fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot provide the locked revision", member.Name)
	}
	tree := staging
	if member.Directory != "." {
		tree = filepath.Join(staging, filepath.FromSlash(member.Directory))
		rel, err := filepath.Rel(staging, tree)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("source_member_invalid: locked directory for %s escapes its Git snapshot", member.Name)
		}
		info, err := os.Lstat(tree)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source_snapshot_changed: locked Git snapshot for %s no longer carries %q", member.Name, member.Directory)
		}
	}
	spec, err := skillspec.Load(tree)
	if err != nil {
		return fmt.Errorf("source_snapshot_changed: locked Git package %s is no longer valid: %v", member.Name, err)
	}
	content, err := closure.ContentHashFor(tree, spec)
	if err != nil {
		return err
	}
	if content != member.ContentSHA256 {
		return fmt.Errorf("source_snapshot_changed: declared Git source for %s no longer matches its locked content", member.Name)
	}
	return nil
}

func fetchLockedCommitFromDeclaredSource(cfg *config.Config, projectRoot string, source draftReplaySource, member sourcelock.Member) (string, error) {
	if source.Source.Git == "" && source.Source.Repository == "" {
		return "", fmt.Errorf("source_snapshot_unavailable: no declared Git endpoint can replay %s", member.Name)
	}
	policy, err := loadReplaySourcePolicy(cfg)
	if err != nil {
		return "", err
	}
	resolution, err := config.ResolveRepositoryEndpoints(policy, source.Source.Git, source.Source.Repository)
	if err != nil {
		if strings.Contains(err.Error(), config.CodeRepositoryEndpointUnavailable) {
			return "", fmt.Errorf("source_snapshot_unavailable: no reachable endpoint for %s", member.Name)
		}
		return "", err
	}
	if !identity.Allowed(resolution.Identity, cfg.AllowedSources) {
		return "", fmt.Errorf("source not allowed for %s (identity %s)", member.Name, resolution.Identity)
	}
	if resolution.Identity != member.Package.Repository {
		return "", fmt.Errorf("source_snapshot_changed: declared repository identity for %s differs from the lock", member.Name)
	}
	projectAbs, _ := filepath.Abs(projectRoot)
	key := projectAbs + "\x00" + source.Alias + "\x00" + resolution.Identity
	sum := sha256.Sum256([]byte(key))
	repo := filepath.Join(cfg.Home(), "source-replay", hex.EncodeToString(sum[:])[:20])
	for index, attempt := range resolution.Attempts {
		target, ok := attempt.ConnectionURL()
		if !ok {
			return "", fmt.Errorf("source_selection_invalid: replay endpoint %d is malformed", index+1)
		}
		if err := gitops.FetchCommitFromURLIsolated(repo, target, member.Package.Commit.Hex); err == nil {
			return repo, nil
		} else if index+1 < len(resolution.Attempts) && buildrepo.AllowSecondAttempt(resolution.Fallback, buildrepo.ClassifyFetchOutput(err.Error())) {
			continue
		}
		break
	}
	return "", fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot provide the locked revision", member.Name)
}

func loadReplaySourcePolicy(cfg *config.Config) (*config.SourcePolicy, error) {
	path := ""
	if cfg.Path != "" {
		path = filepath.Join(filepath.Dir(cfg.Path), config.SourcePolicyFileName)
	}
	return config.LoadSourcePolicy(path)
}

func lockedGitCacheKey(member sourcelock.Member) string {
	if member.Package.Kind == sourcelock.KindConfiguredGit {
		return member.Package.Source
	}
	if member.Selection != nil {
		return member.Package.Repository
	}
	if member.Package.Kind == sourcelock.KindNetworkGit && member.Package.Directory != "." {
		return member.Package.Repository
	}
	return member.Name
}
