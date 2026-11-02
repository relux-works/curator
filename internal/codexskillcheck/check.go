// Package codexskillcheck validates the bounded Codex SKILL.md contract.
// Package-manifest and installation policy validation belong to skillcheck.
package codexskillcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Issue is a stable diagnostic code, without input values or filesystem data.
type Issue struct {
	Code string `json:"code"`
}

// Result is the versioned, single-result validation interface.
type Result struct {
	SchemaVersion int     `json:"schema_version"`
	Contract      string  `json:"contract"`
	Outcome       string  `json:"outcome"`
	Issues        []Issue `json:"issues"`
}

func result(outcome, code string) Result {
	r := Result{SchemaVersion: 1, Contract: "codex-quick-validate-v1", Outcome: outcome, Issues: []Issue{}}
	if code != "" {
		r.Issues = append(r.Issues, Issue{Code: code})
	}
	return r
}

// UsageError reports invalid command arguments, independently of skill content.
func UsageError() Result { return result("usage_error", "usage_invalid") }

// ExitCode separates content validity, usage failure and unavailable input.
func (r Result) ExitCode() int {
	switch r.Outcome {
	case "valid":
		return 0
	case "invalid":
		return 1
	case "usage_error":
		return 2
	default:
		return 3
	}
}

// Check reads only the explicitly selected directory's SKILL.md. It performs
// no configuration loading, dependency resolution, download or input writes.
func Check(dir string) Result {
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md")) // #nosec G304 -- SKILL.md is an explicit caller-selected input; validation performs no writes.
	if err != nil {
		return result("setup_error", "skill_io")
	}
	return Validate(data)
}

var (
	namePattern = regexp.MustCompile(`^[a-z0-9-]+$`)
	// Python's source \d means Unicode decimal digits; Go's \d is ASCII.
	fencePattern = regexp.MustCompile("^[ \\t]*(?:(?:[-+*]|\\p{Nd}+[.)])[ \\t]+)?(`{3,}|~{3,})(.*)$")
	todoPattern  = regexp.MustCompile(`^[ ]{0,3}\[TODO:[^\n]*\][ \t]*$`)
)

// Validate checks the bounded frontmatter and placeholder rules. Empty required
// strings, duplicate/non-string keys, optional-value types and broader YAML
// compatibility have no additional guarantees in this version.
func Validate(data []byte) Result {
	invalid := func(code string) Result { return result("invalid", code) }
	if !utf8.Valid(data) {
		return invalid("utf8_invalid")
	}
	lines := strings.Split(string(data), "\n")
	if strings.TrimSuffix(lines[0], "\r") != "---" {
		return invalid("frontmatter_invalid")
	}
	end := 1
	for end < len(lines) && strings.TrimSuffix(lines[end], "\r") != "---" {
		end++
	}
	if end == len(lines) {
		return invalid("frontmatter_invalid")
	}
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &document); err != nil || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return invalid("frontmatter_invalid")
	}
	fields := make(map[string]*yaml.Node)
	mapping := document.Content[0]
	for i := 0; i < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		switch key.Value {
		case "name", "description", "license", "allowed-tools", "metadata":
			fields[key.Value] = value
		default:
			return invalid("key_unknown")
		}
	}
	nameNode, ok := fields["name"]
	if !ok {
		return invalid("name_missing")
	}
	if nameNode.Kind != yaml.ScalarNode || nameNode.Tag != "!!str" {
		return invalid("name_type")
	}
	name := strings.TrimSpace(nameNode.Value)
	if !namePattern.MatchString(name) || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
		return invalid("name_invalid")
	}
	if utf8.RuneCountInString(name) > 64 {
		return invalid("name_too_long")
	}
	descriptionNode, ok := fields["description"]
	if !ok {
		return invalid("description_missing")
	}
	if descriptionNode.Kind != yaml.ScalarNode || descriptionNode.Tag != "!!str" {
		return invalid("description_type")
	}
	description := strings.TrimSpace(descriptionNode.Value)
	if utf8.RuneCountInString(description) > 1024 {
		return invalid("description_too_long")
	}
	if strings.HasPrefix(description, "[TODO:") {
		return invalid("description_todo")
	}
	if strings.ContainsAny(description, "<>") {
		return invalid("description_angle")
	}
	var marker byte
	var markerLength int
	for _, raw := range lines[end+1:] {
		line := strings.TrimSuffix(raw, "\r")
		if match := fencePattern.FindStringSubmatch(line); match != nil {
			if marker == 0 {
				marker, markerLength = match[1][0], len(match[1])
			} else if match[1][0] == marker && len(match[1]) >= markerLength && strings.TrimSpace(match[2]) == "" {
				marker, markerLength = 0, 0
			}
			continue
		}
		if marker == 0 && todoPattern.MatchString(line) {
			return invalid("body_todo")
		}
	}
	return result("valid", "")
}
