// Package contextpkg reads the two package manifests of the agent-environments
// capability — agent-context.json (environments §2, §3) and agent-mcp.json
// (environments §2.2) — under the strict reader discipline of core §1, and
// validates the module files a context package declares.
package contextpkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/pkgversion"
	"github.com/relux-works/curator/internal/protocoljson"
)

// File names fixed by the protocol.
const (
	ManifestName    = "agent-context.json"
	MCPManifestName = "agent-mcp.json"
	ContextDir      = "context"
	InformativeDoc  = "CONTEXT.md"
	SchemaVersion   = 1
	MaxWeight       = 2147483647
)

// Diagnostic codes (environments §2.1, §3.1).
const (
	DiagManifestInvalid            = "context_manifest_invalid"
	DiagMCPInvalid                 = "mcp_declaration_invalid"
	DiagModuleMissing              = "profile_module_missing"
	DiagModuleBytesInvalid         = "profile_module_bytes_invalid"
	DiagSelectorUnknownEnvironment = "profile_selector_unknown_environment"
)

// Diagnostic is one protocol-named failure or warning.
type Diagnostic struct {
	Code   string
	Detail string
}

func (d *Diagnostic) Error() string { return d.Code + ": " + d.Detail }

func invalid(code, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...)}
}

// Requirement is one entry of requires.contexts, requires.skills, or
// requires.mcp. Exactly one of Range, Tag, and Revision is set.
type Requirement struct {
	Name      string
	Git       string
	Range     string
	Tag       string
	Revision  string
	Directory string
	// Weight is the edge weight of a context requirement; nil when absent.
	Weight *int64
	// Mode and Commands keep their core §4.4 meaning on a skill requirement.
	Mode     string
	Commands []string
}

// Form names the requirement's declaration form: "range", "tag", or "revision".
func (r Requirement) Form() string {
	switch {
	case r.Range != "":
		return "range"
	case r.Tag != "":
		return "tag"
	default:
		return "revision"
	}
}

// Module is one entry of context.modules.
type Module struct {
	Path         string
	Environments []string // nil means every environment
	Class        string   // "root" or "system"
}

// Applies reports whether the module's selector admits the environment.
func (m Module) Applies(environment string) bool {
	if m.Environments == nil {
		return true
	}
	for _, id := range m.Environments {
		if id == environment {
			return true
		}
	}
	return false
}

// Manifest is a validated agent-context.json.
type Manifest struct {
	Name    string
	Version string
	Weight  int64
	Weights map[string]int64
	// HasContext is true when the "context" member is present; a package
	// without it is a pure umbrella and declares no root-context surface.
	HasContext bool
	Modules    []Module
	Contexts   map[string]Requirement
	Skills     map[string]Requirement
	MCP        map[string]Requirement
}

