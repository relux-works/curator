// Package mcp verifies declared MCP server requirements against the
// configuration surfaces of the target agent environments (Spec §11).
//
// The check is read-only: Curator never provisions or launches MCP servers.
// Missing or malformed configuration files count as configuring no servers.
package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
	"github.com/relux-works/curator/internal/skillspec"
)

// projectSources maps agents to project-relative config files (Spec §11.1).
var projectSources = map[string][]string{
	"codex_cli":   {".codex/config.toml"},
	"opencode":    {"opencode.json", "opencode.jsonc"},
}

// homeSources maps agents to home-relative config files.
var homeSources = map[string][]string{
	"codex_cli":   {".codex/config.toml"},
	"opencode":    {".config/opencode/opencode.json", ".config/opencode/opencode.jsonc"},
}

// claudeSettings can reject servers declared in .mcp.json; a rejected server
// never activates, so it counts as not configured.
var claudeSettings []string

func init() {
	for _, id := range harnessregistry.IDs() {
		adapter, err := harnessregistry.Lookup(id)
		if err != nil || !adapter.Capabilities().Has(harnessapi.MCPInspection) {
			continue
		}
		projectContext, err := harnessapi.NewContext(runtime.GOOS, harnessapi.ProjectScope, harnessapi.VerifiedRelease{})
		if err != nil {
			continue
		}
		project, err := adapter.MCP(projectContext, harnessapi.ProjectScope)
		if err == nil {
			for _, path := range project.Paths().Paths() {
				projectSources[string(id)] = append(projectSources[string(id)], path.String())
			}
		}
		homeContext, err := harnessapi.NewContext(runtime.GOOS, harnessapi.UserScope, harnessapi.VerifiedRelease{})
		if err != nil {
			continue
		}
		home, err := adapter.MCP(homeContext, harnessapi.UserScope)
		if err == nil {
			for _, path := range home.Paths().Paths() {
				homeSources[string(id)] = append(homeSources[string(id)], path.String())
			}
		}
		if adapter.Capabilities().Has(harnessapi.TrustObservation) {
			settings, err := adapter.Settings(projectContext)
			if err == nil {
				for _, path := range settings.Paths().Paths() {
					claudeSettings = append(claudeSettings, path.String())
				}
			}
		}
	}
}

// Env fixes the filesystem roots for resolution (tests inject temp dirs).
type Env struct {
	ProjectRoot string
	UserHome    string
}

// ConfiguredServers returns the server names configured for one agent
// across project and user surfaces.
func ConfiguredServers(env Env, agent string) (map[string]any, error) {
	project, err := projectEntries(env, agent)
	if err != nil {
		return nil, err
	}
	home, err := homeEntries(env, agent)
	if err != nil {
		return nil, err
	}
	entries := map[string]any{}
	for name, entry := range project {
		entries[name] = entry
	}
	for name, entry := range home {
		entries[name] = entry
	}
	return entries, nil
}

// ProjectOnlyServers returns names configured only through project-level
// surfaces: agents gate those behind checkout trust (Spec §11.3).
func ProjectOnlyServers(env Env, agent string) (map[string]bool, error) {
	project, err := projectEntries(env, agent)
	if err != nil {
		return nil, err
	}
	home, err := homeEntries(env, agent)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for name := range project {
		if _, present := home[name]; !present {
			out[name] = true
		}
	}
	return out, nil
}

// Finding is the verification outcome for one skill.
type Finding struct {
	// FoundIn maps each declared server to the sorted agents where it was
	// found; recorded in the install marker.
	FoundIn map[string][]string
}

