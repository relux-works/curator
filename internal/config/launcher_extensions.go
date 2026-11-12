package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SchemaVersion7 extends the implemented manager-config-v5 carrier with closed
// launcher selections. Version 6 remains the separate node-source carrier.
const SchemaVersion7 = 7

// LauncherExtensions contains only metadata references. Reading, editing and
// exporting these values never reads a secret store or starts a runtime.
type LauncherExtensions struct {
	GrantService *GrantServiceReference `json:"grant_service,omitempty"`
	Proxy        *LauncherProxy         `json:"proxy,omitempty"`
	ClaudeAuth   *LauncherClaudeAuth    `json:"claude_auth,omitempty"`
}

// LauncherProxy selects a named network profile from the admitted catalog.
type LauncherProxy struct {
	Profile string `json:"profile"`
}

// LauncherClaudeAuth selects OAuth through a file or Keychain reference.
type LauncherClaudeAuth struct {
	Type  string            `json:"type"`
	Store LauncherAuthStore `json:"store"`
}

// LauncherAuthStore is a closed file-or-Keychain reference, never material.
type LauncherAuthStore struct {
	Type    string `json:"type"`
	Path    string `json:"path,omitempty"`
	Service string `json:"service,omitempty"`
	Account string `json:"account,omitempty"`
}

// LauncherConfigProjection is the narrow transport for the effective selected
// machine configuration. ConfigPath is the actual Core-selected source.
type LauncherConfigProjection struct {
	Schema       string             `json:"schema"`
	ConfigSchema int                `json:"config_schema"`
	ConfigPath   string             `json:"config_path"`
	Extensions   LauncherExtensions `json:"extensions"`
}

// LauncherProjection returns the effective launcher metadata without reading credentials.
func (c *Config) LauncherProjection() LauncherConfigProjection {
	p := LauncherConfigProjection{Schema: "curator.launcher-config/1", ConfigSchema: c.Schema, ConfigPath: c.Path}
	if c.Extensions != nil {
		if c.Extensions.Proxy != nil {
			value := *c.Extensions.Proxy
			p.Extensions.Proxy = &value
		}
		if c.Extensions.ClaudeAuth != nil {
			value := *c.Extensions.ClaudeAuth
			p.Extensions.ClaudeAuth = &value
		}
	}
	return p
}

func launcherReference(raw any) (string, bool) {
	s, ok := raw.(string)
	return s, ok && s != "" && len(s) <= 4096 && utf8.ValidString(s) && strings.TrimSpace(s) == s && !strings.ContainsFunc(s, unicode.IsControl)
}

func launcherMembers(raw any, names ...string) (map[string]any, bool) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != len(names) {
		return nil, false
	}
	for _, key := range names {
		if _, present := object[key]; !present {
			return nil, false
		}
	}
	return object, true
}

func parseLauncherExtensions(raw any) (*LauncherExtensions, error) {
	fail := func() (*LauncherExtensions, error) { return nil, fmt.Errorf("extensions: invalid launcher selection") }
	object, ok := raw.(map[string]any)
	if !ok {
		return fail()
	}
	for key := range object {
		if key != "proxy" && key != "claude_auth" && key != "grant_service" {
			return fail()
		}
	}
	result := &LauncherExtensions{}
	if rawGrant, present := object["grant_service"]; present {
		value, err := parseGrantService(rawGrant)
		if err != nil {
			return nil, err
		}
		result.GrantService = value
	}
	if rawProxy, present := object["proxy"]; present {
		proxy, ok := launcherMembers(rawProxy, "profile")
		if !ok {
			return fail()
		}
		profile, ok := launcherReference(proxy["profile"])
		if !ok {
			return fail()
		}
		result.Proxy = &LauncherProxy{Profile: profile}
	}
	if rawAuth, present := object["claude_auth"]; present {
		auth, ok := launcherMembers(rawAuth, "type", "store")
		if !ok || auth["type"] != "oauth" {
			return fail()
		}
		store, ok := auth["store"].(map[string]any)
		if !ok {
			return fail()
		}
		selected := LauncherAuthStore{}
		switch store["type"] {
		case "file":
			if _, ok := launcherMembers(store, "type", "path"); !ok {
				return fail()
			}
			path, ok := launcherReference(store["path"])
			if !ok || !filepath.IsAbs(path) || filepath.Clean(path) != path {
				return fail()
			}
			selected.Type, selected.Path = "file", path
		case "keychain":
			if _, ok := launcherMembers(store, "type", "service", "account"); !ok {
				return fail()
			}
			service, validService := launcherReference(store["service"])
			account, validAccount := launcherReference(store["account"])
			if !validService || !validAccount {
				return fail()
			}
			selected.Type, selected.Service, selected.Account = "keychain", service, account
		default:
			return fail()
		}
		result.ClaudeAuth = &LauncherClaudeAuth{Type: "oauth", Store: selected}
	}
	return result, nil
}
