// Package marker reads and writes install markers (.csk-install.json) and
// implements the up-to-date and tamper-detection semantics of Spec §8.5.
package marker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Name is the marker file name inside an installed skill directory.
const Name = ".csk-install.json"

// SchemaVersion is the supported marker schema.
const SchemaVersion = 1

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

// Marker is the install marker payload (Spec §8.5).
type Marker struct {
	SchemaVersion      int                 `json:"schema_version"`
	Name               string              `json:"name"`
	Source             string              `json:"source"`
	RefKind            string              `json:"ref_kind"`
	Ref                string              `json:"ref"`
	Commit             string              `json:"commit"`
	ContentSHA256      string              `json:"content_sha256"`
	Locale             string              `json:"locale,omitempty"`
	Agents             []string            `json:"agents"`
	Commands           []string            `json:"commands"`
	Dependencies       []string            `json:"dependencies"`
	SkillSchemaVersion int                 `json:"skill_schema_version"`
	RuntimeRoots       []string            `json:"runtime_roots"`
	InstalledAt        string              `json:"installed_at"`
	Files              []string            `json:"files"`
	Git                string              `json:"git,omitempty"`
	Requirements       []string            `json:"requirements,omitempty"`
	McpServers         map[string][]string `json:"mcp_servers,omitempty"`
	Attestation        *Attestation        `json:"attestation,omitempty"`
	Activation         *Activation         `json:"activation,omitempty"`
	Requirers          []string            `json:"requirers,omitempty"`
	Substituted        string              `json:"substituted,omitempty"`
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
	required := []string{
		"schema_version", "name", "source", "ref_kind", "ref", "commit", "content_sha256", "locale",
		"agents", "commands", "dependencies", "skill_schema_version", "runtime_roots", "installed_at", "files",
	}
	for _, field := range required {
		if _, present := raw[field]; !present {
			return false
		}
	}
	if m.SchemaVersion != SchemaVersion || !identifiers.Valid(m.Name) || !identifiers.PortablePath(m.Source) ||
		(m.RefKind != "tag" && m.RefKind != "branch" && m.RefKind != "revision") ||
		m.Ref == "" || utf8.RuneCountInString(m.Ref) > 8192 || !markerCommitRE.MatchString(m.Commit) ||
		!markerSHA256RE.MatchString(m.ContentSHA256) || m.SkillSchemaVersion < 0 || m.SkillSchemaVersion > 5 {
		return false
	}
	if !validNullableLocale(raw["locale"], m.Locale) || !validTimestamp(m.InstalledAt) ||
		!validIdentifierSet(m.Agents) || !validIdentifierSet(m.Commands) || !validIdentifierSet(m.Dependencies) ||
		!validPathSet(m.RuntimeRoots) || !validPathSet(m.Files) {
		return false
	}
	if !validOptionalNonEmptyString(raw, "git", m.Git) || !validOptionalNonEmptyString(raw, "substituted", m.Substituted) {
		return false
	}
	if _, present := raw["requirements"]; present && !validIdentifierSet(m.Requirements) {
		return false
	}
	if _, present := raw["requirers"]; present && !validStringSet(m.Requirers) {
		return false
	}
	if _, present := raw["mcp_servers"]; present {
		if m.McpServers == nil {
			return false
		}
		for name, consumers := range m.McpServers {
			if !identifiers.Valid(name) || !validIdentifierSet(consumers) {
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
			!validBooleanRaw(object["context"]) || !validIdentifierSet(m.Activation.Commands) {
			return false
		}
	}
	return true
}

func validNullableLocale(raw json.RawMessage, decoded string) bool {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return true
	}
	var value string
	return json.Unmarshal(raw, &value) == nil && value == decoded && value != "" && utf8.RuneCountInString(value) <= 64
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

func validIdentifierSet(values []string) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if !identifiers.Valid(value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func validPathSet(values []string) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if !identifiers.PortablePath(value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func validStringSet(values []string) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
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

// Write stores the marker inside dir with sorted keys and a trailing newline.
func Write(dir string, m *Marker) error {
	m.SchemaVersion = SchemaVersion
	m.Agents = nonNilStrings(m.Agents)
	m.Commands = nonNilStrings(m.Commands)
	m.Dependencies = nonNilStrings(m.Dependencies)
	m.RuntimeRoots = nonNilStrings(m.RuntimeRoots)
	m.Files = nonNilStrings(m.Files)
	sort.Strings(m.Agents)
	sort.Strings(m.Commands)
	sort.Strings(m.Dependencies)
	sort.Strings(m.RuntimeRoots)
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
	return os.WriteFile(filepath.Join(dir, Name), append(payload, '\n'), 0o644)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// Current reports whether the installed directory is up to date for the
// expected marker (Spec §8.5): ref kind and value, commit, locale, agents,
// activation, substitution, MCP findings, and attestation must match, and
// the content hash recomputed from disk must equal the recorded one. An
// unsupported marker schema is an error.
func Current(installedDir string, expected *Marker) (bool, error) {
	if version, ok := markerSchemaVersion(installedDir); ok && version != SchemaVersion {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
	}
	recorded := Read(installedDir)
	if recorded == nil {
		return false, nil
	}
	if recorded.SchemaVersion != SchemaVersion {
		return false, fmt.Errorf("unsupported installed marker schema in %s", filepath.Join(installedDir, Name))
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
	actual, err := hashing.ContentSHA256(installedDir, nil)
	if err != nil {
		return false, err
	}
	return recorded.ContentSHA256 == actual, nil
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
