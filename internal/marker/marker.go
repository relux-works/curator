// Package marker reads and writes install markers (.csk-install.json) and
// implements the up-to-date and tamper-detection semantics of Spec §8.5.
package marker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Name is the marker file name inside an installed skill directory.
const Name = ".csk-install.json"

const (
	// LegacySchemaVersion is the historical marker schema retained for reads.
	LegacySchemaVersion = 1
	// SchemaVersion is the marker schema written by every installation mutation.
	SchemaVersion = 2
	// ExternalSchemaVersion is written for schema-7 installations. It can
	// represent local go-v1 and external go-repository-v1 commands together.
	ExternalSchemaVersion = 3
	// PolicySchemaVersion is written for schema-8 installations. It is
	// ExternalSchemaVersion with `schema_version` 4 and `skill_schema_version`
	// 8 and no other difference, so every marker-v3 build-record rule applies
	// to it unchanged.
	PolicySchemaVersion = 4
)

var (
	markerCommitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
	markerSHA256RE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	markerKeyIDRE  = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

// Activation records how the node was activated.
type Activation struct {
	Context  bool     `json:"context"`
	Commands []string `json:"commands"`
}

// Attestation is the authorizing registry record summary (Spec §13.3).
type Attestation struct {
	Registry string `json:"registry"`
	Status   string `json:"status"`
	KeyID    string `json:"key_id,omitempty"`
}

// Build records the portable identities needed to revalidate one compiled
// command without persisting any manager-home path.
type Build struct {
	Driver               string                `json:"driver"`
	ReceiptSchemaVersion int                   `json:"receipt_schema_version,omitempty"`
	ExecutionPolicy      string                `json:"execution_policy,omitempty"`
	Repository           string                `json:"repository,omitempty"`
	DeclaredIdentity     *RepositoryIdentity   `json:"declared_identity,omitempty"`
	DeclaredLockedCommit *RepositoryCommit     `json:"declared_locked_commit,omitempty"`
	DeclaredTag          string                `json:"declared_tag,omitempty"`
	EffectiveIdentity    *RepositoryIdentity   `json:"effective_identity,omitempty"`
	ObjectFormat         string                `json:"object_format,omitempty"`
	Commit               string                `json:"commit,omitempty"`
	Substituted          bool                  `json:"substituted,omitempty"`
	Substitution         *RepositorySubstitute `json:"substitution,omitempty"`
	BuildSource          *buildsource.Identity `json:"build_source,omitempty"`
	DescriptorTarget     string                `json:"descriptor_target,omitempty"`
	CacheKey             buildmeta.CacheKey    `json:"cache_key"`
	ReceiptSHA256        buildmeta.ReceiptHash `json:"receipt_sha256"`
	ArtifactSHA256       string                `json:"artifact_sha256"`
	ArtifactPath         string                `json:"artifact_path"`
}

// MarshalJSON keeps marker-v2 local records byte-compatible while emitting
// the closed marker-v3 record selected by the driver. In particular,
// substituted=false is required for external records and forbidden for local
// ones, which cannot be expressed with a single struct tag.
func (b Build) MarshalJSON() ([]byte, error) {
	value := map[string]any{
		"driver": b.Driver, "cache_key": b.CacheKey, "receipt_sha256": b.ReceiptSHA256,
		"artifact_sha256": b.ArtifactSHA256, "artifact_path": b.ArtifactPath,
	}
	if b.ReceiptSchemaVersion != 0 {
		value["receipt_schema_version"] = b.ReceiptSchemaVersion
	}
	if b.ExecutionPolicy != "" {
		value["execution_policy"] = b.ExecutionPolicy
	}
	if b.Driver == "go-repository-v1" {
		value["repository"] = b.Repository
		value["declared_identity"] = b.DeclaredIdentity
		value["declared_locked_commit"] = b.DeclaredLockedCommit
		if b.DeclaredTag != "" {
			value["declared_tag"] = b.DeclaredTag
		}
		value["effective_identity"] = b.EffectiveIdentity
		value["object_format"] = b.ObjectFormat
		value["commit"] = b.Commit
		value["substituted"] = b.Substituted
		if b.Substitution != nil {
			value["substitution"] = b.Substitution
		}
		value["build_source"] = b.BuildSource
		value["descriptor_target"] = b.DescriptorTarget
	}
	return json.Marshal(value)
}

// RepositoryIdentity is a marker-v3 declared/effective source identity.
type RepositoryIdentity struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// RepositoryCommit is the package-declared immutable Git object lock.
type RepositoryCommit struct {
	ObjectFormat string `json:"object_format"`
	Hex          string `json:"hex"`
}

// RepositoryRef is the typed operator-selected network substitution ref.
type RepositoryRef struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// RepositorySubstitute records substitution semantics without paths or
// credentials. Local selectors deliberately collapse to type only.
type RepositorySubstitute struct {
	Type string         `json:"type"`
	Ref  *RepositoryRef `json:"ref,omitempty"`
}

// Marker is the install marker payload (Spec §8.5).
type Marker struct {
	SchemaVersion      int                   `json:"schema_version"`
	Name               string                `json:"name"`
	Source             string                `json:"source"`
	RefKind            string                `json:"ref_kind"`
	Ref                string                `json:"ref"`
	Commit             string                `json:"commit"`
	ContentSHA256      string                `json:"content_sha256"`
	Locale             string                `json:"locale,omitempty"`
	Agents             []string              `json:"agents"`
	Commands           []string              `json:"commands"`
	Dependencies       []string              `json:"dependencies"`
	SkillSchemaVersion int                   `json:"skill_schema_version"`
	RuntimeRoots       []string              `json:"runtime_roots"`
	BuildRoots         []string              `json:"build_roots"`
	BuildSource        *buildsource.Identity `json:"build_source,omitempty"`
	InstalledAt        string                `json:"installed_at"`
	Files              []string              `json:"files"`
	Builds             map[string]Build      `json:"builds"`
	Git                string                `json:"git,omitempty"`
	Requirements       []string              `json:"requirements,omitempty"`
	McpServers         map[string][]string   `json:"mcp_servers,omitempty"`
	Attestation        *Attestation          `json:"attestation,omitempty"`
	Activation         *Activation           `json:"activation,omitempty"`
	Requirers          []string              `json:"requirers,omitempty"`
	Substituted        string                `json:"substituted,omitempty"`
}

// BuildCurrentness supplies independently derived build state. RawSnapshot
// must return a validated token for the immutable package snapshot; returning
// nil, nil means the snapshot is absent. InspectCache must perform a read-only
// protected-cache inspection. ContextFiles and RuntimeFiles are complete
// relative file sets proving static build-root exclusion; nil means unknown.
type BuildCurrentness struct {
	RawSnapshot  func() (*buildsource.Token, error)
	InspectCache func(command string, expectation buildcache.Expectation) buildcache.Result
	Inputs       map[string]buildmeta.Input
	ContextFiles []string
	RuntimeFiles []string
	// InspectExternal verifies one receipt-v2 protected entry without fetching,
	// auditing, repairing, adopting, signing, or executing it.
	InspectExternal func(command string, recorded Build) (bool, error)
	// VerifyShim checks the manager-derived launcher/PATH relationship for any
	// local or external build command. Nil evidence fails closed for marker v3.
	VerifyShim func(command string, recorded Build) (bool, error)
}

// MarshalJSON keeps the marker object compatible with the protocol wire
// shape. The independent implementation writes an unselected locale as JSON
// null, while mandatory list fields are always arrays rather than null.
func (m Marker) MarshalJSON() ([]byte, error) {
	type plain Marker
	payload, err := json.Marshal(plain(m))
	if err != nil {
		return nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		return nil, err
	}
	if m.Locale == "" {
		object["locale"] = nil
	}
	if m.SchemaVersion == LegacySchemaVersion {
		delete(object, "build_roots")
		delete(object, "build_source")
		delete(object, "builds")
	}
	return json.Marshal(object)
}

// Read loads the marker of an installed directory; nil when absent or
// unreadable (an unreadable marker simply means "not current").
func Read(installedDir string) *Marker {
	payload, err := os.ReadFile(filepath.Join(installedDir, Name)) // #nosec G304 -- path derives from the install root
	if err != nil {
		return nil
	}
	if err := protocoljson.Validate(payload); err != nil {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	var m Marker
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&m); err != nil || !validMarker(&m, raw) {
		return nil
	}
	return &m
}

func validMarker(m *Marker, raw map[string]json.RawMessage) bool {
	commonRequired := []string{
		"schema_version", "name", "source", "ref_kind", "ref", "commit", "content_sha256", "locale",
		"agents", "commands", "dependencies", "skill_schema_version", "runtime_roots", "installed_at", "files",
	}
	commonOptional := []string{
		"git", "requirements", "mcp_servers", "attestation", "activation", "requirers", "substituted",
	}
	required := append([]string(nil), commonRequired...)
	allowed := append(append([]string(nil), commonRequired...), commonOptional...)
	switch m.SchemaVersion {
	case LegacySchemaVersion:
	case SchemaVersion:
		required = append(required, "build_roots", "builds")
		allowed = append(allowed, "build_roots", "build_source", "builds")
	case ExternalSchemaVersion, PolicySchemaVersion:
		required = append(required, "build_roots", "builds")
		allowed = append(allowed, "build_roots", "build_source", "builds")
	default:
		return false
	}
	for _, field := range required {
		if _, present := raw[field]; !present {
			return false
		}
	}
	if !onlyFields(raw, allowed) || !identifiers.Valid(m.Name) || !identifiers.PortablePath(m.Source) ||
		(m.RefKind != "tag" && m.RefKind != "branch" && m.RefKind != "revision") ||
		m.Ref == "" || utf8.RuneCountInString(m.Ref) > 8192 || !markerCommitRE.MatchString(m.Commit) ||
		!markerSHA256RE.MatchString(m.ContentSHA256) || m.SkillSchemaVersion < 0 ||
		(m.SchemaVersion == LegacySchemaVersion && m.SkillSchemaVersion > 5) ||
		(m.SchemaVersion == SchemaVersion && m.SkillSchemaVersion > 6) ||
		(m.SchemaVersion == ExternalSchemaVersion && m.SkillSchemaVersion != 7) ||
		(m.SchemaVersion == PolicySchemaVersion && m.SkillSchemaVersion != 8) {
		return false
	}
	setsSorted := m.SchemaVersion == SchemaVersion || m.SchemaVersion == ExternalSchemaVersion ||
		m.SchemaVersion == PolicySchemaVersion
	if !validNullableLocale(raw["locale"], m.Locale) || !validTimestamp(m.InstalledAt) ||
		!validIdentifierSet(m.Agents, setsSorted) || !validIdentifierSet(m.Commands, setsSorted) ||
		!validIdentifierSet(m.Dependencies, setsSorted) || !validPathSet(m.RuntimeRoots, setsSorted) ||
		!validPathSet(m.Files, setsSorted) {
		return false
	}
	if !validOptionalNonEmptyString(raw, "git", m.Git) || !validOptionalNonEmptyString(raw, "substituted", m.Substituted) {
		return false
	}
	if _, present := raw["requirements"]; present && !validIdentifierSet(m.Requirements, setsSorted) {
		return false
	}
	if _, present := raw["requirers"]; present && !validStringSet(m.Requirers, setsSorted) {
		return false
	}
	if _, present := raw["mcp_servers"]; present {
		if m.McpServers == nil {
			return false
		}
		for name, consumers := range m.McpServers {
			if !identifiers.Valid(name) || !validIdentifierSet(consumers, setsSorted) {
				return false
			}
		}
	}
	if attestationRaw, present := raw["attestation"]; present {
		object, ok := rawObject(attestationRaw)
		if !ok || m.Attestation == nil || object["registry"] == nil || object["status"] == nil ||
			m.Attestation.Registry == "" || utf8.RuneCountInString(m.Attestation.Registry) > 8192 ||
			(m.Attestation.Status != "audited" && m.Attestation.Status != "deprecated") {
			return false
		}
		if keyRaw, present := object["key_id"]; present {
			var keyID string
			if json.Unmarshal(keyRaw, &keyID) != nil || !markerKeyIDRE.MatchString(keyID) {
				return false
			}
		}
	}
	if activationRaw, present := raw["activation"]; present {
		object, ok := rawObject(activationRaw)
		if !ok || m.Activation == nil || object["context"] == nil || object["commands"] == nil ||
			!validBooleanRaw(object["context"]) || !validIdentifierSet(m.Activation.Commands, setsSorted) {
			return false
		}
	}
	if (m.SchemaVersion == SchemaVersion || m.SchemaVersion == ExternalSchemaVersion ||
		m.SchemaVersion == PolicySchemaVersion) && !validBuildState(m, raw) {
		return false
	}
	return true
}

func onlyFields(raw map[string]json.RawMessage, allowed []string) bool {
	fields := make(map[string]bool, len(allowed))
	for _, field := range allowed {
		fields[field] = true
	}
	for field := range raw {
		if !fields[field] {
			return false
		}
	}
	return true
}

func validBuildState(m *Marker, raw map[string]json.RawMessage) bool {
	if !validPathSet(m.BuildRoots, true) || m.Builds == nil {
		return false
	}
	_, sourcePresent := raw["build_source"]
	if len(m.Builds) == 0 {
		return !sourcePresent && m.BuildSource == nil
	}
	hasLocal := false
	for _, build := range m.Builds {
		hasLocal = hasLocal || build.Driver == buildmeta.DriverGoV1
	}
	if hasLocal {
		if len(m.BuildRoots) == 0 || !sourcePresent || m.BuildSource == nil || m.BuildSource.Algorithm != buildsource.Algorithm ||
			!markerSHA256RE.MatchString(m.BuildSource.ContentSHA256) {
			return false
		}
	} else if sourcePresent || m.BuildSource != nil {
		return false
	}
	for command, build := range m.Builds {
		if !identifiers.Valid(command) || !containsString(m.Commands, command) ||
			!markerSHA256RE.MatchString(string(build.CacheKey)) ||
			!markerSHA256RE.MatchString(string(build.ReceiptSHA256)) ||
			!markerSHA256RE.MatchString(build.ArtifactSHA256) || !identifiers.PortablePath(build.ArtifactPath) {
			return false
		}
		unixPath, unixErr := buildmeta.ArtifactPath(command, "linux")
		windowsPath, windowsErr := buildmeta.ArtifactPath(command, "windows")
		if unixErr != nil || windowsErr != nil || (build.ArtifactPath != unixPath && build.ArtifactPath != windowsPath) {
			return false
		}
		if m.SchemaVersion == SchemaVersion {
			if build.Driver != buildmeta.DriverGoV1 || build.ReceiptSchemaVersion != 0 || build.ExecutionPolicy != "" || hasRepositoryState(build) {
				return false
			}
			continue
		}
		if !validV3Build(build) {
			return false
		}
	}
	return true
}

func hasRepositoryState(build Build) bool {
	return build.Repository != "" || build.DeclaredIdentity != nil || build.DeclaredLockedCommit != nil ||
		build.DeclaredTag != "" || build.EffectiveIdentity != nil || build.ObjectFormat != "" ||
		build.Commit != "" || build.Substituted || build.Substitution != nil || build.BuildSource != nil ||
		build.DescriptorTarget != ""
}

func validV3Build(build Build) bool {
	if build.ExecutionPolicy != buildmeta.ExecutionPolicy {
		return false
	}
	if build.Driver == buildmeta.DriverGoV1 {
		return build.ReceiptSchemaVersion == 1 && !hasRepositoryState(build)
	}
	if build.Driver != "go-repository-v1" || build.ReceiptSchemaVersion != 2 ||
		!identifiers.Valid(build.Repository) || !identifiers.Valid(build.DescriptorTarget) ||
		build.DeclaredIdentity == nil || build.DeclaredIdentity.Kind != "network-git" || build.DeclaredIdentity.Value == "" ||
		build.DeclaredLockedCommit == nil || build.DeclaredLockedCommit.ObjectFormat != build.ObjectFormat ||
		build.EffectiveIdentity == nil || (build.EffectiveIdentity.Kind != "network-git" && build.EffectiveIdentity.Kind != "operator-local-git") ||
		build.EffectiveIdentity.Value == "" || build.BuildSource == nil || build.BuildSource.Algorithm != buildsource.Algorithm ||
		!markerSHA256RE.MatchString(build.BuildSource.ContentSHA256) {
		return false
	}
	wantLen := 40
	if build.ObjectFormat == "sha256" {
		wantLen = 64
	} else if build.ObjectFormat != "sha1" {
		return false
	}
	if len(build.Commit) != wantLen || len(build.DeclaredLockedCommit.Hex) != wantLen ||
		!markerCommitRE.MatchString(build.Commit) || !markerCommitRE.MatchString(build.DeclaredLockedCommit.Hex) {
		return false
	}
	if build.Substituted != (build.Substitution != nil) {
		return false
	}
	if build.Substitution != nil && build.Substitution.Type != "local-path" &&
		(build.Substitution.Type != "network-git" || build.Substitution.Ref == nil) {
		return false
	}
	return true
}

func validNullableLocale(raw json.RawMessage, decoded string) bool {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return true
	}
	var value string
	return json.Unmarshal(raw, &value) == nil && value == decoded && identifiers.ValidLocale(value)
}

