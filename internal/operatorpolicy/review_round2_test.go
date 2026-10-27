package operatorpolicy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func r2Grant(t *testing.T, f fixture, id, approval string, rev uint64) SignedCommand {
	t.Helper()
	c := f.command.Command
	p := *c.Policy
	s := *c.Selection
	c.Policy = &p
	c.Selection = &s
	c.EntryID = id
	p.ID = id
	c.EntryRevision = rev
	c.Issuer.Approval = digest([]byte(approval))
	p.Issuer = c.Issuer
	c.Time = time.Now().UTC().Format(time.RFC3339)
	p.Time = c.Time
	return signed(t, c, f.operator, f.actor)
}
func r2Revoke(t *testing.T, f fixture, rev uint64) SignedCommand {
	t.Helper()
	c := Command{Kind: "operator-policy-command/1", Operation: "revoke", Issuer: f.command.Command.Issuer, EntryID: "build", EntryRevision: rev, Time: time.Now().UTC().Format(time.RFC3339), Reason: "Revoke fixture"}
	c.Issuer.Approval = digest([]byte(fmt.Sprintf("revoke-%d", rev)))
	return signed(t, c, f.operator, f.actor)
}

// Claim: removing/recreating an entry through real commands retains its audit
// revision across two cycles; stale/forged counters do not write state.
func TestR2EntryDeletionRecreation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	approve(t, f)
	for cycle := uint64(0); cycle < 2; cycle++ {
		rev := 1 + 2*cycle
		pending := r2Grant(t, f, "build", fmt.Sprintf("pending-%d", cycle), rev)
		if _, e := f.store.Apply(ctx, r2Revoke(t, f, rev)); e != nil {
			t.Fatal(e)
		}
		st, e := f.store.List(ctx)
		if e != nil || len(st.Policies) != 0 || st.EntryRevisions["build"] != rev+1 {
			t.Fatal("revoke did not retain revision", e)
		}
		before := stateBytes(t, f)
		_, e = f.store.Apply(ctx, pending)
		require(t, e, "workspace_trust_policy_authority")
		unchanged(t, f, before)
		forged := pending
		forged.Command.EntryRevision = rev + 1
		_, e = f.store.Apply(ctx, forged)
		require(t, e, "workspace_trust_policy_authority")
		unchanged(t, f, before)
		if _, e = f.store.Apply(ctx, r2Grant(t, f, "build", fmt.Sprintf("fresh-%d", cycle), rev+1)); e != nil {
			t.Fatal(e)
		}
		if got := entryRevision(t, f, "build"); got != rev+2 {
			t.Fatalf("revision=%d", got)
		}
	}
}

// Claim: normal serialization and a later unrelated signed store mutation
// preserve a terminal lapse; a fresh Store instance cannot revive it.
func TestR2LapseSurvivesRewrite(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	approve(t, f)
	moved := f.selection.Path + "-moved"
	if e := os.Rename(f.selection.Path, moved); e != nil {
		t.Fatal(e)
	}
	_, e := f.store.Read(ctx, f.selection)
	require(t, e, "workspace_trust_symlink")
	if e = os.Rename(moved, f.selection.Path); e != nil {
		t.Fatal(e)
	}
	var st ledger
	if e = decode(stateBytes(t, f), &st); e != nil {
		t.Fatal(e)
	}
	b, e := canonical(st)
	if e != nil {
		t.Fatal(e)
	}
	write(t, filepath.Join(f.store.Home, "operator-policy/state.json"), b)
	c := r2Grant(t, f, "other", "rewrite-other", 0).Command
	c.Selection.Lane = "lane:other"
	c.Policy.Lanes = []string{c.Selection.Lane}
	if _, e = f.store.Apply(ctx, signed(t, c, f.operator, f.actor)); e != nil {
		t.Fatal(e)
	}
	before := stateBytes(t, f)
	out, e := (Store{Home: f.store.Home}).Read(ctx, f.selection)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatal("lapsed delivery")
	}
	unchanged(t, f, before)
	list, e := f.store.List(ctx)
	if e != nil || list.Events[0].State.Status != "lapsed" || len(list.Events[0].History) != 1 {
		t.Fatal("lapse lost", e)
	}
}

