package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Engine owns journals beneath one manager home. It does not acquire locks;
// every mutating or journal-reading entry point requires the caller's witness.
type Engine struct {
	mu               sync.Mutex
	journalRoot      string
	hooks            Hooks
	syncStagedParent func(string) error
}

// New constructs an engine without creating manager-home state.
func New(home string, options ...Option) (*Engine, error) {
	if home == "" {
		return nil, fmt.Errorf("manager home is empty")
	}
	absolute, err := filepath.Abs(home)
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		journalRoot:      filepath.Join(filepath.Clean(absolute), "state", "transactions", "v1"),
		syncStagedParent: syncDirectory,
	}
	for _, option := range options {
		if option != nil {
			option(engine)
		}
	}
	return engine, nil
}

// Prepare creates the durable journal before copying deterministic sibling
// staging targets. Live targets are not changed. Preparation is kept under the
// home lock so one transaction id cannot race another process.
func (engine *Engine) Prepare(lock HomeLock, plan Plan) (*Journal, error) {
	if err := requireHomeLock(lock); err != nil {
		return nil, err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()

	journal, _, err := engine.buildJournal(plan)
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); err == nil {
		return nil, fmt.Errorf("transaction %s already exists", journal.TransactionID)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := engine.saveJournal(journal); err != nil {
		return nil, err
	}
	prepared := false
	defer func() {
		if !prepared {
			_ = engine.removePreparedSidecars(journal)
		}
	}()
	for index := range journal.Targets {
		target := &journal.Targets[index]
		if target.StagedPath == "" {
			continue
		}
		if err := engine.stageTarget(journal, index); err != nil {
			return nil, engine.abortPreparation(journal, err)
		}
		got, err := DigestPath(target.StagedPath)
		if err != nil || got != target.DesiredDigest {
			if err == nil {
				err = corruptionf("staged target %s/%s digest %s, want %s", target.Class, target.Identifier, got, target.DesiredDigest)
			}
			return nil, engine.abortPreparation(journal, err)
		}
		if err := engine.syncStagedParent(filepath.Dir(target.StagedPath)); err != nil {
			return nil, engine.abortPreparation(journal, err)
		}
	}
	for index := range journal.Targets {
		journal.Targets[index].StagedSource = ""
		journal.Targets[index].StagingEntries = nil
		journal.Targets[index].StagingIndex = 0
		journal.Targets[index].StagingActive = false
		journal.Targets[index].StagingCreated = false
		journal.Targets[index].StagingBytes = 0
		journal.Targets[index].StagingPrefixDigest = ""
		journal.Targets[index].StagingWriteBytes = 0
		journal.Targets[index].StagingWriteDigest = ""
		journal.Targets[index].StagingDiscarded = false
	}
	journal.Phase = PhasePrepared
	if err := engine.saveJournal(journal); err != nil {
		return nil, engine.abortPreparation(journal, err)
	}
	if err := engine.emit(journal, -1, PointPrepared); err != nil {
		return nil, engine.abortPreparation(journal, err)
	}
	prepared = true
	return cloneJournal(journal), nil
}

func (engine *Engine) buildJournal(plan Plan) (*Journal, []string, error) {
	if !transactionIDPattern.MatchString(plan.TransactionID) {
		return nil, nil, fmt.Errorf("invalid transaction id")
	}
	if plan.ProjectIdentity == "" || !validText(plan.ProjectIdentity) {
		return nil, nil, fmt.Errorf("invalid project identity")
	}
	if len(plan.Targets) == 0 {
		return nil, nil, fmt.Errorf("transaction has no targets")
	}
	keys := append([]string(nil), plan.ReferencedBuildKeys...)
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare([]byte(keys[i]), []byte(keys[j])) < 0 })
	if !sortedUniqueStrings(keys) && len(keys) > 0 {
		return nil, nil, fmt.Errorf("referenced build keys must be nonempty and unique")
	}
	targets := append([]Target(nil), plan.Targets...)
	sort.Slice(targets, func(i, j int) bool {
		if compared := bytes.Compare([]byte(targets[i].Class), []byte(targets[j].Class)); compared != 0 {
			return compared < 0
		}
		return bytes.Compare([]byte(targets[i].Identifier), []byte(targets[j].Identifier)) < 0
	})
	journal := &Journal{
		Schema:              journalSchema,
		TransactionID:       plan.TransactionID,
		ProjectIdentity:     plan.ProjectIdentity,
		Phase:               PhasePreparing,
		ReferencedBuildKeys: keys,
		Targets:             make([]TargetRecord, len(targets)),
	}
	sources := make([]string, len(targets))
	livePaths := make(map[string]struct{}, len(targets))
	for index, target := range targets {
		if target.Class == "" || target.Identifier == "" || !validText(target.Class) || !validText(target.Identifier) {
			return nil, nil, fmt.Errorf("target %d has invalid class or identifier", index)
		}
		if index > 0 && target.Class == targets[index-1].Class && target.Identifier == targets[index-1].Identifier {
			return nil, nil, fmt.Errorf("duplicate transaction target %s/%s", target.Class, target.Identifier)
		}
		if (target.PreimageDigest == "") == (target.ExpectedGeneration == "") {
			return nil, nil, fmt.Errorf("target %s/%s must have exactly one preimage expectation", target.Class, target.Identifier)
		}
		if target.PreimageDigest != "" && !validDigest(target.PreimageDigest) {
			return nil, nil, fmt.Errorf("target %s/%s has invalid preimage digest", target.Class, target.Identifier)
		}
		live, err := filepath.Abs(target.LivePath)
		if err != nil || target.LivePath == "" {
			return nil, nil, fmt.Errorf("target %s/%s live path is invalid", target.Class, target.Identifier)
		}
		live = filepath.Clean(live)
		if _, duplicate := livePaths[live]; duplicate {
			return nil, nil, fmt.Errorf("multiple transaction targets share live path %s", live)
		}
		livePaths[live] = struct{}{}
		if err := validateGenerationPath(target.GenerationPath); err != nil {
			return nil, nil, fmt.Errorf("target %s/%s: %w", target.Class, target.Identifier, err)
		}
		desiredDigest := DigestAbsent
		stagedPath := ""
		stagedSource := ""
		var stagingEntries []RemovalEntry
		if target.StagedSource != "" {
			source, err := filepath.Abs(target.StagedSource)
			if err != nil {
				return nil, nil, err
			}
			source = filepath.Clean(source)
			if source == live {
				return nil, nil, fmt.Errorf("target %s/%s stages from its live path", target.Class, target.Identifier)
			}
			desiredDigest, err = DigestPath(source)
			if err != nil || desiredDigest == DigestAbsent {
				if err == nil {
					err = fmt.Errorf("staged source is absent")
				}
				return nil, nil, fmt.Errorf("target %s/%s: %w", target.Class, target.Identifier, err)
			}
			sources[index] = source
			stagedSource = source
			stagedPath = sidecarPath(live, plan.TransactionID, index, "desired")
			stagingEntries, err = captureRemovalEntries(source, desiredDigest)
			if err != nil {
				return nil, nil, fmt.Errorf("target %s/%s: capture staging manifest: %w", target.Class, target.Identifier, err)
			}
		}
		record := TargetRecord{
			Class:              target.Class,
			Identifier:         target.Identifier,
			LivePath:           live,
			StagedPath:         stagedPath,
			StagedSource:       stagedSource,
			StagingEntries:     stagingEntries,
			PreimageDigest:     target.PreimageDigest,
			ExpectedGeneration: target.ExpectedGeneration,
			GenerationPath:     target.GenerationPath,
			BackupPath:         sidecarPath(live, plan.TransactionID, index, "backup"),
			RollbackPath:       sidecarPath(live, plan.TransactionID, index, "rollback"),
			DesiredDigest:      desiredDigest,
			State:              StatePending,
		}
		for _, sidecar := range []string{record.StagedPath, record.BackupPath, record.RollbackPath} {
			if sidecar == "" {
				continue
			}
			if _, err := os.Lstat(sidecar); err == nil {
				return nil, nil, corruptionf("unreferenced transaction sidecar exists: %s", sidecar)
			} else if !os.IsNotExist(err) {
				return nil, nil, err
			}
		}
		journal.Targets[index] = record
	}
	if err := engine.validateJournal(journal); err != nil {
		return nil, nil, err
	}
	return journal, sources, nil
}

