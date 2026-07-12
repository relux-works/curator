// Package capabilities parses the capability declaration of Spec §5.5.
//
// Capabilities describe the boundaries within which a skill is expected to
// operate. They are an audit and review surface, not a runtime sandbox.
package capabilities

import (
	"strings"

	"github.com/relux-works/curator/internal/verr"
)

// Manifest is a parsed capability declaration. Zero-value list fields mean
// "none". Filesystem is either one of the keywords "repo" or "home-config",
// or nil Keyword with explicit Paths.
type Manifest struct {
	Network     []string
	Filesystem  Filesystem
	Exec        []string
	Secrets     []string
	EnvRead     []string
	PromptScope string
}

// Filesystem is either a keyword ("repo", "home-config") or a path list.
type Filesystem struct {
	Keyword string
	Paths   []string
}

// ImplicitNone is the capability set of manifests older than schema 3:
// nothing declared.
func ImplicitNone() Manifest {
	return Manifest{}
}

// Parse validates a raw capabilities object (Spec §5.5). A nil raw value
// yields ImplicitNone.
func Parse(raw any) (Manifest, error) {
	if raw == nil {
		return ImplicitNone(), nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return Manifest{}, verr.New("capabilities", "must be an object")
	}
	if err := rejectUnknown(obj, map[string]bool{
		"network": true, "filesystem": true, "exec": true,
		"secrets": true, "env_read": true, "prompt_scope": true,
	}, "capabilities"); err != nil {
		return Manifest{}, err
	}

	network, err := noneOrList(valueOr(obj, "network", "none"), "capabilities.network", validateHostGlob)
	if err != nil {
		return Manifest{}, err
	}
	filesystem, err := parseFilesystem(valueOr(obj, "filesystem", "repo"))
	if err != nil {
		return Manifest{}, err
	}
	execs, err := noneOrList(valueOr(obj, "exec", "none"), "capabilities.exec", validateExecutable)
	if err != nil {
		return Manifest{}, err
	}
	secrets, err := noneOrList(valueOr(obj, "secrets", "none"), "capabilities.secrets", nil)
	if err != nil {
		return Manifest{}, err
	}
	envRead, err := noneOrList(valueOr(obj, "env_read", []any{}), "capabilities.env_read", validateEnvVar)
	if err != nil {
		return Manifest{}, err
	}

	promptScope := ""
	if rawScope, present := obj["prompt_scope"]; present && rawScope != nil {
		text, ok := rawScope.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return Manifest{}, verr.New("capabilities.prompt_scope", "must be a non-empty string when present")
		}
		promptScope = strings.TrimSpace(text)
	}

	return Manifest{
		Network:     network,
		Filesystem:  filesystem,
		Exec:        execs,
		Secrets:     secrets,
		EnvRead:     envRead,
		PromptScope: promptScope,
	}, nil
}

func valueOr(obj map[string]any, key string, fallback any) any {
	if value, present := obj[key]; present {
		return value
	}
	return fallback
}

func parseFilesystem(raw any) (Filesystem, error) {
	if keyword, ok := raw.(string); ok && (keyword == "repo" || keyword == "home-config") {
		return Filesystem{Keyword: keyword}, nil
	}
	paths, err := noneOrList(raw, "capabilities.filesystem", validatePath)
	if err != nil {
		return Filesystem{}, err
	}
	return Filesystem{Paths: paths}, nil
}

// noneOrList accepts the literal "none" (empty result) or a list of unique
// non-empty strings, each optionally validated.
func noneOrList(raw any, field string, validate func(string, string) error) ([]string, error) {
	if raw == "none" {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New(field, "must be 'none' or a list of strings")
	}
	values := make([]string, 0, len(list))
	seen := map[string]bool{}
	for index, item := range list {
		text, ok := item.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil, verr.New(field, "[%d] must be a non-empty string", index)
		}
		text = strings.TrimSpace(text)
		if validate != nil {
			if err := validate(text, field+indexSuffix(index)); err != nil {
				return nil, err
			}
		}
		if seen[text] {
			return nil, verr.New(field, "values must be unique")
		}
		seen[text] = true
		values = append(values, text)
	}
	return values, nil
}

func indexSuffix(index int) string {
	return "[" + itoa(index) + "]"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func validateHostGlob(value, field string) error {
	if strings.ContainsAny(value, " \t/\\") {
		return verr.New(field, "must be a host glob, not a URL or path")
	}
	return nil
}

func validateExecutable(value, field string) error {
	if strings.HasPrefix(value, "-") || strings.ContainsAny(value, "/\\ \t") {
		return verr.New(field, "must be an executable name, not a path or command line")
	}
	return nil
}

func validateEnvVar(value, field string) error {
	for i, r := range value {
		isAlpha := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
		isDigit := r >= '0' && r <= '9'
		if i == 0 && !isAlpha {
			return verr.New(field, "must be an environment variable name")
		}
		if !isAlpha && !isDigit {
			return verr.New(field, "must be an environment variable name")
		}
	}
	return nil
}

func validatePath(value, field string) error {
	if strings.ContainsRune(value, 0) {
		return verr.New(field, "must not contain NUL bytes")
	}
	if strings.HasPrefix(value, "-") {
		return verr.New(field, "must not start with '-'")
	}
	if strings.HasPrefix(value, "~/") || strings.HasPrefix(value, "/") || value == "~" {
		return nil
	}
	for _, part := range strings.Split(value, "/") {
		if part == ".." {
			return verr.New(field, "must not contain '..'")
		}
	}
	return nil
}

func rejectUnknown(obj map[string]any, allowed map[string]bool, label string) error {
	var unknown []string
	for key := range obj {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sortStrings(unknown)
	return verr.New(label, "has unsupported field(s): %s", strings.Join(unknown, ", "))
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j-1] > values[j]; j-- {
			values[j-1], values[j] = values[j], values[j-1]
		}
	}
}
