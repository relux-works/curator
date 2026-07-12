// Package config loads the machine configuration (Spec §7.1) and applies the
// enforced system configuration with locked keys (Spec §7.2).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/verr"
)

// SchemaVersion is the only supported machine config schema.
const SchemaVersion = 1

// DefaultWorktreeAliasPattern extracts checkout aliases for git worktrees.
const DefaultWorktreeAliasPattern = `[A-Z]+-[0-9]+`

// DefaultAgents applies when a project declares no agents anywhere.
var DefaultAgents = []string{"codex_cli"}

// LockableKeys are the top-level keys an organization may lock from the
// system config (Spec §7.2).
var LockableKeys = map[string]bool{
	"audit_registries":           true,
	"disable_builtin_registries": true,
	"allowed_sources":            true,
	"audit":                      true,
}

var revocationHashRE = regexp.MustCompile(`^(?:sha256:)?[A-Fa-f0-9]{64}$`)

// Project is a registered project entry.
type Project struct {
	Alias         string
	Path          string
	Agents        []string
	ProjectAlias  string
	CheckoutAlias string
}

// Registry is a trusted audit registry entry (Spec §7.1).
type Registry struct {
	Name       string
	URL        string
	PublicKeys []string
	Enabled    bool
}

// SourcePolicyRule classifies sources for cloud audit egress (Spec §12.1).
type SourcePolicyRule struct {
	Pattern string
	Class   string
}

// Audit is the audit and registry policy configuration. Backend-specific
// validation lives with the audit gate; this layer keeps the raw backends
// object.
type Audit struct {
	Enabled           bool
	Mode              string // advisory | strict
	FailOn            string // off | low | medium | high | critical
	Backend           string
	Model             string
	AllowCloud        bool
	Backends          map[string]any
	Grants            []map[string]any
	Revocations       []string
	SourcePolicyClass string // default_class
	SourcePolicyRules []SourcePolicyRule
	RegistryPolicy    string // advisory | strict
}

// Config is the effective machine configuration.
type Config struct {
	Path                     string
	SkillsRoot               string
	PreferredLocale          string
	DefaultAgents            []string
	AdapterMode              string // auto | symlink | copy
	WorktreeAliasPattern     string
	Projects                 map[string]Project
	Audit                    Audit
	AllowedSources           []string
	AuditRegistries          []Registry
	DisableBuiltinRegistries bool
}

// Home returns the directory holding the config file: the machine home for
// caches, runtime, scopes.
func (c *Config) Home() string { return filepath.Dir(c.Path) }

// TrustedRegistries returns the effective registries: built-in defaults plus
// configured entries; a configured entry with the same URL overrides a
// built-in one; disabled entries drop out (Spec §7.1, §13).
func (c *Config) TrustedRegistries() []Registry {
	byURL := map[string]Registry{}
	var order []string
	if !c.DisableBuiltinRegistries {
		for _, entry := range BuiltinRegistries {
			if _, seen := byURL[entry.URL]; !seen {
				order = append(order, entry.URL)
			}
			byURL[entry.URL] = entry
		}
	}
	for _, entry := range c.AuditRegistries {
		if _, seen := byURL[entry.URL]; !seen {
			order = append(order, entry.URL)
		}
		byURL[entry.URL] = entry
	}
	var result []Registry
	for _, url := range order {
		if entry := byURL[url]; entry.Enabled {
			result = append(result, entry)
		}
	}
	return result
}

// BuiltinRegistries ships with a release; empty until a central registry key
// is pinned.
var BuiltinRegistries = []Registry{}

// UserPath resolves the user config path, honoring the env override.
func UserPath() string {
	if override := os.Getenv("CURATOR_CONFIG"); override != "" {
		return expandHome(override)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".curator", "config.json")
}

// SystemPath resolves the enforced system config path, honoring the env
// override; returns "" when no system config exists.
func SystemPath() string {
	if override := os.Getenv("CURATOR_SYSTEM_CONFIG"); override != "" {
		return expandHome(override)
	}
	var candidate string
	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = `C:\ProgramData`
		}
		candidate = filepath.Join(programData, "curator", "config.json")
	} else {
		candidate = "/etc/curator/config.json"
	}
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

// Load reads the user config, overlays the system config, and parses the
// result. warn receives locked-key override warnings.
func Load(path string, warn func(string)) (*Config, error) {
	if path == "" {
		path = UserPath()
	}
	userData, err := readObject(path)
	if err != nil {
		return nil, err
	}
	if systemPath := SystemPath(); systemPath != "" {
		systemData, err := readObject(systemPath)
		if err != nil {
			return nil, err
		}
		userData, err = applySystem(systemData, userData, systemPath, warn)
		if err != nil {
			return nil, err
		}
	}
	return Parse(userData, path)
}

