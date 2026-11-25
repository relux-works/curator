// Package codex_cli owns the Codex CLI facts expressible by harnessapi v1.
package codex_cli

import (
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

const (
	defaultHome       = ".codex"
	homeSkillsPath    = "skills"
	projectSkillsPath = ".codex/skills"
	instructionsPath  = "AGENTS.md"
	projectMCPPath    = ".codex/config.toml"
	homeMCPPath       = "config.toml"
)

// ID is this module's stable canonical identifier.
const ID harnessapi.ID = "codex_cli"

// Adapter exposes the staged harness surface with explicit operational refusals.
type Adapter struct{ harnessapi.UnportedOperations }

// New constructs the module's staged adapter.
func New() harnessapi.Adapter { return Adapter{} }

// ID returns the module's canonical identifier.
func (Adapter) ID() harnessapi.ID { return ID }

// APIVersion returns the implemented internal contract revision.
func (Adapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only the Codex surfaces described by the v1 contract.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.ManagedHome,
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.CredentialMetadata,
		harnessapi.ReleaseProbing,
	)
	return set
}

// Home describes the Codex home selector and its default directory.
func (Adapter) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	roots, err := namedPaths(defaultHome)
	if err != nil {
		return harnessapi.HomeLayout{}, err
	}
	return harnessapi.NewHomeLayout("CODEX_HOME", roots, false)
}

// Instructions describes the monolithic root instruction file.
func (Adapter) Instructions(_ harnessapi.Context, form harnessapi.Form) (harnessapi.InstructionSpec, error) {
	if form != harnessapi.Monolithic {
		return harnessapi.InstructionSpec{}, harnessapi.NewRefusal("environment_form_unsupported")
	}
	paths, err := namedPaths(instructionsPath)
	if err != nil {
		return harnessapi.InstructionSpec{}, err
	}
	return harnessapi.NewInstructionSpec(paths, harnessapi.Monolithic, nil)
}

// Skills describes the Codex project skill mirror.
func (Adapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	path := projectSkillsPath
	if scope == harnessapi.UserScope {
		path = homeSkillsPath
	}
	paths, err := namedPaths(path)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores = paths
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, false, true)
}

// MCP describes read-only discovery of Codex TOML configuration sources.
func (Adapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	path := homeMCPPath
	if scope == harnessapi.ProjectScope {
		path = projectMCPPath
	}
	paths, err := namedPaths(path)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.TOML, true, false, scope == harnessapi.ProjectScope)
}

// Credentials reports only the known auth-file route metadata.
func (Adapter) Credentials(harnessapi.Context) (harnessapi.CredentialRoute, error) {
	paths, err := namedPaths("auth.json")
	if err != nil {
		return harnessapi.CredentialRoute{}, err
	}
	return harnessapi.NewCredentialRoute(paths, "keyring-preferred", true)
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (Adapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("codex", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
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
