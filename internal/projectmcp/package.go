package projectmcp

import (
	"encoding/json"
	"path/filepath"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/stateread"
)

// LoadPackage composes the unchanged legacy package validator (including byte
// bounds) with the additional stage-A server gates. Inline scalar bounds do not
// loosen released package acceptance.
func LoadPackage(root string) (*contextpkg.MCPManifest, error) {
	file, err := stateread.ReadFile(filepath.Join(root, contextpkg.MCPManifestName))
	if err != nil {
		return nil, err
	}
	if file.Kind == stateread.KindAbsent {
		return nil, invalid("resolved agent-mcp.json is missing")
	}
	payload := file.Bytes
	manifest, err := contextpkg.ParseMCP(payload)
	if err != nil {
		return nil, invalid("resolved agent-mcp.json")
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, invalid("resolved agent-mcp.json")
	}
	normalized, err := NormalizeServer(raw["server"])
	if err != nil {
		return nil, err
	}
	manifest.Server.Args, _ = normalized["args"].([]string)
	manifest.Server.EnvNames, _ = normalized["env_names"].([]string)
	manifest.Server.Environments, _ = normalized["environments"].([]string)
	return manifest, nil
}

// LoadNamedPackage also binds the declaration NAME; package selection does not
// support renaming or last-writer-wins aliases.
func LoadNamedPackage(root, name string) (*contextpkg.MCPManifest, error) {
	manifest, err := LoadPackage(root)
	if err != nil {
		return nil, err
	}
	if manifest.Name != name {
		return nil, invalid("resolved package NAME differs")
	}
	return manifest, nil
}

// NormalizePackageServer composes the stage-A scalar rules with the released
// package validator, so a carrier cannot admit bytes rejected at acquisition.
func NormalizePackageServer(raw any) (map[string]any, error) {
	normalized, err := NormalizeServer(raw)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{"schema_version": 1, "name": "server", "version": "1.0.0", "server": normalized})
	if err != nil {
		return nil, invalid("resolved package server")
	}
	if _, err := contextpkg.ParseMCP(payload); err != nil {
		return nil, invalid("resolved package server")
	}
	return normalized, nil
}