// applySystem overlays the system config (Spec §7.2): locked keys win over
// the user value with a warning; unlocked keys act as defaults; a locked but
// unset key is a configuration error.
func applySystem(systemData, userData map[string]any, systemPath string, warn func(string)) (map[string]any, error) {
	rawLocked, _ := systemData["locked"].([]any)
	if systemData["locked"] != nil && rawLocked == nil {
		return nil, verr.New("locked", "system config %s field 'locked' must be a list of strings", systemPath)
	}
	locked := map[string]bool{}
	for _, item := range rawLocked {
		key, ok := item.(string)
		if !ok {
			return nil, verr.New("locked", "system config %s field 'locked' must be a list of strings", systemPath)
		}
		locked[key] = true
	}
	merged := map[string]any{}
	for key, value := range userData {
		merged[key] = value
	}
	for key, value := range systemData {
		if key == "locked" || key == "schema_version" {
			continue
		}
		if locked[key] {
			if userValue, present := userData[key]; present && !jsonEqual(userValue, value) && warn != nil {
				warn(fmt.Sprintf("config key %q is locked by %s; the user override is ignored", key, systemPath))
			}
			merged[key] = value
		} else if _, present := merged[key]; !present {
			merged[key] = value
		}
	}
	for key := range locked {
		if _, present := systemData[key]; !present {
			return nil, verr.New("locked", "system config %s locks %q but does not set it", systemPath, key)
		}
	}
	return merged, nil
}

