// Package codex_cli delegates the Codex CLI harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/codex and
// exposes read-only skill projection and MCP inspection descriptors.
package codex_cli

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/codex"
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is the canonical Curator harness identifier for Codex CLI.
const ID harnessapi.ID = "codex_cli"

const (
	skillPath = ".codex/skills"
	mcpPath   = ".codex/config.toml"
)

// Adapter wraps the upstream Codex CLI harness adapter with read-only skill
// projection and MCP inspection descriptors.
type Adapter struct{ codex.HarnessAdapter }

// New constructs the Codex CLI harness adapter.
func New() harnessapi.Adapter { return Adapter{} }

// Capabilities returns only actually ported operations.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.ReleaseProbing,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
	)
	return set
}

// Skills describes the project and user skill projection path for Codex CLI.
func (Adapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(skillPath)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores = paths
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, false, true)
}

// MCP describes read-only inspection of the native Codex CLI TOML configuration.
func (Adapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(mcpPath)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.TOML, true, false, false)
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
