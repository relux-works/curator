// Package envregistry is the closed revision-1 environment adapter
// registry (environments §7). Adapters are manager code, never package
// code: every per-adapter fact — home mechanism, root-context target,
// supported forms, system-prompt and MCP channel descriptors, credential
// passthrough, provisioning seeds, the isolation matrix, shadowing paths,
// secondary targets, size advisories, and recorded tool releases — lives in
// this table, validated against the section it cites. Profile bytes never
// select an adapter fact; machine configuration selects among the declared
// values with the section 12.1 defaults.
package envregistry

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// Environment identifiers (environments §7.1, closed in revision 1).
const (
	ClaudeCode = "claude_code"
	CodexCLI   = "codex_cli"
	OpenCode   = "opencode"
	Pi         = "pi"
)

// Diagnostics (environments §7.7, §10.4, §11.1).
const (
	DiagUnknown              = "environment_unknown"
	DiagFormUnsupported      = "environment_form_unsupported"
	DiagIsolatedUnsupported  = "environment_isolated_unsupported"
	DiagSharedUnsupported    = "environment_shared_unsupported"
	DiagTargetUnknown        = "environment_target_unknown"
	DiagTargetConsent        = "environment_target_consent_required"
	DiagShadowingPresent     = "environment_shadowing_path_present"
	DiagPassthroughDetached  = "environment_passthrough_detached"
	DiagSeedUnreadable       = "environment_seed_unreadable"
	DiagSeedShadowed         = "environment_seed_shadowed"
	DiagToolVersionUnverifed = "environment_tool_version_unverified"
	DiagSizeExceeded         = "environment_context_size_exceeded"
	DiagReservedCommand      = "environment_reserved_command_name"
)

// Forms (environments §7.2).
const (
	FormMonolithic = "monolithic"
	FormReferenced = "referenced"
)

// Isolation modes (environments §7.4).
const (
	IsolationShared   = "shared"
	IsolationIsolated = "isolated"
)

// Descriptor kinds and semantics (environments §7.3).
const (
	KindFlag      = "flag"
	KindConfigKey = "config-key"
	KindVariable  = "variable"
	KindFile      = "file"

	SemanticsAppend  = "append"
	SemanticsReplace = "replace"

	ArgumentPath     = "path"
	ArgumentContents = "contents"
	ArgumentName     = "name"
)

// Channel is one closed channel descriptor (environments §7.3): a system-
// prompt descriptor carries Semantics, an MCP descriptor does not.
type Channel struct {
	Kind      string
	Semantics string
	Flag      string
	Argument  string
	Name      string
	With      []string
	Key       string
	Variable  string
	Filename  string
}

// Passthrough is one credential passthrough entry of an adapter
// (environments §7.4).
type Passthrough struct {
	// Path is the home-relative entry shared with the native home.
	Path string
	// Strategy is per-home-keychain, ambient, keyring-preferred, or
	// file-link. FileLinkTarget records which native entry a file-link
	// points at; empty for the other strategies.
	Strategy       string
	FileLinkTarget string
}

// Strategies (environments §7.4).
const (
	StrategyPerHomeKeychain  = "per-home-keychain"
	StrategyAmbient          = "ambient"
	StrategyKeyringPreferred = "keyring-preferred"
	StrategyFileLink         = "file-link"
)

// Shadow is one declared shadowing path (environments §7.5).
type Shadow struct {
	// Path is home-relative; Surface is the shadowed surface key.
	Path    string
	Surface string
}

// Target is one secondary fixed-home target (environments §7.6). The
// table holds one row per adapter: revision 1 declares two rows sharing
// the xcode-coding-assistant identifier, one for claude_code (home as
// CLAUDE_CONFIG_DIR) and one for codex_cli (the same directory as
// CODEX_HOME).
type Target struct {
	ID string
	// Adapter is the environment the target applies to; a target never
	// applies to any other adapter.
	Adapter string
	Probe   string
	Home    string
	// Surfaces honored by the embedded host.
	Surfaces []string
	// Ungoverned names what the embedded host owns that the manager never
	// audits: MCP configuration and commands/.
	Ungoverned string
}

