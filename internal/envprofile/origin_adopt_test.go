package envprofile

import (
	"bytes"
	"encoding/json"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Proves the real adoption owner writes only missing explicit authority,
// preserves the legacy home, and retries byte-identically without discovery.
// Wrong attested bindings and wrong credential targets preserve all preimages.
func TestProvisioningOriginAdoptionOwner(t *testing.T) {
	for _, state := range []string{"healthy", "wrong-lock", "wrong-target", "missing-link"} {
		t.Run(state, func(t *testing.T) {
			fx, _, raw := recordedOriginFixture(t, "a")
			var record envorigin.Record
			recordPtr, err := envorigin.Parse(raw, envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(fx.lockHash, "sha256:"), ManagedTarget: ManagedHomeDir(fx.home, fx.profile, "codex_cli")}, "codex_cli", fx.home)
			if err != nil {
				t.Fatal(err)
			}
			record = *recordPtr
			target := record.Binding.ManagedTarget
			path := filepath.Join(target, envorigin.Name)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if state == "wrong-target" || state == "missing-link" {
				if err := os.Remove(filepath.Join(target, "auth.json")); err != nil {
					t.Fatal(err)
				}
				if state == "wrong-target" {
					if err := os.Symlink(filepath.Join(t.TempDir(), "auth.json"), filepath.Join(target, "auth.json")); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := nativeSnapshot(t, target)
			request := AdoptOriginRequest{Home: fx.home, Machine: envregistry.DefaultMachineConfig(), Attestation: envorigin.AdoptRequest{Adapter: "codex_cli", Binding: record.Binding, OperatorHome: record.OperatorHome, SelectorState: record.Selector.State, SelectorValue: record.Selector.Value, NativeRoot: record.NativeRoot}}
			if state == "wrong-lock" {
				request.Attestation.Binding.LockSHA256 = "sha256:" + strings.Repeat("f", 64)
			}
			err = AdoptOrigin(request)
			if state != "healthy" {
				diagnostic := envorigin.DiagInvalid
				if state != "wrong-lock" {
					diagnostic = envregistry.DiagCredentialConflict
				}
				if err == nil || !strings.Contains(err.Error(), diagnostic) {
					t.Fatalf("intended adoption refusal: %v", err)
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("adoption refusal changed home")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			adopted, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := envorigin.Parse(adopted, record.Binding, "codex_cli", fx.home)
			if err != nil || parsed.Capture.Method != "adoption" {
				t.Fatalf("adopted carrier: %v", err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if nativeSnapshot(t, target) != before {
				t.Fatal("adoption changed another entry")
			}
			if err := os.WriteFile(path, adopted, 0600); err != nil {
				t.Fatal(err)
			}
			if err := AdoptOrigin(request); err != nil {
				t.Fatal(err)
			}
			again, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(again, adopted) {
				t.Fatal("identical attestation rewrote capture")
			}
		})
	}
}

// Two legitimate legacy adapters can be adopted sequentially. The unselected
// home remains byte-identical and quarantined; ordinary resolve is still refused
// until that home's explicit attestation is supplied. No account lookup occurs.
func TestProvisioningOriginTwoLegacyAdoptions(t *testing.T) {
	fx, _, first := recordedOriginFixture(t, "a")
	rr := fx.request("claude_code")
	rr.Repair = true
	if _, err := Resolve(rr); err != nil {
		t.Fatal(err)
	}
	records := [][]byte{first, nativeRead(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, "claude_code"), envorigin.Name))}
	ids := []string{"codex_cli", "claude_code"}
	requests := make([]AdoptOriginRequest, 2)
	for i, id := range ids {
		var r envorigin.Record
		if err := json.Unmarshal(records[i], &r); err != nil {
			t.Fatal(err)
		}
		requests[i] = AdoptOriginRequest{Home: fx.home, Machine: envregistry.DefaultMachineConfig(), Attestation: envorigin.AdoptRequest{Adapter: id, Binding: r.Binding, OperatorHome: r.OperatorHome, SelectorState: r.Selector.State, SelectorValue: r.Selector.Value, NativeRoot: r.NativeRoot}}
		if err := os.Remove(filepath.Join(r.Binding.ManagedTarget, envorigin.Name)); err != nil {
			t.Fatal(err)
		}
	}
	second := ManagedHomeDir(fx.home, fx.profile, ids[1])
	before := nativeSnapshot(t, second)
	if err := AdoptOrigin(requests[0]); err != nil {
		t.Fatal(err)
	}
	if nativeSnapshot(t, second) != before {
		t.Fatal("disjoint legacy neighbor changed")
	}
	rr = fx.request(ids[1])
	rr.Repair = false
	if _, err := Resolve(rr); err == nil || !strings.Contains(err.Error(), envorigin.DiagRequired) {
		t.Fatalf("ordinary continuation borrowed adoption quarantine: %v", err)
	}
	// An idempotent explicit adoption still has exactly the same narrow scope.
	if err := AdoptOrigin(requests[0]); err != nil {
		t.Fatal(err)
	}
	if err := AdoptOrigin(requests[1]); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		rr = fx.request(id)
		if _, err := Resolve(rr); err != nil {
			t.Fatal(err)
		}
	}
	wrong := requests[0]
	wrong.Attestation.NativeRoot = t.TempDir()
	before = nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, ids[0]))
	if err := AdoptOrigin(wrong); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
		t.Fatalf("different origin accepted: %v", err)
	}
	if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, ids[0])) != before {
		t.Fatal("different-origin refusal changed capture")
	}
}