// Verify checks every requirement of every node against the target agents
// (Spec §11.2, §11.3). Returns per-skill findings and warnings; a failed
// requirement is an error carrying the hint.
func Verify(env Env, agents []string, skills map[string]map[string]skillspec.McpServer) (map[string]Finding, []string, error) {
	findings := map[string]Finding{}
	var warnings []string
	var problems []string

	skillNames := make([]string, 0, len(skills))
	for name := range skills {
		skillNames = append(skillNames, name)
	}
	sort.Strings(skillNames)

	for _, skillName := range skillNames {
		requirements := skills[skillName]
		if len(requirements) == 0 {
			continue
		}
		foundIn := map[string][]string{}
		reqNames := make([]string, 0, len(requirements))
		for name := range requirements {
			reqNames = append(reqNames, name)
		}
		sort.Strings(reqNames)
		for _, serverName := range reqNames {
			requirement := requirements[serverName]
			var available, missing []string
			for _, agent := range agents {
				configuredServers, err := ConfiguredServers(env, agent)
				if err != nil {
					return findings, warnings, err
				}
				if _, configured := configuredServers[serverName]; configured {
					available = append(available, agent)
				} else {
					missing = append(missing, agent)
				}
			}
			sort.Strings(available)
			sort.Strings(missing)
			foundIn[serverName] = available

			if requirement.RequiredIn == "all" {
				if len(missing) > 0 {
					problems = append(problems, fmt.Sprintf(
						"MCP server %q required by %s is not configured for agent(s): %s. Hint: %s",
						serverName, skillName, strings.Join(missing, ", "), requirement.Hint))
					continue
				}
			} else if len(available) == 0 {
				problems = append(problems, fmt.Sprintf(
					"MCP server %q required by %s is not configured in any target agent environment. Hint: %s",
					serverName, skillName, requirement.Hint))
				continue
			}

			// Static stdio PATH probe: warn only when every entry for the
			// server is positively a stdio server whose command is missing.
			for _, agent := range available {
				command, missing, err := missingStdioCommand(env, agent, serverName)
				if err != nil {
					return findings, warnings, err
				}
				if missing {
					warnings = append(warnings, fmt.Sprintf(
						"MCP server %q for %s runs %q, which is not on PATH", serverName, agent, command))
				}
			}
			// Project-only pending warning.
			var trustGated []string
			for _, agent := range available {
				projectOnly, err := ProjectOnlyServers(env, agent)
				if err != nil {
					return findings, warnings, err
				}
				if projectOnly[serverName] {
					trustGated = append(trustGated, agent)
				}
			}
			if len(trustGated) > 0 {
				warnings = append(warnings, fmt.Sprintf(
					"MCP server %q is declared only in project-level config for %s; agents keep project servers pending until the checkout is trusted",
					serverName, strings.Join(trustGated, ", ")))
			}
		}
		findings[skillName] = Finding{FoundIn: foundIn}
	}
	if len(problems) > 0 {
		return findings, warnings, fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return findings, warnings, nil
}

func missingStdioCommand(env Env, agent, serverName string) (string, bool, error) {
	var commands []string
	project, err := projectEntries(env, agent)
	if err != nil {
		return "", false, err
	}
	home, err := homeEntries(env, agent)
	if err != nil {
		return "", false, err
	}
	for _, source := range []map[string]any{project, home} {
		entry, present := source[serverName]
		if !present {
			continue
		}
		command, ok := stdioCommand(entry)
		if !ok {
			return "", false, nil // possibly remote: no warning
		}
		commands = append(commands, command)
	}
	if len(commands) == 0 {
		return "", false, nil
	}
	for _, command := range commands {
		if _, err := exec.LookPath(command); err == nil {
			return "", false, nil
		}
	}
	return commands[0], true, nil
}

// stdioCommand extracts the executable of a stdio entry; false when the
// entry may be remote or has an unrecognized shape.
func stdioCommand(entry any) (string, bool) {
	obj, ok := entry.(map[string]any)
	if !ok {
		return "", false
	}
	switch command := obj["command"].(type) {
	case string:
		if command != "" {
			return command, true
		}
	case []any:
		// OpenCode local servers declare the command as an argv list.
		if len(command) > 0 {
			if first, ok := command[0].(string); ok && first != "" {
				return first, true
			}
		}
	}
	return "", false
}

func projectEntries(env Env, agent string) (map[string]any, error) {
	entries := map[string]any{}
	sources, err := mcpSources(agent, harnessapi.ProjectScope, projectSources[agent])
	if err != nil {
		return nil, err
	}
	for _, rel := range sources {
		for name, entry := range entriesInFile(filepath.Join(env.ProjectRoot, filepath.FromSlash(rel))) {
			entries[name] = entry
		}
	}
	if agent == "claude_code" && len(entries) > 0 {
		for name := range claudeDisabledServers(env) {
			delete(entries, name)
		}
	}
	return entries, nil
}

func homeEntries(env Env, agent string) (map[string]any, error) {
	entries := map[string]any{}
	sources, err := mcpSources(agent, harnessapi.UserScope, homeSources[agent])
	if err != nil {
		return nil, err
	}
	for _, rel := range sources {
		for name, entry := range entriesInFile(filepath.Join(env.UserHome, filepath.FromSlash(rel))) {
			entries[name] = entry
		}
	}
	return entries, nil
}

func mcpSources(agent string, scope harnessapi.Scope, legacy []string) ([]string, error) {
	adapter, err := harnessregistry.Lookup(harnessapi.ID(agent))
	if refusalIs(err, harnessapi.UnknownID) {
		return legacy, nil
	}
	if err != nil {
		return nil, err
	}
	context, err := harnessapi.NewContext(runtime.GOOS, scope, harnessapi.VerifiedRelease{})
	if err != nil {
		return nil, err
	}
	channel, err := adapter.MCP(context, scope)
	if refusalIs(err, harnessapi.OperationNotPorted) {
		return legacy, nil
	}
	if err != nil {
		return nil, err
	}
	if !channel.Inspect() {
		return nil, nil
	}
	paths := channel.Paths().Paths()
	sources := make([]string, 0, len(paths))
	for _, path := range paths {
		sources = append(sources, path.String())
	}
	return sources, nil
}

func refusalIs(err error, code string) bool {
	var refusal harnessapi.Refusal
	return errors.As(err, &refusal) && refusal.Code() == code
}

func entriesInFile(path string) map[string]any {
	payload, err := os.ReadFile(path) // #nosec G304 -- third-party config surfaces resolved from fixed relative paths
	if err != nil {
		return nil
	}
	base := filepath.Base(path)
	var servers any
	switch {
	case strings.HasSuffix(base, ".toml"):
		var data map[string]any
		if err := toml.Unmarshal(payload, &data); err != nil {
			return nil
		}
		servers = data["mcp_servers"]
	case base == "opencode.json" || base == "opencode.jsonc":
		var data map[string]any
		if err := json.Unmarshal(stripJSONC(payload), &data); err != nil {
			return nil
		}
		raw, _ := data["mcp"].(map[string]any)
		// an entry switched off with "enabled": false does not count
		filtered := map[string]any{}
		for name, entry := range raw {
			if obj, ok := entry.(map[string]any); ok {
				if enabled, present := obj["enabled"]; present && enabled == false {
					continue
				}
			}
			filtered[name] = entry
		}
		servers = filtered
	default:
		var data map[string]any
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil
		}
		servers = data["mcpServers"]
	}
	obj, ok := servers.(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	for name, entry := range obj {
		if name != "" {
			out[name] = entry
		}
	}
	return out
}

func claudeDisabledServers(env Env) map[string]bool {
	disabled := map[string]bool{}
	for _, rel := range claudeSettings {
		payload, err := os.ReadFile(filepath.Join(env.ProjectRoot, filepath.FromSlash(rel))) // #nosec G304
		if err != nil {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal(payload, &data); err != nil {
			continue
		}
		names, _ := data["disabledMcpjsonServers"].([]any)
		for _, item := range names {
			if name, ok := item.(string); ok && name != "" {
				disabled[name] = true
			}
		}
	}
	return disabled
}

// stripJSONC removes // and /* */ comments outside strings: the minimal
// JSONC grammar OpenCode files use.
func stripJSONC(payload []byte) []byte {
	var out []byte
	inString, escaped := false, false
	for i := 0; i < len(payload); i++ {
		c := payload[i]
		if inString {
			out = append(out, c)
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch {
		case c == '"':
			inString = true
			out = append(out, c)
		case c == '/' && i+1 < len(payload) && payload[i+1] == '/':
			for i < len(payload) && payload[i] != '\n' {
				i++
			}
			if i < len(payload) {
				out = append(out, '\n')
			}
		case c == '/' && i+1 < len(payload) && payload[i+1] == '*':
			i += 2
			for i+1 < len(payload) && (payload[i] != '*' || payload[i+1] != '/') {
				i++
			}
			i++
		default:
			out = append(out, c)
		}
	}
	return out
}
