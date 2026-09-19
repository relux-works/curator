package closure

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/whitelist"
)

// DraftResolveConfig carries the explicit inputs of one draft source
// closure resolution (skillfile-sources §3). Resolution never reads the
// portable lock and never reuses a previous snapshot: every selected
// package is frozen anew and the admitted membership is re-enumerated
// before lock publication. Callers opt in by constructing this config;
// frozen v1 manifests never reach here.
type DraftResolveConfig struct {
	ProjectRoot string
	Home        string
	// SkillsRoot serves legacy and transitive lanes only. Draft roots
	// never use it; hermetic local-only tests leave it empty.
	SkillsRoot string
	// AllowedSources carries the machine allowlist for legacy and
	// transitive acquisitions (Spec §8.2). Empty allows every source,
	// exactly like closure.Options.
	AllowedSources []string
	Manifest       *manifest.Manifest
	// ManifestPayload holds the exact Skillfile bytes the parser consumed.
	// The lock binds their CCJ-1 digest; a changed manifest fails stale.
	ManifestPayload []byte
	Expansion       manifest.ExpansionOptions
	// RootInputs maps source aliases to operator-owned root-input entries
	// for project-root packages. Changing it requires explicit refresh.
	RootInputs map[string][]string
	// Outputs lists extra absolute managed-output roots beyond the
	// project .agents trees and the local snapshot store.
	Outputs []string
	// Fetch refreshes every applicable repository exactly once through
	// the admitted acquisition path: Git aliases in pinGitAliases, legacy
	// and transitive checkouts in ensureRepo. Remote-less checkouts are
	// a no-op and alias trees the caller already refreshed stay skipped
	// through the shared FetchedRepos map. Explicit resolve/refresh sets
	// it; frozen install/launch never fetch.
	Fetch bool
	// FetchedRepos deduplicates fetches across scopes. Nil allocates one.
	FetchedRepos map[string]bool
	FetchRepo    func(string) error
	// GitResolve resolves one alias tree ref to a commit. Nil uses git.
	GitResolve func(repoRoot, kind, value string) (gitops.ResolvedRef, error)
	// StagingParent owns local acquisition staging. Empty uses TempDir default.
	StagingParent string
}

// DraftPlan is one deterministic locked plan: the portable lock, the
// machine-private bindings, the provider-first nodes, and the frozen
// tree per installed name. Frozen trees are immutable store paths,
// never live directories.
type DraftPlan struct {
	Lock        *sourcelock.Lock
	Bindings    *sourcelock.Bindings
	Nodes       []*Node
	Frozen      map[string]string
	Packages    map[string]sourcelock.Package
	ManifestSHA string
}

// draftAcquired carries one frozen root selection through BuildExpanded.
type draftAcquired struct {
	node     *Node
	pkg      sourcelock.Package
	frozen   string
	snapshot string // local inventory digest, empty for Git
}

