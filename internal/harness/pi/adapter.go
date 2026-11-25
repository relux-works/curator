// Package pi declares Pi's value-only harness policy. Shared callers retain
// filesystem effects and behaviors that need a later contract revision.
package pi

import (
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is this module's stable canonical identifier.
const ID harnessapi.ID = "pi"

// Adapter exposes the staged harness surface with explicit operational refusals.
type Adapter struct{ harnessapi.UnportedOperations }

// New constructs the module's staged adapter.
func New() harnessapi.Adapter { return Adapter{} }

// ID returns the module's canonical identifier.
func (Adapter) ID() harnessapi.ID { return ID }

// APIVersion returns the implemented internal contract revision.
func (Adapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.ManagedHome,
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.NativeSkillDiscovery,
		harnessapi.CredentialMetadata,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	)
	return set
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (Adapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("pi", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

func namedPaths(names ...string) (harnessapi.NamedPaths, error) {
	paths := make([]harnessvalue.RelativePath, 0, len(names))
	for _, name := range names {
		path, err := harnessvalue.NewRelativePath(name)
		if err != nil {
			return harnessapi.NamedPaths{}, err
		}
		paths = append(paths, path)
	}
	return harnessapi.NewNamedPaths(paths)
}

// Home declares the direct native home and its existing environment selector.
func (Adapter) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	roots, err := namedPaths(".pi")
	if err != nil {
		return harnessapi.HomeLayout{}, err
	}
	return harnessapi.NewHomeLayout("PI_CODING_AGENT_DIR", roots, false)
}

// Instructions declares the native root-context file. System-prompt file
// selection remains in its shared configuration owner until a later contract.
func (Adapter) Instructions(_ harnessapi.Context, form harnessapi.Form) (harnessapi.InstructionSpec, error) {
	if form != harnessapi.Monolithic {
		return harnessapi.InstructionSpec{}, harnessapi.NewRefusal("environment_form_unsupported")
	}
	paths, err := namedPaths("AGENTS.md")
	if err != nil {
		return harnessapi.InstructionSpec{}, err
	}
	return harnessapi.NewInstructionSpec(paths, form, nil)
}

// Skills declares Pi's native user-home skill directory. The legacy contract
// has no project-scope Pi skill path.
func (Adapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if scope != harnessapi.UserScope {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
	}
	paths, err := namedPaths("skills")
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores, err := namedPaths()
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, true, false)
}

// Credentials reports only the current native route. The shared credential
// owner retains link planning and old-root migration.
func (Adapter) Credentials(harnessapi.Context) (harnessapi.CredentialRoute, error) {
	paths, err := namedPaths("agent/auth.json")
	if err != nil {
		return harnessapi.CredentialRoute{}, err
	}
	return harnessapi.NewCredentialRoute(paths, "file-link", true)
}

// Trust declares that Pi has no project trust requirement; it grants none.
func (Adapter) Trust(harnessapi.Context) (harnessapi.TrustPolicy, error) {
	return harnessapi.NewTrustPolicy(false), nil
}

// MCP explicitly refuses the unsupported native MCP channel. This fact is
// independent of the supplied context and scope.
func (Adapter) MCP(harnessapi.Context, harnessapi.Scope) (harnessapi.MCPChannel, error) {
	return harnessapi.MCPChannel{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
}
