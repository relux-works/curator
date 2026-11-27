package harnessapi

import (
	"time"

	upstream "github.com/relux-works/curator-runtimes/pkg/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

// Refusal codes shared across Curator harness adapters.
const (
	UnknownID             = upstream.UnknownID
	APIVersionUnsupported = upstream.APIVersionUnsupported
	RegistryMismatch      = upstream.RegistryMismatch
	CapabilityUnsupported = upstream.CapabilityUnsupported
	ProbeSpecUnsupported  = upstream.ProbeSpecUnsupported
	RuntimeMissing        = upstream.RuntimeMissing
)

// Refusal is a typed refusal error returned by harness adapters.
type Refusal = upstream.Refusal

// CapabilitySet is an immutable validated set of harness capabilities.
type CapabilitySet = upstream.CapabilitySet

// ProbeSpec describes a harness release-probe invocation.
type ProbeSpec = upstream.ProbeSpec

// VerifiedRelease holds a validated harness release identifier.
type VerifiedRelease = upstream.VerifiedRelease

// ProbeOutput holds captured release-probe stdout, stderr, and exit code.
type ProbeOutput = upstream.ProbeOutput

// Context carries platform, scope, and verified release inputs to an adapter.
type Context = upstream.Context

// Property is a validated key-value harness property.
type Property = upstream.Property

// Snapshot is an immutable collection of observed harness properties.
type Snapshot = upstream.Snapshot

// Request describes a requested harness operation, form, and scope.
type Request = upstream.Request

// NamedPaths is an ordered collection of validated relative paths.
type NamedPaths = upstream.NamedPaths

// HomeLayout describes a harness managed-home environment and roots.
type HomeLayout = upstream.HomeLayout

// InstructionSpec describes harness root instruction paths, form, and notice bytes.
type InstructionSpec = upstream.InstructionSpec

// SkillDiscovery describes native skill discovery and projection paths.
type SkillDiscovery = upstream.SkillDiscovery

// Format identifies a harness configuration serialization format.
type Format = upstream.Format

// SettingsSpec describes harness settings paths, format, and managed fields.
type SettingsSpec = upstream.SettingsSpec

// MCPChannel describes harness MCP inspection and write configuration paths.
type MCPChannel = upstream.MCPChannel

// CredentialRoute describes a harness credential routing strategy and paths.
type CredentialRoute = upstream.CredentialRoute

// TrustPolicy declares whether a harness requires workspace trust records.
type TrustPolicy = upstream.TrustPolicy

// FileAction describes one staged file mutation action with preimage and mode.
type FileAction = upstream.FileAction

// SeedPlan describes staged seed file actions and typed refusals.
type SeedPlan = upstream.SeedPlan

// ProjectContextPlan describes project context projection actions and protected paths.
type ProjectContextPlan = upstream.ProjectContextPlan

// Status describes observed harness status properties.
type Status = upstream.Status

// DriftReport describes harness configuration drift paths and refusals.
type DriftReport = upstream.DriftReport

// Supported harness configuration file formats.
const (
	JSON     = upstream.JSON
	TOML     = upstream.TOML
	JSONC    = upstream.JSONC
	Markdown = upstream.Markdown
)

// NewRefusal constructs a typed harness refusal error.
func NewRefusal(code string) Refusal { return upstream.NewRefusal(code) }

// NewCapabilitySet validates and builds an immutable set of harness capabilities.
func NewCapabilitySet(values ...Capability) (CapabilitySet, error) {
	return upstream.NewCapabilitySet(values...)
}

// NewProbeSpec constructs a release-probe command specification.
func NewProbeSpec(alias string, args []string, timeout time.Duration, env map[string]string) ProbeSpec {
	return upstream.NewProbeSpec(alias, args, timeout, env)
}

// NewVerifiedRelease validates a harness release identifier.
func NewVerifiedRelease(version string) (VerifiedRelease, error) {
	return upstream.NewVerifiedRelease(version)
}

// NewProbeOutput records captured stdout, stderr, and exit code from a probe.
func NewProbeOutput(stdout, stderr []byte, exitCode int) (ProbeOutput, error) {
	return upstream.NewProbeOutput(stdout, stderr, exitCode)
}

// NewContext constructs a bounded adapter evaluation context.
func NewContext(platform string, scope Scope, release VerifiedRelease) (Context, error) {
	return upstream.NewContext(platform, scope, release)
}

// NewProperty constructs a validated key-value harness property.
func NewProperty(name, value string) (Property, error) {
	return upstream.NewProperty(name, value)
}

// NewSnapshot constructs an immutable property snapshot.
func NewSnapshot(properties []Property) (Snapshot, error) {
	return upstream.NewSnapshot(properties)
}

// NewRequest constructs an adapter operation request.
func NewRequest(operation Capability, form Form, scope Scope) (Request, error) {
	return upstream.NewRequest(operation, form, scope)
}

// NewNamedPaths validates an ordered collection of relative paths.
func NewNamedPaths(paths []harnessvalue.RelativePath) (NamedPaths, error) {
	return upstream.NewNamedPaths(paths)
}

// NewHomeLayout constructs a harness home layout declaration.
func NewHomeLayout(environment string, roots NamedPaths, parent bool) (HomeLayout, error) {
	return upstream.NewHomeLayout(environment, roots, parent)
}

// NewInstructionSpec constructs a harness root instruction file specification.
func NewInstructionSpec(paths NamedPaths, form Form, notice []byte) (InstructionSpec, error) {
	return upstream.NewInstructionSpec(paths, form, notice)
}

// NewSkillDiscovery constructs a skill discovery and projection declaration.
func NewSkillDiscovery(paths, ignores NamedPaths, scope Scope, native, mirror bool) (SkillDiscovery, error) {
	return upstream.NewSkillDiscovery(paths, ignores, scope, native, mirror)
}

// NewSettingsSpec constructs a harness settings specification.
func NewSettingsSpec(paths NamedPaths, format Format, fields []Property) (SettingsSpec, error) {
	return upstream.NewSettingsSpec(paths, format, fields)
}

// NewMCPChannel constructs an MCP configuration channel declaration.
func NewMCPChannel(paths NamedPaths, format Format, inspect, write, trust bool) (MCPChannel, error) {
	return upstream.NewMCPChannel(paths, format, inspect, write, trust)
}

// NewCredentialRoute constructs a harness credential-routing declaration.
func NewCredentialRoute(paths NamedPaths, strategy string, qualified bool) (CredentialRoute, error) {
	return upstream.NewCredentialRoute(paths, strategy, qualified)
}

// NewTrustPolicy constructs a workspace trust policy value.
func NewTrustPolicy(required bool) TrustPolicy { return upstream.NewTrustPolicy(required) }

// NewFileAction constructs a validated file mutation action.
func NewFileAction(path harnessvalue.RelativePath, bytes []byte, preimage string, mode uint32) (FileAction, error) {
	return upstream.NewFileAction(path, bytes, preimage, mode)
}

// NewSeedPlan constructs a validated seed file plan.
func NewSeedPlan(actions []FileAction, refusals []Refusal) (SeedPlan, error) {
	return upstream.NewSeedPlan(actions, refusals)
}

// NewProjectContextPlan constructs a project context projection plan.
func NewProjectContextPlan(plan SeedPlan, protected NamedPaths) ProjectContextPlan {
	return upstream.NewProjectContextPlan(plan, protected)
}

// NewStatus constructs a harness status report.
func NewStatus(properties []Property) (Status, error) { return upstream.NewStatus(properties) }

// NewDriftReport constructs a harness drift report.
func NewDriftReport(paths NamedPaths, refusals []Refusal) (DriftReport, error) {
	return upstream.NewDriftReport(paths, refusals)
}
