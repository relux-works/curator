// Package gemini owns the native skill and MCP inspection facts for Gemini CLI.
package gemini

import (
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is this module's stable canonical identifier.
const ID harnessapi.ID = "gemini"

// Adapter exposes the staged harness surface with explicit operational refusals.
type Adapter struct{ harnessapi.UnportedOperations }

// New constructs the module's staged adapter.
func New() harnessapi.Adapter { return Adapter{} }

// ID returns the module's canonical identifier.
func (Adapter) ID() harnessapi.ID { return ID }

// APIVersion returns the implemented internal contract revision.
func (Adapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only the skill projection and MCP inspection operations.
func (Adapter) Capabilities() harnessapi.CapabilitySet { return Capabilities() }

// Capabilities returns the operations registered for this adapter.
func Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	return set
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (Adapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
}

// Skills describes the native destination and project ignore requirement.
func (Adapter) Skills(context harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if !validScope(context, scope) {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPath(".gemini/skills")
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores = paths
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, false, true)
}

// MCP describes the read-only JSON inspection surfaces and project trust gate.
func (Adapter) MCP(context harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	if !validScope(context, scope) {
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPath(".gemini/settings.json")
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.JSON, true, false, scope == harnessapi.ProjectScope)
}

func validScope(context harnessapi.Context, scope harnessapi.Scope) bool {
	return (scope == harnessapi.UserScope || scope == harnessapi.ProjectScope) && context.Scope() == scope
}

func namedPath(value string) (harnessapi.NamedPaths, error) {
	path, err := harnessvalue.NewRelativePath(value)
	if err != nil {
		return harnessapi.NamedPaths{}, err
	}
	return harnessapi.NewNamedPaths([]harnessvalue.RelativePath{path})
}
