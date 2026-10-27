package operatorpolicy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Claim: a fresh keeper request lacking operator approval cannot allocate an
// event, while the identical command with the real approval can. This avoids
// letting duplicate-approval rejection hide a broken signature gate.
func TestReviewerFreshAuthority(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	before := stateBytes(t, f)
	bad := f.command
	bad.ApprovalSignature = bad.ActorSignature
	out, e := f.store.Apply(ctx, bad)
	require(t, e, "workspace_trust_policy_authority")
	if out != nil {
		t.Fatal("unapproved keeper received event")
	}
	unchanged(t, f, before)
	approve(t, f)
}

// Claim: observing a missing/renamed checkout must retain terminal lapse;
// restoring the same inode cannot revive the old approval (CIP-0020 §2.1).
func TestReviewerMissingCheckoutLapse(t *testing.T) {
	f := newFixture(t)
	approve(t, f)
	ctx := context.Background()
	if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
		t.Fatal("positive", e)
	}
	moved := f.selection.Path + "-renamed"
	if e := os.Rename(f.selection.Path, moved); e != nil {
		t.Fatal(e)
	}
	out, e := f.store.Read(ctx, f.selection)
	require(t, e, "workspace_trust_symlink")
	if out != nil {
		t.Fatal("missing checkout delivered")
	}
	list, e := f.store.List(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if list.Events[0].State.Status != "lapsed" {
		t.Errorf("observed missing checkout remains %s, want lapsed", list.Events[0].State.Status)
	}
	lapsed := stateBytes(t, f)
	if e := os.Rename(moved, f.selection.Path); e != nil {
		t.Fatal(e)
	}
	out, e = f.store.Read(ctx, f.selection)
	if e == nil || !strings.HasPrefix(e.Error(), "workspace_untrusted:") || out != nil {
		t.Errorf("public read no longer reports terminal lapse: delivery=%v error=%v", out != nil, e)
	}
	out, e = f.store.read(ctx, f.selection, observeGranted)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatalf("restored checkout revived old consent: delivery=%v error=%v", out != nil, e)
	}
	_, e = f.store.Apply(ctx, f.command)
	require(t, e, "workspace_trust_grant_revoked")
	unchanged(t, f, lapsed)
	// Only a new owner signature at the retained entry revision restores trust;
	// the original event remains terminal and retains its lapse provenance.
	c := f.command.Command
	c.EntryRevision = entryRevision(t, f, c.EntryID)
	c.Issuer.Approval = strings.Repeat("6", 64)
	c.Time = time.Now().UTC().Format(time.RFC3339)
	p := *c.Policy
	p.Time, p.Issuer = c.Time, c.Issuer
	c.Policy = &p
	f.command = signed(t, c, f.operator, f.actor)
	ev := approve(t, f)
	if out, e = f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil || out.GrantID != ev.ID {
		t.Fatal("new approval failed", e)
	}
	list, e = f.store.List(ctx)
	if e != nil || len(list.Events) != 2 || list.Events[0].State.Status != "revoked" || len(list.Events[0].History) != 2 || list.Events[0].History[1].Status != "lapsed" {
		t.Fatalf("lost terminal lapse: %+v %v", list, e)
	}
}

// Claim: two independently signed, active entries for the exact same selection
// are ambiguous and must not silently select the last one (CIP-0020 §2.1).
func TestReviewerAmbiguousEntries(t *testing.T) {
	f := newFixture(t)
	approve(t, f)
	ctx := context.Background()
	if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
		t.Fatal("positive", e)
	}
	c := f.command.Command
	p := *c.Policy
	c.EntryID = "second"
	c.Issuer.Approval = strings.Repeat("9", 64)
	p.ID = c.EntryID
	p.Issuer = c.Issuer
	p.Hooks = []Hook{}
	c.Policy = &p
	before := stateBytes(t, f)
	out, e := f.store.Apply(ctx, signed(t, c, f.operator, f.actor))
	require(t, e, "workspace_trust_policy_authority")
	unchanged(t, f, before)
	if out != nil {
		t.Fatal("ambiguous write emitted event")
	}
	// Another concrete lane is a different full selection, not an ambiguity.
	sel := *c.Selection
	sel.Lane = "lane:second"
	c.Selection = &sel
	p.Lanes = []string{sel.Lane}
	c.Policy = &p
	if _, e = f.store.Apply(ctx, signed(t, c, f.operator, f.actor)); e != nil {
		t.Fatal("distinct lane refused", e)
	}
	if delivery, e := f.store.read(ctx, f.selection, observeGranted); e != nil || delivery == nil {
		t.Fatal("original selection lost", e)
	}
	if delivery, e := f.store.read(ctx, sel, observeGranted); e != nil || delivery == nil || delivery.Policy.ID != "second" {
		t.Fatal("distinct selection failed", e)
	}
}

