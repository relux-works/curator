package transaction

import (
	"errors"
	"fmt"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envunit"
)

// RetainedGlobalConflict reports a committed, intentionally nonlaunchable L1.
// It is distinct from an ordinary failure whose transaction must restore L0.
type RetainedGlobalConflict struct{ Cleanup error }

func (e *RetainedGlobalConflict) Error() string {
	if e.Cleanup != nil {
		return fmt.Sprintf("environment_surface_unmanaged_conflict: retained L1; cleanup pending: %v", e.Cleanup)
	}
	return "environment_surface_unmanaged_conflict: retained L1; explicit profile sync --takeover is required"
}
func (e *RetainedGlobalConflict) Unwrap() error { return e.Cleanup }

func (engine *Engine) rollbackGlobal(j *Journal, cause error) error {
	j.Phase = PhaseRollingBack
	if err := engine.saveJournal(j); err != nil {
		return errors.Join(cause, err)
	}
	return errors.Join(cause, engine.rollback(j))
}

func (engine *Engine) commitGlobalRetention(j *Journal) error {
	if err := envjournal.ValidateGlobal(j); err != nil {
		return engine.rollbackGlobal(j, err)
	}
	if envjournal.GlobalRetained(j) {
		return engine.finishGlobalRetention(j)
	}
	if err := engine.commitDirectorySubset(j, func(path string) bool { return envjournal.GlobalPublishedDirectory(j, path) }); err != nil {
		return engine.rollbackGlobal(j, err)
	}
	// The receipt is not a publication prefix: confirm the exact captured
	// unmanaged entry after L1 is durable, then record that disposition before
	// committing the receipt. No native/home or legacy target has been touched.
	for i, t := range j.Targets {
		if t.Class == "lifecycle" || !envjournal.GlobalPublishedTarget(j, t.LivePath) {
			continue
		}
		if err := engine.commitTarget(j, i); err != nil {
			return engine.rollbackGlobal(j, err)
		}
	}
	if j.Global.Healthy {
		j.Global.State = "forward"
		if err := engine.saveJournal(j); err != nil {
			return err
		}
		return engine.commitRemainingTargets(j)
	}
	// ValidateGlobal binds the conflict to one exact typed unit preimage.
	// Directory digests must use the same closed tree representation at capture
	// and confirmation; observing just the directory entry loses that binding.
	class := ""
	for _, target := range j.Unit.Targets {
		if target.Path == j.Global.Conflict.Target {
			class = target.Class
			break
		}
	}
	actual, err := envunit.ObserveTarget(j.Global.Conflict.Target, class)
	if err != nil {
		return engine.rollbackGlobal(j, err)
	}
	if actual != j.Global.Conflict.Entry {
		return engine.rollbackGlobal(j, corruptionf("global conflict entry changed after lock publication"))
	}
	j.Global.State = "retaining"
	if err := engine.saveJournal(j); err != nil {
		return err
	}
	if err := engine.emit(j, -1, PointGlobalRetained); err != nil {
		return &RetainedGlobalConflict{Cleanup: err}
	}
	return engine.finishGlobalRetention(j)
}

// Recovery only completes receipt publication and cleanup after confirmation.
// Earlier interruption compensates L1; it never resumes a forward home write.
func (engine *Engine) finishGlobalRetention(j *Journal) error {
	if !envjournal.GlobalRetained(j) {
		return corruptionf("global retention was not confirmed")
	}
	for i, t := range j.Targets {
		if t.Class == "lifecycle" {
			if err := engine.commitTarget(j, i); err != nil {
				return &RetainedGlobalConflict{Cleanup: err}
			}
		}
	}
	j.Phase = PhaseCleanup
	if err := engine.saveJournal(j); err != nil {
		return &RetainedGlobalConflict{Cleanup: err}
	}
	if err := engine.emit(j, -1, PointBeforeCleanup); err != nil {
		return &RetainedGlobalConflict{Cleanup: err}
	}
	if err := engine.cleanupCommitted(j); err != nil {
		return &RetainedGlobalConflict{Cleanup: err}
	}
	return &RetainedGlobalConflict{}
}
