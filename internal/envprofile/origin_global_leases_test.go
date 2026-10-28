package envprofile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/projectleaserecord"
	"github.com/relux-works/curator/internal/transaction"
)

// Existing P lease metadata is an unchanged retained preimage, not a new P
// reservation. A real Q reservation may CAS the shared ledger only preserving
// P's exact captured entries. Missing/changed P entries refuse before Q's
// session authority, with no changes to L1, receipt or home/native preimages.
// This exercises fixture composition/session transports, not a real launcher.
func TestProvisioningOriginGlobalLeasePreimages(t *testing.T) {
	for _, mode := range []string{"healthy", "changed-p", "missing-p"} {
		t.Run(mode, func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			root := t.TempDir()
			writePackage(t, root, "lease-q-context", "1.0.0", "Q context\n")
			if _, _, _, err := Install(fx.home, InstallOptions{Operand: root, As: "lease-q"}); err != nil {
				t.Fatal(err)
			}
			ap := &originLeaseAuthority{home: fx.home, profile: fx.profile}
			mp, err := projectlease.New(fx.home, ap, ap)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = mp.Acquire(context.Background(), projectlease.Request{LeaseID: "retained-p", ContextRef: "composition:retained-p", ContextSHA256: strings.Repeat("a", 64), OwnerRef: "principal:fixture", SessionRef: "session:retained-p", Generation: 1}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(fx.home, "state", "project-leases", "v1", "ledger.json")
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			old, err := projectleaserecord.DecodeLedger(original)
			if err != nil {
				t.Fatal(err)
			}
			plantGlobalUnmanaged(t, fx)
			before := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
			info, err := GlobalInstall(fx.home, Policy{}, envregistry.DefaultMachineConfig(), resolver)
			var retained *transaction.RetainedGlobalConflict
			if !errors.As(err, &retained) || retained.Cleanup != nil {
				t.Fatalf("actual retained disposition with existing P lease: %v", err)
			}
			identity, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, receipt, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			if err = envauthority.Validate(fx.home); err != nil {
				t.Fatalf("unchanged existing P references invalidate manager: %v", err)
			}
			mutated := old
			if mode == "changed-p" {
				entry := mutated.Leases["retained-p"]
				entry.State = projectleaserecord.Unknown
				mutated.Leases["retained-p"] = entry
			}
			if mode == "missing-p" {
				delete(mutated.Leases, "retained-p")
			}
			if mode != "healthy" {
				raw, err := projectleaserecord.Canonical(mutated)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			pre, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			aq := &originLeaseAuthority{home: fx.home, profile: "lease-q", homeRef: "home:disjoint-q"}
			mq, err := projectlease.New(fx.home, aq, aq)
			if err != nil {
				t.Fatal(err)
			}
			_, err = mq.Acquire(context.Background(), projectlease.Request{LeaseID: "disjoint-q", ContextRef: "composition:disjoint-q", ContextSHA256: strings.Repeat("b", 64), OwnerRef: "principal:fixture", SessionRef: "session:disjoint-q", Generation: 1})
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if mode == "healthy" {
				if err != nil || aq.observations == 0 {
					t.Fatalf("actual Q continuation blocked: observations=%d %v", aq.observations, err)
				}
				next, err := projectleaserecord.DecodeLedger(after)
				if err != nil {
					t.Fatal(err)
				}
				a, _ := json.Marshal(next.Leases["retained-p"])
				b, _ := json.Marshal(old.Leases["retained-p"])
				if string(a) != string(b) || len(next.Leases) != 2 {
					t.Fatal("Q publication changed retained P references")
				}
				observed := ap.observations
				if err := mp.Reconcile(context.Background(), "retained-p"); err == nil || !strings.Contains(err.Error(), "global_retained_conflict") || ap.observations != observed {
					t.Fatalf("P transition reached session continuation: %v", err)
				}
				exerciseRetainedLeaseGap(t, fx, next, path)

			} else if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || aq.observations != 0 || string(pre) != string(after) {
				t.Fatalf("intended retained reference refusal: observations=%d err=%v", aq.observations, err)
			}
			_, now, err := envlifecycle.Read(identity)
			if err != nil || string(now) != string(receipt) {
				t.Fatalf("Q changed retained lifecycle receipt: %v", err)
			}
			_, hash, err := readLock(fx.home, fx.profile)
			if err != nil || hash != info.LockHash {
				t.Fatal("Q changed retained L1")
			}
			if nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != before {
				t.Fatal("Q changed retained home preimages")
			}
		})
	}
}

// The actual shared-ledger rename gap is admitted from exact before/after
// sidecars, not journal existence. Changing its phase must make admission fail;
// restoring that one fixture record allows fresh recovery and disjoint Q work.
func exerciseRetainedLeaseGap(t *testing.T, fx *managedFixture, before *projectleaserecord.Ledger, path string) {
	t.Helper()
	oldP, _ := json.Marshal(before.Leases["retained-p"])
	fresh := before.Leases["disjoint-q"]
	fresh.Binding.LeaseID = "restart-q"
	fresh.Binding.SessionRef = "session:restart-q"
	before.Leases["restart-q"] = fresh
	raw, err := projectleaserecord.Canonical(before)
	if err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(t.TempDir(), "lease.json")
	if err = os.WriteFile(stage, raw, 0600); err != nil {
		t.Fatal(err)
	}
	digest, err := transaction.DigestPath(path)
	if err != nil {
		t.Fatal(err)
	}
	lm, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := lm.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	}()
	hit := false
	engine, err := transaction.New(fx.home, transaction.WithCommitCheck(projectlease.TransactionCheck(fx.home)), transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
		if e.Point == transaction.PointAfterBackup && e.LivePath == path {
			hit = true
			panic("lease-gap-stop")
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	id := newTransactionID()
	bindingPath := filepath.Join(fx.home, "state", "project-leases", "v1", "bindings", "restart-q.json")
	bindingStage := filepath.Join(filepath.Dir(stage), "binding.json")
	binding, err := projectleaserecord.Canonical(fresh.Binding)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(bindingStage, binding, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Prepare(held, transaction.Plan{TransactionID: id, ProjectIdentity: "project-leases", Targets: []transaction.Target{{Class: "project-lease", Identifier: bindingPath, LivePath: bindingPath, StagedSource: bindingStage, PreimageDigest: transaction.DigestAbsent, PrivateFile: true}, {Class: "project-lease", Identifier: path, LivePath: path, StagedSource: stage, PreimageDigest: digest, PrivateFile: true}}}); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if got := recover(); got != "lease-gap-stop" {
				t.Fatalf("wrong gap stop: %v", got)
			}
		}()
		_ = engine.Commit(held, id)
	}()
	if !hit {
		t.Fatal("actual lease backup boundary not reached")
	}
	if err = envauthority.Validate(fx.home); err != nil {
		t.Fatalf("genuine lease replacement gap refused: %v", err)
	}
	journalPath := filepath.Join(fx.home, "state", "transactions", "v1", id+".json")
	saved, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatal(err)
	}
	var journal transaction.Journal
	if err = json.Unmarshal(saved, &journal); err != nil {
		t.Fatal(err)
	}
	journal.Phase = transaction.PhasePrepared
	poison, err := json.Marshal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(journalPath, append(poison, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err = envauthority.Validate(fx.home); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
		t.Fatalf("phase plant escaped exact gap admission: %v", err)
	}
	if err = os.WriteFile(journalPath, saved, 0600); err != nil {
		t.Fatal(err)
	}
	freshEngine, err := transaction.New(fx.home, transaction.WithCommitCheck(projectlease.TransactionCheck(fx.home)))
	if err != nil {
		t.Fatal(err)
	}
	if recovered, err := projectlease.RecoverUnderLock(fx.home, held, freshEngine); err != nil || !recovered {
		t.Fatalf("actual Q recovery/J2: recovered=%t %v", recovered, err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	next, err := projectleaserecord.DecodeLedger(after)
	if err != nil {
		t.Fatal(err)
	}
	gotP, _ := json.Marshal(next.Leases["retained-p"])
	if string(gotP) != string(oldP) || next.Leases["restart-q"].State != projectleaserecord.Unknown || next.Leases["disjoint-q"].State != projectleaserecord.Unknown {
		t.Fatal("Q recovery/J2 changed frozen P or failed its real unknown transition")
	}
}
