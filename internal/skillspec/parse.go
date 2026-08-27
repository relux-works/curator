package skillspec

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/moduleroots"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

var (
	requirementModes    = map[string]bool{"full": true, "runtime": true, "context": true}
	requirementRefKinds = map[string]bool{"tag": true, "revision": true}
	rangeMarkers        = []string{"^", "~", ">", "<", "*", " "}
	mcpTransports       = map[string]bool{"stdio": true, "http": true}
	mcpRequiredIn       = map[string]bool{"any": true, "all": true}
)

// Load reads the skill spec of a snapshot directory. agent-skill.json is
// canonical, csk-skill.json is a legacy read alias, and agents/runtime.json is
// consulted only when neither modern filename exists (Spec §4).
func Load(snapshot string) (*Spec, error) {
	canonicalPath := filepath.Join(snapshot, CanonicalManifestName)
	legacyPath := filepath.Join(snapshot, LegacyManifestName)
	canonicalExists, err := pathExists(canonicalPath)
	if err != nil {
		return nil, err
	}
	legacyExists, err := pathExists(legacyPath)
	if err != nil {
		return nil, err
	}

	if canonicalExists && legacyExists {
		canonical, canonicalData, err := loadSkillManifest(canonicalPath, CanonicalManifestName)
		if err != nil {
			return nil, err
		}
		_, legacyData, err := loadSkillManifest(legacyPath, LegacyManifestName)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(canonicalData, legacyData) {
			return nil, verr.New("", "conflicting_skill_manifests: %s and %s contain different JSON values", CanonicalManifestName, LegacyManifestName)
		}
		return canonical, nil
	}
	if canonicalExists {
		spec, _, loadErr := loadSkillManifest(canonicalPath, CanonicalManifestName)
		return spec, loadErr
	}
	if legacyExists {
		spec, _, loadErr := loadSkillManifest(legacyPath, LegacyManifestName)
		return spec, loadErr
	}
	runtimePath := filepath.Join(snapshot, filepath.FromSlash(RuntimeFallbackName))
	runtimeExists, err := pathExists(runtimePath)
	if err != nil {
		return nil, err
	}
	if runtimeExists {
		return loadRuntimeFallback(runtimePath)
	}
	return &Spec{Commands: map[string]Command{}, Capabilities: capabilities.ImplicitNone()}, nil
}

// ManifestSourcePath returns the protocol path that Load will report for
// diagnostics. It does not parse the selected file.
func ManifestSourcePath(snapshot string) string {
	if exists, _ := pathExists(filepath.Join(snapshot, CanonicalManifestName)); exists {
		return CanonicalManifestName
	}
	if exists, _ := pathExists(filepath.Join(snapshot, LegacyManifestName)); exists {
		return LegacyManifestName
	}
	if exists, _ := pathExists(filepath.Join(snapshot, filepath.FromSlash(RuntimeFallbackName))); exists {
		return RuntimeFallbackName
	}
	return ""
}

// pathExists reports whether a manifest entry exists at path. Only a genuine
// "does not exist" is absence: an entry that exists but cannot be inspected --
// an unreadable directory, a component that is not a directory -- is reported
// as an error, because degrading it to absence would silently hand a snapshot
// the legacy fallback spec or an empty one (Spec §4). Lstat, not Stat, so a
// manifest symlink counts as present even when its target is gone; the parse
// that follows is then what fails, loudly, naming the file.
func pathExists(path string) (bool, error) {
	if _, err := os.Lstat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("cannot determine whether %s exists: %w", path, err)
	}
	return true, nil
}

