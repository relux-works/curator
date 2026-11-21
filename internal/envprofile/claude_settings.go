package envprofile

import (
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// Settings remain runtime-owned. Curator owns only its recorded exclusions.
type claudeSettingsPlan struct {
	input, payload []byte
	present        bool
	mode           os.FileMode
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
