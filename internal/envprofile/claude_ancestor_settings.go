package envprofile

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/relux-works/curator/internal/claudememory"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

type claudeAncestorSettingsPlan struct {
	settings            *claudeSettingsPlan
	settingsPath        string
	statePath           string
	stateInput          []byte
	statePayload        []byte
	statePresent        bool
	stateMode           os.FileMode
	settingsDigest      string
	stateDigest         string
	ancestorDiagnostics []string
}

func claudeMemoryStateRel(req *ResolveRequest) string {
	identity := filepath.ToSlash(filepath.Join(req.Profile, req.EnvID))
	key := sha256.Sum256([]byte(identity))
	return path.Join("state", "claude-memory", "v1", fmt.Sprintf("%x.json", key))
}

// All inputs, including the ownership record, are validated before recovery or
// any home mutation. The same plan is rebuilt under the held publication lock.
func prepareClaudeAncestorSettings(req *ResolveRequest) (*claudeAncestorSettingsPlan, error) {
	homeRel, err := managedRelative(req.Home, ManagedHomeDir(req.Home, req.Profile, req.EnvID))
	if err != nil {
		return nil, err
	}
	input, present, mode, err := readClaudeSettingsInput(req.Home, homeRel)
	if err != nil {
		return nil, err
	}
	if present && len(input) == 0 {
		return nil, fmt.Errorf("%s: managed Claude settings are empty", envregistry.DiagSeedUnreadable)
	}
	stateRel := claudeMemoryStateRel(req)
	state, statePresent, err := readManagedRegular(req.Home, stateRel)
	if err != nil {
		return nil, err
	}
	statePath, err := managedPath(req.Home, stateRel, false)
	if err != nil {
		return nil, err
	}
	stateMode := os.FileMode(0o600)
	if statePresent {
		if err := privatedir.ValidateFile(statePath); err != nil {
			return nil, fmt.Errorf("%s: Claude memory ownership is not private: %w", envregistry.DiagSeedUnreadable, err)
		}
		info, err := os.Lstat(statePath)
		if err != nil {
			return nil, err
		}
		stateMode = info.Mode()
	}
	// Global composition can omit LaunchDir. Use the working directory only
	// for memory discovery; changing the request would also change seed trust.
	launchDir := req.LaunchDir
	if launchDir == "" {
		launchDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("%s: resolve Claude working directory: %w", envregistry.DiagSeedUnreadable, err)
		}
	}
	projectRoot, err := claudememory.ProjectRoot(launchDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, err)
	}
	managedRoot := filepath.Join(ManagedHomeDir(req.Home, req.Profile, req.EnvID), "CLAUDE.md")
	ancestors, err := claudememory.AncestorFiles(projectRoot, launchDir, managedRoot)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, err)
	}
	native, err := claudememory.NativeFile()
	if err != nil {
		return nil, err
	}
	if err := claudememory.ValidateNativeTargets(native, projectRoot, managedRoot); err != nil {
		return nil, fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, err)
	}
	payload, nextState, err := claudememory.Reconcile(input, state, native, ancestors)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, err)
	}
	settingsPath, err := managedPath(req.Home, managedJoin(homeRel, "settings.json"), false)
	if err != nil {
		return nil, err
	}
	plan := &claudeAncestorSettingsPlan{
		settings:     &claudeSettingsPlan{input: input, present: present, mode: mode},
		settingsPath: settingsPath, statePath: statePath, stateInput: state, statePresent: statePresent, stateMode: stateMode,
	}
	if !bytes.Equal(payload, input) {
		plan.settings.payload = payload
		missing, err := claudememory.MissingAncestorFiles(input, ancestors)
		if err != nil {
			return nil, err
		}
		for _, path := range missing {
			metadata, err := stateread.Lstat(path)
			if err != nil {
				return nil, fmt.Errorf("%s: inspect ancestor memory: %w", envregistry.DiagSeedUnreadable, err)
			}
			if metadata.Kind == stateread.KindAbsent {
				continue
			}
			plan.ancestorDiagnostics = append(plan.ancestorDiagnostics, "managed Claude ancestor exclusion is missing for "+path)
		}
	}
	if !bytes.Equal(nextState, state) {
		plan.statePayload = nextState
	}
	// Digests include metadata and bytes, and are rechecked by the existing
	// engine before either target changes. No filesystem CAS is claimed.
	plan.settingsDigest, err = transaction.DigestPath(settingsPath)
	if err == nil {
		plan.stateDigest, err = transaction.DigestPath(statePath)
	}
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// publishClaudeAncestorSettings uses the existing journal and recovery route;
// neither a new journal protocol nor a vendor settings extension is needed.
func publishClaudeAncestorSettings(op *operation, req *ResolveRequest) error {
	for attempt := 0; attempt < 3; attempt++ {
		err := publishClaudeAncestorSettingsAttempt(op, req)
		if !errors.Is(err, errClaudeMemoryInputsChanged) {
			return err
		}
	}
	return fmt.Errorf("%s: managed Claude memory inputs changed during all 3 publication checks", envregistry.DiagSeedUnreadable)
}

