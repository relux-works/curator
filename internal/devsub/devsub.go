// Package devsub parses Skillfile.dev.json, the non-committed development
// substitution manifest (Spec §6.2).
package devsub

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/verr"
)

// Name is the dev manifest file name next to Skillfile.json.
const Name = "Skillfile.dev.json"

var refKinds = map[string]bool{"tag": true, "revision": true, "branch": true}

// Substitution replaces a provider: exactly one of Path (a local checkout)
// or Git plus a ref. Branches are allowed here by design.
type Substitution struct {
	SkillName string
	Path      string // absolute after resolution
	Git       string
	RefKind   string
	RefValue  string
}

// Describe renders the substitution target for install output and markers.
func (s Substitution) Describe() string {
	if s.Path != "" {
		return "path " + s.Path
	}
	return fmt.Sprintf("git %s %s %s", s.Git, s.RefKind, s.RefValue)
}

// PathIn returns the dev manifest path for a project root.
func PathIn(projectRoot string) string {
	return filepath.Join(projectRoot, Name)
}

// Load reads the substitutions; a missing file yields an empty map.
func Load(projectRoot string) (map[string]Substitution, error) {
	filePath := PathIn(projectRoot)
	payload, err := os.ReadFile(filePath) // #nosec G304 -- path is derived from the project root
	if os.IsNotExist(err) {
		return map[string]Substitution{}, nil
	}
	if err != nil {
		return nil, err
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", filePath, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", filePath)
	}
	for key := range obj {
		if key != "substitutions" {
			return nil, verr.New(Name, "has unsupported field(s): %s", key)
		}
	}
	rawSubs, present := obj["substitutions"]
	if !present {
		return map[string]Substitution{}, nil
	}
	subsObj, ok := rawSubs.(map[string]any)
	if !ok {
		return nil, verr.New("substitutions", "must be an object")
	}

	substitutions := map[string]Substitution{}
	for name, rawEntry := range subsObj {
		label := "substitutions." + name
		if name == "" {
			return nil, verr.New("substitutions", "substitution names must be non-empty strings")
		}
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "substitution name %s", identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		for key := range entry {
			if key != "path" && key != "git" && key != "ref" {
				return nil, verr.New(label, "has unsupported field(s): %s", key)
			}
		}
		_, hasPath := entry["path"]
		_, hasGit := entry["git"]
		if hasPath == hasGit {
			return nil, verr.New(label, "must declare exactly one of 'path' or 'git'")
		}

		if hasPath {
			text, ok := entry["path"].(string)
			if !ok || text == "" {
				return nil, verr.New(label+".path", "must be a non-empty string")
			}
			if _, hasRef := entry["ref"]; hasRef {
				return nil, verr.New(label, "with 'path' reads the local checkout; 'ref' does not apply")
			}
			resolved := text
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(projectRoot, resolved)
			}
			substitutions[name] = Substitution{SkillName: name, Path: resolved}
			continue
		}

		git, ok := entry["git"].(string)
		if !ok || git == "" {
			return nil, verr.New(label+".git", "must be a non-empty string")
		}
		ref, ok := entry["ref"].(map[string]any)
		if !ok {
			return nil, verr.New(label, "with 'git' requires a 'ref' object")
		}
		for key := range ref {
			if key != "kind" && key != "value" {
				return nil, verr.New(label+".ref", "has unsupported field(s): %s", key)
			}
		}
		kind, _ := ref["kind"].(string)
		if !refKinds[kind] {
			return nil, verr.New(label+".ref.kind", "must be one of tag, revision, or branch")
		}
		value, _ := ref["value"].(string)
		if value == "" {
			return nil, verr.New(label+".ref.value", "must be a non-empty string")
		}
		substitutions[name] = Substitution{SkillName: name, Git: git, RefKind: kind, RefValue: value}
	}
	return substitutions, nil
}
