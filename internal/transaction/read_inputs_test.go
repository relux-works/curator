package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Removing a live entry has no staged sidecar. That absent path must not
// overlap an independent frozen input; real sidecar aliases still refuse in
// TestFrozenReadInputValidation. Commit must reach the intended removal.
func TestFrozenReadInputRemoval(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	input := filepath.Join(root, "policy")
	if err := os.WriteFile(input, []byte("allowed"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest, err := DigestPath(input)
	if err != nil {
		t.Fatal(err)
	}
	target := fileTarget(t, "native", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "stage", "one"), "old", "unused")
	target.StagedSource = ""
	home := filepath.Join(root, "home")
	mustMkdirAll(t, home)
	engine := mustEngine(t, home)
	j, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-input-removal", ProjectIdentity: "profile", Targets: []Target{target}, ReadInputs: map[string]string{input: digest}})
	if err != nil {
		t.Fatal(err)
	}
	if err = engine.Commit(testLock{}, j.TransactionID); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(target.LivePath); !os.IsNotExist(err) {
		t.Fatalf("removal did not reach live target: %v", err)
	}
	if got := mustRead(t, input); got != "allowed" {
		t.Fatalf("independent input changed: %q", got)
	}
}

// Frozen bytes survive loss of the process-local guard. In-place policy edits
// refuse both same-process commit and fresh-engine recovery, preserving targets;
// identical input permits recovery of the prepared new generation.
func TestFrozenReadInputsCommitAndRecovery(t *testing.T) {
	for _, restart := range []bool{false, true} {
		for _, changed := range []bool{false, true} {
			t.Run(map[bool]string{false: "commit", true: "recovery"}[restart]+map[bool]string{false: "-valid", true: "-changed"}[changed], func(t *testing.T) {
				root := t.TempDir()
				mustMkdirAll(t, filepath.Join(root, "live"))
				mustMkdirAll(t, filepath.Join(root, "stage"))
				input := filepath.Join(root, "policy")
				if err := os.WriteFile(input, []byte("allowed"), 0o600); err != nil {
					t.Fatal(err)
				}
				digest, err := DigestPath(input)
				if err != nil {
					t.Fatal(err)
				}
				target := fileTarget(t, "native", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "stage", "one"), "old", "new")
				home := filepath.Join(root, "home")
				engine := mustEngine(t, home)
				inputs := map[string]string{input: digest}
				journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-input", ProjectIdentity: "profile", Targets: []Target{target}, ReadInputs: inputs})
				if err != nil {
					t.Fatal(err)
				}
				inputs[input] = DigestAbsent // The plan map is not a mutable authority handle.
				if changed {
					if err = os.WriteFile(input, []byte("denied!"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if restart {
					engine = mustEngine(t, home)
					err = engine.Recover(testLock{})
				} else {
					err = engine.Commit(testLock{}, journal.TransactionID)
				}
				want := "new"
				if changed {
					want = "old"
					if err == nil || !strings.Contains(err.Error(), "transaction_input_changed") {
						t.Fatalf("input change did not refuse: %v", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if got := mustRead(t, target.LivePath); got != want {
					t.Fatalf("target = %q, want %q", got, want)
				}
			})
		}
	}
}

// A journal cannot freeze its own changing target or sidecar namespace, nor
// accept malformed paths/digests; independent inputs are exercised above.
func TestFrozenReadInputValidation(t *testing.T) {
	for _, kind := range []string{"target", "child", "ancestor", "sidecar", "sidecar-child", "relative", "digest"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			mustMkdirAll(t, filepath.Join(root, "live"))
			mustMkdirAll(t, filepath.Join(root, "stage"))
			target := fileTarget(t, "native", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "stage", "one"), "old", "new")
			path := target.LivePath
			digest := target.PreimageDigest
			switch kind {
			case "child":
				path = filepath.Join(path, "child")
			case "ancestor":
				path = filepath.Dir(path)
			case "sidecar", "sidecar-child":
				path = sidecarPath(target.LivePath, "txn-invalid-input", 0, "backup")
				if kind == "sidecar-child" {
					path = filepath.Join(path, "child")
				}
			case "relative":
				path = "relative"
			case "digest":
				path = filepath.Join(root, "policy")
				digest = "bad"
			}
			engine := mustEngine(t, filepath.Join(root, "home"))
			wantReason := "invalid frozen read input"
			if kind == "target" || kind == "child" || kind == "ancestor" {
				wantReason = "read input is a changing transaction target"
			}
			if kind == "sidecar" || kind == "sidecar-child" {
				wantReason = "read input aliases a transaction sidecar"
			}
			if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-invalid-input", ProjectIdentity: "profile", Targets: []Target{target}, ReadInputs: map[string]string{path: digest}}); !errors.Is(err, errInvalidJournal) || !strings.Contains(err.Error(), wantReason) {
				t.Fatalf("frozen input rejection = %v, want invalid journal: %s", err, wantReason)
			}
			if got := mustRead(t, target.LivePath); got != "old" {
				t.Fatal("invalid input changed live target")
			}
		})
	}
}
