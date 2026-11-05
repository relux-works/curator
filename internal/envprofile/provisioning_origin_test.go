package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// A failed home publication must restore changed payloads and backup state,
// together with marker/origin. The fault must hit the actual journal boundary;
// a refusal before publication is not this rollback witness.
func TestProvisioningOriginWholeHomeRollback(t *testing.T) {
	for _, record := range []string{envmarker.Name, envorigin.Name} {
		for _, point := range []transaction.Point{transaction.PointBeforeInstall, transaction.PointAfterInstall} {
			t.Run(record+"/"+string(point), func(t *testing.T) {
				fx, req := managedCallerFixture(t)
				home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
				if err := os.WriteFile(filepath.Join(home, "AGENTS.md"), []byte("fixture drift to preserve on rollback\n"), 0600); err != nil {
					t.Fatal(err)
				}
				before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
				reached := false
				planted := errors.New("planted whole-home publication failure")
				req.Repair = true
				req.PreserveLock = true
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == point && event.LivePath == filepath.Join(home, record) {
						reached = true
						return planted
					}
					return nil
				}})}
				if _, err := Resolve(req); !reached || !errors.Is(err, planted) {
					t.Fatalf("intended publication fault not observed: reached=%t err=%v", reached, err)
				}
				if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
					t.Fatal("whole-home rollback left changed payload or backup state")
				}
			})
		}
	}
	// Nearby valid control: without the fault, repair restores the real surface
	// and its recorded origin remains usable on a fresh read-only operation.
	t.Run("valid-repair", func(t *testing.T) {
		fx, req := managedCallerFixture(t)
		path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), "AGENTS.md")
		original, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, []byte("fixture drift\n"), 0600); err != nil {
			t.Fatal(err)
		}
		req.Repair = true
		req.PreserveLock = true
		if _, err = Resolve(req); err != nil {
			t.Fatal(err)
		}
		if actual, err := os.ReadFile(path); err != nil || string(actual) != string(original) {
			t.Fatal("valid repair did not restore the planned payload")
		}
		req.Repair = false
		req.PreserveLock = false
		if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("valid repaired origin not usable: %v", err)
		}
	})
}

// Durable first-provisioning intent must recover after the marker is installed
// but before origin installation. An ordinary legacy home without such intent
// remains covered by the explicit-adoption refusal tests below.
func TestProvisioningOriginInterruptedProvisionRecovers(t *testing.T) {
	fx, req := managedCallerFixture(t)
	home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
	// Only this synthetic fixture home is removed to exercise first provisioning.
	if err := os.RemoveAll(home); err != nil {
		t.Fatal(err)
	}
	req.Repair = true
	const crash = "planted interruption before origin installation"
	reached := false
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointBeforeInstall && event.LivePath == filepath.Join(home, envorigin.Name) {
			reached = true
			panic(crash)
		}
		return nil
	}})}
	func() {
		defer func() {
			if recovered := recover(); recovered != crash {
				t.Fatalf("intended interruption not observed: %v", recovered)
			}
		}()
		_, _ = Resolve(req)
	}()
	if !reached {
		t.Fatal("origin installation boundary not reached")
	}
	req.transactionOptions = nil
	// Read-only resolution must not complete durable intent or infer the
	// missing record from the installed marker.
	req.Repair = false
	before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
	if result, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envorigin.DiagRequired) || (result != nil && len(result.Document) != 0) {
		t.Fatalf("read-only pending resolution did not refuse: %v", err)
	}
	if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
		t.Fatal("read-only pending resolution changed the home")
	}
	req.Repair = true
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("valid pending home transaction did not recover: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, envorigin.Name)); err != nil {
		t.Fatalf("recovery did not publish the origin: %v", err)
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatalf("recovered home not current: %v", err)
	}
}

