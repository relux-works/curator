package config

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/verr"
)

// Environments is the machine environments configuration (environments
// protocol §12.1, manager profile §12). Every knob below carries the exact
// name, value grammar, and default the §12.1 table declares. Parse applies
// the defaults before use; a knob that is absent takes the table default.
type Environments struct {
	CurrentProfile       *string
	ScopedCurrent        map[string]string
	Overlays             map[string][]OverlayDeclaration
	OverlayDefaultWeight int
	OverlaysAllowed      bool
	Precedence           Precedence
	Forms                map[string]string
	SystemPromptFiles    map[string]SystemPromptFiles
	// Permissions maps profile names to the section 12.1 permission mode.
	// An absent entry is a silent profile level; the system lock replaces
	// this whole map and forces native for every profile.
	Permissions      map[string]string
	Targets          map[string]TargetConfig
	Isolation        map[string]map[string]string
	XDGSeedAllowlist []string
	// PassableEnvNames is the passable_env_names knob (§12.1, default
	// []): nil with PassableEnvNamesSet false means the knob is absent;
	// nil with PassableEnvNamesSet true is the explicit null (unbounded);
	// a non-nil slice — possibly empty — is the bounded allowlist.
	PassableEnvNames    []string
	PassableEnvNamesSet bool
	MCPPackageAllowlist []string
	// ProviderDirectories is the §11 trust-root list after the install
	// directory, in listed order; the §12.1 default is the empty list.
	ProviderDirectories []string
	// SourceSigners is the per-canonical-source signer allowlist
	// (environments §12.1). A present source with an empty list admits no
	// signer; an absent source is unconfigured.
	SourceSigners map[string][]SourceSigner
	// RequireSourceSigners requires every Git source to have a source-level
	// allowlist. The security posture selects true only when this knob is absent;
	// Set distinguishes the default from an explicit false value.
	RequireSourceSigners    bool
	RequireSourceSignersSet bool
	ShadowAcknowledged      []ShadowAcknowledgement
	SecretWaivers           []SecretMaterialWaiver
	// TransitiveSystemModules is the §3/§5.5 admission policy: drop
	// (default) skips transitive system modules at materialization with a
	// warning, error refuses them at resolution.
	TransitiveSystemModules string
	// SystemModuleWaivers admits transitive packages' system modules by
	// name; an entry naming no lock member has no effect.
	SystemModuleWaivers []SystemModuleWaiver
	BackupRetention     int
	RequireCurrent      *string
	InPlaceMode         map[string]string
}

// OverlayDeclaration is one entry of overlays.<profile> (environments §6):
// an ordinary context package named by a git source or a path source, with
// the machine-assigned weight. Weight nil means the declaration carries no
// weight and overlay_default_weight applies at composition time.
type OverlayDeclaration struct {
	Source    string
	Range     string
	Tag       string
	Revision  string
	Directory string
	Weight    *int
}

// Precedence is the section 6 pair of independent primitives.
type Precedence struct {
	Winner    string
	Placement string
}

// SystemPromptFiles is the per-profile pi file selection (environments
// §5.5): exactly off, append, or replace.
type SystemPromptFiles struct {
	Pi string
}

// TargetConfig is one targets.<target-id> entry (environments §7.6).
type TargetConfig struct {
	Participation string
	Consented     bool
}

// ShadowAcknowledgement records one operator deliberate-override entry
// (environments §7.5, §12).
type ShadowAcknowledgement struct {
	Env  string
	Path string
}

// SecretMaterialWaiver is one secret_material_waivers entry (environments
// §9.1): the member pin, file protocol path, exact byte span, and reason.
type SecretMaterialWaiver struct {
	Pin         string
	HashVersion int
	File        string
	Span        [2]int
	Reason      string
}

// SystemModuleWaiver is one system_module_waivers entry (environments
// §12.1): the admitted context member's package name and the free-text
// reason recording why the operator admits its system modules.
type SystemModuleWaiver struct {
	Package string
	Reason  string
}

// SourceSigner is one SSH public key or OpenPGP fingerprint in a source
// allowlist (environments §12.1).
type SourceSigner struct {
	Type        string
	Key         string
	Fingerprint string
}

// Defaults from the environments §12.1 table.
const (
	DefaultOverlayWeight   = 1000
	DefaultBackupRetention = 5
	maxSafeInteger         = 9007199254740991
)

// LockableEnvKeys are the environments §12.2 keys a system file may lock,
// named as the manager §1 locked set names them.
var LockableEnvKeys = map[string]bool{
	"environments.overlays_allowed":          true,
	"environments.precedence":                true,
	"environments.mcp_package_allowlist":     true,
	"environments.passable_env_names":        true,
	"environments.require_current_profile":   true,
	"environments.isolation":                 true,
	"environments.transitive_system_modules": true,
	"environments.provider_directories":      true,
	"environments.permissions":               true,
	"environments.source_signers":            true,
	"environments.require_source_signers":    true,
}

// systemEnvKnobs are the environments keys a system file may carry at all:
// exactly the lockable subset (manager §1 rule 1).
func systemEnvKnobs() map[string]bool {
	knobs := map[string]bool{}
	for key := range LockableEnvKeys {
		knobs[strings.TrimPrefix(key, "environments.")] = true
	}
	return knobs
}

var (
	envRangeRE      = regexp.MustCompile(`^[0-9A-Za-z.*<>=^~| \-]+$`)
	envCommitRE     = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
	providerDriveRE = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
	maxProviderDir  = 4096
)