// ResolveDraft resolves selected local/Git packages and their transitive
// dependencies into a deterministic locked plan.
//
// Local packages freeze admitted filesystem bytes (dirty and untracked
// included, never Git HEAD) via snapshot capture/publication. Git aliases
// resolve to one commit per alias; every member from one alias shares it.
// Transitive requirements unify under the existing closure rules; branch
// refs stay root-only because skill requirements admit tag/revision only.
// The returned lock is sorted by UTF-8 skill name with root selection
// indexes preserved.
func ResolveDraft(cfg DraftResolveConfig) (*DraftPlan, error) {
	if cfg.Manifest == nil || cfg.Manifest.SchemaVersion != 2 {
		return nil, fmt.Errorf("source_selection_invalid: draft resolution requires Skillfile schema 2")
	}
	if cfg.ProjectRoot == "" || cfg.Home == "" {
		return nil, fmt.Errorf("source_selection_invalid: draft resolution requires project root and home")
	}
	if len(cfg.ManifestPayload) == 0 {
		return nil, fmt.Errorf("source_selection_invalid: draft resolution requires the parsed Skillfile bytes")
	}
	manifestSHA, err := sourcelock.ManifestDigest(cfg.ManifestPayload)
	if err != nil {
		return nil, err
	}
	projectAbs, err := filepath.Abs(cfg.ProjectRoot)
	if err != nil {
		return nil, err
	}
	outputs := draftOutputs(projectAbs, cfg.Home, cfg.Outputs)
	knownAliases := map[string]bool{}
	for alias := range cfg.Manifest.Sources {
		knownAliases[alias] = true
	}
	fetched := cfg.FetchedRepos
	if fetched == nil {
		fetched = map[string]bool{}
	}
	acquired := map[string]*draftAcquired{}
	gitResolve := cfg.GitResolve
	if gitResolve == nil {
		gitResolve = func(repoRoot, kind, value string) (gitops.ResolvedRef, error) {
			return gitops.Resolve(repoRoot, kind, value)
		}
	}
	// Pin every selected Git alias to its declared commit and capture
	// that commit's immutable tree BEFORE expansion: selection names,
	// validation and collection membership must come from the resolved
	// ref, never from the acquired working checkout whose HEAD (or stale
	// tree) may differ from it. Each alias resolves exactly once; the
	// proving checkout stays separate for identity and authentication.
	pins, err := pinGitAliases(cfg, gitResolve, fetched)
	if err != nil {
		return nil, err
	}
	expansion := manifest.ExpansionOptions{
		GitRoots:    frozenGitRoots(cfg.Expansion.GitRoots, pins),
		OutputRoots: cfg.Expansion.OutputRoots,
	}
	// Baseline expansion fixes selection indexes and the admitted
	// membership before any capture. A second expansion after capture
	// must agree, otherwise membership raced and the attempt fails.
	// Both run over the pinned commit trees, never the live checkout.
	baseline, err := manifest.Expand(cfg.Manifest, expansion)
	if err != nil {
		return nil, err
	}
	indexByName := map[string]int{}
	dirByName := map[string]string{}
	for _, sel := range baseline {
		// Expand guarantees unique installed names; repeated direct
		// selections already fail with source_name_conflict there.
		indexByName[sel.Decl.Name] = sel.Index
		dirByName[sel.Decl.Name] = sel.Directory
	}
	acquire := func(sel manifest.Selection) (*Node, error) {
		decl := sel.Decl
		if decl.Selector == nil {
			return nil, fmt.Errorf("source_selection_invalid: frozen acquisition requires a selector for %s", decl.Name)
		}
		source, ok := cfg.Manifest.Sources[decl.Selector.From]
		if !ok {
			return nil, fmt.Errorf("source_alias_unknown: %s", decl.Selector.From)
		}
		if source.Path != "" {
			return acquireLocal(cfg, outputs, knownAliases, sel, source, acquired)
		}
		return acquireGitPinned(cfg, knownAliases, sel, source, pins[decl.Selector.From], acquired)
	}
	closureOpts := Options{
		SkillsRoot:     cfg.SkillsRoot,
		Home:           cfg.Home,
		AllowedSources: cfg.AllowedSources,
		FetchExisting:  cfg.Fetch,
		FetchedRepos:   fetched,
		FetchRepo:      cfg.FetchRepo,
		ScratchRoot:    "",
	}
	nodes, err := BuildExpanded(closureOpts, cfg.Manifest, expansion, AcquireSelection(acquire), nil)
	if err != nil {
		return nil, err
	}
	if err := retainBranchRule(nodes); err != nil {
		return nil, err
	}
	// Re-enumerate the admitted member set before lock publication.
	again, err := manifest.Expand(cfg.Manifest, expansion)
	if err != nil {
		return nil, err
	}
	if !sameExpansion(baseline, again) {
		return nil, fmt.Errorf("source_snapshot_changed: admitted membership changed during capture; retry the explicit attempt")
	}
	plan, err := assemblePlan(cfg, manifestSHA, nodes, acquired, indexByName, dirByName)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// RefreshDraft re-runs explicit resolution and atomically replaces the
// lock and machine bindings only after every gate succeeds. Failure
// preserves the prior files: the previous lock bytes are staged before
// publication and restored when the bindings write fails, so a failure
// at the second publication step never leaves a half-published generation.
func RefreshDraft(cfg DraftResolveConfig, lockPath, bindingsPath string) (*DraftPlan, error) {
	plan, err := ResolveDraft(cfg)
	if err != nil {
		return nil, err
	}
	priorLock, priorExisted, priorErr := stagePriorFile(lockPath)
	if priorErr != nil {
		return nil, priorErr
	}
	if err := sourcelock.Write(lockPath, plan.Lock); err != nil {
		return nil, err
	}
	if bindingsPath != "" {
		if err := sourcelock.WriteBindings(bindingsPath, plan.Bindings); err != nil {
			if restoreErr := restorePriorFile(lockPath, priorLock, priorExisted); restoreErr != nil {
				return nil, fmt.Errorf("%v (rollback of %s failed: %v)", err, lockPath, restoreErr)
			}
			return nil, err
		}
	}
	return plan, nil
}

// stagePriorFile snapshots the current bytes of path before publication:
// priorExisted reports whether the file exists. An existing but
// unreadable file fails here, before any publication step runs.
func stagePriorFile(path string) (prior []byte, existed bool, err error) {
	if _, statErr := os.Lstat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, false, nil
		}
		return nil, false, statErr
	}
	prior, err = os.ReadFile(path) // #nosec G304 -- caller-supplied lock path staged for rollback
	if err != nil {
		return nil, true, err
	}
	return prior, true, nil
}