func loadSkillManifest(filePath, sourceFile string) (*Spec, map[string]any, error) {
	data, err := decodeObject(filePath)
	if err != nil {
		return nil, nil, err
	}
	snapshot := filepath.Dir(filePath)

	schema, err := intField(data, "schema_version")
	if err != nil {
		return nil, nil, verr.New("schema_version", "%s field must be an integer", sourceFile)
	}
	if !SupportedSchemaVersions[schema] {
		return nil, nil, verr.New("schema_version", "unsupported value %d in %s; this skill requires a newer tool. %s", schema, sourceFile, UpgradeHint)
	}

	if schema >= 2 {
		allowed := map[string]bool{"schema_version": true, "runtime_roots": true, "commands": true, "dependencies": true}
		if schema >= 3 {
			allowed["capabilities"] = true
		}
		if schema >= 6 {
			allowed["build_roots"] = true
		}
		if schema >= 7 {
			allowed["build_repositories"] = true
		}
		if err := rejectUnknown(data, allowed, sourceFile); err != nil {
			return nil, nil, err
		}
	}
	if schema >= 3 {
		if _, present := data["capabilities"]; !present {
			return nil, nil, verr.New("capabilities", "%s schema v%d requires 'capabilities'", sourceFile, schema)
		}
	}
	caps := capabilities.ImplicitNone()
	if schema >= 3 {
		caps, err = capabilities.Parse(data["capabilities"])
		if err != nil {
			return nil, nil, err
		}
	}

	var runtimeRoots []string
	if schema >= 2 {
		if raw, present := data["runtime_roots"]; present {
			runtimeRoots, err = parseRuntimeRoots(raw, snapshot)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	var buildRoots []string
	if schema >= 6 {
		if raw, present := data["build_roots"]; present {
			buildRoots, err = parseBuildRoots(raw, snapshot, runtimeRoots)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	buildRepositories, err := parseBuildRepositories(data["build_repositories"], schema)
	if err != nil {
		return nil, nil, err
	}

	commands, err := parseCommands(data["commands"], schema, snapshot, runtimeRoots)
	if err != nil {
		return nil, nil, err
	}
	if schema >= 6 {
		if err := validateBuildLayout(snapshot, buildRoots, commands); err != nil {
			return nil, nil, err
		}
	}
	if schema >= 8 {
		if err := validateModuleRoots(snapshot, buildRoots, runtimeRoots, commands); err != nil {
			return nil, nil, err
		}
	}
	if schema >= 7 {
		if err := validateRepositoryCommands(buildRepositories, commands); err != nil {
			return nil, nil, err
		}
	}

	dependencies, requirements, mcpServers, err := parseDependencies(data["dependencies"], schema)
	if err != nil {
		return nil, nil, err
	}

	return &Spec{
		SchemaVersion:     schema,
		SourceFile:        sourceFile,
		RuntimeRoots:      runtimeRoots,
		BuildRoots:        buildRoots,
		BuildRepositories: buildRepositories,
		Capabilities:      caps,
		Commands:          commands,
		Dependencies:      dependencies,
		Requirements:      requirements,
		McpServers:        mcpServers,
	}, data, nil
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
	parseEntry := func(name string, rawEntry any) error {
		label := "commands." + name
		if name == "" {
			return verr.New("commands", "command names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return verr.New(label, "command name %s", identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return verr.New(label, "must be an object")
		}
		switch entry["type"] {
		case "script":
			if schema >= 2 {
				allowed := map[string]bool{"type": true, "unix_path": true, "win_path": true}
				if schema >= 8 {
					allowed["execution_policy"] = true
					allowed["interpreter"] = true
				}
				if err := rejectUnknown(entry, allowed, label); err != nil {
					return err
				}
			}
			unixPath, unixSet, err := optionalPath(entry, "unix_path", label, schema >= 2)
			if err != nil {
				return err
			}
			winPath, winSet, err := optionalPath(entry, "win_path", label, schema >= 2)
			if err != nil {
				return err
			}
			if schema >= 2 && !unixSet && !winSet {
				return verr.New(label, "script command requires 'unix_path' or 'win_path'")
			}
			if schema >= 2 {
				for field, rel := range map[string]string{"unix_path": unixPath, "win_path": winPath} {
					if rel == "" {
						continue
					}
					if err := validateScriptFile(snapshot, rel, runtimeRoots, label+"."+field); err != nil {
						return err
					}
				}
			}
			policy, interpreter, err := parseScriptExecution(entry, schema, label)
			if err != nil {
				return err
			}
			commands[name] = Command{
				Name: name, Type: "script", UnixPath: unixPath, WinPath: winPath,
				ExecutionPolicy: policy, Interpreter: interpreter,
			}
		case "system":
			if schema >= 2 {
				if err := rejectUnknown(entry, map[string]bool{"type": true, "command": true, "hint": true}, label); err != nil {
					return err
				}
			}
			binary, ok := entry["command"].(string)
			if !ok || binary == "" {
				return verr.New(label, "system command requires non-empty 'command'")
			}
			if schema >= 6 && !identifiers.Valid(binary) {
				return verr.New(label+".command", "system command %s", identifiers.Rule)
			}
			hint, err := optionalString(entry, "hint", label)
			if err != nil {
				return err
			}
			if schema >= 6 {
				if _, present := entry["hint"]; present && hint == "" {
					return verr.New(label+".hint", "must be a non-empty string")
				}
			}
			commands[name] = Command{Name: name, Type: "system", Command: binary, Hint: hint}
		case "build":
			if schema < 6 {
				return verr.New(label, "has unsupported type %v", entry["type"])
			}
			driver, ok := entry["driver"].(string)
			if !ok {
				return verr.New(label+".driver", "must be a supported build driver")
			}
			if driver == "go-repository-v1" {
				if schema < 7 {
					return verr.New(label+".driver", "requires schema_version 7")
				}
				if err := rejectUnknown(entry, map[string]bool{"type": true, "driver": true, "repository": true, "target": true}, label); err != nil {
					return err
				}
				repository, repositoryOK := entry["repository"].(string)
				target, targetOK := entry["target"].(string)
				if !repositoryOK || !identifiers.Valid(repository) {
					return verr.New(label+".repository", "must be a portable repository identifier")
				}
				if !targetOK || !identifiers.Valid(target) {
					return verr.New(label+".target", "must be a portable target identifier")
				}
				commands[name] = Command{Name: name, Type: "build", Driver: driver, Repository: repository, Target: target}
				return nil
			}
			if driver != "go-v1" {
				return verr.New(label+".driver", "must be 'go-v1' or 'go-repository-v1'")
			}
			if err := rejectUnknownBuildFields(entry, schema, label); err != nil {
				return err
			}
			rawSource, present := entry["source_dir"]
			sourceDir, ok := rawSource.(string)
			if !present || !ok || sourceDir == "" {
				return verr.New(label+".source_dir", "must be a non-empty string")
			}
			sourceDir, err := validateRelativePath(sourceDir, label+".source_dir", true)
			if err != nil {
				return err
			}
			modules, err := parseModuleDeclaration(entry, schema, label)
			if err != nil {
				return err
			}
			commands[name] = Command{Name: name, Type: "build", Driver: driver, SourceDir: sourceDir, Modules: modules}
		default:
			return verr.New(label, "has unsupported type %v", entry["type"])
		}
		return nil
	}
	if schema >= 6 {
		names := make([]string, 0, len(obj))
		for name := range obj {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if err := parseEntry(name, obj[name]); err != nil {
				return nil, err
			}
		}
	} else {
		for name, rawEntry := range obj {
			if err := parseEntry(name, rawEntry); err != nil {
				return nil, err
			}
		}
	}
	return commands, nil
}

func parseBuildRepositories(raw any, schema int) (map[string]BuildRepository, error) {
	repositories := map[string]BuildRepository{}
	if raw == nil {
		return repositories, nil
	}
	if schema < 7 {
		return nil, verr.New("build_repositories", "requires schema_version 7")
	}
	obj, ok := raw.(map[string]any)
	if !ok || len(obj) == 0 {
		return nil, verr.New("build_repositories", "must be a non-empty object")
	}
	names := make([]string, 0, len(obj))
	for name := range obj {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		label := "build_repositories." + name
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "repository name %s", identifiers.Rule)
		}
		entry, ok := obj[name].(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		if err := rejectUnknown(entry, map[string]bool{"git": true, "locked_commit": true, "tag": true}, label); err != nil {
			return nil, err
		}
		git, ok := entry["git"].(string)
		if !ok {
			return nil, verr.New(label+".git", "must be an HTTPS or SSH repository URL")
		}
		source, err := buildrepo.ParseSource(git)
		if err != nil {
			return nil, verr.New(label+".git", "%v", err)
		}
		lock, err := buildrepo.ParseLockedCommit(entry["locked_commit"], label+".locked_commit")
		if err != nil {
			return nil, err
		}
		tag := ""
		if rawTag, present := entry["tag"]; present {
			tag, ok = rawTag.(string)
			if !ok || !buildrepo.ValidRefName(tag) {
				return nil, verr.New(label+".tag", "must be a safe Git tag name of at most 255 UTF-8 bytes")
			}
		}
		repositories[name] = BuildRepository{
			Name: name, Git: git, Identity: source.Identity, Transport: source.Transport,
			LockedCommit: LockedCommit{ObjectFormat: lock.ObjectFormat, Hex: lock.Hex}, Tag: tag,
		}
	}
	return repositories, nil
}

func validateRepositoryCommands(repositories map[string]BuildRepository, commands map[string]Command) error {
	used := map[string]bool{}
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		command := commands[name]
		if command.Driver != "go-repository-v1" {
			continue
		}
		if _, present := repositories[command.Repository]; !present {
			return verr.New("commands."+name+".repository", "selects undeclared build repository %q", command.Repository)
		}
		used[command.Repository] = true
	}
	repositoryNames := make([]string, 0, len(repositories))
	for name := range repositories {
		repositoryNames = append(repositoryNames, name)
	}
	sort.Strings(repositoryNames)
	for _, name := range repositoryNames {
		if !used[name] {
			return verr.New("build_repositories."+name, "is not selected by any go-repository-v1 command")
		}
	}
	return nil
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

// parseBuildRoots validates the schema v6 module roots without invoking a Go
// toolchain. Filesystem validation uses lstat on every package-controlled path
// component so links cannot redirect checks outside the immutable snapshot.
func parseBuildRoots(raw any, snapshot string, runtimeRoots []string) ([]string, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("build_roots", "must be a list")
	}
	roots := make([]string, 0, len(list))
	for index, item := range list {
		field := fmt.Sprintf("build_roots[%d]", index)
		text, ok := item.(string)
		if !ok {
			return nil, verr.New(field, "must be a non-empty string")
		}
		root, err := validateRelativePath(text, field, true)
		if err != nil {
			return nil, err
		}
		if err := validateLinkFreeDirectory(snapshot, root, field, "build root"); err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	seen := map[string]bool{}
	for _, root := range roots {
		if seen[root] {
			return nil, verr.New("build_roots", "must be unique")
		}
		seen[root] = true
	}
	if left, right, overlaps := overlappingRoots(roots); overlaps {
		return nil, verr.New("build_roots", "must be disjoint: %s overlaps %s", left, right)
	}
	for _, buildRoot := range roots {
		for _, runtimeRoot := range runtimeRoots {
			if pathContains(buildRoot, runtimeRoot) || pathContains(runtimeRoot, buildRoot) {
				return nil, verr.New("build_roots", "must not overlap runtime_roots: %s overlaps %s", buildRoot, runtimeRoot)
			}
		}
	}
	return roots, nil
}

func overlappingRoots(roots []string) (string, string, bool) {
	sorted := append([]string(nil), roots...)
	sort.Strings(sorted)
	for index, left := range sorted {
		for _, right := range sorted[index+1:] {
			if pathContains(left, right) || pathContains(right, left) {
				return left, right, true
			}
		}
	}
	return "", "", false
}

func rejectUnknownBuildFields(entry map[string]any, schema int, label string) error {
	allowed := map[string]bool{"type": true, "driver": true, "source_dir": true}
	if schema >= 8 {
		allowed["modules"] = true
	}
	unknown := make([]string, 0)
	for field := range entry {
		if !allowed[field] {
			unknown = append(unknown, field)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return verr.New(label+"."+unknown[0], "field is not supported for build commands")
}

// parseScriptExecution reads the schema-8 script execution surface. The
// absence of `execution_policy` is the default and the only spelling of
// declared-only, so null, "none", and every other value is rejected: a
// manifest cannot express declared-only twice. `dependentRequired` binds the
// two fields in both directions -- a command declares both or neither.
//
// Schemas 2 through 7 never reach this function's checks: rejectUnknown has
// already refused both field names for them. Schema 1 does reach it, and must
// leave with nothing -- it keeps its deployed extension tolerance, so an
// unknown command field there is ignored, never read as an enforcement claim.
func parseScriptExecution(entry map[string]any, schema int, label string) (string, string, error) {
	if schema < 8 {
		return "", "", nil
	}
	rawPolicy, policyPresent := entry["execution_policy"]
	rawInterpreter, interpreterPresent := entry["interpreter"]
	if !policyPresent && !interpreterPresent {
		return "", "", nil
	}
	if !policyPresent {
		return "", "", verr.New(label+".interpreter", "requires 'execution_policy'")
	}
	if !interpreterPresent {
		return "", "", verr.New(label+".execution_policy", "requires 'interpreter'")
	}
	policy, ok := rawPolicy.(string)
	if !ok || policy != ScriptExecutionPolicy {
		return "", "", verr.New(label+".execution_policy", "must be %q; omit the field for declared-only execution", ScriptExecutionPolicy)
	}
	interpreter, ok := rawInterpreter.(string)
	if !ok || !ScriptInterpreters[interpreter] {
		return "", "", verr.New(label+".interpreter", "must be one of %s", strings.Join(sortedKeys(ScriptInterpreters), ", "))
	}
	return policy, interpreter, nil
}

// parseModuleDeclaration reads the schema-8 `modules` list of a local go-v1
// build command. An absent list is the default; an empty list is admitted and
// means the same thing. Null is not a spelling of either.
//
// The schema guard is not reachable today -- rejectUnknownBuildFields already
// refuses the field below schema 8, and build commands start at schema 6 --
// but it keeps the band stated where the field is read.
func parseModuleDeclaration(entry map[string]any, schema int, label string) ([]string, error) {
	raw, present := entry["modules"]
	if !present || schema < 8 {
		return nil, nil
	}
	field := label + ".modules"
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New(field, "must be a list of portable relative directory paths")
	}
	modules := make([]string, 0, len(list))
	for index, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, verr.New(fmt.Sprintf("%s[%d]", field, index), "must be a non-empty string")
		}
		modules = append(modules, text)
	}
	return modules, nil
}

// validateModuleRoots performs the declaration and containment half of
// Spec §4.2.3 for every local go-v1 build command. It runs at parse time,
// before the driver's fixed `go list`, which is where the failure boundary of
// that section places it. Form and bijection validation belong to the driver,
// after `go list` returns and before `go build`.
func validateModuleRoots(snapshot string, buildRoots, runtimeRoots []string, commands map[string]Command) error {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		command := commands[name]
		if command.Driver != "go-v1" || len(command.Modules) == 0 {
			continue
		}
		field := "commands." + name + ".modules"
		if err := moduleroots.ValidateDeclaration(snapshot, field, command.Modules, buildRoots, runtimeRoots); err != nil {
			return err
		}
	}
	return nil
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func validateBuildLayout(snapshot string, buildRoots []string, commands map[string]Command) error {
	used := make(map[string]bool, len(buildRoots))
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		command := commands[name]
		if command.Driver != "go-v1" {
			continue
		}
		field := "commands." + name + ".source_dir"
		var containing []string
		for _, root := range buildRoots {
			if pathContains(root, command.SourceDir) {
				containing = append(containing, root)
			}
		}
		if len(containing) != 1 {
			return verr.New(field, "must be below exactly one build_roots entry")
		}
		if err := validateLinkFreeDirectory(snapshot, command.SourceDir, field, "source directory"); err != nil {
			return err
		}
		if err := validateNearestGoMod(snapshot, containing[0], command.SourceDir, field); err != nil {
			return err
		}
		used[containing[0]] = true
	}
	for index, root := range buildRoots {
		if !used[root] {
			return verr.New(fmt.Sprintf("build_roots[%d]", index), "build root is not used by any build command: %s", root)
		}
	}
	return nil
}

func validateLinkFreeDirectory(snapshot, rel, field, noun string) error {
	current := snapshot
	for _, component := range strings.Split(rel, "/") {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return verr.New(field, "%s does not exist: %s", noun, rel)
			}
			return verr.New(field, "cannot inspect %s %s: %v", noun, rel, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return verr.New(field, "%s must be link-free: %s", noun, rel)
		}
		if !info.IsDir() {
			return verr.New(field, "%s must be a directory: %s", noun, rel)
		}
	}
	return nil
}

func validateNearestGoMod(snapshot, buildRoot, sourceDir, field string) error {
	current := sourceDir
	for {
		modPath := filepath.Join(snapshot, filepath.FromSlash(current), "go.mod")
		info, err := os.Lstat(modPath)
		switch {
		case err == nil:
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return verr.New(field, "nearest go.mod must be a real regular file in build root %s", buildRoot)
			}
			if current != buildRoot {
				return verr.New(field, "intervening module %s/go.mod is below build root %s", current, buildRoot)
			}
			return nil
		case !os.IsNotExist(err):
			return verr.New(field, "cannot inspect nearest go.mod: %v", err)
		}
		if current == buildRoot {
			return verr.New(field, "build root %s must contain the nearest go.mod directly", buildRoot)
		}
		current = path.Dir(current)
	}
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
	if !identifiers.PortablePath(value) {
		if strictPosix && (strings.Contains(value, `\`) || strings.Contains(value, "//")) {
			return "", verr.New(field, "must be a POSIX-style relative path inside the skill repository")
		}
		return "", verr.New(field, "must be a portable relative path inside the skill repository")
	}
	return value, nil
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
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", filePath, err)
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
