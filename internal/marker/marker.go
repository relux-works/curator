// Package marker reads and writes install markers (.csk-install.json) and
// implements the up-to-date and tamper-detection semantics of Spec §8.5.
package marker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closureexec"
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
	// SchemaV5 is written only for skillfile-sources draft installations
	// (Skillfile schema 2). Its package replaces the legacy
	// source/git/ref_kind/ref/commit fields and its lock_sha256 binds the
	// installed selection and declared ref through the validated lock and
	// matching manifest. Every locked draft member records it, whatever
	// the skill manifest version; legacy lanes never write it.
	SchemaV5 = 5
	// NewestSchemaVersion is the highest marker schema this release reads. It
	// is what an operator is told when a document from a newer manager is
	// refused, so it must advance with every new readable schema.
	NewestSchemaVersion = SchemaV5
)

var (
	markerCommitRE    = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
	markerSHA256RE    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	markerSHA1RE      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	markerSHA256HexRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
	markerKeyIDRE     = regexp.MustCompile(`^[0-9a-f]{16}$`)
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

// Package is the draft marker-v5 frozen package identity. It mirrors
// the disjoint source-types arms: a local snapshot never carries Git
// fields and a Git package never carries a snapshot digest. The digest
// forms are never substituted across arms.
type Package struct {
	Kind       string  `json:"kind"`
	Snapshot   string  `json:"snapshot,omitempty"`
	Repository string  `json:"repository,omitempty"`
	Source     string  `json:"source,omitempty"`
	Commit     *Commit `json:"commit,omitempty"`
	Directory  string  `json:"directory,omitempty"`
}

// Commit is a locked Git object: format-bound lowercase hex, never a
// snapshot digest and never prefixed.
type Commit struct {
	ObjectFormat string `json:"object_format"`
	Hex          string `json:"hex"`
}

// Marker is the install marker payload (Spec §8.5).
type Marker struct {
	SchemaVersion      int                   `json:"schema_version"`
	Name               string                `json:"name"`
	Source             string                `json:"source,omitempty"`
	RefKind            string                `json:"ref_kind,omitempty"`
	Ref                string                `json:"ref,omitempty"`
	Commit             string                `json:"commit,omitempty"`
	Package            *Package              `json:"package,omitempty"`
	LockSHA256         string                `json:"lock_sha256,omitempty"`
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
	Assurance    closureexec.AssuranceBinding
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
	// A draft v5 marker carries no legacy source identity: the replaced
	// fields stay empty and omitempty keeps them absent, so no reader can
	// mistake one for an unattributed legacy marker. A builder that sets
	// them fails v5 validation below instead of being silently masked.
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
	case SchemaV5:
		// The frozen package replaces every legacy source field, so v5
		// carries its own required set without them.
		required = []string{
			"schema_version", "name", "package", "lock_sha256", "content_sha256", "locale",
			"agents", "commands", "dependencies", "skill_schema_version", "runtime_roots",
			"build_roots", "builds", "installed_at", "files",
		}
		allowed = append(append([]string(nil), required...),
			"requirements", "mcp_servers", "attestation", "activation", "requirers", "substituted",
			"build_source")
	default:
		return false
	}
	for _, field := range required {
		if _, present := raw[field]; !present {
			return false
		}
	}
	if !onlyFields(raw, allowed) || !identifiers.Valid(m.Name) {
		return false
	}
	if m.SchemaVersion == SchemaV5 {
		if !validV5Identity(m, raw) {
			return false
		}
	} else if !identifiers.PortablePath(m.Source) || !validLegacyTriple(m) {
		return false
	}
	if !markerSHA256RE.MatchString(m.ContentSHA256) || m.SkillSchemaVersion < 0 ||
		(m.SchemaVersion == LegacySchemaVersion && m.SkillSchemaVersion > 5) ||
		(m.SchemaVersion == SchemaVersion && m.SkillSchemaVersion > 6) ||
		(m.SchemaVersion == ExternalSchemaVersion && m.SkillSchemaVersion != 7) ||
		(m.SchemaVersion == PolicySchemaVersion && m.SkillSchemaVersion != 8) ||
		(m.SchemaVersion == SchemaV5 && (m.SkillSchemaVersion < 1 || m.SkillSchemaVersion > 9)) {
		return false
	}
	setsSorted := m.SchemaVersion == SchemaVersion || m.SchemaVersion == ExternalSchemaVersion ||
		m.SchemaVersion == PolicySchemaVersion || m.SchemaVersion == SchemaV5
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
		m.SchemaVersion == PolicySchemaVersion || m.SchemaVersion == SchemaV5) && !validBuildState(m, raw) {
		return false
	}
	return true
}

