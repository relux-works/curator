package skillspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/verr"
)

var (
	requirementModes    = map[string]bool{"full": true, "runtime": true, "context": true}
	requirementRefKinds = map[string]bool{"tag": true, "revision": true}
	rangeMarkers        = []string{"^", "~", ">", "<", "*", " "}
	mcpTransports       = map[string]bool{"stdio": true, "http": true}
	mcpRequiredIn       = map[string]bool{"any": true, "all": true}
)

// Load reads the skill spec of a snapshot directory: csk-skill.json first,
// the legacy agents/runtime.json second, a pure context skill otherwise
// (Spec §5, §5.10).
func Load(snapshot string) (*Spec, error) {
	cskPath := filepath.Join(snapshot, "csk-skill.json")
	if _, err := os.Stat(cskPath); err == nil {
		return loadCskSkill(cskPath)
	}
	runtimePath := filepath.Join(snapshot, "agents", "runtime.json")
	if _, err := os.Stat(runtimePath); err == nil {
		return loadRuntimeFallback(runtimePath)
	}
	return &Spec{Commands: map[string]Command{}, Capabilities: capabilities.ImplicitNone()}, nil
}

func loadCskSkill(filePath string) (*Spec, error) {
	data, err := decodeObject(filePath)
	if err != nil {
		return nil, err
	}
	snapshot := filepath.Dir(filePath)

	schema, err := intField(data, "schema_version")
	if err != nil {
		return nil, verr.New("schema_version", "must be an integer")
	}
	if !SupportedSchemaVersions[schema] {
		return nil, verr.New("schema_version", "unsupported value %d; this skill requires a newer tool. %s", schema, UpgradeHint)
	}

	if schema >= 2 {
		allowed := map[string]bool{"schema_version": true, "runtime_roots": true, "commands": true, "dependencies": true}
		if schema >= 3 {
			allowed["capabilities"] = true
		}
		if err := rejectUnknown(data, allowed, "csk-skill.json"); err != nil {
			return nil, err
		}
	}
	if schema >= 3 {
		if _, present := data["capabilities"]; !present {
			return nil, verr.New("capabilities", "csk-skill.json schema v%d requires 'capabilities'", schema)
		}
	}
	caps := capabilities.ImplicitNone()
	if schema >= 3 {
		caps, err = capabilities.Parse(data["capabilities"])
		if err != nil {
			return nil, err
		}
	}

	var runtimeRoots []string
	if schema >= 2 {
		if raw, present := data["runtime_roots"]; present {
			runtimeRoots, err = parseRuntimeRoots(raw, snapshot)
			if err != nil {
				return nil, err
			}
		}
	}

	commands, err := parseCommands(data["commands"], schema, snapshot, runtimeRoots)
	if err != nil {
		return nil, err
	}

	dependencies, requirements, mcpServers, err := parseDependencies(data["dependencies"], schema)
	if err != nil {
		return nil, err
	}

	return &Spec{
		SchemaVersion: schema,
		SourceFile:    "csk-skill.json",
		RuntimeRoots:  runtimeRoots,
		Capabilities:  caps,
		Commands:      commands,
		Dependencies:  dependencies,
		Requirements:  requirements,
		McpServers:    mcpServers,
	}, nil
}

