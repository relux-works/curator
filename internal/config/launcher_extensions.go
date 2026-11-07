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

// SchemaVersion8 adds the optional Host notification metadata carrier.
const SchemaVersion8 = 8

// LauncherExtensions contains only metadata references. Reading, editing and
// exporting these values never reads a secret store or starts a runtime.
type LauncherExtensions struct {
	Proxy        *LauncherProxy        `json:"proxy,omitempty"`
	ClaudeAuth   *LauncherClaudeAuth   `json:"claude_auth,omitempty"`
	Notification *LauncherNotification `json:"notification,omitempty"`
}

// LauncherNotification selects an immutable workflow through the admitted Host
// contract. It contains no endpoint, process specification or enrollment proof.
type LauncherNotification struct {
	Backend           string `json:"backend"`
	Contract          string `json:"contract,omitempty"`
	WorkflowReference string `json:"workflow_reference,omitempty"`
	WorkflowSHA256    string `json:"workflow_sha256,omitempty"`
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
	if c.Schema == SchemaVersion8 {
		p.Schema = "curator.launcher-config/2"
	}
	if c.Extensions != nil {
		if c.Extensions.Proxy != nil {
			value := *c.Extensions.Proxy
			p.Extensions.Proxy = &value
		}
		if c.Extensions.ClaudeAuth != nil {
			value := *c.Extensions.ClaudeAuth
			p.Extensions.ClaudeAuth = &value
		}
		if c.Extensions.Notification != nil {
			value := *c.Extensions.Notification
			p.Extensions.Notification = &value
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

func parseLauncherExtensions(raw any, schema int) (*LauncherExtensions, error) {
	fail := func() (*LauncherExtensions, error) { return nil, fmt.Errorf("extensions: invalid launcher selection") }
	object, ok := raw.(map[string]any)
	if !ok {
		return fail()
	}
	for key := range object {
		if key != "proxy" && key != "claude_auth" && (schema != SchemaVersion8 || key != "notification") {
			return fail()
		}
	}
	result := &LauncherExtensions{}
	if rawNotification, present := object["notification"]; present {
		selection, ok := rawNotification.(map[string]any)
		if !ok {
			return fail()
		}
		switch selection["backend"] {
		case "off":
			if _, ok := launcherMembers(selection, "backend"); !ok {
				return fail()
			}
			result.Notification = &LauncherNotification{Backend: "off"}
		case "host":
			if _, ok := launcherMembers(selection, "backend", "contract", "workflow_reference", "workflow_sha256"); !ok || selection["contract"] != "curator-host-notification/1" {
				return fail()
			}
			reference, validReference := selection["workflow_reference"].(string)
			digest, validDigest := selection["workflow_sha256"].(string)
			if !validReference || !launcherWorkflowReference(reference) || !validDigest || !canonicalSHA256RE.MatchString("sha256:"+digest) {
				return fail()
			}
			result.Notification = &LauncherNotification{Backend: "host", Contract: "curator-host-notification/1", WorkflowReference: reference, WorkflowSHA256: digest}
		default:
			return fail()
		}
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

func launcherWorkflowReference(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("._:/@+-", c) {
			continue
		}
		return false
	}
	return true
}