// defaultEnvironments returns the §12.1 defaults.
func defaultEnvironments() Environments {
	return Environments{
		ScopedCurrent:           map[string]string{},
		Overlays:                map[string][]OverlayDeclaration{},
		OverlayDefaultWeight:    DefaultOverlayWeight,
		OverlaysAllowed:         true,
		Precedence:              Precedence{Winner: "higher-weight", Placement: "winner-last"},
		Forms:                   map[string]string{},
		SystemPromptFiles:       map[string]SystemPromptFiles{},
		Permissions:             map[string]string{},
		Targets:                 map[string]TargetConfig{},
		Isolation:               map[string]map[string]string{},
		XDGSeedAllowlist:        []string{"git", "gh", "ssh"},
		MCPPackageAllowlist:     []string{},
		ProviderDirectories:     []string{},
		SourceSigners:           map[string][]SourceSigner{},
		ShadowAcknowledged:      []ShadowAcknowledgement{},
		SecretWaivers:           []SecretMaterialWaiver{},
		TransitiveSystemModules: "drop",
		SystemModuleWaivers:     []SystemModuleWaiver{},
		BackupRetention:         DefaultBackupRetention,
		InPlaceMode:             map[string]string{},
	}
}

// parseEnvironments validates a frozen manager-config schema-2 environments
// object and fills the table defaults.
func parseEnvironments(raw any) (Environments, error) {
	return parseEnvironmentsVersion(raw, SchemaVersion2)
}

// parseEnvironmentsVersion validates the environments object using its
// enclosing manager-config schema so frozen waiver shapes stay frozen.
func parseEnvironmentsVersion(raw any, schema int) (Environments, error) {
	env := defaultEnvironments()
	if raw == nil {
		return env, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return Environments{}, verr.New("environments", "must be an object")
	}
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if !envKnob(key) {
			return Environments{}, verr.New("environments", "has unsupported field %q", key)
		}
	}
	if err := parseProfileNameOrNull(obj, "current_profile", &env.CurrentProfile); err != nil {
		return Environments{}, err
	}
	if err := parseIdentifierMap(obj, "scoped_current", env.ScopedCurrent); err != nil {
		return Environments{}, err
	}
	if rawOverlays, present := obj["overlays"]; present && rawOverlays != nil {
		overlays, err := parseOverlays(rawOverlays)
		if err != nil {
			return Environments{}, err
		}
		env.Overlays = overlays
	}
	if rawWeight, present := obj["overlay_default_weight"]; present {
		weight, err := boundedInteger(rawWeight, 0, maxSafeInteger)
		if err != nil {
			return Environments{}, verr.New("environments.overlay_default_weight", "must be an integer between 0 and %d", maxSafeInteger)
		}
		env.OverlayDefaultWeight = weight
	}
	if rawAllowed, present := obj["overlays_allowed"]; present {
		allowed, ok := rawAllowed.(bool)
		if !ok {
			return Environments{}, verr.New("environments.overlays_allowed", "must be a boolean")
		}
		env.OverlaysAllowed = allowed
	}
	if rawPrecedence, present := obj["precedence"]; present && rawPrecedence != nil {
		precedence, err := parsePrecedence(rawPrecedence)
		if err != nil {
			return Environments{}, err
		}
		env.Precedence = precedence
	}
	if err := parseEnumMap(obj, "forms", map[string]bool{"monolithic": true, "referenced": true}, env.Forms); err != nil {
		return Environments{}, err
	}
	if rawSPF, present := obj["system_prompt_files"]; present && rawSPF != nil {
		spf, err := parseSystemPromptFiles(rawSPF)
		if err != nil {
			return Environments{}, err
		}
		env.SystemPromptFiles = spf
	}
	if rawPermissions, present := obj["permissions"]; present {
		if rawPermissions == nil {
			return Environments{}, verr.New("environments.permissions", "must be an object")
		}
		if err := parseEnumMap(obj, "permissions", map[string]bool{"native": true, "yolo": true}, env.Permissions); err != nil {
			return Environments{}, err
		}
	}
	if rawTargets, present := obj["targets"]; present && rawTargets != nil {
		targets, err := parseTargets(rawTargets)
		if err != nil {
			return Environments{}, err
		}
		env.Targets = targets
	}
	if rawIsolation, present := obj["isolation"]; present && rawIsolation != nil {
		isolation, err := parseIsolation(rawIsolation)
		if err != nil {
			return Environments{}, err
		}
		env.Isolation = isolation
	}
	if rawXDG, present := obj["xdg_seed_allowlist"]; present && rawXDG != nil {
		allowlist, err := parseXDGSeedAllowlist(rawXDG)
		if err != nil {
			return Environments{}, err
		}
		env.XDGSeedAllowlist = allowlist
	}
	if rawPassable, present := obj["passable_env_names"]; present {
		env.PassableEnvNamesSet = true
		if rawPassable == nil {
			env.PassableEnvNames = nil
		} else {
			names, err := identifierSet(rawPassable, "environments.passable_env_names")
			if err != nil {
				return Environments{}, err
			}
			env.PassableEnvNames = names
		}
	}
	if rawMCP, present := obj["mcp_package_allowlist"]; present && rawMCP != nil {
		allowlist, err := parseNonEmptyUniqueStrings(rawMCP, "environments.mcp_package_allowlist")
		if err != nil {
			return Environments{}, err
		}
		env.MCPPackageAllowlist = allowlist
	}
	if rawShadow, present := obj["shadow_acknowledged"]; present && rawShadow != nil {
		acks, err := parseShadowAcknowledged(rawShadow)
		if err != nil {
			return Environments{}, err
		}
		env.ShadowAcknowledged = acks
	}
	if rawWaivers, present := obj["secret_material_waivers"]; present && rawWaivers != nil {
		waivers, err := parseSecretWaivers(rawWaivers, schema)
		if err != nil {
			return Environments{}, err
		}
		env.SecretWaivers = waivers
	}
	if rawTransitive, present := obj["transitive_system_modules"]; present {
		transitive, ok := rawTransitive.(string)
		if !ok || (transitive != "drop" && transitive != "error") {
			return Environments{}, verr.New("environments.transitive_system_modules", "must be drop or error")
		}
		env.TransitiveSystemModules = transitive
	}
	// An explicit null is rejected like any other mistyped list: the
	// schema type is array (§12.1), so null is malformed, never the
	// empty default. Only absence takes the default.
	if rawWaivers, present := obj["system_module_waivers"]; present {
		waivers, err := parseSystemModuleWaivers(rawWaivers)
		if err != nil {
			return Environments{}, err
		}
		env.SystemModuleWaivers = waivers
	}
	if rawRetention, present := obj["backup_retention"]; present {
		retention, err := boundedInteger(rawRetention, 0, maxSafeInteger)
		if err != nil {
			return Environments{}, verr.New("environments.backup_retention", "must be an integer between 0 and %d", maxSafeInteger)
		}
		env.BackupRetention = retention
	}
	if err := parseProfileNameOrNull(obj, "require_current_profile", &env.RequireCurrent); err != nil {
		return Environments{}, err
	}
	if err := parseEnumMap(obj, "in_place_mode", map[string]bool{"linked": true, "copied": true}, env.InPlaceMode); err != nil {
		return Environments{}, err
	}
	// An explicit null is rejected like any other mistyped list: the
	// schema type is array (§12.1), so null is malformed, never the
	// empty default. Only absence takes the default. Unlike
	// passable_env_names, this knob has no null meaning.
	if rawPD, present := obj["provider_directories"]; present {
		dirs, err := parseProviderDirectories(rawPD)
		if err != nil {
			return Environments{}, err
		}
		env.ProviderDirectories = dirs
	}
	if rawSigners, present := obj["source_signers"]; present {
		signers, err := parseSourceSigners(rawSigners)
		if err != nil {
			return Environments{}, err
		}
		env.SourceSigners = signers
	}
	if rawRequired, present := obj["require_source_signers"]; present {
		required, ok := rawRequired.(bool)
		if !ok {
			return Environments{}, verr.New("environments.require_source_signers", "must be a boolean")
		}
		env.RequireSourceSigners = required
		env.RequireSourceSignersSet = true
	}
	return env, nil
}