// restorePriorFile returns the file at path to its staged prior bytes:
// rewrite them atomically when the file existed, remove the new file when
// it did not. Publication writes are atomic renames, so a failed second
// write leaves the first write fully in place to roll back, and the
// restore itself is an atomic rename too, so a crash during rollback
// leaves either the prior or the published bytes behind, never a torn
// file. A crash between the two publication renames instead leaves the
// new lock beside the old bindings; that generation mismatch fails closed
// (source_lock_stale via Bindings.CheckFresh) and the next explicit
// attempt re-publishes both files.
func restorePriorFile(path string, prior []byte, existed bool) error {
	if !existed {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return sourcelock.Restore(path, prior)
}

// FrozenOptions carries the member-scoped inputs of frozen consumption.
// GitRepos maps a locked Git member name to a local repository path that
// holds its locked commit, so the cached snapshot bytes authenticate against
// the lock without network or live-source adoption. Refs maps a locked root
// member name to its accepted declared identity (ref kind/value and source
// declaration from the manifest selection the lock was resolved from), so
// materialization carries real declared values instead of empty fields.
// GitKeys optionally overrides the immutable-cache source key per member:
// legacy network-git roots were snapshotted under their legacy source path
// (decl.Source), not the canonical repository, so their key differs from
// gitCacheKey. When absent, gitCacheKey applies (draft alias roots and
// transitive members).
type FrozenOptions struct {
	GitRepos map[string]string
	Refs     map[string]FrozenRef
	GitKeys  map[string]string
}

// FrozenRef is the accepted declared identity of one locked root member:
// the ref kind/value of its manifest source alias, the declared endpoint,
// and the installed source name that mirrors the legacy source default.
type FrozenRef struct {
	Kind   string
	Ref    string
	Git    string
	Source string
}

// OpenDraftFrozen consumes one locked plan without rescanning collections,
// advancing branches, or replacing a local snapshot. Every local member
// opens its immutable store tree; an absent snapshot fails with
// source_snapshot_unavailable and a diverged store tree fails with
// source_snapshot_changed. Git members authenticate their complete cached
// commit tree (every regular file including runtime/build roots, plus
// permission bits) against the locked commit before serving the locked
// subtree; an absent commit snapshot fails unavailable without network, a
// diverged one fails changed, and a member with no repository to prove its
// bytes fails closed without adopting live bytes.
func OpenDraftFrozen(home string, lock *sourcelock.Lock, opts FrozenOptions) (map[string]string, error) {
	if lock == nil {
		return nil, fmt.Errorf("source_selection_invalid: lock is nil")
	}
	if err := lock.Validate(); err != nil {
		return nil, err
	}
	frozen := make(map[string]string, len(lock.Members))
	// Authenticate each distinct cached commit tree once: every member of
	// one alias shares its locked commit, so one digest comparison covers
	// them all before any subtree is served.
	authenticated := map[string]string{}
	for _, member := range lock.Members {
		pkg := member.Package
		switch pkg.Kind {
		case sourcelock.KindLocalSnapshot:
			path, err := snapshot.OpenLocal(home, pkg.Snapshot)
			if err != nil {
				return nil, err
			}
			frozen[member.Name] = path
		case sourcelock.KindNetworkGit, sourcelock.KindConfiguredGit:
			key := gitCacheKey(member)
			if override, ok := opts.GitKeys[member.Name]; ok && override != "" {
				key = override
			}
			target, ok := authenticated[key+"\x00"+pkg.Commit.Hex]
			if !ok {
				repo, ok := opts.GitRepos[member.Name]
				if !ok || repo == "" {
					return nil, fmt.Errorf("source_snapshot_unavailable: snapshot for %s cannot be authenticated without its locked repository; capture it with an explicit attempt", member.Name)
				}
				var err error
				target, err = openGitFrozen(home, member, key, repo)
				if err != nil {
					return nil, err
				}
				authenticated[key+"\x00"+pkg.Commit.Hex] = target
			}
			subtree, err := serveGitSubtree(target, member)
			if err != nil {
				return nil, err
			}
			frozen[member.Name] = subtree
		default:
			return nil, fmt.Errorf("source_selection_invalid: unknown package kind %q", pkg.Kind)
		}
	}
	return frozen, nil
}

// gitCacheKey returns the immutable-cache source key of one locked Git
// member. Draft alias roots were frozen by alias acquisition under the
// canonical repository identity; transitive members flowed through the
// legacy lane, whose cache key is the requirement source name — the
// installed member name — exactly as snapshotFor publishes it. Legacy
// network-git roots are the exception: they were snapshotted under their
// legacy source path (decl.Source), not the canonical repository, so
// callers must supply FrozenOptions.GitKeys for them; gitCacheKey alone
// cannot distinguish them from draft roots.
func gitCacheKey(member sourcelock.Member) string {
	pkg := member.Package
	if pkg.Kind == sourcelock.KindConfiguredGit {
		return pkg.Source
	}
	if member.Selection != nil {
		return pkg.Repository
	}
	return member.Name
}

// LoadDraftFrozenNodes rebuilds closure nodes from a validated lock and
// its frozen trees. It performs no network, no ref resolution, and no
// collection rescan: the locked membership is the plan. Every requirement
// must name a locked member, otherwise the plan is missing a member.
// Command collisions and requirement narrowing are still enforced.
func LoadDraftFrozenNodes(home string, lock *sourcelock.Lock, opts FrozenOptions) ([]*Node, map[string]string, error) {
	frozen, err := OpenDraftFrozen(home, lock, opts)
	if err != nil {
		return nil, nil, err
	}
	nodes := make([]*Node, 0, len(lock.Members))
	byName := map[string]*Node{}
	memberByName := make(map[string]sourcelock.Member, len(lock.Members))
	for _, member := range lock.Members {
		memberByName[member.Name] = member
		tree, ok := frozen[member.Name]
		if !ok {
			return nil, nil, fmt.Errorf("source_member_missing: %s has no frozen tree", member.Name)
		}
		spec, err := skillspec.Load(tree)
		if err != nil {
			return nil, nil, fmt.Errorf("source_member_invalid: %s: %w", member.Name, err)
		}
		// Recompute the projected context hash over the frozen tree and
		// require the locked value: a tampered or partially published
		// cache that changes any context byte fails here with
		// source_snapshot_changed instead of reaching installers.
		// Local members already carry full-inventory authentication
		// from OpenLocal; Git members additionally passed complete
		// commit-tree byte authentication in openGitFrozen, which also
		// covers the runtime/build bytes this projection excludes.
		content, err := ContentHashFor(tree, spec)
		if err != nil {
			return nil, nil, err
		}
		if content != member.ContentSHA256 {
			return nil, nil, fmt.Errorf("source_snapshot_changed: stored snapshot for %s does not match its locked content", member.Name)
		}
		node := &Node{Name: member.Name, Snapshot: tree, Spec: spec}
		if member.Selection != nil {
			node.Decl = manifest.Decl{Name: member.Name, Selector: &manifest.Selector{Directory: member.Directory}}
			node.Edges = []Edge{{Consumer: ProjectEdge, Mode: "full"}}
			node.Chains = []string{ProjectEdge + " -> " + member.Name}
			// Carry the accepted declared identity into the frozen node:
			// the installed source name mirrors the legacy source
			// default, and the ref kind/value with the declared endpoint
			// come from the manifest selection the lock resolved. Empty
			// fields stay empty: local snapshots declare no ref, and no
			// value is invented to satisfy downstream validation.
			if ref, ok := opts.Refs[member.Name]; ok {
				if ref.Source != "" {
					node.Decl.Source = ref.Source
				}
				node.Decl.Git = ref.Git
				if member.Package.IsGit() && ref.Kind != "" && ref.Ref != "" {
					node.Resolved = gitops.ResolvedRef{Kind: ref.Kind, Ref: ref.Ref, Commit: member.Package.Commit.Hex}
				}
			}
			if node.Decl.Source == "" {
				node.Decl.Source = member.Name
			}
		}
		if pkg := member.Package; pkg.IsGit() {
			commit := pkg.Commit.Hex
			if node.Resolved.Commit == "" {
				node.Resolved = gitops.ResolvedRef{Commit: commit}
			}
			if pkg.Kind == sourcelock.KindNetworkGit {
				node.Identity = pkg.Repository
			}
		}
		nodes = append(nodes, node)
		byName[member.Name] = node
	}
	// Wire transitive edges from frozen specs without resolving refs.
	for _, node := range nodes {
		if node.Spec == nil {
			continue
		}
		names := make([]string, 0, len(node.Spec.Requirements))
		for name := range node.Spec.Requirements {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			req := node.Spec.Requirements[name]
			target, ok := byName[req.Name]
			if !ok {
				return nil, nil, fmt.Errorf("source_member_missing: %s requires %s outside the lock", node.Name, req.Name)
			}
			target.Edges = append(target.Edges, Edge{Consumer: node.Name, Mode: req.Mode, Commands: req.Commands})
			target.Chains = append(target.Chains, node.Name+" -> "+req.Name)
		}
	}
	// Recover the accepted declared identity of transitive members from
	// their requirers' specs: every requirement that unified to the locked
	// commit names the exact ref kind/value and endpoint the legacy lane
	// resolved, and the source defaults to the requirement name exactly as
	// buildQueue records it. The first consumer in sorted order wins, which
	// matches the first-resolution-wins unification the lock was built from.
	// A transitive member no locked consumer requires keeps empty fields.
	recoverTransitiveIdentity(nodes, memberByName)
	if err := DetectActiveCommandCollisions(nodes); err != nil {
		return nil, nil, err
	}
	ordered, err := topologicalOrderLocked(nodes)
	if err != nil {
		return nil, nil, err
	}
	return ordered, frozen, nil
}

// recoverTransitiveIdentity fills the declaration and resolved ref of
// locked transitive members from the specs that require them. Root members
// already carry their manifest selection identity; transitive members carry
// only the lock, so the requirer's requirement record is the accepted
// declaration. Local snapshots never enter this path: memberForNode only
// assigns Git arms to non-acquired members.
func recoverTransitiveIdentity(nodes []*Node, memberByName map[string]sourcelock.Member) {
	ordered := make([]*Node, 0, len(nodes))
	ordered = append(ordered, nodes...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	for _, node := range ordered {
		member, ok := memberByName[node.Name]
		if !ok || member.Selection != nil || !member.Package.IsGit() {
			continue
		}
		for _, consumer := range ordered {
			if consumer.Spec == nil {
				continue
			}
			// The requirement map key is the consumer-local alias; the
			// installed name lives on the record, exactly as the edge
			// wiring above resolves it.
			keys := make([]string, 0, len(consumer.Spec.Requirements))
			for key := range consumer.Spec.Requirements {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				req := consumer.Spec.Requirements[key]
				if req.Name != node.Name {
					continue
				}
				node.Decl = manifest.Decl{
					Name:   node.Name,
					Source: node.Name,
					Ref:    manifest.Ref{Kind: req.RefKind, Value: req.RefValue},
					Git:    req.Git,
				}
				node.Resolved = gitops.ResolvedRef{Kind: req.RefKind, Ref: req.RefValue, Commit: member.Package.Commit.Hex}
				break
			}
			if node.Decl.Source != "" {
				break
			}
		}
		if node.Decl.Source == "" {
			node.Decl.Source = node.Name
		}
	}
}

// ContentHashFor hashes the projected context of one frozen tree: the
// whitelisted context roots minus declared runtime/build roots, with
// scripts/ included only for commandless skills. Runtime-only and
// build-only edits leave this digest unchanged while the snapshot
// inventory digest moves, so refresh observes a new package identity
// for the same context hash.
func ContentHashFor(frozen string, spec *skillspec.Spec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("source_member_invalid: frozen package has no spec")
	}
	includeScripts := len(spec.Commands) == 0
	excludeRoots := whitelist.ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots)
	staging, err := os.MkdirTemp("", "curator-context-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(staging) }()
	destination := filepath.Join(staging, "context")
	if _, err := whitelist.CopyContext(frozen, destination, includeScripts, excludeRoots); err != nil {
		return "", fmt.Errorf("source_member_invalid: frozen package context: %v", err)
	}
	digest, err := hashing.ContentSHA256(destination, nil)
	if err != nil {
		return "", fmt.Errorf("source_member_invalid: frozen package context hash: %v", err)
	}
	return digest, nil
}

func acquireLocal(cfg DraftResolveConfig, outputs []string, knownAliases map[string]bool, sel manifest.Selection, source manifest.Source, acquired map[string]*draftAcquired) (*Node, error) {
	projectAbs, _ := filepath.Abs(cfg.ProjectRoot)
	sourceRoot := source.Path
	if !filepath.IsAbs(sourceRoot) {
		sourceRoot = filepath.Join(projectAbs, sourceRoot)
	}
	stagingParent := cfg.StagingParent
	if stagingParent == "" {
		stagingParent = os.TempDir()
	}
	acquisition, err := snapshot.PrepareLocalAcquisition(sourceRoot, sel.Directory, projectAbs, sel.Decl.Selector.From, outputs, cfg.RootInputs, knownAliases, stagingParent)
	if err != nil {
		return nil, err
	}
	defer func() { _ = acquisition.Close() }()
	inventory, err := snapshot.Capture(acquisition)
	if err != nil {
		return nil, err
	}
	storePath, err := snapshot.PublishLocal(cfg.Home, acquisition.Staging, inventory)
	if err != nil {
		return nil, err
	}
	pkg, err := sourcelock.LocalPackage(inventory.Snapshot)
	if err != nil {
		return nil, err
	}
	spec, err := skillspec.Load(storePath)
	if err != nil {
		return nil, fmt.Errorf("source_member_invalid: %s: %w", sel.Decl.Name, err)
	}
	node := &Node{Name: sel.Decl.Name, Snapshot: storePath, Spec: spec}
	acquired[sel.Decl.Name] = &draftAcquired{node: node, pkg: pkg, frozen: storePath, snapshot: inventory.Snapshot}
	return node, nil
}

// gitPin is one Git alias resolved to its declared commit with that
// commit's immutable tree captured. repo is the acquired working checkout
// that proves the bytes; frozen is the captured commit tree that
// expansion and acquisition read instead of the checkout.
type gitPin struct {
	resolved gitops.ResolvedRef
	frozen   string
	repo     string
}

// pinGitAliases resolves every selected Git alias exactly once to its
// declared ref and captures that commit's immutable tree before any
// expansion runs. Only aliases the manifest selects are pinned: unselected
// aliases never resolve, and unknown aliases stay unknown until expansion
// reports them. Fetch (when enabled) runs here, once per alias, so ref
// resolution observes the fetched refs. The identity check fails a
// resolver that answers for a different ref than declared.
func pinGitAliases(cfg DraftResolveConfig, gitResolve func(string, string, string) (gitops.ResolvedRef, error), fetched map[string]bool) (map[string]*gitPin, error) {
	selected := map[string]bool{}
	for _, decl := range cfg.Manifest.Skills {
		if decl.Selector == nil {
			continue
		}
		alias := decl.Selector.From
		source, ok := cfg.Manifest.Sources[alias]
		if !ok || source.Path != "" {
			continue
		}
		selected[alias] = true
	}
	aliases := make([]string, 0, len(selected))
	for alias := range selected {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	pins := make(map[string]*gitPin, len(aliases))
	for _, alias := range aliases {
		source := cfg.Manifest.Sources[alias]
		root, ok := cfg.Expansion.GitRoots[alias]
		if !ok || root == "" {
			return nil, fmt.Errorf("source_selection_invalid: alias %s requires an acquired Git tree", alias)
		}
		if cfg.Fetch {
			fetch := cfg.FetchRepo
			if fetch == nil {
				fetch = gitops.Fetch
			}
			// The dedup key is the shared closure repository key, so a
			// tree the caller already refreshed (for example through the
			// production alias acquisition) is never fetched twice.
			key := repoKey(root)
			if !fetched[key] {
				if err := fetch(root); err != nil {
					return nil, err
				}
				fetched[key] = true
			}
		}
		resolved, err := gitResolve(root, source.Ref.Kind, source.Ref.Value)
		if err != nil {
			return nil, err
		}
		if resolved.Kind != source.Ref.Kind || resolved.Ref != source.Ref.Value {
			return nil, fmt.Errorf("source_member_invalid: acquired Git identity differs for alias %s", alias)
		}
		frozen, err := snapshot.Get(cfg.Home, source.Identity, root, resolved.Commit)
		if err != nil {
			return nil, err
		}
		pins[alias] = &gitPin{resolved: resolved, frozen: frozen, repo: root}
	}
	return pins, nil
}

// frozenGitRoots points expansion at the captured commit trees: every
// pinned alias reads its resolved commit, never the working checkout.
// Unpinned entries pass through untouched, and the caller's map is never
// mutated, so bindings still record the proving checkout locations.
func frozenGitRoots(roots map[string]string, pins map[string]*gitPin) map[string]string {
	frozen := make(map[string]string, len(roots))
	for alias, root := range roots {
		frozen[alias] = root
	}
	for alias, pin := range pins {
		frozen[alias] = pin.frozen
	}
	return frozen
}

func acquireGitPinned(_ DraftResolveConfig, _ map[string]bool, sel manifest.Selection, source manifest.Source, pin *gitPin, acquired map[string]*draftAcquired) (*Node, error) {
	alias := sel.Decl.Selector.From
	if pin == nil {
		return nil, fmt.Errorf("source_selection_invalid: alias %s requires an acquired Git tree", alias)
	}
	resolved := pin.resolved
	commit, err := commitFor(resolved.Commit)
	if err != nil {
		return nil, err
	}
	pkg, err := sourcelock.NetworkGitPackage(source.Identity, commit, sel.Directory)
	if err != nil {
		return nil, err
	}
	// The subtree comes from the captured commit tree pinned before
	// expansion, not from a fresh read of the checkout: selection and
	// bytes share one resolved commit.
	subtree := pin.frozen
	if sel.Directory != "." {
		subtree = filepath.Join(pin.frozen, filepath.FromSlash(sel.Directory))
	}
	spec, err := skillspec.Load(subtree)
	if err != nil {
		return nil, fmt.Errorf("source_member_invalid: %s:%s: %w", alias, sel.Directory, err)
	}
	node := &Node{
		Name:     sel.Decl.Name,
		Resolved: resolved,
		Repo:     pin.repo,
		Snapshot: subtree,
		Spec:     spec,
		Identity: source.Identity,
	}
	acquired[sel.Decl.Name] = &draftAcquired{node: node, pkg: pkg, frozen: subtree}
	return node, nil
}

func assemblePlan(cfg DraftResolveConfig, manifestSHA string, nodes []*Node, acquired map[string]*draftAcquired, indexByName map[string]int, dirByName map[string]string) (*DraftPlan, error) {
	members := make([]sourcelock.Member, 0, len(nodes))
	packages := make(map[string]sourcelock.Package, len(nodes))
	frozen := make(map[string]string, len(nodes))
	for _, node := range nodes {
		member, pkg, tree, err := memberForNode(cfg, node, acquired, indexByName, dirByName)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
		packages[node.Name] = pkg
		frozen[node.Name] = tree
	}
	lock, err := sourcelock.New(manifestSHA, members)
	if err != nil {
		return nil, err
	}
	bindings, err := bindingsFor(cfg, lock)
	if err != nil {
		return nil, err
	}
	return &DraftPlan{Lock: lock, Bindings: bindings, Nodes: nodes, Frozen: frozen, Packages: packages, ManifestSHA: manifestSHA}, nil
}

func memberForNode(_ DraftResolveConfig, node *Node, acquired map[string]*draftAcquired, indexByName map[string]int, dirByName map[string]string) (sourcelock.Member, sourcelock.Package, string, error) {
	var selection *int
	if index, ok := indexByName[node.Name]; ok {
		// Collection siblings share the zero-based skills index.
		value := index
		selection = &value
	}
	if acquiredEntry, ok := acquired[node.Name]; ok {
		directory := dirByName[node.Name]
		content, err := ContentHashFor(acquiredEntry.frozen, node.Spec)
		if err != nil {
			return sourcelock.Member{}, sourcelock.Package{}, "", err
		}
		member := sourcelock.Member{Name: node.Name, Selection: selection, Directory: directory, Package: acquiredEntry.pkg, ContentSHA256: content}
		if err := member.Validate("members"); err != nil {
			return sourcelock.Member{}, sourcelock.Package{}, "", err
		}
		return member, acquiredEntry.pkg, acquiredEntry.frozen, nil
	}
	// Legacy and transitive members flow through the existing Git lane.
	// Their directory is "." relative to the selected repository and
	// their package arm follows the canonical identity: a canonical
	// repository selects network-git, otherwise the scoped
	// configured-git identity. No snapshot digest ever enters a commit.
	directory := "."
	commit, err := commitFor(node.Resolved.Commit)
	if err != nil {
		return sourcelock.Member{}, sourcelock.Package{}, "", err
	}
	var pkg sourcelock.Package
	if node.Identity != "" {
		pkg, err = sourcelock.NetworkGitPackage(node.Identity, commit, directory)
	} else {
		source := node.Decl.Source
		if source == "" {
			source = node.Name
		}
		pkg, err = sourcelock.ConfiguredGitPackage(source, commit)
	}
	if err != nil {
		return sourcelock.Member{}, sourcelock.Package{}, "", err
	}
	content, err := ContentHashFor(node.Snapshot, node.Spec)
	if err != nil {
		return sourcelock.Member{}, sourcelock.Package{}, "", err
	}
	member := sourcelock.Member{Name: node.Name, Selection: selection, Directory: directory, Package: pkg, ContentSHA256: content}
	if err := member.Validate("members"); err != nil {
		return sourcelock.Member{}, sourcelock.Package{}, "", err
	}
	return member, pkg, node.Snapshot, nil
}

func bindingsFor(cfg DraftResolveConfig, lock *sourcelock.Lock) (*sourcelock.Bindings, error) {
	projectAbs, _ := filepath.Abs(cfg.ProjectRoot)
	sources := make(map[string]sourcelock.SourceBinding, len(cfg.Manifest.Sources))
	for alias, source := range cfg.Manifest.Sources {
		var location string
		if source.Path != "" {
			joined := source.Path
			if !filepath.IsAbs(joined) {
				joined = filepath.Join(projectAbs, joined)
			}
			resolved, err := filepath.EvalSymlinks(joined)
			if err != nil {
				return nil, fmt.Errorf("source_member_invalid: alias %s: %v", alias, err)
			}
			location = filepath.Clean(resolved)
		} else {
			root, ok := cfg.Expansion.GitRoots[alias]
			if !ok {
				return nil, fmt.Errorf("source_selection_invalid: alias %s requires an acquired Git tree", alias)
			}
			resolved, err := filepath.EvalSymlinks(root)
			if err != nil {
				resolved = root
			}
			absolute, err := filepath.Abs(resolved)
			if err != nil {
				return nil, fmt.Errorf("source_selection_invalid: alias %s: %v", alias, err)
			}
			location = filepath.Clean(absolute)
		}
		sources[alias] = sourcelock.SourceBinding{Location: location, RootInputs: append([]string(nil), cfg.RootInputs[alias]...)}
	}
	return sourcelock.NewBindings(lock.LockSHA256, sources)
}

func commitFor(hexCommit string) (sourcelock.Commit, error) {
	format := ""
	switch len(hexCommit) {
	case 40:
		format = "sha1"
	case 64:
		format = "sha256"
	default:
		return sourcelock.Commit{}, fmt.Errorf("source_member_invalid: commit %q must be 40 or 64 lowercase hex", hexCommit)
	}
	for _, c := range hexCommit {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return sourcelock.Commit{}, fmt.Errorf("source_member_invalid: commit %q must be lowercase hex", hexCommit)
		}
	}
	return sourcelock.Commit{ObjectFormat: format, Hex: hexCommit}, nil
}

// retainBranchRule keeps the root-only branch rule: skill requirements
// (transitive edges) admit exact tag or revision refs only. The parser
// already rejects branch requirements; this defensive check fails a
// branch that reaches closure through any other path.
func retainBranchRule(nodes []*Node) error {
	for _, node := range nodes {
		if node.Spec == nil {
			continue
		}
		for name, req := range node.Spec.Requirements {
			if req.RefKind == "branch" {
				return fmt.Errorf("source_selection_invalid: %s requires %s with branch %q; transitive branches are not admitted", node.Name, name, req.RefValue)
			}
		}
	}
	return nil
}

func sameExpansion(first, second []manifest.Selection) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i].Index != second[i].Index || first[i].Decl.Name != second[i].Decl.Name || first[i].Directory != second[i].Directory {
			return false
		}
	}
	return true
}