// Adapter is the closed per-adapter record.
type Adapter struct {
	ID string
	// EnvVar is the fragment variable naming the home. For opencode it
	// names the managed parent: the tool reads <parent>/opencode/ as the
	// home (environments §7.1).
	EnvVar string
	// ParentVar is true where EnvVar names a parent of the home.
	ParentVar bool
	// RootTarget is the root-context file name in the home.
	RootTarget string
	// SkillsDir is the skills directory name in the home; empty where the
	// adapter uses the machine-global native surface (opencode,
	// split-brain by construction, §7.1).
	SkillsDir   string
	Forms       []string
	DefaultForm string
	// SystemPrompt holds the §7.3 channel descriptors, nil for opencode.
	SystemPrompt []Channel
	// MCP holds the single §7.8 channel descriptor, nil for pi.
	MCP *Channel
	// Passthrough holds the per-platform entries. The key is the GOOS
	// value; "default" applies where no platform row exists.
	Passthrough map[string][]Passthrough
	// Seeds holds the per-adapter provisioning seed class (§7.4).
	Seeds []string
	// SeedWritten marks seeds the manager writes rather than copies
	// (claude_code .claude.json).
	SeedWritten map[string]bool
	Shadows     []Shadow
	// VerifiedRelease is the tool release the facts were verified on
	// (§7.9); empty where the tool was not installed on the recording
	// machine (opencode, docs-confidence throughout).
	VerifiedRelease string
	// Probe runs "<tool> --version" read-only for the detected release.
	Probe []string
	// SizeAdvisoryBytes is root_context_size_advisory_bytes (§7.9).
	SizeAdvisoryBytes int64
	// CredentialScope, AuthWrite, GlobalContextCap, ExecFlags, and
	// ProfileFlag are the §7.9 recorded members.
	CredentialScope  string
	AuthWrite        string
	GlobalContextCap string
	ExecFlags        string
	ProfileFlag      string
}

