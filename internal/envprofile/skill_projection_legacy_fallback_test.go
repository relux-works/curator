package envprofile

import (
	"errors"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
)

// Historical fallback oracle only. Production replacement is owned by the
// whole-home transaction; no legacy helper confers admission to a managed home.
// Replace the complete copied tree through the existing recoverable entry
// transaction. Overlay-copying would retain excluded files from the old tree;
// renaming a symlink over a non-empty directory is unavailable on some hosts.
func legacyPublishSkillFallback(op *operation, root, rel, source string) (resultErr error) {
	live, err := managedPath(root, rel, true)
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(live), ".curator-skill-copy-*")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(stage)) }()
	stageRel, err := managedRelative(root, stage)
	if err != nil {
		return err
	}
	nextRel := managedJoin(stageRel, "next")
	if err := copyTree(source, root, nextRel); err != nil {
		return err
	}
	digest, err := transaction.DigestTarget(transaction.KindEntry, live)
	if err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-skill-projection", Targets: []transaction.Target{{
		Class: "skill-context", Identifier: rel, Kind: transaction.KindEntry, LivePath: live,
		StagedSource: filepath.Join(stage, "next"), PreimageDigest: digest,
	}}}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	return op.engine.Commit(op.lock, plan.TransactionID)
}
