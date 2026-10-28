package envauthority

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
)

func overlaps(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	return a == b || strings.HasPrefix(a, b+string(filepath.Separator)) || strings.HasPrefix(b, a+string(filepath.Separator))
}

// AdmitJournal checks the actual executor effect set, independently of the
// caller's selection. A valid receipt classifies custody, not arbitrary writes.
// Invoke before the first journal publication and again before forward work.
// Rollback/committed cleanup instead use the exact reachable journal proof.
func AdmitJournal(manager string, j *envjournal.Journal) error {
	leasePath := filepath.Join(manager, "state", "project-leases", "v1", "ledger.json")
	for _, target := range j.Targets {
		if overlaps(target.LivePath, leasePath) {
			removal := j.Unit != nil && j.Unit.Action == envunit.Remove && target.LivePath == leasePath && target.Class == "reference-record"
			if !removal && (j.ProjectIdentity != "project-leases" || target.Class != "project-lease" || target.LivePath != leasePath) {
				return fmt.Errorf("%s: project lease references require their typed publication owner", envorigin.DiagInvalid)
			}
		}
	}
	paths := []string{}
	for _, target := range j.Targets {
		paths = append(paths, target.LivePath)
	}
	for _, dir := range j.Directories {
		paths = append(paths, dir.LivePath)
	}
	for _, opaque := range j.Opaque {
		paths = append(paths, opaque.LivePath)
	}
	if err := AdmitEffects(manager, j.Unit, paths); err != nil {
		return err
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return err
	}
	ledger, raw, err := envlifecycle.Read(identity)
	if err != nil {
		return err
	}
	if err := admitDesiredLeaseReferences(manager, j, ledger); err != nil {
		return err
	}
	for _, target := range j.Targets {
		if target.LivePath == envlifecycle.Path(manager) {
			if j.Unit == nil || target.Class != "lifecycle" || j.Unit.LedgerPublication == nil || j.Unit.LedgerPublication.Path != target.LivePath {
				return fmt.Errorf("%s: lifecycle ledger requires a structured unit CAS", envorigin.DiagInvalid)
			}
		}
	}
	if len(raw) == 0 {
		return nil
	}
	for _, entry := range ledger.Entries {
		admitted, err := consumesOrCreates(entry, j.Unit)
		if err != nil {
			return err
		}
		if admitted {
			continue
		}
		protected := entry.Targets()
		for _, name := range entry.Profiles() {
			protected = append(protected, filepath.Join(manager, "profiles", name), filepath.Join(manager, "environments", name))
		}
		paths := []string{}
		for _, target := range j.Targets {
			// The separately validated structured ledger CAS preserves unrelated
			// entries; its physical shared-container path is not an overlap waiver.
			if target.Class == "lifecycle" {
				continue
			}
			paths = append(paths, target.LivePath)
			if target.LivePath == filepath.Join(manager, "profiles", "current") || strings.HasPrefix(target.LivePath, filepath.Join(manager, "profiles", "scoped")+string(filepath.Separator)) {
				source := target.StagedPath
				if j.Phase == envjournal.PhasePreparing {
					source = target.StagedSource
				}
				if source != "" {
					record, e := stateread.ReadRegularFile(source)
					if e != nil {
						return e
					}
					for _, name := range entry.Profiles() {
						if record.Kind == stateread.KindPresent && strings.TrimSpace(string(record.Bytes)) == name {
							return fmt.Errorf("%s: lifecycle binding cannot be activated", envorigin.DiagInvalid)
						}
					}
				}
			}
		}
		for _, dir := range j.Directories {
			paths = append(paths, dir.LivePath)
		}
		for _, opaque := range j.Opaque {
			paths = append(paths, opaque.LivePath)
		}
		for _, path := range paths {
			for _, root := range protected {
				if overlaps(path, root) {
					return fmt.Errorf("%s: effect intersects protected lifecycle namespace", envorigin.DiagInvalid)
				}
			}
		}
	}
	return nil
}

func consumesOrCreates(entry envlifecycle.Entry, unit *envunit.Unit) (bool, error) {
	if unit == nil {
		return false, nil
	}
	if entry.ID() == unit.ID {
		var creating *envunit.Unit
		if entry.Removed != nil {
			creating = &entry.Removed.CreatedByUnit
		}
		if entry.Global != nil {
			creating = &entry.Global.CreatedByUnit
		}
		a, _ := json.Marshal(creating)
		b, _ := json.Marshal(unit)
		if creating != nil && bytes.Equal(a, b) {
			return true, nil
		}
		if entry.Purged != nil && unit.Action == envunit.Purge {
			raw, err := unit.Marshal()
			if err != nil {
				return false, err
			}
			if entry.Purged.PurgeCommit.Evidence == envorigin.Digest(raw) {
				return true, nil
			}
		}
	}
	if unit.RetainedEntry == nil || unit.RetainedEntry.ID != entry.ID() {
		return false, nil
	}
	raw, err := entry.Marshal()
	if err != nil {
		return false, err
	}
	if unit.RetainedEntry.Digest != envorigin.Digest(raw) || !bytes.Equal(unit.RetainedEntry.Serialized, raw) {
		return false, fmt.Errorf("%s: consumption receipt binding changed", envorigin.DiagInvalid)
	}
	if unit.LedgerPublication == nil || !((entry.Removed != nil && unit.Action == envunit.Purge) || (entry.Global != nil && unit.Action == envunit.Takeover) || (entry.Purged != nil && unit.Action == envunit.Reuse)) {
		return false, fmt.Errorf("%s: lifecycle action cannot consume receipt", envorigin.DiagInvalid)
	}
	if err := unit.Validate(); err != nil {
		return false, err
	}
	// Full action-specific old/new inventory is independently checked by the
	// ledger builder/reader, not by a caller-provided boolean admission verdict.
	return true, nil
}