// EnvKnobNames is the closed §12.1 knob list: the single source the reader
// (envKnob, SplitEnvKnob) and the system-file lockable-subset test derive
// from, so a knob added to the reader without test coverage — or to the
// test without the reader — fails instead of widening a gate silently.
var EnvKnobNames = []string{
	"current_profile", "scoped_current", "overlays", "overlay_default_weight",
	"overlays_allowed", "precedence", "forms", "system_prompt_files",
	"targets", "isolation", "xdg_seed_allowlist", "passable_env_names",
	"mcp_package_allowlist", "shadow_acknowledged", "secret_material_waivers",
	"transitive_system_modules", "system_module_waivers",
	"backup_retention", "require_current_profile", "in_place_mode",
	"provider_directories", "permissions", "source_signers", "require_source_signers",
}

var gpgFingerprintRE = regexp.MustCompile(`^[A-F0-9]{40}$`)

var sshKeyTypes = map[string]bool{
	"ssh-ed25519":                        true,
	"ssh-rsa":                            true,
	"ecdsa-sha2-nistp256":                true,
	"ecdsa-sha2-nistp384":                true,
	"ecdsa-sha2-nistp521":                true,
	"sk-ssh-ed25519@openssh.com":         true,
	"sk-ecdsa-sha2-nistp256@openssh.com": true,
}

func parseSourceSigners(raw any) (map[string][]SourceSigner, error) {
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.source_signers", "must be an object")
	}
	out := make(map[string][]SourceSigner, len(entries))
	sources := make([]string, 0, len(entries))
	for source := range entries {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		if !identity.ValidCanonical(source) {
			return nil, verr.New("environments.source_signers", "source %q must be a canonical source identity", source)
		}
		items, ok := entries[source].([]any)
		if !ok {
			return nil, verr.New("environments.source_signers."+source, "must be an array")
		}
		seen := map[string]bool{}
		parsed := make([]SourceSigner, 0, len(items))
		for index, item := range items {
			object, ok := item.(map[string]any)
			if !ok {
				return nil, verr.New("environments.source_signers."+source, "entry %d must be an object", index)
			}
			kind, ok := object["type"].(string)
			if !ok {
				return nil, verr.New("environments.source_signers."+source, "entry %d type must be ssh or gpg", index)
			}
			signer := SourceSigner{Type: kind}
			switch kind {
			case "ssh":
				if len(object) != 2 {
					return nil, verr.New("environments.source_signers."+source, "ssh entry %d must contain only type and key", index)
				}
				key, ok := object["key"].(string)
				if !ok {
					return nil, verr.New("environments.source_signers."+source, "ssh entry %d key must be an OpenSSH public key line", index)
				}
				keyType, material, err := parseOpenSSHPublicKey(key)
				if err != nil {
					return nil, verr.New("environments.source_signers."+source, "ssh entry %d: %v", index, err)
				}
				signer.Key = strings.TrimSpace(key)
				identity := keyType + " " + normalizeOpenSSHKeyMaterial(material)
				if seen[identity] {
					return nil, verr.New("environments.source_signers."+source, "contains duplicate signer %q", identity)
				}
				seen[identity] = true
			case "gpg":
				if len(object) != 2 {
					return nil, verr.New("environments.source_signers."+source, "gpg entry %d must contain only type and fingerprint", index)
				}
				fingerprint, ok := object["fingerprint"].(string)
				if !ok || !gpgFingerprintRE.MatchString(fingerprint) {
					return nil, verr.New("environments.source_signers."+source, "gpg entry %d fingerprint must be 40 uppercase hexadecimal characters", index)
				}
				signer.Fingerprint = fingerprint
				if seen["gpg "+fingerprint] {
					return nil, verr.New("environments.source_signers."+source, "contains duplicate signer %q", fingerprint)
				}
				seen["gpg "+fingerprint] = true
			default:
				return nil, verr.New("environments.source_signers."+source, "entry %d type must be ssh or gpg", index)
			}
			parsed = append(parsed, signer)
		}
		out[source] = parsed
	}
	return out, nil
}

