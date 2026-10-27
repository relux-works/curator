package operatorpolicy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

// hookObservation may only be supplied by a version-qualified internal adapter.
// The CLI accepts neither observations nor qualification booleans from callers.
type hookObservation struct{ Effective, NativeApprovals []Hook }
type observe func(context.Context, Selection, Event) (hookObservation, error)

// Read evaluates current consent for one start. No vendor adapter is qualified
// for CIP-0020 policy hooks/destination routing yet, so known valid grants fail
// closed with workspace_trust_adapter_unqualified. Unknown projects still fail
// with workspace_untrusted before reaching that qualification boundary.
func (s Store) Read(ctx context.Context, sel Selection) (*Delivery, error) {
	return s.read(ctx, sel, func(context.Context, Selection, Event) (hookObservation, error) {
		return hookObservation{}, refusal("workspace_trust_adapter_unqualified")
	})
}
func (s Store) read(ctx context.Context, sel Selection, observer observe) (*Delivery, error) {
	// Proven absence is distinct from an unreadable existing ledger allocation.
	if _, err := os.Lstat(filepath.Join(s.Home, "operator-policy")); errors.Is(err, os.ErrNotExist) {
		return nil, refusal("workspace_untrusted")
	}
	op, e := s.begin(ctx, false)
	if e != nil {
		return nil, e
	}
	defer op.close()
	var ev *Event
	profileMismatch := false
	active := 0
	for i := len(op.state.Events) - 1; i >= 0; i-- {
		candidate := &op.state.Events[i]
		saved := op.state.Selections[candidate.ID]
		if saved.Checkout == sel.Checkout && saved.Path == sel.Path && saved.Environment == sel.Environment && saved.Lane == sel.Lane {
			if saved.Profile != sel.Profile {
				profileMismatch = true
				continue
			}
			if candidate.State.Status == "active" {
				active++
				if active > 1 {
					return nil, refusalDetail("workspace_trust_policy_authority", "ambiguous active entries for selection")
				}
				ev = candidate
			} else if ev == nil {
				ev = candidate
			}
		}
	}
	if ev == nil {
		if profileMismatch {
			return nil, refusal("project_context_drift")
		}
		return nil, refusal("workspace_untrusted")
	}
	if ev.State.Status == "revoked" {
		return nil, refusal("workspace_trust_grant_revoked")
	}
	if ev.State.Status == "lapsed" {
		return nil, refusal("workspace_untrusted")
	}
	p, ok := op.state.Policies[ev.Policy.ID]
	if !ok {
		return nil, refusal("workspace_untrusted")
	}
	if hash(p) != ev.Policy.SHA256 {
		return nil, refusal("workspace_trust_policy_digest")
	}
	if !has(p.Environments, sel.Environment) || !has(p.Lanes, sel.Lane) {
		return nil, refusal("workspace_untrusted")
	}
	deadline, _ := timestamp(ev.ReviewBy)
	lapse := func(code string) (*Delivery, error) {
		ev.History = append(ev.History, ev.State)
		ev.State = State{"lapsed", time.Now().UTC().Format(time.RFC3339), "Current authorization identity or deadline lost", nil}
		if err := op.save(); err != nil {
			return nil, err
		}
		return nil, refusal(code)
	}
	if !time.Now().UTC().Before(deadline) {
		return lapse("workspace_untrusted")
	}
	live, e := subject(s.Home, op.held.Held(), sel)
	if e != nil {
		if strings.HasPrefix(e.Error(), "workspace_trust_repository_changed:") {
			return lapse("workspace_trust_repository_changed")
		}
		if strings.HasPrefix(e.Error(), "workspace_trust_symlink:") {
			// An unavailable or linked registered root is observed identity loss,
			// not a transient opportunity to restore the old consent later.
			return lapse("workspace_trust_symlink")
		}
		return nil, e
	}
	if ev.Subject.RegistrationSHA256 != live.RegistrationSHA256 || ev.Identity.Root != live.Registration.RootIdentity || ev.Subject.Repository != (Repository{live.RepositoryRoot, live.PrimaryRemote, live.RootCommit}) {
		return lapse("workspace_trust_repository_changed")
	}
	if live.PrimaryRemote != p.Destination.remote() || !reflect.DeepEqual(ev.Subject.Destination, p.Destination) {
		return nil, refusal("workspace_trust_destination_mismatch")
	}
	if observer == nil {
		return nil, refusal("workspace_trust_adapter_unqualified")
	}
	observed, e := observer(ctx, sel, *ev)
	if e != nil {
		return nil, e
	}
	if observed.Effective == nil || observed.NativeApprovals == nil {
		return nil, refusal("workspace_trust_hook_digest")
	}
	for _, h := range append(append([]Hook{}, observed.Effective...), observed.NativeApprovals...) {
		found := false
		for _, g := range ev.Surfaces.Hooks {
			if h == g {
				found = true
				break
			}
		}
		if !found {
			return nil, refusal("workspace_trust_hook_digest")
		}
	}
	if e = op.held.AssertCurrent(); e != nil {
		return nil, e
	}
	return &Delivery{ev.ID, ev.Policy, sel, ev.Surfaces}, nil
}

// Inspection excludes signatures, command payloads and native contents. Events
// contain only sanitized canonical destinations and hook IDs/digests. Native
// home projections are unsupported, so touched homes/cleanup are always empty.
type Inspection struct {
	EntryRevisions map[string]uint64 `json:"entry_revisions"`
	Policies       map[string]Policy `json:"policies"`
	Current        []PolicyStatus    `json:"current"`
	Events         []Event           `json:"events"`
	TouchedHomes   []string          `json:"touched_homes"`
	PendingCleanup []string          `json:"pending_cleanup"`
}

// PolicyStatus reports current digest and effective status without repairing.
type PolicyStatus struct {
	Policy PolicyRef `json:"policy"`
	Status string    `json:"status"`
}

// List never grants, repairs, renews, projects or lapses an event.
func (s Store) List(ctx context.Context) (Inspection, error) {
	op, e := s.begin(ctx, false)
	if e != nil {
		return Inspection{}, e
	}
	defer op.close()
	current := []PolicyStatus{}
	for id, policy := range op.state.Policies {
		status := "lapsed"
		deadline, _ := timestamp(policy.ReviewBy)
		if time.Now().UTC().Before(deadline) {
			for _, event := range op.state.Events {
				if event.Policy.ID == id && event.Policy.SHA256 == hash(policy) && event.State.Status == "active" {
					status = "active"
				}
			}
		}
		current = append(current, PolicyStatus{PolicyRef{id, hash(policy)}, status})
	}
	sort.Slice(current, func(i, j int) bool { return current[i].Policy.ID < current[j].Policy.ID })
	return Inspection{EntryRevisions: op.entryRevisions(), Policies: op.state.Policies, Current: current, Events: op.state.Events, TouchedHomes: []string{}, PendingCleanup: []string{}}, nil
}
