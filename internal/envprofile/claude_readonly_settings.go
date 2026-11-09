package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// Owner-only read access matches managed homes' single-owner 0700 boundary.
const claudeReadOnlySettingsMode os.FileMode = 0o400

var claudeSettingsPaths = []string{"settings.json", "cowork_settings.json", ".config.json"}

// prepareClaudeSettings preserves operator settings bytes. The opt-in owns
// their regular-file/no-write-bit invariant, not their JSON contents. These
// files are deliberately separate from marker-owned context surfaces: unset
// leaves them in place, and they are never refreshed from a native home.
func (p *homePlan) prepareClaudeSettings(req *ResolveRequest) error {
	if p.adapter.ID != envregistry.ClaudeCode {
		return nil
	}
	switch mode := req.Machine.ClaudeSettingsMode[req.Profile]; mode {
	case "", "native":
		return nil
	case "read-only":
	default:
		return fmt.Errorf("environments.claude_settings_mode.%s: must be one of native, read-only", req.Profile)
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("environment_settings_mode_unsupported: read-only Claude settings require Unix permission bits")
	}
	p.claudeSettings = map[string][]byte{}
	root := EnvRoot(req.Home)
	rootState, err := stateread.Lstat(root)
	if err != nil {
		return err
	}
	if rootState.Kind == stateread.KindAbsent {
		p.claudeSettings["settings.json"] = []byte("{}\n")
		return nil
	}
	homeRel, err := managedRelative(root, p.homeDir)
	if err != nil {
		return err
	}
	for _, path := range claudeSettingsPaths {
		payload, present, err := readManagedRegular(root, managedJoin(homeRel, path))
		if err != nil {
			return fmt.Errorf("environment_claude_settings_invalid: %s: %w", path, err)
		}
		// readManagedRegular returns absent for directories; the contract refuses
		// every present nonregular entry rather than converting it into a file.
		metadata, err := stateread.Lstat(filepath.Join(p.homeDir, path))
		if err != nil {
			return fmt.Errorf("environment_claude_settings_invalid: %s: %w", path, err)
		}
		if metadata.Kind == stateread.KindPresent && !metadata.Info.Mode().IsRegular() {
			return fmt.Errorf("environment_claude_settings_invalid: %s must be a regular file", path)
		}
		if !present {
			if path != "settings.json" {
				continue
			}
			payload = []byte("{}\n")
		}
		p.claudeSettings[path] = payload
	}
	return nil
}

// checkClaudeSettings adds permission drift to the common status/resolve
// verifier. It never chmods or writes, including during a dry-run repair.
func (v *verification) checkClaudeSettings(p *homePlan) {
	for _, path := range claudeSettingsPaths {
		if _, enabled := p.claudeSettings[path]; !enabled {
			continue
		}
		metadata, err := stateread.Lstat(filepath.Join(p.homeDir, path))
		switch {
		case err != nil:
			v.reasons = append(v.reasons, fmt.Sprintf("Claude settings %s is unreadable: %v", path, err))
		case metadata.Kind == stateread.KindAbsent:
			v.reasons = append(v.reasons, fmt.Sprintf("Claude settings %s is missing (claude_settings_mode=read-only)", path))
		case metadata.Info == nil || !metadata.Info.Mode().IsRegular():
			v.reasons = append(v.reasons, fmt.Sprintf("Claude settings %s must be a regular file", path))
		case metadata.Info.Mode().Perm()&0o222 != 0:
			v.reasons = append(v.reasons, fmt.Sprintf("Claude settings %s has write bits (%04o); claude_settings_mode=read-only requires no write bits", path, metadata.Info.Mode().Perm()))
		}
	}
}