func parseOpenSSHPublicKey(line string) (string, string, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 || !sshKeyTypes[fields[0]] {
		return "", "", fmt.Errorf("key must use a supported OpenSSH key type and base64 material")
	}
	// Environments §12.1 defines the SSH key as an OpenSSH public-key line
	// with base64 key material. Validate that grammar with or without padding;
	// cryptographic verification later decides whether the blob is a real key.
	if _, err := base64.StdEncoding.DecodeString(fields[1]); err != nil {
		if _, rawErr := base64.RawStdEncoding.DecodeString(fields[1]); rawErr != nil {
			return "", "", fmt.Errorf("key material is not valid base64")
		}
	}
	return fields[0], fields[1], nil
}

// Environments §12.1 defines SSH identity as key type plus base64 material.
// Padding is optional in the public-key line, so remove trailing '=' on
// both identities before comparing them.
func normalizeOpenSSHKeyMaterial(material string) string {
	return strings.TrimRight(material, "=")
}

// envKnob reports whether key is a §12.1 knob name.
func envKnob(key string) bool {
	for _, name := range EnvKnobNames {
		if key == name {
			return true
		}
	}
	return false
}

func parseProfileNameOrNull(obj map[string]any, field string, target **string) error {
	raw, present := obj[field]
	if !present || raw == nil {
		*target = nil
		return nil
	}
	name, ok := raw.(string)
	if !ok || !identifiers.Valid(name) {
		return verr.New("environments."+field, "%s", identifiers.Rule)
	}
	*target = &name
	return nil
}

func parseIdentifierMap(obj map[string]any, field string, target map[string]string) error {
	raw, present := obj[field]
	if !present || raw == nil {
		return nil
	}
	entries, ok := raw.(map[string]any)
	if !ok {
		return verr.New("environments."+field, "must be an object")
	}
	for key, value := range entries {
		name, ok := value.(string)
		if !identifiers.Valid(key) || !ok || !identifiers.Valid(name) {
			return verr.New("environments."+field, "must map portable identifiers to portable identifiers")
		}
		target[key] = name
	}
	return nil
}

func parseEnumMap(obj map[string]any, field string, allowed map[string]bool, target map[string]string) error {
	raw, present := obj[field]
	if !present || raw == nil {
		return nil
	}
	entries, ok := raw.(map[string]any)
	if !ok {
		return verr.New("environments."+field, "must be an object")
	}
	var keys []string
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var allowedNames []string
	for name := range allowed {
		allowedNames = append(allowedNames, name)
	}
	sort.Strings(allowedNames)
	for _, key := range keys {
		if !identifiers.Valid(key) {
			return verr.New("environments."+field, "key %q %s", key, identifiers.Rule)
		}
		value, ok := entries[key].(string)
		if !ok || !allowed[value] {
			return verr.New("environments."+field+"."+key, "must be one of %s", strings.Join(allowedNames, ", "))
		}
		target[key] = value
	}
	return nil
}

func parseOverlays(raw any) (map[string][]OverlayDeclaration, error) {
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.overlays", "must be an object")
	}
	overlays := map[string][]OverlayDeclaration{}
	for profile, rawList := range entries {
		if !identifiers.Valid(profile) {
			return nil, verr.New("environments.overlays", "profile %q %s", profile, identifiers.Rule)
		}
		list, ok := rawList.([]any)
		if !ok {
			return nil, verr.New("environments.overlays."+profile, "must be a list")
		}
		for index, item := range list {
			label := fmt.Sprintf("environments.overlays.%s[%d]", profile, index)
			decl, err := parseOverlay(item, label)
			if err != nil {
				return nil, err
			}
			overlays[profile] = append(overlays[profile], decl)
		}
		if _, present := overlays[profile]; !present {
			overlays[profile] = []OverlayDeclaration{}
		}
	}
	return overlays, nil
}

