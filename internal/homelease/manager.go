// Package homelease implements profile-home reservations under the existing
// manager lock and transaction journal. Project leases keep their own authority.
package homelease

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"reflect"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease/persistence"
	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/transaction"
)

type Holder = state.Holder
type Record = state.Record
type Admission = state.Admission
type Refusal = state.Refusal

// Authority performs read-only current-home/materialization verification under
// the borrowed witness. A selector never supplies its own filesystem authority.
type Authority interface {
	Admit(context.Context, managerlock.Held, string, string, string) (Admission, error)
}

// Evidence is an internal trusted-observer result, not a CLI payload or a claim
// supplied by the holder. A consumer adapter must independently populate every
// identity and proof; absent integration cannot synthesize successful evidence.
type Evidence struct {
	Holder                Holder
	NoStart               bool
	Terminal              bool
	DescendantsTerminated bool
	OwnershipReconciled   bool
}
type Observer interface {
	Observe(context.Context, Holder) (Evidence, error)
}

type Request struct {
	Environment string `json:"environment"`
	Profile     string `json:"profile"`
	LockSHA256  string `json:"lock_sha256"`
	RequestID   string `json:"request_id"`
	Holder      Holder `json:"holder"`
}
type Outcome struct {
	Changed bool    `json:"changed"`
	Lease   *Record `json:"lease"`
}
type Manager struct {
	home      string
	authority Authority
	observer  Observer
	hooks     transaction.Hooks
}

func New(home string, authority Authority, observer Observer) (*Manager, error) {
	m, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	return &Manager{home: m.Home(), authority: authority, observer: observer}, nil
}

func hash(raw []byte) string { d := sha256.Sum256(raw); return hex.EncodeToString(d[:]) }
func nonce() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
func terminal(r Record) string {
	if r.Phase == "released" {
		return "already_released"
	}
	if r.Phase == "broken" {
		return "already_broken"
	}
	return ""
}