// SortedNames returns the keys of a requirement map in bytewise order.
func SortedNames(requirements map[string]Requirement) []string {
	names := make([]string, 0, len(requirements))
	for name := range requirements {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LoadManifest reads and validates the manifest at the package root.
func LoadManifest(root string) (*Manifest, error) {
	payload, err := os.ReadFile(filepath.Join(root, ManifestName)) // #nosec G304 -- package root chosen by the caller
	if err != nil {
		if os.IsNotExist(err) {
			return nil, invalid(DiagManifestInvalid, "%s is absent at %s", ManifestName, root)
		}
		return nil, fmt.Errorf("read %s: %w", ManifestName, err)
	}
	manifest, err := ParseManifest(payload)
	if err != nil {
		return nil, err
	}
	if manifest.HasContext {
		info, err := os.Stat(filepath.Join(root, ContextDir))
		if err != nil || !info.IsDir() {
			return nil, invalid(DiagManifestInvalid, "package %s declares context but has no %s/ directory", manifest.Name, ContextDir)
		}
	}
	return manifest, nil
}

// ParseManifest validates agent-context.json bytes.
func ParseManifest(payload []byte) (*Manifest, error) {
	object, err := decodeObject(payload, DiagManifestInvalid, ManifestName)
	if err != nil {
		return nil, err
	}
	if err := requireKeys(object, DiagManifestInvalid, "", "schema_version", "name", "version", "weight", "weights", "context", "requires"); err != nil {
		return nil, err
	}
	if err := requireSchemaVersion(object, DiagManifestInvalid); err != nil {
		return nil, err
	}
	manifest := &Manifest{}
	if manifest.Name, err = requireIdentifier(object, "name", DiagManifestInvalid); err != nil {
		return nil, err
	}
	if manifest.Version, err = requireVersion(object, DiagManifestInvalid); err != nil {
		return nil, err
	}
	if raw, present := object["weight"]; present {
		weight, ok := weightValue(raw)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "weight must be an integer between 0 and %d", MaxWeight)
		}
		manifest.Weight = weight
	}
	if raw, present := object["weights"]; present {
		weights, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "weights must be an object")
		}
		manifest.Weights = map[string]int64{}
		for name, value := range weights {
			if !identifiers.Valid(name) {
				return nil, invalid(DiagManifestInvalid, "weights key %q is not a portable identifier", name)
			}
			weight, ok := weightValue(value)
			if !ok {
				return nil, invalid(DiagManifestInvalid, "weights.%s must be an integer between 0 and %d", name, MaxWeight)
			}
			manifest.Weights[name] = weight
		}
	}
	if raw, present := object["context"]; present {
		context, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "context must be an object")
		}
		if err := requireKeys(context, DiagManifestInvalid, "context.", "modules"); err != nil {
			return nil, err
		}
		modulesRaw, ok := context["modules"].([]any)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "context.modules must be an array")
		}
		manifest.HasContext = true
		manifest.Modules = []Module{}
		seen := map[string]bool{}
		for index, entry := range modulesRaw {
			module, err := parseModule(entry, index)
			if err != nil {
				return nil, err
			}
			if seen[module.Path] {
				return nil, invalid(DiagManifestInvalid, "context.modules names %q twice", module.Path)
			}
			seen[module.Path] = true
			manifest.Modules = append(manifest.Modules, module)
		}
	}
	if raw, present := object["requires"]; present {
		requires, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "requires must be an object")
		}
		if err := requireKeys(requires, DiagManifestInvalid, "requires.", "contexts", "skills", "mcp"); err != nil {
			return nil, err
		}
		if manifest.Contexts, err = parseRequirements(requires["contexts"], "contexts"); err != nil {
			return nil, err
		}
		if manifest.Skills, err = parseRequirements(requires["skills"], "skills"); err != nil {
			return nil, err
		}
		if manifest.MCP, err = parseRequirements(requires["mcp"], "mcp"); err != nil {
			return nil, err
		}
	}
	return manifest, nil
}

func parseModule(entry any, index int) (Module, error) {
	object, ok := entry.(map[string]any)
	if !ok {
		return Module{}, invalid(DiagManifestInvalid, "context.modules[%d] must be an object", index)
	}
	if err := requireKeys(object, DiagManifestInvalid, fmt.Sprintf("context.modules[%d].", index), "path", "environments", "class"); err != nil {
		return Module{}, err
	}
	path, ok := object["path"].(string)
	if !ok || !identifiers.PortablePath(path) {
		return Module{}, invalid(DiagManifestInvalid, "context.modules[%d].path must be a portable relative path", index)
	}
	module := Module{Path: path, Class: "root"}
	if raw, present := object["environments"]; present {
		selector, err := identifierList(raw, true)
		if err != nil {
			return Module{}, invalid(DiagManifestInvalid, "context.modules[%d].environments %v", index, err)
		}
		module.Environments = selector
	}
	if raw, present := object["class"]; present {
		class, ok := raw.(string)
		if !ok || (class != "root" && class != "system") {
			return Module{}, invalid(DiagManifestInvalid, "context.modules[%d].class must be root or system", index)
		}
		module.Class = class
	}
	return module, nil
}

