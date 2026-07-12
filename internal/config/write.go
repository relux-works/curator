package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Bootstrap writes a minimal schema 1 user configuration. Existing files are
// preserved unless force is true.
func Bootstrap(path, skillsRoot, preferredLocale string, defaultAgents []string, force bool) error {
	if skillsRoot == "" {
		return fmt.Errorf("skills_root must be a non-empty path")
	}
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("config already exists at %s; pass --force to overwrite", path)
	}
	if len(defaultAgents) == 0 {
		defaultAgents = append([]string(nil), DefaultAgents...)
	}
	payload := map[string]any{
		"schema_version": 1,
		"skills_root":    skillsRoot,
		"default_agents": defaultAgents,
		"adapter_mode":   "auto",
		"projects":       map[string]any{},
	}
	if preferredLocale != "" {
		payload["preferred_locale"] = preferredLocale
	}
	if _, err := Parse(toFloatJSON(payload), path); err != nil {
		return err
	}
	return writeObjectAtomic(path, payload)
}

// AddProject records one project in the user configuration without
// disturbing unrelated user fields or enforced-system defaults.
func AddProject(path, alias, projectPath string, agents []string) error {
	if alias == "" {
		return fmt.Errorf("project alias must be non-empty")
	}
	resolved, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("project path does not exist: %s", resolved)
	}
	object, err := readObject(path)
	if err != nil {
		return err
	}
	projects, ok := object["projects"].(map[string]any)
	if !ok {
		return fmt.Errorf("projects must be an object")
	}
	entry := map[string]any{"path": resolved}
	if len(agents) > 0 {
		entry["agents"] = agents
	}
	projects[alias] = toFloatJSON(entry)
	if _, err := Parse(object, path); err != nil {
		return err
	}
	return writeObjectAtomic(path, object)
}

func writeObjectAtomic(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(append(payload, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		return os.Rename(tmpPath, path)
	}
	return nil
}

// toFloatJSON applies the same number representation as json.Unmarshal,
// which is what Parse receives from disk.
func toFloatJSON(value any) map[string]any {
	payload, _ := json.Marshal(value)
	var object map[string]any
	_ = json.Unmarshal(payload, &object)
	return object
}