func draftOutputs(projectAbs, home string, extra []string) []string {
	outputs := append([]string(nil), extra...)
	outputs = append(outputs, filepath.Join(projectAbs, ".agents"))
	for _, rel := range adapters.AgentPaths {
		outputs = append(outputs, filepath.Join(projectAbs, filepath.FromSlash(rel)))
	}
	outputs = append(outputs, snapshot.LocalStoreDir(home))
	return outputs
}

// openGitFrozen authenticates one cached commit tree against its locked
// commit and returns the commit-tree root. A commit-shaped cache directory
// is not evidence of its bytes: the complete inventory (every regular file
// including runtime/build roots, plus permission bits) must equal a fresh
// read-back of the pinned commit, otherwise a runtime-only tamper that the
// projected context hash excludes would reach installers. Link and special
// files fail before the byte comparison so their diagnostics name the path.
func openGitFrozen(home string, member sourcelock.Member, key, repo string) (string, error) {
	pkg := member.Package
	target := gitSnapshotDir(home, key, pkg.Commit.Hex)
	info, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source_snapshot_unavailable: snapshot for %s is not in the store; capture it with an explicit attempt", member.Name)
		}
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s is not a directory", member.Name)
	}
	// A commit-shaped cache directory is not evidence of its bytes:
	// refuse any link or special file anywhere in the cached repository
	// tree before serving a member from it. This covers runtime and
	// build files that the projected context hash excludes.
	if err := refuseGitCacheLinks(target, member.Name); err != nil {
		return "", err
	}
	// Authenticate the complete locked snapshot content against the locked
	// commit: a tampered regular file, a flipped executable bit, or a
	// missing inventory member fails here with source_snapshot_changed
	// instead of reaching installers. No network and no live-source
	// adoption back this comparison, only the pinned commit read back
	// from the locked repository.
	if _, err := snapshot.AuthenticateGit(home, key, repo, pkg.Commit.Hex); err != nil {
		return "", err
	}
	return target, nil
}

