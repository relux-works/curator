// Package harnessregistry owns the closed, compile-time harness catalog.
// Native operations are staged until their individual migrations are qualified.
package harnessregistry

import (
	"github.com/relux-works/curator/internal/harness/claude_code"
	"github.com/relux-works/curator/internal/harness/codex_cli"
	"github.com/relux-works/curator/internal/harness/cursor"
	"github.com/relux-works/curator/internal/harness/gemini"
	"github.com/relux-works/curator/internal/harness/muse"
	"github.com/relux-works/curator/internal/harness/opencode"
	"github.com/relux-works/curator/internal/harness/pi"
	"github.com/relux-works/curator/internal/harness/windsurf"
	"github.com/relux-works/curator/internal/harnessapi"
)

// DefaultID retains the existing default canonical identifier.
const DefaultID harnessapi.ID = codex_cli.ID

type registration struct {
	id           harnessapi.ID
	version      harnessapi.APIVersion
	capabilities harnessapi.CapabilitySet
	constructor  func() harnessapi.Adapter
	probe        harnessapi.ProbeSpec
	hasProbe     bool
}

func probing() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

func skillAndMCPInspection() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	return set
}

func nativeSkillsAndMCPInspection() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.NativeSkillDiscovery, harnessapi.MCPInspection)
	return set
}

// This is the only registration table. No runtime registration path exists.
// Probe aliases and arguments are pinned to the original environment catalog;
// the five-second deadline and empty overrides are the v1 contract policy.
var entries = [...]registration{
	{claude_code.ID, 1, probing(), claude_code.New, harnessapi.NewProbeSpec("claude", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), true},
	{codex_cli.ID, 1, probing(), codex_cli.New, harnessapi.NewProbeSpec("codex", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), true},
	{cursor.ID, 1, skillAndMCPInspection(), cursor.New, harnessapi.ProbeSpec{}, false},
	{gemini.ID, 1, harnessapi.CapabilitySet{}, gemini.New, harnessapi.ProbeSpec{}, false},
	{muse.ID, 1, probing(), muse.New, harnessapi.NewProbeSpec("muse", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), true},
	{opencode.ID, 1, probing(), opencode.New, harnessapi.NewProbeSpec("opencode", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), true},
	{pi.ID, 1, func() harnessapi.CapabilitySet {
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
	}(), pi.New, harnessapi.NewProbeSpec("pi", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), true},
	{windsurf.ID, 1, nativeSkillsAndMCPInspection(), windsurf.New, harnessapi.ProbeSpec{}, false},
}

// CanonicalID normalizes the existing CLI aliases without choosing a fallback.
func CanonicalID(id harnessapi.ID) harnessapi.ID {
	switch id {
	case "claude":
		return claude_code.ID
	case "codex":
		return codex_cli.ID
	}
	return id
}

// IDs returns an independent list of canonical registrations in static order.
func IDs() []harnessapi.ID {
	ids := make([]harnessapi.ID, len(entries))
	for i, entry := range entries {
		ids[i] = entry.id
	}
	return ids
}

// IDsWithCapability returns the static registrations that declare a capability.
func IDsWithCapability(capability harnessapi.Capability) []harnessapi.ID {
	var ids []harnessapi.ID
	for _, entry := range entries {
		if entry.capabilities.Has(capability) {
			ids = append(ids, entry.id)
		}
	}
	return ids
}

// Lookup returns only a checked static registration or a typed refusal.
func Lookup(id harnessapi.ID) (harnessapi.Adapter, error) {
	id = CanonicalID(id)
	for _, entry := range entries {
		if entry.id == id {
			return verify(entry.version, entry.id, entry.capabilities, entry.constructor)
		}
	}
	return nil, harnessapi.NewRefusal(harnessapi.UnknownID)
}

// verify checks registration metadata in refusal order without adding an entry
// to the catalog. Keeping the verifier separate makes the version guard and
// zero-operation contract directly testable with inert constructors.
func verify(version harnessapi.APIVersion, id harnessapi.ID, capabilities harnessapi.CapabilitySet, constructor func() harnessapi.Adapter) (harnessapi.Adapter, error) {
	if version != harnessapi.AdapterAPIVersionV1 {
		return nil, harnessapi.NewRefusal(harnessapi.APIVersionUnsupported)
	}
	if constructor == nil {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	adapter := constructor()
	if adapter == nil {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	// Version precedes every operational call and other metadata comparisons.
	if adapter.APIVersion() != version {
		return nil, harnessapi.NewRefusal(harnessapi.APIVersionUnsupported)
	}
	if adapter.ID() != id || adapter.Capabilities() != capabilities {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	return adapter, nil
}

// ProbePolicy returns a defensive copy of the canonical ID's closed probe tuple.
func ProbePolicy(id harnessapi.ID) (harnessapi.ProbeSpec, error) {
	id = CanonicalID(id)
	for _, entry := range entries {
		if entry.id == id {
			if !entry.hasProbe {
				return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
			}
			return harnessapi.NewProbeSpec(entry.probe.ExecutableAlias(), entry.probe.Arguments(), entry.probe.Timeout(), entry.probe.Environment()), nil
		}
	}
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.UnknownID)
}