// validV5Identity enforces the draft marker-v5 identity: the frozen package
// replaces every legacy source field, the lock binds the installed selection
// through the validated lock, and a local snapshot admits neither a network
// attestation nor a development substitution.
func validV5Identity(m *Marker, raw map[string]json.RawMessage) bool {
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
		if _, present := raw[field]; present {
			return false
		}
	}
	if !validV5PackageShape(raw["package"]) || !validV5Package(m.Package) {
		return false
	}
	if !markerSHA256RE.MatchString(m.LockSHA256) {
		return false
	}
	if m.Package.Kind == "local-snapshot" {
		if _, present := raw["attestation"]; present {
			return false
		}
		if _, present := raw["substituted"]; present {
			return false
		}
	}
	return true
}

// validLegacyTriple reports whether the declared ref identity uses the
// legacy grammar: a known ref kind, a non-empty bounded ref and a hex
// commit. Legacy markers require it.
func validLegacyTriple(m *Marker) bool {
	return (m.RefKind == "tag" || m.RefKind == "branch" || m.RefKind == "revision") &&
		m.Ref != "" && utf8.RuneCountInString(m.Ref) <= 8192 && markerCommitRE.MatchString(m.Commit)
}

// v5PackageArms is the closed raw shape of every source-types package arm:
// the exact member set the schema admits, each bound to its JSON type. The
// shared Package struct knows the fields of every arm, so a decoded value
// cannot tell a foreign field carrying null or "" from an absent one; the
// raw object can, and must be checked before the lossy decode.
var v5PackageArms = map[string]map[string]string{
	"local-snapshot": {"kind": "string", "snapshot": "string"},
	"network-git":    {"kind": "string", "repository": "string", "commit": "object", "directory": "string"},
	"configured-git": {"kind": "string", "source": "string", "commit": "object", "directory": "string"},
}

// v5CommitShape is the closed lockedCommit object: both members required,
// both strings.
var v5CommitShape = map[string]string{"object_format": "string", "hex": "string"}

// v5ExternalBuildTypes is the closed member-to-JSON-type map of the v5
// external go-repository-v1 build record arm (install-marker-v5.schema.json
// external oneOf branch, receipt_schema_version 3). The decoded Build value
// cannot tell an absent `substituted` from an explicit false, nor a null
// member from an absent one; the raw object can, and must be checked before
// the lossy decode, the same discipline as the package arms above.
var v5ExternalBuildTypes = map[string]string{
	"driver":                 "string",
	"receipt_schema_version": "number",
	"execution_policy":       "string",
	"repository":             "string",
	"declared_identity":      "object",
	"declared_locked_commit": "object",
	"effective_identity":     "object",
	"object_format":          "string",
	"commit":                 "string",
	"substituted":            "boolean",
	"build_source":           "object",
	"descriptor_target":      "string",
	"cache_key":              "string",
	"receipt_sha256":         "string",
	"artifact_sha256":        "string",
	"artifact_path":          "string",
	"declared_tag":           "string",
	"substitution":           "object",
}

// v5ExternalBuildRequired is the required member set of the external arm:
// every type member above except the optional declared_tag and the
// substituted-conditional substitution.
var v5ExternalBuildRequired = []string{
	"driver", "receipt_schema_version", "execution_policy", "repository",
	"declared_identity", "declared_locked_commit", "effective_identity",
	"object_format", "commit", "substituted", "build_source",
	"descriptor_target", "cache_key", "receipt_sha256", "artifact_sha256",
	"artifact_path",
}

// Closed nested shapes of the external arm. Declared and effective
// identities share the same member names (kind/value) across their
// network/local variants; the const values stay the job of the decoded
// validation. The locked commit reuses v5CommitShape.
var (
	v5IdentityShape            = map[string]string{"kind": "string", "value": "string"}
	v5BuildSourceShape         = map[string]string{"algorithm": "string", "content_sha256": "string"}
	v5SubstitutionLocalShape   = map[string]string{"type": "string"}
	v5SubstitutionNetworkShape = map[string]string{"type": "string", "ref": "object"}
	v5SubstitutionRefShape     = map[string]string{"kind": "string", "value": "string"}
)

