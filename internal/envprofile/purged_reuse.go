package envprofile

import (
	"encoding/json"
	"fmt"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// Terminal name reuse publishes only fresh installed records and consumes the
// exact purged receipt in one CAS unit. It neither provisions a home nor grants
// launch authority. Subsequent Resolve performs genuine initial capture after
// the old namespace's permanent absence has been validated and retired.
func (op *operation) publishPurgedReuse(files map[string][]byte, removals []string, boundaries *staging.Boundaries, validate func() error) (_ bool, resultErr error) {
	home := op.engine.Home()
	identity, err := envunit.ObserveManager(home)
	if err != nil {
		return false, err
	}
	ledger, prior, err := envlifecycle.Read(identity)
	if err != nil {
		return false, err
	}
	for _, entry := range ledger.Entries {
		if entry.Purged == nil {
			continue
		}
		name := entry.Purged.ProfileName
		source, hasSource := files[sourcePath(home, name)]
		lockBytes, hasLock := files[lockPath(home, name)]
		if !hasSource && !hasLock {
			continue
		}
		if !hasSource || !hasLock || len(files) != 2 || len(removals) != 0 {
			return true, fmt.Errorf("%s: terminal name reuse requires exact fresh source and lock publication", envorigin.DiagInvalid)
		}
		if err := ledger.VerifyAbsence(); err != nil {
			return true, err
		}
		lock, err := contextlock.Parse(lockBytes)
		if err != nil {
			return true, err
		}
		canonical, err := lock.Canonical()
		if err != nil || string(canonical) != string(lockBytes) {
			return true, fmt.Errorf("%s: name reuse lock is not canonical", envorigin.DiagInvalid)
		}
		raw, err := entry.Marshal()
		if err != nil {
			return true, err
		}
		var closure []string
		for _, member := range lock.Members {
			raw, err := json.Marshal(member)
			if err != nil {
				return true, err
			}
			closure = append(closure, envorigin.Digest(raw))
		}
		unit := &envunit.Unit{Record: envunit.Version, ID: newTransactionID(), Manager: identity, Action: envunit.Reuse, Homes: []envunit.Home{}, Profiles: []envunit.ProfileTransition{{Name: name, Source: envorigin.Digest(source), Closure: closure, OldLock: envunit.RecordImage{Kind: "absent"}, NewLock: envunit.RecordImage{Kind: "file", Bytes: canonical, Digest: envorigin.Digest(canonical)}}}, RetainedEntry: &envunit.RetainedEntry{ID: entry.ID(), Digest: envorigin.Digest(raw), Serialized: raw}}
		// There are deliberately no native/home targets. Only these fixed
		// installed records may precede a later genuine provisioning capture.
		g := newHomePublicationGroup()
		g.unit = unit
		defer finishPublicationCleanup(&resultErr, g.close)
		previous := op.publishPlan
		op.publishPlan = func(plan transaction.Plan) error {
			plan.Lifecycle = &envlifecycle.Change{Prior: prior}
			if _, err := op.engine.Prepare(op.lock, plan); err != nil {
				return err
			}
			return op.engine.Commit(op.lock, plan.TransactionID)
		}
		defer func() { op.publishPlan = previous }()
		return true, op.commitHomeGroup(home, g, files, nil, boundaries, validate)
	}
	return false, nil
}