// serveGitSubtree serves the authenticated locked subtree, never the
// repository root: the lock binds one source-relative directory per Git
// member and the member and package directories must agree (lock
// validation).
func serveGitSubtree(target string, member sourcelock.Member) (string, error) {
	pkg := member.Package
	directory := pkg.Directory
	if directory == "" {
		directory = member.Directory
	}
	if directory == "" || directory == "." {
		return target, nil
	}
	cleaned := filepath.Clean(filepath.FromSlash(directory))
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("source_member_invalid: locked directory %q for %s escapes its snapshot", directory, member.Name)
	}
	subtree := filepath.Join(target, cleaned)
	rel, err := filepath.Rel(target, subtree)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("source_member_invalid: locked directory %q for %s escapes its snapshot", directory, member.Name)
	}
	subInfo, err := os.Lstat(subtree)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source_snapshot_changed: stored snapshot for %s no longer carries %q", member.Name, directory)
		}
		return "", err
	}
	if !subInfo.IsDir() || subInfo.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("source_snapshot_changed: stored snapshot subtree for %s is not a directory", member.Name)
	}
	return subtree, nil
}

// refuseGitCacheLinks proves a cached Git repository tree carries no
// links or special files: every entry below root must be a real
// directory or a regular file. Anything else fails with
// source_snapshot_changed and is never passed to installers. The walk
// covers the whole commit tree, including runtime and build files that
// never enter the projected context hash.
func refuseGitCacheLinks(root, member string) error {
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source_snapshot_changed: stored snapshot for %s carries a link at %s", member, path)
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("source_snapshot_changed: stored snapshot for %s carries a special file at %s", member, path)
		}
		return nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "source_snapshot_changed") {
			return err
		}
		return fmt.Errorf("source_snapshot_changed: stored snapshot for %s cannot be read: %v", member, err)
	}
	return nil
}