// Canonical transaction records are necessary but not sufficient recovery
// authority. Semantically wrong bindings/owners, a missing paired marker and
// altered staging refuse before recovery; an unrelated journal never adopts a
// legacy home. The same public Prepare path supplies a valid positive control.
func TestProvisioningOriginPendingAdmission(t *testing.T) {
	for _, mode := range []string{"valid", "wrong-owner", "wrong-lock", "wrong-target", "wrong-marker", "unpaired", "unrelated", "changed-staging", "present-invalid", "competing-journal"} {
		t.Run(mode, func(t *testing.T) {
			fx, req := managedCallerFixtureAt(t, true)
			home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			originPath := filepath.Join(home, envorigin.Name)
			markerPath := filepath.Join(home, envmarker.Name)
			originBytes, err := os.ReadFile(originPath)
			if err != nil {
				t.Fatal(err)
			}
			markerBytes, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}
			record, err := envorigin.Parse(originBytes)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "wrong-lock":
				record.Binding.LockSHA256 = "sha256:" + strings.Repeat("b", 64)
			case "wrong-target":
				record.Binding.ManagedTarget = filepath.Join(home, "other")
			case "wrong-marker":
				marker, e := envmarker.Parse(markerBytes)
				if e != nil {
					t.Fatal(e)
				}
				marker.Profile.Name = "other"
				markerBytes, err = marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}
			originBytes, err = record.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(originPath); err != nil {
				t.Fatal(err)
			}
			if mode != "unrelated" {
				if err := os.Remove(markerPath); err != nil {
					t.Fatal(err)
				}
			}
			files := map[string][]byte{originPath: originBytes, markerPath: markerBytes}
			if mode == "unpaired" {
				delete(files, markerPath)
			}
			if mode == "unrelated" {
				files = map[string][]byte{filepath.Join(home, "unrelated-pending-file"): []byte("unrelated intent")}
			}
			plan := transaction.Plan{TransactionID: "origin-admission-control", ProjectIdentity: "profiles"}
			if mode == "wrong-owner" {
				plan.ProjectIdentity = "other"
			}
			for path, payload := range files {
				stage := filepath.Join(t.TempDir(), "record")
				if err := os.WriteFile(stage, payload, 0600); err != nil {
					t.Fatal(err)
				}
				digest, err := transaction.DigestPath(path)
				if err != nil {
					t.Fatal(err)
				}
				plan.Targets = append(plan.Targets, transaction.Target{Class: "profile", Identifier: path, LivePath: path, StagedSource: stage, PreimageDigest: digest})
			}
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			journal, prepareErr := op.engine.Prepare(op.lock, plan)
			if prepareErr == nil && mode == "competing-journal" {
				stage := filepath.Join(t.TempDir(), "other")
				if err := os.WriteFile(stage, []byte("separate home intent"), 0600); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(home, "other-pending-entry")
				_, prepareErr = op.engine.Prepare(op.lock, transaction.Plan{TransactionID: "competing-home-intent", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: path, LivePath: path, StagedSource: stage, PreimageDigest: transaction.DigestAbsent}}})
			}
			closeErr := op.close()
			if err := errors.Join(prepareErr, closeErr); err != nil {
				t.Fatal(err)
			}
			if mode == "changed-staging" {
				for _, target := range journal.Targets {
					if target.LivePath == originPath {
						if err := os.WriteFile(target.StagedPath, []byte("{}"), 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			if mode == "present-invalid" {
				if err := os.WriteFile(originPath, []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			journalPath := filepath.Join(fx.home, "state", "transactions", "v1", plan.TransactionID+".json")
			journalBefore, err := os.ReadFile(journalPath)
			if err != nil {
				t.Fatal(err)
			}
			before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
			recoveryCalls := 0
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveryCalls++ }})}
			req.Repair = true
			result, err := Resolve(req)
			if mode == "valid" {
				if err != nil || result == nil || len(result.Document) == 0 || recoveryCalls != 1 {
					t.Fatalf("valid pending authority: recovery=%d err=%v", recoveryCalls, err)
				}
				adapter, _ := envregistry.ByID(req.EnvID)
				if _, present, err := req.readProvisioningOrigin(adapter); err != nil || !present {
					t.Fatalf("valid recovery did not install origin: %v", err)
				}
				return
			}
			diagnostic := envorigin.DiagInvalid
			if mode == "unrelated" {
				diagnostic = envorigin.DiagRequired
			}
			if err == nil || !strings.Contains(err.Error(), diagnostic) || recoveryCalls != 0 || (result != nil && len(result.Document) != 0) {
				t.Fatalf("pending authority refusal: recovery=%d err=%v", recoveryCalls, err)
			}
			if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
				t.Fatal("pending authority refusal changed fixture state")
			}
			journalAfter, err := os.ReadFile(journalPath)
			if err != nil || string(journalAfter) != string(journalBefore) {
				t.Fatal("pending authority refusal changed durable intent")
			}
		})
	}
}

