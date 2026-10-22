// Package projectmcp implements the separately versioned Skillfile MCP stage-A
// declaration model. Parsing never grants inline approval or executes a server.
package projectmcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	Capability        = "skillfile-mcp-stage-a-v1"
	Invalid           = "mcp_declaration_invalid"
	InlineUnsupported = "mcp_inline_approval_unsupported"
)

// Error carries a stable refusal class and a non-secret field description.
type Error struct{ Code, Field string }

func (e *Error) Error() string   { return e.Code + ": " + e.Field }
func invalid(field string) error { return &Error{Invalid, field} }

// Declaration retains normalized JSON-domain data, including ignored options.
// Its package arm has no inline server until exact committed bytes are resolved.
type Declaration struct {
	Server  map[string]any
	Package map[string]any
	Install map[string]any
}

var nameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)
var envRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var urlRE = regexp.MustCompile(`^https://[A-Za-z0-9][A-Za-z0-9.-]*(?::[0-9]{1,5})?(?:/[^\s?#@]*)?$`)
var npmRE = regexp.MustCompile(`^(?:@([a-z0-9][a-z0-9._-]*)/)?([a-z0-9][a-z0-9._-]*)@((?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*))$`)
var uvRE = regexp.MustCompile(`^([a-z0-9](?:[a-z0-9._-]*[a-z0-9])?)(@|==)((?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*))$`)

// Environments is the persisted environment vocabulary, not CLI aliases.
var environments = []string{"claude_code", "codex_cli", "muse", "opencode", "pi"}

// SetError retains only independently valid declarations for refusal reporting.
// Its valid siblings never authorize partial admission after a failed parse.
type SetError struct {
	Cause error
	Valid map[string]Declaration
}

func (e *SetError) Error() string { return e.Cause.Error() }
func (e *SetError) Unwrap() error { return e.Cause }

// ParseSet recognizes all three arms. On failure, no declaration set is
// admitted; valid siblings remain available solely through the typed error.
func ParseSet(raw any) (map[string]Declaration, error) {
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, invalid("mcp_servers must be an object")
	}
	out := map[string]Declaration{}
	var first error
	for _, name := range sortedKeys(object) {
		var decl Declaration
		var err error
		if !nameRE.MatchString(name) {
			err = invalid("mcp_servers NAME")
		} else {
			decl, err = ParseDeclaration(object[name])
		}
		if err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		out[name] = decl
	}
	if first != nil {
		return nil, &SetError{Cause: first, Valid: out}
	}
	return out, nil
}

// ParseDeclaration applies semantic validation before computing any identity.
func ParseDeclaration(raw any) (Declaration, error) {
	if text, ok := raw.(string); ok {
		raw = map[string]any{"server": map[string]any{"transport": "http", "url": text}}
	}
	obj, ok := raw.(map[string]any)
	if !ok || !closed(obj, "server", "package", "install") {
		return Declaration{}, invalid("declaration arm")
	}
	_, server := obj["server"]
	_, pkg := obj["package"]
	if server == pkg {
		return Declaration{}, invalid("exactly one declaration arm required")
	}
	options := map[string]any{}
	if raw, present := obj["install"]; present {
		var valid bool
		options, valid = raw.(map[string]any)
		if !valid {
			return Declaration{}, invalid("install must be an object")
		}
	}
	if err := auditValue(options, "install"); err != nil {
		return Declaration{}, err
	}
	if _, err := protocoljson.MarshalCanonical(options); err != nil {
		return Declaration{}, invalid("install JSON scalars")
	}
	decl := Declaration{Install: options}
	if server {
		normalized, err := NormalizeServer(obj["server"])
		if err != nil {
			return Declaration{}, err
		}
		decl.Server = normalized
	} else {
		selector, ok := obj["package"].(map[string]any)
		if !ok {
			return Declaration{}, invalid("package must be an object")
		}
		// Reuse the released requirement validator, including range/tag/directory
		// semantics, without extending that released grammar.
		payload, err := json.Marshal(map[string]any{"schema_version": 1, "name": "project-mcp", "version": "1.0.0", "requires": map[string]any{"mcp": map[string]any{"server": selector}}})
		if err != nil {
			return Declaration{}, invalid("package JSON")
		}
		if _, err = contextpkg.ParseManifest(payload); err != nil {
			return Declaration{}, invalid("package selector")
		}
		git, _ := selector["git"].(string)
		if _, err := identity.Parse(git); err != nil || identity.ClassifySource(git) != identity.SourceGit {
			return Declaration{}, invalid("package.git")
		}
		decl.Package = selector
	}
	return decl, nil
}

// NormalizeServer validates the unchanged closed transport arms plus the new
// lexical/pin/selector rules. Only the two set-valued fields are sorted.
func NormalizeServer(raw any) (map[string]any, error) {
	server, ok := raw.(map[string]any)
	if !ok {
		return nil, invalid("server must be an object")
	}
	transport, _ := server["transport"].(string)
	normalized := map[string]any{"transport": transport}
	switch transport {
	case "http":
		if !closed(server, "transport", "url", "env_names", "environments") {
			return nil, invalid("http server members")
		}
		value, ok := server["url"].(string)
		if !ok || !validURL(value) {
			return nil, invalid("server.url")
		}
		normalized["url"] = value
	case "stdio":
		if !closed(server, "transport", "command", "args", "env_names", "environments") {
			return nil, invalid("stdio server members")
		}
		command, ok := server["command"].(string)
		if !ok || !bare(command) {
			return nil, invalid("server.command")
		}
		rawArgs, ok := stringArray(server["args"])
		if !ok {
			return nil, invalid("server.args")
		}
		args := make([]string, 0, len(rawArgs))
		for _, raw := range rawArgs {
			arg, ok := raw.(string)
			if !ok || !utf8.ValidString(arg) || utf8.RuneCountInString(arg) > 8192 {
				return nil, invalid("server.args")
			}
			args = append(args, arg)
		}
		if !validPin(command, args) {
			return nil, invalid("server runtime pin")
		}
		normalized["command"], normalized["args"] = command, args
	default:
		return nil, invalid("server.transport")
	}
	names := []string{}
	if raw, present := server["env_names"]; present {
		var err error
		names, err = stringSet(raw, false, func(s string) bool { return envRE.MatchString(s) && !contextpkg.ReservedEnvName(s) })
		if err != nil {
			return nil, invalid("server.env_names")
		}
	}
	normalized["env_names"] = names
	if raw, present := server["environments"]; present {
		ids, err := stringSet(raw, true, func(s string) bool {
			for _, id := range environments {
				if id == s {
					return true
				}
			}
			return false
		})
		if err != nil {
			return nil, invalid("server.environments")
		}
		normalized["environments"] = ids
	}
	if err := auditValue(normalized, "server url/args"); err != nil {
		return nil, err
	}
	return normalized, nil
}

