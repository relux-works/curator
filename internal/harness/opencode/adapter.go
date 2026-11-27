// Package opencode delegates the OpenCode harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/opencode and
// exposes read-only MCP inspection descriptors.
package opencode

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/opencode"
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is the canonical Curator harness identifier for OpenCode CLI.
const ID harnessapi.ID = "opencode"

// Adapter wraps the upstream OpenCode harness adapter with read-only MCP
// inspection descriptors.
type Adapter struct{ opencode.HarnessAdapter }

// New constructs the OpenCode harness adapter.
func New() harnessapi.Adapter { return Adapter{} }

// Capabilities returns only actually ported operations.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.ReleaseProbing,
		harnessapi.MCPInspection,
	)
	return set
}

// MCP describes read-only inspection of native OpenCode JSON/JSONC configuration.
func (Adapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	switch scope {
	case harnessapi.ProjectScope:
		paths, err := namedPaths("opencode.json", "opencode.jsonc")
		if err != nil {
			return harnessapi.MCPChannel{}, err
		}
		return harnessapi.NewMCPChannel(paths, harnessapi.JSONC, true, false, false)
	case harnessapi.UserScope:
		paths, err := namedPaths(".config/opencode/opencode.json", ".config/opencode/opencode.jsonc")
		if err != nil {
			return harnessapi.MCPChannel{}, err
		}
		return harnessapi.NewMCPChannel(paths, harnessapi.JSONC, true, false, false)
	default:
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
}

func namedPaths(values ...string) (harnessapi.NamedPaths, error) {
	paths := make([]harnessvalue.RelativePath, 0, len(values))
	for _, value := range values {
		path, err := harnessvalue.NewRelativePath(value)
		if err != nil {
			return harnessapi.NamedPaths{}, err
		}
		paths = append(paths, path)
	}
	return harnessapi.NewNamedPaths(paths)
}
