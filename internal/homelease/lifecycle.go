package homelease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/transaction"
)

func (m *Manager) lookup(l *state.Ledger, id string) (state.Entry, bool, error) {
	e, ok := l.Leases[id]
	if ok {
		return e, true, nil
	}
	// The project lifecycle has already validated the complete project ledger.
	// Its immutable binding identifies kind without granting profile authority.
	raw, err := state.Read(m.home, filepath.Join(m.home, "state", "project-leases", "v1", "bindings", id+".json"))
	if err == nil {
		if _, err := projectlease.ParseBinding(raw); err != nil {
			return e, false, state.Refuse("environment_lease_unreadable", "invalid_record")
		}
		return e, false, state.Refuse("environment_lease_authority_refused", "unsupported_lease_kind")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return e, false, state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	return e, false, nil
}

func (m *Manager) Bind(ctx context.Context, id string, revision int64, holder Holder) (out Outcome, reason string, err error) {
	if !state.ValidID(id) {
		return out, "", state.Refuse("usage", "invalid_selector")
	}
	if revision < 1 || revision > protocoljson.MaxSafeInteger {
		return out, "", state.Refuse("usage", "revision_required")
	}
	raw, er := state.Canonical(holder)
	if er != nil || state.Validate("holder", raw) != nil || holder.Kind != "bound" {
		return out, "", state.Refuse("usage", "invalid_holder")
	}
	err = m.withState(ctx, func(l *state.Ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		e, exists, err := m.lookup(l, id)
		if err != nil {
			return err
		}
		if !exists {
			return state.Refuse("environment_lease_unknown", "unknown_lease")
		}
		// Terminal disposition precedes historical successful bind replay.
		if !e.Record.Blocking() {
			return state.Refuse("environment_lease_revision_conflict", "revision_changed")
		}
		if e.BindRevision != nil && *e.BindRevision == revision && reflect.DeepEqual(e.Record.Holder, holder) {
			r := e.Record
			out.Lease = &r
			reason = "bound"
			return nil
		}
		if e.Record.Revision != revision {
			return state.Refuse("environment_lease_revision_conflict", "revision_changed")
		}
		if holder.ManagerID != e.Record.Holder.ManagerID || holder.ManagerRegisteredAt != e.Record.Holder.ManagerRegisteredAt {
			return state.Refuse("environment_lease_holder_mismatch", "manager_incarnation_changed")
		}
		if !sameReservation(e.Record.Holder, holder) {
			return state.Refuse("environment_lease_holder_mismatch", "request_conflict")
		}
		if e.Record.Holder.Kind != "prepared" {
			return state.Refuse("environment_lease_holder_mismatch", "session_generation_changed")
		}
		if m.observer == nil {
			return state.Refuse("environment_lease_authority_refused", "authority_unavailable")
		}
		observed, err := m.observer.Observe(ctx, holder)
		if err != nil {
			return state.Refuse("environment_lease_authority_refused", "authority_unavailable")
		}
		if !reflect.DeepEqual(observed.Holder, holder) {
			return state.Refuse("environment_lease_holder_mismatch", "session_generation_changed")
		}
		e.BindRevision = &revision
		e.Record.Holder = holder
		e.Record.Phase = "active"
		e.Record.Revision++
		at := now()
		e.Record.Observation = state.Observation{State: "not_observed", ObservedAt: &at}
		l.Leases[id] = e
		if err := m.publish(l, lock, engine, map[string][]byte{}); err != nil {
			return err
		}
		r := e.Record
		out = Outcome{Changed: true, Lease: &r}
		reason = "bound"
		return nil
	})
	return
}

func (m *Manager) Release(ctx context.Context, id string) (out Outcome, reason string, err error) {
	if !state.ValidID(id) {
		return out, "", state.Refuse("usage", "invalid_selector")
	}
	err = m.withState(ctx, func(l *state.Ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		e, exists, err := m.lookup(l, id)
		if err != nil {
			return err
		}
		if !exists {
			reason = "unknown_lease"
			return nil
		}
		if terminal(e.Record) != "" {
			r := e.Record
			out.Lease = &r
			reason = terminal(r)
			return nil
		}
		if m.observer == nil {
			return state.Refuse("environment_lease_release_unconfirmed", "session_state_unknown")
		}
		proof, err := m.observer.Observe(ctx, e.Record.Holder)
		if err != nil || !reflect.DeepEqual(proof.Holder, e.Record.Holder) {
			return state.Refuse("environment_lease_release_unconfirmed", "session_state_unknown")
		}
		confirmed := proof.Terminal && proof.DescendantsTerminated && proof.OwnershipReconciled
		if e.Record.Holder.Kind == "prepared" {
			confirmed = proof.NoStart && proof.DescendantsTerminated && proof.OwnershipReconciled
		}
		if !confirmed {
			return state.Refuse("environment_lease_release_unconfirmed", "termination_unconfirmed")
		}
		observation := "session_down"
		if proof.NoStart {
			observation = "never_started"
		}
		r, err := m.finish(l, e, lock, engine, "release", "authoritative reconciliation", observation)
		if err != nil {
			return err
		}
		out = Outcome{Changed: true, Lease: &r}
		reason = "released"
		return nil
	})
	return
}

func (m *Manager) Break(ctx context.Context, id string, revision int64, operator bool, acknowledgement string) (out Outcome, reason string, err error) {
	if !operator {
		return out, "", state.Refuse("usage", "operator_confirmation_required")
	}
	if !validReason(acknowledgement) {
		return out, "", state.Refuse("usage", "reason_required")
	}
	if revision < 1 || revision > protocoljson.MaxSafeInteger {
		return out, "", state.Refuse("usage", "revision_required")
	}
	if !state.ValidID(id) {
		return out, "", state.Refuse("usage", "invalid_selector")
	}
	err = m.withState(ctx, func(l *state.Ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		e, exists, err := m.lookup(l, id)
		if err != nil {
			return err
		}
		if !exists {
			return state.Refuse("environment_lease_unknown", "unknown_lease")
		}
		if e.Record.Phase == "broken" && e.TerminalEvent != nil {
			raw, err := state.Read(m.home, state.EventPath(m.home, *e.TerminalEvent))
			if err != nil {
				return err
			}
			var event state.Event
			if err := protocoljson.UnmarshalCanonical(raw, &event); err != nil {
				return err
			}
			if event.Previous.Revision == revision && event.Reason == acknowledgement {
				r := e.Record
				out.Lease = &r
				reason = "already_broken"
				return nil
			}
		}
		if !e.Record.Blocking() || revision != e.Record.Revision {
			return state.Refuse("environment_lease_revision_conflict", "revision_changed")
		}
		r, err := m.finish(l, e, lock, engine, "break", acknowledgement, "operator_break")
		if err != nil {
			return err
		}
		out = Outcome{Changed: true, Lease: &r}
		reason = "broken"
		return nil
	})
	return
}

func validReason(s string) bool {
	if !utf8.ValidString(s) || len([]rune(s)) > 1024 || strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if r < 32 || r == 127 || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func (m *Manager) finish(l *state.Ledger, e state.Entry, lock *managerlock.HomeLock, engine *transaction.Engine, operation, reason, observation string) (Record, error) {
	previous := e.Record
	if previous.Revision >= protocoljson.MaxSafeInteger {
		return Record{}, state.Refuse("environment_lease_revision_conflict", "revision_changed")
	}
	id, err := nonce()
	if err != nil {
		return Record{}, err
	}
	phase := "released"
	if operation == "break" {
		phase = "broken"
	}
	at := now()
	e.Record.Phase = phase
	e.Record.Revision++
	e.Record.Observation = state.Observation{State: observation, ObservedAt: &at}
	event := state.Event{Kind: "home-lease-event/1", EventID: "event-" + id, Operation: operation, LeaseID: e.Record.LeaseID, Previous: previous, NewRevision: e.Record.Revision, Reason: reason, RecordedAt: at}
	e.TerminalEvent = &event.EventID
	l.Leases[e.Record.LeaseID] = e
	raw, err := state.Canonical(event)
	if err != nil {
		return Record{}, err
	}
	if err := state.Validate("event", raw); err != nil {
		return Record{}, err
	}
	if err := m.publish(l, lock, engine, map[string][]byte{state.EventPath(m.home, event.EventID): raw}); err != nil {
		return Record{}, err
	}
	return e.Record, nil
}