func validateGenerationPath(path string) error {
	if path == "" {
		return nil
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("generation path escapes the live target")
	}
	return nil
}

func (engine *Engine) abortPreparation(journal *Journal, cause error) error {
	if journal.Phase == PhasePreparing {
		if err := engine.discardPreparing(journal); err != nil {
			return errors.Join(cause, err)
		}
		return cause
	}
	journal.Phase = PhaseRollingBack
	if err := engine.saveJournal(journal); err != nil {
		return errors.Join(cause, err)
	}
	if err := engine.rollback(journal); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

// Commit resumes a prepared or interrupted commit by id. Any ordinary commit
// failure durably switches the journal to rollback before restoring targets.
func (engine *Engine) Commit(lock HomeLock, transactionID string) error {
	if err := requireHomeLock(lock); err != nil {
		return err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	journal, err := engine.loadJournal(transactionID)
	if err != nil {
		return err
	}
	return engine.resume(journal)
}

// Recover processes every journal in unsigned transaction-id order. Recovery
// is home-scoped and intentionally does not accept a current project filter.
func (engine *Engine) Recover(lock HomeLock) error {
	if err := requireHomeLock(lock); err != nil {
		return err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	ids, err := engine.journalIDs()
	if err != nil {
		return err
	}
	for _, id := range ids {
		journal, err := engine.loadJournal(id)
		if err != nil {
			return err
		}
		if err := engine.resume(journal); err != nil {
			return fmt.Errorf("recover transaction %s: %w", id, err)
		}
	}
	return nil
}

func (engine *Engine) resume(journal *Journal) error {
	switch journal.Phase {
	case PhasePreparing:
		return engine.discardPreparing(journal)
	case PhasePrepared, PhaseCommitting:
		return engine.commit(journal)
	case PhaseCleanup:
		return engine.cleanupCommitted(journal)
	case PhaseRollingBack, PhaseRollbackCleanup:
		return engine.rollback(journal)
	default:
		return journalf("unsupported recovery phase %q", journal.Phase)
	}
}

func (engine *Engine) commit(journal *Journal) error {
	if journal.Phase == PhasePrepared {
		journal.Phase = PhaseCommitting
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	for index := range journal.Targets {
		if err := engine.commitTarget(journal, index); err != nil {
			journal.Phase = PhaseRollingBack
			if saveErr := engine.saveJournal(journal); saveErr != nil {
				return errors.Join(err, saveErr)
			}
			return errors.Join(err, engine.rollback(journal))
		}
	}
	journal.Phase = PhaseCleanup
	if err := engine.saveJournal(journal); err != nil {
		return err
	}
	if err := engine.emit(journal, -1, PointBeforeCleanup); err != nil {
		return err
	}
	return engine.cleanupCommitted(journal)
}

func (engine *Engine) commitTarget(journal *Journal, index int) error {
	target := &journal.Targets[index]
	switch target.State {
	case StatePending:
		if err := engine.recoverPendingBackup(target); err != nil {
			return err
		}
		if target.State == StateBackedUp {
			// A prior process may have crashed after the durable backup rename
			// but before recording it. Persist the recovered state before the
			// desired swap, exactly like the ordinary backup path below.
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
		}
		if target.State == StatePending {
			if err := engine.emit(journal, index, PointBeforeBackup); err != nil {
				return err
			}
			if err := verifyPreimage(target); err != nil {
				return err
			}
			digest, err := DigestPath(target.LivePath)
			if err != nil {
				return err
			}
			if target.PreimageDigest != "" && digest != target.PreimageDigest {
				return corruptionf("target %s/%s changed before backup", target.Class, target.Identifier)
			}
			if digest != DigestAbsent {
				if err := durableRenameNoReplace(target.LivePath, target.BackupPath); err != nil {
					return err
				}
				if err := syncTree(target.BackupPath); err != nil {
					return err
				}
				captured, err := DigestPath(target.BackupPath)
				if err != nil {
					return err
				}
				if captured != digest {
					if restoreErr := durableRenameNoReplace(target.BackupPath, target.LivePath); restoreErr != nil {
						return errors.Join(corruptionf("target %s/%s changed while backing up", target.Class, target.Identifier), restoreErr)
					}
					return corruptionf("target %s/%s changed while backing up", target.Class, target.Identifier)
				}
				digest = captured
			}
			target.BackupDigest = digest
			if err := engine.emit(journal, index, PointAfterBackup); err != nil {
				return err
			}
			target.State = StateBackedUp
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
		}
		fallthrough
	case StateBackedUp:
		current, err := DigestPath(target.LivePath)
		if err != nil {
			return err
		}
		if current == target.DesiredDigest {
			target.State = StateCommitted
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
			return engine.emit(journal, index, PointTargetCommitted)
		}
		if current != DigestAbsent {
			return corruptionf("target %s/%s contains unknown bytes before install", target.Class, target.Identifier)
		}
		if err := engine.emit(journal, index, PointBeforeInstall); err != nil {
			return err
		}
		if target.DesiredDigest != DigestAbsent {
			stagedDigest, err := DigestPath(target.StagedPath)
			if err != nil || stagedDigest != target.DesiredDigest {
				if err == nil {
					err = corruptionf("staged target %s/%s digest changed", target.Class, target.Identifier)
				}
				return err
			}
			if err := durableRenameNoReplace(target.StagedPath, target.LivePath); err != nil {
				return err
			}
			if err := syncTree(target.LivePath); err != nil {
				return err
			}
		}
		if err := engine.emit(journal, index, PointAfterInstall); err != nil {
			return err
		}
		target.State = StateCommitted
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
		return engine.emit(journal, index, PointTargetCommitted)
	case StateCommitted:
		current, err := DigestPath(target.LivePath)
		if err != nil {
			return err
		}
		if current != target.DesiredDigest {
			return corruptionf("committed target %s/%s digest %s, want %s", target.Class, target.Identifier, current, target.DesiredDigest)
		}
		return nil
	default:
		return journalf("target %s/%s has state %q during commit", target.Class, target.Identifier, target.State)
	}
}

func (engine *Engine) recoverPendingBackup(target *TargetRecord) error {
	backupDigest, err := DigestPath(target.BackupPath)
	if err != nil {
		return err
	}
	if backupDigest == DigestAbsent {
		return nil
	}
	liveDigest, err := DigestPath(target.LivePath)
	if err != nil {
		return err
	}
	if liveDigest != DigestAbsent {
		return corruptionf("target %s/%s has both a pending backup and live bytes", target.Class, target.Identifier)
	}
	if target.PreimageDigest != "" && backupDigest != target.PreimageDigest {
		return corruptionf("pending backup for %s/%s has unknown bytes", target.Class, target.Identifier)
	}
	if target.ExpectedGeneration != "" {
		if err := verifyGenerationAt(target, target.BackupPath); err != nil {
			return err
		}
	}
	target.BackupDigest = backupDigest
	target.State = StateBackedUp
	return nil
}

func verifyPreimage(target *TargetRecord) error {
	if target.PreimageDigest != "" {
		current, err := DigestPath(target.LivePath)
		if err != nil {
			return err
		}
		if current != target.PreimageDigest {
			return corruptionf("target %s/%s preimage digest %s, want %s", target.Class, target.Identifier, current, target.PreimageDigest)
		}
		return nil
	}
	return verifyGenerationAt(target, target.LivePath)
}

func verifyGenerationAt(target *TargetRecord, root string) error {
	path := root
	if target.GenerationPath != "" {
		path = filepath.Join(root, filepath.FromSlash(target.GenerationPath))
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 1024*1024 {
		return corruptionf("target %s/%s generation is unavailable or unsafe", target.Class, target.Identifier)
	}
	file, err := os.Open(path) // #nosec G304 -- path is constrained to the selected live target
	if err != nil {
		return err
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(info, after) || !after.Mode().IsRegular() {
		_ = file.Close()
		return corruptionf("target %s/%s generation changed while opening", target.Class, target.Identifier)
	}
	payload, readErr := io.ReadAll(io.LimitReader(file, 1024*1024+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return errors.Join(readErr, closeErr)
	}
	if len(payload) > 1024*1024 {
		return corruptionf("target %s/%s generation exceeds the size limit", target.Class, target.Identifier)
	}
	if string(payload) != target.ExpectedGeneration {
		return corruptionf("target %s/%s generation does not match", target.Class, target.Identifier)
	}
	return nil
}

func (engine *Engine) rollback(journal *Journal) error {
	if journal.Phase != PhaseRollbackCleanup {
		journal.Phase = PhaseRollingBack
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
		for index := len(journal.Targets) - 1; index >= 0; index-- {
			if err := engine.rollbackTarget(journal, index); err != nil {
				return err
			}
		}
		journal.Phase = PhaseRollbackCleanup
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	if err := engine.removePreparedSidecars(journal); err != nil {
		return err
	}
	return engine.removeJournalDurably(journal)
}

func (engine *Engine) rollbackTarget(journal *Journal, index int) error {
	target := &journal.Targets[index]
	if target.State == StatePending {
		if backup, err := DigestPath(target.BackupPath); err != nil {
			return err
		} else if backup == DigestAbsent {
			return nil
		}
		target.State = StateBackedUp
	}
	if target.State == StateRolledBack {
		return nil
	}
	current, err := DigestPath(target.LivePath)
	if err != nil {
		return err
	}
	rollbackDigest, err := DigestPath(target.RollbackPath)
	if err != nil {
		return err
	}
	backupDigest, err := DigestPath(target.BackupPath)
	if err != nil {
		return err
	}
	// Recovery may observe the preimage already restored but StateRolledBack
	// not yet durable. Recognize only the exact recorded backup digest and only
	// after the backup path has disappeared; unknown current bytes still fail.
	if current != DigestAbsent && target.BackupDigest != "" && current == target.BackupDigest && backupDigest == DigestAbsent {
		if (rollbackDigest == target.DesiredDigest && rollbackDigest != DigestAbsent) || target.DesiredDigest == DigestAbsent {
			target.State = StateRolledBack
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
			return engine.emit(journal, index, PointTargetRolledBack)
		}
	}
	if rollbackDigest != DigestAbsent {
		if current != DigestAbsent {
			return corruptionf("target %s/%s has both live and rollback bytes", target.Class, target.Identifier)
		}
		if rollbackDigest != target.DesiredDigest {
			if existing, err := DigestPath(target.LivePath); err != nil {
				return err
			} else if existing != DigestAbsent {
				return corruptionf("rollback refused to overwrite current bytes for %s/%s", target.Class, target.Identifier)
			}
			if err := durableRenameNoReplace(target.RollbackPath, target.LivePath); err != nil {
				return errors.Join(corruptionf("rollback target %s/%s has unknown bytes", target.Class, target.Identifier), err)
			}
			return corruptionf("rollback target %s/%s desired digest mismatch", target.Class, target.Identifier)
		}
	} else if current != DigestAbsent {
		if current != target.DesiredDigest {
			return corruptionf("rollback refused unknown current bytes for %s/%s", target.Class, target.Identifier)
		}
		if err := durableRenameNoReplace(target.LivePath, target.RollbackPath); err != nil {
			return err
		}
		rollbackDigest, err = DigestPath(target.RollbackPath)
		if err != nil {
			return err
		}
		if rollbackDigest != target.DesiredDigest {
			if existing, existingErr := DigestPath(target.LivePath); existingErr != nil {
				return existingErr
			} else if existing != DigestAbsent {
				return corruptionf("rollback captured unknown bytes for %s/%s and current bytes appeared", target.Class, target.Identifier)
			}
			if restoreErr := durableRenameNoReplace(target.RollbackPath, target.LivePath); restoreErr != nil {
				return errors.Join(corruptionf("rollback captured unknown bytes for %s/%s", target.Class, target.Identifier), restoreErr)
			}
			return corruptionf("rollback captured unknown bytes for %s/%s", target.Class, target.Identifier)
		}
	} else if target.State == StateCommitted && target.DesiredDigest != DigestAbsent {
		return corruptionf("committed target %s/%s disappeared before rollback", target.Class, target.Identifier)
	}

	if target.BackupDigest == "" {
		target.BackupDigest = backupDigest
	}
	if backupDigest != target.BackupDigest {
		return corruptionf("backup for %s/%s changed before rollback", target.Class, target.Identifier)
	}
	if backupDigest != DigestAbsent {
		if existing, err := DigestPath(target.LivePath); err != nil {
			return err
		} else if existing != DigestAbsent {
			return corruptionf("rollback refused to overwrite current bytes for %s/%s", target.Class, target.Identifier)
		}
		if err := durableRenameNoReplace(target.BackupPath, target.LivePath); err != nil {
			return err
		}
		if err := syncTree(target.LivePath); err != nil {
			return err
		}
	}
	if err := engine.emit(journal, index, PointAfterRestore); err != nil {
		return err
	}
	target.State = StateRolledBack
	if err := engine.saveJournal(journal); err != nil {
		return err
	}
	return engine.emit(journal, index, PointTargetRolledBack)
}

func (engine *Engine) cleanupCommitted(journal *Journal) error {
	for index := range journal.Targets {
		target := &journal.Targets[index]
		current, err := DigestPath(target.LivePath)
		if err != nil {
			return err
		}
		if current != target.DesiredDigest {
			return corruptionf("cleanup refused changed target %s/%s", target.Class, target.Identifier)
		}
	}
	if err := engine.resumeRecordedRemoval(journal); err != nil {
		return err
	}
	for _, removal := range committedRemovalCandidates(journal) {
		if err := engine.removeRecordedSidecar(journal, removal); err != nil {
			return err
		}
	}
	return engine.removeJournalDurably(journal)
}

func (engine *Engine) removePreparedSidecars(journal *Journal) error {
	if err := engine.resumeRecordedRemoval(journal); err != nil {
		return err
	}
	for _, removal := range rollbackRemovalCandidates(journal) {
		if err := engine.removeRecordedSidecar(journal, removal); err != nil {
			return err
		}
	}
	return nil
}

type removalCandidate struct {
	targetIndex int
	path        string
	digest      string
}

func committedRemovalCandidates(journal *Journal) []removalCandidate {
	removals := make([]removalCandidate, 0, len(journal.Targets)*3)
	for index := len(journal.Targets) - 1; index >= 0; index-- {
		target := &journal.Targets[index]
		removals = appendRemovalCandidate(removals, index, target.BackupPath, target.BackupDigest)
		removals = appendRemovalCandidate(removals, index, target.RollbackPath, target.DesiredDigest)
		removals = appendRemovalCandidate(removals, index, target.StagedPath, target.DesiredDigest)
	}
	return removals
}

func rollbackRemovalCandidates(journal *Journal) []removalCandidate {
	removals := make([]removalCandidate, 0, len(journal.Targets)*3)
	for index := len(journal.Targets) - 1; index >= 0; index-- {
		target := &journal.Targets[index]
		removals = appendRemovalCandidate(removals, index, target.StagedPath, target.DesiredDigest)
		removals = appendRemovalCandidate(removals, index, target.BackupPath, target.BackupDigest)
		removals = appendRemovalCandidate(removals, index, target.RollbackPath, target.DesiredDigest)
	}
	return removals
}

func appendRemovalCandidate(removals []removalCandidate, targetIndex int, path, digest string) []removalCandidate {
	if path == "" {
		return removals
	}
	return append(removals, removalCandidate{targetIndex: targetIndex, path: path, digest: digest})
}

func findRemovalCandidate(journal *Journal, path string) (removalCandidate, bool) {
	for _, removal := range committedRemovalCandidates(journal) {
		if removal.path == path {
			return removal, true
		}
	}
	return removalCandidate{}, false
}

func (engine *Engine) resumeRecordedRemoval(journal *Journal) error {
	if journal.RemovalPath == "" {
		return nil
	}
	removal, found := findRemovalCandidate(journal, journal.RemovalPath)
	if !found {
		return journalf("removal path is not a canonical transaction sidecar")
	}
	return engine.finishRecordedRemoval(journal, removal)
}

func (engine *Engine) removeRecordedSidecar(journal *Journal, removal removalCandidate) error {
	if journal.RemovalPath != "" {
		return corruptionf("transaction %s has unfinished removal %s before %s", journal.TransactionID, journal.RemovalPath, removal.path)
	}
	present, err := validateRemovalStart(removal.path, removal.digest)
	if err != nil {
		return err
	}
	if !present {
		return nil
	}
	entries, err := captureRemovalEntries(removal.path, removal.digest)
	if err != nil {
		return err
	}
	journal.RemovalPath = removal.path
	journal.RemovalDigest = removal.digest
	journal.RemovalEntries = entries
	if err := engine.saveJournal(journal); err != nil {
		journal.RemovalPath = ""
		journal.RemovalDigest = ""
		journal.RemovalEntries = nil
		return err
	}
	return engine.finishRecordedRemoval(journal, removal)
}

func (engine *Engine) finishRecordedRemoval(journal *Journal, removal removalCandidate) error {
	if journal.RemovalPath != removal.path {
		return corruptionf("transaction %s removal ownership does not match %s", journal.TransactionID, removal.path)
	}
	if err := removeDurablyWithEntries(
		removal.path,
		journal.RemovalDigest,
		journal.RemovalEntries,
		func() error { return engine.emit(journal, removal.targetIndex, PointAfterCleanupRename) },
		func() error { return engine.emit(journal, removal.targetIndex, PointDuringCleanupRemoval) },
	); err != nil {
		return err
	}
	entries := journal.RemovalEntries
	digest := journal.RemovalDigest
	discarded := false
	stagingActive := false
	stagingCreated := false
	stagingBytes := int64(0)
	stagingPrefixDigest := ""
	stagingWriteBytes := int64(0)
	stagingWriteDigest := ""
	if journal.Phase == PhasePreparing {
		discarded = journal.Targets[removal.targetIndex].StagingDiscarded
		stagingActive = journal.Targets[removal.targetIndex].StagingActive
		stagingCreated = journal.Targets[removal.targetIndex].StagingCreated
		stagingBytes = journal.Targets[removal.targetIndex].StagingBytes
		stagingPrefixDigest = journal.Targets[removal.targetIndex].StagingPrefixDigest
		stagingWriteBytes = journal.Targets[removal.targetIndex].StagingWriteBytes
		stagingWriteDigest = journal.Targets[removal.targetIndex].StagingWriteDigest
		journal.Targets[removal.targetIndex].StagingDiscarded = true
		journal.Targets[removal.targetIndex].StagingActive = false
		journal.Targets[removal.targetIndex].StagingCreated = false
		journal.Targets[removal.targetIndex].StagingBytes = 0
		journal.Targets[removal.targetIndex].StagingPrefixDigest = ""
		journal.Targets[removal.targetIndex].StagingWriteBytes = 0
		journal.Targets[removal.targetIndex].StagingWriteDigest = ""
	}
	journal.RemovalPath = ""
	journal.RemovalDigest = ""
	journal.RemovalEntries = nil
	if err := engine.saveJournal(journal); err != nil {
		if journal.Phase == PhasePreparing {
			journal.Targets[removal.targetIndex].StagingDiscarded = discarded
			journal.Targets[removal.targetIndex].StagingActive = stagingActive
			journal.Targets[removal.targetIndex].StagingCreated = stagingCreated
			journal.Targets[removal.targetIndex].StagingBytes = stagingBytes
			journal.Targets[removal.targetIndex].StagingPrefixDigest = stagingPrefixDigest
			journal.Targets[removal.targetIndex].StagingWriteBytes = stagingWriteBytes
			journal.Targets[removal.targetIndex].StagingWriteDigest = stagingWriteDigest
		}
		journal.RemovalPath = removal.path
		journal.RemovalDigest = digest
		journal.RemovalEntries = entries
		return err
	}
	return nil
}

func (engine *Engine) emit(journal *Journal, targetIndex int, point Point) error {
	event := Event{Point: point, TransactionID: journal.TransactionID, TargetIndex: targetIndex}
	if targetIndex >= 0 {
		event.Class = journal.Targets[targetIndex].Class
		event.Identifier = journal.Targets[targetIndex].Identifier
	}
	if engine.hooks.Observe != nil {
		engine.hooks.Observe(event)
	}
	if engine.hooks.Fault != nil {
		return engine.hooks.Fault(event)
	}
	return nil
}

// ReferencedBuildKeys returns the sorted union from every live journal. GC can
// therefore retain in-flight artifacts without understanding current projects.
func (engine *Engine) ReferencedBuildKeys(lock HomeLock) ([]string, error) {
	if err := requireHomeLock(lock); err != nil {
		return nil, err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	ids, err := engine.journalIDs()
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, id := range ids {
		journal, err := engine.loadJournal(id)
		if err != nil {
			return nil, err
		}
		for _, key := range journal.ReferencedBuildKeys {
			set[key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare([]byte(keys[i]), []byte(keys[j])) < 0 })
	return keys, nil
}

func cloneJournal(journal *Journal) *Journal {
	clone := *journal
	clone.ReferencedBuildKeys = append([]string(nil), journal.ReferencedBuildKeys...)
	clone.RemovalEntries = append([]RemovalEntry(nil), journal.RemovalEntries...)
	clone.Targets = append([]TargetRecord(nil), journal.Targets...)
	for index := range clone.Targets {
		clone.Targets[index].StagingEntries = append([]RemovalEntry(nil), journal.Targets[index].StagingEntries...)
	}
	return &clone
}
