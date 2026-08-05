// Package devsub parses Skillfile.dev.json, the non-committed development
// substitution manifest (Spec §6.2).
package devsub

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
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

// BuildRepositorySubstitution is an operator-selected effective source for a
// named repository of one skill. Command, driver, target, output, credentials,
// and signing remain package/manager owned and are intentionally absent.
type BuildRepositorySubstitution struct {
	SkillName      string
	RepositoryName string
	Path           string // absolute local checkout, when local
	Selector       string // normalized project-relative selector, when local
	Git            string
	Identity       string
	Transport      string
	RefKind        string
	RefValue       string
}

// Manifest is the parsed development-substitution document.
type Manifest struct {
	SchemaVersion                int
	Substitutions                map[string]Substitution
	BuildRepositorySubstitutions map[string]map[string]BuildRepositorySubstitution
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
	manifest, err := LoadManifest(projectRoot)
	if err != nil {
		return nil, err
	}
	return manifest.Substitutions, nil
}

// LoadManifest reads both ordinary and schema-2 build-repository
// substitutions. Load remains the byte/behavior-compatible schema-1 facade.
func LoadManifest(projectRoot string) (*Manifest, error) {
	filePath := PathIn(projectRoot)
	payload, err := os.ReadFile(filePath) // #nosec G304 -- path is derived from the project root
	if os.IsNotExist(err) {
		return &Manifest{SchemaVersion: 1, Substitutions: map[string]Substitution{}, BuildRepositorySubstitutions: map[string]map[string]BuildRepositorySubstitution{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseManifestBytes(payload, projectRoot)
}

// ParseBytes parses one substitution payload that a caller already read. It is
// the byte-level entry point Load itself is built from, so a caller that has to
// bind a generation digest to the exact bytes its closure was resolved from can
// read the file once and parse those bytes. The project root stays a parameter
// because a relative substitution path resolves against it.
func ParseBytes(payload []byte, projectRoot string) (map[string]Substitution, error) {
	manifest, err := ParseManifestBytes(payload, projectRoot)
	if err != nil {
		return nil, err
	}
	return manifest.Substitutions, nil
}

// ParseManifestBytes validates substitutions and resolves them below projectRoot.
func ParseManifestBytes(payload []byte, projectRoot string) (*Manifest, error) {
	filePath := PathIn(projectRoot)
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
	schema := 1
	if rawSchema, present := obj["schema_version"]; present {
		number, ok := rawSchema.(float64)
		if !ok || number != 2 {
			return nil, verr.New("schema_version", "must be integer 2 when present")
		}
		schema = 2
	}
	for key := range obj {
		//nolint:staticcheck // This expression mirrors the closed schema alternatives directly.
		if key != "substitutions" && !(schema == 2 && (key == "schema_version" || key == "build_repository_substitutions")) {
			return nil, verr.New(Name, "has unsupported field(s): %s", key)
		}
	}
	rawSubs, present := obj["substitutions"]
	if !present {
		if schema == 2 {
			return nil, verr.New("substitutions", "schema 2 requires an object")
		}
		return &Manifest{SchemaVersion: 1, Substitutions: map[string]Substitution{}, BuildRepositorySubstitutions: map[string]map[string]BuildRepositorySubstitution{}}, nil
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
			if schema == 2 {
				var err error
				text, err = schema2String(entry["path"], 8192, label+".path")
				if err != nil {
					return nil, err
				}
			} else if !ok || text == "" {
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
		if schema == 2 {
			var err error
			git, err = schema2String(entry["git"], 8192, label+".git")
			if err != nil {
				return nil, err
			}
		} else if !ok || git == "" {
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
		if schema == 2 {
			var err error
			value, err = schema2String(ref["value"], 1024, label+".ref.value")
			if err != nil {
				return nil, err
			}
		} else if value == "" {
			return nil, verr.New(label+".ref.value", "must be a non-empty string")
		}
		substitutions[name] = Substitution{SkillName: name, Git: git, RefKind: kind, RefValue: value}
	}
	buildSubstitutions := map[string]map[string]BuildRepositorySubstitution{}
	if schema == 2 {
		rawBuild := obj["build_repository_substitutions"]
		if rawBuild != nil {
			buildObj, ok := rawBuild.(map[string]any)
			if !ok {
				return nil, verr.New("build_repository_substitutions", "must be an object")
			}
			skillNames := sortedKeys(buildObj)
			for _, skillName := range skillNames {
				skillLabel := "build_repository_substitutions." + skillName
				if !identifiers.Valid(skillName) {
					return nil, verr.New(skillLabel, "skill name %s", identifiers.Rule)
				}
				repositories, ok := buildObj[skillName].(map[string]any)
				if !ok || len(repositories) == 0 {
					return nil, verr.New(skillLabel, "must be a non-empty object")
				}
				parsed := map[string]BuildRepositorySubstitution{}
				for _, repositoryName := range sortedKeys(repositories) {
					label := skillLabel + "." + repositoryName
					if !identifiers.Valid(repositoryName) {
						return nil, verr.New(label, "repository name %s", identifiers.Rule)
					}
					entry, ok := repositories[repositoryName].(map[string]any)
					if !ok {
						return nil, verr.New(label, "must be an object")
					}
					parsedEntry, err := parseBuildRepositorySubstitution(entry, projectRoot, skillName, repositoryName, label)
					if err != nil {
						return nil, err
					}
					parsed[repositoryName] = parsedEntry
				}
				buildSubstitutions[skillName] = parsed
			}
		}
	}
	return &Manifest{SchemaVersion: schema, Substitutions: substitutions, BuildRepositorySubstitutions: buildSubstitutions}, nil
}

func schema2String(raw any, maxScalars int, label string) (string, error) {
	value, ok := raw.(string)
	if !ok || value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxScalars {
		return "", verr.New(label, "must be a non-empty string of at most %d Unicode scalars", maxScalars)
	}
	return value, nil
}

func parseBuildRepositorySubstitution(entry map[string]any, projectRoot, skillName, repositoryName, label string) (BuildRepositorySubstitution, error) {
	for key := range entry {
		if key != "path" && key != "git" && key != "ref" {
			return BuildRepositorySubstitution{}, verr.New(label, "has unsupported field(s): %s", key)
		}
	}
	_, hasPath := entry["path"]
	_, hasGit := entry["git"]
	if hasPath == hasGit {
		return BuildRepositorySubstitution{}, verr.New(label, "must declare exactly one of 'path' or 'git'")
	}
	result := BuildRepositorySubstitution{SkillName: skillName, RepositoryName: repositoryName}
	if hasPath {
		if _, hasRef := entry["ref"]; hasRef {
			return result, verr.New(label, "local path substitution must not declare ref")
		}
		rawPath, ok := entry["path"].(string)
		if !ok || rawPath == "" || !utf8.ValidString(rawPath) || utf8.RuneCountInString(rawPath) > 8192 {
			return result, verr.New(label+".path", "must be a non-empty string of at most 8192 Unicode scalars")
		}
		if strings.ContainsRune(rawPath, '\\') || path.IsAbs(rawPath) || hasWindowsVolumePrefix(rawPath) {
			return result, verr.New(label+".path", "must be a project-relative POSIX selector")
		}
		selector, err := buildrepo.NormalizeLocalSelector(rawPath)
		if err != nil {
			return result, verr.New(label+".path", "%v", err)
		}
		resolved := filepath.Join(projectRoot, filepath.FromSlash(selector))
		result.Path, result.Selector = resolved, selector
		return result, nil
	}
	git, ok := entry["git"].(string)
	if !ok {
		return result, verr.New(label+".git", "must be an HTTPS or SSH repository URL")
	}
	source, err := buildrepo.ParseSource(git)
	if err != nil {
		return result, verr.New(label+".git", "%v", err)
	}
	ref, ok := entry["ref"].(map[string]any)
	if !ok {
		return result, verr.New(label+".ref", "requires a structured ref")
	}
	for key := range ref {
		if key != "kind" && key != "value" {
			return result, verr.New(label+".ref", "has unsupported field(s): %s", key)
		}
	}
	kind, _ := ref["kind"].(string)
	value, _ := ref["value"].(string)
	valid := kind == "revision" && (len(value) == 40 || len(value) == 64) && isLowerHex(value) || (kind == "tag" || kind == "branch") && buildrepo.ValidRefName(value)
	if !valid {
		return result, verr.New(label+".ref", "must be a full lowercase revision or safe tag/branch")
	}
	result.Git, result.Identity, result.Transport = git, source.Identity, source.Transport
	result.RefKind, result.RefValue = kind, value
	return result, nil
}

func hasWindowsVolumePrefix(value string) bool {
	if len(value) < 2 || value[1] != ':' {
		return false
	}
	first := value[0]
	return first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z'
}

func isLowerHex(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return value != ""
}

func sortedKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
