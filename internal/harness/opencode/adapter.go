// Package opencode describes OpenCode's v1 harness surfaces.
package opencode

import (
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is this module's stable canonical identifier.
const ID harnessapi.ID = "opencode"

// Adapter exposes the staged harness surface with explicit operational refusals.
type Adapter struct{ harnessapi.UnportedOperations }

// New constructs the module's staged adapter.
func New() harnessapi.Adapter { return Adapter{} }

// ID returns the module's canonical identifier.
func (Adapter) ID() harnessapi.ID { return ID }

// APIVersion returns the implemented internal contract revision.
func (Adapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns the operations represented by this adapter.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.ManagedHome,
		harnessapi.InstructionOutput,
		harnessapi.NativeSkillDiscovery,
		harnessapi.MCPInspection,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	)
	return set
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (Adapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("opencode", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

// Home returns the parent-valued XDG config home and OpenCode's child root.
func (Adapter) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	roots, err := namedPaths("opencode")
	if err != nil {
		return harnessapi.HomeLayout{}, err
	}
	return harnessapi.NewHomeLayout("XDG_CONFIG_HOME", roots, true)
}

// Instructions describes the root context and the referenced-form wrapper path.
func (Adapter) Instructions(_ harnessapi.Context, form harnessapi.Form) (harnessapi.InstructionSpec, error) {
	if form == "" {
		form = harnessapi.Monolithic
	}
	var paths []string
	switch form {
	case harnessapi.Monolithic:
		paths = []string{"AGENTS.md"}
	case harnessapi.Referenced:
		paths = []string{"AGENTS.md", "opencode.json"}
	default:
		return harnessapi.InstructionSpec{}, harnessapi.NewRefusal("environment_form_unsupported")
	}
	named, err := namedPaths(paths...)
	if err != nil {
		return harnessapi.InstructionSpec{}, err
	}
	return harnessapi.NewInstructionSpec(named, form, nil)
}

// Skills describes native .agents discovery and the existing user-scope mirror.
func (Adapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	paths, err := namedPaths(".agents/skills")
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores, err = namedPaths(".agents")
		if err != nil {
			return harnessapi.SkillDiscovery{}, err
		}
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, true, scope == harnessapi.UserScope)
}

// MCP describes OpenCode's read-only native inspection sources by scope.
func (Adapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	var paths []string
	switch scope {
	case harnessapi.UserScope:
		paths = []string{".config/opencode/opencode.json", ".config/opencode/opencode.jsonc"}
	case harnessapi.ProjectScope:
		paths = []string{"opencode.json", "opencode.jsonc"}
	default:
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	named, err := namedPaths(paths...)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(named, harnessapi.JSONC, true, false, scope == harnessapi.ProjectScope)
}

// Trust describes project admission without granting or changing trust.
func (Adapter) Trust(context harnessapi.Context) (harnessapi.TrustPolicy, error) {
	switch context.Scope() {
	case harnessapi.UserScope:
		return harnessapi.NewTrustPolicy(false), nil
	case harnessapi.ProjectScope:
		return harnessapi.NewTrustPolicy(true), nil
	default:
		return harnessapi.TrustPolicy{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
}

// Preflight leaves generic validation to the shared execution owner.
func (Adapter) Preflight(harnessapi.Context, harnessapi.Snapshot, harnessapi.Request) ([]harnessapi.Refusal, error) {
	return []harnessapi.Refusal{}, nil
}

// PlanSeed has no byte-file seed actions; XDG linked seeds remain in their shared owner.
func (Adapter) PlanSeed(harnessapi.Context, harnessapi.Snapshot, harnessapi.Request) (harnessapi.SeedPlan, error) {
	return harnessapi.NewSeedPlan(nil, nil)
}

// Status returns only the approved observations supplied by the shared collector.
func (Adapter) Status(_ harnessapi.Context, snapshot harnessapi.Snapshot) (harnessapi.Status, error) {
	return harnessapi.NewStatus(snapshot.Properties())
}

// Drift leaves ownership and file comparison to the shared marker engine.
func (Adapter) Drift(harnessapi.Context, harnessapi.Snapshot) (harnessapi.DriftReport, error) {
	paths, err := harnessapi.NewNamedPaths(nil)
	if err != nil {
		return harnessapi.DriftReport{}, err
	}
	return harnessapi.NewDriftReport(paths, nil)
}

func namedPaths(values ...string) (harnessapi.NamedPaths, error) {
	paths := make([]harnessvalue.RelativePath, len(values))
	for i, value := range values {
		path, err := harnessvalue.NewRelativePath(value)
		if err != nil {
			return harnessapi.NamedPaths{}, err
		}
		paths[i] = path
	}
	return harnessapi.NewNamedPaths(paths)
}
