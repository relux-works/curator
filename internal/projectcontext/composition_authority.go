package projectcontext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

// CompositionAuthority mirrors the I-LEASE contract until its owning package
// is integrated. These exact shapes will be replaced with that package import.
type CompositionAuthority interface {
	Resolve(context.Context, managerlock.Held, string, string) (Composition, error)
	Authorize(context.Context, managerlock.Held, Composition, string) error
}

// Authority is bound to an operator-controlled store and the invocation CWD.
// Empty CWD means the process CWD, never the registered project root as fallback.
// It grants no agent authority: the current I-REG lane admits the local manager
// principal only. A future grant entry point must enforce its own live ceiling.
type Authority struct {
	Store Store
	CWD   string
}

var _ CompositionAuthority = Authority{}

// ErrAuthorityRevoked mirrors projectlease.ErrRevoked until lease wiring.
var ErrAuthorityRevoked = errors.New("project authority revoked")

type compositionEnvelope struct {
	Version        int
	Request        ComposeRequest
	Preview        CompositionPreview
	RuntimeEntries []string
	BuildKeys      []string
}

// PublishComposition retains a derived record and supporting metadata only.
// Rendering is previewed first and recomputed under the same manager lock just
// before publication. Generated files are consumed by the separate lease lane;
// this method does not mutate the managed home, profile lock or lease state.
func (a Authority) PublishComposition(ctx context.Context, req ComposeRequest, expected string, ownerRef string) (Composition, error) {
	req.Policy = canonicalCompositionPolicy(req.Policy)
	op, err := a.Store.beginValidated(false, func() error { return envauthority.Validate(a.Store.Home) })
	if err != nil {
		return Composition{}, err
	}
	defer op.close()
	if ownerRef != op.state.Principal {
		return Composition{}, fail("project_item_unapproved", "owner has no current manager authority")
	}
	if err = a.bound(op, req.CheckoutID); err != nil {
		return Composition{}, err
	}
	plan, err := op.compose(ctx, req)
	if err != nil {
		return Composition{}, err
	}
	if !hexRE.MatchString(expected) || plan.Preview.SHA256 != expected {
		return Composition{}, fail("project_context_drift", "composition differs from operator preview")
	}
	// A fresh recomputation catches changed protected bytes, approvals, policy or
	// registration between the first read and the publication boundary.
	again, err := op.compose(ctx, req)
	if err != nil {
		return Composition{}, err
	}
	if identity(plan) != identity(again) {
		return Composition{}, fail("project_context_drift", "composition changed before publication")
	}
	if err = a.bound(op, req.CheckoutID); err != nil {
		return Composition{}, err
	}
	// Retain only canonical request representations after admission: aliases for
	// the same record must publish identical immutable metadata.
	req.Mode = again.Preview.Record.Mode
	req.Selections = canonicalCompositionSelections(req.Selections, op.state.Reviews[req.CheckoutID].Review)
	envelope := compositionEnvelope{1, req, again.Preview, again.RuntimeEntries, again.BuildKeys}
	b, err := canonical(envelope)
	if err != nil {
		return Composition{}, err
	}
	if len(b) > MaxRecord {
		return Composition{}, fail("project_context_invalid", "composition metadata exceeds record bound")
	}
	dir := filepath.Join(op.root, "compositions")
	if err = privatedir.MakeAll(dir); err != nil {
		return Composition{}, err
	}
	if err = pathboundary.Validate(dir); err != nil {
		return Composition{}, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return Composition{}, err
	}
	defer func() { _ = root.Close() }()
	name := again.Preview.SHA256 + ".json"
	metadata, e := stateread.LstatWith(filepath.Join(dir, name), func(string) (os.FileInfo, error) { return root.Lstat(name) })
	if e != nil {
		return Composition{}, fail("project_context_unreadable", "composition state unknown")
	}
	if metadata.Kind == stateread.KindPresent {
		old, _, e := read(root, name, MaxRecord)
		if e != nil {
			return Composition{}, e
		}
		if !bytes.Equal(old, b) {
			return Composition{}, fail("project_context_drift", "immutable composition metadata changed")
		}
	} else {
		if err = ctx.Err(); err != nil {
			return Composition{}, err
		}
		if err = op.lock.Held().Validate(filepath.Dir(op.root)); err != nil {
			return Composition{}, err
		}
		if err = envauthority.Validate(a.Store.Home); err != nil {
			return Composition{}, err
		}
		if err = writeAtomic(dir, name, b); err != nil {
			return Composition{}, err
		}
	}
	return compositionOf(again), nil
}
func (a Authority) bound(op *operation, id string) error {
	r, err := op.registered(id)
	if err != nil {
		return err
	}
	cwd := a.CWD
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return fail("project_registration_mismatch", "CWD unavailable")
		}
	}
	selected, err := op.locateFor(cwd, r.Record.Profile.Name, r.Record.Environment)
	if err != nil {
		return err
	}
	if selected.CheckoutID != id {
		return fail("project_registration_mismatch", "composition belongs to another CWD binding")
	}
	return nil
}