// Registry is the closed revision-1 adapter set (environments §7.1).
// #nosec G101 -- no credential material here: the strings name Keychain service schemes and file roles, never secret values.
var Registry = []Adapter{
	{
		ID:          ClaudeCode,
		EnvVar:      "CLAUDE_CONFIG_DIR",
		RootTarget:  "CLAUDE.md",
		SkillsDir:   "skills",
		Forms:       []string{FormMonolithic, FormReferenced},
		DefaultForm: FormMonolithic,
		SystemPrompt: []Channel{
			{Kind: KindFlag, Semantics: SemanticsAppend, Flag: "--append-system-prompt-file", Argument: ArgumentPath},
			{Kind: KindFlag, Semantics: SemanticsReplace, Flag: "--system-prompt-file", Argument: ArgumentPath},
		},
		MCP: &Channel{Kind: KindFlag, Flag: "--mcp-config", Argument: ArgumentPath, With: []string{"--strict-mcp-config"}},
		Passthrough: map[string][]Passthrough{
			"darwin": {},
			"linux":  {{Path: ".credentials.json", Strategy: StrategyFileLink, FileLinkTarget: ".credentials.json"}},
		},
		Seeds:             []string{".claude.json"},
		SeedWritten:       map[string]bool{".claude.json": true},
		VerifiedRelease:   "2.1.261",
		Probe:             []string{"claude", "--version"},
		SizeAdvisoryBytes: 32768,
		CredentialScope:   "per CLAUDE_CONFIG_DIR (keychain service suffix sha256[0:8]) on macOS; home on Linux",
		GlobalContextCap:  "none recorded",
	},
	{
		ID:          CodexCLI,
		EnvVar:      "CODEX_HOME",
		RootTarget:  "AGENTS.md",
		SkillsDir:   "skills",
		Forms:       []string{FormMonolithic},
		DefaultForm: FormMonolithic,
		SystemPrompt: []Channel{
			{Kind: KindConfigKey, Semantics: SemanticsReplace, Key: "model_instructions_file"},
		},
		MCP: &Channel{Kind: KindFlag, Flag: "-p", Argument: ArgumentName, Name: "curator-mcp"},
		Passthrough: map[string][]Passthrough{
			"default": {{Path: "auth.json", Strategy: StrategyKeyringPreferred, FileLinkTarget: "auth.json"}},
		},
		Seeds:             []string{"config.toml"},
		SeedWritten:       map[string]bool{},
		VerifiedRelease:   "0.153.2",
		Probe:             []string{"codex", "--version"},
		SizeAdvisoryBytes: 32768,
		CredentialScope:   "home",
		AuthWrite:         "in-place",
		GlobalContextCap:  "none",
		ExecFlags:         "--skip-git-repo-check required outside git",
		ProfileFlag:       "-p (single, silent-if-missing)",
	},
	{
		ID:                OpenCode,
		EnvVar:            "XDG_CONFIG_HOME",
		ParentVar:         true,
		RootTarget:        "AGENTS.md",
		SkillsDir:         "",
		Forms:             []string{FormMonolithic, FormReferenced},
		DefaultForm:       FormMonolithic,
		SystemPrompt:      nil,
		MCP:               &Channel{Kind: KindVariable, Variable: "OPENCODE_CONFIG"},
		Passthrough:       map[string][]Passthrough{"default": {}},
		Seeds:             nil,
		SeedWritten:       map[string]bool{},
		VerifiedRelease:   "",
		Probe:             []string{"opencode", "--version"},
		SizeAdvisoryBytes: 32768,
		CredentialScope:   "xdg-data",
		GlobalContextCap:  "docs-confidence",
	},
	{
		ID:          Pi,
		EnvVar:      "PI_CODING_AGENT_DIR",
		RootTarget:  "AGENTS.md",
		SkillsDir:   "skills",
		Forms:       []string{FormMonolithic},
		DefaultForm: FormMonolithic,
		SystemPrompt: []Channel{
			{Kind: KindFlag, Semantics: SemanticsAppend, Flag: "--append-system-prompt", Argument: ArgumentPath},
			{Kind: KindFile, Semantics: SemanticsAppend, Filename: "APPEND_SYSTEM.md"},
			{Kind: KindFile, Semantics: SemanticsReplace, Filename: "SYSTEM.md"},
		},
		MCP: nil,
		Passthrough: map[string][]Passthrough{
			"default": {{Path: "auth.json", Strategy: StrategyFileLink, FileLinkTarget: "auth.json"}},
		},
		Seeds:             []string{"settings.json", "models.json"},
		SeedWritten:       map[string]bool{},
		Shadows:           []Shadow{{Path: "AGENTS.override.md", Surface: "root-context"}},
		VerifiedRelease:   "0.84.2",
		Probe:             []string{"pi", "--version"},
		SizeAdvisoryBytes: 32768,
		CredentialScope:   "home",
		AuthWrite:         "in-place (lockfile)",
		GlobalContextCap:  "none recorded",
	},
}

// Targets is the closed revision-1 secondary-target list (§7.6): the Xcode
// embedded coding agents. The probe home's default path requires an
// operator to record (the directory is created on the first agent launch),
// so Probe names the role, not a literal path.
var Targets = []Target{
	{
		ID:         "xcode-coding-assistant",
		Adapter:    ClaudeCode,
		Probe:      "Xcode-internal agentic home directory (IDEChatOverrideAgenticHomeDirectory when set, otherwise Xcode's default)",
		Home:       "that directory as CLAUDE_CONFIG_DIR",
		Surfaces:   []string{"root-context", "skills"},
		Ungoverned: "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability",
	},
	{
		ID:         "xcode-coding-assistant",
		Adapter:    CodexCLI,
		Probe:      "the same directory",
		Home:       "that directory as CODEX_HOME",
		Surfaces:   []string{"root-context", "skills"},
		Ungoverned: "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability",
	},
}