// Claim: the new package's startup pins all standard roots and clears vendor
// overrides before test cases, as the CLI package separately demonstrates.
func TestReviewerPolicyHomeIsolation(t *testing.T) {
	h := os.Getenv("HOME")
	if !strings.Contains(h, "operator-policy-home-") || os.Getenv("USERPROFILE") != h || os.Getenv("XDG_CONFIG_HOME") != h || os.Getenv("CODEX_HOME") != "" || os.Getenv("CLAUDE_CONFIG_DIR") != "" {
		t.Fatal("operator policy test home boundary missing")
	}
}

// Claim: candidate trust declarations in project/profile files and environment
// do not replace signed commands. These are injection probes, not supported APIs.
func TestReviewerAmbientClaimsDoNotGrant(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	before := stateBytes(t, f)
	payload := []byte(`{"project_trust":true,"hooks":[],"issuer":{"role":"operator"}}`)
	write(t, filepath.Join(f.selection.Path, ".curator", "policy.json"), payload)
	write(t, filepath.Join(f.selection.Path, ".claude", "settings.json"), payload)
	write(t, filepath.Join(f.store.Home, "profiles", "base", "operator-policy.json"), payload)
	t.Setenv("CURATOR_WORKSPACE_TRUST", "true")
	t.Setenv("CURATOR_OPERATOR_POLICY", string(payload))
	out, e := f.store.Read(ctx, f.selection)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatal("ambient declaration granted authority")
	}
	unchanged(t, f, before)
	peer := f.command
	peer.ApprovalSignature = peer.ActorSignature
	_, e = f.store.Apply(ctx, peer)
	require(t, e, "workspace_trust_policy_authority")
	unchanged(t, f, before)
	approve(t, f)
}

// Claim: revocation supersedes still-unconsumed commands approved earlier;
// a genuinely later approval is the adjacent valid regrant control.
func TestReviewerDelayedApprovalAfterRevoke(t *testing.T) {
	for _, name := range []string{"fresh-approval", "old-approval", "old-approval-same-clock", "old-approval-newer-clock"} {
		t.Run(name, func(t *testing.T) {
			fresh := name == "fresh-approval"
			f := newFixture(t)
			ctx := context.Background()
			now := time.Now().UTC()
			c := f.command.Command
			p := *c.Policy
			c.Time = now.Add(-30 * time.Second).Format(time.RFC3339)
			p.Time = c.Time
			c.Policy = &p
			f.command = signed(t, c, f.operator, f.actor)
			approve(t, f)
			rev := Command{Kind: c.Kind, Operation: "revoke", Issuer: c.Issuer, Time: now.Add(-10 * time.Second).Format(time.RFC3339), Reason: "Revoke all consent", EntryID: c.EntryID}
			rev.Issuer.Approval = strings.Repeat("8", 64)
			rev.EntryRevision = entryRevision(t, f, c.EntryID)
			c.EntryRevision = rev.EntryRevision
			c.Issuer.Approval = strings.Repeat("7", 64)
			c.Time = now.Add(-20 * time.Second).Format(time.RFC3339)
			if fresh {
				c.Time = now.Add(-5 * time.Second).Format(time.RFC3339)
			}
			if name == "old-approval-same-clock" {
				c.Time = rev.Time
			}
			if name == "old-approval-newer-clock" {
				c.Time = now.Add(-5 * time.Second).Format(time.RFC3339)
			}
			p.Time = c.Time
			p.Issuer = c.Issuer
			c.Policy = &p
			var pending SignedCommand
			if !fresh {
				pending = signed(t, c, f.operator, f.actor)
			}
			if _, e := f.store.Apply(ctx, signed(t, rev, f.operator, f.actor)); e != nil {
				t.Fatal(e)
			}
			if fresh {
				c.EntryRevision = entryRevision(t, f, c.EntryID)
				pending = signed(t, c, f.operator, f.actor)
			}
			before := stateBytes(t, f)
			out, e := f.store.Apply(ctx, pending)
			if fresh {
				if e != nil || out == nil {
					t.Fatal("fresh approval rejected", e)
				}
				if delivery, e := f.store.read(ctx, f.selection, observeGranted); e != nil || delivery == nil || delivery.GrantID != out.ID {
					t.Fatal("fresh approval not effective", e)
				}
				return
			}
			if e == nil || out != nil {
				t.Errorf("pre-revocation approval created new active consent: event=%v error=%v", out != nil, e)
			}
			require(t, e, "workspace_trust_policy_authority")
			delivery, e := f.store.Read(ctx, f.selection)
			require(t, e, "workspace_trust_grant_revoked")
			if delivery != nil {
				t.Fatal("stale approval revived revoked entry")
			}
			unchanged(t, f, before)
		})
	}
}

