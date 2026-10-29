package envprofile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// purgeRetained branches before readSource only for proven absence of the whole
// installed namespace. Partial/unreadable installed state is never absence.
// The permanent receipt supplies exactly the old owned homes, not native
// discovery, a new capture, a generic orphan exemption or unmanage authority.
func purgeRetained(op *operation, manager, profile string) (bool, error) {
	installed, err := stateread.Lstat(ProfileDir(manager, profile))
	if err != nil {
		return true, err
	}
	if installed.Kind != stateread.KindAbsent {
		return false, nil
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return true, err
	}
	ledger, prior, err := envlifecycle.Read(identity)
	if err != nil {
		return true, err
	}
	for _, entry := range ledger.Entries {
		matches := false
		for _, name := range entry.Profiles() {
			matches = matches || name == profile
		}
		if !matches {
			continue
		}
		if entry.Purged != nil {
			if err := ledger.VerifyAbsence(); err != nil {
				return true, err
			}
			return true, nil
		}
		if entry.Removed == nil {
			return true, fmt.Errorf("%s: profile lacks a retained removal receipt", envorigin.DiagInvalid)
		}
		removed := entry.Removed
		if len(removed.ProfileTransitions) != 1 || removed.ProfileTransitions[0].Name != profile {
			return true, fmt.Errorf("%s: removal operand differs from receipt", envorigin.DiagInvalid)
		}
		// The common under-lock scan validated the complete receipt, home roots,
		// origins, installed absence and namespace inventory immediately above.
		raw, err := entry.Marshal()
		if err != nil {
			return true, err
		}
		unit := &envunit.Unit{Record: envunit.Version, ID: newTransactionID(), Manager: identity, Action: envunit.Purge, Homes: []envunit.Home{}, Profiles: append([]envunit.ProfileTransition{}, removed.ProfileTransitions...), RetainedEntry: &envunit.RetainedEntry{ID: entry.ID(), Digest: envorigin.Digest(raw), Serialized: raw}}
		plan := transaction.Plan{TransactionID: unit.ID, ProjectIdentity: "managed-home-unit", Unit: unit, Lifecycle: &envlifecycle.Change{Prior: prior}}
		for _, home := range removed.Homes {
			next := home
			next.Before = home.After
			next.DirectoryBefore = home.DirectoryAfter
			next.After = envunit.HomeImage{Kind: "absent"}
			next.DirectoryAfter = envunit.Image{Kind: "absent"}
			unit.Homes = append(unit.Homes, next)
			plan.Opaque = append(plan.Opaque, transaction.OpaqueTarget{LivePath: home.Root, Before: home.DirectoryAfter})
		}
		if _, err := op.engine.Prepare(op.lock, plan); err != nil {
			return true, err
		}
		if err := op.engine.Commit(op.lock, plan.TransactionID); err != nil {
			return true, err
		}
		return true, nil
	}
	// A plain absent source has no deletion authority. Preserve unknown homes,
	// including an unrecognized orphan sidecar, rather than guessing an owner.
	if _, err := os.Lstat(filepath.Join(EnvRoot(manager), profile)); err == nil {
		return true, fmt.Errorf("%s: unexplained orphan has no purge authority", envorigin.DiagInvalid)
	} else if !os.IsNotExist(err) {
		return true, err
	}
	return true, fmt.Errorf("%s: no retained removal receipt for absent profile", envorigin.DiagInvalid)
}
