package envprofile

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// The real MCP writer may temporarily back up its own lock target. That
// absence is a journal state, not changed external input (PR197 R1 witness).
func TestProfileMCPLockBackupValidation(t *testing.T) {
	fx, req := nativeWriterFixture(t)
	path := lockPath(fx.home, fx.profile)
	before := nativeRead(t, path)
	backedUp := false
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
		if event.Point == transaction.PointAfterBackup && event.LivePath == path {
			backedUp = true
			nativeAbsent(t, path)
		}
	}})}
	result, err := ProfileMCP(req)
	if err != nil {
		t.Fatal(err)
	}
	if !backedUp || !result.Published || !bytes.Equal(before, nativeRead(t, path)) {
		t.Fatal("lock backup boundary not reached or exact lock not published")
	}
}

// Once the operation owns the home lock, an external edit before publication
// still refuses, preserves the edit and publishes no native/ownership files.
// The adjacent unchanged case must publish through the same production writer.
func TestProfileMCPExternalLockValidation(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			path := lockPath(fx.home, fx.profile)
			before := nativeRead(t, path)
			external := append(append([]byte(nil), before...), '\n')
			reached := false
			req.beforePublish = func() {
				reached = true
				if changed {
					nativeWrite(t, path, external)
				}
			}
			result, err := ProfileMCP(req)
			if !reached {
				t.Fatal("external mutation boundary not reached")
			}
			if !changed {
				if err != nil || !result.Published || !bytes.Equal(before, nativeRead(t, path)) {
					t.Fatalf("unchanged lock refused: %+v %v", result, err)
				}
				return
			}
			requireMCPError(t, err, "transaction_input_changed")
			if result.Published || !bytes.Equal(external, nativeRead(t, path)) {
				t.Fatal("refusal published or overwrote external lock bytes")
			}
			for _, id := range req.Environments {
				nativeAbsent(t, nativeWriterPath(fx, id))
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
			}
		})
	}
}

// An external lock frozen as ReadInputs remains guarded during swaps, after
// the one-time input callback has expired. Drift rolls back both outputs and
// preserves the foreign edit; the unchanged control commits both outputs.
func TestOperationExternalLockDuringSwaps(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			home := t.TempDir()
			input := lockPath(home, "external")
			nativeWrite(t, input, []byte("original lock"))
			digest, err := transaction.DigestPath(input)
			if err != nil {
				t.Fatal(err)
			}
			targets := make([]transaction.Target, 2)
			for i := range targets {
				live := filepath.Join(home, "outputs", fmt.Sprint(i))
				stage := filepath.Join(t.TempDir(), "staged")
				nativeWrite(t, live, []byte("old"))
				nativeWrite(t, stage, []byte("new"))
				preimage, e := transaction.DigestPath(live)
				if e != nil {
					t.Fatal(e)
				}
				targets[i] = transaction.Target{Class: "fixture", Identifier: fmt.Sprint(i), LivePath: live, StagedSource: stage, PreimageDigest: preimage}
			}
			checked, reached := 0, false
			op, err := beginOperationValidated(home, func() error {
				checked++
				current, e := transaction.DigestPath(input)
				if e != nil {
					return e
				}
				if current != digest {
					return fmt.Errorf("transaction_input_changed: external fixture lock")
				}
				return nil
			}, transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
				if event.Point == transaction.PointAfterInstall && event.LivePath == targets[0].LivePath {
					reached = true
					if changed {
						nativeWrite(t, input, []byte("external edit"))
					}
				}
			}}))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = op.close() }()
			journal, err := op.engine.Prepare(op.lock, transaction.Plan{TransactionID: "external-lock-control", ProjectIdentity: "fixture", Targets: targets, ReadInputs: map[string]string{input: digest}})
			if err != nil {
				t.Fatal(err)
			}
			err = op.engine.Commit(op.lock, journal.TransactionID)
			if !reached || checked != 1 {
				t.Fatalf("mutation reached=%v, input checks=%d; want one pre-recovery check", reached, checked)
			}
			wantOutput, wantInput := "new", "original lock"
			if changed {
				requireMCPError(t, err, "transaction_input_changed")
				wantOutput, wantInput = "old", "external edit"
			} else if err != nil {
				t.Fatal(err)
			}
			for _, target := range targets {
				if string(nativeRead(t, target.LivePath)) != wantOutput {
					t.Fatal("incoherent output generation")
				}
			}
			if string(nativeRead(t, input)) != wantInput {
				t.Fatal("external input overwritten")
			}
		})
	}
}
