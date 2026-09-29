package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/stateread"
)

var errBootstrapCheckpointSignature = errors.New("checkpoint signature failed verification")

// readBootstrapCheckpoint reads an operator-provided checkpoint only after
// proving its containing tree is manager-owned and privately mutable. The
// final open is no-follow and checked against the stateread metadata so a
// raced replacement cannot be mistaken for the validated file.
func readBootstrapCheckpoint(path string, pinnedKeys []string) (parsedSnapshot, error) {
	if path == "" {
		return parsedSnapshot{}, fmt.Errorf("checkpoint path is empty")
	}
	parent := filepath.Dir(path)
	if err := pathboundary.Validate(parent); err != nil {
		return parsedSnapshot{}, fmt.Errorf("checkpoint path %s failed manager protection checks: %w", path, err)
	}
	metadata, err := stateread.Lstat(path)
	if err != nil {
		return parsedSnapshot{}, fmt.Errorf("checkpoint path %s is unreadable: %w", path, err)
	}
	if metadata.Kind == stateread.KindAbsent {
		return parsedSnapshot{}, stateread.AbsentError(path)
	}
	info := metadata.Info
	if info == nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return parsedSnapshot{}, stateread.UnusableError(path, fmt.Errorf("checkpoint is not a regular file"))
	}
	if err := pathboundary.CheckPrivateFile(path, info); err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, fmt.Errorf("checkpoint permits foreign mutation: %w", err))
	}
	file, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, err)
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, err)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return parsedSnapshot{}, stateread.UnusableError(path, fmt.Errorf("checkpoint changed while opening"))
	}
	payload, err := io.ReadAll(file)
	if err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, err)
	}
	var snapshot map[string]any
	if err := decodeJSON(payload, &snapshot); err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, err)
	}
	parsed, err := parseSnapshot(snapshot)
	if err != nil {
		return parsedSnapshot{}, stateread.UnusableError(path, err)
	}
	if !VerifySigned(snapshot, pinnedKeys) {
		return parsedSnapshot{}, stateread.UnusableError(path, errBootstrapCheckpointSignature)
	}
	return parsed, nil
}

// reconcileBootstrapCheckpoint applies the operator checkpoint to protected
// state. A missing state requires a valid checkpoint. With existing state, a
// checkpoint with an invalid signature is rejected without changing the
// established high-water; unreadable or malformed checkpoint inputs fail
// closed. A valid lower or conflicting equal checkpoint is reported as a
// regression.
func reconcileBootstrapCheckpoint(reg Registry, state snapshotState, stateExists bool) (snapshotState, bool, bool, error) {
	checkpoint, err := readBootstrapCheckpoint(reg.BootstrapCheckpoint, reg.PublicKeys)
	if err != nil {
		if stateExists && errors.Is(err, errBootstrapCheckpointSignature) {
			return state, false, false, nil
		}
		return state, false, false, err
	}
	checkpointID := bootstrapCheckpointID(checkpoint)
	if stateExists && state.BootstrapSource == "checkpoint" && state.BootstrapCheckpointID == checkpointID {
		// The configured checkpoint already established this registry's trust
		// source. Its high-water may since have advanced from accepted network
		// views, so do not reinterpret the unchanged bootstrap input as a new
		// rebootstrap request on every operation or status read.
		return state, false, false, nil
	}
	if !stateExists {
		return snapshotState{
			HighestVersion:        checkpoint.Version,
			Head:                  checkpoint.Head,
			MerkleRoot:            checkpoint.MerkleRoot,
			LogSize:               checkpoint.LogSize,
			BootstrapSource:       "checkpoint",
			BootstrapCheckpointID: checkpointID,
		}, true, false, nil
	}
	if checkpoint.Version < state.HighestVersion ||
		(checkpoint.Version == state.HighestVersion && statesConflict(state, snapshotState{
			Head: checkpoint.Head, MerkleRoot: checkpoint.MerkleRoot, LogSize: checkpoint.LogSize,
		})) {
		return state, false, true, nil
	}
	if checkpoint.Version == state.HighestVersion {
		return state, false, false, nil
	}
	next := snapshotState{
		HighestVersion:        checkpoint.Version,
		Head:                  checkpoint.Head,
		MerkleRoot:            checkpoint.MerkleRoot,
		LogSize:               checkpoint.LogSize,
		BootstrapSource:       "checkpoint",
		BootstrapCheckpointID: checkpointID,
	}
	return next, true, false, nil
}

func bootstrapCheckpointID(checkpoint parsedSnapshot) string {
	identity := fmt.Sprintf("%d\x00%d\x00%s\x00%s", checkpoint.Version, checkpoint.LogSize, checkpoint.Head, checkpoint.MerkleRoot)
	sum := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(sum[:])
}

func checkpointRegressionMessage(reg Registry) string {
	return fmt.Sprintf("registry_checkpoint_regression (error): registry %s checkpoint is below or conflicts with its persisted high-water", reg.Name)
}
