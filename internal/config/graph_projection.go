package config

// ParseGraphEnvironments parses only the rendering/composition knobs selected
// by the graph command. Authentication, runtime and provider settings never
// enter the ordinary full-machine decoder.
func ParseGraphEnvironments(raw map[string]any) (Environments, error) {
	selected := map[string]any{}
	for _, key := range []string{"overlays", "overlay_default_weight", "overlays_allowed", "precedence", "forms", "system_prompt_files", "transitive_system_modules", "system_module_waivers"} {
		if value, ok := raw[key]; ok {
			selected[key] = value
		}
	}
	return parseEnvironmentsVersion(selected, SchemaVersion7)
}

// MergeGraphEnvironments reuses the machine precedence/lock merge on the
// restricted surface. No unrelated configuration values are accepted here.
func MergeGraphEnvironments(user, system map[string]any, locked map[string]bool) (Environments, error) {
	if _, err := parseSystemEnvironments(system); err != nil {
		return Environments{}, err
	}
	merged := mergeSystemEnvironments(map[string]any{"environments": system}, map[string]any{"environments": user}, locked, "", nil)
	return parseEnvironmentsVersion(merged, SchemaVersion7)
}
