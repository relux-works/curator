package envprofile

import (
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/manifest"
)

// DirectSkill is one machine-level direct skill declaration accepted by the
// global skill commands. Current Skillfile declarations carry the machine
// constraints; the lock records their resolved result.
type DirectSkill struct {
	Name     string
	Git      string
	Tag      string
	Revision string
}

// GlobalAdd extends the current profile lock with one direct skill and then
// materializes the resulting profile. The lock and immutable entries publish
// before any managed or in-place surface is touched. A surface conflict
// therefore leaves the lock available to profile sync --takeover.
func GlobalAdd(home string, skill DirectSkill, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error)) (Info, error) {
	if skill.Name == "" || skill.Git == "" || (skill.Tag == "") == (skill.Revision == "") {
		return Info{}, fmt.Errorf("%s: global add requires a skill name, Git source, and exactly one of tag or revision", DiagRefConflict)
	}
	return globalSkillOperation(home, policy, machine, &skill, nativeHomeOf)
}

// GlobalInstall resolves the current Skillfile and live profile requirements,
// then materializes them. Previous resolved members are not declarations.
func GlobalInstall(home string, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error)) (Info, error) {
	return globalSkillOperation(home, policy, machine, nil, nativeHomeOf)
}

func globalSkillOperation(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error)) (Info, error) {
	return globalSkillOperationWithOperation(home, policy, machine, addition, nativeHomeOf, nil)
}

func globalSkillOperationWithOperation(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error), borrowed *operation) (Info, error) {
	op, err := borrowed, error(nil)
	if op == nil {
		op, err = beginOperation(home)
	}
	if err != nil {
		return Info{}, err
	}
	op.policy = policy
	op.policy.EnvironmentMachine = machine
	closed := borrowed != nil
	closeOperation := func() {
		if !closed {
			_ = op.close()
			closed = true
		}
	}
	defer closeOperation()
	profile, err := currentOrDefault(home)
	if err != nil {
		return Info{}, err
	}
	source, err := readSource(home, profile)
	var oldLock *contextlock.Lock
	var oldHash string
	if isStateAbsent(err) && profile == DefaultProfile {
		// Only a genuinely absent installed namespace is first use. Partial,
		// unreadable and retained namespaces never acquire new authority here.
		state, e := stateread.Lstat(ProfileDir(home, profile))
		if e != nil {
			return Info{}, e
		}
		if state.Kind != stateread.KindAbsent {
			return Info{}, fmt.Errorf("%s: first-use profile namespace is not absent", DiagSourceInvalid)
		}
		oldLock, payload, e := planDefault(home, policy)
		if e != nil {
			return Info{}, e
		}
		canonical, e := oldLock.Canonical()
		if e != nil {
			return Info{}, e
		}
		op.initialProfile = map[string][]byte{sourcePath(home, profile): payload, lockPath(home, profile): canonical}
		defer func() { op.initialProfile = nil }()
		source = Source{Kind: KindLocal}
		// Preserve the planned base for resolution; no live L0 exists.
		return globalResolvedSkillOperation(op, home, profile, source, oldLock, "absent", policy, machine, addition, nativeHomeOf)
	}
	if err != nil {
		return Info{}, err
	}
	oldLock, oldHash, err = readLock(home, profile)
	if err != nil {
		return Info{}, err
	}
	return globalResolvedSkillOperation(op, home, profile, source, oldLock, oldHash, policy, machine, addition, nativeHomeOf)
}

