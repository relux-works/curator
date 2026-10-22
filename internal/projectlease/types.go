// Package projectlease owns stable project homes and durable session leases.
// Its authority interfaces are manager-side integration boundaries, never
// repository input or launcher assertions. See protocol/project-context §6.
package projectlease

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/protocoljson"
)

var ErrRevoked = errors.New("project authority revoked")

// ErrArtifactsUnavailable means admission lost a referenced artifact before
// committing the lease. Callers must not publish or start this reservation.
var ErrArtifactsUnavailable = errors.New("project_artifacts_unavailable")

// ErrPendingRecovery refuses collection until durable transaction intent is reconciled.
var ErrPendingRecovery = errors.New("pending project lease transaction; recover before collection")

const Busy = "project_view_busy"
const maxRecord = 1 << 20

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var hexPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Binding is the closed, immutable project-lease/1 carrier. State and process
// observations are deliberately excluded from these identity bytes.
type Binding struct {
	Kind          string `json:"kind"`
	LeaseID       string `json:"lease_id"`
	ContextRef    string `json:"context_ref"`
	ContextSHA256 string `json:"context_sha256"`
	CheckoutID    string `json:"checkout_id"`
	Environment   string `json:"environment"`
	HomeRef       string `json:"home_ref"`
	OwnerRef      string `json:"owner_ref"`
	SessionRef    string `json:"session_ref"`
	Generation    int64  `json:"generation"`
}

// Surface is the protected receipt projection for one generated regular file.
// Mode is the numeric value of 0600 or 0700; native mutable state is excluded.
type Surface struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}

// Composition is a verified protected projection, including its CCJ-1 digest
// and registration binding. Cache keys name the existing GC namespaces.
type Composition struct {
	Ref             string             `json:"ref"`
	SHA256          string             `json:"sha256"`
	CheckoutID      string             `json:"checkout_id"`
	Profile         string             `json:"profile"`
	Environment     string             `json:"environment"`
	HomeRef         string             `json:"home_ref"`
	ConcurrentState bool               `json:"concurrent_state"`
	BuildKeys       []string           `json:"build_keys"`
	RuntimeEntries  []string           `json:"runtime_entries"`
	Generated       map[string]Surface `json:"generated"`
}

// Authority resolves only issuing-manager protected allocations and checks
// current admission/revocation. Implementations must not fetch, repair or
// accept caller-supplied records. Missing authority always refuses a launch.
type Authority interface {
	Resolve(context.Context, managerlock.Held, string, string) (Composition, error)
	Authorize(context.Context, managerlock.Held, Composition, string) error
}

// Evidence is supplied by the tracked-session authority, not a PID probe.
// Identity combines session generation with a runner/process incarnation.
// Terminated means the session AND descendants ended and ownership reconciled.
// NoChildStarted proves a reserved session never started a provider or child.
type Evidence struct {
	SessionRef            string
	Generation            int64
	OwnerRef              string
	ProcessGeneration     string
	Running               bool
	Terminated            bool
	OwnershipReconciled   bool
	DescendantsTerminated bool
	NoChildStarted        bool
}

type Sessions interface {
	Observe(context.Context, string, int64) (Evidence, error)
}

type Request struct {
	LeaseID       string
	ContextRef    string
	ContextSHA256 string
	OwnerRef      string
	SessionRef    string
	Generation    int64
}

type Phase string

const (
	Prepared Phase = "prepared"
	Active   Phase = "active"
	Unknown  Phase = "unknown"
	Released Phase = "released"
)

// Status never repairs or reconciles. Authority/read errors remain visible.
type Status struct {
	Binding          Binding `json:"binding"`
	State            Phase   `json:"state"`
	Busy             bool    `json:"busy"`
	Liveness         string  `json:"liveness"`
	Revoked          bool    `json:"revoked"`
	AuthorityUnknown bool    `json:"authority_unknown"`
	Home             string  `json:"home"`
}

// Roots is a conservative protected-state mark set. Released records are also
// retained for resume, so collection never silently discards their provenance.
type Roots struct {
	Homes          []string
	Compositions   []string
	BuildKeys      []string
	RuntimeEntries []string
}

