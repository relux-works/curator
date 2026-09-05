// Package contextlock is the profile lock of the agent-environments
// capability, context-lock-v1 (environments §1.3): the strict schema-1
// object naming the root and every closure member, its canonical CCJ-1 bytes,
// and the lock hash that is the profile's effective pin.
package contextlock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/pkgversion"
	"github.com/relux-works/curator/internal/protocoljson"
)

// SchemaVersion is the only lock schema this release reads or writes.
const SchemaVersion = 1

// Member kinds, in their bytewise order.
const (
	KindContext = "context"
	KindMCP     = "mcp"
	KindSkill   = "skill"
)

// Member is one closure member.
type Member struct {
	Kind      string
	Name      string
	Source    string // canonical network identity; empty for a path or local package
	Directory string
	Version   string // empty for a skill with no version tag peeling to its commit
	Commit    string
	StateHash string // 64 lowercase hex; the state_sha256 pin
	Weight    int64
	// RequiredBy is the sorted list of direct requirers; empty for the root.
	RequiredBy []string
	Overlay    bool
}

// Pin renders the member's pin in the generation-header grammar:
// "commit <hex>" or "state sha256:<hex>".
func (m Member) Pin() string {
	if m.Commit != "" {
		return "commit " + m.Commit
	}
	return "state sha256:" + m.StateHash
}

// PinKey is the bare pin as machine configuration and the store key spell it:
// the commit, or the 64-hex state hash.
func (m Member) PinKey() string {
	if m.Commit != "" {
		return m.Commit
	}
	return m.StateHash
}

// Lock is one context-lock-v1 object.
type Lock struct {
	Root    string
	Members []Member
}

// Find returns the member of the given kind and name.
func (lock *Lock) Find(kind, name string) (Member, bool) {
	for _, member := range lock.Members {
		if member.Kind == kind && member.Name == name {
			return member, true
		}
	}
	return Member{}, false
}

// Contexts returns the context members in lock order.
func (lock *Lock) Contexts() []Member {
	var out []Member
	for _, member := range lock.Members {
		if member.Kind == KindContext {
			out = append(out, member)
		}
	}
	return out
}

// RootMember returns the root context member.
func (lock *Lock) RootMember() (Member, bool) { return lock.Find(KindContext, lock.Root) }