func validOptionalNonEmptyString(raw map[string]json.RawMessage, field, decoded string) bool {
	value, present := raw[field]
	if !present {
		return true
	}
	var text string
	return json.Unmarshal(value, &text) == nil && text == decoded && text != "" && utf8.RuneCountInString(text) <= 8192
}

func validTimestamp(value string) bool {
	parsed, err := time.Parse(time.RFC3339, value)
	return err == nil && parsed.UTC().Format("2006-01-02T15:04:05Z") == value
}

func validIdentifierSet(values []string, sorted bool) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for index, value := range values {
		if !identifiers.Valid(value) || seen[value] {
			return false
		}
		if sorted && index > 0 && values[index-1] >= value {
			return false
		}
		seen[value] = true
	}
	return true
}

func validPathSet(values []string, sorted bool) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for index, value := range values {
		if !identifiers.PortablePath(value) || seen[value] {
			return false
		}
		if sorted && index > 0 && values[index-1] >= value {
			return false
		}
		seen[value] = true
	}
	return true
}

func validStringSet(values []string, sorted bool) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for index, value := range values {
		if !utf8.ValidString(value) || seen[value] {
			return false
		}
		if sorted && index > 0 && values[index-1] >= value {
			return false
		}
		seen[value] = true
	}
	return true
}

