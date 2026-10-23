package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/transaction"
)

// The default envprofile recovery engine applies the project-lease policy instead
// of publishing an invalid lease journal with no desired ledger. A journal
// owned by profiles retains the prior recovery behavior and commits normally.
func TestDefaultRecoveryChecksLeaseJournal(t *testing.T) {
	for _, owner := range []string{"profiles", "project-leases"} {
		t.Run(owner, func(t *testing.T) {
			home := t.TempDir()
			if err := os.Chmod(home, 0700); err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			home = lm.Home()
			lock, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(home, "owned.txt")
			stage := filepath.Join(t.TempDir(), "stage")
			if err = os.WriteFile(target, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(stage, []byte("new"), 0600); err != nil {
				t.Fatal(err)
			}
			pre, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(home)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Prepare(lock, transaction.Plan{TransactionID: "recovery-owner", ProjectIdentity: owner, Targets: []transaction.Target{{Class: "project-lease", Identifier: target, LivePath: target, StagedSource: stage, PreimageDigest: pre}}}); err != nil {
				t.Fatal(err)
			}
			if err = lock.Close(); err != nil {
				t.Fatal(err)
			}
			op, err := beginOperation(home)
			if op != nil {
				defer op.close()
			}
			raw, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if owner == "project-leases" {
				if err == nil || !strings.Contains(err.Error(), "lacks its ledger") || string(raw) != "old" {
					t.Fatalf("default lease recovery policy missing: %v bytes=%q", err, raw)
				}
			} else if err != nil || string(raw) != "new" {
				t.Fatal("legacy profile recovery changed", err, string(raw))
			}
		})
	}
}

// Recovery of prepared/active publication leaves runner life unknown; it never
// upgrades to active or releases a reservation. Already unknown/released states
// are controls, and immutable binding bytes survive the lifecycle publication.
func TestRecoveredLeaseLifecycle(t *testing.T) {
	encode := func(v any) ([]byte, error) {
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		return protocoljson.MarshalCanonical(value)
	}

	for _, phase := range []projectlease.Phase{projectlease.Prepared, projectlease.Active, projectlease.Unknown, projectlease.Released} {
		t.Run(string(phase), func(t *testing.T) {
			home := t.TempDir()
			if err := os.Chmod(home, 0700); err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			home = lm.Home()
			lock, err := lm.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(home)
			if err != nil {
				t.Fatal(err)
			}
			b := projectlease.Binding{Kind: "project-lease/1", LeaseID: "next", ContextRef: "composition:next", ContextSHA256: strings.Repeat("a", 64), CheckoutID: "checkout", Environment: "codex_cli", HomeRef: "home:next", OwnerRef: "principal:operator", SessionRef: "session:next", Generation: 1}
			c := projectlease.Composition{Ref: b.ContextRef, SHA256: b.ContextSHA256, CheckoutID: b.CheckoutID, Profile: "default", Environment: b.Environment, HomeRef: b.HomeRef, ConcurrentState: true, BuildKeys: []string{}, RuntimeEntries: []string{}, Generated: map[string]projectlease.Surface{}}
			bindingRaw, err := encode(b)
			if err != nil {
				t.Fatal(err)
			}
			state := map[string]any{"kind": "project-lease-state/1", "leases": map[string]any{"next": map[string]any{"binding": b, "composition": c, "state": phase, "process_generation": "runner:fixture"}}, "views": map[string]any{}}
			raw, err := encode(state)
			if err != nil {
				t.Fatal(err)
			}
			ledger := filepath.Join(home, "state", "project-leases", "v1", "ledger.json")
			binding := filepath.Join(home, "state", "project-leases", "v1", "bindings", "next.json")
			plan := transaction.Plan{TransactionID: "interrupted-lease", ProjectIdentity: "project-leases"}
			for index, item := range []struct {
				path string
				raw  []byte
			}{{binding, bindingRaw}, {ledger, raw}} {
				if err := os.MkdirAll(filepath.Dir(item.path), 0700); err != nil {
					t.Fatal(err)
				}
				source := filepath.Join(t.TempDir(), fmt.Sprint(index))
				if err := os.WriteFile(source, item.raw, 0600); err != nil {
					t.Fatal(err)
				}
				plan.Targets = append(plan.Targets, transaction.Target{Class: "project-lease", Identifier: item.path, LivePath: item.path, StagedSource: source, PreimageDigest: transaction.DigestAbsent})
			}
			if _, err = engine.Prepare(lock, plan); err != nil {
				t.Fatal(err)
			}
			if err = lock.Close(); err != nil {
				t.Fatal(err)
			}
			op, err := beginOperation(home)
			if err != nil {
				t.Fatal(err)
			}
			if err = op.close(); err != nil {
				t.Fatal(err)
			}
			m, err := projectlease.New(home, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			statuses, err := m.Status(t.Context())
			want := projectlease.Unknown
			if phase == projectlease.Released {
				want = projectlease.Released
			}
			if err != nil || len(statuses) != 1 || statuses[0].State != want {
				t.Fatalf("recovered lifecycle remained prepared/active or released: %v %v want=%s", statuses, err, want)
			}
			durable, err := os.ReadFile(ledger)
			var stateDocument struct {
				Leases map[string]struct {
					State projectlease.Phase `json:"state"`
				} `json:"leases"`
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(durable, &stateDocument); err != nil || stateDocument.Leases["next"].State != want {
				t.Fatalf("recovered lifecycle is not durable: %s %v", durable, err)
			}
			journals, err := filepath.Glob(filepath.Join(home, "state", "transactions", "v1", "*.json"))
			if err != nil || len(journals) != 0 {
				t.Fatal("recovery left unsettled intent", journals, err)
			}
			got, err := os.ReadFile(binding)
			if err != nil || string(got) != string(bindingRaw) {
				t.Fatal("recovery changed immutable binding", err)
			}
		})
	}
}
