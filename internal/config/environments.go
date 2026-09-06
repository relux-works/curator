package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
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
	Targets              map[string]TargetConfig
	Isolation            map[string]map[string]string
	XDGSeedAllowlist     []string
	// PassableEnvNames is nil for the null (unbounded) default; a non-nil
	// slice — possibly empty — is the bounded allowlist.
	PassableEnvNames    []string
	MCPPackageAllowlist []string
	ShadowAcknowledged  []ShadowAcknowledgement
	SecretWaivers       []SecretMaterialWaiver
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
	Pin    string
	File   string
	Span   [2]int
	Reason string
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
	"environments.overlays_allowed":        true,
	"environments.precedence":              true,
	"environments.mcp_package_allowlist":   true,
	"environments.passable_env_names":      true,
	"environments.require_current_profile": true,
	"environments.isolation":               true,
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
	envRangeRE  = regexp.MustCompile(`^[0-9A-Za-z.*<>=^~| \-]+$`)
	envCommitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
)

// defaultEnvironments returns the §12.1 defaults.
func defaultEnvironments() Environments {
	return Environments{
		ScopedCurrent:        map[string]string{},
		Overlays:             map[string][]OverlayDeclaration{},
		OverlayDefaultWeight: DefaultOverlayWeight,
		OverlaysAllowed:      true,
		Precedence:           Precedence{Winner: "higher-weight", Placement: "winner-last"},
		Forms:                map[string]string{},
		SystemPromptFiles:    map[string]SystemPromptFiles{},
		Targets:              map[string]TargetConfig{},
		Isolation:            map[string]map[string]string{},
		XDGSeedAllowlist:     []string{"git", "gh", "ssh"},
		MCPPackageAllowlist:  []string{},
		ShadowAcknowledged:   []ShadowAcknowledgement{},
		SecretWaivers:        []SecretMaterialWaiver{},
		BackupRetention:      DefaultBackupRetention,
		InPlaceMode:          map[string]string{},
	}
}

// parseEnvironments validates the manager-config schema-2 environments
// object with the §12.1 value grammars and fills the table defaults.
func parseEnvironments(raw any) (Environments, error) {
	env := defaultEnvironments()
	if raw == nil {
		return env, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return Environments{}, verr.New("environments", "must be an object")
	}
	for key := range obj {
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
	if rawTargets, present := obj["targets"]; present && rawTargets != nil {
		targets, err := parseTargets(rawTargets)
		if err != nil {
			return Environments{}, err
		}
		env.Targets = targets
	}
	if rawIsolation, present := obj["isolation"]; present && rawIsolation != nil {
		isolation, err := parseIsolation(rawIsolation, false)
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
		waivers, err := parseSecretWaivers(rawWaivers)
		if err != nil {
			return Environments{}, err
		}
		env.SecretWaivers = waivers
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
	"backup_retention", "require_current_profile", "in_place_mode",
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
	// Every overlay carries exactly one requirement form — git or path
	// alike (environments §12.1). A form on a `path` source is reader
	// grammar but resolution-invalid: installing or updating a profile
	// whose overlay declaration puts range, tag, revision, or directory
	// on a path source fails with profile_source_invalid (section 1).
	if forms != 1 {
		return OverlayDeclaration{}, verr.New(label, "requires exactly one of range, tag, or revision")
	}
	if rawDir, present := entry["directory"]; present {
		dir, ok := rawDir.(string)
		if !ok || !identifiers.PortablePath(dir) {
			return OverlayDeclaration{}, verr.New(label+".directory", "must be a portable relative path")
		}
		decl.Directory = dir
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
// isolated; systemOnly restricts values to shared, the one direction
// environments §12.2 makes lockable.
func parseIsolation(raw any, systemOnly bool) (map[string]map[string]string, error) {
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
			if !ok || (mode != "shared" && mode != "isolated") || (systemOnly && mode != "shared") {
				if systemOnly {
					return nil, verr.New("environments.isolation."+profile+"."+env, "a system file locks isolation only toward shared")
				}
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

func parseSecretWaivers(raw any) ([]SecretMaterialWaiver, error) {
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
			if key != "pin" && key != "file" && key != "span" && key != "reason" {
				return nil, verr.New(label, "has unsupported field %q", key)
			}
		}
		pin, _ := entry["pin"].(string)
		if !envCommitRE.MatchString(pin) {
			return nil, verr.New(label+".pin", "must be 40 or 64 lowercase hex characters")
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
		waivers = append(waivers, SecretMaterialWaiver{Pin: pin, File: file, Span: bounds, Reason: reason})
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
// exactly the §12.2 lockable subset with the §12.1 value grammars, and
// isolation only toward shared. Anything else is not carriable by the
// system file (manager §1 rule 1).
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
		if _, err := parseIsolation(rawIsolation, true); err != nil {
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
			switch {
			case decl.Range != "":
				entry["range"] = decl.Range
			case decl.Tag != "":
				entry["tag"] = decl.Tag
			default:
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
		waivers = append(waivers, map[string]any{
			"pin": waiver.Pin, "file": waiver.File,
			"span": []any{waiver.Span[0], waiver.Span[1]}, "reason": waiver.Reason,
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
	var passable any
	if e.PassableEnvNames != nil {
		list := []any{}
		for _, name := range e.PassableEnvNames {
			list = append(list, name)
		}
		passable = list
	}
	var current any
	if e.CurrentProfile != nil {
		current = *e.CurrentProfile
	}
	var require any
	if e.RequireCurrent != nil {
		require = *e.RequireCurrent
	}
	return map[string]any{
		"current_profile": current, "scoped_current": scoped, "overlays": overlays,
		"overlay_default_weight": e.OverlayDefaultWeight, "overlays_allowed": e.OverlaysAllowed,
		"precedence": map[string]any{"winner": e.Precedence.Winner, "placement": e.Precedence.Placement},
		"forms":      forms, "system_prompt_files": spf, "targets": targets,
		"isolation": isolation, "xdg_seed_allowlist": xdg, "passable_env_names": passable,
		"mcp_package_allowlist": mcp, "shadow_acknowledged": shadow,
		"secret_material_waivers": waivers, "backup_retention": e.BackupRetention,
		"require_current_profile": require, "in_place_mode": inPlace,
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
	case "overlays_allowed", "mcp_package_allowlist", "passable_env_names", "require_current_profile":
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