func parseRequirements(raw any, family string) (map[string]Requirement, error) {
	if raw == nil {
		return nil, nil
	}
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, invalid(DiagManifestInvalid, "requires.%s must be an object", family)
	}
	out := map[string]Requirement{}
	for name, value := range object {
		if !identifiers.Valid(name) {
			return nil, invalid(DiagManifestInvalid, "requires.%s key %q is not a portable identifier", family, name)
		}
		entry, ok := value.(map[string]any)
		if !ok {
			return nil, invalid(DiagManifestInvalid, "requires.%s.%s must be an object", family, name)
		}
		prefix := fmt.Sprintf("requires.%s.%s.", family, name)
		allowed := []string{"git", "range", "tag", "revision"}
		switch family {
		case "contexts":
			allowed = append(allowed, "directory", "weight")
		case "skills":
			allowed = append(allowed, "mode", "commands")
		case "mcp":
			allowed = append(allowed, "directory")
		}
		if err := requireKeys(entry, DiagManifestInvalid, prefix, allowed...); err != nil {
			return nil, err
		}
		requirement := Requirement{Name: name}
		git, ok := entry["git"].(string)
		if !ok || git == "" || len(git) > 8192 {
			return nil, invalid(DiagManifestInvalid, "%sgit must be a non-empty string", prefix)
		}
		requirement.Git = git
		forms := 0
		if raw, present := entry["range"]; present {
			text, ok := raw.(string)
			if !ok {
				return nil, invalid(DiagManifestInvalid, "%srange must be a string", prefix)
			}
			if _, err := pkgversion.ParseRange(text); err != nil {
				return nil, invalid(DiagManifestInvalid, "%srange %q does not parse", prefix, text)
			}
			requirement.Range = text
			forms++
		}
		if raw, present := entry["tag"]; present {
			text, ok := raw.(string)
			if !ok || !ValidGitRefName(text) {
				return nil, invalid(DiagManifestInvalid, "%stag is not a valid git ref name", prefix)
			}
			requirement.Tag = text
			forms++
		}
		if raw, present := entry["revision"]; present {
			text, ok := raw.(string)
			if !ok || !ValidCommit(text) {
				return nil, invalid(DiagManifestInvalid, "%srevision is not a full lowercase commit", prefix)
			}
			requirement.Revision = text
			forms++
		}
		if forms != 1 {
			return nil, invalid(DiagManifestInvalid, "%s must carry exactly one of range, tag, revision", strings.TrimSuffix(prefix, "."))
		}
		if raw, present := entry["directory"]; present {
			text, ok := raw.(string)
			if !ok || !identifiers.PortablePath(text) {
				return nil, invalid(DiagManifestInvalid, "%sdirectory must be a portable relative path", prefix)
			}
			requirement.Directory = text
		}
		if raw, present := entry["weight"]; present {
			weight, ok := weightValue(raw)
			if !ok {
				return nil, invalid(DiagManifestInvalid, "%sweight must be an integer between 0 and %d", prefix, MaxWeight)
			}
			requirement.Weight = &weight
		}
		if raw, present := entry["mode"]; present {
			mode, ok := raw.(string)
			if !ok || (mode != "full" && mode != "runtime" && mode != "context") {
				return nil, invalid(DiagManifestInvalid, "%smode must be full, runtime, or context", prefix)
			}
			requirement.Mode = mode
		}
		if raw, present := entry["commands"]; present {
			commands, err := identifierList(raw, true)
			if err != nil {
				return nil, invalid(DiagManifestInvalid, "%scommands %v", prefix, err)
			}
			if requirement.Mode != "runtime" {
				return nil, invalid(DiagManifestInvalid, "%scommands requires mode runtime", prefix)
			}
			requirement.Commands = commands
		}
		out[name] = requirement
	}
	return out, nil
}

// Server is the server member of agent-mcp.json.
type Server struct {
	Transport    string
	Command      string
	Args         []string
	URL          string
	EnvNames     []string
	Environments []string // nil means every adapter
}

// Applies reports whether the server's selector admits the environment.
func (s Server) Applies(environment string) bool {
	if s.Environments == nil {
		return true
	}
	for _, id := range s.Environments {
		if id == environment {
			return true
		}
	}
	return false
}

// MCPManifest is a validated agent-mcp.json.
type MCPManifest struct {
	Name    string
	Version string
	Server  Server
}

var (
	bareCommandRE = regexp.MustCompile(`^[^/\\\s\x00-\x1f\x7f]+$`)
	httpsURLRE    = regexp.MustCompile(`^https://[A-Za-z0-9][A-Za-z0-9.-]*(?::[0-9]{1,5})?(?:/[^\s?#@]*)?$`)
	// reservedEnvNameRE is the union of every manager-reserved variable
	// (environments §2.2; agent-mcp-v1.schema.json envNames).
	reservedEnvNameRE = regexp.MustCompile(`^(?:PATH|HOME|TMPDIR|TEMP|TMP|XDG_CONFIG_HOME|XDG_CACHE_HOME|XDG_DATA_HOME|XDG_STATE_HOME|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY|FTP_PROXY|NO_PROXY|http_proxy|https_proxy|all_proxy|ftp_proxy|no_proxy|RES_OPTIONS|HOSTALIASES|LOCALDOMAIN|IFS|CSK_PROJECT_ROOT|USERPROFILE|APPDATA|LOCALAPPDATA|PATHEXT|COMSPEC|WINDIR|SYSTEMROOT|__PYVENV_LAUNCHER__|LD_.*|DYLD_.*|PYTHON.*|NODE_.*|NPM_CONFIG_.*)$`)
)

