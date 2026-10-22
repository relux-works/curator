package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual GC recovery entry point applies the project-lease commit policy:
// intact prepared intent commits and pins runtime; missing runtime rolls back
// the admission before GC and creates neither its ledger nor its binding.
// These are isolated synthetic protected journal records, not input admission.
func TestGCRecoversProjectLeaseArtifacts(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "present"}[present], func(t *testing.T) {
			home := gcHome(t)
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			home = lm.Home()
			lock, err := lm.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			if present {
				if err = os.MkdirAll(filepath.Join(home, "runtime", "tool", "next"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			b := projectlease.Binding{Kind: "project-lease/1", LeaseID: "next", ContextRef: "composition:next", ContextSHA256: strings.Repeat("a", 64), CheckoutID: "checkout-next", Environment: "codex_cli", HomeRef: "home:next", OwnerRef: "principal:operator", SessionRef: "session:next", Generation: 1}
			c := projectlease.Composition{Ref: b.ContextRef, SHA256: b.ContextSHA256, CheckoutID: b.CheckoutID, Profile: "default", Environment: b.Environment, HomeRef: b.HomeRef, ConcurrentState: true, BuildKeys: []string{}, RuntimeEntries: []string{"tool/next"}, Generated: map[string]projectlease.Surface{}}
			encode := func(v any) []byte {
				t.Helper()
				raw, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				decoder := json.NewDecoder(bytes.NewReader(raw))
				decoder.UseNumber()
				var value any
				if err = decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				raw, err = protocoljson.MarshalCanonical(value)
				if err != nil {
					t.Fatal(err)
				}
				return raw
			}
			state := map[string]any{"kind": "project-lease-state/1", "leases": map[string]any{"next": map[string]any{"binding": b, "composition": c, "state": "prepared", "process_generation": ""}}, "views": map[string]any{}}
			ledger := filepath.Join(home, "state", "project-leases", "v1", "ledger.json")
			binding := filepath.Join(home, "state", "project-leases", "v1", "bindings", "next.json")
			plan := transaction.Plan{TransactionID: "project-lease-pending", ProjectIdentity: "project-leases"}
			for i, target := range []struct {
				path string
				raw  []byte
			}{{ledger, encode(state)}, {binding, encode(b)}} {
				if err = os.MkdirAll(filepath.Dir(target.path), 0700); err != nil {
					t.Fatal(err)
				}
				staged := filepath.Join(t.TempDir(), string(rune('a'+i)))
				if err = os.WriteFile(staged, target.raw, 0600); err != nil {
					t.Fatal(err)
				}
				plan.Targets = append(plan.Targets, transaction.Target{Class: "project-lease", Identifier: target.path, LivePath: target.path, StagedSource: staged, PreimageDigest: transaction.DigestAbsent})
			}
			engine, err := transaction.New(home)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Prepare(lock, plan); err != nil {
				t.Fatal(err)
			}
			_, err = collectUnderLock(home, lock, envprofile.Policy{})
			if present {
				if err != nil {
					t.Fatal("GC refused intact prepared lease", err)
				}
				for _, path := range []string{ledger, binding, filepath.Join(home, "runtime", "tool", "next")} {
					if _, err = os.Stat(path); err != nil {
						t.Fatal("GC lost recovered lease root", err)
					}
				}
			} else {
				if !errors.Is(err, projectlease.ErrArtifactsUnavailable) {
					t.Fatalf("GC recovery committed missing artifact or wrong refusal: %v", err)
				}
				for _, path := range []string{ledger, binding} {
					if _, err = os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
						t.Fatal("GC failed admission left publication", err)
					}
				}
			}
		})
	}
}
