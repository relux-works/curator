package scopes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/verr"
)

// HybridDecl is a hybrid skill declaration with its activation targets.
type HybridDecl struct {
	Decl    manifest.Decl
	Targets []string
}

// HybridManifestPath returns the machine-level hybrid manifest location.
func HybridManifestPath(home string) string {
	return filepath.Join(home, "hybrid", "Skillfile.json")
}

// HybridSkillsRoot returns the hybrid store.
func HybridSkillsRoot(home string) string {
	return filepath.Join(home, "hybrid", "skills")
}

// LoadHybridDecls parses the hybrid manifest: standard schema 1 plus a
// required non-empty per-skill targets list (Spec §9.3).
func LoadHybridDecls(home string) ([]HybridDecl, error) {
	path := HybridManifestPath(home)
	payload, err := os.ReadFile(path) // #nosec G304 -- machine home
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", path, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", path)
	}
	parsed, err := manifest.Parse(obj, path)
	if err != nil {
		return nil, err
	}
	targetsByName, err := hybridTargets(obj)
	if err != nil {
		return nil, err
	}
	var decls []HybridDecl
	for _, decl := range parsed.Skills {
		decls = append(decls, HybridDecl{Decl: decl, Targets: targetsByName[decl.Name]})
	}
	return decls, nil
}

func hybridTargets(obj map[string]any) (map[string][]string, error) {
	targets := map[string][]string{}
	rawSkills, _ := obj["skills"].([]any)
	for index, rawEntry := range rawSkills {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		rawTargets, present := entry["targets"]
		if !present {
			return nil, verr.New(fmt.Sprintf("skills[%d]", index),
				"hybrid skill requires a non-empty 'targets' list (project alias, absolute path, or path glob)")
		}
		list, ok := rawTargets.([]any)
		if !ok || len(list) == 0 {
			return nil, verr.New(fmt.Sprintf("skills[%d].targets", index), "must be a non-empty list of strings")
		}
		var values []string
		for _, item := range list {
			text, ok := item.(string)
			if !ok || text == "" {
				return nil, verr.New(fmt.Sprintf("skills[%d].targets", index), "must be a non-empty list of strings")
			}
			values = append(values, text)
		}
		if name != "" {
			targets[name] = values
		}
	}
	return targets, nil
}

// AppliesToProject reports whether a hybrid declaration targets the project:
// by declared alias, exact resolved path, or path glob (Spec §9.3).
func AppliesToProject(decl HybridDecl, aliases []string, projectPath string) bool {
	resolved, err := filepath.Abs(projectPath)
	if err != nil {
		resolved = projectPath
	}
	resolved = strings.ReplaceAll(resolved, `\`, "/")
	aliasSet := map[string]bool{}
	for _, alias := range aliases {
		if alias != "" {
			aliasSet[alias] = true
		}
	}
	for _, target := range decl.Targets {
		if aliasSet[target] {
			return true
		}
		candidate := strings.ReplaceAll(target, `\`, "/")
		if candidate == resolved {
			return true
		}
		if matched, _ := filepath.Match(candidate, resolved); matched {
			return true
		}
	}
	return false
}

// AddHybridDecl adds or replaces a hybrid declaration with its targets.
func AddHybridDecl(home, name, refKind, refValue, git string, targets []string) error {
	path := HybridManifestPath(home)
	obj := map[string]any{"schema_version": float64(manifest.SchemaVersion), "skills": []any{}}
	if payload, err := os.ReadFile(path); err == nil { // #nosec G304 -- machine home
		var raw any
		if err := json.Unmarshal(payload, &raw); err != nil {
			return fmt.Errorf("malformed JSON in %s: %w", path, err)
		}
		if existing, ok := raw.(map[string]any); ok {
			obj = existing
		}
	}
	entry := map[string]any{"name": name, refKind: refValue, "targets": toAny(targets)}
	if git != "" {
		entry["git"] = git
	}
	skills, _ := obj["skills"].([]any)
	replaced := false
	for index, rawEntry := range skills {
		if existing, ok := rawEntry.(map[string]any); ok && existing["name"] == name {
			skills[index] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		skills = append(skills, entry)
	}
	obj["skills"] = skills
	if _, err := manifest.Parse(obj, path); err != nil {
		return err
	}
	if _, err := hybridTargets(obj); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

// RemoveHybridDecl removes a hybrid declaration by name.
func RemoveHybridDecl(home, name string) error {
	path := HybridManifestPath(home)
	payload, err := os.ReadFile(path) // #nosec G304 -- machine home
	if err != nil {
		return fmt.Errorf("skill not declared in hybrid Skillfile: %s", name)
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return fmt.Errorf("malformed JSON in %s: %w", path, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must contain a JSON object", path)
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
		return fmt.Errorf("skill not declared in hybrid Skillfile: %s", name)
	}
	obj["skills"] = kept
	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func toAny(values []string) []any {
	out := make([]any, len(values))
	for index, value := range values {
		out[index] = value
	}
	return out
}