func validURL(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}
	if n := utf8.RuneCountInString(value); n < 9 || n > 4096 {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || strings.ContainsRune("${}\\", r) {
			return false
		}
	}
	return urlRE.MatchString(value)
}
func bare(value string) bool {
	if !utf8.ValidString(value) || value == "" || value == "." || value == ".." || utf8.RuneCountInString(value) > 255 {
		return false
	}
	for _, r := range value {
		if r == '/' || r == '\\' || unicode.IsSpace(r) || r < 32 || r == 127 {
			return false
		}
	}
	return true
}
func validPin(command string, args []string) bool {
	lower := strings.ToLower(command)
	for _, runtime := range []string{"npx", "uvx"} {
		if lower == runtime || strings.HasPrefix(lower, runtime+".") {
			if command != runtime {
				return false
			}
		}
	}
	switch command {
	case "npx":
		if len(args) > 0 && (args[0] == "-y" || args[0] == "--yes") {
			args = args[1:]
		}
		if len(args) == 0 {
			return false
		}
		m := npmRE.FindStringSubmatch(args[0])
		return m != nil && len(m[1]) <= 214 && len(m[2]) <= 214
	case "uvx":
		if len(args) == 0 {
			return false
		}
		separator := "@"
		operand := args[0]
		if operand == "--from" {
			if len(args) < 3 || !bare(args[2]) || strings.HasPrefix(args[2], "-") {
				return false
			}
			operand = args[1]
			separator = "=="
		}
		m := uvRE.FindStringSubmatch(operand)
		return m != nil && len(m[1]) <= 214 && m[2] == separator
	}
	return true
}
func closed(obj map[string]any, keys ...string) bool {
	for key := range obj {
		found := false
		for _, allowed := range keys {
			if key == allowed {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// stringArray accepts JSON arrays and the typed representation produced by
// normalization, allowing validation to be idempotent without scalar coercion.
func stringArray(raw any) ([]any, bool) {
	switch values := raw.(type) {
	case []any:
		return values, true
	case []string:
		out := make([]any, len(values))
		for i, value := range values {
			out[i] = value
		}
		return out, true
	default:
		return nil, false
	}
}
func stringSet(raw any, nonempty bool, valid func(string) bool) ([]string, error) {
	values, ok := stringArray(raw)
	if !ok || (nonempty && len(values) == 0) {
		return nil, invalid("string set")
	}
	seen := map[string]bool{}
	out := []string{}
	for _, raw := range values {
		value, ok := raw.(string)
		if !ok || seen[value] || !valid(value) {
			return nil, invalid("string set")
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}
func sortedKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func auditValue(value any, field string) error {
	switch v := value.(type) {
	case string:
		report := contextaudit.DetectFilesAtVersion(map[string][]byte{contextpkg.MCPManifestName: []byte(v)}, "", hashing.VersionV2, nil)
		if report.Blocking() {
			return &Error{contextaudit.ClassSecretMaterial, field}
		}
	case map[string]any:
		for key, item := range v {
			if err := auditValue(key, field); err != nil {
				return err
			}
			if err := auditValue(item, field); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range v {
			if err := auditValue(item, field); err != nil {
				return err
			}
		}
	case []string:
		for _, item := range v {
			if err := auditValue(item, field); err != nil {
				return err
			}
		}
	}
	return nil
}

// Model preserves every option, including names such as sig, in the identity.
func (d Declaration) Model() map[string]any {
	obj := map[string]any{"install": d.Install}
	if d.Package != nil {
		obj["package"] = d.Package
	} else {
		obj["server"] = d.Server
	}
	return obj
}

// Digest hashes exactly the supplied JSON-domain model with CCJ-1 and no LF.
func Digest(value any) (string, error) {
	payload, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// Names returns deterministic declaration order.
func Names(decls map[string]Declaration) []string {
	names := make([]string, 0, len(decls))
	for name := range decls {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RequirePackage refuses recognized inline declarations without consulting any
// repository-supplied install option or purported approval record.
func (d Declaration) RequirePackage() error {
	if d.Package == nil {
		return &Error{InlineUnsupported, "inline project contribution"}
	}
	return nil
}

// OptionNames names ignored options without exposing their values.
func (d Declaration) OptionNames() []string { return sortedKeys(d.Install) }

// ValidName reports whether a case-sensitive stage-A NAME is portable.
func ValidName(name string) bool { return nameRE.MatchString(name) }

// ValidEnvironment recognizes persisted registry IDs, never CLI aliases.
func ValidEnvironment(name string) bool {
	for _, id := range environments {
		if id == name {
			return true
		}
	}
	return false
}