func globalResolvedSkillOperation(op *operation, home, profile string, source Source, oldLock *contextlock.Lock, oldHash string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error)) (_ Info, resultErr error) {
	manager := newGitManager(home).withPolicy(policy)
	if err := fetchLockedSources(manager, oldLock); err != nil {
		return Info{}, err
	}
	input, err := globalSkillInput(home, profile, source, oldLock, manager, policy, addition)
	if err != nil {
		return Info{}, err
	}
	result, err := contextresolve.Resolve(manager, input)
	if err != nil {
		return Info{}, err
	}
	if err := checkAdmissionPrePublish(home, manager, result, policy); err != nil {
		return Info{}, err
	}
	warnings, err := auditAndStore(home, manager, result, policy)
	if err != nil {
		return Info{}, err
	}
	if err := ensureResolvedEntries(home, manager, result); err != nil {
		return Info{}, err
	}
	canonical, err := result.Lock.Canonical()
	if err != nil {
		return Info{}, err
	}
	newHash := contextlock.HashBytes(canonical)
	info := Info{Name: profile, Source: source, Lock: result.Lock, LockHash: newHash, Current: true, Warnings: warnings}
	g := newHomePublicationGroup()
	defer finishPublicationCleanup(&resultErr, g.close)
	op.publicationGroup = g
	defer func() { op.publicationGroup = nil }()
	records := map[string][]byte{}
	if op.initialProfile != nil {
		records[sourcePath(home, profile)] = op.initialProfile[sourcePath(home, profile)]
	}
	if newHash != oldHash {
		records[lockPath(home, profile)] = canonical
	}
	// Rebind every materialized adapter from its original capture before any
	// publication. This collector does not publish L1 or any home.
	if len(records) != 0 {
		if _, err := op.publishManagedRebind(home, records, nil, nil, nil); err != nil {
			return info, err
		}
	}
	var conflict *globalInPlaceConflict
	environments, err := globalNativeEnvironments(home, profile)
	if err != nil {
		return info, err
	}
	for _, environment := range environments {
		conflictErr := preflightInPlaceObserved(home, profile, environment, policy, nativeHomeOf, source, result.Lock, newHash)
		var candidate *globalInPlaceConflict
		if conflictErr != nil && !errors.As(conflictErr, &candidate) {
			return info, conflictErr
		}
		if conflict == nil {
			conflict = candidate
		}
	}
	planningPolicy := policy
	if conflict != nil {
		planningPolicy.Takeover = true
	}
	for _, environment := range environments {
		nativeResults, err := materializeScopeObserved(op, home, profile, environment, planningPolicy, nativeHomeOf, source, result.Lock, newHash)
		if err != nil {
			return info, err
		}
		for _, entry := range nativeResults {
			if !entry.OK {
				if entry.cause != nil {
					return info, entry.cause
				}
				return info, fmt.Errorf("%s", entry.Detail)
			}
		}
	}
	if g.unit == nil {
		return info, fmt.Errorf("global publication has no bound home inventory")
	}
	g.unit.Action = envunit.GlobalExtend
	var lifecycle *envlifecycle.Change
	if conflict != nil {
		identity, err := envunit.ObserveManager(home)
		if err != nil {
			return info, err
		}
		_, prior, err := envlifecycle.Read(identity)
		if err != nil {
			return info, err
		}
		lifecycle = &envlifecycle.Change{Prior: prior, Conflict: &conflict.capture}
	}
	op.publishPlan = func(plan transaction.Plan) error {
		plan.Unit.Action = envunit.GlobalExtend
		plan.Lifecycle = lifecycle
		if op.stageLegacy != nil && conflict == nil {
			legacy, err := op.stageLegacy(plan.Unit)
			if err != nil {
				return err
			}
			op.joinedLegacy = &legacy
		}
		if op.joinedLegacy != nil {
			var err error
			plan, err = joinLegacyPlan(plan, *op.joinedLegacy)
			if err != nil {
				return err
			}
		}
		if _, err := op.engine.Prepare(op.lock, plan); err != nil {
			return err
		}
		return op.engine.Commit(op.lock, plan.TransactionID)
	}
	defer func() { op.publishPlan = nil }()
	if err := op.commitHomeGroup(home, g, records, nil, nil, nil); err != nil {
		return info, err
	}
	return info, nil
}

// A profile lock publication affects its scoped native consumers as well as
// unscoped machine consumers. A scope selecting another profile is disjoint;
// a same-profile scope cannot hide an adapter from the complete rebind group.
func globalNativeEnvironments(home, profile string) ([]string, error) {
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	var environments []string
	for _, adapter := range Adapters {
		selected, hasScope := scoped["env:"+adapter.ID]
		if !hasScope || selected == profile {
			environments = append(environments, adapter.ID)
		}
	}
	return environments, nil
}

func currentOrDefault(home string) (string, error) {
	current, err := Current(home)
	if err != nil {
		return "", err
	}
	if current == "" {
		return DefaultProfile, nil
	}
	return current, nil
}

