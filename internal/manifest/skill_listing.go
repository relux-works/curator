package manifest

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"gopkg.in/yaml.v3"
)

// ParseSkillListing shares install's frontmatter rules while retaining the
// body offset in the original bytes. Decoded listing text is distinct from
// both raw YAML and the body; CRLF never changes source byte coordinates.
func ParseSkillListing(data []byte) (name, description string, bodyStart int, err error) {
	var frontmatter []byte
	closed := false
	for offset, line := 0, 0; offset < len(data); line++ {
		end := bytes.IndexByte(data[offset:], '\n')
		if end < 0 {
			end = len(data)
		} else {
			end += offset + 1
		}
		text := bytes.TrimSuffix(bytes.TrimSuffix(data[offset:end], []byte("\n")), []byte("\r"))
		if line == 0 {
			if string(text) != "---" {
				return "", "", 0, fmt.Errorf("missing YAML frontmatter")
			}
		} else if string(text) == "---" {
			bodyStart = end
			closed = true
			break
		} else {
			frontmatter = append(frontmatter, text...)
			frontmatter = append(frontmatter, '\n')
		}
		offset = end
	}
	if !closed {
		return "", "", 0, fmt.Errorf("unterminated YAML frontmatter")
	}
	frontmatter = bytes.TrimSuffix(frontmatter, []byte("\n"))
	var metadata map[string]any
	if err := yaml.Unmarshal(frontmatter, &metadata); err != nil {
		return "", "", 0, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	name, ok := metadata["name"].(string)
	if !ok || !identifiers.Valid(name) {
		return "", "", 0, fmt.Errorf("invalid SKILL.md name")
	}
	description, ok = metadata["description"].(string)
	if !ok || strings.TrimSpace(description) == "" {
		return "", "", 0, fmt.Errorf("invalid SKILL.md description")
	}
	if raw, ok := metadata["triggers"]; ok {
		triggers, ok := raw.([]any)
		if !ok {
			return "", "", 0, fmt.Errorf("invalid triggers")
		}
		for _, raw := range triggers {
			value, ok := raw.(string)
			if !ok || strings.TrimSpace(value) == "" {
				return "", "", 0, fmt.Errorf("invalid trigger")
			}
		}
	}
	return name, description, bodyStart, nil
}
