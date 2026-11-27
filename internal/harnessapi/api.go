// Package harnessapi delegates the versioned value-only harness contract to
// github.com/relux-works/curator-runtimes/pkg/harnessapi.
package harnessapi

import upstream "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// ID identifies a registered harness.
type ID = upstream.ID

// APIVersion identifies the internal harness adapter contract revision.
type APIVersion = upstream.APIVersion

// Form identifies a root context delivery form.
type Form = upstream.Form

// Scope identifies user or project evaluation scope.
type Scope = upstream.Scope

// Capability identifies an implemented harness adapter capability.
type Capability = upstream.Capability

// Adapter is the versioned value-only harness policy interface.
type Adapter = upstream.Adapter

// Core harness contract constants delegated to curator-runtimes.
const (
	AdapterAPIVersionV1 = upstream.AdapterAPIVersionV1
	ProbeTimeoutV1      = upstream.ProbeTimeoutV1

	Monolithic = upstream.Monolithic
	Referenced = upstream.Referenced

	UserScope    = upstream.UserScope
	ProjectScope = upstream.ProjectScope

	ManagedHome          = upstream.ManagedHome
	InstructionOutput    = upstream.InstructionOutput
	SkillProjection      = upstream.SkillProjection
	NativeSkillDiscovery = upstream.NativeSkillDiscovery
	MCPInspection        = upstream.MCPInspection
	NativeMCPWrite       = upstream.NativeMCPWrite
	ProjectContextOutput = upstream.ProjectContextOutput
	HomeLease            = upstream.HomeLease
	CredentialMetadata   = upstream.CredentialMetadata
	TrustObservation     = upstream.TrustObservation
	ReleaseProbing       = upstream.ReleaseProbing
)