func globalSkillInput(home, profile string, source Source, oldLock *contextlock.Lock, manager *gitManager, policy Policy, addition *DirectSkill) (contextresolve.Input, error) {
	root, ok := oldLock.RootMember()
	if !ok {
		return contextresolve.Input{}, fmt.Errorf("%s: profile %q lock has no root member", DiagSourceInvalid, profile)
	}
	var input contextresolve.Input
	switch source.Kind {
	case KindGit:
		resolved, err := manager.inputFor(source)
		if err != nil {
			return contextresolve.Input{}, err
		}
		input = resolved
		if root.Commit == "" {
			return contextresolve.Input{}, fmt.Errorf("%s: profile %q lock has no Git root commit", DiagSourceInvalid, profile)
		}
		input.Root.Range, input.Root.Tag = "", ""
		input.Root.Revision = root.Commit
	case KindPath, KindLocal:
		if root.StateHash == "" {
			return contextresolve.Input{}, fmt.Errorf("%s: profile %q lock has no state pin for its local root", DiagSourceInvalid, profile)
		}
		entry := contextstore.EntryDir(home, contextlock.KindContext, oldLock.Root, root.StateHash)
		manifest, err := contextpkg.LoadManifest(packageRoot(entry, root.Directory))
		if err != nil {
			return contextresolve.Input{}, fmt.Errorf("%s: profile %q root snapshot cannot be read: %v", DiagSourceInvalid, profile, err)
		}
		if manifest.Name != oldLock.Root {
			return contextresolve.Input{}, fmt.Errorf("%s: profile %q root snapshot names %q, want %q", DiagSourceInvalid, profile, manifest.Name, oldLock.Root)
		}
		input = contextresolve.Input{
			Root:      contextresolve.Requirement{Kind: contextlock.KindContext, Name: oldLock.Root},
			RootState: &contextresolve.StatePackage{StateHash: root.StateHash, Manifest: packageOf(manifest)},
		}
	default:
		return contextresolve.Input{}, fmt.Errorf("%s: profile %q has unsupported source kind %q", DiagSourceInvalid, profile, source.Kind)
	}
	overlays, err := resolveOverlays(home, manager, profile, policy)
	if err != nil {
		return contextresolve.Input{}, err
	}
	input.Overlays = overlays
	input.MCPAllowlist = policy.MCPAllowlist
	// Exact Skillfile declarations constrain the selected profile alongside its
	// requirements (§9.4). Resolve disagreements before any surface is written;
	// a late adapter conflict would already have rematerialized the old profile.
	global, err := manifest.Load(filepath.Join(home, "global"))
	if err != nil {
		return contextresolve.Input{}, fmt.Errorf("%s: global Skillfile: %v", DiagSourceInvalid, err)
	}
	if global != nil {
		for _, decl := range global.Skills {
			if decl.Git == "" || (decl.Ref.Kind != "tag" && decl.Ref.Kind != "revision") || (addition != nil && decl.Name == addition.Name) {
				continue
			}
			identity, err := canonicalGit(decl.Git)
			if err != nil {
				return contextresolve.Input{}, fmt.Errorf("%s: global skill %q: %v", DiagSourceInvalid, decl.Name, err)
			}
			manager.recordRaw(identity, decl.Git)
			requirement := contextresolve.Requirement{Kind: contextlock.KindSkill, Name: decl.Name, Source: identity}
			if decl.Ref.Kind == "tag" {
				requirement.Tag = decl.Ref.Value
			} else {
				requirement.Revision = decl.Ref.Value
			}
			input.Direct = append(input.Direct, requirement)
		}
	}
	if addition != nil {
		identity, err := canonicalGit(addition.Git)
		if err != nil {
			return contextresolve.Input{}, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		manager.recordRaw(identity, addition.Git)
		input.Direct = append(input.Direct, contextresolve.Requirement{
			Kind: contextlock.KindSkill, Name: addition.Name, Source: identity,
			Tag: addition.Tag, Revision: addition.Revision,
		})
	}
	sort.Slice(input.Direct, func(i, j int) bool { return input.Direct[i].Name < input.Direct[j].Name })
	overlayInputDefaults(&input, policy)
	return input, nil
}

// SyncManagedHomesWithPolicy re-materializes every installed profile's
// managed homes from its lock. It complements the native in-place pass in
// SyncWithPolicy; both source the skill set from the same profile lock.
func SyncManagedHomesWithPolicy(home string, policy Policy, machine envregistry.MachineConfig) ([]EntryResult, error) {
	if err := EnsureDefaultWithPolicy(home, policy); err != nil {
		return nil, err
	}
	profiles, err := ListWithPolicy(home, policy)
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		if err := preflightManagedProfile(home, profile.Name, policy, machine, nil); err != nil {
			return nil, err
		}
	}
	var results []EntryResult
	var failures []string
	for _, profile := range profiles {
		profileResults, syncErr := syncManagedProfile(home, profile.Name, policy, machine, nil)
		results = append(results, profileResults...)
		if syncErr != nil {
			failures = append(failures, syncErr.Error())
		}
	}
	if len(failures) > 0 {
		return results, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return results, nil
}

func preflightManagedProfile(home, profile string, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error)) error {
	source, lock, hash, err := loadResolveInputs(home, profile)
	if err != nil {
		return err
	}
	precedence := policy.Precedence()
	for _, adapter := range envregistry.Registry {
		// Muse is managed-home-only and is provisioned by an explicit resolve.
		// Global skill changes reconcile an existing Muse home without requiring
		// every operator to have logged into the newly admitted tool.
		participates, err := managedAdapterParticipates(home, profile, adapter)
		if err != nil {
			return err
		}
		if !participates {
			continue
		}
		req := ResolveRequest{Home: home, Profile: profile, EnvID: adapter.ID, Machine: machine, Repair: true, Policy: policy, NativeHomeOf: nativeHomeOf}
		verdict := verifyHome(&req, adapter, source, lock, hash)
		if verdict.markerReadFailed {
			return fmt.Errorf("%s: %s", envmarker.DiagMarkerUnreadable, strings.Join(verdict.reasons, "; "))
		}
		if verdict.passthroughReadFailed {
			return fmt.Errorf("%s: %s", envregistry.DiagPassthroughUnreadable, strings.Join(verdict.reasons, "; "))
		}
		if verdict.current {
			continue
		}
		order, err := contextmaterialize.EmittedOrder(lock, precedence)
		if err != nil {
			return fmt.Errorf("%s: %v", DiagRepairFailed, err)
		}
		plan, err := assembleHome(&req, source, lock, hash, precedence, order, adapter, verdict.marker)
		if err != nil {
			return err
		}
		recorded := map[string]bool{}
		if verdict.marker != nil {
			for _, surface := range verdict.marker.Surfaces {
				for _, path := range surface.Paths {
					recorded[path] = true
				}
			}
		}
		want := unmanagedPlanTargets(plan, recorded)
		if len(want) > 0 && !policy.Takeover {
			if _, err := inventoryUnmanaged(plan.homeDir, want, contextstore.Root(home)); err != nil {
				return err
			}
		}
	}
	return nil
}

