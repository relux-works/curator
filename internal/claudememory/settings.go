// Package claudememory supplies the managed Claude parent-memory exclusion.
package claudememory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// NativeFile names the operator's default memory, independently of the selected
// CLAUDE_CONFIG_DIR. It does not inspect that file or its contents.
func NativeFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("operator home must be absolute")
	}
	return filepath.Join(home, ".claude", "CLAUDE.md"), nil
}

// Merge preserves existing settings and exclusions, adding only the native
// memory path. A satisfied document is returned byte-identical. Invalid input
// is never replaced with defaults. Number lexemes in unrelated values survive.
func Merge(payload []byte, nativeFile string) ([]byte, error) {
	if !filepath.IsAbs(nativeFile) {
		return nil, fmt.Errorf("native Claude memory path must be absolute")
	}
	object := map[string]json.RawMessage{}
	if payload != nil {
		if err := protocoljson.Validate(payload); err != nil {
			return nil, fmt.Errorf("invalid Claude settings JSON")
		}
		if err := json.Unmarshal(payload, &object); err != nil || object == nil {
			return nil, fmt.Errorf("Claude settings must be an object")
		}
	}
	var excludes []string
	if value, ok := object["claudeMdExcludes"]; ok {
		var entries []any
		if err := json.Unmarshal(value, &entries); err != nil || entries == nil {
			return nil, fmt.Errorf("claudeMdExcludes must be an array of strings")
		}
		for _, entry := range entries {
			path, ok := entry.(string)
			if !ok {
				return nil, fmt.Errorf("claudeMdExcludes must be an array of strings")
			}
			excludes = append(excludes, path)
		}
	}
	nativeFile = literalPattern(filepath.ToSlash(nativeFile))
	for _, path := range excludes {
		if path == nativeFile {
			return payload, nil
		}
	}
	excludes = append(excludes, nativeFile)
	encoded, err := json.Marshal(excludes)
	if err != nil {
		return nil, err
	}
	object["claudeMdExcludes"] = encoded
	encoded, err = json.Marshal(object)
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

// Claude interprets absolute paths as glob patterns too. Character classes
// keep metacharacters in an operator directory literal without a broad glob.
func literalPattern(path string) string {
	var out strings.Builder
	for _, r := range path {
		switch r {
		case '*', '?', '[', ']', '{', '}', '(', ')', '!', '+', '@':
			out.WriteRune('[')
			out.WriteRune(r)
			out.WriteRune(']')
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
