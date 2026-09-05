// Package envmarker reads and writes the environment marker of environments
// §8.2, agent-environment-marker-v1: the per-home ledger of record for
// environment surfaces (.agent-environment.json). Readers reject an
// unsupported version and unknown fields; an unreadable or invalid marker
// fails closed and is reported as environment_marker_invalid.
package envmarker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Name is the marker file name beside the managed surfaces.
const Name = ".agent-environment.json"

// Version is the only marker version this release reads or writes.
const Version = 1

// Modes (environments §8.1).
const (
	ModeManagedHome = "managed-home"
	ModeLinked      = "linked"
	ModeCopied      = "copied"
)

// Surface keys (environments §8.2).
const (
	SurfaceRootContext  = "root-context"
	SurfaceSkills       = "skills"
	SurfaceSystemPrompt = "system-prompt"
	SurfaceMCP          = "mcp"
)

// Copy reasons.
const (
	ReasonSymlinkFallback       = "symlink-fallback"
	ReasonClaudeCodeRootContext = "claude-code-root-context"
)

// DiagMarkerInvalid is the fail-closed diagnostic for an unreadable,
// malformed, or unsupported marker.
const DiagMarkerInvalid = "environment_marker_invalid"

// Requirement is the declared requirement of a git root, as written.
type Requirement struct {
	Range    string `json:"range,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// Profile identifies the materialized profile.
type Profile struct {
	Name               string       `json:"name"`
	Root               string       `json:"root"`
	Kind               string       `json:"kind"`
	LockSHA256         string       `json:"lock_sha256"`
	Source             string       `json:"source,omitempty"`
	Requirement        *Requirement `json:"requirement,omitempty"`
	Directory          string       `json:"directory,omitempty"`
	SourcePath         string       `json:"source_path,omitempty"`
	ImportedFromNative bool         `json:"imported_from_native,omitempty"`
}

// Member is one context member in emitted order.
type Member struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Commit      string `json:"commit,omitempty"`
	StateSHA256 string `json:"state_sha256,omitempty"`
	Weight      int64  `json:"weight"`
	Overlay     bool   `json:"overlay"`
	SourcePath  string `json:"source_path,omitempty"`
}

// Precedence carries both primitives.
type Precedence struct {
	Winner    string `json:"winner"`
	Placement string `json:"placement"`
}

// Copy records one entry materialized as a copy rather than a link.
type Copy struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Surface is one managed surface entry.
type Surface struct {
	Paths         []string `json:"paths"`
	Form          string   `json:"form,omitempty"`
	ContentSHA256 string   `json:"content_sha256"`
	// Copies is present (possibly empty) in linked and managed-home modes and
	// absent in copied mode.
	Copies *[]Copy `json:"copies,omitempty"`
}

// Marker is one agent-environment-marker-v1 object. The managed-home members
// (passthrough, seeds, seeded_projects, seed_links) are declared so a reader
// accepts a managed-home marker; this stage writes in-place markers only.
type Marker struct {
	Version        int                `json:"version"`
	Profile        Profile            `json:"profile"`
	Members        []Member           `json:"members"`
	Precedence     Precedence         `json:"precedence"`
	Mode           string             `json:"mode"`
	Surfaces       map[string]Surface `json:"surfaces"`
	Passthrough    *[]Passthrough     `json:"passthrough,omitempty"`
	Seeds          *[]string          `json:"seeds,omitempty"`
	SeededProjects []string           `json:"seeded_projects,omitempty"`
	SeedLinks      []string           `json:"seed_links,omitempty"`
}

// Passthrough is one recorded credential passthrough entry of a managed home.
type Passthrough struct {
	Path     string `json:"path"`
	Strategy string `json:"strategy"`
}

var (
	hex256RE = regexp.MustCompile(`^[0-9a-f]{64}$`)
	sha256RE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	commitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
)

// Validate applies the schema rules.
func (m *Marker) Validate() error {
	if m.Version != Version {
		return fmt.Errorf("unsupported marker version %d", m.Version)
	}
	if !identifiers.Valid(m.Profile.Name) || !identifiers.Valid(m.Profile.Root) {
		return fmt.Errorf("profile name or root is not a portable identifier")
	}
	if !hex256RE.MatchString(m.Profile.LockSHA256) {
		return fmt.Errorf("profile lock_sha256 is not 64 lowercase hex")
	}
	switch m.Profile.Kind {
	case "git":
		if m.Profile.Source == "" || m.Profile.Requirement == nil {
			return fmt.Errorf("a git profile records its source and requirement")
		}
		forms := 0
		for _, value := range []string{m.Profile.Requirement.Range, m.Profile.Requirement.Tag, m.Profile.Requirement.Revision} {
			if value != "" {
				forms++
			}
		}
		if forms != 1 {
			return fmt.Errorf("a git profile requirement carries exactly one form")
		}
		if m.Profile.SourcePath != "" || m.Profile.ImportedFromNative {
			return fmt.Errorf("a git profile carries no source_path or imported_from_native")
		}
	case "local":
		if m.Profile.Source != "" || m.Profile.Requirement != nil || m.Profile.Directory != "" || m.Profile.SourcePath != "" || m.Profile.ImportedFromNative {
			return fmt.Errorf("a local profile carries only name, root, kind, and lock_sha256")
		}
	case "path":
		if m.Profile.SourcePath == "" || m.Profile.Source != "" || m.Profile.Requirement != nil || m.Profile.Directory != "" {
			return fmt.Errorf("a path profile records source_path and no git members")
		}
	default:
		return fmt.Errorf("profile kind %q is not git, local, or path", m.Profile.Kind)
	}
	if len(m.Members) == 0 {
		return fmt.Errorf("members is empty")
	}
	for _, member := range m.Members {
		if !identifiers.Valid(member.Name) {
			return fmt.Errorf("member %q is not a portable identifier", member.Name)
		}
		switch {
		case member.Commit != "" && member.StateSHA256 == "" && member.SourcePath == "":
			if !commitRE.MatchString(member.Commit) {
				return fmt.Errorf("member %s commit is invalid", member.Name)
			}
		case member.Commit == "" && member.StateSHA256 != "":
			if !hex256RE.MatchString(member.StateSHA256) {
				return fmt.Errorf("member %s state_sha256 is invalid", member.Name)
			}
		default:
			return fmt.Errorf("member %s carries exactly one pin", member.Name)
		}
	}
	if (m.Precedence.Winner != "higher-weight" && m.Precedence.Winner != "lower-weight") ||
		(m.Precedence.Placement != "winner-last" && m.Precedence.Placement != "winner-first") {
		return fmt.Errorf("precedence primitives are invalid")
	}
	switch m.Mode {
	case ModeManagedHome:
		if m.Passthrough == nil || m.Seeds == nil {
			return fmt.Errorf("a managed-home marker records passthrough and seeds")
		}
	case ModeLinked, ModeCopied:
		if m.Passthrough != nil || m.Seeds != nil || m.SeededProjects != nil || m.SeedLinks != nil {
			return fmt.Errorf("an in-place marker carries no managed-home members")
		}
	default:
		return fmt.Errorf("mode %q is not managed-home, linked, or copied", m.Mode)
	}
	if m.Surfaces == nil {
		return fmt.Errorf("surfaces is absent")
	}
	for key, surface := range m.Surfaces {
		switch key {
		case SurfaceRootContext, SurfaceSkills, SurfaceSystemPrompt, SurfaceMCP:
		default:
			return fmt.Errorf("surface key %q is unknown", key)
		}
		if key == SurfaceRootContext {
			if surface.Form != "monolithic" && surface.Form != "referenced" {
				return fmt.Errorf("root-context form %q is invalid", surface.Form)
			}
		} else if surface.Form != "" {
			return fmt.Errorf("surface %s carries a form", key)
		}
		if surface.Paths == nil {
			return fmt.Errorf("surface %s has no paths array", key)
		}
		seen := map[string]bool{}
		for _, path := range surface.Paths {
			if !identifiers.PortablePath(path) || seen[path] {
				return fmt.Errorf("surface %s path %q is not a unique portable path", key, path)
			}
			seen[path] = true
		}
		if !sha256RE.MatchString(surface.ContentSHA256) {
			return fmt.Errorf("surface %s content_sha256 is invalid", key)
		}
		if m.Mode == ModeCopied {
			if surface.Copies != nil {
				return fmt.Errorf("surface %s records copies in copied mode", key)
			}
		} else if surface.Copies == nil {
			return fmt.Errorf("surface %s lacks the copies array", key)
		} else {
			for _, copy := range *surface.Copies {
				if !seen[copy.Path] {
					return fmt.Errorf("surface %s copy %q is not one of its paths", key, copy.Path)
				}
				if copy.Reason != ReasonSymlinkFallback && copy.Reason != ReasonClaudeCodeRootContext {
					return fmt.Errorf("surface %s copy reason %q is unknown", key, copy.Reason)
				}
			}
		}
	}
	return nil
}

// Marshal renders the marker bytes: indented JSON with sorted keys and one
// trailing LF.
func (m *Marker) Marshal() ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// Parse reads marker bytes under the strict reader discipline.
func Parse(payload []byte) (*Marker, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("%s: %w", DiagMarkerInvalid, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var marker Marker
	if err := decoder.Decode(&marker); err != nil {
		return nil, fmt.Errorf("%s: %w", DiagMarkerInvalid, err)
	}
	if err := marker.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", DiagMarkerInvalid, err)
	}
	return &marker, nil
}

// Read loads the marker of a home. It distinguishes absence (nil, nil) from a
// failed or invalid read (nil, error carrying DiagMarkerInvalid).
func Read(home string) (*Marker, error) {
	payload, err := os.ReadFile(filepath.Join(home, Name)) // #nosec G304 -- home chosen by the caller
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", DiagMarkerInvalid, err)
	}
	return Parse(payload)
}

// SortedSurfaceKeys returns the surface keys in bytewise order.
func (m *Marker) SortedSurfaceKeys() []string {
	keys := make([]string, 0, len(m.Surfaces))
	for key := range m.Surfaces {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