// Claim: missing hook digests and extra effective hooks refuse beside a valid
// observation; the internal observer establishes no vendor qualification.
func TestReviewerMissingDigestAndExtraHook(t *testing.T) {
	for _, name := range []string{"missing-digest", "extra-hook"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			approve(t, f)
			ctx := context.Background()
			if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
				t.Fatal("positive", e)
			}
			before := stateBytes(t, f)
			observer := func(ctx context.Context, s Selection, e Event) (hookObservation, error) {
				o, _ := observeGranted(ctx, s, e)
				if name == "missing-digest" {
					o.Effective[0].SHA256 = ""
				} else {
					o.Effective = append(o.Effective, Hook{"extra", strings.Repeat("d", 64)})
				}
				return o, nil
			}
			out, e := f.store.read(ctx, f.selection, observer)
			require(t, e, "workspace_trust_hook_digest")
			if out != nil {
				t.Fatal("forbidden delivery")
			}
			unchanged(t, f, before)
		})
	}
}

// Claim: independently valid legacy grants merged into one ledger refuse on
// both public and core reads in either insertion order, before any observer.
// The single-entry fixture is the adjacent core positive; no vendor is qualified.
func TestReviewerLegacyAmbiguousRead(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			f := newFixture(t)
			ctx := context.Background()
			empty := stateBytes(t, f)
			approve(t, f)
			if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
				t.Fatal("single-entry control", e)
			}
			var first ledger
			if e := decode(stateBytes(t, f), &first); e != nil {
				t.Fatal(e)
			}
			// Simulate pre-fix/corrupt persisted state without bypassing signature,
			// registration, schema or policy-digest validation on either entry.
			write(t, filepath.Join(f.store.Home, "operator-policy", "state.json"), empty)
			c := f.command.Command
			p := *c.Policy
			c.EntryID, p.ID = "second", "second"
			c.Issuer.Approval = strings.Repeat("9", 64)
			p.Issuer, p.Hooks = c.Issuer, []Hook{}
			c.Policy = &p
			if _, e := f.store.Apply(ctx, signed(t, c, f.operator, f.actor)); e != nil {
				t.Fatal(e)
			}
			var second ledger
			if e := decode(stateBytes(t, f), &second); e != nil {
				t.Fatal(e)
			}
			for k, v := range second.Audit {
				first.Audit[k] = v
			}
			for k, v := range second.Policies {
				first.Policies[k] = v
			}
			for k, v := range second.Selections {
				first.Selections[k] = v
			}
			first.Events = append(first.Events, second.Events...)
			if reverse {
				first.Events[0], first.Events[1] = first.Events[1], first.Events[0]
			}
			b, e := canonical(first)
			if e != nil {
				t.Fatal(e)
			}
			write(t, filepath.Join(f.store.Home, "operator-policy", "state.json"), b)
			if _, e = f.store.List(ctx); e != nil {
				t.Fatal("fixture failed ledger authentication", e)
			}
			for _, public := range []bool{true, false} {
				var out *Delivery
				if public {
					out, e = f.store.Read(ctx, f.selection)
				} else {
					out, e = f.store.read(ctx, f.selection, func(context.Context, Selection, Event) (hookObservation, error) {
						t.Fatal("ambiguous read reached observer")
						return hookObservation{}, nil
					})
				}
				require(t, e, "workspace_trust_policy_authority")
				if out != nil {
					t.Fatal("ambiguous delivery")
				}
				unchanged(t, f, b)
			}
		})
	}
}

// Claim: revisions survive a revoke of an absent entry; its pre-revoke first
// approval cannot create consent, while a freshly signed current revision can.
// Altering the revision without both signatures confers no authority.
func TestPolicyRevisionRetainedWithoutEvent(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	c := f.command.Command
	rev := Command{Kind: c.Kind, Operation: "revoke", EntryID: c.EntryID, Issuer: c.Issuer, Time: time.Now().UTC().Format(time.RFC3339), Reason: "Revoke pending approvals"}
	rev.Issuer.Approval = strings.Repeat("8", 64)
	signedRev := signed(t, rev, f.operator, f.actor)
	if _, e := f.store.Apply(ctx, signedRev); e != nil {
		t.Fatal(e)
	}
	if _, e := f.store.Apply(ctx, signedRev); e != nil {
		t.Fatal("idempotent revoke", e)
	}
	if got := entryRevision(t, f, c.EntryID); got != 1 {
		t.Fatalf("revision=%d want 1", got)
	}
	before := stateBytes(t, f)
	out, e := f.store.Apply(ctx, f.command)
	require(t, e, "workspace_trust_policy_authority")
	if out != nil {
		t.Fatal("old first approval accepted after revoke")
	}
	unchanged(t, f, before)
	forged := f.command
	forged.Command.EntryRevision = 1
	_, e = f.store.Apply(ctx, forged)
	require(t, e, "workspace_trust_policy_authority")
	unchanged(t, f, before)
	c.EntryRevision = entryRevision(t, f, c.EntryID)
	f.command = signed(t, c, f.operator, f.actor)
	approve(t, f)
	if got := entryRevision(t, f, c.EntryID); got != 2 {
		t.Fatalf("revision=%d want 2", got)
	}
	if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
		t.Fatal("fresh approval", e)
	}
}
