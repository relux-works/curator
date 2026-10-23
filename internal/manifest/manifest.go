// Package manifest parses project Skillfile.json schemas 1, 2 and 3
// declarations and edits legacy schema-1 declarations.
package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SchemaVersion is the legacy Skillfile schema and the required global schema.
const SchemaVersion = 1

// Name is the manifest file name at the project root.
const Name = "Skillfile.json"

// Ref is an exact reference declaration: kind is "tag", "branch", or
// "revision" (branches are allowed at the project level only, Spec §6.1).
type Ref struct {
	Kind  string
	Value string
}

// Decl is one declared skill.
type Decl struct {
	Name   string
	Source string // path under skills_root; defaults to Name
	Ref    Ref
	Git    string // optional clone URL
	// Selector is non-nil only for a schema-2 source selection. Legacy Source,
	// Git and Ref remain empty for that arm.
	Selector *Selector
}

// Manifest is the parsed project manifest.
type Manifest struct {
	Path          string
	SchemaVersion int
	Sources       map[string]Source
	ProjectAlias  string
	Agents        []string
	Locale        string
	Skills        []Decl
	MCPServers    map[string]projectmcp.Declaration
}

// PathIn returns the manifest path for a project root.
func PathIn(projectRoot string) string {
	return filepath.Join(projectRoot, Name)
}

// Load reads and parses the project manifest. A missing file returns
// (nil, nil): the project is simply not initialized.
func Load(projectRoot string) (*Manifest, error) {
	filePath := PathIn(projectRoot)
	payload, err := os.ReadFile(filePath) // #nosec G304 -- path is derived from the project root
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseBytes(payload, filePath)
}

// ParseBytes parses one manifest payload that a caller already read. It is the
// byte-level entry point Load itself is built from, so a caller that has to bind
// a generation digest to the exact bytes its closure was resolved from can read
// the file once and parse those bytes, instead of reading the path a second time
// and parsing a possibly different generation.
func ParseBytes(payload []byte, filePath string) (*Manifest, error) {
	if err := protocoljson.Validate(payload); err != nil {
		if schema3Prefix(payload) {
			return nil, &projectmcp.Error{Code: projectmcp.Invalid, Field: "Skillfile JSON"}
		}

		return nil, fmt.Errorf("malformed JSON in %s: %w", filePath, err)
	}
	// Dispatch with lexical numbers before legacy float conversion. In particular,
	// an oversized ignored option must not leak through encoding/json's error.
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var precise map[string]any
	decodeErr := decoder.Decode(&precise)
	_, numericSchema := precise["schema_version"].(json.Number)
	if decodeErr == nil && numericSchema && schema3Prefix(payload) {
		if !boundedMCPNumbers(precise) {
			return nil, &projectmcp.Error{Code: projectmcp.Invalid, Field: "Skillfile JSON scalars"}
		}
		if _, err := protocoljson.MarshalCanonical(precise); err != nil {
			return nil, &projectmcp.Error{Code: projectmcp.Invalid, Field: "Skillfile JSON scalars"}
		}
		precise["schema_version"] = float64(3)
		return Parse(precise, filePath)
	}
	// Retain the released schema-1/2 decoding and error behavior.
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", filePath, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", filePath)
	}
	return Parse(obj, filePath)
}