func gitSnapshotDir(home, source, commit string) string {
	return filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
}

func topologicalOrderLocked(nodes []*Node) ([]*Node, error) {
	// Reuse the provider-first ordering without network: order by the
	// frozen dependency edges already wired on the nodes.
	byName := map[string]*Node{}
	for _, node := range nodes {
		byName[node.Name] = node
	}
	visited := map[string]bool{}
	temporary := map[string]bool{}
	var ordered []*Node
	var visit func(*Node) error
	visit = func(node *Node) error {
		if visited[node.Name] {
			return nil
		}
		if temporary[node.Name] {
			return fmt.Errorf("dependency cycle involving %s", node.Name)
		}
		temporary[node.Name] = true
		names := make([]string, 0, len(node.Spec.Requirements))
		if node.Spec != nil {
			for name := range node.Spec.Requirements {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			req := node.Spec.Requirements[name]
			target, ok := byName[req.Name]
			if !ok {
				return fmt.Errorf("source_member_missing: %s requires %s outside the lock", node.Name, req.Name)
			}
			if err := visit(target); err != nil {
				return err
			}
		}
		delete(temporary, node.Name)
		visited[node.Name] = true
		ordered = append(ordered, node)
		return nil
	}
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := visit(byName[name]); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

var _ = devsub.Substitution{}
var _ = gitops.ResolvedRef{}
