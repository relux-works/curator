// Package locale analyzes and renders skill localization (Spec §4.3).
//
// The rules apply only when the project selects a locale; with no selected
// locale, skills install without localization checks or rendering.
package locale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Issue is one localization finding.
type Issue struct {
	Severity string // "error" or "warning"
	Code     string
	Path     string
	Message  string
}

// Analysis is the outcome of locale analysis for one snapshot.
type Analysis struct {
	LocaleToRender string
	Issues         []Issue
}

// Failed reports whether any issue is an error.
func (a Analysis) Failed() bool {
	for _, issue := range a.Issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

// Analyze inspects the snapshot against the selected locale.
func Analyze(snapshot, selected string) Analysis {
	if selected == "" {
		return Analysis{}
	}
	metadataPath := filepath.Join(snapshot, "locales", "metadata.json")
	triggersRoot := filepath.Join(snapshot, ".skill_triggers")
	_, metadataErr := os.Stat(metadataPath)
	triggersInfo, triggersErr := os.Stat(triggersRoot)
	if metadataErr != nil && triggersErr != nil {
		return Analysis{} // no localization metadata at all
	}
	if triggersErr == nil && !triggersInfo.IsDir() {
		return Analysis{Issues: []Issue{{
			Severity: "error", Code: "locale.triggers_not_directory", Path: ".skill_triggers",
			Message: "locale trigger catalog must be a directory: " + triggersRoot,
		}}}
	}
	if metadataErr != nil {
		return Analysis{Issues: []Issue{{
			Severity: "error", Code: "locale.metadata_missing", Path: "locales/metadata.json",
			Message: "locale metadata missing: " + metadataPath,
		}}}
	}
	locales, err := readLocales(metadataPath)
	if err != nil {
		return Analysis{Issues: []Issue{{
			Severity: "error", Code: "locale.metadata_invalid", Path: "locales/metadata.json",
			Message: err.Error(),
		}}}
	}

	triggerLocales := map[string]bool{}
	if triggersErr == nil {
		entries, _ := os.ReadDir(triggersRoot)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				triggerLocales[strings.TrimSuffix(entry.Name(), ".md")] = true
			}
		}
	}
	var consistent []string
	for code := range locales {
		if triggerLocales[code] {
			consistent = append(consistent, code)
		}
	}
	sort.Strings(consistent)
	if len(consistent) == 0 {
		return Analysis{Issues: []Issue{{
			Severity: "error", Code: "locale.no_consistent_catalog", Path: "locales/metadata.json",
			Message: "locale metadata and trigger catalogs have no matching supported locale",
		}}}
	}
	for _, code := range consistent {
		if code == selected {
			return Analysis{LocaleToRender: selected}
		}
	}
	return Analysis{Issues: []Issue{{
		Severity: "warning", Code: "locale.selected_unavailable", Path: "locales/metadata.json",
		Message: fmt.Sprintf(
			"locale %q is not fully available; using source SKILL.md without localized rendering. Available locale catalogs: %s",
			selected, strings.Join(consistent, ", ")),
	}}}
}

// Render applies the selected locale to an installed context directory:
// rewrite the SKILL.md frontmatter description and triggers, and rewrite
// agents/openai.yaml when present (Spec §4.3). No-op when analysis selects
// nothing; an error when analysis failed.
func Render(snapshot, installedDir, selected string) ([]Issue, error) {
	analysis := Analyze(snapshot, selected)
	if analysis.Failed() {
		for _, issue := range analysis.Issues {
			if issue.Severity == "error" {
				return analysis.Issues, fmt.Errorf("%s", issue.Message)
			}
		}
	}
	if analysis.LocaleToRender == "" {
		return analysis.Issues, nil
	}
	locales, err := readLocales(filepath.Join(snapshot, "locales", "metadata.json"))
	if err != nil {
		return analysis.Issues, err
	}
	data, present := locales[analysis.LocaleToRender]
	if !present {
		return analysis.Issues, fmt.Errorf("locale %q is not supported by the metadata", analysis.LocaleToRender)
	}
	description, _ := data["description"].(string)
	if strings.TrimSpace(description) != "" {
		triggers := parseTriggers(filepath.Join(snapshot, ".skill_triggers", analysis.LocaleToRender+".md"))
		if err := rewriteFrontmatter(filepath.Join(installedDir, "SKILL.md"), strings.TrimSpace(description), triggers); err != nil {
			return analysis.Issues, err
		}
	}
	openaiPath := filepath.Join(installedDir, "agents", "openai.yaml")
	if _, err := os.Stat(openaiPath); err == nil {
		if err := rewriteOpenAI(openaiPath, data); err != nil {
			return analysis.Issues, err
		}
	}
	return analysis.Issues, nil
}