// Sort orders the members by (kind, name), bytewise, and each required_by
// list bytewise.
func (lock *Lock) Sort() {
	for index := range lock.Members {
		sort.Strings(lock.Members[index].RequiredBy)
	}
	sort.SliceStable(lock.Members, func(i, j int) bool {
		a, b := lock.Members[i], lock.Members[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
}

var (
	commitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
	hex256RE = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Validate applies the context-lock-v1 schema rules.
func (lock *Lock) Validate() error {
	if !identifiers.Valid(lock.Root) {
		return fmt.Errorf("lock root %q is not a portable identifier", lock.Root)
	}
	if len(lock.Members) == 0 {
		return fmt.Errorf("lock has no members")
	}
	seen := map[string]bool{}
	rootSeen := false
	for index, member := range lock.Members {
		switch member.Kind {
		case KindContext, KindMCP, KindSkill:
		default:
			return fmt.Errorf("member %d has unknown kind %q", index, member.Kind)
		}
		if !identifiers.Valid(member.Name) {
			return fmt.Errorf("member %d name %q is not a portable identifier", index, member.Name)
		}
		key := member.Kind + "\x00" + member.Name
		if seen[key] {
			return fmt.Errorf("member %s %s appears twice", member.Kind, member.Name)
		}
		seen[key] = true
		if member.Kind == KindContext && member.Name == lock.Root {
			rootSeen = true
		}
		if index > 0 {
			previous := lock.Members[index-1]
			if previous.Kind > member.Kind || (previous.Kind == member.Kind && previous.Name >= member.Name) {
				return fmt.Errorf("members are not sorted by (kind, name) at %s %s", member.Kind, member.Name)
			}
		}
		switch {
		case member.Commit != "" && member.StateHash != "":
			return fmt.Errorf("member %s carries both commit and state_sha256", member.Name)
		case member.Commit != "":
			if !commitRE.MatchString(member.Commit) {
				return fmt.Errorf("member %s commit %q is invalid", member.Name, member.Commit)
			}
			if member.Source == "" {
				return fmt.Errorf("member %s pins a commit without a source", member.Name)
			}
		case member.StateHash != "":
			if !hex256RE.MatchString(member.StateHash) {
				return fmt.Errorf("member %s state_sha256 %q is invalid", member.Name, member.StateHash)
			}
			if member.Kind != KindContext || member.Source != "" || member.Directory != "" {
				return fmt.Errorf("member %s: a state pin admits only a context member without source or directory", member.Name)
			}
		default:
			return fmt.Errorf("member %s carries no pin", member.Name)
		}
		if member.Directory != "" && !identifiers.PortablePath(member.Directory) {
			return fmt.Errorf("member %s directory %q is not a portable path", member.Name, member.Directory)
		}
		if member.Kind == KindSkill {
			if member.Directory != "" {
				return fmt.Errorf("skill member %s carries a directory", member.Name)
			}
		} else if member.Version == "" {
			return fmt.Errorf("member %s carries no version", member.Name)
		}
		if member.Version != "" {
			if _, err := pkgversion.ParseVersion(member.Version); err != nil {
				return fmt.Errorf("member %s version %q: %v", member.Name, member.Version, err)
			}
		}
		if member.Weight < 0 || member.Weight > 2147483647 {
			return fmt.Errorf("member %s weight %d is out of range", member.Name, member.Weight)
		}
		if member.Overlay && member.Kind != KindContext {
			return fmt.Errorf("member %s: only a context member may be an overlay", member.Name)
		}
		for i, requirer := range member.RequiredBy {
			if !identifiers.Valid(requirer) {
				return fmt.Errorf("member %s required_by %q is not a portable identifier", member.Name, requirer)
			}
			if i > 0 && member.RequiredBy[i-1] >= requirer {
				return fmt.Errorf("member %s required_by is not sorted and unique", member.Name)
			}
		}
	}
	if !rootSeen {
		return fmt.Errorf("lock root %s is not a context member", lock.Root)
	}
	return nil
}

// Object renders the lock as the JSON-domain value CCJ-1 encodes.
func (lock *Lock) Object() map[string]any {
	members := make([]any, 0, len(lock.Members))
	for _, member := range lock.Members {
		object := map[string]any{
			"kind":    member.Kind,
			"name":    member.Name,
			"weight":  member.Weight,
			"overlay": member.Overlay,
		}
		requiredBy := make([]any, 0, len(member.RequiredBy))
		for _, requirer := range member.RequiredBy {
			requiredBy = append(requiredBy, requirer)
		}
		object["required_by"] = requiredBy
		if member.Source != "" {
			object["source"] = member.Source
		}
		if member.Directory != "" {
			object["directory"] = member.Directory
		}
		if member.Version != "" {
			object["version"] = member.Version
		}
		if member.Commit != "" {
			object["commit"] = member.Commit
		}
		if member.StateHash != "" {
			object["state_sha256"] = member.StateHash
		}
		members = append(members, object)
	}
	return map[string]any{
		"schema_version": SchemaVersion,
		"root":           lock.Root,
		"members":        members,
	}
}

// Canonical returns the CCJ-1 bytes of the lock after validation.
func (lock *Lock) Canonical() ([]byte, error) {
	if err := lock.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lock: %w", err)
	}
	return protocoljson.MarshalCanonical(lock.Object())
}

// Hash returns the lock hash, "sha256:<64 lowercase hex>" over the CCJ-1 bytes.
func (lock *Lock) Hash() (string, error) {
	canonical, err := lock.Canonical()
	if err != nil {
		return "", err
	}
	return HashBytes(canonical), nil
}

// HashBytes hashes already-canonical lock bytes.
func HashBytes(canonical []byte) string {
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Parse reads a lock from its bytes under the strict reader discipline.
func Parse(payload []byte) (*Lock, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("invalid lock: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var raw struct {
		SchemaVersion json.Number `json:"schema_version"`
		Root          string      `json:"root"`
		Members       []struct {
			Kind       string      `json:"kind"`
			Name       string      `json:"name"`
			Source     string      `json:"source"`
			Directory  string      `json:"directory"`
			Version    string      `json:"version"`
			Commit     string      `json:"commit"`
			StateHash  string      `json:"state_sha256"`
			Weight     json.Number `json:"weight"`
			RequiredBy []string    `json:"required_by"`
			Overlay    *bool       `json:"overlay"`
		} `json:"members"`
	}
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("invalid lock: %w", err)
	}
	if raw.SchemaVersion.String() != strconv.Itoa(SchemaVersion) {
		return nil, fmt.Errorf("invalid lock: schema_version must be %d", SchemaVersion)
	}
	lock := &Lock{Root: raw.Root}
	for _, member := range raw.Members {
		weight, err := strconv.ParseInt(member.Weight.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid lock: member %s weight: %v", member.Name, err)
		}
		if member.RequiredBy == nil || member.Overlay == nil {
			return nil, fmt.Errorf("invalid lock: member %s lacks required_by or overlay", member.Name)
		}
		lock.Members = append(lock.Members, Member{
			Kind: member.Kind, Name: member.Name, Source: member.Source, Directory: member.Directory,
			Version: member.Version, Commit: member.Commit, StateHash: member.StateHash, Weight: weight,
			RequiredBy: member.RequiredBy, Overlay: *member.Overlay,
		})
	}
	if err := lock.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lock: %w", err)
	}
	return lock, nil
}

// Read loads a lock file. The file holds exactly the CCJ-1 bytes, so the
// bytes on disk hash to the lock hash.
func Read(path string) (*Lock, string, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- manager-home path
	if err != nil {
		return nil, "", err
	}
	if err := protocoljson.RequireCanonical(payload); err != nil {
		return nil, "", fmt.Errorf("lock %s: %w", path, err)
	}
	lock, err := Parse(payload)
	if err != nil {
		return nil, "", fmt.Errorf("lock %s: %w", path, err)
	}
	return lock, HashBytes(payload), nil
}

// Write stores the lock's canonical bytes at path atomically and returns the
// lock hash.
func Write(path string, lock *Lock) (string, error) {
	canonical, err := lock.Canonical()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".lock-*")
	if err != nil {
		return "", err
	}
	if _, err := tmp.Write(canonical); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	return HashBytes(canonical), nil
}
