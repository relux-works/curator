// Package skillcheck validates one skill package without a consuming
// project (Spec §15, curator skill check).
package skillcheck

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/locale"
	"github.com/relux-works/curator/internal/skillspec"
)

// Issue is one validation finding.
type Issue struct {
	Severity string `json:"severity"` // error | warning
	Code     string `json:"code"`
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

// Validate checks a skill directory: SKILL.md presence, manifest parsing
// with every Spec §5 rule, and locale consistency when a locale is given.
func Validate(skillDir, localeValue string) []Issue {
	var issues []Issue
	if info, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil || info.IsDir() {
		issues = append(issues, Issue{
			Severity: "error", Code: "skill.missing_skill_md",
			Path: "SKILL.md", Message: "required SKILL.md not found",
		})
	}
	if _, err := skillspec.Load(skillDir); err != nil {
		issues = append(issues, Issue{
			Severity: "error", Code: "skill.manifest_invalid",
			Path: "csk-skill.json", Message: err.Error(),
		})
	}
	analysis := locale.Analyze(skillDir, localeValue)
	for _, item := range analysis.Issues {
		issues = append(issues, Issue{Severity: item.Severity, Code: item.Code, Path: item.Path, Message: item.Message})
	}
	return issues
}

// HasErrors reports whether any issue is an error.
func HasErrors(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

// Format renders one issue for terminal output.
func Format(issue Issue) string {
	location := ""
	if issue.Path != "" {
		location = " (" + issue.Path + ")"
	}
	return fmt.Sprintf("%s: %s%s: %s", issue.Severity, issue.Code, location, issue.Message)
}
