// Package manifest parses the project manifest Skillfile.json, schema 1
// (Spec §6.1), and edits skill declarations in place.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SchemaVersion is the only supported Skillfile schema.
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
}

// Manifest is the parsed project manifest.
type Manifest struct {
	Path         string
	ProjectAlias string
	Agents       []string
	Locale       string
	Skills       []Decl
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
	return Parse(obj, filePath)
}

// Parse validates a raw manifest object (Spec §6.1).
func Parse(obj map[string]any, filePath string) (*Manifest, error) {
	if unknown := unknownFields(obj, "schema_version", "project", "agents", "locale", "skills"); len(unknown) > 0 {
		return nil, verr.New("Skillfile", "unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	schema, present := obj["schema_version"]
	if !present {
		return nil, verr.New("schema_version", "missing required field")
	}
	number, ok := schema.(float64)
	if !ok || number != float64(int(number)) {
		return nil, verr.New("schema_version", "must be an integer, got %v", schema)
	}
	if int(number) != SchemaVersion {
		return nil, verr.New("schema_version", "unsupported Skillfile schema_version %d; this Skillfile requires a newer tool", int(number))
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

	return &Manifest{Path: filePath, ProjectAlias: alias, Agents: agents, Locale: locale, Skills: skills}, nil
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
