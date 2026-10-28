package envlifecycle

import (
	"encoding/json"
	"github.com/relux-works/curator/internal/contextlock"
	"path/filepath"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// Change carries a typed terminal action, never a precomputed authority verdict.
// The executor builds the receipt only after binding the real unit's complete
// targets, prepared directory identities and boundary proof. It publishes the
// resulting bytes as an ordinary private CAS target in that same transaction.
type Change struct {
	Prior          []byte
	RemovedRecords []OwnedRecord
	Conflict       *ConflictReceipt
	Revocations    []Revocation
}

func (c *Change) Bind(unit *envunit.Unit) error {
	if unit == nil || (unit.Action != envunit.Remove && unit.Action != envunit.Purge && unit.Action != envunit.GlobalExtend && unit.Action != envunit.Takeover && unit.Action != envunit.Reuse) {
		return invalid("unsupported terminal change action")
	}
	ledger, err := ParseOrAbsent(c.Prior, unit.Manager)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(ledger.Entries)+1)
	consumed := false
	for _, entry := range ledger.Entries {
		if (unit.Action == envunit.Purge || unit.Action == envunit.Takeover || unit.Action == envunit.Reuse) && unit.RetainedEntry != nil && entry.ID() == unit.RetainedEntry.ID {
			if unit.Action == envunit.Purge && entry.Removed == nil || unit.Action == envunit.Takeover && entry.Global == nil || unit.Action == envunit.Reuse && entry.Purged == nil {
				return invalid("consumption variant differs")
			}
			raw, err := entry.Marshal()
			if err != nil {
				return err
			}
			if unit.RetainedEntry.Digest != envorigin.Digest(raw) {
				return invalid("purge receipt changed")
			}
			unit.RetainedEntry.Serialized = raw
			consumed = true
			continue
		}
		ids = append(ids, entry.ID())
	}
	if (unit.Action == envunit.Purge || unit.Action == envunit.Takeover || unit.Action == envunit.Reuse) && !consumed {
		return invalid("purge receipt absent")
	}
	if unit.Action != envunit.Takeover && unit.Action != envunit.Reuse {
		ids = append(ids, unit.ID)
	}
	digest := "absent"
	if len(c.Prior) != 0 {
		digest = envorigin.Digest(c.Prior)
	}
	unit.LedgerPublication = &envunit.LedgerPublication{Path: Path(unit.Manager.Path), PreimageDigest: digest, Revision: ledger.Revision, NextRevision: ledger.Revision + 1, EntryIDs: ids}
	return nil
}

func ParseOrAbsent(raw []byte, manager envunit.Manager) (*Ledger, error) {
	if len(raw) == 0 {
		return &Ledger{Record: Version, Manager: manager, Entries: []Entry{}}, nil
	}
	return Parse(raw, manager)
}