func validBooleanRaw(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false"))
}

func rawObject(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || object == nil {
		return nil, false
	}
	return object, true
}

// buildBearingSchema reports whether a marker of this schema carries the
// build_roots/builds/build_source triple.
func buildBearingSchema(version int) bool {
	return version == SchemaVersion || externalCapableSchema(version)
}

// externalCapableSchema reports whether a marker of this schema can record
// external go-repository-v1 commands alongside local ones. Marker v4 is v3
// with the version bumped and nothing else changed, so both answer yes.
func externalCapableSchema(version int) bool {
	return version == ExternalSchemaVersion || version == PolicySchemaVersion
}

// supportedMarkerSchema reports whether version is a marker schema this
// release reads. Every listed version stays readable for the whole of protocol
// 1.x; only the written version advances with the manifest band.
func supportedMarkerSchema(version int) bool {
	return version == LegacySchemaVersion || version == SchemaVersion ||
		version == ExternalSchemaVersion || version == PolicySchemaVersion
}

// Write stores the marker inside dir with sorted keys and a trailing newline.
func Write(dir string, m *Marker) error {
	if m == nil {
		return errors.New("install marker is nil")
	}
	switch {
	case m.SkillSchemaVersion >= 8:
		m.SchemaVersion = PolicySchemaVersion
	case m.SkillSchemaVersion == 7:
		m.SchemaVersion = ExternalSchemaVersion
	default:
		m.SchemaVersion = SchemaVersion
	}
	m.Agents = nonNilStrings(m.Agents)
	m.Commands = nonNilStrings(m.Commands)
	m.Dependencies = nonNilStrings(m.Dependencies)
	m.RuntimeRoots = nonNilStrings(m.RuntimeRoots)
	m.BuildRoots = nonNilStrings(m.BuildRoots)
	m.Files = nonNilStrings(m.Files)
	if m.Builds == nil {
		m.Builds = map[string]Build{}
	}
	sort.Strings(m.Agents)
	sort.Strings(m.Commands)
	sort.Strings(m.Dependencies)
	sort.Strings(m.RuntimeRoots)
	sort.Strings(m.BuildRoots)
	sort.Strings(m.Files)
	if m.Requirements != nil {
		sort.Strings(m.Requirements)
	}
	for name := range m.McpServers {
		sort.Strings(m.McpServers[name])
	}
	if m.Activation != nil {
		m.Activation.Commands = nonNilStrings(m.Activation.Commands)
		sort.Strings(m.Activation.Commands)
	}
	if m.Requirers != nil {
		sort.Strings(m.Requirers)
	}
	payload, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if protocoljson.Validate(payload) != nil {
		return errors.New("install marker cannot be encoded as strict JSON")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil || !validMarker(m, raw) {
		return fmt.Errorf("install marker is invalid for schema %d", m.SchemaVersion)
	}
	return os.WriteFile(filepath.Join(dir, Name), append(payload, '\n'), 0o644)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// Current reports whether the installed directory is up to date for the
// effective marker. Schema 1 remains eligible for schema 1 through 5 installs.
// A build-enabled schema-2 marker additionally requires one BuildCurrentness
// value; missing validation evidence fails closed as non-current.
func Current(installedDir string, expected *Marker, buildState ...BuildCurrentness) (bool, error) {
	if expected == nil {
		return false, errors.New("expected install marker is nil")
	}
	if len(buildState) > 1 {
		return false, errors.New("multiple build currentness values supplied")
	}
	if version, ok := markerSchemaVersion(installedDir); ok && !supportedMarkerSchema(version) {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
	}
	recorded := Read(installedDir)
	if recorded == nil {
		return false, nil
	}
	if !supportedMarkerSchema(recorded.SchemaVersion) {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
	}
	if recorded.SchemaVersion == LegacySchemaVersion &&
		(expected.SkillSchemaVersion > 5 || len(expected.BuildRoots) != 0 || len(expected.Builds) != 0 || expected.BuildSource != nil) {
		return false, nil
	}
	if recorded.RefKind != expected.RefKind || recorded.Ref != expected.Ref || recorded.Commit != expected.Commit {
		return false, nil
	}
	if recorded.Locale != expected.Locale {
		return false, nil
	}
	if !equalStrings(recorded.Agents, expected.Agents) {
		return false, nil
	}
	if !reflect.DeepEqual(normalizeActivation(recorded.Activation), normalizeActivation(expected.Activation)) {
		return false, nil
	}
	if recorded.Substituted != expected.Substituted {
		return false, nil
	}
	if expected.McpServers != nil && !reflect.DeepEqual(recorded.McpServers, expected.McpServers) {
		return false, nil
	}
	if !reflect.DeepEqual(recorded.Attestation, expected.Attestation) {
		return false, nil
	}
	if buildBearingSchema(recorded.SchemaVersion) {
		if !equalStrings(recorded.BuildRoots, normalizedStrings(expected.BuildRoots)) ||
			!reflect.DeepEqual(recorded.Builds, normalizedBuilds(expected.Builds)) ||
			!reflect.DeepEqual(recorded.BuildSource, expected.BuildSource) {
			return false, nil
		}
	}
	actual, err := hashing.ContentSHA256(installedDir, nil)
	if err != nil {
		return false, err
	}
	if recorded.ContentSHA256 != actual {
		return false, nil
	}
	if len(recorded.Builds) == 0 {
		return true, nil
	}
	if !buildBearingSchema(recorded.SchemaVersion) || len(buildState) != 1 {
		return false, nil
	}
	return currentBuilds(installedDir, recorded, buildState[0])
}

func currentBuilds(installedDir string, recorded *Marker, state BuildCurrentness) (bool, error) {
	if state.ContextFiles == nil || state.RuntimeFiles == nil {
		return false, nil
	}
	contextFiles := normalizedStrings(state.ContextFiles)
	if !equalStrings(contextFiles, recorded.Files) || pathsUnderRoots(contextFiles, recorded.BuildRoots) ||
		pathsUnderRoots(state.RuntimeFiles, recorded.BuildRoots) {
		return false, nil
	}
	for _, root := range recorded.BuildRoots {
		if _, err := os.Lstat(filepath.Join(installedDir, filepath.FromSlash(root))); err == nil {
			return false, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
	}
	localBuilds := map[string]Build{}
	externalBuilds := map[string]Build{}
	for command, build := range recorded.Builds {
		if build.Driver == buildmeta.DriverGoV1 {
			localBuilds[command] = build
		} else {
			externalBuilds[command] = build
		}
	}
	if len(localBuilds) != len(state.Inputs) {
		return false, nil
	}
	for command := range localBuilds {
		if _, present := state.Inputs[command]; !present {
			return false, nil
		}
	}
	for command, build := range externalBuilds {
		if state.InspectExternal == nil || state.VerifyShim == nil {
			return false, nil
		}
		ok, err := state.InspectExternal(command, build)
		if err != nil || !ok {
			return false, err
		}
		ok, err = state.VerifyShim(command, build)
		if err != nil || !ok {
			return false, err
		}
	}
	if len(localBuilds) == 0 {
		return true, nil
	}
	if state.RawSnapshot == nil || state.InspectCache == nil || state.Inputs == nil || recorded.BuildSource == nil {
		return false, nil
	}

	token, err := state.RawSnapshot()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, buildsource.ErrInvalidSnapshot) {
			return false, nil
		}
		return false, err
	}
	if token == nil {
		return false, nil
	}
	current := true
	useErr := token.Use(func(token *buildsource.Token) error {
		identity := token.Identity()
		if identity != *recorded.BuildSource {
			current = false
			return nil
		}
		commands := make([]string, 0, len(localBuilds))
		for command := range localBuilds {
			commands = append(commands, command)
		}
		sort.Strings(commands)
		for _, command := range commands {
			build := localBuilds[command]
			input := state.Inputs[command]
			if input.Command != command || input.Driver != build.Driver || input.BuildSource != identity ||
				!containsString(recorded.BuildRoots, input.BuildRoot) || input.Validate() != nil {
				current = false
				return nil
			}
			key, keyErr := input.CacheKey()
			artifactPath, pathErr := buildmeta.ArtifactPath(command, input.Target.GOOS)
			if keyErr != nil || pathErr != nil || key != build.CacheKey || artifactPath != build.ArtifactPath {
				current = false
				return nil
			}
			result := state.InspectCache(command, buildcache.Expectation{Input: input, ReceiptHash: build.ReceiptSHA256})
			if !validCacheResult(result, input, build) {
				current = false
				return nil
			}
			if externalCapableSchema(recorded.SchemaVersion) {
				if state.VerifyShim == nil {
					current = false
					return nil
				}
				shimOK, shimErr := state.VerifyShim(command, build)
				if shimErr != nil {
					return shimErr
				}
				if !shimOK {
					current = false
					return nil
				}
			}
		}
		return nil
	})
	return buildCurrentnessResult(current, useErr, token.Close())
}

func buildCurrentnessResult(current bool, useErr, closeErr error) (bool, error) {
	if closeErr != nil {
		return false, errors.Join(useErr, closeErr)
	}
	if useErr != nil {
		if errors.Is(useErr, buildsource.ErrSnapshotMutated) || errors.Is(useErr, buildsource.ErrInvalidSnapshot) {
			return false, nil
		}
		return false, useErr
	}
	return current, nil
}

func validCacheResult(result buildcache.Result, input buildmeta.Input, build Build) bool {
	if result.Status != buildcache.Hit || result.ArtifactPath == "" ||
		!reflect.DeepEqual(result.Receipt.Input, input) || result.Receipt.CacheKey != build.CacheKey ||
		result.ReceiptHash != build.ReceiptSHA256 || result.Receipt.Artifact.Path != build.ArtifactPath ||
		result.Receipt.Artifact.SHA256 != build.ArtifactSHA256 {
		return false
	}
	hash, err := buildmeta.HashReceiptBytes(result.ReceiptBytes)
	if err != nil || hash != build.ReceiptSHA256 {
		return false
	}
	receipt, err := buildmeta.DecodeExpectedReceipt(result.ReceiptBytes, input)
	return err == nil && reflect.DeepEqual(receipt, result.Receipt)
}

func pathsUnderRoots(paths, roots []string) bool {
	for _, path := range paths {
		for _, root := range roots {
			if path == root || len(path) > len(root) && path[:len(root)] == root && path[len(root)] == '/' {
				return true
			}
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	index := sort.SearchStrings(values, want)
	return index < len(values) && values[index] == want
}

func normalizedStrings(values []string) []string {
	result := append([]string(nil), values...)
	if result == nil {
		result = []string{}
	}
	sort.Strings(result)
	return result
}

func normalizedBuilds(values map[string]Build) map[string]Build {
	if values == nil {
		return map[string]Build{}
	}
	return values
}

func markerSchemaVersion(installedDir string) (int, bool) {
	payload, err := os.ReadFile(filepath.Join(installedDir, Name)) // #nosec G304 -- path derives from the install root
	if err != nil || protocoljson.Validate(payload) != nil {
		return 0, false
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(payload, &object) != nil {
		return 0, false
	}
	var version int
	if json.Unmarshal(object["schema_version"], &version) != nil {
		return 0, false
	}
	return version, true
}

// ReplaceDir atomically swaps newDir into target: back up, rename, roll back
// on failure, drop the backup on success (Spec §8.5).
func ReplaceDir(newDir, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	backup := filepath.Join(filepath.Dir(target), fmt.Sprintf(".%s.backup-%d", filepath.Base(target), os.Getpid()))
	if err := os.RemoveAll(backup); err != nil {
		return err
	}
	if _, err := os.Lstat(target); err == nil {
		if err := os.Rename(target, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(newDir, target); err != nil {
		if _, statErr := os.Lstat(backup); statErr == nil {
			if _, targetErr := os.Lstat(target); targetErr != nil {
				_ = os.Rename(backup, target)
			}
		}
		return err
	}
	return os.RemoveAll(backup)
}

func normalizeActivation(a *Activation) *Activation {
	if a == nil {
		return nil
	}
	commands := append([]string(nil), a.Commands...)
	sort.Strings(commands)
	if len(commands) == 0 {
		commands = []string{}
	}
	return &Activation{Context: a.Context, Commands: commands}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
