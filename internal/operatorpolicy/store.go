package operatorpolicy

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectcontext"
)

// Store owns only private policy/audit/event state. Native homes and profile
// locks are never written, so no home lease is created or bypassed here.
type Store struct{ Home string }
type ledger struct {
	Version    int                      `json:"version"`
	Keys       map[string]string        `json:"operator_keys"`
	Audit      map[string]SignedCommand `json:"audit"`
	Policies   map[string]Policy        `json:"policies"`
	Events     []Event                  `json:"events"`
	Selections map[string]Selection     `json:"selections"`
	// No vendor keys are projected by this implementation. A nonempty journal
	// refuses rather than pretending to understand another writer's ownership.
	Journal []jsonJournal `json:"ownership_journal"`
}
type jsonJournal struct {
	Grant string `json:"grant_id"`
}
type operation struct {
	root  string
	held  *managerlock.HomeLock
	state ledger
}

func (s Store) begin(ctx context.Context, bootstrap bool) (*operation, error) {
	m, e := managerlock.New(s.Home)
	if e != nil {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	var lock *managerlock.HomeLock
	if bootstrap {
		lock, e = m.AcquireHomeOnly(ctx, false)
	} else {
		lock, e = m.AcquireExistingHomeOnly(ctx)
	}
	if e != nil {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	op := &operation{root: filepath.Join(m.Home(), "operator-policy"), held: lock}
	ok := false
	defer func() {
		if !ok {
			_ = lock.Close()
		}
	}()
	if pathboundary.ValidateRoot(m.Home()) != nil {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	info, e := os.Lstat(op.root)
	if errors.Is(e, os.ErrNotExist) && bootstrap {
		if e = privatedir.Make(op.root); e != nil {
			return nil, e
		}
		op.state = ledger{1, map[string]string{}, map[string]SignedCommand{}, map[string]Policy{}, []Event{}, map[string]Selection{}, []jsonJournal{}}
	} else {
		if e != nil || !info.IsDir() || privatedir.Validate(op.root) != nil || pathboundary.Validate(op.root) != nil {
			return nil, refusal("workspace_trust_ledger_unreadable")
		}
		p := filepath.Join(op.root, "state.json")
		if privatedir.ValidateFile(p) != nil {
			return nil, refusal("workspace_trust_ledger_unreadable")
		}
		payload, err := readBounded(op.root, "state.json", maxState)
		if err != nil || decode(payload, &op.state) != nil {
			return nil, refusal("workspace_trust_ledger_unreadable")
		}
		b, _ := canonical(op.state)
		if !bytes.Equal(b, payload) {
			return nil, refusal("workspace_trust_ledger_unreadable")
		}
		if e = op.validate(time.Now().UTC()); e != nil {
			return nil, e
		}
	}
	ok = true
	return op, nil
}
func (op *operation) close() { _ = op.held.Close() }

// entryRevisions counts accepted commands and retained terminal lapses. Lapse
// remains part of the boundary after revocation moves it into event history;
// reads and serialization cannot reset it or count it twice. Existing signed
// audit bytes are unchanged, and no wall clock establishes approval freshness.
func (op *operation) entryRevisions() map[string]uint64 {
	revisions := map[string]uint64{}
	for _, a := range op.state.Audit {
		revisions[a.Command.EntryID]++
	}
	for _, ev := range op.state.Events {
		if ev.State.Status == "lapsed" {
			revisions[ev.Policy.ID]++
		}
		for _, state := range ev.History {
			if state.Status == "lapsed" {
				revisions[ev.Policy.ID]++
			}
		}
	}
	return revisions
}
func (op *operation) validate(now time.Time) error {
	s := op.state
	if s.Version != 1 || s.Keys == nil || s.Audit == nil || s.Policies == nil || s.Events == nil || s.Selections == nil || s.Journal == nil || len(s.Journal) != 0 || len(s.Keys) != 1 || len(s.Events) > 1024 || len(s.Audit) > 2048 {
		return refusal("workspace_trust_ledger_unreadable")
	}
	for principal, key := range s.Keys {
		pub, err := base64.StdEncoding.DecodeString(key)
		if !idPattern.MatchString(principal) || err != nil || len(pub) != ed25519.PublicKeySize {
			return refusal("workspace_trust_ledger_unreadable")
		}
	}
	for ref, a := range s.Audit {
		if a.Command.Issuer.Approval != ref || authenticate(s.Keys, a, now) != nil {
			return refusal("workspace_trust_policy_authority")
		}
	}
	for id, p := range s.Policies {
		a, ok := s.Audit[p.Issuer.Approval]
		if !ok || a.Command.Policy == nil || id != p.ID {
			return refusal("workspace_trust_policy_authority")
		}
		if hash(p) != hash(*a.Command.Policy) {
			return refusal("workspace_trust_policy_digest")
		}
	}
	seen := map[string]bool{}
	for _, ev := range s.Events {
		if validEvent(ev) != nil || seen[ev.ID] {
			return refusal("workspace_trust_ledger_unreadable")
		}
		seen[ev.ID] = true
		a, ok := s.Audit[ev.Issuer.Approval]
		selection, selected := s.Selections[ev.ID]
		if !ok || a.Command.Policy == nil || a.Command.Selection == nil || !selected || selection != *a.Command.Selection || ev.Policy.SHA256 != hash(*a.Command.Policy) || ev.Policy.ID != a.Command.EntryID || ev.Issuer != a.Command.Issuer {
			return refusal("workspace_trust_policy_authority")
		}
		if ev.Subject.Path != selection.Path || ev.Subject.Checkout != selection.Checkout || ev.Identity.Environment != selection.Environment || ev.Identity.Lane != selection.Lane || !reflect.DeepEqual(ev.Surfaces, Surfaces{a.Command.Policy.ProjectTrust, a.Command.Policy.Hooks}) {
			return refusal("workspace_trust_policy_authority")
		}
		p := a.Command.Policy
		if ev.Time != a.Command.Time || ev.ReviewBy != p.ReviewBy || ev.Reason != a.Command.Reason || ev.Subject.Repository.RootCommit != p.RootCommit || !reflect.DeepEqual(ev.Subject.Destination, p.Destination) {
			return refusal("workspace_trust_policy_digest")
		}
		at, _ := timestamp(ev.State.Time)
		if at.After(now) {
			return refusal("workspace_trust_ledger_unreadable")
		}
		if ev.State.Status == "revoked" {
			by := ev.State.By
			rev, ok := s.Audit[by.Approval]
			if !ok || rev.Command.EntryID != ev.Policy.ID || rev.Command.Issuer != *by || rev.Command.Time != ev.State.Time {
				return refusal("workspace_trust_policy_authority")
			}
		}
	}
	if len(s.Selections) != len(seen) {
		return refusal("workspace_trust_ledger_unreadable")
	}
	return nil
}
func (op *operation) save() error {
	if e := op.held.AssertCurrent(); e != nil {
		return e
	}
	b, e := canonical(op.state)
	if e != nil || len(b) > maxState {
		return refusal("workspace_trust_ledger_unreadable")
	}
	if e = privatedir.Validate(op.root); e != nil {
		return refusal("workspace_trust_ledger_unreadable")
	}
	f, e := privatedir.CreateTemp(op.root, ".state-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer func() { _ = f.Close(); _ = os.Remove(name) }()
	if _, e = f.Write(b); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = op.held.AssertCurrent(); e != nil {
		return e
	}
	if e = os.Rename(name, filepath.Join(op.root, "state.json")); e != nil {
		return e
	}
	d, e := os.Open(op.root)
	if e != nil {
		return e
	}
	defer func() { _ = d.Close() }()
	return d.Sync()
}

// PinOperator is a one-time local operator enrollment command, never invoked
// by package ingestion. Subsequent key replacement is deliberately refused.
// Same-UID tampering remains the explicit pre-CIP-0024 residual limitation.
func (s Store) PinOperator(ctx context.Context, principal, key string) error {
	pub, e := base64.StdEncoding.DecodeString(key)
	if e != nil || len(pub) != ed25519.PublicKeySize || !idPattern.MatchString(principal) {
		return refusal("workspace_trust_policy_authority")
	}
	op, e := s.begin(ctx, true)
	if e != nil {
		return e
	}
	defer op.close()
	if len(op.state.Keys) != 0 {
		if op.state.Keys[principal] == key {
			return nil
		}
		return refusal("workspace_trust_policy_authority")
	}
	op.state.Keys[principal] = key
	return op.save()
}
func subject(home string, held managerlock.Held, s Selection) (projectcontext.PolicySubject, error) {
	if !exactPath(s.Path) || !idPattern.MatchString(s.Lane) || !digestPattern.MatchString(s.Profile.LockSHA256) {
		return projectcontext.PolicySubject{}, refusal("workspace_untrusted")
	}
	if userHome, err := os.UserHomeDir(); err != nil || s.Path == userHome {
		return projectcontext.PolicySubject{}, refusal("workspace_trust_path_invalid")
	}
	return (projectcontext.Store{Home: home}).ReadPolicySubject(held, s.Checkout, s.Path, s.Environment, s.Profile)
}
func revoke(events []Event, id string, issuer Issuer, at, reason string) error {
	t, _ := timestamp(at)
	for i := range events {
		ev := &events[i]
		if ev.Policy.ID != id || ev.State.Status == "revoked" {
			continue
		}
		last, _ := timestamp(ev.State.Time)
		if t.Before(last) {
			return refusal("workspace_trust_policy_authority")
		}
		ev.History = append(ev.History, ev.State)
		ev.State = State{"revoked", at, reason, &issuer}
	}
	return nil
}

// Apply authenticates both signatures before creating or editing consent.
// An entry edit durably revokes every old derived event in the same atomic
// state write; no native cleanup or provider execution occurs.
func (s Store) Apply(ctx context.Context, a SignedCommand) (*Event, error) {
	op, e := s.begin(ctx, false)
	if e != nil {
		return nil, e
	}
	defer op.close()
	now := time.Now().UTC()
	if e = authenticate(op.state.Keys, a, now); e != nil {
		return nil, e
	}
	c := a.Command
	if old, ok := op.state.Audit[c.Issuer.Approval]; ok {
		if !reflect.DeepEqual(old, a) {
			return nil, refusal("workspace_trust_policy_authority")
		}
		for _, ev := range op.state.Events {
			if ev.Issuer.Approval == c.Issuer.Approval {
				if ev.State.Status != "active" {
					return nil, refusal("workspace_trust_grant_revoked")
				}
				return &ev, nil
			}
		}
		if c.Operation == "revoke" {
			return nil, nil
		}
		return nil, refusal("workspace_trust_policy_authority")
	}
	if c.EntryRevision != op.entryRevisions()[c.EntryID] {
		return nil, refusalDetail("workspace_trust_policy_authority", "signed entry revision is not current")
	}
	if c.Operation == "revoke" {
		if e = revoke(op.state.Events, c.EntryID, c.Issuer, c.Time, c.Reason); e != nil {
			return nil, e
		}
		delete(op.state.Policies, c.EntryID)
		op.state.Audit[c.Issuer.Approval] = a
		return nil, op.save()
	}
	p := *c.Policy
	sel := *c.Selection
	for _, existing := range op.state.Events {
		if existing.State.Status == "active" && existing.Policy.ID != c.EntryID && op.state.Selections[existing.ID] == sel {
			return nil, refusalDetail("workspace_trust_policy_authority", "ambiguous active entries for selection")
		}
	}
	until, _ := timestamp(p.ReviewBy)
	if !now.Before(until) || !has(p.Environments, sel.Environment) || !has(p.Lanes, sel.Lane) {
		return nil, refusal("workspace_untrusted")
	}
	observed, e := subject(s.Home, op.held.Held(), sel)
	if e != nil {
		return nil, e
	}
	if observed.RootCommit != p.RootCommit {
		return nil, refusal("workspace_trust_repository_changed")
	}
	if observed.PrimaryRemote != p.Destination.remote() {
		return nil, refusal("workspace_trust_destination_mismatch")
	}
	if _, ok := op.state.Policies[p.ID]; ok {
		if e = revoke(op.state.Events, p.ID, c.Issuer, c.Time, "Policy entry replaced"); e != nil {
			return nil, e
		}
	}
	ev := Event{"workspace-trust-policy-grant-v1", "policy-grant", "grant:" + hash(c), PolicyRef{p.ID, hash(p)}, Subject{sel.Checkout, observed.RegistrationSHA256, sel.Path, Repository{observed.RepositoryRoot, observed.PrimaryRemote, observed.RootCommit}, p.Destination}, Identity{observed.Registration.RootIdentity, sel.Environment, sel.Lane}, Surfaces{p.ProjectTrust, p.Hooks}, c.Issuer, c.Time, p.ReviewBy, c.Reason, State{"active", c.Time, c.Reason, &c.Issuer}, []State{}}
	if e = validEvent(ev); e != nil {
		return nil, e
	}
	op.state.Policies[p.ID] = p
	op.state.Audit[c.Issuer.Approval] = a
	op.state.Events = append(op.state.Events, ev)
	op.state.Selections[ev.ID] = sel
	if e = op.validate(now); e != nil {
		return nil, e
	}
	if e = op.save(); e != nil {
		return nil, e
	}
	return &ev, nil
}