func parseOverlay(raw any, label string) (OverlayDeclaration, error) {
	entry, ok := raw.(map[string]any)
	if !ok {
		return OverlayDeclaration{}, verr.New(label, "must be an object")
	}
	for key := range entry {
		switch key {
		case "source", "range", "tag", "revision", "directory", "weight":
		default:
			return OverlayDeclaration{}, verr.New(label, "has unsupported field %q", key)
		}
	}
	source, _ := entry["source"].(string)
	if !nonEmptyString(source) {
		return OverlayDeclaration{}, verr.New(label+".source", "requires a non-empty string")
	}
	forms := 0
	var decl OverlayDeclaration
	decl.Source = source
	if rawRange, present := entry["range"]; present {
		forms++
		rng, ok := rawRange.(string)
		if !ok || !validRange(rng) {
			return OverlayDeclaration{}, verr.New(label+".range", "must be a version range over 1-1024 characters of [0-9A-Za-z.*<>=^~| -]")
		}
		decl.Range = rng
	}
	if rawTag, present := entry["tag"]; present {
		forms++
		tag, ok := rawTag.(string)
		if !ok || !validGitRef(tag) {
			return OverlayDeclaration{}, verr.New(label+".tag", "must be a valid git ref name of 1-255 characters")
		}
		decl.Tag = tag
	}
	if rawRevision, present := entry["revision"]; present {
		forms++
		revision, ok := rawRevision.(string)
		if !ok || !envCommitRE.MatchString(revision) {
			return OverlayDeclaration{}, verr.New(label+".revision", "must be 40 or 64 lowercase hex characters")
		}
		decl.Revision = revision
	}
	if rawDir, present := entry["directory"]; present {
		dir, ok := rawDir.(string)
		if !ok || !identifiers.PortablePath(dir) {
			return OverlayDeclaration{}, verr.New(label+".directory", "must be a portable relative path")
		}
		decl.Directory = dir
	}
	// The requirement form is required only for a `git` source
	// (manager-config-v2 $defs/overlay, environments §1 and §12.1). A
	// `path` overlay carries no range, tag, revision, or directory —
	// section 1 makes any of them profile_source_invalid, so the reader
	// refuses the declaration outright rather than admitting a shape no
	// resolution can accept. A spelling that is neither kind is refused
	// here too: core §6.1 rejects an invalid network form instead of
	// taking it as local.
	switch identity.ClassifySource(source) {
	case identity.SourceGit:
		if forms != 1 {
			return OverlayDeclaration{}, verr.New(label, "a git overlay requires exactly one of range, tag, or revision")
		}
	case identity.SourcePath:
		if forms != 0 {
			return OverlayDeclaration{}, verr.New(label, "a path overlay carries no range, tag, or revision")
		}
		if decl.Directory != "" {
			return OverlayDeclaration{}, verr.New(label, "a path overlay carries no directory")
		}
	default:
		return OverlayDeclaration{}, verr.New(label+".source", "is neither a git source nor a path")
	}
	if rawWeight, present := entry["weight"]; present {
		weight, err := boundedInteger(rawWeight, 0, maxSafeInteger)
		if err != nil {
			return OverlayDeclaration{}, verr.New(label+".weight", "must be an integer between 0 and %d", maxSafeInteger)
		}
		decl.Weight = &weight
	}
	return decl, nil
}

func parsePrecedence(raw any) (Precedence, error) {
	precedence := Precedence{Winner: "higher-weight", Placement: "winner-last"}
	obj, ok := raw.(map[string]any)
	if !ok {
		return Precedence{}, verr.New("environments.precedence", "must be an object")
	}
	for key := range obj {
		if key != "winner" && key != "placement" {
			return Precedence{}, verr.New("environments.precedence", "has unsupported field %q", key)
		}
	}
	if rawWinner, present := obj["winner"]; present {
		winner, ok := rawWinner.(string)
		if !ok || (winner != "higher-weight" && winner != "lower-weight") {
			return Precedence{}, verr.New("environments.precedence.winner", "must be higher-weight or lower-weight")
		}
		precedence.Winner = winner
	}
	if rawPlacement, present := obj["placement"]; present {
		placement, ok := rawPlacement.(string)
		if !ok || (placement != "winner-last" && placement != "winner-first") {
			return Precedence{}, verr.New("environments.precedence.placement", "must be winner-last or winner-first")
		}
		precedence.Placement = placement
	}
	return precedence, nil
}

func parseSystemPromptFiles(raw any) (map[string]SystemPromptFiles, error) {
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.system_prompt_files", "must be an object")
	}
	result := map[string]SystemPromptFiles{}
	for profile, rawEntry := range entries {
		if !identifiers.Valid(profile) {
			return nil, verr.New("environments.system_prompt_files", "profile %q %s", profile, identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New("environments.system_prompt_files."+profile, "must be an object")
		}
		for key := range entry {
			if key != "pi" {
				return nil, verr.New("environments.system_prompt_files."+profile, "has unsupported field %q", key)
			}
		}
		setting := SystemPromptFiles{Pi: "off"}
		if rawPi, present := entry["pi"]; present {
			pi, ok := rawPi.(string)
			if !ok || (pi != "off" && pi != "append" && pi != "replace") {
				return nil, verr.New("environments.system_prompt_files."+profile+".pi", "must be off, append, or replace")
			}
			setting.Pi = pi
		}
		result[profile] = setting
	}
	return result, nil
}

func parseTargets(raw any) (map[string]TargetConfig, error) {
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.targets", "must be an object")
	}
	result := map[string]TargetConfig{}
	for id, rawEntry := range entries {
		if !identifiers.Valid(id) {
			return nil, verr.New("environments.targets", "target %q %s", id, identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New("environments.targets."+id, "must be an object")
		}
		for key := range entry {
			if key != "participation" && key != "consented" {
				return nil, verr.New("environments.targets."+id, "has unsupported field %q", key)
			}
		}
		target := TargetConfig{Participation: "auto"}
		if rawParticipation, present := entry["participation"]; present {
			participation, ok := rawParticipation.(string)
			if !ok || (participation != "auto" && participation != "off" && participation != "enabled") {
				return nil, verr.New("environments.targets."+id+".participation", "must be auto, off, or enabled")
			}
			target.Participation = participation
		}
		if rawConsented, present := entry["consented"]; present {
			consented, ok := rawConsented.(bool)
			if !ok {
				return nil, verr.New("environments.targets."+id+".consented", "must be a boolean")
			}
			target.Consented = consented
		}
		result[id] = target
	}
	return result, nil
}

// parseIsolation validates isolation.<profile>.<env> with values shared or
// isolated. System configuration uses the same grammar because environments
// §12.2 makes both directions lockable.
func parseIsolation(raw any) (map[string]map[string]string, error) {
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.isolation", "must be an object")
	}
	result := map[string]map[string]string{}
	for profile, rawEntry := range entries {
		if !identifiers.Valid(profile) {
			return nil, verr.New("environments.isolation", "profile %q %s", profile, identifiers.Rule)
		}
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New("environments.isolation."+profile, "must be an object")
		}
		modes := map[string]string{}
		for env, rawMode := range entry {
			if !identifiers.Valid(env) {
				return nil, verr.New("environments.isolation."+profile, "environment %q %s", env, identifiers.Rule)
			}
			mode, ok := rawMode.(string)
			if !ok || (mode != "shared" && mode != "isolated") {
				return nil, verr.New("environments.isolation."+profile+"."+env, "must be shared or isolated")
			}
			modes[env] = mode
		}
		result[profile] = modes
	}
	return result, nil
}

