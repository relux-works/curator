package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envunit"
)

func buildOpaque(plan Plan) ([]OpaqueRecord, error) {
	if len(plan.Opaque) == 0 {
		return nil, nil
	}
	if plan.Unit == nil {
		return nil, journalf("opaque action requires typed custody")
	}
	targets := append([]OpaqueTarget(nil), plan.Opaque...)
	sort.Slice(targets, func(i, j int) bool { return targets[i].LivePath < targets[j].LivePath })
	var result []OpaqueRecord
	for index, target := range targets {
		if target.Class == envunit.BackupQuarantine {
			adapter, ok := plan.Unit.BackupAdapter(target.LivePath)
			if !ok {
				return nil, journalf("opaque backup outside enumerated home")
			}
			if err := envunit.CheckBackupGeneration(target.LivePath, adapter, plan.Unit.BackupManaged(target.LivePath)); err != nil {
				return nil, err
			}
		} else if target.Class != "" || plan.Unit.Action != envunit.Purge {
			return nil, journalf("unknown opaque custody")
		}
		before, err := envunit.ObserveDirectory(target.LivePath)
		if err != nil {
			return nil, err
		}
		if before != target.Before || before.Kind != "directory" {
			return nil, corruptionf("opaque home preimage changed")
		}
		quarantine := envjournal.OpaquePath(target.LivePath, plan.TransactionID, index)
		state, err := envunit.ObserveDirectory(quarantine)
		if err != nil {
			return nil, err
		}
		if state.Kind != "absent" {
			return nil, corruptionf("opaque quarantine slot is occupied")
		}
		result = append(result, OpaqueRecord{Class: target.Class, LivePath: target.LivePath, QuarantinePath: quarantine, Before: before, State: StatePending})
	}
	return result, nil
}

func (engine *Engine) opaqueEvent(j *Journal, point Point, path string) error {
	event := Event{Point: point, TransactionID: j.TransactionID, TargetIndex: -1, Class: "quarantine", Identifier: path, LivePath: path}
	if engine.hooks.Observe != nil {
		engine.hooks.Observe(event)
	}
	if engine.hooks.Fault != nil {
		return engine.hooks.Fault(event)
	}
	return nil
}

func (engine *Engine) commitOpaque(j *Journal) error {
	for index := range j.Opaque {
		if err := engine.checkCommit(j); err != nil {
			return err
		}
		action := &j.Opaque[index]
		if action.State == StateCommitted {
			continue
		}
		live, err := envunit.ObserveDirectory(action.LivePath)
		if err != nil {
			return err
		}
		if live == action.Before {
			if err := engine.renameNoReplace(action.LivePath, action.QuarantinePath); err != nil {
				return err
			}
			if err := syncDirectory(filepath.Dir(action.LivePath)); err != nil {
				return err
			}
			if err := engine.opaqueEvent(j, PointAfterQuarantine, action.LivePath); err != nil {
				return err
			}
		}
		action.State = StateCommitted
		delete(engine.nsCache, j.TransactionID)
		if err := engine.saveJournal(j); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) rollbackOpaque(j *Journal) error {
	for index := len(j.Opaque) - 1; index >= 0; index-- {
		if err := engine.checkOwnership(); err != nil {
			return err
		}
		if err := envjournal.VerifyOpaqueState(j); err != nil {
			return err
		}
		action := &j.Opaque[index]
		live, err := envunit.ObserveDirectory(action.LivePath)
		if err != nil {
			return err
		}
		if live.Kind == "absent" {
			if err := engine.renameNoReplace(action.QuarantinePath, action.LivePath); err != nil {
				return err
			}
			if err := syncDirectory(filepath.Dir(action.LivePath)); err != nil {
				return err
			}
			if err := engine.opaqueEvent(j, PointAfterQuarantineRestore, action.LivePath); err != nil {
				return err
			}
		}
		action.State = StateRolledBack
		delete(engine.nsCache, j.TransactionID)
		if err := engine.saveJournal(j); err != nil {
			return err
		}
	}
	return nil
}

// removeOpaque traverses only names/metadata inside the exact quarantined inode.
// It never opens, reads, hashes, copies or follows a file/link payload. Each
// recursive directory is pinned before removing any child and rechecked after.
func (engine *Engine) removeOpaque(j *Journal, action OpaqueRecord, path string) error {
	if err := engine.checkOwnership(); err != nil {
		return err
	}
	root, err := envunit.ObserveDirectory(action.QuarantinePath)
	if err != nil {
		return err
	}
	if root != action.Before {
		return corruptionf("opaque cleanup root was substituted")
	}
	before, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if before.IsDir() && before.Mode()&os.ModeSymlink == 0 {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			current, err := os.Lstat(path)
			if err != nil || !os.SameFile(before, current) {
				return corruptionf("opaque cleanup parent changed")
			}
			if err := engine.removeOpaque(j, action, filepath.Join(path, entry.Name())); err != nil {
				return err
			}
		}
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) {
		return corruptionf("opaque cleanup entry changed")
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	return engine.opaqueEvent(j, PointDuringQuarantineRemoval, action.LivePath)
}

func (engine *Engine) cleanupOpaque(j *Journal) error {
	if j.Phase != PhaseCleanup && len(j.Opaque) != 0 {
		return fmt.Errorf("opaque cleanup outside committed phase")
	}
	if err := envjournal.VerifyOpaqueState(j); err != nil {
		return err
	}
	for _, action := range j.Opaque {
		if envjournal.GlobalRetained(j) {
			continue
		}
		state, err := envunit.ObserveDirectory(action.QuarantinePath)
		if err != nil {
			return err
		}
		if state.Kind == "absent" {
			continue
		}
		if err := engine.removeOpaque(j, action, action.QuarantinePath); err != nil {
			return err
		}
	}
	return nil
}
