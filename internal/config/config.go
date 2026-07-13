// Package config loads the machine configuration (Spec §7.1) and applies the
// enforced system configuration with locked keys (Spec §7.2).
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SchemaVersion is the only supported machine config schema.
const SchemaVersion = 1

// DefaultWorktreeAliasPattern extracts checkout aliases for git worktrees.
const DefaultWorktreeAliasPattern = `[A-Z]+-[0-9]+`

// Protocol defaults for audit requests, registry snapshots, and record cache.
const (
	DefaultMaxRequestBytes          = 1_048_576
	MaximumMaxRequestBytes          = 10_485_760
	DefaultSnapshotMaxAgeSeconds    = 604_800
	DefaultSnapshotClockSkewSeconds = 300
	DefaultCacheTTLSeconds          = 3_600
	DefaultOfflineGraceSeconds      = 604_800
	maximumDurationSeconds          = 2_147_483_647
)

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

var (
	revocationHashRE = regexp.MustCompile(`^(?:sha256:)?[A-Fa-f0-9]{64}$`)
	registryHostRE   = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$`)
	pinnedKeyRE      = regexp.MustCompile(`^(?:ed25519:)?[A-Za-z0-9+/]{43}=$`)
)

var managerKeys = map[string]bool{
	"schema_version": true, "skills_root": true, "default_agents": true,
	"preferred_locale": true, "adapter_mode": true, "worktree_alias_pattern": true,
	"projects": true, "allowed_sources": true, "audit": true,
	"audit_registries": true, "disable_builtin_registries": true,
}

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
	Enabled                  bool
	Mode                     string // advisory | strict
	FailOn                   string // off | low | medium | high | critical
	Backend                  string
	Model                    string
	AllowCloud               bool
	Backends                 map[string]any
	Grants                   []map[string]any
	Revocations              []string
	SourcePolicyClass        string // default_class
	SourcePolicyRules        []SourcePolicyRule
	RegistryPolicy           string // advisory | strict
	MaxRequestBytes          int
	SnapshotMaxAgeSeconds    int
	SnapshotClockSkewSeconds int
	CacheTTLSeconds          int
	OfflineGraceSeconds      int
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
	if schema, ok := integerValue(systemData["schema_version"]); !ok || schema != SchemaVersion {
		return nil, verr.New("schema_version", "system config %s requires schema_version 1", systemPath)
	}
	for key := range systemData {
		if key != "locked" && !managerKeys[key] {
			return nil, verr.New(key, "system config %s has unsupported field", systemPath)
		}
	}
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
		if !LockableKeys[key] {
			return nil, verr.New("locked", "system config %s cannot lock %q", systemPath, key)
		}
		if locked[key] {
			return nil, verr.New("locked", "system config %s lists %q more than once", systemPath, key)
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
	for key := range data {
		if !managerKeys[key] {
			return nil, verr.New(key, "unsupported top-level configuration field")
		}
	}
	schema, ok := integerValue(data["schema_version"])
	if !ok {
		return nil, verr.New("schema_version", "must be an integer")
	}
	if schema != SchemaVersion {
		return nil, verr.New("schema_version", "unsupported config schema_version %d; this config requires a newer tool", schema)
	}

	skillsRoot, ok := data["skills_root"].(string)
	if !ok || skillsRoot == "" || utf8.RuneCountInString(skillsRoot) > 4096 {
		return nil, verr.New("skills_root", "requires a non-empty string")
	}

	defaultAgents := append([]string(nil), DefaultAgents...)
	if raw, present := data["default_agents"]; present {
		list, err := identifierSet(raw, "default_agents")
		if err != nil {
			return nil, err
		}
		defaultAgents = list
	}

	preferredLocale := ""
	if raw, present := data["preferred_locale"]; present && raw != nil {
		preferredLocale, ok = raw.(string)
		if !ok || !identifiers.ValidLocale(preferredLocale) {
			return nil, verr.New("preferred_locale", "must be null or a 1-64 character ASCII locale selector")
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
		if !ok || worktreePattern == "" || utf8.RuneCountInString(worktreePattern) > 1024 {
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
		seen := map[string]bool{}
		for _, item := range allowedSources {
			if strings.TrimSpace(item) == "" || utf8.RuneCountInString(item) > 4096 || seen[item] {
				return nil, verr.New("allowed_sources", "must be a list of non-empty strings")
			}
			seen[item] = true
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
		if !identifiers.Valid(alias) {
			return nil, verr.New("projects", "project alias %q %s", alias, identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		for key := range entry {
			if key != "path" && key != "agents" && key != "project_alias" && key != "checkout_alias" {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		projectPath, ok := entry["path"].(string)
		if !ok || projectPath == "" || utf8.RuneCountInString(projectPath) > 4096 {
			return nil, verr.New(label+".path", "requires a non-empty string")
		}
		var agents []string
		if raw, present := entry["agents"]; present {
			agents, err = identifierSet(raw, label+".agents")
			if err != nil {
				return nil, err
			}
		}
		projectAlias := alias
		if raw, present := entry["project_alias"]; present && raw != nil {
			projectAlias, ok = raw.(string)
			if !ok || !identifiers.Valid(projectAlias) {
				return nil, verr.New(label+".project_alias", "%s when present", identifiers.Rule)
			}
		}
		checkoutAlias := alias
		if raw, present := entry["checkout_alias"]; present && raw != nil {
			checkoutAlias, ok = raw.(string)
			if !ok || !identifiers.Valid(checkoutAlias) {
				return nil, verr.New(label+".checkout_alias", "%s when present", identifiers.Rule)
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
	audit := Audit{
		Mode: "advisory", FailOn: "high", Backend: "null", RegistryPolicy: "advisory",
		SourcePolicyClass: "internal", MaxRequestBytes: DefaultMaxRequestBytes,
		SnapshotMaxAgeSeconds:    DefaultSnapshotMaxAgeSeconds,
		SnapshotClockSkewSeconds: DefaultSnapshotClockSkewSeconds,
		CacheTTLSeconds:          DefaultCacheTTLSeconds, OfflineGraceSeconds: DefaultOfflineGraceSeconds,
	}
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
		"snapshot_max_age_seconds": true, "snapshot_clock_skew_seconds": true,
		"cache_ttl_seconds": true, "offline_grace_seconds": true,
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
		if audit.Backend == "" || utf8.RuneCountInString(audit.Backend) > 128 {
			return Audit{}, verr.New("audit.backend", "must be a non-empty string")
		}
	}
	if raw, present := obj["model"]; present && raw != nil {
		audit.Model, ok = raw.(string)
		if !ok || audit.Model == "" || utf8.RuneCountInString(audit.Model) > 1024 {
			return Audit{}, verr.New("audit.model", "must be a non-empty string when present")
		}
	}
	if raw, present := obj["allow_cloud"]; present {
		audit.AllowCloud, ok = raw.(bool)
		if !ok {
			return Audit{}, verr.New("audit.allow_cloud", "must be a boolean")
		}
	}
	var settingErr error
	if raw, present := obj["max_request_bytes"]; present {
		audit.MaxRequestBytes, settingErr = boundedInteger(raw, 1, MaximumMaxRequestBytes)
		if settingErr != nil {
			return Audit{}, verr.New("audit.max_request_bytes", "must be an integer between 1 and %d", MaximumMaxRequestBytes)
		}
	}
	for key, target := range map[string]*int{
		"snapshot_max_age_seconds":    &audit.SnapshotMaxAgeSeconds,
		"snapshot_clock_skew_seconds": &audit.SnapshotClockSkewSeconds,
		"cache_ttl_seconds":           &audit.CacheTTLSeconds,
		"offline_grace_seconds":       &audit.OfflineGraceSeconds,
	} {
		raw, present := obj[key]
		if !present {
			continue
		}
		minimum := 0
		if key == "snapshot_max_age_seconds" {
			minimum = 1
		}
		*target, settingErr = boundedInteger(raw, minimum, maximumDurationSeconds)
		if settingErr != nil {
			return Audit{}, verr.New("audit."+key, "must be an integer between %d and %d", minimum, maximumDurationSeconds)
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
		for key := range policy {
			if key != "default_class" && key != "rules" {
				return Audit{}, verr.New("audit.source_policy", "has unsupported field %q", key)
			}
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
				for key := range rule {
					if key != "pattern" && key != "class" {
						return Audit{}, verr.New(fmt.Sprintf("audit.source_policy.rules[%d]", index), "has unsupported field %q", key)
					}
				}
				pattern, _ := rule["pattern"].(string)
				class, _ := rule["class"].(string)
				if pattern == "" || utf8.RuneCountInString(pattern) > 4096 || (class != "internal" && class != "public") {
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
		for key := range entry {
			if key != "name" && key != "url" && key != "public_keys" && key != "enabled" {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		name, _ := entry["name"].(string)
		if !identifiers.Valid(name) {
			return nil, verr.New(label+".name", "%s", identifiers.Rule)
		}
		rawURL, _ := entry["url"].(string)
		registryURL, err := canonicalRegistryURL(rawURL)
		if err != nil {
			return nil, verr.New(label+".url", "%v", err)
		}
		if seen[registryURL] {
			return nil, verr.New(label, "duplicates url %q", registryURL)
		}
		seen[registryURL] = true
		var keys []string
		if rawKeys, present := entry["public_keys"]; present {
			var err error
			keys, err = stringList(rawKeys, label+".public_keys")
			if err != nil {
				return nil, err
			}
			seenKeys := map[string]bool{}
			for _, key := range keys {
				if !pinnedKeyRE.MatchString(key) || seenKeys[key] {
					return nil, verr.New(label+".public_keys", "must contain unique canonical Ed25519 public keys")
				}
				seenKeys[key] = true
			}
		}
		enabled := true
		if rawEnabled, present := entry["enabled"]; present {
			enabled, ok = rawEnabled.(bool)
			if !ok {
				return nil, verr.New(label+".enabled", "must be a boolean")
			}
		}
		registries = append(registries, Registry{Name: name, URL: registryURL, PublicKeys: keys, Enabled: enabled})
	}
	return registries, nil
}

func canonicalRegistryURL(value string) (string, error) {
	if value == "" || utf8.RuneCountInString(value) > 4096 {
		return "", fmt.Errorf("must contain at most 4096 characters")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("requires an http(s) URL with a host")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if parsed.Hostname() == "" || (scheme != "http" && scheme != "https") {
		return "", fmt.Errorf("requires an http(s) URL with a host")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || strings.Contains(value, "%") {
		return "", fmt.Errorf("must not contain credentials, a query, a fragment, or percent escapes")
	}
	host := strings.ToLower(parsed.Hostname())
	address := net.ParseIP(host)
	if address == nil && (len(host) > 253 || !registryHostRE.MatchString(host)) {
		return "", fmt.Errorf("requires a portable ASCII DNS host or IP literal")
	}
	if address != nil {
		host = address.String()
	}
	if scheme == "http" {
		if host != "localhost" && (address == nil || !address.IsLoopback()) {
			return "", fmt.Errorf("plain HTTP is permitted only for an explicitly configured loopback host")
		}
	}
	if strings.Contains(parsed.Path, "//") || strings.Contains(parsed.Path, `\`) {
		return "", fmt.Errorf("path must not contain doubled separators or backslashes")
	}
	path := strings.TrimRight(parsed.Path, "/")
	for _, character := range path {
		if character > unicode.MaxASCII || unicode.IsSpace(character) || unicode.IsControl(character) {
			return "", fmt.Errorf("path must contain only printable non-space ASCII characters")
		}
	}
	for _, component := range strings.Split(strings.Trim(path, "/"), "/") {
		if component == "." || component == ".." {
			return "", fmt.Errorf("path must not contain dot segments")
		}
	}
	port := parsed.Port()
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		port = ""
	}
	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + host + "]"
	}
	if port != "" {
		authority = net.JoinHostPort(host, port)
	}
	canonical := (&url.URL{Scheme: scheme, Host: authority, Path: path}).String()
	if utf8.RuneCountInString(canonical) > 4096 {
		return "", fmt.Errorf("canonical URL exceeds 4096 characters")
	}
	return canonical, nil
}

func readObject(path string) (map[string]any, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- config path comes from the operator
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("global config not found: %s", path)
		}
		return nil, err
	}
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("malformed JSON in config %s: %w", path, err)
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

func identifierSet(raw any, field string) ([]string, error) {
	values, err := stringList(raw, field)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, value := range values {
		if !identifiers.Valid(value) || seen[value] {
			return nil, verr.New(field, "must contain unique portable identifiers")
		}
		seen[value] = true
	}
	return values, nil
}

func integerValue(raw any) (int, bool) {
	switch value := raw.(type) {
	case int:
		return value, true
	case int64:
		if int64(int(value)) != value {
			return 0, false
		}
		return int(value), true
	case float64:
		if value != float64(int(value)) {
			return 0, false
		}
		return int(value), true
	case json.Number:
		parsed, err := value.Int64()
		if err != nil || int64(int(parsed)) != parsed {
			return 0, false
		}
		return int(parsed), true
	default:
		return 0, false
	}
}

func boundedInteger(raw any, minimum, maximum int) (int, error) {
	value, ok := integerValue(raw)
	if !ok || value < minimum || value > maximum {
		return 0, fmt.Errorf("outside range")
	}
	return value, nil
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