func tagged(ref, tag string) bool {
	return strings.HasPrefix(ref, tag+":") && len(ref) <= 96 && idPattern.MatchString(strings.TrimPrefix(ref, tag+":"))
}
func validEnvironment(s string) bool { return s == "claude_code" || s == "codex_cli" || s == "muse" }
func (b Binding) validate() error {
	if b.Kind != "project-lease/1" || !idPattern.MatchString(b.LeaseID) || !tagged(b.ContextRef, "composition") || !hexPattern.MatchString(b.ContextSHA256) || !idPattern.MatchString(b.CheckoutID) || !validEnvironment(b.Environment) || !tagged(b.HomeRef, "home") || !tagged(b.OwnerRef, "principal") || !tagged(b.SessionRef, "session") || b.Generation < 1 || b.Generation > protocoljson.MaxSafeInteger {
		return fmt.Errorf("invalid project-lease/1 binding")
	}
	return nil
}
func (c Composition) validate() error {
	if !tagged(c.Ref, "composition") || !hexPattern.MatchString(c.SHA256) || !idPattern.MatchString(c.CheckoutID) || !identifiers.Valid(c.Profile) || !validEnvironment(c.Environment) || !tagged(c.HomeRef, "home") || c.BuildKeys == nil || c.RuntimeEntries == nil || c.Generated == nil {
		return fmt.Errorf("invalid protected composition projection")
	}
	aliases := map[string]bool{}
	for p, file := range c.Generated {
		if !safeSurface(p) || !hexPattern.MatchString(file.SHA256) || file.Size < 0 || file.Size > 16<<20 || (file.Mode != 0600 && file.Mode != 0700) {
			return fmt.Errorf("invalid protected generated surface")
		}
		key := foldPath(p)
		if aliases[key] {
			return fmt.Errorf("generated surface path alias")
		}
		aliases[key] = true
	}
	for path := range aliases {
		parts := strings.Split(path, "/")
		for i := 1; i < len(parts); i++ {
			if aliases[strings.Join(parts[:i], "/")] {
				return fmt.Errorf("generated surface file/ancestor collision")
			}
		}
	}

	total := int64(0)
	for _, file := range c.Generated {
		total += file.Size
	}
	if len(c.Generated) > 10000 || total > 64<<20 {
		return fmt.Errorf("generated inventory exceeds limits")
	}
	for _, k := range c.BuildKeys {
		if !strings.HasPrefix(k, "sha256:") || !hexPattern.MatchString(strings.TrimPrefix(k, "sha256:")) {
			return fmt.Errorf("invalid composition build key")
		}
	}
	for _, k := range c.RuntimeEntries {
		parts := strings.Split(k, "/")
		if len(parts) != 2 || !identifiers.Valid(parts[0]) || !identifiers.Valid(parts[1]) {
			return fmt.Errorf("invalid composition runtime entry")
		}
	}
	return nil
}
func homeRelative(c Composition) string {
	return filepath.Join("environments", "project-views", c.CheckoutID, c.Profile, c.Environment)
}
func canonical(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Preserve integer precision in the canonical encoder.
	var object any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err = dec.Decode(&object); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(object)
}
func digest(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }

// ParseBinding reads an exact, bounded CCJ-1 binding before authority lookup.
func ParseBinding(raw []byte) (Binding, error) {
	var b Binding
	if len(raw) > maxRecord {
		return b, fmt.Errorf("project lease binding exceeds record limit")
	}
	if err := protocoljson.UnmarshalCanonical(raw, &b); err != nil {
		return b, err
	}
	// Re-encoding enforces exact member spelling and completeness; the JSON
	// struct decoder alone also accepts case-insensitive field aliases.
	exact, err := canonical(b)
	if err != nil || string(exact) != string(raw) {
		return b, fmt.Errorf("project lease binding members must match exactly")
	}
	if err := b.validate(); err != nil {
		return b, err
	}
	return b, nil
}

func foldPath(path string) string {
	return cases.Fold().String(norm.NFC.String(filepath.ToSlash(path)))
}
