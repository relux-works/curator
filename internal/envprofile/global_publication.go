package envprofile

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// NativeSkillOwnership is read-only planned custody, valid only while the
// caller's held lock and private publication staging remain alive.
type NativeSkillOwnership func(string, string, *contextlock.Member) (bool, error)

// PublishGlobalSkillPlanWithPlanner derives both publishers before executing
// either. The callback may only produce private staged targets. It receives
// the exact planned profile owner so initial native leaves cannot be claimed
// by both publishers merely because neither has published its marker yet.
func PublishGlobalSkillPlanWithPlanner(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error), lock transaction.HomeLock, stage func(NativeSkillOwnership) (transaction.Plan, error), options ...transaction.Option) (Info, error) {
	if err := lock.AssertHeld(); err != nil {
		return Info{}, err
	}
	if err := envauthority.Validate(home); err != nil {
		return Info{}, err
	}
	if stage == nil {
		return Info{}, fmt.Errorf("global legacy planner is required")
	}
	if addition != nil && (addition.Name == "" || addition.Git == "" || (addition.Tag == "") == (addition.Revision == "")) {
		return Info{}, fmt.Errorf("%s: invalid global addition", DiagRefConflict)
	}
	engine, err := newOperationEngine(home, options...)
	if err != nil {
		return Info{}, err
	}
	op := &operation{lock: borrowedGlobalLock{lock}, engine: engine, policy: policy}
	op.stageLegacy = func(unit *envunit.Unit) (transaction.Plan, error) {
		return stage(func(root, name string, requested *contextlock.Member) (bool, error) {
			path := filepath.Join(root, name)
			for _, native := range unit.Native {
				for _, rel := range native.Paths {
					if path != filepath.Join(native.Root, filepath.FromSlash(rel)) || rel != filepath.ToSlash(filepath.Join(inPlaceSkillsDir(native.Adapter), name)) {
						continue
					}
					for _, profile := range unit.Profiles {
						if profile.Name != native.Profile {
							continue
						}
						locked, err := contextlock.Parse(profile.NewLock.Bytes)
						if err != nil {
							return false, err
						}
						member, ok := locked.Find(contextlock.KindSkill, name)
						if !ok {
							return false, fmt.Errorf("%s: planned native skill lacks exact lock member", DiagSurfaceDrift)
						}
						if requested != nil && (member.Source != requested.Source || member.Commit != requested.Commit || member.Directory != requested.Directory) {
							return false, fmt.Errorf("context_range_conflict: planned native skill %q differs from global declaration", name)
						}
						return true, nil
					}
				}
			}
			return OwnsNativeSkill(home, root, name, requested)
		})
	}
	return globalSkillOperationWithOperation(home, policy, machine, addition, nativeHomeOf, op)
}

// PublishGlobalSkillPlan joins the actual staged install targets to the full
// profile/home group before Prepare. There is no post-profile mutable callback:
// every ordinary legacy failure belongs to the same durable L0 rollback.
func PublishGlobalSkillPlan(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error), lock transaction.HomeLock, legacy transaction.Plan) (Info, error) {
	if err := lock.AssertHeld(); err != nil {
		return Info{}, err
	}
	if err := envauthority.Validate(home); err != nil {
		return Info{}, err
	}
	if legacy.Unit != nil || legacy.Legacy == nil || legacy.Lifecycle != nil || len(legacy.Opaque) != 0 {
		return Info{}, fmt.Errorf("global install requires an unexecuted, closed legacy publication plan")
	}
	if addition != nil && (addition.Name == "" || addition.Git == "" || (addition.Tag == "") == (addition.Revision == "")) {
		return Info{}, fmt.Errorf("%s: invalid global addition", DiagRefConflict)
	}
	engine, err := newOperationEngine(home)
	if err != nil {
		return Info{}, err
	}
	op := &operation{lock: borrowedGlobalLock{lock}, engine: engine, policy: policy, joinedLegacy: &legacy}
	return globalSkillOperationWithOperation(home, policy, machine, addition, nativeHomeOf, op)
}

