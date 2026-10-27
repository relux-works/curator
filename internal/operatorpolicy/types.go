// Package operatorpolicy records authenticated CIP-0020 policy grants. It
// neither projects vendor state nor selects permissions, sandbox or memory.
package operatorpolicy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/projectcontext"
	"github.com/relux-works/curator/internal/protocoljson"
)

const maxState = 4 << 20

var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func refusal(code string) error               { return fmt.Errorf("%s: operator policy admission refused", code) }
func refusalDetail(code, detail string) error { return fmt.Errorf("%w (%s)", refusal(code), detail) }
func canonical(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var x any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e = d.Decode(&x); e != nil {
		return nil, e
	}
	return protocoljson.MarshalCanonical(x)
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func hash(v any) string      { b, _ := canonical(v); return digest(b) }
func decode(b []byte, v any) error {
	if len(b) > maxState || protocoljson.Validate(b) != nil {
		return refusal("workspace_trust_ledger_unreadable")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return refusal("workspace_trust_ledger_unreadable")
	}
	// Go's struct decoder accepts case aliases and absent required zero values.
	// Compare JSON-domain objects to enforce the closed spelling/field set.
	var raw any
	r := json.NewDecoder(bytes.NewReader(b))
	r.UseNumber()
	if r.Decode(&raw) != nil {
		return refusal("workspace_trust_ledger_unreadable")
	}
	original, err := protocoljson.MarshalCanonical(raw)
	typed, typedErr := canonical(v)
	if err != nil || typedErr != nil || !bytes.Equal(original, typed) {
		return refusal("workspace_trust_ledger_unreadable")
	}
	return nil
}

// ParseSignedCommand rejects duplicate/unknown members before authentication.
func ParseSignedCommand(b []byte) (SignedCommand, error) {
	var s SignedCommand
	if e := decode(b, &s); e != nil {
		return s, refusal("workspace_trust_policy_authority")
	}
	return s, nil
}
func timestamp(s string) (time.Time, error) {
	t, e := time.Parse(time.RFC3339, s)
	if e != nil || !strings.HasSuffix(s, "Z") {
		return t, refusal("workspace_trust_ledger_unreadable")
	}
	return t, nil
}

// Hook binds the complete effective definition, never just its executable.
type Hook struct {
	ID     string `json:"hook_id"`
	SHA256 string `json:"sha256"`
}

// Surfaces is deliberately limited to trust and exact hook digests.
type Surfaces struct {
	ProjectTrust bool   `json:"project_trust"`
	Hooks        []Hook `json:"hooks"`
}

// Actor identifies the actual command signer, separately from its approver.
type Actor struct {
	Kind      string `json:"kind"`
	Principal string `json:"principal_ref"`
}

// Issuer is provenance; signatures against enrolled keys establish authority.
type Issuer struct {
	Role      string `json:"role"`
	Principal string `json:"principal_ref"`
	Approval  string `json:"approval_ref"`
	Actor     Actor  `json:"actor"`
}

// Destination is the closed direct canonical-host tuple from CIP-0020.
type Destination struct {
	Transport string   `json:"transport"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	Path      string   `json:"repository_path"`
	User      string   `json:"user,omitempty"`
	Pins      []string `json:"host_key_fingerprints,omitempty"`
}

// Policy is the complete approved entry. Every field participates in CCJ-1 hashing.
type Policy struct {
	ID           string      `json:"entry_id"`
	RootCommit   string      `json:"root_commit"`
	Destination  Destination `json:"canonical_destination"`
	Spellings    []string    `json:"remote_spellings"`
	Environments []string    `json:"environments"`
	Lanes        []string    `json:"lanes"`
	ProjectTrust bool        `json:"project_trust"`
	Hooks        []Hook      `json:"hooks"`
	Issuer       Issuer      `json:"issuer"`
	Time         string      `json:"time"`
	Reason       string      `json:"reason"`
	ReviewBy     string      `json:"review_by"`
}

// Selection binds a command and read to one exact protected registration/lock.
type Selection struct {
	Checkout    string                 `json:"checkout_id"`
	Path        string                 `json:"path"`
	Environment string                 `json:"environment"`
	Lane        string                 `json:"lane_id"`
	Profile     projectcontext.Profile `json:"profile"`
}

// Command is signed in its entirety by the approving operator AND the actor.
// Approval is an immutable opaque digest reference, not a claimed capability.
type Command struct {
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Issuer    Issuer `json:"issuer"`
	Time      string `json:"time"`
	Reason    string `json:"reason"`
	EntryID   string `json:"entry_id"`
	// EntryRevision counts retained successful commands and lapses. Zero
	// is omitted for first commands and compatibility with existing audit bytes.
	EntryRevision  uint64     `json:"entry_revision,omitempty"`
	Policy         *Policy    `json:"policy,omitempty"`
	Selection      *Selection `json:"selection,omitempty"`
	ActorPublicKey string     `json:"actor_public_key"`
}

// SignedCommand contains base64 Ed25519 signatures over canonical Command bytes.
type SignedCommand struct {
	Command           Command `json:"command"`
	ApprovalSignature string  `json:"approval_signature"`
	ActorSignature    string  `json:"actor_signature"`
}

// PolicyRef binds the complete approved entry.
type PolicyRef struct {
	ID     string `json:"entry_id"`
	SHA256 string `json:"sha256"`
}

// Repository is the literal repository identity, not a moving HEAD.
type Repository struct {
	Root       string `json:"root"`
	Remote     string `json:"primary_remote"`
	RootCommit string `json:"root_commit"`
}

// Subject binds the retained registration and canonical destination.
type Subject struct {
	Checkout           string      `json:"checkout_id"`
	RegistrationSHA256 string      `json:"registration_sha256"`
	Path               string      `json:"path"`
	Repository         Repository  `json:"repository"`
	Destination        Destination `json:"canonical_destination"`
}

// Identity scopes one grant to one environment and concrete lane.
type Identity struct {
	Root        string `json:"root_identity"`
	Environment string `json:"environment"`
	Lane        string `json:"lane_id"`
}

// State retains revocation and lapse provenance; lapses have no issuer.
type State struct {
	Status string  `json:"status"`
	Time   string  `json:"time"`
	Reason string  `json:"reason"`
	By     *Issuer `json:"by"`
}

// Event is exactly workspace-trust-policy-grant-v1, never an A1 explicit grant.
type Event struct {
	Record   string    `json:"record"`
	Event    string    `json:"event"`
	ID       string    `json:"grant_id"`
	Policy   PolicyRef `json:"policy"`
	Subject  Subject   `json:"subject"`
	Identity Identity  `json:"identity"`
	Surfaces Surfaces  `json:"surfaces"`
	Issuer   Issuer    `json:"issuer"`
	Time     string    `json:"time"`
	ReviewBy string    `json:"review_by"`
	Reason   string    `json:"reason"`
	State    State     `json:"state"`
	History  []State   `json:"state_history"`
}

// Delivery is an internal per-start read result, not the proposed #197 carrier.
// Callers must consume it immediately; retaining it never extends consent.
type Delivery struct {
	GrantID   string    `json:"grant_id"`
	Policy    PolicyRef `json:"policy"`
	Selection Selection `json:"selection"`
	Surfaces  Surfaces  `json:"surfaces"`
}