func loadRuntimeFallback(filePath string) (*Spec, error) {
	data, err := decodeObject(filePath)
	if err != nil {
		return nil, err
	}
	rawCommands, _ := data["commands"].(map[string]any)
	if data["commands"] != nil && rawCommands == nil {
		return nil, verr.New("commands", "agents/runtime.json field 'commands' must be an object")
	}
	commands := map[string]Command{}
	for name, rawPath := range rawCommands {
		if name == "" {
			return nil, verr.New("commands", "runtime command names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New("commands."+name, "runtime command name %s", identifiers.Rule)
		}
		rel, ok := rawPath.(string)
		if !ok || rel == "" {
			return nil, verr.New("commands."+name, "path must be a non-empty string")
		}
		if _, err := validateRelativePath(rel, "commands."+name, false); err != nil {
			return nil, err
		}
		command := Command{Name: name, Type: "script", UnixPath: rel}
		if strings.HasSuffix(rel, ".cmd") {
			command.WinPath = rel
		}
		commands[name] = command
	}
	return &Spec{
		SchemaVersion: 1,
		SourceFile:    "agents/runtime.json",
		Capabilities:  capabilities.ImplicitNone(),
		Commands:      commands,
	}, nil
}

func parseCommands(raw any, schema int, snapshot string, runtimeRoots []string) (map[string]Command, error) {
	if raw == nil {
		return map[string]Command{}, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("commands", "must be an object")
	}
	commands := map[string]Command{}
	for name, rawEntry := range obj {
		label := "commands." + name
		if name == "" {
			return nil, verr.New("commands", "command names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "command name %s", identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		switch entry["type"] {
		case "script":
			if schema >= 2 {
				if err := rejectUnknown(entry, map[string]bool{"type": true, "unix_path": true, "win_path": true}, label); err != nil {
					return nil, err
				}
			}
			unixPath, unixSet, err := optionalPath(entry, "unix_path", label, schema >= 2)
			if err != nil {
				return nil, err
			}
			winPath, winSet, err := optionalPath(entry, "win_path", label, schema >= 2)
			if err != nil {
				return nil, err
			}
			if schema >= 2 && !unixSet && !winSet {
				return nil, verr.New(label, "script command requires 'unix_path' or 'win_path'")
			}
			if schema >= 2 {
				for field, rel := range map[string]string{"unix_path": unixPath, "win_path": winPath} {
					if rel == "" {
						continue
					}
					if err := validateScriptFile(snapshot, rel, runtimeRoots, label+"."+field); err != nil {
						return nil, err
					}
				}
			}
			commands[name] = Command{Name: name, Type: "script", UnixPath: unixPath, WinPath: winPath}
		case "system":
			if schema >= 2 {
				if err := rejectUnknown(entry, map[string]bool{"type": true, "command": true, "hint": true}, label); err != nil {
					return nil, err
				}
			}
			binary, ok := entry["command"].(string)
			if !ok || binary == "" {
				return nil, verr.New(label, "system command requires non-empty 'command'")
			}
			hint, err := optionalString(entry, "hint", label)
			if err != nil {
				return nil, err
			}
			commands[name] = Command{Name: name, Type: "system", Command: binary, Hint: hint}
		default:
			return nil, verr.New(label, "has unsupported type %v", entry["type"])
		}
	}
	return commands, nil
}

func parseDependencies(raw any, schema int) (map[string]CommandDependency, map[string]Requirement, map[string]McpServer, error) {
	empty := func() (map[string]CommandDependency, map[string]Requirement, map[string]McpServer, error) {
		return map[string]CommandDependency{}, map[string]Requirement{}, map[string]McpServer{}, nil
	}
	if raw == nil {
		return empty()
	}
	if schema < 2 {
		return nil, nil, nil, verr.New("dependencies", "requires schema_version 2 or newer")
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, nil, nil, verr.New("dependencies", "must be an object")
	}
	if _, present := obj["skills"]; present && schema < 4 {
		return nil, nil, nil, verr.New("dependencies.skills", "requires schema_version 4")
	}
	if _, present := obj["mcp_servers"]; present && schema < 5 {
		return nil, nil, nil, verr.New("dependencies.mcp_servers", "requires schema_version 5")
	}
	allowed := map[string]bool{"commands": true}
	if schema >= 4 {
		allowed["skills"] = true
	}
	if schema >= 5 {
		allowed["mcp_servers"] = true
	}
	if err := rejectUnknown(obj, allowed, "dependencies"); err != nil {
		return nil, nil, nil, err
	}

	requirements, err := parseRequirements(obj["skills"])
	if err != nil {
		return nil, nil, nil, err
	}
	mcpServers, err := parseMcpServers(obj["mcp_servers"])
	if err != nil {
		return nil, nil, nil, err
	}

	dependencies := map[string]CommandDependency{}
	if rawCommands := obj["commands"]; rawCommands != nil {
		commandsObj, ok := rawCommands.(map[string]any)
		if !ok {
			return nil, nil, nil, verr.New("dependencies.commands", "must be an object")
		}
		for name, rawEntry := range commandsObj {
			label := "dependencies.commands." + name
			if name == "" {
				return nil, nil, nil, verr.New("dependencies.commands", "dependency command names must be non-empty strings")
			}
			if !identifiers.Valid(name) {
				return nil, nil, nil, verr.New(label, "dependency command name %s", identifiers.Rule)
			}
			entry, ok := rawEntry.(map[string]any)
			if !ok {
				return nil, nil, nil, verr.New(label, "must be an object")
			}
			hint, err := optionalString(entry, "hint", label)
			if err != nil {
				return nil, nil, nil, err
			}
			switch entry["type"] {
			case "system":
				if err := rejectUnknown(entry, map[string]bool{"type": true, "command": true, "hint": true}, label); err != nil {
					return nil, nil, nil, err
				}
				binary, ok := entry["command"].(string)
				if !ok || binary == "" {
					return nil, nil, nil, verr.New(label, "system dependency requires non-empty 'command'")
				}
				dependencies[name] = CommandDependency{Name: name, Type: "system", Command: binary, Hint: hint}
			case "skill":
				if err := rejectUnknown(entry, map[string]bool{"type": true, "skill": true, "command": true, "hint": true}, label); err != nil {
					return nil, nil, nil, err
				}
				skill, ok := entry["skill"].(string)
				if !ok || skill == "" || !identifiers.Valid(skill) {
					return nil, nil, nil, verr.New(label, "skill dependency requires a valid 'skill' name (%s)", identifiers.Rule)
				}
				command, ok := entry["command"].(string)
				if !ok || command == "" || !identifiers.Valid(command) {
					return nil, nil, nil, verr.New(label, "skill dependency requires a valid 'command' name (%s)", identifiers.Rule)
				}
				dependencies[name] = CommandDependency{Name: name, Type: "skill", Skill: skill, Command: command, Hint: hint}
			default:
				return nil, nil, nil, verr.New(label, "has unsupported type %v", entry["type"])
			}
		}
	}
	return dependencies, requirements, mcpServers, nil
}

func parseRequirements(raw any) (map[string]Requirement, error) {
	requirements := map[string]Requirement{}
	if raw == nil {
		return requirements, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("dependencies.skills", "must be an object")
	}
	for name, rawEntry := range obj {
		label := "dependencies.skills." + name
		if name == "" {
			return nil, verr.New("dependencies.skills", "skill requirement names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "skill requirement name %s", identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		if _, present := entry["version"]; present {
			return nil, verr.New(label, "declares 'version'; version ranges are not supported. Pin an exact ref: {\"kind\": \"tag\" | \"revision\", \"value\": ...}")
		}
		if err := rejectUnknown(entry, map[string]bool{"git": true, "ref": true, "mode": true, "commands": true}, label); err != nil {
			return nil, err
		}

		git, ok := entry["git"].(string)
		if !ok || git == "" {
			return nil, verr.New(label, "requires a non-empty 'git' source URL")
		}

		ref, ok := entry["ref"].(map[string]any)
		if !ok {
			return nil, verr.New(label, "requires a 'ref' object with 'kind' and 'value'")
		}
		if err := rejectUnknown(ref, map[string]bool{"kind": true, "value": true}, label+".ref"); err != nil {
			return nil, err
		}
		kind, _ := ref["kind"].(string)
		if kind == "branch" {
			return nil, verr.New(label+".ref", "pins a branch; skill requirements accept exact 'tag' or 'revision' refs only")
		}
		if !requirementRefKinds[kind] {
			return nil, verr.New(label+".ref.kind", "must be 'tag' or 'revision'")
		}
		value, _ := ref["value"].(string)
		if value == "" {
			return nil, verr.New(label+".ref.value", "must be a non-empty string")
		}
		for _, marker := range rangeMarkers {
			if strings.Contains(value, marker) {
				return nil, verr.New(label+".ref.value", "%q looks like a version range; skill requirements accept exact tags or revisions only", value)
			}
		}

		mode := "full"
		if rawMode, present := entry["mode"]; present {
			mode, _ = rawMode.(string)
			if !requirementModes[mode] {
				return nil, verr.New(label+".mode", "must be one of full, runtime, or context")
			}
		}

		var commands []string
		if rawCommands, present := entry["commands"]; present {
			if mode != "runtime" {
				return nil, verr.New(label+".commands", "applies to runtime requirements only")
			}
			list, ok := rawCommands.([]any)
			if !ok || len(list) == 0 {
				return nil, verr.New(label+".commands", "must be a non-empty list of command names")
			}
			seen := map[string]bool{}
			for _, item := range list {
				command, ok := item.(string)
				if !ok || command == "" {
					return nil, verr.New(label+".commands", "entries must be non-empty strings")
				}
				if !identifiers.Valid(command) {
					return nil, verr.New(label+".commands", "entry %q %s", command, identifiers.Rule)
				}
				if !seen[command] {
					seen[command] = true
					commands = append(commands, command)
				}
			}
		}

		requirements[name] = Requirement{Name: name, Git: git, RefKind: kind, RefValue: value, Mode: mode, Commands: commands}
	}
	return requirements, nil
}

func parseMcpServers(raw any) (map[string]McpServer, error) {
	servers := map[string]McpServer{}
	if raw == nil {
		return servers, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("dependencies.mcp_servers", "must be an object")
	}
	for name, rawEntry := range obj {
		label := "dependencies.mcp_servers." + name
		if name == "" {
			return nil, verr.New("dependencies.mcp_servers", "MCP server names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "MCP server name %s", identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		if err := rejectUnknown(entry, map[string]bool{"hint": true, "transport": true, "required_in": true}, label); err != nil {
			return nil, err
		}
		hint, ok := entry["hint"].(string)
		if !ok || hint == "" {
			return nil, verr.New(label, "requires a non-empty 'hint' describing how to connect the server")
		}
		transport := ""
		if rawTransport, present := entry["transport"]; present && rawTransport != nil {
			transport, _ = rawTransport.(string)
			if !mcpTransports[transport] {
				return nil, verr.New(label+".transport", "must be 'stdio' or 'http'")
			}
		}
		requiredIn := "any"
		if rawRequired, present := entry["required_in"]; present {
			requiredIn, _ = rawRequired.(string)
			if !mcpRequiredIn[requiredIn] {
				return nil, verr.New(label+".required_in", "must be 'any' or 'all'")
			}
		}
		servers[name] = McpServer{Name: name, Hint: hint, Transport: transport, RequiredIn: requiredIn}
	}
	return servers, nil
}

// parseRuntimeRoots validates runtime_roots per Spec §5.3: POSIX-relative,
// existing directories, unique, pairwise disjoint.
func parseRuntimeRoots(raw any, snapshot string) ([]string, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("runtime_roots", "must be a list")
	}
	roots := make([]string, 0, len(list))
	for index, item := range list {
		field := fmt.Sprintf("runtime_roots[%d]", index)
		text, ok := item.(string)
		if !ok {
			return nil, verr.New(field, "must be a non-empty string")
		}
		root, err := validateRelativePath(text, field, true)
		if err != nil {
			return nil, err
		}
		info, statErr := os.Stat(filepath.Join(snapshot, filepath.FromSlash(root)))
		if statErr != nil {
			return nil, verr.New(field, "runtime root does not exist: %s", root)
		}
		if !info.IsDir() {
			return nil, verr.New(field, "runtime root must be a directory: %s", root)
		}
		roots = append(roots, root)
	}
	seen := map[string]bool{}
	for _, root := range roots {
		if seen[root] {
			return nil, verr.New("runtime_roots", "must be unique after normalization")
		}
		seen[root] = true
	}
	sorted := append([]string(nil), roots...)
	sort.Strings(sorted)
	for i, left := range sorted {
		for _, right := range sorted[i+1:] {
			if pathContains(left, right) || pathContains(right, left) {
				container, contained := left, right
				if pathContains(right, left) {
					container, contained = right, left
				}
				return nil, verr.New("runtime_roots", "must be disjoint: %s contains %s", container, contained)
			}
		}
	}
	return roots, nil
}

func validateScriptFile(snapshot, rel string, runtimeRoots []string, label string) error {
	info, err := os.Stat(filepath.Join(snapshot, filepath.FromSlash(rel)))
	if err != nil {
		return verr.New(label, "source file not found: %s", rel)
	}
	if info.IsDir() {
		return verr.New(label, "must point to a file: %s", rel)
	}
	if len(runtimeRoots) > 0 {
		for _, root := range runtimeRoots {
			if pathContains(root, rel) {
				return nil
			}
		}
		return verr.New(label, "command path %q is not inside any runtime_roots", rel)
	}
	return nil
}

// validateRelativePath enforces the relative path rules of Spec §5.4/§5.3.
// strictPosix additionally rejects backslashes, doubled slashes, and "."
// segments.
func validateRelativePath(value, field string, strictPosix bool) (string, error) {
	if value == "" {
		return "", verr.New(field, "must be a non-empty string")
	}
	normalized := strings.TrimRight(value, "/")
	if normalized == "" {
		return "", verr.New(field, "must be a non-empty string")
	}
	if strictPosix {
		if strings.Contains(normalized, `\`) || strings.Contains(normalized, "//") {
			return "", verr.New(field, "must be a POSIX-style relative path inside the skill repository")
		}
		for _, part := range strings.Split(normalized, "/") {
			if part == "" || part == "." {
				return "", verr.New(field, "must be a POSIX-style relative path inside the skill repository")
			}
		}
	}
	clean := path.Clean(normalized)
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") || clean == "." {
		return "", verr.New(field, "must be a relative path inside the skill repository")
	}
	for _, part := range strings.Split(clean, "/") {
		if part == ".." {
			return "", verr.New(field, "must be a relative path inside the skill repository")
		}
	}
	return clean, nil
}

func pathContains(root, rel string) bool {
	rootParts := strings.Split(root, "/")
	relParts := strings.Split(rel, "/")
	if len(relParts) < len(rootParts) {
		return false
	}
	for i, part := range rootParts {
		if relParts[i] != part {
			return false
		}
	}
	return true
}

func decodeObject(filePath string) (map[string]any, error) {
	payload, err := os.ReadFile(filePath) // #nosec G304 -- manifest path is derived from the snapshot directory
	if err != nil {
		return nil, err
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", filePath, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", filePath)
	}
	return obj, nil
}

// intField extracts an integer field, rejecting booleans and fractions.
func intField(obj map[string]any, key string) (int, error) {
	raw, present := obj[key]
	if !present {
		return 0, fmt.Errorf("missing %s", key)
	}
	number, ok := raw.(float64)
	if !ok || number != float64(int(number)) {
		return 0, fmt.Errorf("%s is not an integer", key)
	}
	return int(number), nil
}

func optionalString(entry map[string]any, key, label string) (string, error) {
	raw, present := entry[key]
	if !present || raw == nil {
		return "", nil
	}
	text, ok := raw.(string)
	if !ok {
		return "", verr.New(label+"."+key, "must be a string")
	}
	return text, nil
}

func optionalPath(entry map[string]any, key, label string, strictPosix bool) (string, bool, error) {
	raw, present := entry[key]
	if !present || raw == nil {
		return "", false, nil
	}
	text, ok := raw.(string)
	if !ok {
		return "", false, verr.New(label+"."+key, "must be a non-empty string")
	}
	clean, err := validateRelativePath(text, label+"."+key, strictPosix)
	if err != nil {
		return "", false, err
	}
	return clean, true, nil
}

func rejectUnknown(obj map[string]any, allowed map[string]bool, label string) error {
	var unknown []string
	for key := range obj {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return verr.New(label, "has unsupported field(s): %s", strings.Join(unknown, ", "))
}
