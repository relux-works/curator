package envjournal

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
)

func verifyLifecycleTarget(j *Journal, target TargetRecord) error {
	publication := j.Unit.LedgerPublication
	if publication == nil {
		return unitError("lifecycle target lacks typed CAS")
	}
	read := func(path string) ([]byte, bool, error) {
		return envorigin.ReadPrivateAuthority(j.Unit.Manager.Path, path, envlifecycle.MaxBytes)
	}
	before := func(path string) (bool, error) {
		raw, present, err := read(path)
		if err != nil {
			return false, err
		}
		if publication.PreimageDigest == "absent" {
			return !present, nil
		}
		return present && envorigin.Digest(raw) == publication.PreimageDigest, nil
	}
	after := func(path string) (bool, error) {
		raw, present, err := read(path)
		if err != nil || !present {
			return false, err
		}
		ledger, err := envlifecycle.Parse(raw, j.Unit.Manager)
		if err != nil {
			return false, err
		}
		if ledger.Revision != publication.NextRevision {
			return false, nil
		}
		ids := make([]string, 0, len(ledger.Entries))
		found := false
		for _, entry := range ledger.Entries {
			ids = append(ids, entry.ID())
			if entry.ID() != j.Unit.ID {
				continue
			}
			if j.Unit.Action == "remove-purge" {
				if entry.Purged == nil {
					return false, unitError("purge lacks terminal variant")
				}
				actual, _ := j.Unit.Marshal()
				if entry.Purged.PurgeCommit.Evidence != envorigin.Digest(actual) {
					return false, unitError("terminal purge evidence differs from real unit")
				}
				found = true
				continue
			}
			if j.Unit.Action == "global-extend" {
				if entry.Global == nil || j.Global == nil {
					return false, unitError("global receipt variant absent")
				}
				captured, _ := json.Marshal(entry.Global.CreatedByUnit)
				actual, _ := json.Marshal(j.Unit)
				conflict, _ := json.Marshal(entry.Global.ConflictReceipt)
				intended, _ := json.Marshal(j.Global.Conflict)
				if !bytes.Equal(captured, actual) || !bytes.Equal(conflict, intended) {
					return false, unitError("global receipt differs from actual unit/conflict")
				}
				found = true
				continue
			}
			if entry.Removed == nil {
				return false, unitError("terminal variant differs from admitted action")
			}
			captured, _ := json.Marshal(entry.Removed.CreatedByUnit)
			actual, _ := json.Marshal(j.Unit)
			if !bytes.Equal(captured, actual) {
				return false, unitError("terminal receipt differs from complete real unit")
			}
			found = true
		}
		if j.Unit.Action == "retained-takeover" || j.Unit.Action == "purged-name-reuse" {
			// Exact receipt consumption and unchanged L1/complete homes are checked by
			// the same closed semantic validator independently of an old ledger file.
			if err := envlifecycle.ValidateConsumption(j.Unit, ledger); err != nil {
				return false, err
			}
			found = true
		}
		want := append([]string{}, publication.EntryIDs...)
		sort.Strings(ids)
		sort.Strings(want)
		a, _ := json.Marshal(ids)
		b, _ := json.Marshal(want)
		return found && bytes.Equal(a, b), nil
	}
	absent := func(path string) (bool, error) { _, present, err := read(path); return !present, err }
	liveBefore, err := before(target.LivePath)
	if err != nil {
		return err
	}
	liveAfter, err := after(target.LivePath)
	if err != nil {
		return err
	}
	liveAbsent, err := absent(target.LivePath)
	if err != nil {
		return err
	}
	backupBefore, err := before(target.BackupPath)
	if err != nil {
		return err
	}
	backupAbsent, err := absent(target.BackupPath)
	if err != nil {
		return err
	}
	moved := backupBefore && publication.PreimageDigest != "absent" || backupAbsent && publication.PreimageDigest == "absent"
	valid := false
	switch j.Phase {
	case PhasePreparing, PhasePrepared:
		valid = target.State == StatePending && liveBefore && backupAbsent
	case PhaseCommitting:
		switch target.State {
		case StatePending:
			valid = liveBefore && backupAbsent || moved && (liveAbsent || liveAfter)
		case StateBackedUp:
			valid = moved && (liveAbsent || liveAfter)
		case StateCommitted:
			valid = liveAfter && moved
		}
	case PhaseCleanup:
		valid = target.State == StateCommitted && liveAfter && (moved || backupAbsent)
	case PhaseRollingBack:
		valid = liveBefore && backupAbsent || moved && (liveAbsent || liveAfter)
	case PhaseRollbackCleanup:
		valid = target.State == StateRolledBack && liveBefore && backupAbsent
	}
	if !valid {
		return unitError("unreachable lifecycle CAS state")
	}
	if j.Phase == PhasePrepared || j.Phase == PhaseCommitting {
		missing, err := absent(target.StagedPath)
		if err != nil {
			return err
		}
		if !missing {
			valid, err := after(target.StagedPath)
			if err != nil {
				return err
			}
			if !valid {
				return unitError("staged terminal ledger differs")
			}
		}
	}
	return nil
}