// Explicit adoption can leave a different installed profile quarantined. A
// malformed neighbor, divergent installed binding, or unrelated valid pending
// journal cannot borrow that exception. Ordinary mutation remains blocked.
func TestProvisioningOriginAdoptionQuarantine(t *testing.T) {
	for _, state := range []string{"two-profiles", "malformed", "shared-lock", "unrelated-journal"} {
		t.Run(state, func(t *testing.T) {
			fx, _, raw := recordedOriginFixture(t, "a")
			source := t.TempDir()
			writePackage(t, source, "other-legacy", "1.0.0", "neighbor\n")
			if _, _, _, err := Install(fx.home, InstallOptions{Operand: source}); err != nil {
				t.Fatal(err)
			}
			other := *fx
			other.profile = "other-legacy"
			rr := other.request("codex_cli")
			rr.Repair = true
			if _, err := Resolve(rr); err != nil {
				t.Fatal(err)
			}
			var record envorigin.Record
			if err := json.Unmarshal(raw, &record); err != nil {
				t.Fatal(err)
			}
			request := AdoptOriginRequest{Home: fx.home, Machine: envregistry.DefaultMachineConfig(), Attestation: envorigin.AdoptRequest{Adapter: "codex_cli", Binding: record.Binding, OperatorHome: record.OperatorHome, SelectorState: record.Selector.State, SelectorValue: record.Selector.Value, NativeRoot: record.NativeRoot}}
			otherTarget := ManagedHomeDir(fx.home, other.profile, "codex_cli")
			otherRaw := nativeRead(t, filepath.Join(otherTarget, envorigin.Name))
			var otherRecord envorigin.Record
			if err := json.Unmarshal(otherRaw, &otherRecord); err != nil {
				t.Fatal(err)
			}
			if state == "unrelated-journal" {
				op, err := beginOperation(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(t.TempDir(), "entry")
				stage := filepath.Join(t.TempDir(), "stage")
				nativeWrite(t, stage, []byte("desired"))
				_, err = op.engine.Prepare(op.lock, transaction.Plan{TransactionID: "unrelated-adopt", ProjectIdentity: "fixture", Targets: []transaction.Target{{Class: "fixture", Identifier: target, LivePath: target, StagedSource: stage, PreimageDigest: transaction.DigestAbsent}}})
				if closeErr := op.close(); closeErr != nil {
					t.Fatal(closeErr)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range []string{filepath.Join(record.Binding.ManagedTarget, envorigin.Name), filepath.Join(otherTarget, envorigin.Name)} {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			if state == "malformed" {
				nativeWrite(t, filepath.Join(otherTarget, envorigin.Name), []byte("{}"))
			}
			if state == "shared-lock" {
				marker, err := envmarker.Read(otherTarget)
				if err != nil {
					t.Fatal(err)
				}
				marker.Profile.LockSHA256 = strings.Repeat("f", 64)
				body, err := marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				nativeWrite(t, filepath.Join(otherTarget, envmarker.Name), body)
			}
			before := nativeSnapshot(t, fx.home)
			neighborBefore := nativeSnapshot(t, otherTarget)
			if state == "two-profiles" {
				if op, err := beginOperation(fx.home); err == nil {
					if e := op.close(); e != nil {
						t.Error(e)
					}
					t.Fatal("ordinary mutation borrowed adoption quarantine")
				}
				if err := AdoptOrigin(request); err != nil {
					t.Fatal(err)
				}
				if nativeSnapshot(t, otherTarget) != neighborBefore {
					t.Fatal("neighbor authority was changed")
				}
				second := request
				second.Attestation.Binding = otherRecord.Binding
				second.Attestation.OperatorHome = otherRecord.OperatorHome
				second.Attestation.NativeRoot = otherRecord.NativeRoot
				second.Attestation.SelectorState = otherRecord.Selector.State
				second.Attestation.SelectorValue = otherRecord.Selector.Value
				if err := AdoptOrigin(second); err != nil {
					t.Fatal(err)
				}
				for _, profile := range []string{fx.profile, other.profile} {
					rr := fx.request("codex_cli")
					rr.Profile = profile
					if _, err := Resolve(rr); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				if err := AdoptOrigin(request); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
					t.Fatalf("quarantine boundary refusal: %v", err)
				}
				if nativeSnapshot(t, fx.home) != before {
					t.Fatal("quarantine refusal changed state")
				}
			}
		})
	}
}