// ByID resolves a registered environment or reports environment_unknown
// (§7.7, §10.4). An explicit operand naming an unregistered environment is
// an error, never a warning.
func ByID(id string) (Adapter, error) {
	for _, adapter := range Registry {
		if adapter.ID == id {
			return adapter, nil
		}
	}
	return Adapter{}, fmt.Errorf("%s: unregistered environment %q", DiagUnknown, id)
}

// TargetByID resolves a declared secondary target or reports
// environment_target_unknown (§7.6, §7.7). The identifier alone is
// global; use TargetFor to resolve it for one adapter.
func TargetByID(id string) (Target, error) {
	for _, target := range Targets {
		if target.ID == id {
			return target, nil
		}
	}
	return Target{}, fmt.Errorf("%s: undeclared target %q", DiagTargetUnknown, id)
}

// TargetsFor reports the declared secondary targets of one adapter.
// Adapters without a §7.6 row (opencode, pi) report none.
func TargetsFor(adapterID string) []Target {
	var out []Target
	for _, target := range Targets {
		if target.Adapter == adapterID {
			out = append(out, target)
		}
	}
	return out
}

// TargetFor resolves a declared secondary target for one adapter or
// reports environment_target_unknown (§7.6, §7.7).
func TargetFor(adapterID, id string) (Target, error) {
	for _, target := range Targets {
		if target.Adapter == adapterID && target.ID == id {
			return target, nil
		}
	}
	return Target{}, fmt.Errorf("%s: undeclared target %q for %s", DiagTargetUnknown, id, adapterID)
}

// PassthroughFor returns the adapter's passthrough entries for the runtime
// platform.
func (a Adapter) PassthroughFor(goos string) []Passthrough {
	if entries, ok := a.Passthrough[goos]; ok {
		return entries
	}
	return a.Passthrough["default"]
}

// ResolveForm maps the configured form to the effective form: the adapter
// default when unconfigured, else the requested form when the adapter
// supports it, else environment_form_unsupported (§7.2). Profile data
// cannot select a form; the caller passes machine configuration only.
func (a Adapter) ResolveForm(configured string) (string, error) {
	if configured == "" {
		return a.DefaultForm, nil
	}
	for _, form := range a.Forms {
		if configured == form {
			return form, nil
		}
	}
	return "", fmt.Errorf("%s: the %s adapter supports no %s form", DiagFormUnsupported, a.ID, configured)
}

// ResolveIsolation maps the configured isolation mode to the effective one
// under the §7.4 matrix. configured empty means the default: shared,
// except claude_code on macOS at or above the pinned release, which is
// isolated by construction (the per-CLAUDE_CONFIG_DIR Keychain item) and
// rejects shared. atOrAbovePinned reports whether the detected tool
// release is at or above the adapter's recorded verified release; an
// unknown release is reported by the caller as
// environment_tool_version_unverified and passed here as true — the
// selection scheme is verified from the bundle's service-name builder, so
// an unproven release warns rather than blocks.
func (a Adapter) ResolveIsolation(configured string, atOrAbovePinned bool) (string, error) {
	return a.resolveIsolation(runtime.GOOS, configured, atOrAbovePinned)
}