func parseXDGSeedAllowlist(raw any) ([]string, error) {
	names, err := stringList(raw, "environments.xdg_seed_allowlist")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, name := range names {
		if !validXDGEntryName(name) || seen[name] {
			return nil, verr.New("environments.xdg_seed_allowlist", "must contain unique XDG config entry names: 1-255 characters, no path separator, never opencode")
		}
		seen[name] = true
	}
	return names, nil
}

func validXDGEntryName(name string) bool {
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 255 {
		return false
	}
	if name == "." || name == ".." || name == "opencode" {
		return false
	}
	for _, character := range name {
		if character == '/' || character == '\\' || character == 0 || (character < 0x20 || character == 0x7f) {
			return false
		}
	}
	return true
}

func parseNonEmptyUniqueStrings(raw any, field string) ([]string, error) {
	names, err := stringList(raw, field)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, name := range names {
		if !nonEmptyString(name) || seen[name] {
			return nil, verr.New(field, "must contain unique non-empty strings")
		}
		seen[name] = true
	}
	return names, nil
}

func nonEmptyString(value string) bool {
	length := utf8.RuneCountInString(value)
	return length >= 1 && length <= 8192
}

// parseProviderDirectories validates the §11/§12.1 provider_directories
// knob with the manager-config-v2 grammar: a list of unique absolute
// paths, each POSIX-absolute or Windows drive-absolute and 1-4096
// characters long — the two spellings the schema admits.
func parseProviderDirectories(raw any) ([]string, error) {
	const field = "environments.provider_directories"
	dirs, err := stringList(raw, field)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := []string{}
	for _, dir := range dirs {
		if !validProviderDir(dir) || seen[dir] {
			return nil, verr.New(field, "must contain unique absolute paths: POSIX-absolute or Windows drive-absolute, 1-4096 characters")
		}
		seen[dir] = true
		out = append(out, dir)
	}
	return out, nil
}

func validProviderDir(dir string) bool {
	length := utf8.RuneCountInString(dir)
	if length < 1 || length > maxProviderDir {
		return false
	}
	if strings.HasPrefix(dir, "/") {
		return true
	}
	return providerDriveRE.MatchString(dir)
}

