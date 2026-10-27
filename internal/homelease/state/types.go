// Package state owns the closed home-lease projection and read-only admission
// policy. It has no launcher, journal or lock-acquisition dependency: all writers,
// including transaction recovery, can apply the same policy without a cycle.
package state

import (
	"fmt"
	"strings"
)

type Process struct {
	PID       int64  `json:"pid"`
	StartedAt string `json:"started_at"`
}

type Holder struct {
	Kind                string   `json:"kind"`
	Launcher            string   `json:"launcher"`
	ManagerID           string   `json:"manager_id"`
	ManagerRegisteredAt string   `json:"manager_registered_at"`
	RequestID           string   `json:"request_id"`
	Werkstatt           *string  `json:"werkstatt"`
	Session             *string  `json:"session"`
	Generation          *int64   `json:"generation"`
	Process             *Process `json:"process"`
}

type Observation struct {
	State      string  `json:"state"`
	ObservedAt *string `json:"observed_at"`
}

type Record struct {
	Kind          string      `json:"kind"`
	LeaseID       string      `json:"lease_id"`
	ManagerRootID string      `json:"manager_root_id"`
	HomeID        string      `json:"home_id"`
	Environment   string      `json:"environment"`
	Profile       string      `json:"profile"`
	LockSHA256    string      `json:"lock_sha256"`
	Revision      int64       `json:"revision"`
	Phase         string      `json:"phase"`
	AcquiredAt    string      `json:"acquired_at"`
	Holder        Holder      `json:"holder"`
	Observation   Observation `json:"observation"`
}

func (r Record) Blocking() bool {
	return r.Phase == "prepared" || r.Phase == "active" || r.Phase == "unknown"
}

type Event struct {
	Kind        string `json:"kind"`
	EventID     string `json:"event_id"`
	Operation   string `json:"operation"`
	LeaseID     string `json:"lease_id"`
	Previous    Record `json:"previous"`
	NewRevision int64  `json:"new_revision"`
	Reason      string `json:"reason"`
	RecordedAt  string `json:"recorded_at"`
}

type Details struct {
	LeaseID  *string  `json:"lease_id"`
	Revision *int64   `json:"revision"`
	Holders  []Record `json:"holders"`
	Count    int      `json:"count"`
}

// Refusal carries only closed, manager-authored diagnostics. Invalid holder
// bytes and observer output never become diagnostic text.
type Refusal struct {
	Code    string
	Reason  string
	Details Details
}

func Refuse(code, reason string) *Refusal {
	return &Refusal{Code: code, Reason: reason, Details: Details{Holders: []Record{}}}
}

func (e *Refusal) Error() string {
	var holders []string
	for _, r := range e.Details.Holders {
		holders = append(holders, fmt.Sprintf("%s revision %d %s %s/%s manager %s registered %s session %v", r.LeaseID, r.Revision, r.Phase, r.Environment, r.Profile, r.Holder.ManagerID, r.Holder.ManagerRegisteredAt, sessionName(r.Holder)))
	}
	message := e.Code + "/" + e.Reason
	if len(holders) > 0 {
		message += ": " + strings.Join(holders, "; ") + "; inspect curator env lease list; stop and reconcile each holder, or deliberately use an audited operator break"
	}
	return message
}

func sessionName(h Holder) string {
	if h.Session == nil {
		return "unbound"
	}
	return *h.Session
}

// Admission is Curator-authored, never accepted from the holder or CLI.
// Protected includes the actual home, profile authority and shared mutable
// dependencies. Runtime/build references are retained independently of scopes.
type Admission struct {
	Home           string   `json:"home"`
	Protected      []string `json:"protected"`
	BuildKeys      []string `json:"build_keys"`
	RuntimeEntries []string `json:"runtime_entries"`
}

type Entry struct {
	Record        Record    `json:"record"`
	Initial       Record    `json:"initial"`
	PayloadSHA256 string    `json:"payload_sha256"`
	Admission     Admission `json:"admission"`
	BindRevision  *int64    `json:"bind_revision"`
	TerminalEvent *string   `json:"terminal_event"`
}

type Ledger struct {
	Kind     string            `json:"kind"`
	Leases   map[string]Entry  `json:"leases"`
	Requests map[string]string `json:"requests"`
}

func Empty() *Ledger {
	return &Ledger{Kind: "home-lease-state/1", Leases: map[string]Entry{}, Requests: map[string]string{}}
}
