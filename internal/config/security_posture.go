package config

import "fmt"

// SecurityPostureRevision is the one release switch for the machine posture
// default. Revision A ships first with the permissive warning; Revision B is
// the later default flip owned by its bounded rollout task.
const SecurityPostureRevision = "A"

const (
	// SecurityPosturePermissive selects today's default-open gate values.
	SecurityPosturePermissive = "permissive"
	// SecurityPostureHardened selects the strict end of each posture gate.
	SecurityPostureHardened = "hardened"
	// DiagSecurityPosturePermissive identifies the revision-A migration warning.
	DiagSecurityPosturePermissive = "security_posture_permissive"
	// DiagSourceAllowlistEmpty identifies the hardened install/update refusal.
	DiagSourceAllowlistEmpty = "source_allowlist_empty"
	// DiagMCPAllowlistEmpty identifies the hardened MCP-resolution refusal.
	DiagMCPAllowlistEmpty = "mcp_package_allowlist_empty"
	// DiagPassableEnvUnbounded identifies explicit-null passthrough refusal.
	DiagPassableEnvUnbounded = "passable_env_names_unbounded_refused" // #nosec G101 -- diagnostic identifier, not a credential.
	// SecurityPostureMigrationHint is shown while revision A defaults permissive.
	SecurityPostureMigrationHint = "set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default"
)

// SecurityPostureDiagnostic is a posture operation finding. Optional fields
// are emitted only when that diagnostic defines them in manager §7.1.
type SecurityPostureDiagnostic struct {
	Code          string `json:"code"`
	Severity      string `json:"severity"`
	Count         *int   `json:"count,omitempty"`
	NamesKnob     bool   `json:"names_knob,omitempty"`
	MigrationHint bool   `json:"migration_hint,omitempty"`
}

