package install

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Both install APIs can recover manager-wide transactions even when their
// selected scope has no environment marker. An orphaned node carrier blocks
// that recovery, compilation and publication. Per-home controls install the
// declared skill and actually recover the prepared intent. Snapshots cover
// bytes, inventory and link targets, not timestamps or permission bits.
func TestInstallCredentialAuthority(t *testing.T) {
	for _, scope := range []string{"global", "project"} {
		for _, node := range []bool{true, false} {
			label := "per-home"
			if node {
				label = "node"
			}
			t.Run(scope+"/"+label, func(t *testing.T) {
				e := newEnv(t)
				e.skill("skill-a")
				e.declare("skill-a")
				e.globalDeclare("skill-a")
				target := filepath.Join(e.home, "state", "authority-witness")
				e.write(e.home, "state/authority-witness", "old intent")
				manager, err := managerlock.New(e.home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := manager.AcquireHomeOnly(context.Background(), false)
				if err != nil {
					t.Fatal(err)
				}
				engine, err := transaction.New(e.home)
				if err != nil {
					t.Fatal(err)
				}
				staged := filepath.Join(t.TempDir(), "intent")
				if err = os.WriteFile(staged, []byte("recovered intent"), 0600); err != nil {
					t.Fatal(err)
				}
				digest, err := transaction.DigestTarget(transaction.KindBytes, target)
				if err != nil {
					t.Fatal(err)
				}
				_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "authority-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: target, Kind: transaction.KindBytes, LivePath: target, StagedSource: staged, PreimageDigest: digest}}})
				if err != nil {
					t.Fatal(err)
				}
				if pending, err := engine.PendingRecovery(lock); err != nil || !pending {
					t.Fatalf("pending setup: %v %v", pending, err)
				}
				if err = lock.Close(); err != nil {
					t.Fatal(err)
				}
				raw, err := os.ReadFile("../envmarker/testdata/seed-c/node.json")
				if err != nil {
					t.Fatal(err)
				}
				if !node {
					var doc map[string]any
					if err = json.Unmarshal(raw, &doc); err != nil {
						t.Fatal(err)
					}
					doc["credential_source"] = map[string]any{"mode": "per-home"}
					doc["seeds"] = []string{"config.toml"}
					raw, err = json.Marshal(doc)
					if err != nil {
						t.Fatal(err)
					}
				}
				if _, err = envmarker.Parse(raw); err != nil {
					t.Fatal(err)
				}
				dir := envprofile.ManagedHomeDir(e.home, "orphan", "codex_cli")
				if err = os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, envmarker.Name), raw, 0600); err != nil {
					t.Fatal(err)
				}
				before, projectBefore := entryDigest(e.home), entryDigest(e.project)
				recoveries := 0
				opts := Options{Platform: installPlatform(), Commit: CommitDeps{Hooks: transaction.Hooks{ObserveRecovery: func() { recoveries++ }}}}
				var result Result
				if scope == "global" {
					result = Global(e.cfg, t.TempDir(), opts)
				} else {
					result = Project(e.cfg, e.project, "fixture", opts)
				}
				if node {
					if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, " "), "credential_injection_unavailable:") {
						t.Fatalf("node install refusal: %+v", result)
					}
					if recoveries != 0 {
						t.Errorf("node install entered recovery: %d", recoveries)
					}
					if before != entryDigest(e.home) || projectBefore != entryDigest(e.project) {
						t.Error("node install refusal mutated state")
					}
				} else {
					if result.Status != "ok" || recoveries == 0 {
						t.Fatalf("per-home positive: %+v; recovery calls %d", result, recoveries)
					}
					got, err := os.ReadFile(target)
					if err != nil || string(got) != "recovered intent" {
						t.Fatalf("per-home recovery intent: %q %v", got, err)
					}
				}
			})
		}
	}
}
