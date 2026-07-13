// Package skillcheck validates one skill package without a consuming
// project (Spec §15, curator skill check).
package skillcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/locale"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
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
	spec, err := skillspec.Load(skillDir)
	if err != nil {
		issues = append(issues, Issue{
			Severity: "error", Code: "skill.manifest_invalid",
			Path: "csk-skill.json", Message: err.Error(),
		})
	} else {
		issues = append(issues, runtimeRootReferenceWarnings(skillDir, spec)...)
		issues = append(issues, commandResolutionWarnings(skillDir, spec)...)
	}
	analysis := locale.Analyze(skillDir, localeValue)
	for _, item := range analysis.Issues {
		issues = append(issues, Issue{Severity: item.Severity, Code: item.Code, Path: item.Path, Message: item.Message})
	}
	return issues
}

func runtimeRootReferenceWarnings(skillDir string, spec *skillspec.Spec) []Issue {
	hasProviderRuntime := false
	for _, dependency := range spec.Dependencies {
		if dependency.Type == "skill" {
			hasProviderRuntime = true
			break
		}
	}
	if !hasProviderRuntime {
		for _, requirement := range spec.Requirements {
			if requirement.Mode == "full" || requirement.Mode == "runtime" {
				hasProviderRuntime = true
				break
			}
		}
	}
	if len(spec.RuntimeRoots) == 0 && !hasProviderRuntime {
		return nil
	}

	var issues []Issue
	for _, path := range promptMarkdownFiles(skillDir) {
		payload, err := os.ReadFile(path) // #nosec G304 -- path is below the validated skill directory
		if err != nil {
			continue
		}
		text := string(payload)
		for _, runtimeRoot := range spec.RuntimeRoots {
			windowsRoot := strings.ReplaceAll(runtimeRoot, "/", `\`)
			tokens := []string{runtimeRoot + "/", windowsRoot + `\`}
			for _, token := range tokens {
				if !strings.Contains(text, token) {
					continue
				}
				relative, _ := filepath.Rel(skillDir, path)
				issues = append(issues, Issue{
					Severity: "warning",
					Code:     "skill.runtime_root_in_prompt_context",
					Path:     filepath.ToSlash(relative),
					Message: fmt.Sprintf(
						"prompt-visible text references runtime-only path %q; Curator removes that root from installed skill context. Use exported command shims for installed execution and keep manifest-relative paths source-checkout-only",
						token,
					),
				})
				break
			}
		}
		if hasProviderRuntime && (strings.Contains(text, "scripts/") || strings.Contains(text, `scripts\`)) {
			relative, _ := filepath.Rel(skillDir, path)
			issues = append(issues, Issue{
				Severity: "warning",
				Code:     "skill.provider_runtime_path_in_prompt_context",
				Path:     filepath.ToSlash(relative),
				Message:  "prompt-visible text references a source scripts path while this skill consumes another skill's runtime. Resolve the provider's exported command shim instead of guessing its source layout",
			})
		}
	}
	return issues
}

func commandResolutionWarnings(skillDir string, spec *skillspec.Spec) []Issue {
	hasScriptCommands := false
	hasWindowsCommand := false
	for _, command := range spec.Commands {
		if command.Type != "script" {
			continue
		}
		hasScriptCommands = true
		if command.WinPath != "" {
			hasWindowsCommand = true
		}
	}
	hasProviderCommands := false
	for _, dependency := range spec.Dependencies {
		if dependency.Type == "skill" {
			hasProviderCommands = true
			break
		}
	}
	if !hasProviderCommands {
		for _, requirement := range spec.Requirements {
			if requirement.Mode == "full" || requirement.Mode == "runtime" {
				hasProviderCommands = true
				break
			}
		}
	}
	if !hasScriptCommands && !hasProviderCommands {
		return nil
	}

	var content strings.Builder
	for _, path := range promptMarkdownFiles(skillDir) {
		payload, err := os.ReadFile(path) // #nosec G304 -- path is below the validated skill directory
		if err == nil {
			content.Write(payload)
			content.WriteByte('\n')
		}
	}
	text := content.String()
	var missing []string
	if !strings.Contains(text, ".agents/bin") && !strings.Contains(text, `.agents\bin`) {
		missing = append(missing, "project .agents/bin lookup")
	}
	if !strings.Contains(text, "global/bin") && !strings.Contains(text, `global\bin`) {
		missing = append(missing, "manager global/bin fallback")
	}
	if !strings.Contains(text, "command -v") || !strings.Contains(text, "Get-Command") {
		missing = append(missing, "validated POSIX and PowerShell bare-command fallbacks")
	}
	if hasWindowsCommand && !strings.Contains(text, ".cmd") {
		missing = append(missing, "Windows .cmd shim suffix")
	}
	if len(missing) == 0 {
		return nil
	}
	return []Issue{{
		Severity: "warning",
		Code:     "skill.command_resolution_contract_missing",
		Path:     "SKILL.md",
		Message: "prompt-visible instructions export managed runtime commands but do not document a shell-neutral resolver (" +
			strings.Join(missing, ", ") +
			"). Agents must resolve project shims first, then global shims, then a validated bare command; shell profile activation is optional",
	}}
}

func promptMarkdownFiles(skillDir string) []string {
	seen := map[string]bool{}
	var paths []string
	for _, root := range whitelist.IncludeRoots {
		candidate := filepath.Join(skillDir, root)
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			if strings.EqualFold(filepath.Ext(candidate), ".md") && !seen[candidate] {
				seen[candidate] = true
				paths = append(paths, candidate)
			}
			continue
		}
		_ = filepath.WalkDir(candidate, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") || seen[path] {
				return nil
			}
			seen[path] = true
			paths = append(paths, path)
			return nil
		})
	}
	sort.Strings(paths)
	return paths
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