// Claim under attack: a pre-lapse unused approval must not renew consent after
// observed identity loss. Future and equal-second signed times remove accidental
// clock-order protection; no owner key is used after lapse in either stale case.
func TestR2PendingApprovalAcrossLapse(t *testing.T) {
	for _, tc := range []struct{ fresh, sameSecond bool }{{true, false}, {false, false}, {false, true}} {
		fresh := tc.fresh
		name := "pre-lapse-signature"
		if fresh {
			name = "fresh-post-lapse-signature"
		} else if tc.sameSecond {
			name = "pre-lapse-same-second"
		}
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			ctx := context.Background()
			approve(t, f)
			c := r2Grant(t, f, "build", "pending-lapse-control", 1).Command
			at := time.Now().UTC().Add(5 * time.Second).Truncate(time.Second)
			if tc.sameSecond {
				// Align only the fixture so a current-time approval and lapse fit
				// within the public timestamp's one-second precision.
				time.Sleep(time.Until(time.Now().Truncate(time.Second).Add(time.Second)) + 20*time.Millisecond)
				at = time.Now().UTC().Truncate(time.Second)
			}
			c.Time = at.Format(time.RFC3339)
			c.Policy.Time = c.Time
			var pending SignedCommand
			if !fresh {
				pending = signed(t, c, f.operator, f.actor)
			}
			moved := f.selection.Path + "-moved"
			if e := os.Rename(f.selection.Path, moved); e != nil {
				t.Fatal(e)
			}
			_, e := f.store.Read(ctx, f.selection)
			require(t, e, "workspace_trust_symlink")
			if e = os.Rename(moved, f.selection.Path); e != nil {
				t.Fatal(e)
			}
			_, e = f.store.Read(ctx, f.selection)
			require(t, e, "workspace_untrusted")
			list, e := f.store.List(ctx)
			if e != nil {
				t.Fatal(e)
			}
			lapsedAt, e := timestamp(list.Events[0].State.Time)
			if e != nil || lapsedAt.After(at) {
				t.Fatal("timing control did not establish pending timestamp >= lapse", e)
			}
			if fresh {
				c.EntryRevision = entryRevision(t, f, c.EntryID)
				pending = signed(t, c, f.operator, f.actor)
			}
			time.Sleep(time.Until(at) + 10*time.Millisecond)
			before := stateBytes(t, f)
			out, e := f.store.Apply(ctx, pending)
			if fresh {
				if e != nil || out == nil {
					t.Fatal("fresh post-lapse approval rejected", e)
				}
				return
			}
			if e == nil || !strings.HasPrefix(e.Error(), "workspace_trust_policy_authority:") || out != nil {
				t.Errorf("unused pre-lapse approval: event=%v error=%v; want workspace_trust_policy_authority and no event", out != nil, e)
			}
			if e == nil {
				list, _ := f.store.List(ctx)
				t.Logf("revision=%d event states=%s,%s", list.EntryRevisions["build"], list.Events[0].State.Status, list.Events[1].State.Status)
			}
			unchanged(t, f, before)
		})
	}
}

// Characterization, NOT an anti-rollback claim: deleting the allocation or
// restoring an earlier authenticated state can rewind this same-UID store.
// Missing an allocated state file alone refuses; preservation of old bytes in
// these fixture attacks shows exactly which external boundary is absent.
func TestR2StoreRollbackBoundary(t *testing.T) {
	for _, kind := range []string{"missing-state", "remove-allocation", "restore-backup", "restore-before-lapse"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			ctx := context.Background()
			approve(t, f)
			backup := stateBytes(t, f)
			if kind == "restore-before-lapse" {
				moved := f.selection.Path + "-moved"
				if e := os.Rename(f.selection.Path, moved); e != nil {
					t.Fatal(e)
				}
				_, e := f.store.Read(ctx, f.selection)
				require(t, e, "workspace_trust_symlink")
				if e = os.Rename(moved, f.selection.Path); e != nil {
					t.Fatal(e)
				}
				_, e = f.store.Read(ctx, f.selection)
				require(t, e, "workspace_untrusted")
			} else {
				if _, e := f.store.Apply(ctx, r2Revoke(t, f, 1)); e != nil {
					t.Fatal(e)
				}
				_, e := f.store.Read(ctx, f.selection)
				require(t, e, "workspace_trust_grant_revoked")
			}
			state := filepath.Join(f.store.Home, "operator-policy/state.json")
			if kind == "missing-state" {
				if e := os.Rename(state, state+".saved"); e != nil {
					t.Fatal(e)
				}
				_, e := f.store.Read(ctx, f.selection)
				require(t, e, "workspace_trust_ledger_unreadable")
				_, e = f.store.Apply(ctx, f.command)
				require(t, e, "workspace_trust_ledger_unreadable")
				var st ledger
				if e = decode(backup, &st); e != nil {
					t.Fatal(e)
				}
				e = f.store.PinOperator(ctx, "principal:operator", st.Keys["principal:operator"])
				require(t, e, "workspace_trust_ledger_unreadable")
				return
			}
			if kind == "remove-allocation" {
				dir := filepath.Dir(state)
				if e := os.Rename(dir, dir+".saved"); e != nil {
					t.Fatal(e)
				}
				_, e := f.store.Read(ctx, f.selection)
				require(t, e, "workspace_untrusted")
				if e = f.store.PinOperator(ctx, "principal:operator", f.command.Command.Policy.Issuer.Principal); e == nil {
					t.Fatal("invalid key accepted")
				}
				// Decode the original pinned public key from the retained fixture backup.
				var st ledger
				if e = decode(backup, &st); e != nil {
					t.Fatal(e)
				}
				if e = f.store.PinOperator(ctx, "principal:operator", st.Keys["principal:operator"]); e != nil {
					t.Fatal(e)
				}
				if got := entryRevision(t, f, "build"); got != 0 {
					t.Fatalf("expected observed reset to zero, got %d", got)
				}
				if _, e = f.store.Apply(ctx, f.command); e != nil {
					t.Fatal("rollback characterization changed", e)
				}
			} else {
				write(t, state, backup)
			}
			list, e := f.store.List(ctx)
			if e != nil || list.EntryRevisions["build"] != 1 || list.Events[0].State.Status != "active" {
				t.Fatal("expected rollback boundary witness", e)
			}
			out, e := f.store.read(ctx, f.selection, observeGranted)
			if e != nil || out == nil {
				t.Fatal("expected core rollback witness", e)
			}
			_, e = f.store.Read(ctx, f.selection)
			require(t, e, "workspace_trust_adapter_unqualified")
			t.Log("OBSERVED same-UID rollback: revision=1, old event active; production adapter gate still closed")
		})
	}
}