// ReservedEnvName reports whether a variable name is manager-reserved.
func ReservedEnvName(name string) bool { return reservedEnvNameRE.MatchString(name) }

// LoadMCP reads and validates the declaration at the package root.
func LoadMCP(root string) (*MCPManifest, error) {
	payload, err := os.ReadFile(filepath.Join(root, MCPManifestName)) // #nosec G304 -- package root chosen by the caller
	if err != nil {
		if os.IsNotExist(err) {
			return nil, invalid(DiagMCPInvalid, "%s is absent at %s", MCPManifestName, root)
		}
		return nil, fmt.Errorf("read %s: %w", MCPManifestName, err)
	}
	return ParseMCP(payload)
}

// ParseMCP validates agent-mcp.json bytes.
func ParseMCP(payload []byte) (*MCPManifest, error) {
	object, err := decodeObject(payload, DiagMCPInvalid, MCPManifestName)
	if err != nil {
		return nil, err
	}
	if err := requireKeys(object, DiagMCPInvalid, "", "schema_version", "name", "version", "server"); err != nil {
		return nil, err
	}
	if err := requireSchemaVersion(object, DiagMCPInvalid); err != nil {
		return nil, err
	}
	manifest := &MCPManifest{}
	if manifest.Name, err = requireIdentifier(object, "name", DiagMCPInvalid); err != nil {
		return nil, err
	}
	if manifest.Version, err = requireVersion(object, DiagMCPInvalid); err != nil {
		return nil, err
	}
	server, ok := object["server"].(map[string]any)
	if !ok {
		return nil, invalid(DiagMCPInvalid, "server is required and must be an object")
	}
	transport, _ := server["transport"].(string)
	switch transport {
	case "stdio":
		if err := requireKeys(server, DiagMCPInvalid, "server.", "transport", "command", "args", "env_names", "environments"); err != nil {
			return nil, err
		}
		command, ok := server["command"].(string)
		if !ok || command == "" || len(command) > 255 || command == "." || command == ".." || !bareCommandRE.MatchString(command) {
			return nil, invalid(DiagMCPInvalid, "server.command must be a bare executable name without a path separator")
		}
		rawArgs, ok := server["args"].([]any)
		if !ok {
			return nil, invalid(DiagMCPInvalid, "server.args is required and must be an array of strings")
		}
		args := []string{}
		for index, raw := range rawArgs {
			arg, ok := raw.(string)
			if !ok || len(arg) > 8192 {
				return nil, invalid(DiagMCPInvalid, "server.args[%d] must be a string of at most 8192 bytes", index)
			}
			args = append(args, arg)
		}
		manifest.Server = Server{Transport: transport, Command: command, Args: args}
	case "http":
		if err := requireKeys(server, DiagMCPInvalid, "server.", "transport", "url", "env_names", "environments"); err != nil {
			return nil, err
		}
		url, ok := server["url"].(string)
		if !ok || len(url) < 9 || len(url) > 4096 || !httpsURLRE.MatchString(url) {
			return nil, invalid(DiagMCPInvalid, "server.url must be an https URL with an ASCII host and no userinfo, query, or fragment")
		}
		manifest.Server = Server{Transport: transport, URL: url}
	default:
		return nil, invalid(DiagMCPInvalid, "server.transport must be stdio or http")
	}
	if raw, present := server["env_names"]; present {
		names, err := identifierList(raw, false)
		if err != nil {
			return nil, invalid(DiagMCPInvalid, "server.env_names %v", err)
		}
		for _, name := range names {
			if ReservedEnvName(name) {
				return nil, invalid(DiagMCPInvalid, "server.env_names names the manager-reserved variable %s", name)
			}
		}
		manifest.Server.EnvNames = names
	}
	if raw, present := server["environments"]; present {
		selector, err := identifierList(raw, true)
		if err != nil {
			return nil, invalid(DiagMCPInvalid, "server.environments %v", err)
		}
		manifest.Server.Environments = selector
	}
	return manifest, nil
}

