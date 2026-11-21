package claudememory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/protocoljson"
)

// Ownership records only patterns appended by Curator, never pre-existing user
// entries. Settings and this record must be published in one locked transaction.
type Ownership struct {
	Version       int      `json:"version"`
	OwnedPatterns []string `json:"owned_patterns"`
}

// Reconcile refreshes dynamic ancestor exclusions without accumulating paths
// from previous projects. No record means all existing entries are user-owned.
// Operator-native memory/rules remain excluded independently of project root.
// Returned bytes are plans only: errors never authorize either file to change.
func Reconcile(payload, record []byte, nativeFile string, ancestors []string) ([]byte, []byte, error) {
	object, original, err := exclusionSettings(payload)
	if err != nil {
		return nil, nil, err
	}
	ownership, err := readOwnership(record)
	if err != nil {
		return nil, nil, err
	}
	excludes := append([]string{}, original...)
	for _, owned := range ownership.OwnedPatterns {
		for i, pattern := range excludes {
			if pattern == owned {
				excludes = append(excludes[:i], excludes[i+1:]...)
				break // Only the one Curator-owned occurrence is removed.
			}
		}
	}
	if !filepath.IsAbs(nativeFile) {
		return nil, nil, fmt.Errorf("native Claude memory path must be absolute")
	}
	nativeCanonical, err := canonicalMemoryPath(nativeFile)
	if err != nil {
		return nil, nil, err
	}
	rules := filepath.Join(filepath.Dir(nativeFile), "rules")
	rulesCanonical, err := canonicalMemoryPath(rules)
	if err != nil {
		return nil, nil, err
	}
	contains := func(pattern string) bool {
		for _, present := range excludes {
			if present == pattern {
				return true
			}
		}
		return false
	}
	// These two operator patterns are the explicit exception to the project
	// boundary. Neither this exception nor ancestor enumeration covers AGENTS.
	for _, path := range []string{nativeFile, nativeCanonical, rules, rulesCanonical} {
		pattern := literalPattern(filepath.ToSlash(path))
		if path == rules || path == rulesCanonical {
			pattern += "/**"
		}
		if !contains(pattern) {
			excludes = append(excludes, pattern)
		}
	}
	updated := Ownership{Version: 1, OwnedPatterns: []string{}}
	for _, path := range ancestors {
		if !filepath.IsAbs(path) || (filepath.Base(path) != "CLAUDE.md" && filepath.Base(path) != "CLAUDE.local.md") {
			return nil, nil, fmt.Errorf("invalid Claude ancestor filename")
		}
		pattern := literalPattern(filepath.ToSlash(filepath.Clean(path)))
		if !contains(pattern) {
			excludes = append(excludes, pattern)
			updated.OwnedPatterns = append(updated.OwnedPatterns, pattern)
		}
	}
	settings := payload
	if !reflect.DeepEqual(excludes, original) {
		encoded, err := json.Marshal(excludes)
		if err != nil {
			return nil, nil, err
		}
		object["claudeMdExcludes"] = encoded
		settings, err = json.Marshal(object)
		if err != nil {
			return nil, nil, err
		}
		settings = append(settings, '\n')
	}
	next, err := json.Marshal(updated)
	if err != nil {
		return nil, nil, err
	}
	return settings, append(next, '\n'), nil
}

func readOwnership(payload []byte) (Ownership, error) {
	var result Ownership
	if payload == nil {
		return Ownership{Version: 1, OwnedPatterns: []string{}}, nil
	}
	if err := protocoljson.Validate(payload); err != nil {
		return result, fmt.Errorf("invalid Claude memory ownership JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil || result.Version != 1 || result.OwnedPatterns == nil {
		return Ownership{}, fmt.Errorf("invalid Claude memory ownership schema")
	}
	seen := map[string]bool{}
	for _, pattern := range result.OwnedPatterns {
		path, valid := literalMemoryPath(pattern)
		if !valid || seen[pattern] || (filepath.Base(path) != "CLAUDE.md" && filepath.Base(path) != "CLAUDE.local.md") {
			return Ownership{}, fmt.Errorf("invalid owned Claude ancestor pattern")
		}
		seen[pattern] = true
	}
	return result, nil
}

// Ownership accepts only the literal encoding produced here, never broad
// user globs that could remove an unrelated user exclusion during refresh.
func literalMemoryPath(pattern string) (string, bool) {
	var decoded strings.Builder
	for len(pattern) > 0 {
		r, size := utf8.DecodeRuneInString(pattern)
		if r == '[' {
			if len(pattern) <= size {
				return "", false
			}
			r, n := utf8.DecodeRuneInString(pattern[size:])
			if !strings.ContainsRune("*?[]{}()!+@", r) || len(pattern) <= size+n || pattern[size+n] != ']' {
				return "", false
			}
			decoded.WriteRune(r)
			pattern = pattern[size+n+1:]
		} else {
			if strings.ContainsRune("*?]{}()!+@", r) || r == 0 {
				return "", false
			}
			decoded.WriteRune(r)
			pattern = pattern[size:]
		}
	}
	path := filepath.FromSlash(decoded.String())
	return path, filepath.IsAbs(path) && filepath.Clean(path) == path
}

func exclusionSettings(payload []byte) (map[string]json.RawMessage, []string, error) {
	object := map[string]json.RawMessage{}
	if payload != nil {
		if err := protocoljson.Validate(payload); err != nil {
			return nil, nil, fmt.Errorf("invalid Claude settings JSON")
		}
		if err := json.Unmarshal(payload, &object); err != nil || object == nil {
			return nil, nil, fmt.Errorf("settings must be an object")
		}
	}
	excludes := []string{}
	if value, ok := object["claudeMdExcludes"]; ok {
		var entries []any
		if err := json.Unmarshal(value, &entries); err != nil || entries == nil {
			return nil, nil, fmt.Errorf("claudeMdExcludes must be an array of strings")
		}
		for _, entry := range entries {
			path, ok := entry.(string)
			if !ok {
				return nil, nil, fmt.Errorf("claudeMdExcludes must be an array of strings")
			}
			excludes = append(excludes, path)
		}
	}
	return object, excludes, nil
}