// resolveIsolation is ResolveIsolation over an explicit platform so the
// matrix is asserted on every CI runner, not only on macOS.
func (a Adapter) resolveIsolation(goos, configured string, atOrAbovePinned bool) (string, error) {
	if a.ID == OpenCode && configured == IsolationIsolated {
		return "", fmt.Errorf("%s: isolated is a no-op for opencode: auth lives outside the swapped config home", DiagIsolatedUnsupported)
	}
	if a.ID == ClaudeCode && goos == "darwin" {
		if atOrAbovePinned {
			if configured == "" || configured == IsolationIsolated {
				return IsolationIsolated, nil
			}
			if configured == IsolationShared {
				return "", fmt.Errorf("%s: shared is unsupported for claude_code on macOS at or above %s: there is no Keychain item a manager could link", DiagSharedUnsupported, a.VerifiedRelease)
			}
		} else {
			if configured == IsolationIsolated {
				return "", fmt.Errorf("%s: isolated is unsupported for claude_code on macOS below %s", DiagIsolatedUnsupported, a.VerifiedRelease)
			}
			if configured == "" || configured == IsolationShared {
				return IsolationShared, nil
			}
		}
	}
	if configured == "" {
		return IsolationShared, nil
	}
	if configured != IsolationShared && configured != IsolationIsolated {
		return "", fmt.Errorf("isolation %q is not shared or isolated", configured)
	}
	return configured, nil
}

// DefaultIsolation is ResolveIsolation with no configured value.
func (a Adapter) DefaultIsolation(atOrAbovePinned bool) (string, error) {
	return a.ResolveIsolation("", atOrAbovePinned)
}

// CompareVersions compares dotted decimal releases: -1, 0, or +1. A
// non-numeric suffix on a component (2.1.261-beta) compares below the bare
// component; an unparseable release compares as unknown (0, false).
func CompareVersions(have, want string) (int, bool) {
	parse := func(release string) ([]int, string, bool) {
		parts := strings.Split(release, ".")
		numbers := make([]int, 0, len(parts))
		for _, part := range parts {
			digits := part
			suffix := ""
			for i := 0; i < len(part); i++ {
				if part[i] < '0' || part[i] > '9' {
					digits = part[:i]
					suffix = part[i:]
					break
				}
			}
			if digits == "" {
				return nil, "", false
			}
			number, err := strconv.Atoi(digits)
			if err != nil {
				return nil, "", false
			}
			numbers = append(numbers, number)
			if suffix != "" {
				return numbers, suffix, true
			}
		}
		return numbers, "", true
	}
	haveNumbers, haveSuffix, ok := parse(have)
	if !ok {
		return 0, false
	}
	wantNumbers, _, ok := parse(want)
	if !ok {
		return 0, false
	}
	for i := 0; i < len(haveNumbers) || i < len(wantNumbers); i++ {
		var h, w int
		if i < len(haveNumbers) {
			h = haveNumbers[i]
		}
		if i < len(wantNumbers) {
			w = wantNumbers[i]
		}
		if h != w {
			if h < w {
				return -1, true
			}
			return 1, true
		}
	}
	if haveSuffix != "" {
		return -1, true
	}
	return 0, true
}

// AtOrAbovePinned reports whether the detected release is at or above the
// adapter's recorded verified release (§7.9). An empty verified release
// (opencode, docs-confidence throughout) or an undetectable release is
// unknown: the caller warns environment_tool_version_unverified and treats
// the release as unverified, never as matching.
func (a Adapter) AtOrAbovePinned(detected string) (bool, bool) {
	if a.VerifiedRelease == "" || detected == "" || detected == "unknown" {
		return false, false
	}
	comparison, ok := CompareVersions(detected, a.VerifiedRelease)
	if !ok {
		return false, false
	}
	return comparison >= 0, true
}

