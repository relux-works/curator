package envauthority

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
)

type lifecycleState struct {
	ledger   *envlifecycle.Ledger
	retained map[string]bool
}

// inspectLifecycle observes only the fixed private receipt, installed absence,
// home directory identity and exact marker/origin bytes. It never visits a
// native account/root/backend or retained session/credential payload.
func inspectLifecycle(manager string, pending *Pending) (*lifecycleState, error) {
	state := &lifecycleState{retained: map[string]bool{}}
	if _, err := os.Lstat(manager); os.IsNotExist(err) {
		return state, nil
	} else if err != nil {
		return nil, err
	}
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return nil, err
	}
	ledger, raw, err := envlifecycle.Read(identity)
	if err != nil {
		return nil, err
	}
	state.ledger = ledger
	if len(raw) == 0 {
		return state, nil
	}
	for _, entry := range ledger.Entries {
		consumedPending := false
		for _, unit := range pending.Units {
			if (unit.Action != envunit.Purge && unit.Action != envunit.Takeover && unit.Action != envunit.Reuse) || unit.RetainedEntry == nil || unit.RetainedEntry.ID != entry.ID() {
				continue
			}
			raw, err := entry.Marshal()
			if err != nil {
				return nil, err
			}
			if unit.RetainedEntry.Digest != envorigin.Digest(raw) || !bytes.Equal(unit.RetainedEntry.Serialized, raw) {
				return nil, fmt.Errorf("%s: pending purge receipt differs", envorigin.DiagInvalid)
			}
			consumedPending = true
		}
		if consumedPending {
			continue
		}
		if unit, ok := pending.Units[entry.ID()]; ok {
			if entry.Purged != nil && unit.Action == envunit.Purge {
				raw, err := unit.Marshal()
				if err != nil {
					return nil, err
				}
				if entry.Purged.PurgeCommit.Evidence != envorigin.Digest(raw) || unit.RetainedEntry == nil || entry.Purged.PriorRemovalID != unit.RetainedEntry.ID {
					return nil, fmt.Errorf("%s: pending terminal purge differs", envorigin.DiagInvalid)
				}
				continue
			}
			var creating *envunit.Unit
			if entry.Removed != nil {
				creating = &entry.Removed.CreatedByUnit
			}
			if entry.Global != nil {
				creating = &entry.Global.CreatedByUnit
			}
			a, _ := json.Marshal(creating)
			b, _ := json.Marshal(unit)
			if creating == nil || !bytes.Equal(a, b) {
				return nil, fmt.Errorf("%s: lifecycle receipt differs from real pending unit", envorigin.DiagInvalid)
			}
			// The independent journal reader already checked the exact phase,
			// lock/preimages and all target state. This is no stable waiver.
			continue
		}
		if entry.Purged != nil {
			only := &envlifecycle.Ledger{Record: envlifecycle.Version, Manager: identity, Revision: ledger.Revision, Entries: []envlifecycle.Entry{entry}}
			if err := only.VerifyAbsence(); err != nil {
				return nil, err
			}
			root := filepath.Join(manager, "environments", entry.Purged.ProfileName)
			listing, err := os.ReadDir(root)
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			for _, child := range listing {
				if !pending.Artifacts[filepath.Join(root, child.Name())] {
					return nil, fmt.Errorf("%s: purged namespace has an unexpected entry", envorigin.DiagInvalid)
				}
			}
			continue
		}
		if entry.Global != nil {
			if err := envjournal.VerifyRetainedGlobal(entry.Global); err != nil {
				return nil, err
			}
			for _, home := range entry.Global.Homes {
				state.retained[home.Target] = true
			}
			continue
		}
		removed := entry.Removed
		if removed == nil {
			return nil, fmt.Errorf("%s: unknown lifecycle state", envorigin.DiagInvalid)
		}
		only := &envlifecycle.Ledger{Record: envlifecycle.Version, Manager: identity, Revision: ledger.Revision, Entries: []envlifecycle.Entry{entry}}
		if err := only.VerifyAbsence(); err != nil {
			return nil, err
		}
		if err := retainedNamespaces(manager, removed); err != nil {
			return nil, err
		}
		for _, home := range removed.Homes {
			directory, err := envunit.ObserveDirectory(home.Root)
			if err != nil {
				return nil, err
			}
			if directory != home.DirectoryAfter {
				return nil, fmt.Errorf("%s: retained home directory identity changed", envorigin.DiagInvalid)
			}
			marker, err := stateread.ReadRegularFile(filepath.Join(home.Target, envmarker.Name))
			if err != nil {
				return nil, err
			}
			origin, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(home.Target, envorigin.Name), envorigin.MaxRecordBytes)
			if err != nil {
				return nil, err
			}
			if marker.Kind != stateread.KindPresent || !bytes.Equal(marker.Bytes, home.After.Marker) || !present || !bytes.Equal(origin, home.After.Origin) {
				return nil, fmt.Errorf("%s: retained marker/origin was substituted", envorigin.DiagInvalid)
			}
			state.retained[home.Target] = true
		}
	}
	return state, nil
}

func retainedNamespaces(manager string, removed *envlifecycle.RemovedRetained) error {
	byProfile := map[string]map[string]bool{}
	for _, home := range removed.Homes {
		if byProfile[home.Profile] == nil {
			byProfile[home.Profile] = map[string]bool{}
		}
		byProfile[home.Profile][home.Adapter] = true
	}
	for profile, known := range byProfile {
		entries, err := os.ReadDir(filepath.Join(manager, "environments", profile))
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !known[entry.Name()] || !entry.IsDir() {
				return fmt.Errorf("%s: retained namespace inventory changed", envorigin.DiagInvalid)
			}
		}
	}
	return nil
}

// SelectedLifecycle is an early launch/capture barrier. Intentional custody is
// not an installed binding, discovery input, repair permission or cache result.
func SelectedLifecycle(manager, profile string) error {
	if _, err := ReadCompletedLegacyMigration(manager); err != nil {
		return err
	}
	pending, err := PendingUnits(manager)
	if err != nil {
		return err
	}
	state, err := inspectLifecycle(manager, pending)
	if err != nil {
		return err
	}
	if state.ledger == nil {
		return nil
	}
	for _, entry := range state.ledger.Entries {
		for _, name := range entry.Profiles() {
			if name != profile {
				continue
			}
			if entry.Removed != nil {
				return fmt.Errorf("%s: retained_profile_removed; explicit profile remove --purge is required for later deletion", envorigin.DiagInvalid)
			}
			if entry.Purged != nil {
				return fmt.Errorf("%s: purged namespace requires a new genuine provisioning unit before activation", envorigin.DiagInvalid)
			}
			if entry.Global != nil {
				return fmt.Errorf("environment_materialize_recovery_required: global_retained_conflict; explicit takeover continuation required")
			}
		}
	}
	return nil
}
