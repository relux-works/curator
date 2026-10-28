package envjournal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
)

func OpaquePath(live, id string, index int) string {
	return filepath.Join(filepath.Dir(live), fmt.Sprintf(".curator-quarantine-%s-%03d", entrydigest.ShortIdentity(id), index))
}

func OpaqueClass(action OpaqueRecord) string {
	if action.Class == "" {
		return "quarantine"
	}
	return action.Class
}

func ValidateOpaque(j *Journal) error {
	if len(j.Opaque) == 0 {
		if j.Unit != nil && j.Unit.Action == envunit.Purge && len(j.Unit.Homes) != 0 {
			return unitError("purge lacks opaque complete home set")
		}
		return nil
	}
	if j.Unit == nil {
		return unitError("opaque action lacks typed custody")
	}
	if j.Unit.Action == envunit.Purge && len(j.Opaque) != len(j.Unit.Homes) {
		return unitError("opaque purge home inventory differs")
	}
	byRoot := map[string]envunit.Home{}
	for _, home := range j.Unit.Homes {
		byRoot[home.Root] = home
	}
	paths := make([]string, 0, len(j.Opaque))
	for index, action := range j.Opaque {
		if index > 0 && j.Opaque[index-1].LivePath >= action.LivePath {
			return unitError("opaque inventory is not canonical")
		}
		switch OpaqueClass(action) {
		case "quarantine":
			home, ok := byRoot[action.LivePath]
			if j.Unit.Action != envunit.Purge || !ok || action.Before != home.DirectoryBefore || home.DirectoryAfter.Kind != "absent" {
				return unitError("opaque home identity/namespace differs")
			}
		case envunit.BackupQuarantine:
			if _, ok := j.Unit.BackupAdapter(action.LivePath); !ok {
				return unitError("opaque backup generation binding differs")
			}
		default:
			return unitError("unknown opaque class")
		}
		if action.Before.Kind != "directory" || action.QuarantinePath != OpaquePath(action.LivePath, j.TransactionID, index) {
			return unitError("opaque identity/namespace differs")
		}
		if action.State != StatePending && action.State != StateCommitted && action.State != StateRolledBack {
			return unitError("opaque action state")
		}
		if (j.Phase == PhasePreparing || j.Phase == PhasePrepared) && action.State != StatePending {
			return unitError("opaque progress before commit")
		}
		paths = append(paths, action.LivePath, action.QuarantinePath)
	}
	for _, path := range paths {
		for _, t := range j.Targets {
			for _, other := range []string{t.LivePath, t.BackupPath, t.StagedPath, t.RollbackPath} {
				if other != "" && (path == other || strings.HasPrefix(path, other+string(filepath.Separator)) || strings.HasPrefix(other, path+string(filepath.Separator))) {
					return unitError("opaque namespace overlaps file publication")
				}
			}
		}
		for _, d := range j.Directories {
			if d.LivePath == path || strings.HasPrefix(d.LivePath, path+string(filepath.Separator)) {
				return unitError("opaque namespace overlaps directory publication")
			}
		}
		for input := range j.ReadInputs {
			if input == path || strings.HasPrefix(input, path+string(filepath.Separator)) || strings.HasPrefix(path, input+string(filepath.Separator)) {
				return unitError("opaque namespace overlaps frozen input")
			}
		}
	}
	sort.Strings(paths)
	for i := 1; i < len(paths); i++ {
		if paths[i] == paths[i-1] || strings.HasPrefix(paths[i], paths[i-1]+string(filepath.Separator)) {
			return unitError("duplicate opaque namespace")
		}
	}
	return nil
}

func VerifyPurgeReceipt(unit *envunit.Unit) error {
	if unit.Action != envunit.Purge {
		return nil
	}
	retained := unit.RetainedEntry
	if retained == nil || len(retained.Serialized) == 0 || envorigin.Digest(retained.Serialized) != retained.Digest {
		return unitError("purge lacks exact durable removal receipt")
	}
	entry, err := envlifecycle.ParseEntry(retained.Serialized, unit.Manager)
	if err != nil {
		return err
	}
	if entry.ID() != retained.ID || entry.Removed == nil {
		return unitError("purge receipt kind/identity differs")
	}
	a, _ := json.Marshal(entry.Removed.ProfileTransitions)
	b, _ := json.Marshal(unit.Profiles)
	if !bytes.Equal(a, b) || len(entry.Removed.Homes) != len(unit.Homes) {
		return unitError("purge differs from complete removal inventory")
	}
	for index, home := range unit.Homes {
		prior := entry.Removed.Homes[index]
		a, _ := json.Marshal(home.Before)
		b, _ := json.Marshal(prior.After)
		if home.Profile != prior.Profile || home.Adapter != prior.Adapter || home.Target != prior.Target || home.Root != prior.Root || home.DirectoryBefore != prior.DirectoryAfter || !bytes.Equal(a, b) || home.After.Kind != "absent" || home.DirectoryAfter.Kind != "absent" {
			return unitError("purge recaptured or substituted retained home")
		}
	}
	return nil
}

func VerifyOpaqueState(j *Journal) error {
	if err := ValidateOpaque(j); err != nil {
		return err
	}
	if len(j.Opaque) == 0 {
		return nil
	}
	if err := VerifyPurgeReceipt(j.Unit); err != nil {
		return err
	}
	for _, action := range j.Opaque {
		live, err := envunit.ObserveDirectory(action.LivePath)
		if err != nil {
			return err
		}
		quarantine, err := envunit.ObserveDirectory(action.QuarantinePath)
		if err != nil {
			return err
		}
		before := live == action.Before && quarantine.Kind == "absent"
		moved := live.Kind == "absent" && quarantine == action.Before
		clean := live.Kind == "absent" && quarantine.Kind == "absent"
		valid := false
		switch j.Phase {
		case PhasePreparing, PhasePrepared:
			valid = action.State == StatePending && before
		case PhaseCommitting:
			valid = action.State == StatePending && (before || moved) || action.State == StateCommitted && moved
		case PhaseCleanup:
			valid = action.State == StateCommitted && (moved || clean)
			if GlobalRetained(j) {
				valid = action.State == StatePending && before
			}
		case PhaseRollingBack:
			valid = (action.State == StatePending || action.State == StateCommitted) && (before || moved) || action.State == StateRolledBack && before
		case PhaseRollbackCleanup:
			valid = action.State == StateRolledBack && before
		}
		if !valid {
			return unitError("unreachable opaque home/quarantine state")
		}
		if OpaqueClass(action) == envunit.BackupQuarantine && !clean {
			root := action.LivePath
			if moved {
				root = action.QuarantinePath
			}
			adapter, _ := j.Unit.BackupAdapter(action.LivePath)
			if err := envunit.CheckBackupGeneration(root, adapter, j.Unit.BackupManaged(action.LivePath)); err != nil {
				return err
			}
		}
		if err := staging.RecheckDurableOne(staging.Target{LivePath: action.LivePath, Kind: staging.KindEntry}, j.Unit.BoundaryProof); err != nil {
			return err
		}
	}
	return nil
}