// Each legacy/invalid authority refuses both read and repair without changing
// home/profile/native metadata, beside a valid original-origin control.
func TestProvisioningOriginRefusals(t *testing.T) {
	for _, mode := range []string{"missing", "malformed", "directory", "wrong-profile", "wrong-lock", "wrong-target"} {
		for _, repair := range []bool{false, true} {
			t.Run(mode+map[bool]string{false: "-read", true: "-repair"}[repair], func(t *testing.T) {
				fx, req := managedCallerFixtureAt(t, true)
				home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
				recordPath := filepath.Join(home, envorigin.Name)
				raw, err := os.ReadFile(recordPath)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "missing" {
					err = os.Remove(recordPath)
				} else if mode == "malformed" {
					err = os.WriteFile(recordPath, []byte("{"), 0600)
				} else if mode == "directory" {
					if err = os.Remove(recordPath); err == nil {
						err = os.Mkdir(recordPath, 0700)
					}
				} else {
					record, e := envorigin.Parse(raw)
					if e != nil {
						t.Fatal(e)
					}
					switch mode {
					case "wrong-profile":
						record.Binding.Profile = "other"
					case "wrong-lock":
						record.Binding.LockSHA256 = "sha256:" + strings.Repeat("b", 64)
					case "wrong-target":
						record.Binding.ManagedTarget = filepath.Join(home, "other")
					}
					raw, e = record.Marshal()
					if e != nil {
						t.Fatal(e)
					}
					err = os.WriteFile(recordPath, raw, 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
				before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
				req.Repair = repair
				result, err := Resolve(req)
				diag := envorigin.DiagInvalid
				if mode == "missing" {
					diag = envorigin.DiagRequired
				}
				if err == nil || !strings.Contains(err.Error(), diag) {
					t.Fatalf("refusal=%v", err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("refusal emitted a fragment")
				}
				if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
					t.Fatal("authority refusal changed fixture state")
				}
			})
		}
	}
}

// A returned failure at either side of the record rename rolls adoption back
// to the legacy preimage; retry without the fault can adopt that same home.
// These are in-process commit failures, not process-crash recovery evidence.
func TestProvisioningOriginAdoptionCommitRollback(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointBeforeInstall, transaction.PointAfterInstall} {
		t.Run(string(point), func(t *testing.T) {
			fx, req := managedCallerFixtureAt(t, true)
			path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envorigin.Name)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			record, err := envorigin.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Remove(path); err != nil {
				t.Fatal(err)
			}
			before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
			planted := errors.New("planted origin adoption commit failure")
			reached := false
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == point && event.LivePath == path {
					reached = true
					return planted
				}
				return nil
			}})}
			input := OriginAdoptRequest{Resolve: req, OperatorHome: record.OperatorHome, NativeRoot: record.NativeRoot, SelectorState: record.Selector.State, SelectorValue: record.Selector.Value}
			if changed, err := AdoptOrigin(input); changed || !errors.Is(err, planted) || !reached {
				t.Fatalf("commit failure changed=%v err=%v reached=%v", changed, err, reached)
			}
			if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
				t.Fatal("failed adoption left a changed home/native preimage")
			}
			input.Resolve.transactionOptions = nil
			if changed, err := AdoptOrigin(input); err != nil || !changed {
				t.Fatalf("retry changed=%v err=%v", changed, err)
			}
		})
	}
}

// Explicit trusted input admits an otherwise current legacy home once;
// repeated identical adoption is read-only and a changed origin cannot replace it.
func TestProvisioningOriginExplicitAdoption(t *testing.T) {
	fx, req := managedCallerFixtureAt(t, true)
	recordPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envorigin.Name)
	raw, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatal(err)
	}
	original, err := envorigin.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	input := OriginAdoptRequest{Resolve: req, OperatorHome: original.OperatorHome, NativeRoot: original.NativeRoot, SelectorState: original.Selector.State, SelectorValue: original.Selector.Value}
	changed, err := AdoptOrigin(input)
	if err != nil || !changed {
		t.Fatalf("adopt=(%v,%v)", changed, err)
	}
	raw, err = os.ReadFile(recordPath)
	if err != nil {
		t.Fatal(err)
	}
	var adopted envorigin.Record
	if err = json.Unmarshal(raw, &adopted); err != nil {
		t.Fatal(err)
	}
	if adopted.Capture.Method != "adoption" || adopted.Capture.Authority != "explicit-operator-input" {
		t.Fatal("adoption provenance missing")
	}
	before := managedCallerSnapshot(t, fx, fx.native["codex_cli"])
	changed, err = AdoptOrigin(input)
	if err != nil || changed {
		t.Fatalf("repeat=(%v,%v)", changed, err)
	}
	if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
		t.Fatal("idempotent adoption changed state")
	}
	input.NativeRoot = t.TempDir()
	if _, err = AdoptOrigin(input); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
		t.Fatalf("changed origin=%v", err)
	}
	if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, fx.native["codex_cli"])) {
		t.Fatal("refused replacement changed state")
	}
	if _, err = Resolve(req); err != nil {
		t.Fatalf("adopted origin fails valid resolve: %v", err)
	}
}
