package projectcontext

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Claim: held-lock readers retain main's durable lapse boundary: a restored
// inode cannot revive a composition after registration inspection observed loss.
// Missing or malformed lapse evidence also refuses beside a valid control.
func TestCompositionRetainedRootLapse(t *testing.T) {
	for _, damage := range []string{"restored-root", "missing-lapses", "malformed-lapses"} {
		t.Run(damage, func(t *testing.T) {
			f := compositionFixture(t, "codex_cli", nil)
			packageApproved(t, f, false)
			req := compositionRequest(f)
			a := Authority{Store: f.store, CWD: f.root}
			owner := principal(t, f)
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
				t.Fatal(err)
			}
			if err = authorizeComposition(t, a, c, owner); err != nil {
				t.Fatal(err)
			}
			record := filepath.Join(f.store.Home, "project-context", "compositions", c.SHA256+".json")
			before, err := os.ReadFile(record)
			if err != nil {
				t.Fatal(err)
			}
			code := "project_context_drift"
			switch damage {
			case "restored-root":
				moved := f.root + "-moved"
				if err = os.Rename(f.root, moved); err != nil {
					t.Fatal(err)
				}
				_, err = f.store.Inspect(f.registration.CheckoutID)
				required(t, err, "project_registration_mismatch")
				if err = os.Rename(moved, f.root); err != nil {
					t.Fatal(err)
				}
				code = "project_registration_mismatch"
			case "missing-lapses":
				if err = os.Remove(filepath.Join(f.store.Home, "project-context", "lapses.json")); err != nil {
					t.Fatal(err)
				}
			case "malformed-lapses":
				write(t, filepath.Join(f.store.Home, "project-context", "lapses.json"), `{}`)
			}
			_, err = resolveComposition(t, a, c.Ref, c.SHA256)
			required(t, err, code)
			required(t, authorizeComposition(t, a, c, owner), code)
			after, err := os.ReadFile(record)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("refused lapse evidence changed composition")
			}
		})
	}
}

// Claim: resolving a parent composition from a missing nested registration
// observes that child's root loss durably, while an unrelated missing child CWD
// cannot lapse its valid parent. Restored nested roots still require rebind.
func TestCompositionMissingCWDObservesRootLoss(t *testing.T) {
	for _, registered := range []bool{false, true} {
		t.Run(map[bool]string{false: "unregistered-child", true: "registered-child"}[registered], func(t *testing.T) {
			f := compositionFixture(t, "codex_cli", nil)
			packageApproved(t, f, false)
			req := compositionRequest(f)
			a := Authority{Store: f.store, CWD: f.root}
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			c, err := a.PublishComposition(context.Background(), req, p.SHA256, principal(t, f))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
				t.Fatal(err)
			}
			child := filepath.Join(f.root, "nested")
			initRepo(t, child)
			var nested Registration
			if registered {
				nested, err = f.store.Register("nested", child, "base", "codex_cli")
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.store.Inspect(nested.CheckoutID); err != nil {
					t.Fatal(err)
				}
			}
			moved := child + "-moved"
			if err = os.Rename(child, moved); err != nil {
				t.Fatal(err)
			}
			a.CWD = child
			_, err = resolveComposition(t, a, c.Ref, c.SHA256)
			required(t, err, "project_registration_mismatch")
			if err = os.Rename(moved, child); err != nil {
				t.Fatal(err)
			}
			if registered {
				_, err = f.store.Inspect(nested.CheckoutID)
				if err == nil {
					t.Fatal("restored nested registration revived after composition observed root loss")
				}
				required(t, err, "project_registration_mismatch")
			}
			a.CWD = f.root
			if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
				t.Fatalf("valid parent lost authority: %v", err)
			}
		})
	}
}

// Claim: nil and explicit empty waiver sets publish the same resolvable
// metadata; a null waiver map planted in that retained record is rejected by
// main's exact decoder without replacing the immutable record.
func TestCompositionPolicyMetadata(t *testing.T) {
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	req := compositionRequest(f)
	a := Authority{Store: f.store, CWD: f.root}
	owner := principal(t, f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
		t.Fatalf("default policy publication cannot resolve: %v", err)
	}
	req.Policy.System.Waivers = map[string]bool{}
	q, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil || q.SHA256 != p.SHA256 {
		t.Fatalf("empty waiver set changed composition: %v", err)
	}
	if _, err = a.PublishComposition(context.Background(), req, q.SHA256, owner); err != nil {
		t.Fatalf("empty waiver set changed retained metadata: %v", err)
	}
	record := filepath.Join(f.store.Home, "project-context", "compositions", c.SHA256+".json")
	before, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	var envelope compositionEnvelope
	if err = decode(before, &envelope); err != nil {
		t.Fatal(err)
	}
	envelope.Request.Policy.System.Waivers = nil
	planted, err := canonical(envelope)
	if err != nil {
		t.Fatal(err)
	}
	write(t, record, string(planted))
	_, err = resolveComposition(t, a, c.Ref, c.SHA256)
	required(t, err, "project_context_invalid")
	after, err := os.ReadFile(record)
	if err != nil || !bytes.Equal(planted, after) {
		t.Fatal("null waiver refusal replaced retained metadata")
	}
	write(t, record, string(before))
	if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
		t.Fatal(err)
	}
}