var errClaudeMemoryInputsChanged = errors.New("managed Claude memory inputs changed before publication")

func publishClaudeAncestorSettingsAttempt(op *operation, req *ResolveRequest) error {
	plan, err := prepareClaudeAncestorSettings(req)
	if err != nil || (plan.settings.payload == nil && plan.statePayload == nil) {
		return err
	}
	if err := op.capture(plan.settingsPath, plan.statePath); err != nil {
		return err
	}
	tx := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "claude-memory-policy", ReadInputs: map[string]string{}}
	var temps []string
	defer func() {
		for _, path := range temps {
			_ = os.Remove(path)
		}
	}()
	for _, file := range []struct {
		path, digest string
		payload      []byte
		mode         os.FileMode
		private      bool
	}{
		{plan.settingsPath, plan.settingsDigest, plan.settings.payload, plan.settings.mode, false},
		{plan.statePath, plan.stateDigest, plan.statePayload, plan.stateMode, true},
	} {
		if file.payload == nil {
			tx.ReadInputs[file.path] = file.digest
			continue
		}
		if err := privatedir.MakeAll(filepath.Dir(file.path)); err != nil {
			return err
		}
		stage, err := privatedir.CreateTemp(filepath.Dir(file.path), ".curator-write-*.tmp")
		if err != nil {
			return err
		}
		temps = append(temps, stage.Name())
		_, err = stage.Write(file.payload)
		if err == nil {
			err = stage.Chmod(file.mode)
		}
		if err == nil {
			err = stage.Sync()
		}
		closeErr := stage.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			return err
		}
		tx.Targets = append(tx.Targets, transaction.Target{Class: "claude-memory", Identifier: file.path, LivePath: file.path, StagedSource: stage.Name(), PreimageDigest: file.digest, PrivateFile: file.private})
	}
	// Close the input-read window before entering the engine. The engine then
	// checks both pinned preimages and recovers ordered replacements normally.
	latest, err := prepareClaudeAncestorSettings(req)
	if err != nil {
		return err
	}
	if !bytes.Equal(latest.settings.input, plan.settings.input) || latest.settings.present != plan.settings.present || latest.settings.mode != plan.settings.mode || !bytes.Equal(latest.stateInput, plan.stateInput) || latest.statePresent != plan.statePresent || latest.stateMode != plan.stateMode {
		return errClaudeMemoryInputsChanged
	}
	if _, err := op.engine.Prepare(op.witness(), tx); err != nil {
		return err
	}
	if err := op.engine.Commit(op.witness(), tx.TransactionID); err != nil {
		return err
	}
	// Ownership is outside the managed adapter home. Record this publisher's
	// pair before a later global phase can fail and compensate it.
	return op.recordPostimages(plan.settingsPath, plan.statePath)
}
