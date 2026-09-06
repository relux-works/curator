package main

import (
	"encoding/json"
	"fmt"

	"github.com/relux-works/curator/internal/config"
)

// cmdEnvConfig implements `env config show|set|unset` (cli/curator.md): it
// reads or edits one environments §12.1 knob of manager-config schema 2 by
// its table name. Show reads the effective configuration with the §12.1
// defaults applied; set and unset rewrite the machine file only and refuse
// a locked knob with the system-file warning.
func (c cli) cmdEnvConfig(cfg *config.Config, args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env config show|set|unset [<knob> [<value>]]")
		return exitUsage
	}
	switch args[0] {
	case "show":
		return c.cmdEnvConfigShow(cfg, args[1:])
	case "set":
		return c.cmdEnvConfigSet(cfg, args[1:])
	case "unset":
		return c.cmdEnvConfigUnset(cfg, args[1:])
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown env config subcommand %q\n", args[0])
	return exitUsage
}

func (c cli) cmdEnvConfigShow(cfg *config.Config, args []string) int {
	if len(args) > 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env config show [<knob>]")
		return exitUsage
	}
	rendered := cfg.EffectiveJSON()
	env, _ := rendered["environments"].(map[string]any)
	if len(args) == 0 {
		return c.printJSON(env)
	}
	segments, err := config.SplitEnvKnob(args[0])
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	value, present := lookupKnob(env, segments)
	if !present {
		value = nil
	}
	return c.printJSON(value)
}

func (c cli) cmdEnvConfigSet(cfg *config.Config, args []string) int {
	if len(args) != 2 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env config set <knob> <value>")
		return exitUsage
	}
	knob, rawValue := args[0], args[1]
	segments, err := config.SplitEnvKnob(knob)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	if lock := config.EnvLockKey(knob); lock != "" && cfg.LockedBySystem(lock) {
		_, _ = fmt.Fprintf(c.stderr, "curator: environments knob %q is locked by %s; the system value wins\n", knob, cfg.SystemConfigPath)
		return exitFail
	}
	var value any
	if err := json.Unmarshal([]byte(rawValue), &value); err != nil {
		value = rawValue
	}
	object, err := config.ReadRaw(cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	env, _ := object["environments"].(map[string]any)
	if env == nil {
		env = map[string]any{}
		object["environments"] = env
	}
	setKnob(env, segments, value)
	if schema, _ := object["schema_version"].(float64); schema == float64(config.SchemaVersion) {
		object["schema_version"] = float64(config.SchemaVersion2)
	}
	updated, err := config.WriteRaw(cfg.Path, object)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return c.printJSON(lookupEffective(updated, segments))
}

func (c cli) cmdEnvConfigUnset(cfg *config.Config, args []string) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env config unset <knob>")
		return exitUsage
	}
	knob := args[0]
	segments, err := config.SplitEnvKnob(knob)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	if lock := config.EnvLockKey(knob); lock != "" && cfg.LockedBySystem(lock) {
		_, _ = fmt.Fprintf(c.stderr, "curator: environments knob %q is locked by %s; the system value wins\n", knob, cfg.SystemConfigPath)
		return exitFail
	}
	object, err := config.ReadRaw(cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	env, _ := object["environments"].(map[string]any)
	if env == nil {
		_, _ = fmt.Fprintf(c.stderr, "curator: environments knob %q is not set\n", knob)
		return exitFail
	}
	if !deleteKnob(env, segments) {
		_, _ = fmt.Fprintf(c.stderr, "curator: environments knob %q is not set\n", knob)
		return exitFail
	}
	if _, err := config.WriteRaw(cfg.Path, object); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "unset %s\n", knob)
	return exitOK
}

func (c cli) printJSON(value any) int {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, string(payload))
	return exitOK
}

func lookupKnob(node map[string]any, segments []string) (any, bool) {
	var current any = node
	for _, segment := range segments {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = obj[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func setKnob(node map[string]any, segments []string, value any) {
	for _, segment := range segments[:len(segments)-1] {
		next, _ := node[segment].(map[string]any)
		if next == nil {
			next = map[string]any{}
			node[segment] = next
		}
		node = next
	}
	node[segments[len(segments)-1]] = value
}

func deleteKnob(node map[string]any, segments []string) bool {
	for _, segment := range segments[:len(segments)-1] {
		next, _ := node[segment].(map[string]any)
		if next == nil {
			return false
		}
		node = next
	}
	leaf := segments[len(segments)-1]
	if _, present := node[leaf]; !present {
		return false
	}
	delete(node, leaf)
	return true
}

func lookupEffective(cfg *config.Config, segments []string) any {
	env, _ := cfg.EffectiveJSON()["environments"].(map[string]any)
	value, present := lookupKnob(env, segments)
	if !present {
		return nil
	}
	return value
}
