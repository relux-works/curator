package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/claudememory"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

// Settings remain runtime-owned; the manager merges only its native-memory
// exclusion. Never read the native settings or memory file to produce it.
type claudeSettingsPlan struct {
	input, payload []byte
	present        bool
	mode           os.FileMode
}

func prepareClaudeSettings(root, homeRel string) (*claudeSettingsPlan, error) {
	input, present, mode, err := readClaudeSettingsInput(root, homeRel)
	if err != nil {
		return nil, err
	}
	native, err := claudememory.NativeFile()
	if err != nil {
		return nil, err
	}
	if present && len(input) == 0 {
		return nil, fmt.Errorf("%s: managed Claude settings are empty", envregistry.DiagSeedUnreadable)
	}
	payload, err := claudememory.Merge(input, native)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, err)
	}
	plan := &claudeSettingsPlan{input: input, present: present, mode: mode}
	if !bytes.Equal(payload, input) {
		plan.payload = payload
	}
	return plan, nil
}

// Re-read and re-merge before publication. Native rename preparation precedes
// the final read, as for the onboarding seed. Three changed inputs refuse;
// this narrows the external-writer window but does not provide filesystem CAS.
func publishClaudeSettings(root, homeRel string) error {
	for attempt := 0; attempt < 3; attempt++ {
		plan, err := prepareClaudeSettings(root, homeRel)
		if err != nil || plan.payload == nil {
			return err
		}
		full, err := managedPath(root, managedJoin(homeRel, "settings.json"), true)
		if err != nil {
			return err
		}
		file, err := privatedir.CreateTemp(filepath.Dir(full), ".curator-write-*.tmp")
		if err != nil {
			return err
		}
		temp := file.Name()
		_, err = file.Write(plan.payload)
		if err == nil {
			// This publisher owns only the exclusion. Preserve the existing
			// mode, including a read-only policy applied by another owner.
			// Chmod follows the write so writing cannot clear special bits.
			err = file.Chmod(plan.mode)
		}
		if err == nil {
			err = file.Sync()
		}
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(temp)
			return fmt.Errorf("%s: stage managed Claude settings without changing mode: %w", envregistry.DiagSeedUnreadable, err)
		}
		rename, release, err := prepareClaudeSeedRename(temp, full)
		if err != nil {
			_ = os.Remove(temp)
			return err
		}
		input, present, mode, readErr := readClaudeSettingsInput(root, homeRel)
		if readErr != nil || present != plan.present || mode != plan.mode || !bytes.Equal(input, plan.input) {
			release()
			_ = os.Remove(temp)
			if readErr != nil {
				return readErr
			}
			continue
		}
		err = rename()
		release()
		_ = os.Remove(temp)
		return err
	}
	return fmt.Errorf("%s: managed Claude settings changed during all 3 publication checks", envregistry.DiagSeedUnreadable)
}

func readClaudeSettingsInput(root, homeRel string) ([]byte, bool, os.FileMode, error) {
	input, present, err := readManagedRegular(root, managedJoin(homeRel, "settings.json"))
	if err != nil {
		return nil, false, 0, err
	}
	full, err := managedPath(root, managedJoin(homeRel, "settings.json"), false)
	if err != nil {
		return nil, false, 0, err
	}
	state, err := stateread.Lstat(full)
	if err != nil {
		return nil, false, 0, err
	}
	if !present && state.Kind == stateread.KindAbsent {
		return nil, false, managedAuthoredFileMode, nil
	}
	if !present || state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.Mode().IsRegular() {
		return nil, false, 0, fmt.Errorf("%s: managed Claude settings are not a stable regular file", envregistry.DiagSeedUnreadable)
	}
	return input, true, state.Info.Mode(), nil
}
