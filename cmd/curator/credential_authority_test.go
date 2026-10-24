package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// Global declaration, install and GC commands refuse node authority before locks,
// pending recovery or publication. Per-home add/install complete real intent;
// remove/init mutate their declaration without invoking transaction recovery.
// GC also refuses without creating lock state when no journal/lock exists.
// Inventory snapshots include manager and synthetic native bytes/types/modes;
// timestamps are outside the claim.
func TestGlobalCredentialAuthorityEntries(t *testing.T) {
	for _, command := range []string{"add", "install", "remove", "init", "gc", "gc-no-intent"} {
		for _, node := range []bool{true, false} {
			label := "per-home"
			if node {
				label = "node"
			}
			t.Run(command+"/"+label, func(t *testing.T) {
				fx := newGlobalLockCLIFixture(t)
				fx.writeLegacyGlobalDeclaration(t)
				target := filepath.Join(fx.home, "state", "authority-witness")
				if command != "gc-no-intent" {
					if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(target, []byte("old intent"), 0600); err != nil {
						t.Fatal(err)
					}
					manager, e := managerlock.New(fx.home)
					if e != nil {
						t.Fatal(e)
					}
					lock, e := manager.AcquireHomeOnly(context.Background(), false)
					if e != nil {
						t.Fatal(e)
					}
					engine, e := transaction.New(fx.home)
					if e != nil {
						t.Fatal(e)
					}
					staged := filepath.Join(t.TempDir(), "intent")
					if e = os.WriteFile(staged, []byte("recovered intent"), 0600); e != nil {
						t.Fatal(e)
					}
					digest, e := transaction.DigestTarget(transaction.KindBytes, target)
					if e != nil {
						t.Fatal(e)
					}
					_, e = engine.Prepare(lock, transaction.Plan{TransactionID: "authority-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: digest}}})
					if e != nil {
						t.Fatal(e)
					}
					if pending, e := engine.PendingRecovery(lock); e != nil || !pending {
						t.Fatalf("pending setup: %v %v", pending, e)
					}
					if e = lock.Close(); e != nil {
						t.Fatal(e)
					}
				}
				raw, e := os.ReadFile("../../internal/envmarker/testdata/seed-c/node.json")
				if e != nil {
					t.Fatal(e)
				}
				if !node {
					var doc map[string]any
					if e = json.Unmarshal(raw, &doc); e != nil {
						t.Fatal(e)
					}
					doc["credential_source"] = map[string]any{"mode": "per-home"}
					doc["seeds"] = []string{"config.toml"}
					raw, e = json.Marshal(doc)
					if e != nil {
						t.Fatal(e)
					}
				}
				if _, e = envmarker.Parse(raw); e != nil {
					t.Fatal(e)
				}
				dir := envprofile.ManagedHomeDir(fx.home, "orphan", "codex_cli")
				if e = privatedir.MakeAll(dir); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, envmarker.Name), raw, 0600); e != nil {
					t.Fatal(e)
				}
				if command == "gc-no-intent" {
					if _, err := os.Lstat(filepath.Join(fx.home, "state", "locks")); !os.IsNotExist(err) {
						t.Fatalf("GC fresh-home fixture already has lock state: %v", err)
					}
				}
				before := nativeOwnerSnapshot(t, fx.home)
				nativeBefore := nativeOwnerSnapshot(t, os.Getenv("HOME"))
				args := []string{"global", command}
				switch command {
				case "add":
					args = append(args, "new-skill", "--git", fx.gitURL, "--tag", "v1")
				case "remove":
					args = append(args, "new-skill")
				case "gc", "gc-no-intent":
					args = []string{"gc"}
				}
				code, stdout, stderr := fx.run(t, args...)
				if node {
					if code != exitFail || !strings.Contains(stderr, "credential_injection_unavailable:") {
						t.Fatalf("node global diagnostic: %d %s %s", code, stdout, stderr)
					}
					if !reflect.DeepEqual(before, nativeOwnerSnapshot(t, fx.home)) || !reflect.DeepEqual(nativeBefore, nativeOwnerSnapshot(t, os.Getenv("HOME"))) {
						t.Fatal("global node refusal mutated state")
					}
				} else {
					if code != exitOK {
						t.Fatalf("per-home global positive: %d %s %s", code, stdout, stderr)
					}
					if command == "gc-no-intent" {
						return
					}
					want := "old intent"
					if command == "add" || command == "install" || command == "gc" {
						want = "recovered intent"
					}
					got, e := os.ReadFile(target)
					if e != nil || string(got) != want {
						t.Fatalf("per-home recovery intent: %q %v", got, e)
					}
				}
			})
		}
	}
}