func parseShadowAcknowledged(raw any) ([]ShadowAcknowledgement, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("environments.shadow_acknowledged", "must be a list")
	}
	acks := []ShadowAcknowledgement{}
	for index, item := range list {
		label := fmt.Sprintf("environments.shadow_acknowledged[%d]", index)
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		for key := range entry {
			if key != "env" && key != "path" {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		env, _ := entry["env"].(string)
		path, _ := entry["path"].(string)
		if !identifiers.Valid(env) {
			return nil, verr.New(label+".env", "%s", identifiers.Rule)
		}
		if !nonEmptyString(path) {
			return nil, verr.New(label+".path", "requires a non-empty string")
		}
		acks = append(acks, ShadowAcknowledgement{Env: env, Path: path})
	}
	return acks, nil
}

func parseSecretWaivers(raw any, schema int) ([]SecretMaterialWaiver, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("environments.secret_material_waivers", "must be a list")
	}
	waivers := []SecretMaterialWaiver{}
	for index, item := range list {
		label := fmt.Sprintf("environments.secret_material_waivers[%d]", index)
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		for key := range entry {
			if key != "pin" && key != "file" && key != "span" && key != "reason" && (schema != SchemaVersion3 || key != "hash_version") {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		pin, _ := entry["pin"].(string)
		if !envCommitRE.MatchString(pin) {
			return nil, verr.New(label+".pin", "must be 40 or 64 lowercase hex characters")
		}
		hashVersion := 0
		if rawVersion, present := entry["hash_version"]; present {
			if schema != SchemaVersion3 {
				return nil, verr.New(label+".hash_version", "requires manager-config schema_version 3")
			}
			value, ok := integerValue(rawVersion)
			if !ok || value != 2 {
				return nil, verr.New(label+".hash_version", "must be 2")
			}
			if len(pin) != 64 {
				return nil, verr.New(label+".pin", "hash_version requires a 64-character state hash pin")
			}
			hashVersion = 2
		} else if len(pin) == 64 {
			// A state hash without the additive member retains its frozen v1
			// interpretation, including when read from schema 3.
			hashVersion = 1
		}
		file, _ := entry["file"].(string)
		if !identifiers.PortablePath(file) {
			return nil, verr.New(label+".file", "must be a portable relative path")
		}
		rawSpan, present := entry["span"]
		if !present {
			return nil, verr.New(label+".span", "requires a two-element [start, end] span")
		}
		span, ok := rawSpan.([]any)
		if !ok || len(span) != 2 {
			return nil, verr.New(label+".span", "requires a two-element [start, end] span")
		}
		var bounds [2]int
		for i, bound := range span {
			value, err := boundedInteger(bound, 0, maxSafeInteger)
			if err != nil {
				return nil, verr.New(label+".span", "span bounds must be integers between 0 and %d", maxSafeInteger)
			}
			bounds[i] = value
		}
		reason, _ := entry["reason"].(string)
		if !nonEmptyString(reason) {
			return nil, verr.New(label+".reason", "requires a non-empty string")
		}
		waivers = append(waivers, SecretMaterialWaiver{Pin: pin, HashVersion: hashVersion, File: file, Span: bounds, Reason: reason})
	}
	return waivers, nil
}

func parseSystemModuleWaivers(raw any) ([]SystemModuleWaiver, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, verr.New("environments.system_module_waivers", "must be a list")
	}
	waivers := []SystemModuleWaiver{}
	for index, item := range list {
		label := fmt.Sprintf("environments.system_module_waivers[%d]", index)
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		for key := range entry {
			if key != "package" && key != "reason" {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		pkg, _ := entry["package"].(string)
		if !identifiers.Valid(pkg) {
			return nil, verr.New(label+".package", "%s", identifiers.Rule)
		}
		reason, _ := entry["reason"].(string)
		if !nonEmptyString(reason) {
			return nil, verr.New(label+".reason", "requires a non-empty string")
		}
		waivers = append(waivers, SystemModuleWaiver{Package: pkg, Reason: reason})
	}
	return waivers, nil
}

// validRange is the closed range-grammar character check of
// manager-config-v2 (length 1-1024 over [0-9A-Za-z.*<>=^~| -]). Full range
// semantics belong to the resolver (pkgversion); the config layer pins the
// spelling so an unparseable range fails profile_source_invalid at the
// declaration, never silently.
func validRange(value string) bool {
	length := utf8.RuneCountInString(value)
	return length >= 1 && length <= 1024 && envRangeRE.MatchString(value)
}

// validGitRef is the core §6.3 tag grammar the schema encodes: 1-255
// characters, no leading slash, no doubled slash, no dot-dot, no @{, no
// leading-dot component, no .lock component suffix, no ASCII control,
// space, ~ ^ : ? * [ backslash, never @ alone, never ending in slash or
// dot. Go's RE2 has no lookahead, so the check is explicit.
func validGitRef(value string) bool {
	length := utf8.RuneCountInString(value)
	if length < 1 || length > 255 || value == "@" {
		return false
	}
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.HasSuffix(value, ".") {
		return false
	}
	if strings.Contains(value, "//") || strings.Contains(value, "..") || strings.Contains(value, "@{") {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || strings.HasPrefix(component, ".") {
			return false
		}
		trimmed := component
		if strings.HasSuffix(trimmed, ".lock") {
			return false
		}
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f || strings.ContainsRune("~^:?*[\\", character) {
			return false
		}
	}
	return true
}

// EffectiveOverlays returns the overlay declarations for profile after the
// §12.2 composition policy: a locked or configured overlays_allowed false
// empties every overlay list, so resolution joins the root alone.
func (e Environments) EffectiveOverlays(profile string) []OverlayDeclaration {
	if !e.OverlaysAllowed {
		return nil
	}
	return e.Overlays[profile]
}

// LockedBySystem reports whether the manager §1 locked set names key.
func (c *Config) LockedBySystem(key string) bool {
	return c != nil && c.Locked[key]
}

// parseSystemEnvironments validates the system file's environments object:
// exactly the §12.2 lockable subset with the §12.1 value grammars,
// and permissions only toward native.
// Anything else is not carriable by the system file (manager §1 rule 1).
func parseSystemEnvironments(raw any) (map[string]any, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments", "must be an object")
	}
	allowed := systemEnvKnobs()
	for key := range obj {
		if !allowed[key] {
			return nil, verr.New("environments."+key, "is not lockable and must not be carried by the system file")
		}
	}
	if _, err := parseEnvironments(raw); err != nil {
		return nil, err
	}
	if rawIsolation, present := obj["isolation"]; present && rawIsolation != nil {
		if _, err := parseIsolation(rawIsolation); err != nil {
			return nil, err
		}
	}
	// Environments §12.2 locks transitive_system_modules only toward
	// error: a system file MUST NOT admit a transitive package's system
	// modules by locking (or defaulting) the knob to drop.
	if rawTransitive, present := obj["transitive_system_modules"]; present && rawTransitive != nil {
		if transitive, _ := rawTransitive.(string); transitive == "drop" {
			return nil, verr.New("environments.transitive_system_modules", "a system file locks transitive_system_modules only toward error")
		}
	}
	if _, present := obj["permissions"]; present {
		if obj["permissions"] == nil {
			return nil, verr.New("environments.permissions", "must be an object")
		}
		if err := parseEnumMap(obj, "permissions", map[string]bool{"native": true}, map[string]string{}); err != nil {
			return nil, err
		}
	}
	return obj, nil
}

// EffectiveJSON renders the effective configuration the
// manager-config-v2 vector family expects: adapter_mode, default_agents,
// and the environments object with every §12.1 default applied.
func (c *Config) EffectiveJSON() map[string]any {
	env := map[string]any{}
	if c != nil {
		env = c.Env.render()
	}
	agents := DefaultAgents
	mode := "auto"
	if c != nil {
		agents = append([]string(nil), c.DefaultAgents...)
		mode = c.AdapterMode
	}
	return map[string]any{
		"adapter_mode":   mode,
		"default_agents": agents,
		"environments":   env,
	}
}

func (e Environments) render() map[string]any {
	overlays := map[string]any{}
	for profile, list := range e.Overlays {
		rendered := []any{}
		for _, decl := range list {
			entry := map[string]any{"source": decl.Source}
			// A path overlay carries no requirement form, so the
			// effective rendering carries none either: the §12.1 row
			// is {source, range|tag|revision, directory?, weight?}
			// for a git source and {source, weight?} for a path one.
			switch {
			case decl.Range != "":
				entry["range"] = decl.Range
			case decl.Tag != "":
				entry["tag"] = decl.Tag
			case decl.Revision != "":
				entry["revision"] = decl.Revision
			}
			if decl.Directory != "" {
				entry["directory"] = decl.Directory
			}
			if decl.Weight != nil {
				entry["weight"] = *decl.Weight
			}
			rendered = append(rendered, entry)
		}
		overlays[profile] = rendered
	}
	forms := map[string]any{}
	for key, value := range e.Forms {
		forms[key] = value
	}
	spf := map[string]any{}
	for profile, setting := range e.SystemPromptFiles {
		spf[profile] = map[string]any{"pi": setting.Pi}
	}
	targets := map[string]any{}
	for id, target := range e.Targets {
		targets[id] = map[string]any{"participation": target.Participation, "consented": target.Consented}
	}
	isolation := map[string]any{}
	for profile, modes := range e.Isolation {
		entry := map[string]any{}
		for env, mode := range modes {
			entry[env] = mode
		}
		isolation[profile] = entry
	}
	scoped := map[string]any{}
	for key, value := range e.ScopedCurrent {
		scoped[key] = value
	}
	inPlace := map[string]any{}
	for key, value := range e.InPlaceMode {
		inPlace[key] = value
	}
	shadow := []any{}
	for _, ack := range e.ShadowAcknowledged {
		shadow = append(shadow, map[string]any{"env": ack.Env, "path": ack.Path})
	}
	waivers := []any{}
	for _, waiver := range e.SecretWaivers {
		entry := map[string]any{
			"pin": waiver.Pin, "file": waiver.File,
			"span": []any{waiver.Span[0], waiver.Span[1]}, "reason": waiver.Reason,
		}
		if waiver.HashVersion == 2 {
			entry["hash_version"] = 2
		}
		waivers = append(waivers, entry)
	}
	systemWaivers := []any{}
	for _, waiver := range e.SystemModuleWaivers {
		systemWaivers = append(systemWaivers, map[string]any{
			"package": waiver.Package, "reason": waiver.Reason,
		})
	}
	xdg := []any{}
	for _, name := range e.XDGSeedAllowlist {
		xdg = append(xdg, name)
	}
	mcp := []any{}
	for _, name := range e.MCPPackageAllowlist {
		mcp = append(mcp, name)
	}
	providerDirs := []any{}
	for _, dir := range e.ProviderDirectories {
		providerDirs = append(providerDirs, dir)
	}
	sourceSigners := map[string]any{}
	for source, signers := range e.SourceSigners {
		entries := make([]any, 0, len(signers))
		for _, signer := range signers {
			entry := map[string]any{"type": signer.Type}
			if signer.Type == "ssh" {
				entry["key"] = signer.Key
			} else {
				entry["fingerprint"] = signer.Fingerprint
			}
			entries = append(entries, entry)
		}
		sourceSigners[source] = entries
	}
	// The §12.1 schema default is []: an absent knob renders the empty
	// list, an explicit null renders null (unbounded), a configured
	// list renders itself. The S4 runtime default for an absent knob —
	// unbounded with a warning under s4-warn — lives in the resolver,
	// not in this rendering.
	var passable any = []any{}
	if e.PassableEnvNamesSet {
		if e.PassableEnvNames != nil {
			list := []any{}
			for _, name := range e.PassableEnvNames {
				list = append(list, name)
			}
			passable = list
		} else {
			passable = nil
		}
	}
	var current any
	if e.CurrentProfile != nil {
		current = *e.CurrentProfile
	}
	var require any
	if e.RequireCurrent != nil {
		require = *e.RequireCurrent
	}
	permissions := map[string]any{}
	for profile, mode := range e.Permissions {
		permissions[profile] = mode
	}
	return map[string]any{
		"current_profile": current, "scoped_current": scoped, "overlays": overlays,
		"overlay_default_weight": e.OverlayDefaultWeight, "overlays_allowed": e.OverlaysAllowed,
		"precedence": map[string]any{"winner": e.Precedence.Winner, "placement": e.Precedence.Placement},
		"forms":      forms, "system_prompt_files": spf, "targets": targets,
		"isolation": isolation, "xdg_seed_allowlist": xdg, "passable_env_names": passable,
		"mcp_package_allowlist": mcp, "shadow_acknowledged": shadow,
		"secret_material_waivers":   waivers,
		"transitive_system_modules": e.TransitiveSystemModules, "system_module_waivers": systemWaivers,
		"backup_retention":        e.BackupRetention,
		"require_current_profile": require, "in_place_mode": inPlace,
		"provider_directories": providerDirs, "permissions": permissions,
		"source_signers":         sourceSigners,
		"require_source_signers": e.RequireSourceSigners,
	}
}

// EnvLockKey maps an env config knob path to the manager §1 locked key that
// guards it, or "" when the knob is not lockable. Dotted paths nest under
// their table knob: precedence.* shares the precedence lock, isolation.*.*
// shares the isolation lock.
func EnvLockKey(knob string) string {
	head := knob
	if index := strings.Index(head, "."); index >= 0 {
		head = head[:index]
	}
	switch head {
	case "precedence":
		return "environments.precedence"
	case "isolation":
		return "environments.isolation"
	case "overlays_allowed", "mcp_package_allowlist", "passable_env_names", "require_current_profile", "transitive_system_modules", "provider_directories", "permissions", "source_signers", "require_source_signers":
		return "environments." + head
	}
	return ""
}

// SplitEnvKnob validates the knob spelling against the §12.1 table and
// splits it into environments-relative path segments.
func SplitEnvKnob(knob string) ([]string, error) {
	segments := strings.Split(knob, ".")
	if len(segments) == 0 || segments[0] == "" {
		return nil, verr.New("knob", "requires a section 12.1 knob name")
	}
	if envKnob(segments[0]) {
		for _, segment := range segments[1:] {
			if segment == "" {
				return nil, verr.New("knob", "requires a section 12.1 knob name")
			}
		}
		return segments, nil
	}
	return nil, verr.New("knob", "unknown environments knob %q", knob)
}

// ReadRaw loads the user config file without merging the system file, for
// read-modify-write commands that edit the machine file only.
func ReadRaw(path string) (map[string]any, error) {
	return readObject(path)
}

// WriteRaw validates the object and stores it atomically.
func WriteRaw(path string, object map[string]any) (*Config, error) {
	cfg, err := Parse(object, path)
	if err != nil {
		return nil, err
	}
	return cfg, writeObjectAtomic(path, object)
}