func syncManagedProfile(home, profile string, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error)) ([]EntryResult, error) {
	return syncManagedProfileWithOperation(home, profile, policy, machine, nativeHomeOf, nil)
}

func syncManagedProfileWithOperation(home, profile string, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error), op *operation) ([]EntryResult, error) {
	var results []EntryResult
	var failures []string
	for _, adapter := range envregistry.Registry {
		// Muse is managed-home-only and is provisioned by an explicit resolve.
		// Global skill changes reconcile an existing Muse home without requiring
		// every operator to have logged into the newly admitted tool.
		participates, err := managedAdapterParticipates(home, profile, adapter)
		if err != nil {
			return nil, err
		}
		if !participates {
			continue
		}
		_, err = Resolve(ResolveRequest{
			Home: home, Profile: profile, EnvID: adapter.ID, Machine: machine,
			Repair: true, Policy: policy, NativeHomeOf: nativeHomeOf, borrowedOperation: op,
		})
		result := EntryResult{Adapter: adapter.ID, Home: ManagedHomeDir(home, profile, adapter.ID), OK: err == nil}
		if err != nil {
			result.Detail = err.Error()
			failures = append(failures, err.Error())
		}
		results = append(results, result)
	}
	if len(failures) > 0 {
		return results, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return results, nil
}

func unmanagedPlanTargets(plan *homePlan, recorded map[string]bool) map[string]bool {
	want := map[string]bool{}
	for path := range plan.copies {
		if !recorded[path] {
			want[path] = true
		}
	}
	for path := range plan.links {
		if !recorded[path] {
			want[path] = true
		}
	}
	return want
}