type r2WorkerResult struct {
	Error string
	Event bool
}

// Test helper runs real Apply in another OS process, with separately acquired
// manager lock and a barrier shared only within the disposable fixture.
func TestR2ConcurrentChild(t *testing.T) {
	home := os.Getenv("PR195_R2_HOME")
	if home == "" {
		return
	}
	input := os.Getenv("PR195_R2_INPUT")
	b, e := os.ReadFile(input)
	if e != nil {
		t.Fatal(e)
	}
	var c SignedCommand
	if e = json.Unmarshal(b, &c); e != nil {
		t.Fatal(e)
	}
	ready := os.Getenv("PR195_R2_READY")
	if e = os.WriteFile(ready, []byte("ready"), 0600); e != nil {
		t.Fatal(e)
	}
	barrier := os.Getenv("PR195_R2_START")
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, e = os.Stat(barrier); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("barrier timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ev, e := (Store{Home: home}).Apply(ctx, c)
	r := r2WorkerResult{Event: ev != nil}
	if e != nil {
		r.Error = e.Error()
	}
	b, e = json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(os.Getenv("PR195_R2_OUTPUT"), b, 0600); e != nil {
		t.Fatal(e)
	}
}
func r2Compete(t *testing.T, f fixture, a, b SignedCommand) []r2WorkerResult {
	t.Helper()
	dir := t.TempDir()
	start := filepath.Join(dir, "start")
	cmds := []*exec.Cmd{}
	for i, c := range []SignedCommand{a, b} {
		stem := filepath.Join(dir, fmt.Sprint(i))
		payload, e := json.Marshal(c)
		if e != nil {
			t.Fatal(e)
		}
		write(t, stem+".json", payload)
		cmd := exec.Command(os.Args[0], "-test.run=^TestR2ConcurrentChild$")
		cmd.Env = append(os.Environ(), "PR195_R2_HOME="+f.store.Home, "PR195_R2_INPUT="+stem+".json", "PR195_R2_READY="+stem+".ready", "PR195_R2_START="+start, "PR195_R2_OUTPUT="+stem+".out")
		log, e := os.Create(stem + ".log")
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() { _ = log.Close() })
		cmd.Stdout = log
		cmd.Stderr = log
		if e = cmd.Start(); e != nil {
			t.Fatal(e)
		}
		t.Logf("fixture writer pid=%d", cmd.Process.Pid)
		cmds = append(cmds, cmd)
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		_, a := os.Stat(filepath.Join(dir, "0.ready"))
		_, b := os.Stat(filepath.Join(dir, "1.ready"))
		if a == nil && b == nil {
			break
		}
		if time.Now().After(deadline) {
			for _, c := range cmds {
				_ = c.Process.Kill()
				_ = c.Wait()
			}
			t.Fatal("ready timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}
	write(t, start, []byte("start"))
	results := []r2WorkerResult{}
	for i, cmd := range cmds {
		if e := cmd.Wait(); e != nil {
			t.Fatal("child failed", e)
		}
		b, e := os.ReadFile(filepath.Join(dir, fmt.Sprint(i)+".out"))
		if e != nil {
			t.Fatal(e)
		}
		var r r2WorkerResult
		if e = json.Unmarshal(b, &r); e != nil {
			t.Fatal(e)
		}
		results = append(results, r)
	}
	return results
}

// Claim: two OS writers sharing one revision or one exact selection serialize
// to one accepted command and one precise authority refusal, without lost audit.
// A refreshed, newly signed command then succeeds; a stale retry remains inert.
func TestR2ConcurrentWriters(t *testing.T) {
	for _, kind := range []string{"revoke-grant", "ambiguous-grants"} {
		for attempt := 0; attempt < 3; attempt++ {
			t.Run(fmt.Sprintf("%s-%d", kind, attempt), func(t *testing.T) {
				f := newFixture(t)
				ctx := context.Background()
				var a, b SignedCommand
				if kind == "revoke-grant" {
					approve(t, f)
					a = r2Revoke(t, f, 1)
					b = r2Grant(t, f, "build", "competing-grant", 1)
				} else {
					a = f.command
					b = r2Grant(t, f, "second", "competing-entry", 0)
				}
				results := r2Compete(t, f, a, b)
				accepted := 0
				for _, r := range results {
					if r.Error == "" {
						accepted++
					} else if !strings.HasPrefix(r.Error, "workspace_trust_policy_authority:") || r.Event {
						t.Fatalf("wrong losing result: %+v", r)
					}
				}
				if accepted != 1 {
					t.Fatalf("accepted writers=%d, results=%+v", accepted, results)
				}
				list, e := f.store.List(ctx)
				if e != nil {
					t.Fatal(e)
				}
				if kind == "revoke-grant" {
					if list.EntryRevisions["build"] != 2 {
						t.Fatal("lost audit revision")
					}
					before := stateBytes(t, f)
					loser := a
					if results[1].Error != "" {
						loser = b
					}
					_, e = f.store.Apply(ctx, loser)
					require(t, e, "workspace_trust_policy_authority")
					unchanged(t, f, before)
					if _, e = f.store.Apply(ctx, r2Grant(t, f, "build", "post-race-fresh", 2)); e != nil {
						t.Fatal("fresh post-race approval", e)
					}
				} else {
					if len(list.Events) != 1 || len(list.Policies) != 1 || list.EntryRevisions["build"]+list.EntryRevisions["second"] != 1 {
						t.Fatal("ambiguous writes or lost audit")
					}
					if out, e := f.store.read(ctx, f.selection, observeGranted); e != nil || out == nil {
						t.Fatal("single winner not readable", e)
					}
				}
			})
		}
	}
}

// Claim: each terminal lapse advances the boundary once. Repeated denied reads
// and inspection do not advance it again; moving lapse into history on a fresh
// grant and then revoking that grant retain both ordering transitions.
func TestPolicyLapseRevisionCountsOnce(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	approve(t, f)
	moved := f.selection.Path + "-moved"
	if e := os.Rename(f.selection.Path, moved); e != nil {
		t.Fatal(e)
	}
	_, e := f.store.Read(ctx, f.selection)
	require(t, e, "workspace_trust_symlink")
	lapsed := stateBytes(t, f)
	for i := 0; i < 2; i++ {
		if rev := entryRevision(t, f, "build"); rev != 2 {
			t.Fatalf("revision=%d want 2", rev)
		}
		_, e = f.store.Read(ctx, f.selection)
		require(t, e, "workspace_untrusted")
		unchanged(t, f, lapsed)
	}
	if e = os.Rename(moved, f.selection.Path); e != nil {
		t.Fatal(e)
	}
	fresh := r2Grant(t, f, "build", "fresh-counting-control", entryRevision(t, f, "build"))
	ev, e := f.store.Apply(ctx, fresh)
	if e != nil || ev == nil {
		t.Fatal("fresh approval", e)
	}
	if rev := entryRevision(t, f, "build"); rev != 3 {
		t.Fatalf("historical lapse lost: revision=%d want 3", rev)
	}
	if delivery, e := f.store.read(ctx, f.selection, observeGranted); e != nil || delivery == nil || delivery.GrantID != ev.ID {
		t.Fatal("fresh consent", e)
	}
	if _, e = f.store.Apply(ctx, r2Revoke(t, f, 3)); e != nil {
		t.Fatal(e)
	}
	if rev := entryRevision(t, f, "build"); rev != 4 {
		t.Fatalf("revision=%d want 4", rev)
	}
	_, e = f.store.Read(ctx, f.selection)
	require(t, e, "workspace_trust_grant_revoked")
}
