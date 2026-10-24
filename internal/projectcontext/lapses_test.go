package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim R12: production Register/Preview/Approve can fill the ledger, including
// exactly MaxRecord bytes, without crowding out a durable root lapse. Restoring
// permissions cannot revive either active slot, even in another process. The
// 64-byte sibling and ample-capacity controls still admit before loss; explicit
// rebind produces a fresh ID that requires a new approval. The legacy variant
// changes only the private envelope version and removes the new companion to
// model the previous writer; registrations, reviews and approvals are unedited.
func TestRootLapseAtCapacity(t *testing.T) {
	if os.Getenv("CURATOR_LAPSE_CHILD") == "1" {
		store := Store{Home: os.Getenv("CURATOR_LAPSE_HOME")}
		_, err := store.Admitted(os.Getenv("CURATOR_LAPSE_ID"), []Selection{{"guide", false}})
		if !isCode(err, "project_registration_mismatch") {
			t.Fatalf("fresh process revived old authority: %v", err)
		}
		return
	}
	for _, tc := range []struct {
		name string
		gap  int
	}{{"full-ledger", 0}, {"five-byte-headroom", 5}, {"64-byte-headroom", 64}, {"ample-headroom", MaxRecord / 2}, {"legacy-full-ledger", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			packageApproved(t, f, false)
			sibling, err := f.store.Register("fixture", f.root, "base", "claude_code")
			if err != nil {
				t.Fatal(err)
			}
			other := filepath.Join(filepath.Dir(f.root), "other")
			initRepo(t, other)
			r, err := f.store.Register("other", other, "base", "codex_cli")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(other, "skill/SKILL.md"), validSkillDocument)
			d := declaration("guide", "skill", true)
			d.Items[0].Type = "skill"
			statePath := filepath.Join(f.store.Home, "project-context/state.json")
			hintSize := 1000
			for attempt := 0; attempt < 3; attempt++ {
				commands, err := json.Marshal(map[string]any{"inspect-data": map[string]any{"type": "system", "command": "true", "hint": strings.Repeat("x", hintSize)}})
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(other, "skill/agent-skill.json"), skillManifest(string(commands), `{}`))
				if _, err := f.store.Preview(r.CheckoutID, d); err != nil {
					t.Fatal("bounded preview control", err)
				}
				b, err := os.ReadFile(statePath)
				if err != nil {
					t.Fatal(err)
				}
				if len(b) == MaxRecord-tc.gap {
					break
				}
				hintSize += MaxRecord - tc.gap - len(b)
			}
			before, err := os.ReadFile(statePath)
			if err != nil || len(before) != MaxRecord-tc.gap {
				t.Fatalf("capacity fixture not exact: %d, %v", len(before), err)
			}
			t.Logf("production-created ledger bytes=%d, headroom=%d", len(before), tc.gap)
			if tc.name == "legacy-full-ledger" {
				legacyLapseFixture(t, f.store, false)
			}
			if got, err := f.store.Admitted(id, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
				t.Fatal("approved positive control", err)
			}
			if tc.name == "legacy-full-ledger" {
				migrated, err := os.ReadFile(statePath)
				if err != nil || string(migrated) != string(before) {
					t.Fatal("same-size legacy migration changed registration, review or approval", err)
				}
			}
			if err := os.Chmod(f.root, 0100); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Chmod(f.root, 0700) }()
			_, loss := f.store.Inspect(id)
			// Continue through a failed retention attempt as well: the actual
			// regression witness is forbidden authority after restoration.
			if !isCode(loss, "project_registration_mismatch") && !isCode(loss, "project_context_drift") {
				t.Fatal("root loss did not refuse", loss)
			}
			if err := os.Chmod(f.root, 0700); err != nil {
				t.Fatal(err)
			}
			fresh := Store{Home: f.store.Home}
			if got, err := fresh.Admitted(id, []Selection{{"guide", false}}); !isCode(err, "project_registration_mismatch") {
				t.Fatalf("capacity failure revived old authority: items=%d, error=%v", len(got.Items), err)
			}
			required(t, loss, "project_registration_mismatch")
			_, err = fresh.Inspect(sibling.CheckoutID)
			required(t, err, "project_registration_mismatch")
			after, err := os.ReadFile(statePath)
			if err != nil || string(after) != string(before) {
				t.Fatal("lapse retention changed the full ledger", err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestRootLapseAtCapacity$")
			cmd.Env = append(os.Environ(), "CURATOR_LAPSE_CHILD=1", "CURATOR_LAPSE_HOME="+f.store.Home, "CURATOR_LAPSE_ID="+id)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fresh-process lapse read: %v\n%s", err, out)
			}
			// The unrelated root remains usable, and releasing only its large
			// preview must not clear either old slot's retained lapse.
			if _, err := fresh.Preview(r.CheckoutID, Declaration{1, []Request{}}); err != nil {
				t.Fatal("release unrelated filler preview", err)
			}
			_, err = fresh.Admitted(id, []Selection{{"guide", false}})
			required(t, err, "project_registration_mismatch")
			rebound, err := fresh.Rebind(id, f.root)
			if err != nil || rebound.CheckoutID == id {
				t.Fatal("fresh-ID recovery control", err)
			}
			_, err = fresh.Admitted(rebound.CheckoutID, []Selection{{"guide", false}})
			required(t, err, "project_item_unapproved")
			f.registration = rebound
			packageApproved(t, f, false)
			if got, err := fresh.Admitted(rebound.CheckoutID, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
				t.Fatal("new approval control", err)
			}
			_, err = fresh.Inspect(sibling.CheckoutID)
			required(t, err, "project_registration_mismatch")
		})
	}
}

