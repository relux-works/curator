package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/codexskillcheck"
)

// Parse the entire recognized handler before inspecting its input. In
// particular, an earlier unsupported option must not hide a later --json.
func (c cli) cmdSkillValidateCodex(args []string) int {
	jsonOut, operandsOnly := false, false
	valid := true
	var dirs []string
	var badOption string
	for _, arg := range args {
		switch {
		case operandsOnly:
			dirs = append(dirs, arg)
		case arg == "--":
			operandsOnly = true
		case arg == "--json":
			jsonOut = true
		case strings.HasPrefix(arg, "-"):
			valid = false
			if badOption == "" {
				badOption = arg
			}
		default:
			dirs = append(dirs, arg)
		}
	}
	result := codexskillcheck.UsageError()
	if valid && len(dirs) == 1 && dirs[0] != "" {
		result = codexskillcheck.Check(dirs[0])
	}
	if jsonOut {
		_ = json.NewEncoder(c.stdout).Encode(result)
	} else if result.ExitCode() == 0 {
		_, _ = fmt.Fprintln(c.stdout, "Codex skill: valid")
	} else {
		messages := map[string]string{
			"usage_invalid":        "expected exactly one directory and optional --json",
			"skill_io":             "cannot read the selected directory's SKILL.md",
			"utf8_invalid":         "SKILL.md must contain valid UTF-8",
			"frontmatter_invalid":  "SKILL.md must start with a closed YAML frontmatter mapping",
			"key_unknown":          "frontmatter contains an unsupported key",
			"name_missing":         "frontmatter requires name",
			"description_missing":  "frontmatter requires description",
			"name_type":            "name must be a string",
			"description_type":     "description must be a string",
			"name_invalid":         "name requires lowercase ASCII letters, digits and single internal hyphens",
			"name_too_long":        "name exceeds 64 code points",
			"description_too_long": "description exceeds 1024 code points",
			"description_todo":     "description starts with an unfinished TODO placeholder",
			"description_angle":    "description contains an angle bracket",
			"body_todo":            "body contains a standalone TODO placeholder outside a code fence",
		}
		_, _ = fmt.Fprintln(c.stderr, "curator skill validate-codex: "+messages[result.Issues[0].Code])
		if badOption != "" {
			_, _ = fmt.Fprintf(c.stderr, "unsupported option: %s\n", badOption)
		}
	}
	return result.ExitCode()
}