func joinLegacyPlan(profile transaction.Plan, legacy transaction.Plan) (transaction.Plan, error) {
	if profile.Unit == nil || profile.Unit.Legacy != nil || legacy.Legacy == nil {
		return transaction.Plan{}, fmt.Errorf("global plan lacks exact legacy inventory")
	}
	if legacy.BoundaryProof != nil {
		for _, target := range legacy.Targets {
			if err := staging.RecheckDurableOne(staging.Target{LivePath: target.LivePath, Kind: string(target.Kind)}, *legacy.BoundaryProof); err != nil {
				return transaction.Plan{}, err
			}
		}
	}
	profile.Unit.Legacy = legacy.Legacy
	seen := map[string]bool{}
	for _, target := range profile.Targets {
		seen[target.LivePath] = true
	}
	for _, target := range legacy.Targets {
		if seen[target.LivePath] {
			return transaction.Plan{}, fmt.Errorf("global publishers overlap: %s", target.LivePath)
		}
		if err := profile.Unit.ValidateReadScope(target.Class, target.LivePath); err != nil {
			return transaction.Plan{}, err
		}
		seen[target.LivePath] = true
		target.Identifier = target.LivePath
		profile.Targets = append(profile.Targets, target)
	}
	dirs := map[string]transaction.DirectoryAction{}
	for _, dir := range profile.Directories {
		dirs[dir.LivePath] = dir
	}
	for _, dir := range legacy.Directories {
		if prior, exists := dirs[dir.LivePath]; exists {
			if !reflect.DeepEqual(prior, dir) {
				return transaction.Plan{}, fmt.Errorf("global publishers disagree on directory preimage or protection: %s", dir.LivePath)
			}
		} else {
			profile.Directories = append(profile.Directories, dir)
			dirs[dir.LivePath] = dir
		}
	}
	if profile.ReadInputs == nil {
		profile.ReadInputs = map[string]string{}
	}
	for path, digest := range legacy.ReadInputs {
		if seen[path] {
			return transaction.Plan{}, fmt.Errorf("legacy read input overlaps global target")
		}
		if prior, ok := profile.ReadInputs[path]; ok && prior != digest {
			return transaction.Plan{}, fmt.Errorf("global input changed")
		}
		profile.ReadInputs[path] = digest
	}
	profile.ReferencedBuildKeys = append(profile.ReferencedBuildKeys, legacy.ReferencedBuildKeys...)
	// Rebuild the complete destination proof while retaining all admitted input
	// identities from both producers. Per-write durable verification survives
	// loss of the in-memory install callback/check.
	var targets []staging.Target
	for _, target := range profile.Targets {
		targets = append(targets, staging.Target{LivePath: target.LivePath, Kind: staging.KindEntry})
	}
	for _, target := range profile.Opaque {
		targets = append(targets, staging.Target{LivePath: target.LivePath, Kind: staging.KindEntry})
	}
	for _, dir := range profile.Directories {
		targets = append(targets, staging.Target{LivePath: dir.LivePath, Kind: staging.KindEntry})
	}
	var admitted []string
	for _, proof := range []*staging.DurableSnapshot{profile.BoundaryProof, legacy.BoundaryProof} {
		if proof != nil {
			for path := range proof.Admitted {
				admitted = append(admitted, path)
			}
		}
	}
	snapshot, err := (staging.Plan{Targets: targets}).Snapshot(admitted)
	if err != nil {
		return transaction.Plan{}, err
	}
	proof, err := snapshot.Durable()
	if err != nil {
		return transaction.Plan{}, err
	}
	profile.BoundaryProof = &proof
	priorCheck := profile.BoundaryCheck
	profile.BoundaryCheck = func(index int, path string) error {
		if priorCheck != nil {
			if err := priorCheck(index, path); err != nil {
				return err
			}
		}
		if legacy.BoundaryCheck != nil && legacy.Legacy != nil {
			for i, target := range legacy.Targets {
				if target.LivePath == path {
					return legacy.BoundaryCheck(i, path)
				}
			}
		}
		return nil
	}
	return profile, nil
}