// ValidateModules checks every declared module file of a context package at
// root under environments §3: presence, regular-file shape, UTF-8, LF-only
// line endings, and exactly one trailing LF. registered is the adapter
// registry; a selector naming an identifier outside it is a returned warning
// carrying profile_selector_unknown_environment.
func ValidateModules(root string, manifest *Manifest, registered map[string]bool) ([]string, error) {
	var warnings []string
	for _, module := range manifest.Modules {
		path := filepath.Join(root, ContextDir, filepath.FromSlash(module.Path))
		info, err := os.Lstat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return warnings, invalid(DiagModuleMissing, "module %s of %s is absent", module.Path, manifest.Name)
			}
			return warnings, fmt.Errorf("stat module %s of %s: %w", module.Path, manifest.Name, err)
		}
		if !info.Mode().IsRegular() {
			return warnings, invalid(DiagModuleMissing, "module %s of %s is not a regular file", module.Path, manifest.Name)
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- path is a declared module below the package root
		if err != nil {
			return warnings, fmt.Errorf("read module %s of %s: %w", module.Path, manifest.Name, err)
		}
		if err := ValidateModuleBytes(payload); err != nil {
			return warnings, invalid(DiagModuleBytesInvalid, "module %s of %s: %v", module.Path, manifest.Name, err)
		}
		for _, id := range module.Environments {
			if registered != nil && !registered[id] {
				warnings = append(warnings, fmt.Sprintf("%s: module %s of %s selects the unregistered environment %q, which selects nothing", DiagSelectorUnknownEnvironment, module.Path, manifest.Name, id))
			}
		}
	}
	return warnings, nil
}

// ValidateModuleBytes applies the byte rules of environments §3 to one
// module's content.
func ValidateModuleBytes(payload []byte) error {
	if !utf8.Valid(payload) {
		return fmt.Errorf("not valid UTF-8")
	}
	if bytes.IndexByte(payload, '\r') >= 0 {
		return fmt.Errorf("contains a line ending other than LF")
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		return fmt.Errorf("does not end with a trailing LF")
	}
	if len(payload) >= 2 && payload[len(payload)-2] == '\n' {
		return fmt.Errorf("ends with more than one trailing LF")
	}
	return nil
}

// --- shared readers ---------------------------------------------------------

func decodeObject(payload []byte, code, name string) (map[string]any, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, invalid(code, "%s is malformed: %v", name, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, invalid(code, "%s is malformed: %v", name, err)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, invalid(code, "%s must be a JSON object", name)
	}
	return object, nil
}

func requireKeys(object map[string]any, code, prefix string, allowed ...string) error {
	permitted := map[string]bool{}
	for _, key := range allowed {
		permitted[key] = true
	}
	for key := range object {
		if !permitted[key] {
			return invalid(code, "unknown field %s%s", prefix, key)
		}
	}
	return nil
}

func requireSchemaVersion(object map[string]any, code string) error {
	number, ok := object["schema_version"].(json.Number)
	if !ok || number.String() != strconv.Itoa(SchemaVersion) {
		return invalid(code, "schema_version must be %d", SchemaVersion)
	}
	return nil
}

func requireIdentifier(object map[string]any, key, code string) (string, error) {
	value, ok := object[key].(string)
	if !ok || !identifiers.Valid(value) {
		return "", invalid(code, "%s must be a portable identifier", key)
	}
	return value, nil
}

func requireVersion(object map[string]any, code string) (string, error) {
	value, ok := object["version"].(string)
	if !ok {
		return "", invalid(code, "version must be a string")
	}
	if _, err := pkgversion.ParseVersion(value); err != nil {
		return "", invalid(code, "version %q is not a strict semantic version", value)
	}
	return value, nil
}

func weightValue(raw any) (int64, bool) {
	number, ok := raw.(json.Number)
	if !ok {
		return 0, false
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || value < 0 || value > MaxWeight {
		return 0, false
	}
	return value, true
}

func identifierList(raw any, nonEmpty bool) ([]string, error) {
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("must be an array of identifiers")
	}
	if nonEmpty && len(items) == 0 {
		return nil, fmt.Errorf("must not be empty")
	}
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		value, ok := item.(string)
		if !ok || !identifiers.Valid(value) {
			return nil, fmt.Errorf("carries a value that is not a portable identifier")
		}
		if seen[value] {
			return nil, fmt.Errorf("repeats %q", value)
		}
		seen[value] = true
		out = append(out, value)
	}
	return out, nil
}

var commitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)

// ValidCommit reports whether value is a full lowercase commit object id.
func ValidCommit(value string) bool { return commitRE.MatchString(value) }

// ValidGitRefName applies the core §6.3 tag grammar (git-check-ref-format
// without the leading refs/ component).
func ValidGitRefName(value string) bool {
	if value == "" || len(value) > 255 || value == "@" {
		return false
	}
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.HasSuffix(value, ".") {
		return false
	}
	if strings.Contains(value, "//") || strings.Contains(value, "..") || strings.Contains(value, "@{") {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r == 0x7f || strings.ContainsRune(`~^:?*[\`, r) {
			return false
		}
	}
	for _, component := range strings.Split(value, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}