// withState never recovers before admission: pending mutations require their
// explicit authorized recovery owner. Project lifecycle validation is reused,
// and all namespace uncertainty remains a typed refusal.
func (m *Manager) withState(ctx context.Context, fn func(*state.Ledger, *managerlock.HomeLock, *transaction.Engine) error) error {
	locks, err := managerlock.New(m.home)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	lock, err := locks.AcquireHomeOnly(ctx, false)
	if err != nil {
		return state.Refuse("environment_lock_unavailable", "mutation_lock_busy")
	}
	defer func() { _ = lock.Close() }()
	if err := lock.Held().Validate(m.home); err != nil {
		return err
	}
	if err := state.LegacyPending(m.home); err != nil {
		return err
	}
	e, err := transaction.New(m.home, transaction.WithHooks(m.hooks), transaction.WithCommitCheck(projectlease.TransactionCheck(m.home)))
	if err != nil {
		return err
	}
	pending, err := e.PendingRecovery(lock)
	if err != nil {
		return state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	if pending {
		return state.Refuse("environment_lease_recovery_required", "pending_mutation")
	}
	if _, err := projectlease.ReferencesUnderLock(m.home, lock); err != nil {
		return state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	l, err := state.Load(m.home)
	if err != nil {
		return err
	}
	return fn(l, lock, e)
}

func (m *Manager) Acquire(ctx context.Context, req Request) (out Outcome, reason string, err error) {
	req.Environment = envregistry.NormalizeEnvID(req.Environment)
	if req.Environment != "claude_code" && req.Environment != "codex_cli" && req.Environment != "muse" {
		return out, "", state.Refuse("environment_unknown", "environment_absent")
	}
	if !state.ValidID(req.Profile) || !state.ValidID(req.RequestID) || !state.ValidDigest(req.LockSHA256) {
		return out, "", state.Refuse("usage", "invalid_selector")
	}
	raw, er := state.Canonical(req.Holder)
	if er != nil || state.Validate("holder", raw) != nil || req.Holder.Kind != "prepared" || req.Holder.RequestID != req.RequestID {
		return out, "", state.Refuse("usage", "invalid_holder")
	}
	payload, er := state.Canonical(req)
	if er != nil {
		return out, "", state.Refuse("usage", "invalid_holder")
	}
	err = m.withState(ctx, func(l *state.Ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		if id, ok := l.Requests[req.RequestID]; ok {
			e := l.Leases[id]
			if e.PayloadSHA256 != hash(payload) {
				return state.Refuse("environment_lease_holder_mismatch", "request_conflict")
			}
			r := e.Record
			out.Lease = &r
			reason = terminal(r)
			if reason == "" {
				reason = "acquired"
			}
			return nil
		}
		if m.authority == nil {
			return state.Refuse("environment_lease_authority_refused", "authority_unavailable")
		}
		admitted, err := m.authority.Admit(ctx, lock.Held(), req.Environment, req.Profile, req.LockSHA256)
		if err != nil {
			return err
		}
		identity, err := managerlock.New(admitted.Home)
		if err != nil {
			return state.Refuse("environment_home_stale", "home_not_current")
		}
		admitted.Home = identity.Home()
		for i, path := range admitted.Protected {
			if !filepath.IsAbs(path) || filepath.Clean(path) != path {
				return state.Refuse("environment_home_stale", "home_not_current")
			}
			physical, err := managerlock.New(path)
			if err != nil {
				return state.Refuse("environment_home_stale", "home_not_current")
			}
			admitted.Protected[i] = physical.Home()
		}
		for _, existing := range l.Leases {
			same, err := state.Overlap(admitted.Home, existing.Admission.Home)
			if err != nil {
				return err
			}
			if same && existing.Record.Blocking() && existing.Record.LockSHA256 != req.LockSHA256 {
				return state.Refuse("environment_lease_lock_mismatch", "profile_lock_changed")
			}
		}
		id, err := nonce()
		if err != nil {
			return err
		}
		if _, present, err := m.lookup(l, "home-"+id); err != nil || present {
			return state.Refuse("environment_lease_write_failed", "publication_failed")
		}
		r := Record{Kind: "home-lease/1", LeaseID: "home-" + id, ManagerRootID: "root-" + hash([]byte(m.home)), HomeID: "home-" + hash([]byte(admitted.Home)), Environment: req.Environment, Profile: req.Profile, LockSHA256: req.LockSHA256, Revision: 1, Phase: "prepared", AcquiredAt: now(), Holder: req.Holder, Observation: state.Observation{State: "not_observed"}}
		raw, err := state.Canonical(r)
		if err != nil || state.Validate("record", raw) != nil {
			return state.Refuse("usage", "invalid_selector")
		}
		l.Leases[r.LeaseID] = state.Entry{Record: r, Initial: r, PayloadSHA256: hash(payload), Admission: admitted}
		l.Requests[req.RequestID] = r.LeaseID
		if err := m.publish(l, lock, engine, map[string][]byte{state.BindingPath(m.home, r.LeaseID): raw}); err != nil {
			return err
		}
		out = Outcome{Changed: true, Lease: &r}
		reason = "acquired"
		return nil
	})
	return
}

func (m *Manager) publish(l *state.Ledger, lock *managerlock.HomeLock, engine *transaction.Engine, files map[string][]byte) error {
	if err := lock.Held().Validate(m.home); err != nil {
		return err
	}
	raw, err := state.Canonical(l)
	if err != nil || len(raw) > state.MaxRecord {
		return state.Refuse("environment_lease_write_failed", "publication_failed")
	}
	files[state.LedgerPath(m.home)] = raw
	id, err := nonce()
	if err != nil {
		return err
	}
	err = persistence.Publish(persistence.Publication{
		Home: m.home, StageDir: state.Namespace(m.home), TransactionID: "home-lease-" + id,
		Owner: "home-leases", Class: "home-lease", Files: files, PrivateFile: true, Route: state.CheckRoute,
		Immutable: func(path string) bool { return path != state.LedgerPath(m.home) },
	}, lock, engine)
	if err != nil {
		return state.Refuse("environment_lease_write_failed", "publication_failed")
	}
	return nil
}

func sameReservation(a, b Holder) bool {
	a.Kind, b.Kind = "prepared", "prepared"
	a.Session, b.Session = nil, nil
	a.Generation, b.Generation = nil, nil
	a.Process, b.Process = nil, nil
	return reflect.DeepEqual(a, b)
}