// SecurityPostureRow is one closed manager status row.
type SecurityPostureRow struct {
	Gate   string `json:"gate"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

// SecurityPostureRevisions are the independently shipped gate revisions
// rendered beside the posture policy. Status callers supply their actual
// compiled revisions; conformance vectors supply the revision tuple they pin.
type SecurityPostureRevisions struct {
	HookTrust          string `json:"hook_trust"`
	EnvPassthrough     string `json:"env_passthrough"`
	ProviderTrustRoots string `json:"provider_trust_roots"`
	UpdateConfirmation string `json:"update_confirmation"`
	CodexSeed          string `json:"codex_seed"`
}

func defaultSecurityPosture(schema int) string {
	if schema == SchemaVersion {
		return SecurityPosturePermissive
	}
	switch SecurityPostureRevision {
	case "A":
		return SecurityPosturePermissive
	case "B":
		return SecurityPostureHardened
	default:
		panic("unknown security posture revision " + SecurityPostureRevision)
	}
}

// EffectiveSecurityPosture returns the effective profile. A zero Config is
// treated like an older permissive caller so adding this field does not turn
// existing programmatic callers into hardened operations accidentally.
func (c *Config) EffectiveSecurityPosture() string {
	if c == nil || c.SecurityPosture == "" {
		return SecurityPosturePermissive
	}
	return c.SecurityPosture
}

// SecurityPostureHardened exposes the effective posture to downstream
// resolvers without requiring them to reconstruct defaults or provenance.
func (c *Config) SecurityPostureHardened() bool {
	return c != nil && c.EffectiveSecurityPosture() == SecurityPostureHardened
}

// SecurityPostureWarning returns the revision-A migration warning once per
// loaded manager operation when the effective posture is permissive.
func (c *Config) SecurityPostureWarning() string {
	if c == nil || c.SecurityPostureHardened() {
		return ""
	}
	return fmt.Sprintf("%s: security_posture is permissive; %s", DiagSecurityPosturePermissive, SecurityPostureMigrationHint)
}

// SecurityPostureSources returns a copy of the per-knob provenance map.
func (c *Config) SecurityPostureSources() map[string]string {
	result := make(map[string]string, len(c.postureSources))
	for knob, source := range c.postureSources {
		result[knob] = source
	}
	return result
}

// SecurityPostureDiagnostics computes posture-specific diagnostics for one
// manager operation. The permissive warning is operation-scoped. Empty MCP
// allowlists retain their warning when no MCP declarations are present;
// hardened contradictions refuse only the affected operation.
func (c *Config) SecurityPostureDiagnostics(operation string, mcpDeclarationsPresent bool) []SecurityPostureDiagnostic {
	if c == nil {
		return nil
	}
	diagnostics := []SecurityPostureDiagnostic{}
	if !c.SecurityPostureHardened() {
		count := 1
		diagnostics = append(diagnostics, SecurityPostureDiagnostic{
			Code: DiagSecurityPosturePermissive, Severity: "warning", Count: &count,
			NamesKnob: true, MigrationHint: true,
		})
		return diagnostics
	}

	switch operation {
	case "install", "update":
		if len(c.AllowedSources) == 0 {
			diagnostics = append(diagnostics, SecurityPostureDiagnostic{Code: DiagSourceAllowlistEmpty, Severity: "error"})
		}
	case "profile-install", "profile-update", "launch-composition":
		if mcpDeclarationsPresent && len(c.Env.MCPPackageAllowlist) == 0 {
			diagnostics = append(diagnostics, SecurityPostureDiagnostic{Code: DiagMCPAllowlistEmpty, Severity: "error"})
		} else if !mcpDeclarationsPresent && len(c.Env.MCPPackageAllowlist) == 0 {
			diagnostics = append(diagnostics, SecurityPostureDiagnostic{Code: DiagMCPAllowlistEmpty, Severity: "warning"})
		}
		if c.Env.PassableEnvNamesSet && c.Env.PassableEnvNames == nil {
			diagnostics = append(diagnostics, SecurityPostureDiagnostic{
				Code: DiagPassableEnvUnbounded, Severity: "error", NamesKnob: true,
			})
		}
	}
	return diagnostics
}

// CheckSecurityPostureSourceAllowlist enforces the install/update source gate.
func CheckSecurityPostureSourceAllowlist(posture string, allowedSources []string) error {
	if posture == SecurityPostureHardened && len(allowedSources) == 0 {
		return fmt.Errorf("%s: operation refused under the hardened security posture", DiagSourceAllowlistEmpty)
	}
	return nil
}

// CheckSecurityPostureMCPAllowlist enforces the closure-dependent MCP gate.
func CheckSecurityPostureMCPAllowlist(posture string, allowlist []string, declarationsPresent bool) error {
	if posture == SecurityPostureHardened && declarationsPresent && len(allowlist) == 0 {
		return fmt.Errorf("%s: operation refused under the hardened security posture", DiagMCPAllowlistEmpty)
	}
	return nil
}

// CheckSecurityPosturePassableEnv refuses explicit unbounded passthrough
// under hardened posture, while absent values remain governed by S4.
func CheckSecurityPosturePassableEnv(posture string, names []string, set bool) error {
	if posture == SecurityPostureHardened && set && names == nil {
		return fmt.Errorf("%s: operation refused under the hardened security posture", DiagPassableEnvUnbounded)
	}
	return nil
}

// SecurityPostureContradiction reports an effective configuration that makes
// an environment status check non-current. A profile status without --check
// remains informational; source allowlist emptiness affects install/update,
// while MCP declarations and explicit unbounded passthrough affect the
// currently resolved environment state.
func (c *Config) SecurityPostureContradiction(mcpDeclarationsPresent bool) bool {
	if !c.SecurityPostureHardened() {
		return false
	}
	return len(c.AllowedSources) == 0 ||
		(mcpDeclarationsPresent && len(c.Env.MCPPackageAllowlist) == 0) ||
		(c.Env.PassableEnvNamesSet && c.Env.PassableEnvNames == nil)
}

// CheckSecurityPosture returns the first operation refusal, if any.
func (c *Config) CheckSecurityPosture(operation string, mcpDeclarationsPresent bool) error {
	if c == nil {
		return nil
	}
	posture := c.EffectiveSecurityPosture()
	switch operation {
	case "install", "update":
		return CheckSecurityPostureSourceAllowlist(posture, c.AllowedSources)
	case "profile-install", "profile-update", "launch-composition":
		if err := CheckSecurityPostureMCPAllowlist(posture, c.Env.MCPPackageAllowlist, mcpDeclarationsPresent); err != nil {
			return err
		}
		return CheckSecurityPosturePassableEnv(posture, c.Env.PassableEnvNames, c.Env.PassableEnvNamesSet)
	}
	return nil
}

// SecurityPostureStatusRows returns the ordered inventory shared by curator
// status and env status. Schema 1 carries only its four legacy gates and no
// env-status inventory.
func (c *Config) SecurityPostureStatusRows(revisions SecurityPostureRevisions) []SecurityPostureRow {
	if c == nil {
		return nil
	}
	postureSource := c.SecurityPostureSource
	if postureSource == "" {
		postureSource = "profile"
	}
	rows := []SecurityPostureRow{
		{Gate: "security_posture", Value: c.EffectiveSecurityPosture(), Source: postureSource},
		{Gate: "hook-trust", Value: revisions.HookTrust, Source: "shipped"},
		{Gate: "registry-policy", Value: c.Audit.RegistryPolicy, Source: c.PostureSourceFor("audit.registry_policy")},
		{Gate: "audit-mode", Value: c.Audit.Mode, Source: c.PostureSourceFor("audit.mode")},
		{Gate: "source-allowlist", Value: len(c.AllowedSources), Source: c.PostureSourceFor("allowed_sources")},
	}
	if c.Schema == SchemaVersion {
		return rows
	}
	return append(rows,
		SecurityPostureRow{Gate: "env-passthrough", Value: revisions.EnvPassthrough, Source: "shipped"},
		SecurityPostureRow{Gate: "transitive-system-modules", Value: c.Env.TransitiveSystemModules, Source: c.PostureSourceFor("environments.transitive_system_modules")},
		SecurityPostureRow{Gate: "provider-trust-roots", Value: revisions.ProviderTrustRoots, Source: "shipped"},
		SecurityPostureRow{Gate: "source-signers", Value: fmt.Sprintf("%t", c.Env.RequireSourceSigners), Source: c.PostureSourceFor("environments.require_source_signers")},
		SecurityPostureRow{Gate: "update-confirmation", Value: revisions.UpdateConfirmation, Source: "shipped"},
		SecurityPostureRow{Gate: "codex-seed", Value: revisions.CodexSeed, Source: "shipped"},
		SecurityPostureRow{Gate: "store-boundary", Value: "enforced", Source: "shipped"},
		SecurityPostureRow{Gate: "write-discipline", Value: "enforced", Source: "shipped"},
		SecurityPostureRow{Gate: "mcp-package-allowlist", Value: len(c.Env.MCPPackageAllowlist), Source: c.PostureSourceFor("environments.mcp_package_allowlist")},
	)
}

// PostureSourceFor returns a gate's provenance, defaulting unknown or unset
// fields to profile provenance.
func (c *Config) PostureSourceFor(knob string) string {
	if c == nil {
		return "profile"
	}
	if source := c.postureSources[knob]; source != "" {
		return source
	}
	return "profile"
}

func hasConfigKey(data map[string]any, key string) bool {
	if data == nil {
		return false
	}
	_, ok := data[key]
	return ok
}

func hasNestedConfigKey(data map[string]any, object, key string) bool {
	if data == nil {
		return false
	}
	values, _ := data[object].(map[string]any)
	if values == nil {
		return false
	}
	_, ok := values[key]
	return ok
}

func securityPostureSource(effective, machine, system map[string]any, locked map[string]bool, schema int) string {
	if schema == SchemaVersion {
		return "profile"
	}
	if locked["security_posture"] {
		return "lock"
	}
	if hasConfigKey(machine, "security_posture") || hasConfigKey(system, "security_posture") || hasConfigKey(effective, "security_posture") {
		return "explicit"
	}
	return "profile"
}

func securityPostureSources(effective, machine, system map[string]any, locked map[string]bool) map[string]string {
	keys := []struct {
		name      string
		lock      string
		top       string
		topNested string
		env       string
	}{
		{name: "security_posture", lock: "security_posture", top: "security_posture"},
		{name: "audit.mode", lock: "audit", top: "audit", topNested: "mode"},
		{name: "audit.registry_policy", lock: "audit", top: "audit", topNested: "registry_policy"},
		{name: "allowed_sources", lock: "allowed_sources", top: "allowed_sources"},
		{name: "environments.mcp_package_allowlist", lock: "environments.mcp_package_allowlist", env: "mcp_package_allowlist"},
		{name: "environments.passable_env_names", lock: "environments.passable_env_names", env: "passable_env_names"},
		{name: "environments.transitive_system_modules", lock: "environments.transitive_system_modules", env: "transitive_system_modules"},
		{name: "environments.require_source_signers", lock: "environments.require_source_signers", env: "require_source_signers"},
	}
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		switch {
		case locked[key.lock]:
			result[key.name] = "lock"
		case key.env != "" && hasNestedConfigKey(machine, "environments", key.env):
			result[key.name] = "explicit"
		case key.env != "" && hasNestedConfigKey(system, "environments", key.env):
			result[key.name] = "explicit"
		case key.env != "" && hasNestedConfigKey(effective, "environments", key.env):
			result[key.name] = "explicit"
		case key.topNested != "" && hasNestedConfigKey(machine, key.top, key.topNested):
			result[key.name] = "explicit"
		case key.topNested != "" && hasNestedConfigKey(system, key.top, key.topNested):
			result[key.name] = "explicit"
		case key.topNested != "" && hasNestedConfigKey(effective, key.top, key.topNested):
			result[key.name] = "explicit"
		case key.top != "" && key.topNested == "" && hasConfigKey(machine, key.top):
			result[key.name] = "explicit"
		case key.top != "" && key.topNested == "" && hasConfigKey(system, key.top):
			result[key.name] = "explicit"
		case key.top != "" && key.topNested == "" && hasConfigKey(effective, key.top):
			result[key.name] = "explicit"
		default:
			result[key.name] = "profile"
		}
	}
	return result
}

func applyHardenedDefaults(cfg *Config, data map[string]any) {
	if cfg == nil || !cfg.SecurityPostureHardened() {
		return
	}
	if !hasNestedConfigKey(data, "audit", "mode") {
		cfg.Audit.Mode = "strict"
	}
	if !hasNestedConfigKey(data, "audit", "registry_policy") {
		cfg.Audit.RegistryPolicy = "strict"
	}
	if !hasNestedConfigKey(data, "environments", "transitive_system_modules") {
		cfg.Env.TransitiveSystemModules = "error"
	}
	if !hasNestedConfigKey(data, "environments", "require_source_signers") {
		cfg.Env.RequireSourceSigners = true
	}
}
