// Package claude_code owns the Claude Code policy that harnessapi v1 can
// describe. Dynamic seed merging, launch arguments, and the per-home Keychain
// route remain with their existing shared owners until a later API revision.
package claude_code

import (
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// ID is this module's stable canonical identifier.
const ID harnessapi.ID = "claude_code"

// Adapter implements the Claude Code surfaces expressible by harnessapi v1.
type Adapter struct{ harnessapi.UnportedOperations }

// New constructs the adapter registered for Claude Code.
func New() harnessapi.Adapter { return Adapter{} }

// ID returns the canonical harness identifier.
func (Adapter) ID() harnessapi.ID { return ID }

// APIVersion returns the supported adapter contract version.
func (Adapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities lists only the v1 operations this adapter can describe.
func (Adapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.NativeMCPWrite,
		harnessapi.CredentialMetadata,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	)
	return set
}

// ReleaseProbe returns Claude Code's pinned read-only version command.
func (Adapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("claude", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

// Home returns the direct Claude configuration home and its default location.
func (Adapter) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	roots, err := namedPaths(".claude")
	if err != nil {
		return harnessapi.HomeLayout{}, err
	}
	return harnessapi.NewHomeLayout("CLAUDE_CONFIG_DIR", roots, false)
}

// Instructions describes the shared Claude root-instruction file.
func (Adapter) Instructions(_ harnessapi.Context, form harnessapi.Form) (harnessapi.InstructionSpec, error) {
	paths, err := namedPaths("CLAUDE.md")
	if err != nil {
		return harnessapi.InstructionSpec{}, err
	}
	return harnessapi.NewInstructionSpec(paths, form, nil)
}

// Skills describes Claude's native user and project skill locations.
func (Adapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	var path string
	var ignores []string
	switch scope {
	case harnessapi.UserScope:
		path = "skills"
	case harnessapi.ProjectScope:
		path = ".claude/skills"
		ignores = []string{path}
	default:
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(path)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignorePaths, err := namedPaths(ignores...)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	return harnessapi.NewSkillDiscovery(paths, ignorePaths, scope, false, true)
}

// Settings describes the native settings file for the requested scope.
func (Adapter) Settings(context harnessapi.Context) (harnessapi.SettingsSpec, error) {
	paths := []string{"settings.json"}
	if context.Scope() == harnessapi.ProjectScope {
		paths = []string{".claude/settings.json", ".claude/settings.local.json"}
	} else if context.Scope() != harnessapi.UserScope {
		return harnessapi.SettingsSpec{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	named, err := namedPaths(paths...)
	if err != nil {
		return harnessapi.SettingsSpec{}, err
	}
	return harnessapi.NewSettingsSpec(named, harnessapi.JSON, nil)
}

// MCP describes the Claude native MCP files. User-home seed writes remain in
// the shared owner because v1 cannot express their merge input.
func (Adapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	var path string
	var write, trust bool
	switch scope {
	case harnessapi.UserScope:
		path = ".claude.json"
	case harnessapi.ProjectScope:
		path, write, trust = ".mcp.json", true, true
	default:
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(path)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.JSON, true, write, trust)
}

// Credentials describes the Linux file-link route. The Darwin per-home
// Keychain strategy cannot be represented by the v1 CredentialRoute.
func (Adapter) Credentials(context harnessapi.Context) (harnessapi.CredentialRoute, error) {
	switch context.Platform() {
	case "linux":
		paths, err := namedPaths(".credentials.json")
		if err != nil {
			return harnessapi.CredentialRoute{}, err
		}
		return harnessapi.NewCredentialRoute(paths, "file-link", true)
	case "darwin":
		return harnessapi.CredentialRoute{}, harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	default:
		paths, err := namedPaths()
		if err != nil {
			return harnessapi.CredentialRoute{}, err
		}
		return harnessapi.NewCredentialRoute(paths, "unsupported", false)
	}
}

// Trust reports project admission as observation-only policy.
func (Adapter) Trust(context harnessapi.Context) (harnessapi.TrustPolicy, error) {
	return harnessapi.NewTrustPolicy(context.Scope() == harnessapi.ProjectScope), nil
}

// Preflight leaves generic trust, ownership, and transaction checks to their
// shared owners; this adapter adds no extra v1 refusal for the requested action.
func (Adapter) Preflight(harnessapi.Context, harnessapi.Snapshot, harnessapi.Request) ([]harnessapi.Refusal, error) {
	return []harnessapi.Refusal{}, nil
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