// validV5PackageShape validates the raw package object against the closed
// shape of its arm: every arm member present with the right JSON type,
// nothing else present even as null or empty, and the nested commit object
// closed the same way. Duplicate keys and trailing data were already
// refused by protocoljson.Validate over the whole document.
func validV5PackageShape(raw json.RawMessage) bool {
	object, ok := rawObject(raw)
	if !ok {
		return false
	}
	var kind string
	if kindRaw, present := object["kind"]; !present || json.Unmarshal(kindRaw, &kind) != nil {
		return false
	}
	arm, known := v5PackageArms[kind]
	if !known || !closedRawShape(object, arm) {
		return false
	}
	if commitRaw, present := object["commit"]; present {
		commit, ok := rawObject(commitRaw)
		if !ok || !closedRawShape(commit, v5CommitShape) {
			return false
		}
	}
	return true
}

// closedRawShape reports whether object carries exactly the members of
// shape, each with the declared JSON type. A null member is neither absent
// nor of its type, so it is refused either way.
func closedRawShape(object map[string]json.RawMessage, shape map[string]string) bool {
	if len(object) != len(shape) {
		return false
	}
	for member, kind := range shape {
		value, present := object[member]
		if !present || rawJSONType(value) != kind {
			return false
		}
	}
	return true
}

// rawJSONType names the JSON type of a raw value: string, object,
// boolean, number, or "" for anything else (null, arrays, malformed).
func rawJSONType(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	switch {
	case len(trimmed) == 0:
		return ""
	case trimmed[0] == '"':
		var s string
		if json.Unmarshal(trimmed, &s) != nil {
			return ""
		}
		return "string"
	case trimmed[0] == '{':
		if _, ok := rawObject(trimmed); !ok {
			return ""
		}
		return "object"
	case trimmed[0] == 't' || trimmed[0] == 'f':
		if bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false")) {
			return "boolean"
		}
		return ""
	case trimmed[0] == '-' || (trimmed[0] >= '0' && trimmed[0] <= '9'):
		var n json.Number
		if json.Unmarshal(trimmed, &n) != nil {
			return ""
		}
		return "number"
	default:
		return ""
	}
}

