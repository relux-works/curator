package main

import (
	"github.com/relux-works/curator/internal/envregistry"
)

// CLI environment aliases (manager profile §12.1): the command-line
// environment operand accepts claude and codex as aliases of the
// canonical claude_code and codex_cli registry ids. Every site below
// normalizes with envregistry.NormalizeEnvID before any validation or
// lookup, so an alias never reaches the environment_unknown refusal and
// never persists: outputs, markers, fragments, configuration records,
// and locks keep the canonical id.

// envAliasUsage is printed beneath the usage rows that take an
// environment operand. envKnobAliasUsage is its knob-path counterpart
// for the env config rows.
const (
	envAliasUsage     = "curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id"
	envKnobAliasUsage = "curator: environment positions in <knob> accept claude and codex as aliases of claude_code and codex_cli; the file keeps the canonical id"
)

// normalizeEnvKnob rewrites alias spellings in the environment positions
// of an environments knob path to their canonical ids: forms.<env>,
// in_place_mode.<env>, scoped_current.<env-or-target>, and
// isolation.<profile>.<env>. Target ids pass through unchanged — they
// are never aliases — and every other knob passes through untouched.
// The input slice is not modified.
func normalizeEnvKnob(segments []string) []string {
	rewritten := append([]string{}, segments...)
	switch {
	case len(rewritten) >= 2 && (rewritten[0] == "forms" ||
		rewritten[0] == "in_place_mode" || rewritten[0] == "scoped_current"):
		rewritten[1] = envregistry.NormalizeEnvID(rewritten[1])
	case len(rewritten) >= 3 && rewritten[0] == "isolation":
		rewritten[2] = envregistry.NormalizeEnvID(rewritten[2])
	}
	return rewritten
}

// normalizeEnvKnobValue rewrites alias spellings inside a wholesale value
// set on an environment-keyed knob: the map keys of forms,
// in_place_mode, scoped_current, and isolation.<profile>, the
// profile-nested map keys of a whole isolation object, and the env
// members of a shadow_acknowledged list. Any other knob or value shape
// passes through untouched for the configuration grammar to accept or
// refuse as today. Where an alias and its canonical id both name keys,
// the explicit canonical spelling wins, so the rewrite is deterministic.
func normalizeEnvKnobValue(segments []string, value any) any {
	switch {
	case len(segments) == 1 && (segments[0] == "forms" ||
		segments[0] == "in_place_mode" || segments[0] == "scoped_current"):
		return normalizeEnvKeyedMap(value)
	case len(segments) == 2 && segments[0] == "isolation":
		return normalizeEnvKeyedMap(value)
	case len(segments) == 1 && segments[0] == "isolation":
		obj, ok := value.(map[string]any)
		if !ok {
			return value
		}
		rewritten := make(map[string]any, len(obj))
		for profile, nested := range obj {
			rewritten[profile] = normalizeEnvKeyedMap(nested)
		}
		return rewritten
	case len(segments) == 1 && segments[0] == "shadow_acknowledged":
		list, ok := value.([]any)
		if !ok {
			return value
		}
		rewritten := make([]any, len(list))
		for i, item := range list {
			entry, ok := item.(map[string]any)
			if !ok {
				rewritten[i] = item
				continue
			}
			next := make(map[string]any, len(entry))
			for key, member := range entry {
				next[key] = member
			}
			if env, ok := next["env"].(string); ok {
				next["env"] = envregistry.NormalizeEnvID(env)
			}
			rewritten[i] = next
		}
		return rewritten
	default:
		return value
	}
}

// normalizeEnvKeyedMap rewrites alias keys of one environment-keyed map
// to their canonical ids. A non-map value passes through untouched.
func normalizeEnvKeyedMap(value any) any {
	obj, ok := value.(map[string]any)
	if !ok {
		return value
	}
	rewritten := make(map[string]any, len(obj))
	for key, member := range obj {
		if envregistry.NormalizeEnvID(key) == key {
			rewritten[key] = member
		}
	}
	for key, member := range obj {
		if canonical := envregistry.NormalizeEnvID(key); canonical != key {
			if _, present := rewritten[canonical]; !present {
				rewritten[canonical] = member
			}
		}
	}
	return rewritten
}