// Parse validates a raw config object (Spec §7.1).
func Parse(data map[string]any, path string) (*Config, error) {
	schema, ok := data["schema_version"].(float64)
	if !ok || schema != float64(int(schema)) {
		return nil, verr.New("schema_version", "must be an integer")
	}
	if int(schema) != SchemaVersion {
		return nil, verr.New("schema_version", "unsupported config schema_version %d; this config requires a newer tool", int(schema))
	}

	skillsRoot, ok := data["skills_root"].(string)
	if !ok || skillsRoot == "" {
		return nil, verr.New("skills_root", "requires a non-empty string")
	}

	defaultAgents := append([]string(nil), DefaultAgents...)
	if raw, present := data["default_agents"]; present {
		list, err := stringList(raw, "default_agents")
		if err != nil {
			return nil, err
		}
		defaultAgents = list
	}

	preferredLocale := ""
	if raw, present := data["preferred_locale"]; present && raw != nil {
		preferredLocale, ok = raw.(string)
		if !ok {
			return nil, verr.New("preferred_locale", "must be a string when present")
		}
	}

	adapterMode := "auto"
	if raw, present := data["adapter_mode"]; present {
		adapterMode, _ = raw.(string)
		if adapterMode != "auto" && adapterMode != "symlink" && adapterMode != "copy" {
			return nil, verr.New("adapter_mode", "must be auto, symlink, or copy")
		}
	}

	worktreePattern := DefaultWorktreeAliasPattern
	if raw, present := data["worktree_alias_pattern"]; present {
		worktreePattern, ok = raw.(string)
		if !ok || worktreePattern == "" {
			return nil, verr.New("worktree_alias_pattern", "must be a non-empty string")
		}
		if _, err := regexp.Compile(worktreePattern); err != nil {
			return nil, verr.New("worktree_alias_pattern", "is not a valid regex: %v", err)
		}
	}

	audit, err := parseAudit(data["audit"])
	if err != nil {
		return nil, err
	}

	var allowedSources []string
	if raw, present := data["allowed_sources"]; present {
		allowedSources, err = stringList(raw, "allowed_sources")
		if err != nil {
			return nil, err
		}
		for _, item := range allowedSources {
			if strings.TrimSpace(item) == "" {
				return nil, verr.New("allowed_sources", "must be a list of non-empty strings")
			}
		}
	}

	registries, err := parseRegistries(data["audit_registries"])
	if err != nil {
		return nil, err
	}

	disableBuiltin := false
	if raw, present := data["disable_builtin_registries"]; present {
		disableBuiltin, ok = raw.(bool)
		if !ok {
			return nil, verr.New("disable_builtin_registries", "must be a boolean")
		}
	}

	rawProjects, present := data["projects"]
	if !present {
		return nil, verr.New("projects", "required field")
	}
	projectsObj, ok := rawProjects.(map[string]any)
	if !ok {
		return nil, verr.New("projects", "must be an object")
	}
	projects := map[string]Project{}
	for alias, rawEntry := range projectsObj {
		label := "projects." + alias
		if alias == "" {
			return nil, verr.New("projects", "project alias must be a non-empty string")
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		projectPath, ok := entry["path"].(string)
		if !ok || projectPath == "" {
			return nil, verr.New(label+".path", "requires a non-empty string")
		}
		var agents []string
		if raw, present := entry["agents"]; present {
			agents, err = stringList(raw, label+".agents")
			if err != nil {
				return nil, err
			}
		}
		projectAlias := alias
		if raw, present := entry["project_alias"]; present && raw != nil {
			projectAlias, ok = raw.(string)
			if !ok || projectAlias == "" {
				return nil, verr.New(label+".project_alias", "must be a non-empty string when present")
			}
		}
		checkoutAlias := alias
		if raw, present := entry["checkout_alias"]; present && raw != nil {
			checkoutAlias, ok = raw.(string)
			if !ok || checkoutAlias == "" {
				return nil, verr.New(label+".checkout_alias", "must be a non-empty string when present")
			}
		}
		projects[alias] = Project{
			Alias:         alias,
			Path:          expandHome(projectPath),
			Agents:        agents,
			ProjectAlias:  projectAlias,
			CheckoutAlias: checkoutAlias,
		}
	}

	return &Config{
		Path:                     path,
		SkillsRoot:               expandHome(skillsRoot),
		PreferredLocale:          preferredLocale,
		DefaultAgents:            defaultAgents,
		AdapterMode:              adapterMode,
		WorktreeAliasPattern:     worktreePattern,
		Projects:                 projects,
		Audit:                    audit,
		AllowedSources:           allowedSources,
		AuditRegistries:          registries,
		DisableBuiltinRegistries: disableBuiltin,
	}, nil
}

func parseAudit(raw any) (Audit, error) {
	audit := Audit{Mode: "advisory", FailOn: "high", Backend: "null", RegistryPolicy: "advisory", SourcePolicyClass: "internal"}
	if raw == nil {
		return audit, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return Audit{}, verr.New("audit", "must be an object")
	}
	allowed := map[string]bool{
		"enabled": true, "mode": true, "fail_on": true, "backend": true, "model": true,
		"allow_cloud": true, "max_request_bytes": true, "backends": true, "grants": true,
		"revocations": true, "source_policy": true, "registry_policy": true,
	}
	var unknown []string
	for key := range obj {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return Audit{}, verr.New("audit", "has unsupported field(s): %s", strings.Join(unknown, ", "))
	}

	if raw, present := obj["enabled"]; present {
		audit.Enabled, ok = raw.(bool)
		if !ok {
			return Audit{}, verr.New("audit.enabled", "must be a boolean")
		}
	}
	if raw, present := obj["mode"]; present {
		audit.Mode, _ = raw.(string)
		if audit.Mode != "advisory" && audit.Mode != "strict" {
			return Audit{}, verr.New("audit.mode", "must be advisory or strict")
		}
	}
	if raw, present := obj["fail_on"]; present {
		audit.FailOn, _ = raw.(string)
		switch audit.FailOn {
		case "off", "low", "medium", "high", "critical":
		default:
			return Audit{}, verr.New("audit.fail_on", "must be off, low, medium, high, or critical")
		}
	}
	if raw, present := obj["backend"]; present {
		audit.Backend, _ = raw.(string)
		if audit.Backend == "" {
			return Audit{}, verr.New("audit.backend", "must be a non-empty string")
		}
	}
	if raw, present := obj["model"]; present && raw != nil {
		audit.Model, ok = raw.(string)
		if !ok || audit.Model == "" {
			return Audit{}, verr.New("audit.model", "must be a non-empty string when present")
		}
	}
	if raw, present := obj["allow_cloud"]; present {
		audit.AllowCloud, ok = raw.(bool)
		if !ok {
			return Audit{}, verr.New("audit.allow_cloud", "must be a boolean")
		}
	}
	if raw, present := obj["backends"]; present {
		audit.Backends, ok = raw.(map[string]any)
		if !ok {
			return Audit{}, verr.New("audit.backends", "must be an object")
		}
	}
	if raw, present := obj["grants"]; present {
		list, ok := raw.([]any)
		if !ok {
			return Audit{}, verr.New("audit.grants", "must be a list of objects")
		}
		for _, item := range list {
			grant, ok := item.(map[string]any)
			if !ok {
				return Audit{}, verr.New("audit.grants", "must be a list of objects")
			}
			audit.Grants = append(audit.Grants, grant)
		}
	}
	if raw, present := obj["revocations"]; present {
		list, err := stringList(raw, "audit.revocations")
		if err != nil {
			return Audit{}, err
		}
		for index, item := range list {
			if revocationHashRE.MatchString(item) {
				continue
			}
			if strings.HasPrefix(item, "source:") && strings.TrimSpace(strings.TrimPrefix(item, "source:")) != "" {
				continue
			}
			return Audit{}, verr.New(fmt.Sprintf("audit.revocations[%d]", index), "must be a SHA256 hash or source:<pattern>")
		}
		audit.Revocations = list
	}
	if raw, present := obj["source_policy"]; present && raw != nil {
		policy, ok := raw.(map[string]any)
		if !ok {
			return Audit{}, verr.New("audit.source_policy", "must be an object")
		}
		if rawClass, present := policy["default_class"]; present {
			audit.SourcePolicyClass, _ = rawClass.(string)
			if audit.SourcePolicyClass != "internal" && audit.SourcePolicyClass != "public" {
				return Audit{}, verr.New("audit.source_policy.default_class", "must be internal or public")
			}
		}
		if rawRules, present := policy["rules"]; present {
			rules, ok := rawRules.([]any)
			if !ok {
				return Audit{}, verr.New("audit.source_policy.rules", "must be a list")
			}
			for index, item := range rules {
				rule, ok := item.(map[string]any)
				if !ok {
					return Audit{}, verr.New(fmt.Sprintf("audit.source_policy.rules[%d]", index), "must be an object")
				}
				pattern, _ := rule["pattern"].(string)
				class, _ := rule["class"].(string)
				if pattern == "" || (class != "internal" && class != "public") {
					return Audit{}, verr.New(fmt.Sprintf("audit.source_policy.rules[%d]", index), "requires 'pattern' and 'class' of internal or public")
				}
				audit.SourcePolicyRules = append(audit.SourcePolicyRules, SourcePolicyRule{Pattern: pattern, Class: class})
			}
		}
	}
	if raw, present := obj["registry_policy"]; present {
		audit.RegistryPolicy, _ = raw.(string)
		if audit.RegistryPolicy != "advisory" && audit.RegistryPolicy != "strict" {
			return Audit{}, verr.New("audit.registry_policy", "must be advisory or strict")
		}
	}
	return audit, nil
}

func parseRegistries(raw any) ([]Registry, error) {
	if raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("audit_registries", "must be a list")
	}
	var registries []Registry
	seen := map[string]bool{}
	for index, item := range list {
		label := fmt.Sprintf("audit_registries[%d]", index)
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		name, _ := entry["name"].(string)
		if name == "" {
			return nil, verr.New(label, "requires a non-empty string 'name'")
		}
		url, _ := entry["url"].(string)
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return nil, verr.New(label, "requires an http(s) 'url'")
		}
		if seen[url] {
			return nil, verr.New(label, "duplicates url %q", url)
		}
		seen[url] = true
		var keys []string
		if rawKeys, present := entry["public_keys"]; present {
			var err error
			keys, err = stringList(rawKeys, label+".public_keys")
			if err != nil {
				return nil, err
			}
			for _, key := range keys {
				if strings.TrimSpace(key) == "" {
					return nil, verr.New(label+".public_keys", "must be a list of non-empty strings")
				}
			}
		}
		enabled := true
		if rawEnabled, present := entry["enabled"]; present {
			enabled, ok = rawEnabled.(bool)
			if !ok {
				return nil, verr.New(label+".enabled", "must be a boolean")
			}
		}
		registries = append(registries, Registry{Name: name, URL: url, PublicKeys: keys, Enabled: enabled})
	}
	return registries, nil
}

func readObject(path string) (map[string]any, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- config path comes from the operator
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("global config not found: %s", path)
		}
		return nil, err
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON in config %s: %w", path, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config %s must be a JSON object", path)
	}
	return obj, nil
}

func stringList(raw any, field string) ([]string, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New(field, "must be a list of strings")
	}
	values := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, verr.New(field, "must be a list of strings")
		}
		values = append(values, text)
	}
	return values, nil
}

func jsonEqual(a, b any) bool {
	left, errA := json.Marshal(a)
	right, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(left) == string(right)
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~"), "/"))
		}
	}
	return path
}