func (c *Change) Build(unit *envunit.Unit, preparedSources map[string][]byte) ([]byte, error) {
	if unit == nil || (unit.Action != envunit.Remove && unit.Action != envunit.Purge && unit.Action != envunit.GlobalExtend && unit.Action != envunit.Takeover && unit.Action != envunit.Reuse) {
		return nil, invalid("unsupported terminal change action")
	}
	if err := unit.Validate(); err != nil {
		return nil, err
	}
	ledger, err := ParseOrAbsent(c.Prior, unit.Manager)
	if err != nil {
		return nil, err
	}
	if unit.Action == envunit.Takeover || unit.Action == envunit.Reuse {
		next := []Entry{}
		for _, entry := range ledger.Entries {
			if entry.ID() != unit.RetainedEntry.ID {
				next = append(next, entry)
			}
		}
		return PlanReplace(c.Prior, unit.Manager, ledger.Revision, next, unit)
	}
	if unit.Action == envunit.Purge {
		var next []Entry
		var removed *RemovedRetained
		for _, entry := range ledger.Entries {
			if entry.ID() == unit.RetainedEntry.ID {
				removed = entry.Removed
				continue
			}
			next = append(next, entry)
		}
		if removed == nil {
			return nil, invalid("purge missing complete removal evidence")
		}
		raw, err := unit.Marshal()
		if err != nil {
			return nil, err
		}
		roots := make([]string, 0, len(unit.Homes))
		absence := append([]string{}, removed.ExpectedAbsence...)
		for _, home := range unit.Homes {
			roots = append(roots, home.Root)
			absence = append(absence, home.Root)
		}
		entry := Entry{Purged: &Purged{ID: unit.ID, Kind: "purged", Phase: "complete", Manager: unit.Manager, PriorRemovalID: removed.ID, ProfileName: unit.Profiles[0].Name, TargetNamespaces: roots, RemovalCommit: removed.RemovalCommit, PurgeCommit: Commit{Unit: unit.ID, Evidence: envorigin.Digest(raw)}, ExpectedAbsence: absence}}
		next = append(next, entry)
		return PlanReplace(c.Prior, unit.Manager, ledger.Revision, next, unit)
	}
	if unit.Action == envunit.GlobalExtend {
		if c.Conflict == nil {
			return nil, invalid("global conflict capture missing")
		}
		conflict := *c.Conflict
		preimages, err := json.Marshal(unit.Targets)
		if err != nil {
			return nil, err
		}
		conflict.UnitPreimages = envorigin.Digest(preimages)
		records := []OwnedRecord{}
		closure := []string{}
		seen := map[string]bool{}
		for _, p := range unit.Profiles {
			records = append(records, OwnedRecord{Path: filepath.Join(unit.Manager.Path, "profiles", p.Name, "lock.json"), Image: p.NewLock})
			for _, ref := range p.Closure {
				if !seen[ref] {
					closure = append(closure, ref)
					seen[ref] = true
				}
			}
		}
		for _, p := range unit.Profiles {
			path := filepath.Join(unit.Manager.Path, "profiles", p.Name, "source.json")
			raw, ok := preparedSources[path]
			if !ok || envorigin.Digest(raw) != p.Source {
				return nil, invalid("global source bytes differ from bound prepared publication")
			}
			records = append(records, OwnedRecord{Path: path, Image: envunit.RecordImage{Kind: "file", Bytes: raw, Digest: p.Source}})
			// L0 remains custody for unchanged homes while L1 is retained.
			// Pin its immutable closure as well as the desired L1 dependency set.
			if p.OldLock.Kind == "file" {
				old, err := contextlock.Parse(p.OldLock.Bytes)
				if err != nil {
					return nil, err
				}
				for _, member := range old.Members {
					raw, err := json.Marshal(member)
					if err != nil {
						return nil, err
					}
					ref := envorigin.Digest(raw)
					if !seen[ref] {
						closure = append(closure, ref)
						seen[ref] = true
					}
				}
			}
		}
		entry := Entry{Global: &GlobalRetained{ID: unit.ID, Kind: "global-retained", Phase: "retained_conflict", CreatedByUnit: *unit, ProfileTransitions: unit.Profiles, Homes: unit.Homes, BoundaryProof: unit.BoundaryProof, ClosureRefs: closure, RetainedRecords: records, MutablePreimages: unit.Targets, ConflictReceipt: conflict, ContinuationCut: "profile_in_place_conflict"}}
		if err := entry.Seal(); err != nil {
			return nil, err
		}
		next := append(append([]Entry{}, ledger.Entries...), entry)
		return PlanReplace(c.Prior, unit.Manager, ledger.Revision, next, unit)
	}
	absence := make([]string, 0, len(unit.Profiles))
	for _, profile := range unit.Profiles {
		absence = append(absence, filepath.Join(unit.Manager.Path, "profiles", profile.Name))
	}
	// Commit evidence is a digest of the noncredential bound unit. The ledger
	// postimage is intentionally excluded, avoiding a recursive receipt hash.
	raw, err := unit.Marshal()
	if err != nil {
		return nil, err
	}
	revocations := c.Revocations
	if revocations == nil {
		revocations = []Revocation{}
	}
	entry := Entry{Removed: &RemovedRetained{ID: unit.ID, Kind: "removed-retained", Phase: "retained", CreatedByUnit: *unit, ProfileTransitions: unit.Profiles, Homes: unit.Homes, BoundaryProof: unit.BoundaryProof, RemovedRecords: c.RemovedRecords, ExpectedAbsence: absence, ReferenceRevocations: revocations, RemovalCommit: Commit{Unit: unit.ID, Evidence: envorigin.Digest(raw)}}}
	if err := entry.Seal(); err != nil {
		return nil, err
	}
	next := append(append([]Entry{}, ledger.Entries...), entry)
	return PlanReplace(c.Prior, unit.Manager, ledger.Revision, next, unit)
}
