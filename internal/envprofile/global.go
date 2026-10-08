package envprofile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// DirectSkill is one machine-level direct skill declaration accepted by the
// global skill commands. The lock stores the resolved commit; its source and
// revision are sufficient to replay that exact declaration later.
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

// GlobalInstall replays the current profile lock's resolved skill members and
// materializes them. It is also the retry path after a sync takeover.
func GlobalInstall(home string, policy Policy, machine envregistry.MachineConfig, nativeHomeOf func(string) (string, error)) (Info, error) {
	return globalSkillOperation(home, policy, machine, nil, nativeHomeOf)
}

func globalSkillOperation(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error)) (Info, error) {
	op, err := beginOperation(home)
	if err != nil {
		return Info{}, err
	}
	closed := false
	closeOperation := func() {
		if !closed {
			_ = op.close()
			closed = true
		}
	}
	defer closeOperation()
	if err := ensureDefault(op, home, policy); err != nil {
		return Info{}, err
	}
	profile, err := currentOrDefault(home)
	if err != nil {
		return Info{}, err
	}
	source, err := readSource(home, profile)
	if err != nil {
		return Info{}, err
	}
	oldLock, oldHash, err := readLock(home, profile)
	if err != nil {
		return Info{}, err
	}
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
	if newHash != oldHash {
		if err := op.publish(map[string][]byte{lockPath(home, profile): canonical}); err != nil {
			return Info{}, err
		}
	}
	info := Info{Name: profile, Source: source, Lock: result.Lock, LockHash: newHash, Current: true, Warnings: warnings}
	// Preflight both managed and in-place homes while the manager lock is held.
	// A static unmanaged conflict is found before any surface is written, while
	// the extended lock and immutable entries are already published.
	if err := preflightManagedProfile(home, profile, policy, machine, nativeHomeOf); err != nil {
		return info, err
	}
	if err := preflightInPlaceScope(home, profile, "", policy, nativeHomeOf); err != nil {
		return info, err
	}
	nativeResults, err := materializeScopeWithNativeHome(home, profile, "", policy, nativeHomeOf)
	if err != nil {
		return info, err
	}
	for _, result := range nativeResults {
		if !result.OK {
			return info, fmt.Errorf("%s", result.Detail)
		}
	}
	closeOperation()
	if _, err := syncManagedProfile(home, profile, policy, machine, nativeHomeOf); err != nil {
		return info, err
	}
	return info, nil
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
	for _, member := range oldLock.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		if addition != nil && member.Name == addition.Name {
			continue
		}
		if member.Source == "" || member.Commit == "" {
			return contextresolve.Input{}, fmt.Errorf("%s: skill %q in profile %q lock has no replayable Git pin", DiagSourceInvalid, member.Name, profile)
		}
		input.Direct = append(input.Direct, contextresolve.Requirement{
			Kind: contextlock.KindSkill, Name: member.Name, Source: member.Source, Revision: member.Commit,
		})
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
			Repair: true, Policy: policy, NativeHomeOf: nativeHomeOf,
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
