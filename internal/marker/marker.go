// Package marker reads and writes install markers (.csk-install.json) and
// implements the up-to-date and tamper-detection semantics of Spec §8.5.
package marker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/relux-works/curator/internal/hashing"
)

// Name is the marker file name inside an installed skill directory.
const Name = ".csk-install.json"

// SchemaVersion is the supported marker schema.
const SchemaVersion = 1

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

// Read loads the marker of an installed directory; nil when absent or
// unreadable (an unreadable marker simply means "not current").
func Read(installedDir string) *Marker {
	payload, err := os.ReadFile(filepath.Join(installedDir, Name)) // #nosec G304 -- path derives from the install root
	if err != nil {
		return nil
	}
	var m Marker
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil
	}
	return &m
}

// Write stores the marker inside dir with sorted keys and a trailing newline.
func Write(dir string, m *Marker) error {
	m.SchemaVersion = SchemaVersion
	sort.Strings(m.Commands)
	sort.Strings(m.Dependencies)
	if m.Requirements != nil {
		sort.Strings(m.Requirements)
	}
	for name := range m.McpServers {
		sort.Strings(m.McpServers[name])
	}
	payload, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, Name), append(payload, '\n'), 0o644)
}

// Current reports whether the installed directory is up to date for the
// expected marker (Spec §8.5): ref kind and value, commit, locale, agents,
// activation, substitution, MCP findings, and attestation must match, and
// the content hash recomputed from disk must equal the recorded one. An
// unsupported marker schema is an error.
func Current(installedDir string, expected *Marker) (bool, error) {
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
