package envauthority

import (
	"fmt"
	"path/filepath"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/projectleaserecord"
	"github.com/relux-works/curator/internal/stateread"
)

// Logical references are not made disjoint by a shared ledger pathname. Decode
// the actual private ledger with the lease owner's closed codec. Historical
// released entries still cannot resurrect a reference revoked by removal.
func validateLeaseReferences(manager string, state *lifecycleState, pending *Pending) error {
	if state.ledger == nil || len(state.ledger.Entries) == 0 {
		return nil
	}
	path := filepath.Join(manager, "state", "project-leases", "v1", "ledger.json")
	raw, present, err := envorigin.ReadPrivateAuthority(manager, path, projectleaserecord.MaxRecord)
	if err != nil {
		return err
	}
	needsGap := false
	if !present {
		for _, entry := range state.ledger.Entries {
			if entry.Global != nil && entry.Global.CreatedByUnit.CheckLeaseReferences(nil) != nil {
				needsGap = true
			}
		}
	}
	if needsGap {
		matched := false
		for _, j := range pending.LeaseJournals {
			before, after, gap, err := envjournal.LeaseReferenceGap(j, manager)
			if err != nil {
				return err
			}
			if !gap {
				continue
			}
			if matched {
				return fmt.Errorf("%s: ambiguous pending lease reference replacement", envorigin.DiagInvalid)
			}
			for _, entry := range state.ledger.Entries {
				if entry.Global != nil {
					if err := entry.Global.CreatedByUnit.CheckLeaseReferences(before); err != nil {
						return err
					}
					if after != nil {
						if err := entry.Global.CreatedByUnit.CheckLeaseReferences(after); err != nil {
							return err
						}
					}
				}
			}
			raw = before
			present = true
			matched = true
		}
	}
	for _, entry := range state.ledger.Entries {
		if entry.Global != nil {
			if err := entry.Global.CreatedByUnit.CheckLeaseReferences(raw); err != nil {
				return err
			}
		}
	}
	if !present {
		return nil
	}
	leases, err := projectleaserecord.DecodeLedger(raw)
	if err != nil {
		return fmt.Errorf("%s: lease reference ledger: %w", envorigin.DiagInvalid, err)
	}
	protected := map[string]bool{}
	for _, entry := range state.ledger.Entries {
		for _, profile := range entry.Profiles() {
			protected[profile] = true
		}
	}
	check := func(profile string) error {
		if !protected[profile] {
			return nil
		}
		for _, entry := range state.ledger.Entries {
			if entry.Global != nil {
				for _, p := range entry.Global.ProfileTransitions {
					if p.Name == profile {
						return nil
					}
				}
			}
		}
		// A genuine remove journal may restore its exact original reference bytes
		// during rollback. This exception binds the actual profile, target and raw
		// pre/postimage; a different journal or lease ledger cannot borrow it.
		for _, unit := range pending.Units {
			if unit.Action != envunit.Remove {
				continue
			}
			bound := false
			for _, p := range unit.Profiles {
				bound = bound || p.Name == profile
			}
			if !bound {
				continue
			}
			for _, target := range unit.Targets {
				if target.Path == path && target.Class == "reference-record" && (target.Before.Digest == envorigin.Digest(raw) || target.After.Digest == envorigin.Digest(raw)) {
					return nil
				}
			}
		}
		return fmt.Errorf("%s: lease reference intersects protected lifecycle profile", envorigin.DiagInvalid)
	}
	for _, entry := range leases.Leases {
		if err := check(entry.Composition.Profile); err != nil {
			return err
		}
	}
	for _, view := range leases.Views {
		if err := check(view.Composition.Profile); err != nil {
			return err
		}
	}
	return nil
}

// AdmitLeaseReferenceBytes checks the desired closed ledger before its owner
// creates parents or staging. Existing retained references are custody-only:
// exact equality is required, while new reservations still pass SelectedLifecycle.
func AdmitLeaseReferenceBytes(manager string, raw []byte) error {
	if err := Validate(manager); err != nil {
		return err
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return err
	}
	ledger, _, err := envlifecycle.Read(identity)
	if err != nil {
		return err
	}
	return checkLeaseReferenceProjection(ledger, raw)
}
func checkLeaseReferenceProjection(ledger *envlifecycle.Ledger, raw []byte) error {
	leases := &projectleaserecord.Ledger{Kind: "project-lease-state/1", Leases: map[string]projectleaserecord.Entry{}, Views: map[string]projectleaserecord.View{}}
	if len(raw) != 0 {
		var err error
		leases, err = projectleaserecord.DecodeLedger(raw)
		if err != nil {
			return fmt.Errorf("%s: desired lease record: %w", envorigin.DiagInvalid, err)
		}
	}
	blocked := map[string]bool{}
	for _, entry := range ledger.Entries {
		if entry.Global != nil {
			if err := entry.Global.CreatedByUnit.CheckLeaseReferences(raw); err != nil {
				return err
			}
			continue
		}
		for _, profile := range entry.Profiles() {
			blocked[profile] = true
		}
	}
	for _, lease := range leases.Leases {
		if blocked[lease.Composition.Profile] {
			return fmt.Errorf("%s: desired lease reference intersects protected lifecycle profile", envorigin.DiagInvalid)
		}
	}
	for _, view := range leases.Views {
		if blocked[view.Composition.Profile] {
			return fmt.Errorf("%s: desired view reference intersects protected lifecycle profile", envorigin.DiagInvalid)
		}
	}
	return nil
}

// Check the actual desired shared record at prepare and forward admission. A
// caller's owner label or callback cannot grant a semantic-reference waiver.
// The exact remove rollback exception above applies only to reachable live
// preimages; it never authorizes publishing those revoked references anew.
func admitDesiredLeaseReferences(manager string, j *envjournal.Journal, ledger *envlifecycle.Ledger) error {
	path := envunit.LeaseReferencePath(manager)
	for _, target := range j.Targets {
		if target.LivePath != path {
			continue
		}
		if target.DesiredDigest == "absent" {
			if err := checkLeaseReferenceProjection(ledger, nil); err != nil {
				return err
			}
			continue
		}
		candidates := []string{target.StagedPath, target.LivePath}
		if j.Phase == envjournal.PhasePreparing {
			candidates = []string{target.StagedSource}
		}
		found := false
		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}
			got, err := target.Digest(candidate)
			if err != nil {
				return err
			}
			if got != target.DesiredDigest {
				continue
			}
			state, err := stateread.ReadRegularFile(candidate)
			if err != nil {
				return err
			}
			if state.Kind != stateread.KindPresent || len(state.Bytes) > projectleaserecord.MaxRecord {
				return fmt.Errorf("%s: desired lease record unavailable", envorigin.DiagInvalid)
			}
			if err := checkLeaseReferenceProjection(ledger, state.Bytes); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("%s: desired lease record unavailable", envorigin.DiagInvalid)
		}
	}
	return nil
}

// CapturedLeaseProfiles identifies immutable retained reference subsets for
// automatic recovery classification. It is data, not a publication waiver:
// every resulting shared-ledger write must still pass exact subset equality.
func CapturedLeaseProfiles(manager string) (map[string]bool, error) {
	if err := Validate(manager); err != nil {
		return nil, err
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return nil, err
	}
	ledger, _, err := envlifecycle.Read(identity)
	if err != nil {
		return nil, err
	}
	profiles := map[string]bool{}
	for _, entry := range ledger.Entries {
		if entry.Global != nil {
			for _, p := range entry.Global.ProfileTransitions {
				profiles[p.Name] = true
			}
		}
	}
	return profiles, nil
}