// Parse validates a raw manifest object (Spec §6.1).
func Parse(obj map[string]any, filePath string) (*Manifest, error) {
	schema, present := obj["schema_version"]
	if !present {
		return nil, verr.New("schema_version", "missing required field")
	}
	number, ok := schema.(float64)
	if !ok || number != float64(int(number)) {
		if _, extension := obj["mcp_servers"]; extension {
			return nil, &projectmcp.Error{Code: projectmcp.Invalid, Field: "Skillfile schema_version"}
		}
		return nil, verr.New("schema_version", "must be an integer, got %v", schema)
	}
	if number != SchemaVersion && number != 2 && number != 3 {
		return nil, verr.New("schema_version", "unsupported Skillfile schema_version %d; this Skillfile requires a newer tool", int(number))
	}

	allowed := []string{"schema_version", "project", "agents", "locale", "skills"}
	if number >= 2 {
		allowed = append(allowed, "sources")
	}
	if number == 3 {
		allowed = append(allowed, "mcp_servers")
	}
	if unknown := unknownFields(obj, allowed...); len(unknown) > 0 {
		if number == 3 {
			return nil, &projectmcp.Error{Code: projectmcp.Invalid, Field: "Skillfile members"}
		}
		return nil, verr.New("Skillfile", "unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	var sources map[string]Source
	if number >= 2 {
		var err error
		sources, err = parseSources(obj)
		if err != nil {
			return nil, err
		}
	}

	alias, err := parseProjectAlias(obj)
	if err != nil {
		return nil, err
	}

	agents, err := stringList(obj["agents"], "agents")
	if err != nil {
		return nil, err
	}
	seenAgents := map[string]bool{}
	for _, agent := range agents {
		if !identifiers.Valid(agent) || seenAgents[agent] {
			return nil, verr.New("agents", "must contain unique portable identifiers")
		}
		seenAgents[agent] = true
	}

	locale := ""
	if rawLocale, present := obj["locale"]; present && rawLocale != nil {
		locale, ok = rawLocale.(string)
		if !ok || !identifiers.ValidLocale(locale) {
			return nil, verr.New("locale", "must be a 1-64 character ASCII locale selector")
		}
	}

	rawSkills, ok := obj["skills"].([]any)
	if !ok {
		return nil, verr.New("skills", "Skillfile requires field 'skills' as a list")
	}

	var skills []Decl
	seen := map[string]bool{}
	for index, rawEntry := range rawSkills {
		label := fmt.Sprintf("skills[%d]", index)
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		if _, selected := entry["from"]; selected && number >= 2 {
			decl, err := parseSelector(entry, sources, label)
			if err != nil {
				return nil, err
			}
			if decl.Name != "" {
				if seen[decl.Name] {
					return nil, verr.New(label, "source_name_conflict: duplicate skill name %s", decl.Name)
				}
				seen[decl.Name] = true
			}
			skills = append(skills, decl)
			continue
		}
		if unknown := unknownFields(entry, "name", "source", "git", "tag", "branch", "revision"); len(unknown) > 0 {
			return nil, verr.New(label, "unsupported field(s): %s", strings.Join(unknown, ", "))
		}
		name, _ := entry["name"].(string)
		if name == "" {
			return nil, verr.New(label, "requires non-empty string 'name'")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New(label+".name", "skill name %s", identifiers.Rule)
		}
		if seen[name] {
			return nil, verr.New(label+".name", "duplicate skill name in Skillfile: %s", name)
		}
		seen[name] = true

		source := name
		if rawSource, present := entry["source"]; present {
			source, ok = rawSource.(string)
			if !ok || source == "" {
				return nil, verr.New(label+".source", "must be a non-empty string")
			}
			if !identifiers.PortablePath(source) {
				return nil, verr.New(label+".source", "must be a portable relative path")
			}
		}

		git := ""
		if rawGit, present := entry["git"]; present && rawGit != nil {
			git, ok = rawGit.(string)
			if !ok || git == "" {
				return nil, verr.New(label+".git", "must be a non-empty string when present")
			}
		}

		var refKeys []string
		for _, key := range []string{"tag", "branch", "revision"} {
			if _, present := entry[key]; present {
				refKeys = append(refKeys, key)
			}
		}
		if len(refKeys) != 1 {
			return nil, verr.New(label, "must specify exactly one of tag, branch, or revision")
		}
		refValue, _ := entry[refKeys[0]].(string)
		if refValue == "" {
			return nil, verr.New(label+"."+refKeys[0], "must be a non-empty string")
		}

		skills = append(skills, Decl{Name: name, Source: source, Ref: Ref{Kind: refKeys[0], Value: refValue}, Git: git})
	}

	var mcpServers map[string]projectmcp.Declaration
	if number == 3 {
		if raw, present := obj["mcp_servers"]; present {
			mcpServers, err = projectmcp.ParseSet(raw)
			if err != nil {
				return nil, err
			}
		}
	}
	return &Manifest{MCPServers: mcpServers, Path: filePath, SchemaVersion: int(number), Sources: sources, ProjectAlias: alias, Agents: agents, Locale: locale, Skills: skills}, nil
}

func parseProjectAlias(obj map[string]any) (string, error) {
	raw, present := obj["project"]
	if !present || raw == nil {
		return "", nil
	}
	project, ok := raw.(map[string]any)
	if !ok {
		return "", verr.New("project", "must be an object when present")
	}
	if unknown := unknownFields(project, "alias"); len(unknown) > 0 {
		return "", verr.New("project", "unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	rawAlias, present := project["alias"]
	if !present || rawAlias == nil {
		return "", nil
	}
	alias, ok := rawAlias.(string)
	if !ok || alias == "" || utf8.RuneCountInString(alias) > 128 || containsControl(alias) {
		return "", verr.New("project.alias", "must be a non-empty control-free string of at most 128 characters")
	}
	return alias, nil
}

func stringList(raw any, field string) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New(field, "must be a list of strings")
	}
	values := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, verr.New(field, "must be a list of strings")
		}
		values = append(values, text)
	}
	return values, nil
}