func readLocales(metadataPath string) (map[string]map[string]any, error) {
	payload, err := os.ReadFile(metadataPath) // #nosec G304 -- path derives from the snapshot
	if err != nil {
		return nil, err
	}
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("malformed locale metadata %s: %v", metadataPath, err)
	}
	var raw struct {
		Locales map[string]any `json:"locales"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed locale metadata %s: %v", metadataPath, err)
	}
	if raw.Locales == nil {
		return nil, fmt.Errorf("locale metadata %s must contain object field 'locales'", metadataPath)
	}
	locales := map[string]map[string]any{}
	for code, value := range raw.Locales {
		entry, ok := value.(map[string]any)
		if !identifiers.ValidLocale(code) || !ok {
			return nil, fmt.Errorf("locale metadata %s has invalid locale entry %q", metadataPath, code)
		}
		locales[code] = entry
	}
	return locales, nil
}

// parseTriggers collects "- phrase" list items outside fenced code blocks.
func parseTriggers(path string) []string {
	payload, err := os.ReadFile(path) // #nosec G304 -- path derives from the snapshot
	if err != nil {
		return nil
	}
	var triggers []string
	inCode := false
	for _, line := range strings.Split(string(payload), "\n") {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		if strings.HasPrefix(stripped, "- ") {
			value := strings.TrimSpace(stripped[2:])
			value = strings.Trim(value, `"'`)
			if value != "" {
				triggers = append(triggers, value)
			}
		}
	}
	return triggers
}

func rewriteFrontmatter(skillPath, description string, triggers []string) error {
	payload, err := os.ReadFile(skillPath) // #nosec G304 -- path derives from the install dir
	if err != nil {
		return err
	}
	lines := strings.Split(string(payload), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	end := -1
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			end = index
			break
		}
	}
	if end < 0 {
		return nil
	}
	var nameLine string
	for _, line := range lines[1:end] {
		if strings.HasPrefix(line, "name:") {
			nameLine = line
			break
		}
	}
	rendered := []string{"---"}
	if nameLine != "" {
		rendered = append(rendered, nameLine)
	}
	rendered = append(rendered, "description: "+quote(description))
	if len(triggers) > 0 {
		rendered = append(rendered, "triggers:")
		for _, trigger := range triggers {
			rendered = append(rendered, "  - "+quote(trigger))
		}
	}
	rendered = append(rendered, "---")
	rendered = append(rendered, lines[end+1:]...)
	return os.WriteFile(skillPath, []byte(strings.Join(rendered, "\n")+"\n"), 0o644)
}

func rewriteOpenAI(path string, data map[string]any) error {
	displayName, _ := data["display_name"].(string)
	shortDescription, _ := data["short_description"].(string)
	defaultPrompt, _ := data["default_prompt"].(string)
	if displayName == "" || shortDescription == "" || defaultPrompt == "" {
		return nil
	}
	content := "interface:\n" +
		"  display_name: " + quote(displayName) + "\n" +
		"  short_description: " + quote(shortDescription) + "\n" +
		"  default_prompt: " + quote(defaultPrompt) + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func quote(value string) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}
