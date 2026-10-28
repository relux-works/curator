package envprofile

import (
	"bytes"
	"fmt"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/transaction"
)

// observeRetainedOrigin is confined to explicit whole-group takeover. The
// independently verified receipt supplies the exact historical L0 binding;
// current HOME/selectors/link text cannot authorize a recapture or drift waiver.
func observeRetainedOrigin(req *ResolveRequest, marker *envmarker.Marker, retained *envlifecycle.GlobalRetained) error {
	for _, h := range retained.Homes {
		if h.Profile != req.Profile || h.Adapter != req.EnvID {
			continue
		}
		if marker == nil || h.Target != ManagedHomeDir(req.Home, req.Profile, req.EnvID) {
			return fmt.Errorf("%s: retained home target differs", envorigin.DiagInvalid)
		}
		var lock envunit.RecordImage
		for _, p := range retained.ProfileTransitions {
			if p.Name == req.Profile {
				lock = p.OldLock
			}
		}
		binding := envorigin.Binding{Profile: req.Profile, ManagedTarget: h.Target, LockSHA256: lock.Digest}
		record, raw, err := envorigin.Read(req.Home, binding, req.EnvID)
		if err != nil {
			return err
		}
		if !bytes.Equal(raw, h.Before.Origin) {
			return fmt.Errorf("%s: retained origin preimage changed", envorigin.DiagInvalid)
		}
		req.origin = &originObservation{record: record, raw: raw}
		return nil
	}
	return fmt.Errorf("%s: adapter not in complete retained group", envorigin.DiagInvalid)
}

// A sync/use takeover is one complete L1 convergence, never an L2 retry. The
// receipt's complete native set controls its own writes; today's selector does
// not silently move a retained in-place target to another account/home.
func continueRetainedGlobals(op *operation, home, selected string, policy Policy, activation *scopePublication) ([]EntryResult, bool, error) {
	identity, err := envunit.ObserveManager(home)
	if err != nil {
		return nil, false, err
	}
	ledger, prior, err := envlifecycle.Read(identity)
	if err != nil {
		return nil, false, err
	}
	var entry *envlifecycle.Entry
	for i := range ledger.Entries {
		e := &ledger.Entries[i]
		if e.Global == nil {
			continue
		}
		matches := selected == ""
		for _, p := range e.Global.ProfileTransitions {
			matches = matches || p.Name == selected
		}
		if matches {
			if entry != nil {
				return nil, true, fmt.Errorf("environment_materialize_recovery_required: multiple retained groups require separate explicit profile operands")
			}
			entry = e
		}
	}
	if entry == nil {
		return nil, false, nil
	}
	if !policy.Takeover {
		return nil, true, fmt.Errorf("environment_materialize_recovery_required: global_retained_conflict; explicit sync/use --takeover required")
	}
	retained := entry.Global
	if err := envjournal.VerifyRetainedGlobal(retained); err != nil {
		return nil, true, err
	}
	op.policy = policy
	op.retained = retained
	defer func() { op.retained = nil }()
	g := newHomePublicationGroup()
	defer g.close()
	op.publicationGroup = g
	defer func() { op.publicationGroup = nil }()
	files := map[string][]byte{}
	for _, p := range retained.ProfileTransitions {
		files[lockPath(home, p.Name)] = p.NewLock.Bytes
	}
	if _, err := op.publishManagedRebind(home, files, nil, nil, nil); err != nil {
		return nil, true, err
	}
	nativeByAdapter := map[string]string{}
	for _, n := range retained.CreatedByUnit.Native {
		nativeByAdapter[n.Adapter] = n.Root
	}
	resolver := func(id string) (string, error) {
		root, ok := nativeByAdapter[id]
		if !ok {
			return "", fmt.Errorf("%s: native target not in retained group", envorigin.DiagInvalid)
		}
		return root, nil
	}
	var results []EntryResult
	for _, n := range retained.CreatedByUnit.Native {
		source, err := readSource(home, n.Profile)
		if err != nil {
			return nil, true, err
		}
		lock, hash, err := readLock(home, n.Profile)
		if err != nil {
			return nil, true, err
		}
		rows, err := materializeScopeObserved(op, home, n.Profile, n.Adapter, policy, resolver, source, lock, hash)
		if err != nil {
			return nil, true, err
		}
		for _, row := range rows {
			if !row.OK {
				if row.cause != nil {
					return rows, true, row.cause
				}
				return rows, true, fmt.Errorf("%s", row.Detail)
			}
		}
		results = append(results, rows...)
	}
	if g.unit == nil {
		return nil, true, fmt.Errorf("%s: empty retained group", envorigin.DiagInvalid)
	}
	raw, err := entry.Marshal()
	if err != nil {
		return nil, true, err
	}
	oldLocks := map[string]envunit.RecordImage{}
	for _, p := range retained.ProfileTransitions {
		oldLocks[p.Name] = p.OldLock
	}
	g.unit.Action = envunit.Takeover
	g.unit.RetainedEntry = &envunit.RetainedEntry{ID: entry.ID(), Digest: envorigin.Digest(raw), Serialized: raw, HomeLocks: oldLocks}
	op.publishPlan = func(plan transaction.Plan) error {
		plan.Lifecycle = &envlifecycle.Change{Prior: prior}
		if _, err := op.engine.Prepare(op.lock, plan); err != nil {
			return err
		}
		return op.engine.Commit(op.lock, plan.TransactionID)
	}
	defer func() { op.publishPlan = nil }()
	var removals []string
	if activation != nil {
		for path, raw := range activation.files {
			files[path] = raw
		}
		removals = activation.removals
	}
	if err := op.commitHomeGroup(home, g, files, removals, nil, nil); err != nil {
		return results, true, err
	}
	// Every inactive managed adapter was enlisted above even if no native scope
	// currently selects it; its authority capture was preserved by Rebind.
	return results, true, nil
}
