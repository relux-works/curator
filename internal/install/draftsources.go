package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
			// selected from one schema-9 repository share that checkout, so map
			// each selected package to a same-repository member's checkout.
			if skillsRoot == "" {
				return opts, sources, nil, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository; capture it with an explicit attempt", member.Name)
			}
			if pkg.Kind == sourcelock.KindNetworkGit {
				if member.Directory != "." {
					location, err := lockedNetworkRepository(skillsRoot, lock, pkg.Repository)
					if err != nil {
						return opts, sources, nil, err
					}
					opts.GitRepos[member.Name] = location
					continue
				}
				opts.GitRepos[member.Name] = filepath.Join(skillsRoot, filepath.FromSlash(member.Name))
				continue
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
		if _, err := os.Lstat(location); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("source_snapshot_unavailable: cannot inspect repository checkout for %s: %w", candidate.Name, err)
		}
		if err := gitops.EnsureRepo(location); err != nil {
			return "", fmt.Errorf("source_snapshot_unavailable: repository checkout for %s is unusable: %w", candidate.Name, err)
		}
		return location, nil
	}
	return "", fmt.Errorf("source_snapshot_unavailable: no checkout for locked repository %s", repository)
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
	for _, member := range lock.Members {
		pkg := member.Package
		switch pkg.Kind {
		case sourcelock.KindLocalSnapshot:
			if _, err := snapshot.OpenLocal(cfg.Home(), pkg.Snapshot); err == nil {
				continue
			} else if !strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
				return err
			}
			source, ok := sources[member.Name]
			if !ok {
				return fmt.Errorf("source_snapshot_unavailable: no declared path source can replay %s", member.Name)
			}
			if err := replayLocalDraftSnapshot(cfg, projectRoot, member, source, projectManifest); err != nil {
				return err
			}
		case sourcelock.KindNetworkGit, sourcelock.KindConfiguredGit:
			key := lockedGitCacheKey(member)
			if override := opts.GitKeys[member.Name]; override != "" {
				key = override
			}
			repo := opts.GitRepos[member.Name]
			if _, err := snapshot.AuthenticateGit(cfg.Home(), key, repo, pkg.Commit.Hex); err == nil {
				continue
			} else if !strings.HasPrefix(err.Error(), "source_snapshot_unavailable:") {
				return err
			}
			if pkg.Kind == sourcelock.KindNetworkGit && member.Selection != nil {
				source, ok := sources[member.Name]
				if !ok || source.Source.Identity != pkg.Repository {
					return fmt.Errorf("source_snapshot_changed: declared repository identity for %s differs from the lock", member.Name)
				}
			}
			if err := replayGitDraftSnapshot(cfg, projectRoot, member, key, repo, sources[member.Name], opts.GitRepos); err != nil {
				return err
			}
		default:
			return fmt.Errorf("source_selection_invalid: unknown package kind %q", pkg.Kind)
		}
	}
	return nil
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
		if member.Package.Kind != sourcelock.KindNetworkGit || member.Selection == nil {
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
		} else if member.Package.Kind == sourcelock.KindNetworkGit && member.Selection != nil {
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
		return fmt.Errorf("source_snapshot_unavailable: declared Git source for %s cannot provide the locked revision", member.Name)
	}
	if _, err := snapshot.AuthenticateGit(cfg.Home(), key, repo, commit); err != nil {
		return err
	}
	return nil
}

func verifyLockedGitMember(repo string, member sourcelock.Member) error {
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
	return member.Name
}