// MachineConfig carries the section 12.1 environments knobs. Manager-config
// schema 2 (a later batch) persists these; until then every knob takes its
// stated default when absent, which this constructor applies. Profile bytes
// never populate this struct.
type MachineConfig struct {
	// Forms maps env-id to the configured root-context form.
	Forms map[string]string
	// Isolation maps profile to env-id to the configured isolation mode.
	Isolation map[string]map[string]string
	// InPlaceMode maps env-id to linked or copied.
	InPlaceMode map[string]string
	// SystemPromptFiles maps profile to off, append, or replace (pi only).
	SystemPromptFiles map[string]string
	// TargetParticipation maps target-id to auto, off, or enabled.
	TargetParticipation map[string]string
	// TargetConsented maps target-id to recorded one-time consent.
	TargetConsented map[string]bool
	// XDGSeedAllowlist bounds the opencode XDG seeds.
	XDGSeedAllowlist []string
	// PassableEnvNames bounds env_names; nil means unbounded.
	PassableEnvNames []string
	// ShadowAcknowledged holds env + home-relative path entries the
	// operator recorded as deliberate.
	ShadowAcknowledged []ShadowAck
	// BackupRetention is the generations kept; 0 keeps every generation.
	BackupRetention int
	backupSet       bool
}

// ShadowAck is one shadow_acknowledged entry (§12.1).
type ShadowAck struct {
	Env  string
	Path string
}

// DefaultXDGSeedAllowlist is the §7.1 default.
var DefaultXDGSeedAllowlist = []string{"git", "gh", "ssh"}

// DefaultBackupRetention is the §8.3 default of 5.
const DefaultBackupRetention = 5

// DefaultMachineConfig returns the knob set with every default applied.
func DefaultMachineConfig() MachineConfig {
	return MachineConfig{
		Forms:               map[string]string{},
		Isolation:           map[string]map[string]string{},
		InPlaceMode:         map[string]string{},
		SystemPromptFiles:   map[string]string{},
		TargetParticipation: map[string]string{},
		TargetConsented:     map[string]bool{},
		XDGSeedAllowlist:    append([]string{}, DefaultXDGSeedAllowlist...),
		PassableEnvNames:    nil,
		ShadowAcknowledged:  nil,
		BackupRetention:     DefaultBackupRetention,
		backupSet:           true,
	}
}

// Retention returns the effective backup retention.
func (c MachineConfig) Retention() int {
	if !c.backupSet {
		return DefaultBackupRetention
	}
	return c.BackupRetention
}

// EffectiveForm resolves the machine-configured form for the adapter.
func (c MachineConfig) EffectiveForm(adapter Adapter) (string, error) {
	return adapter.ResolveForm(c.Forms[adapter.ID])
}

// EffectiveIsolation resolves the machine-configured isolation mode for a
// profile × environment.
func (c MachineConfig) EffectiveIsolation(profile string, adapter Adapter, atOrAbovePinned bool) (string, error) {
	configured := ""
	if perProfile, ok := c.Isolation[profile]; ok {
		configured = perProfile[adapter.ID]
	}
	return adapter.ResolveIsolation(configured, atOrAbovePinned)
}

// ShadowAcknowledges reports whether the operator recorded the shadowing
// path as deliberate, downgrading the row to a current warning (§7.5).
func (c MachineConfig) ShadowAcknowledges(env, path string) bool {
	for _, ack := range c.ShadowAcknowledged {
		if ack.Env == env && ack.Path == path {
			return true
		}
	}
	return false
}

// TargetParticipates evaluates target participation: off never, enabled
// always (an explicit enable is consent by that act), auto exactly when
// the probe path exists (§7.6).
func (c MachineConfig) TargetParticipates(target Target, probeExists func(string) bool) bool {
	switch c.TargetParticipation[target.ID] {
	case "off":
		return false
	case "enabled":
		return true
	default:
		return probeExists(target.Probe)
	}
}

// TargetConsent reports whether the first auto write into the target's
// home may proceed: recorded consent or an explicit per-target enable.
// Without either the manager stops with
// environment_target_consent_required (§7.6).
func (c MachineConfig) TargetConsent(target Target) error {
	if c.TargetParticipation[target.ID] == "enabled" || c.TargetConsented[target.ID] {
		return nil
	}
	return fmt.Errorf("%s: the first write into %s (%s) needs one-time consent: record targets.%s.consented or enable the target explicitly", DiagTargetConsent, target.ID, target.Home, target.ID)
}