// Claim: malformed, noncanonical, unrelated, special and oversized companion
// records cannot be mistaken for no lapse. A valid sibling record is readable,
// and pre-companion ledgers with inline lapses retain their refusal semantics.
func TestRetainedLapseRecordRefusesMalformed(t *testing.T) {
	for _, kind := range []string{"absent", "unknown", "alias", "missing", "version", "duplicate", "unsorted", "foreign", "noncanonical", "oversized", "symlink", "directory"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			p := filepath.Join(f.store.Home, "project-context/lapses.json")
			b, err := canonical(rootLapses{1, []string{id}})
			if err != nil {
				t.Fatal(err)
			}
			write(t, p, string(b))
			_, err = f.store.Inspect(id)
			required(t, err, "project_registration_mismatch")
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			empty, _ := canonical(rootLapses{1, []string{}})
			write(t, p, string(empty))
			if _, err := f.store.Inspect(id); err != nil {
				t.Fatal("empty companion positive control", err)
			}
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "absent":
				// Once version 2 requires a companion, deletion is not first use.
			case "unknown":
				write(t, p, strings.Replace(string(b), `"version":1`, `"version":1,"extra":true`, 1))
			case "alias":
				write(t, p, strings.Replace(string(b), `checkout_ids`, `Checkout_ids`, 1))
			case "missing":
				write(t, p, `{"version":1}`)
			case "version":
				write(t, p, strings.Replace(string(b), `"version":1`, `"version":2`, 1))
			case "duplicate":
				bad, _ := canonical(rootLapses{1, []string{id, id}})
				write(t, p, string(bad))
			case "unsorted":
				write(t, p, string(empty))
				r, err := f.store.Register("fixture", f.root, "base", "claude_code")
				if err != nil {
					t.Fatal(err)
				}
				ids := []string{id, r.CheckoutID}
				if ids[0] < ids[1] {
					ids[0], ids[1] = ids[1], ids[0]
				}
				bad, _ := canonical(rootLapses{1, ids})
				write(t, p, string(bad))
			case "foreign":
				bad, _ := canonical(rootLapses{1, []string{newID()}})
				write(t, p, string(bad))
			case "noncanonical":
				write(t, p, string(b)+"\n")
			case "oversized":
				write(t, p, strings.Repeat(" ", maxLapseRecord+1))
			case "symlink":
				target := filepath.Join(filepath.Dir(f.root), "lapses.json")
				write(t, target, string(b))
				if err := os.Symlink(target, p); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(p, 0700); err != nil {
					t.Fatal(err)
				}
			}
			_, err = f.store.Inspect(id)
			required(t, err, "project_context_drift")
		})
	}
	// This fixture models the previous writer's canonical inline representation.
	f := newFixture(t)
	legacyLapseFixture(t, f.store, true)
	_, err := f.store.Inspect(f.registration.CheckoutID)
	required(t, err, "project_registration_mismatch")
}

func legacyLapseFixture(t *testing.T, store Store, lapsed bool) {
	t.Helper()
	op, err := store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	op.state.Version = 1
	for id, r := range op.state.Registrations {
		r.Lapsed = lapsed
		op.state.Registrations[id] = r
	}
	err = op.save()
	op.close()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(store.Home, "project-context/lapses.json")); err != nil {
		t.Fatal(err)
	}
}

// Claim: a synthetic valid maximum-size registration envelope can retain all
// 256 lapse IDs in less than the companion's independent bound, then read every
// lapse through a fresh Store. This checks the bound, not registration policy.
func TestMaximumLapseRecordFits(t *testing.T) {
	f := newFixture(t)
	op, err := f.store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	defer op.close()
	template := op.state.Registrations[f.registration.CheckoutID]
	op.state.Registrations = map[string]registered{}
	for len(op.state.Registrations) < 256 {
		r := template
		id := "p" + strings.Repeat("x", 30) + newID()
		r.Record.CheckoutID = id
		r.Record.HomeRef = "home:" + id
		r.Record.WorktreeRef = "worktree:" + id
		op.state.Registrations[id] = r
	}
	if err := op.save(); err != nil {
		t.Fatal("maximum envelope control", err)
	}
	for id, r := range op.state.Registrations {
		r.Lapsed = true
		op.state.Registrations[id] = r
	}
	if err := op.retainLapses(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(op.root, "lapses.json"))
	if err != nil || len(b) >= maxLapseRecord {
		t.Fatalf("maximum lapse record: %d bytes, %v", len(b), err)
	}
	t.Logf("256 lapse IDs require %d bytes under %d-byte bound", len(b), maxLapseRecord)
	op.close()
	fresh, err := f.store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.close()
	for _, r := range fresh.state.Registrations {
		if !r.Lapsed {
			t.Fatal("maximum companion dropped a lapse")
		}
	}
}