// EnsureEmpty creates an empty manifest when none exists and returns its path.
func EnsureEmpty(projectRoot string) (string, error) {
	info, err := os.Stat(projectRoot)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("project path does not exist: %s", projectRoot)
	}
	filePath := PathIn(projectRoot)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}
	payload := map[string]any{"schema_version": SchemaVersion, "agents": []string{}, "skills": []any{}}
	return filePath, writeJSON(filePath, payload)
}

// AddDecl adds or replaces a declaration and validates the result before
// writing (Spec §6.1 add semantics).
func AddDecl(projectRoot, name, refKind, refValue, git, source string) error {
	filePath := PathIn(projectRoot)
	obj, err := readObject(filePath)
	if err != nil {
		return err
	}
	decl := map[string]any{"name": name, refKind: refValue}
	if git != "" {
		decl["git"] = git
	}
	if source != "" {
		decl["source"] = source
	}
	skills, _ := obj["skills"].([]any)
	replaced := false
	for index, rawEntry := range skills {
		if entry, ok := rawEntry.(map[string]any); ok && entry["name"] == name {
			skills[index] = decl
			replaced = true
			break
		}
	}
	if !replaced {
		skills = append(skills, decl)
	}
	obj["skills"] = skills
	if _, err := Parse(obj, filePath); err != nil {
		return err
	}
	return writeJSON(filePath, obj)
}

// RemoveDecl removes a declaration by name.
func RemoveDecl(projectRoot, name string) error {
	filePath := PathIn(projectRoot)
	obj, err := readObject(filePath)
	if err != nil {
		return err
	}
	skills, _ := obj["skills"].([]any)
	kept := make([]any, 0, len(skills))
	for _, rawEntry := range skills {
		if entry, ok := rawEntry.(map[string]any); ok && entry["name"] == name {
			continue
		}
		kept = append(kept, rawEntry)
	}
	if len(kept) == len(skills) {
		return fmt.Errorf("skill not declared in Skillfile: %s", name)
	}
	obj["skills"] = kept
	if _, err := Parse(obj, filePath); err != nil {
		return err
	}
	return writeJSON(filePath, obj)
}

func readObject(filePath string) (map[string]any, error) {
	payload, err := os.ReadFile(filePath) // #nosec G304 -- path is derived from the project root
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s not found at %s; run 'curator init' first", Name, filePath)
	}
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

func unknownFields(object map[string]any, allowed ...string) []string {
	set := map[string]bool{}
	for _, field := range allowed {
		set[field] = true
	}
	var unknown []string
	for field := range object {
		if !set[field] {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func containsControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) {
			return true
		}
	}
	return false
}

func writeJSON(filePath string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, append(data, '\n'), 0o644)
}

// schema3Prefix classifies an already rejected document from a well-formed
// leading schema member. It never admits input or exposes malformed field data.
func schema3Prefix(payload []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return false
	}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return false
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return false
		}
		if key == "schema_version" {
			var version json.Number
			if json.Unmarshal(raw, &version) == nil {
				number, err := version.Float64()
				if err == nil && number == 3 {
					return true
				}
			}
		}
	}
	return false
}

// Reject oversized integer tokens before any numeric conversion. CCJ-1 then
// checks shortest-form integer spelling and the narrower safe-integer range.
func boundedMCPNumbers(value any) bool {
	switch typed := value.(type) {
	case json.Number:
		return len(strings.TrimPrefix(string(typed), "-")) <= 16
	case map[string]any:
		for _, child := range typed {
			if !boundedMCPNumbers(child) {
				return false
			}
		}
	case []any:
		for _, child := range typed {
			if !boundedMCPNumbers(child) {
				return false
			}
		}
	}
	return true
}