// validV5ExternalBuildShape validates the raw external build record against
// the closed shape of its arm: every required member present with the right
// JSON type, nothing else present even as null, the substituted/substitution
// conditional, and the nested identity, commit, build-source and substitution
// objects closed the same way. Value grammars behind the member types stay
// the job of the decoded validation; this check pins the closed shape, the
// class that admitted an absent substituted as an explicit false.
func validV5ExternalBuildShape(raw json.RawMessage) bool {
	object, ok := rawObject(raw)
	if !ok {
		return false
	}
	for _, field := range v5ExternalBuildRequired {
		if _, present := object[field]; !present {
			return false
		}
	}
	for member, value := range object {
		want, allowed := v5ExternalBuildTypes[member]
		if !allowed || rawJSONType(value) != want {
			return false
		}
	}
	// Presence of substituted was enforced by the required loop above,
	// which is the single refusal point for an absent member: the decode
	// below tolerates absence so a mutant dropping the requirement admits
	// the record instead of surviving behind a redundant refusal.
	var substituted bool
	if subRaw, present := object["substituted"]; present {
		if json.Unmarshal(subRaw, &substituted) != nil {
			return false
		}
	}
	_, hasSubstitution := object["substitution"]
	if substituted != hasSubstitution {
		return false
	}
	for _, field := range []string{"declared_identity", "effective_identity"} {
		nested, ok := rawObject(object[field])
		if !ok || !closedRawShape(nested, v5IdentityShape) {
			return false
		}
	}
	if nested, ok := rawObject(object["declared_locked_commit"]); !ok || !closedRawShape(nested, v5CommitShape) {
		return false
	}
	if nested, ok := rawObject(object["build_source"]); !ok || !closedRawShape(nested, v5BuildSourceShape) {
		return false
	}
	if subRaw, present := object["substitution"]; present {
		sub, ok := rawObject(subRaw)
		if !ok {
			return false
		}
		var subType string
		if typeRaw, present := sub["type"]; !present || json.Unmarshal(typeRaw, &subType) != nil {
			return false
		}
		switch subType {
		case "local-path":
			if !closedRawShape(sub, v5SubstitutionLocalShape) {
				return false
			}
		case "network-git":
			if !closedRawShape(sub, v5SubstitutionNetworkShape) {
				return false
			}
			ref, ok := rawObject(sub["ref"])
			if !ok || !closedRawShape(ref, v5SubstitutionRefShape) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// validV5Package mirrors the disjoint source-types arms: a local snapshot
// never carries Git fields and a Git package never carries a snapshot
// digest. Digest forms are never substituted across arms.
func validV5Package(p *Package) bool {
	if p == nil {
		return false
	}
	switch p.Kind {
	case "local-snapshot":
		if !markerSHA256RE.MatchString(p.Snapshot) {
			return false
		}
		if p.Repository != "" || p.Source != "" || p.Commit != nil || p.Directory != "" {
			return false
		}
	case "network-git":
		if p.Snapshot != "" || p.Source != "" {
			return false
		}
		if p.Repository == "" || strings.HasSuffix(p.Repository, ".git") {
			return false
		}
		if !validV5Commit(p.Commit) {
			return false
		}
		if p.Directory != "." && (!identifiers.PortablePath(p.Directory) || strings.ContainsAny(p.Directory, "*?[]")) {
			return false
		}
	case "configured-git":
		if p.Snapshot != "" || p.Repository != "" {
			return false
		}
		if !identifiers.PortablePath(p.Source) {
			return false
		}
		if !validV5Commit(p.Commit) {
			return false
		}
		if p.Directory != "." {
			return false
		}
	default:
		return false
	}
	return true
}

// validV5Commit enforces the lockedCommit shape: sha1 binds 40 hex, sha256
// binds 64. A snapshot digest never satisfies it: it carries a prefix the
// bare hex grammar rejects.
func validV5Commit(c *Commit) bool {
	if c == nil {
		return false
	}
	switch c.ObjectFormat {
	case "sha1":
		return markerSHA1RE.MatchString(c.Hex)
	case "sha256":
		return markerSHA256HexRE.MatchString(c.Hex)
	default:
		return false
	}
}

// object renders the canonical package preimage: exactly the disjoint
// source-types arms the lock validation hashes, so a marker package and
// its lock member digest identically.
func (p *Package) object() map[string]any {
	if p == nil {
		return nil
	}
	switch p.Kind {
	case "network-git":
		return map[string]any{
			"kind":       p.Kind,
			"repository": p.Repository,
			"commit":     map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory":  p.Directory,
		}
	case "configured-git":
		return map[string]any{
			"kind":      p.Kind,
			"source":    p.Source,
			"commit":    map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory": p.Directory,
		}
	default:
		return map[string]any{"kind": p.Kind, "snapshot": p.Snapshot}
	}
}

// Digest returns the sha256:<hex> content identity of the package over its
// CCJ-1 bytes. It must agree with the lock package digest for the same
// identity: the runtime store, status, repair, rollback and GC all follow
// this key, and a mismatch would strand or sweep live trees.
func (p *Package) Digest() (string, error) {
	if !validV5Package(p) {
		return "", fmt.Errorf("draft marker package is invalid")
	}
	canonical, err := protocoljson.MarshalCanonical(p.object())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
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
	var buildsRaw map[string]json.RawMessage
	if m.SchemaVersion == SchemaV5 {
		if err := json.Unmarshal(raw["builds"], &buildsRaw); err != nil || buildsRaw == nil {
			return false
		}
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
		// A draft v5 marker binds receipt version 3 on every build entry,
		// whatever the skill schema: its cache entries are receipt-3 package
		// wrappers, so a legacy record shape can never describe them.
		if m.SchemaVersion == SchemaV5 {
			if build.Driver == "go-repository-v1" {
				rawBuild, present := buildsRaw[command]
				if !present || !validV5ExternalBuildShape(rawBuild) {
					return false
				}
			}
			if !validV5Build(build) {
				return false
			}
			continue
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
	return validDriverBuild(build, 1, 2)
}

// validV5Build is the marker-5 record: every retained driver field of the
// v3/v4 record with receipt_schema_version 3 on both arms.
func validV5Build(build Build) bool {
	return validDriverBuild(build, buildmeta.SourceAwareSchemaVersion, buildmeta.SourceAwareSchemaVersion)
}

func validDriverBuild(build Build, localReceipt, externalReceipt int) bool {
	if build.ExecutionPolicy != buildmeta.ExecutionPolicy {
		return false
	}
	if build.Driver == buildmeta.DriverGoV1 {
		return build.ReceiptSchemaVersion == localReceipt && !hasRepositoryState(build)
	}
	if build.Driver != "go-repository-v1" || build.ReceiptSchemaVersion != externalReceipt ||
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

// BuildBearingSchema reports whether a marker of this schema carries the
// build_roots/builds/build_source triple.
//
// Every reader that decides whether a recorded compiled command is knowable
// asks this, so a new build-bearing schema is admitted in one place instead of
// in each reader's own inequality. A reader that bands on a single schema
// silently reports a perfectly current installation as needing reinstallation
// the moment the written schema advances.
func BuildBearingSchema(version int) bool {
	return version == SchemaVersion || version == SchemaV5 || externalCapableSchema(version)
}

// externalCapableSchema reports whether a marker of this schema can record
// external go-repository-v1 commands alongside local ones. Marker v4 is v3
// with the version bumped and nothing else changed, so both answer yes.
func externalCapableSchema(version int) bool {
	return version == ExternalSchemaVersion || version == PolicySchemaVersion
}

// SupportedSchema reports whether version is a marker schema this release
// reads. Every listed version stays readable for the whole of protocol 1.x;
// only the written version advances with the manifest band.
func SupportedSchema(version int) bool {
	return version == LegacySchemaVersion || version == SchemaVersion ||
		version == ExternalSchemaVersion || version == PolicySchemaVersion ||
		version == SchemaV5
}

// Write stores the marker inside dir with sorted keys and a trailing newline.
//
// A marker carrying a frozen draft package is always schema 5, regardless
// of skill manifest version; every other marker keeps the legacy
// skill-schema band byte-identically.
func Write(dir string, m *Marker) error {
	if m == nil {
		return errors.New("install marker is nil")
	}
	switch {
	case m.Package != nil:
		m.SchemaVersion = SchemaV5
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
	if version, ok := markerSchemaVersion(installedDir); ok && !SupportedSchema(version) {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
	}
	recorded := Read(installedDir)
	if recorded == nil {
		return false, nil
	}
	if !SupportedSchema(recorded.SchemaVersion) {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
	}
	if recorded.SchemaVersion == LegacySchemaVersion &&
		(expected.SkillSchemaVersion > 5 || len(expected.BuildRoots) != 0 || len(expected.Builds) != 0 || expected.BuildSource != nil) {
		return false, nil
	}
	if recorded.RefKind != expected.RefKind || recorded.Ref != expected.Ref || recorded.Commit != expected.Commit {
		return false, nil
	}
	// A draft marker is current only for the exact frozen package and lock
	// generation it was installed from: a runtime-only refresh changes the
	// package identity while the projected context stays identical, so the
	// lock comparison below is what observes it. The expectation is staged
	// fresh on every run, so its schema version is unset until Write; the
	// frozen package it carries is what selects the v5 comparison. A v5
	// marker never matches a legacy one; changed registry, substitution,
	// declared ref, package or lock makes the installation non-current.
	if recorded.SchemaVersion == SchemaV5 || expected.Package != nil {
		if recorded.SchemaVersion != SchemaV5 || expected.Package == nil {
			return false, nil
		}
		if !reflect.DeepEqual(recorded.Package, expected.Package) || recorded.LockSHA256 != expected.LockSHA256 {
			return false, nil
		}
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
	if BuildBearingSchema(recorded.SchemaVersion) {
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
	if !BuildBearingSchema(recorded.SchemaVersion) || len(buildState) != 1 {
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
			logicalKey, keyErr := input.CacheKey()
			assuredID, assuranceErr := (closureexec.AssuredBuildCacheInput{BuildInput: input, Binding: state.Assurance}).ID()
			key := buildmeta.CacheKey(assuredID)
			artifactPath, pathErr := buildmeta.ArtifactPath(command, input.Target.GOOS)
			if keyErr != nil || assuranceErr != nil || pathErr != nil || key != build.CacheKey || artifactPath != build.ArtifactPath {
				current = false
				return nil
			}
			result := state.InspectCache(command, buildcache.Expectation{Input: input, ReceiptHash: build.ReceiptSHA256, Assurance: state.Assurance})
			if !validCacheResult(result, input, logicalKey, build) {
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

func validCacheResult(result buildcache.Result, input buildmeta.Input, logicalKey buildmeta.CacheKey, build Build) bool {
	if result.Status != buildcache.Hit || result.ArtifactPath == "" ||
		!reflect.DeepEqual(result.Receipt.Input, input) || result.Receipt.CacheKey != logicalKey || result.CacheKey != build.CacheKey ||
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
