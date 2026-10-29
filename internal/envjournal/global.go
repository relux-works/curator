package envjournal

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// GlobalPublishedTarget identifies the fixed installed publication and receipt
// that survive a handled global conflict. Homes, activation, native surfaces and legacy consumers do
// not become stable merely because their intended L1 images were journaled.
func GlobalPublishedTarget(j *Journal, path string) bool {
	if j.Global == nil || j.Unit == nil {
		return false
	}
	if j.Unit.LedgerPublication != nil && path == j.Unit.LedgerPublication.Path {
		return true
	}
	for _, p := range j.Unit.Profiles {
		root := filepath.Join(j.Unit.Manager.Path, "profiles", p.Name)
		if path == filepath.Join(root, "lock.json") || path == filepath.Join(root, "source.json") {
			return true
		}
	}
	return false
}

// GlobalPublishedDirectory identifies a parent needed by the retained installed records or ledger.
func GlobalPublishedDirectory(j *Journal, path string) bool {
	if j.Unit != nil && j.Unit.LedgerPublication != nil && strings.HasPrefix(j.Unit.LedgerPublication.Path, path+string(filepath.Separator)) {
		return true
	}
	for _, target := range j.Targets {
		if GlobalPublishedTarget(j, target.LivePath) && (path == filepath.Dir(target.LivePath) || strings.HasPrefix(target.LivePath, path+string(filepath.Separator))) {
			return true
		}
	}
	return false
}

// GlobalRetained reports the explicitly journaled retaining phase.
func GlobalRetained(j *Journal) bool { return j.Global != nil && j.Global.State == "retaining" }

// ValidateGlobal checks the typed global disposition and complete preimage bindings.
func ValidateGlobal(j *Journal) error {
	if j.Global == nil {
		// Before binding, the initial construction has no durable global disposition.
		if j.Unit != nil && j.Unit.Action == envunit.GlobalExtend {
			return unitError("global intent missing")
		}
		return nil
	}
	if j.Unit == nil || j.Unit.Action != envunit.GlobalExtend {
		return unitError("global intent action/CAS differs")
	}
	for _, action := range j.Opaque {
		if OpaqueClass(action) != envunit.BackupQuarantine {
			return unitError("global opaque action outside backup custody")
		}
		if GlobalRetained(j) && action.State != StatePending {
			return unitError("retained backup was mutated")
		}
	}
	if j.Global.Healthy {
		if j.Unit.LedgerPublication != nil || j.Global.Conflict != (envlifecycle.ConflictReceipt{}) || (j.Global.State != "planned" && j.Global.State != "forward") {
			return unitError("healthy global disposition differs")
		}
		return nil
	}
	if j.Unit.LedgerPublication == nil {
		return unitError("retained global lacks structured CAS")
	}
	if j.Global.State != "planned" && j.Global.State != "retaining" {
		return unitError("unknown global disposition")
	}
	if GlobalRetained(j) && (j.Phase == PhasePreparing || j.Phase == PhasePrepared || j.Phase == PhaseRollingBack || j.Phase == PhaseRollbackCleanup) {
		return unitError("unreachable retained phase")
	}
	c := j.Global.Conflict
	if c.Code != "environment_surface_unmanaged_conflict" || c.Phase != "profile_in_place" || c.Sequence != 1 || c.Entry.Kind == "absent" {
		return unitError("global typed conflict differs")
	}
	raw, err := json.Marshal(j.Unit.Targets)
	if err != nil {
		return err
	}
	if c.UnitPreimages != envorigin.Digest(raw) {
		return unitError("global preimage set differs")
	}
	foundProfile, foundNative, foundTarget := false, false, false
	for _, p := range j.Unit.Profiles {
		if p.Name == c.Profile && p.NewLock.Digest == c.Publication {
			foundProfile = true
		}
	}
	for _, n := range j.Unit.Native {
		if n.Profile == c.Profile && n.Adapter == c.Adapter && (c.Target == n.Root || strings.HasPrefix(c.Target, n.Root+string(filepath.Separator))) {
			foundNative = true
		}
	}
	for _, t := range j.Unit.Targets {
		if t.Path == c.Target && t.Before == c.Entry && (t.Class == "owned-bytes" || t.Class == "owned-tree") {
			foundTarget = true
		}
	}
	if !foundProfile || !foundNative || !foundTarget || c.Surface == "" {
		return unitError("global conflict not bound to exact profile/native entry")
	}
	if GlobalRetained(j) {
		for _, t := range j.Targets {
			if !GlobalPublishedTarget(j, t.LivePath) && t.State != StatePending {
				return unitError("retained home/consumer was mutated")
			}
		}
		for _, d := range j.Directories {
			if !GlobalPublishedDirectory(j, d.LivePath) && d.State != StatePending {
				return unitError("retained home parent was mutated")
			}
		}
	}
	return nil
}

// VerifyRetainedGlobal checks the permanent disposition after journal cleanup.
// Its complete preimage inventory is custody evidence, never launch authority.
func VerifyRetainedGlobal(g *envlifecycle.GlobalRetained) error {
	j := &Journal{Unit: &g.CreatedByUnit, Global: &GlobalRetention{Conflict: g.ConflictReceipt, State: "retaining"}, Phase: PhaseCleanup}
	for _, t := range g.MutablePreimages {
		if t.Class != "parent" && t.Class != "home-directory" {
			j.Targets = append(j.Targets, TargetRecord{Class: t.Class, LivePath: t.Path})
		}
	}
	if err := g.CreatedByUnit.Manager.Verify(); err != nil {
		return err
	}
	if err := g.CreatedByUnit.VerifyNativeRoots(); err != nil {
		return err
	}
	for _, t := range g.MutablePreimages {
		want := t.Before
		if GlobalPublishedTarget(j, t.Path) || (t.Class == "parent" || t.Class == "home-directory") && GlobalPublishedDirectory(j, t.Path) {
			want = t.After
		}
		actual, err := envunit.ObserveTarget(t.Path, t.Class)
		if err != nil {
			return err
		}
		if !sameImage(actual, want) {
			return unitError("retained global mutable preimage changed: " + t.Class)
		}
	}
	return nil
}
