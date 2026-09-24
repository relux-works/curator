package install

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

// EnvDraftSourcesV1 names the draft/opt-in switch for Skillfile schema 2
// source closure (skillfile-sources revision 1). Exactly "1" enables the
// frozen resolution lane; every other value, including unset and empty,
// keeps the frozen v1 lane byte-identically. The switch is operator-owned:
// package data can neither set it nor observe it.
const EnvDraftSourcesV1 = "CURATOR_DRAFT_SOURCES_V1"

// DraftSourcesEnabled reports whether the switch selects the draft lane.
func DraftSourcesEnabled(getenv func(string) string) bool {
	return getenv != nil && getenv(EnvDraftSourcesV1) == "1"
}

func draftSourcesEnabled(opts Options) bool {
	if opts.DraftSourcesV1 {
		return true
	}
	return DraftSourcesEnabled(os.Getenv)
}

// readManifestDocumentDraft reads one Skillfile with explicit draft
// admission. Without admission a schema-2 manifest fails with the
// parser upgrade error and the frozen lane is never entered.
func readManifestDocumentDraft(root string, draft bool) (*manifest.Manifest, string, []byte, error) {
	path := manifest.PathIn(root)
	current, err := readDocument(path)
	if err != nil {
		return nil, "", nil, err
	}
	if !current.exists {
		return nil, current.generation, nil, nil
	}
	if !draft {
		parsed, err := manifest.ParseBytes(current.payload, path)
		if err != nil {
			return nil, "", nil, err
		}
		return parsed, current.generation, current.payload, nil
	}
	parsed, err := manifest.ParseBytesWithOptions(current.payload, path, manifest.ParseOptions{DraftSourcesV1: true})
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
// It never rescans collections, advances branches, or replaces a local
// snapshot: the locked membership and immutable store bytes are the
// plan. A missing lock fails without resolving; a changed manifest
// fails stale; a missing store tree fails unavailable.
func draftFrozenNodes(home, projectRoot, skillsRoot string, payload []byte) ([]*closure.Node, map[string]string, *sourcelock.Lock, error) {
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
	opts, err := draftFrozenInput(home, projectRoot, skillsRoot, payload, lock)
	if err != nil {
		return nil, nil, nil, err
	}
	nodes, frozen, err := closure.LoadDraftFrozenNodes(home, lock, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	return nodes, frozen, lock, nil
}

// draftFrozenInput maps every locked Git member to the local repository
// that proves its bytes and every locked root member to its accepted
// declared identity. Draft selector roots recover their declaring alias
// from the lock member's selection index (skillfile-sources §3 binds each
// root member to its declaring selection) and resolve the proving
// repository through the machine bindings written beside the lock by the
// explicit attempt; unchanged legacy named entries (§1 retains them) recover
// their own indexed declaration (name/source/git/tag/branch/revision) and
// resolve through the SkillsRoot checkout the legacy lane snapshotted them
// from; transitive members flowed through the legacy lane, so their
// repository is the SkillsRoot checkout the lane persists. A stale
// bindings generation fails stale; a missing repository fails closed at
// consumption without network or live adoption.
func draftFrozenInput(home, projectRoot, skillsRoot string, payload []byte, lock *sourcelock.Lock) (closure.FrozenOptions, error) {
	opts := closure.FrozenOptions{GitRepos: map[string]string{}, Refs: map[string]closure.FrozenRef{}, GitKeys: map[string]string{}}
	projectManifest, err := manifest.ParseBytesWithOptions(payload, manifest.PathIn(projectRoot), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		return opts, err
	}
	var bindings *sourcelock.Bindings
	for _, member := range lock.Members {
		pkg := member.Package
		ref := closure.FrozenRef{Source: member.Name}
		if member.Selection != nil {
			index := *member.Selection
			if index < 0 || index >= len(projectManifest.Skills) {
				return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
			}
			decl := projectManifest.Skills[index]
			if decl.Selector != nil {
				// Draft selector root (§3). Local draft roots declare no
				// ref and keep Source only; Git roots recover the alias.
				if pkg.Kind == sourcelock.KindNetworkGit {
					alias, source, err := declaringGitSource(projectManifest, member)
					if err != nil {
						return opts, err
					}
					if bindings == nil {
						bindings, err = readFreshBindings(home, projectRoot, lock)
						if err != nil {
							return opts, err
						}
					}
					binding, ok := bindings.Sources[alias]
					if !ok || binding.Location == "" {
						return opts, fmt.Errorf("source_snapshot_unavailable: no machine bindings for %s; run explicit resolve first", member.Name)
					}
					opts.GitRepos[member.Name] = binding.Location
					ref.Kind, ref.Ref, ref.Git = source.Ref.Kind, source.Ref.Value, source.Git
				} else if pkg.IsGit() {
					return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
			} else {
				// Unchanged legacy named entry (§1): the indexed declaration
				// is the root itself (name/source/git/tag/branch/revision),
				// never an alias table entry.
				if decl.Name != member.Name {
					return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if pkg.Kind == sourcelock.KindLocalSnapshot {
					return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if !pkg.IsGit() {
					return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				ref.Source = decl.Source
				if ref.Source == "" {
					ref.Source = member.Name
				}
				ref.Git = decl.Git
				ref.Kind = decl.Ref.Kind
				ref.Ref = decl.Ref.Value
				if ref.Kind == "" || ref.Ref == "" {
					return opts, fmt.Errorf("source_member_invalid: locked %s has no declaring source", member.Name)
				}
				if skillsRoot == "" {
					return opts, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository; capture it with an explicit attempt", member.Name)
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
				return opts, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository; capture it with an explicit attempt", member.Name)
			}
			if pkg.Kind == sourcelock.KindNetworkGit {
				if member.Directory != "." {
					location, err := lockedNetworkRepository(skillsRoot, lock, pkg.Repository)
					if err != nil {
						return opts, err
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
	return opts, nil
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

// readFreshBindings loads the machine bindings of one project and requires
// them to belong to the consumed lock generation. Changed machine bindings
// require explicit refresh and never reinterpret an existing lock.
func readFreshBindings(home, projectRoot string, lock *sourcelock.Lock) (*sourcelock.Bindings, error) {
	path := DraftBindingsPath(home, projectRoot)
	bindings, err := sourcelock.ReadBindings(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source_snapshot_unavailable: no machine bindings at %s; run explicit resolve first", path)
		}
		return nil, err
	}
	if err := bindings.CheckFresh(lock); err != nil {
		return nil, err
	}
	return bindings, nil
}