// Resolve rechecks a retained composition under the caller's held home lock.
func (a Authority) Resolve(ctx context.Context, held managerlock.Held, ref, sha256 string) (Composition, error) {
	if err := envauthority.Validate(a.Store.Home); err != nil {
		return Composition{}, err
	}
	if err := ctx.Err(); err != nil {
		return Composition{}, err
	}
	op, err := a.Store.readHeld(held)
	if err != nil {
		return Composition{}, err
	}
	c, err := a.resolve(ctx, op, ref, sha256)
	if err != nil {
		return Composition{}, err
	}
	if err = held.Validate(filepath.Dir(op.root)); err != nil {
		return Composition{}, err
	}
	return c, nil
}

func (a Authority) resolve(ctx context.Context, op *operation, ref, hash string) (Composition, error) {
	if !hexRE.MatchString(hash) || ref != "composition:"+hashID(hash) {
		return Composition{}, fail("project_context_invalid", "invalid composition reference or digest")
	}
	dir := filepath.Join(op.root, "compositions")
	if err := pathboundary.Validate(dir); err != nil {
		return Composition{}, fail("project_context_drift", "composition boundary unavailable")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return Composition{}, fail("project_context_unreadable", "composition record unavailable")
	}
	defer func() { _ = root.Close() }()
	b, _, err := read(root, hash+".json", MaxRecord)
	if err != nil {
		return Composition{}, fail("project_context_unreadable", "composition record unavailable")
	}
	var saved compositionEnvelope
	if err = decode(b, &saved); err != nil {
		return Composition{}, err
	}
	record, err := canonical(saved.Preview.Record)
	if err != nil {
		return Composition{}, err
	}
	if saved.Version != 1 || digest(record) != hash || saved.Preview.SHA256 != hash || saved.Preview.Ref != ref {
		return Composition{}, fail("project_context_drift", "composition digest differs")
	}
	if _, err = ParseCompositionRecord(record); err != nil {
		return Composition{}, err
	}
	if err = a.bound(op, saved.Request.CheckoutID); err != nil {
		return Composition{}, err
	}
	for _, item := range saved.Preview.Generation.Items {
		approval, ok := op.state.Approvals[saved.Request.CheckoutID+":"+item.ItemID]
		if ok && (approval.Revoked || approval.Record.ItemSHA256 != item.ItemSHA256) {
			return Composition{}, fmt.Errorf("project_item_unapproved: %w", ErrAuthorityRevoked)
		}
	}
	plan, err := op.compose(ctx, saved.Request)
	if err != nil {
		return Composition{}, err
	}
	fresh := compositionEnvelope{1, saved.Request, plan.Preview, plan.RuntimeEntries, plan.BuildKeys}
	if identity(fresh) != identity(saved) {
		return Composition{}, fail("project_context_drift", "composition supporting records or receipt differ")
	}
	return compositionOf(plan), nil
}

// Authorize rechecks current principal authority and the complete composition.
func (a Authority) Authorize(ctx context.Context, held managerlock.Held, c Composition, ownerRef string) error {
	if err := envauthority.Validate(a.Store.Home); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	op, err := a.Store.readHeld(held)
	if err != nil {
		return err
	}
	if !tagged(ownerRef, "principal") || ownerRef != op.state.Principal {
		return fail("project_item_unapproved", "owner has no current manager authority")
	}
	resolved, err := a.resolve(ctx, op, c.Ref, c.SHA256)
	if err != nil {
		return err
	}
	if identity(resolved) != identity(c) {
		return fail("project_context_drift", "resolved composition binding or receipt differs")
	}
	if err = held.Validate(filepath.Dir(op.root)); err != nil {
		return err
	}
	return ctx.Err()
}

func compositionOf(plan compositionPlan) Composition {
	p := plan.Preview
	c := Composition{Ref: p.Ref, SHA256: p.SHA256, CheckoutID: strings.TrimPrefix(p.Record.RegistrationRef, "registration:"), Profile: p.Record.Profile.Name, Environment: p.Record.Environment, HomeRef: p.Record.HomeRef, ConcurrentState: plan.ConcurrentState, BuildKeys: append([]string{}, plan.BuildKeys...), RuntimeEntries: append([]string{}, plan.RuntimeEntries...), Generated: map[string]Surface{}}
	for _, f := range p.Receipt.Files {
		mode, _ := strconv.ParseUint(f.Mode, 8, 32)
		c.Generated[f.Path] = Surface{f.SHA256, f.Size, uint32(mode)}
	}
	return c
}

// canonicalCompositionSelections runs after admission has rejected duplicates,
// unknown items and implicit omissions. Retain the approved review order and an
// explicit empty list, without changing any enabled/disabled choice.
func canonicalCompositionSelections(selections []Selection, review Review) []Selection {
	selected := make(map[string]Selection, len(selections))
	for _, selection := range selections {
		selected[selection.ItemID] = selection
	}
	out := []Selection{}
	for _, candidate := range review.Items {
		if selection, ok := selected[candidate.Item.ID]; ok {
			out = append(out, selection)
		}
	}
	return out
}
