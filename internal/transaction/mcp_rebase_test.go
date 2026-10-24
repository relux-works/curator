package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// A private MCP replacement retains both the landed commit-admission check and
// frozen-input check, including fresh-engine recovery. Each independent refusal
// preserves old live bytes; the valid sibling publishes a private replacement.
func TestMCPPrivateCommitAndInputGates(t *testing.T) {
	for _, restart := range []bool{false, true} {
		for _, plant := range []string{"valid", "commit-refusal", "input-change"} {
			t.Run(map[bool]string{false: "commit", true: "recovery"}[restart]+"/"+plant, func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				input := filepath.Join(root, "input")
				if err := os.WriteFile(input, []byte("allowed"), 0o600); err != nil {
					t.Fatal(err)
				}
				digest, err := DigestPath(input)
				if err != nil {
					t.Fatal(err)
				}
				target := fileTarget(t, "native", "one", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
				target.PrivateFile = true
				if err := privatedir.ProtectFile(target.StagedSource); err != nil {
					t.Fatal(err)
				}
				refused := false
				injected := errors.New("MCP commit admission refused")
				check := func(HomeLock, Journal) error {
					if refused {
						return injected
					}
					return nil
				}
				engine := mustEngine(t, home, WithCommitCheck(check))
				journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "mcp-combined", ProjectIdentity: "profile", Targets: []Target{target}, ReadInputs: map[string]string{input: digest}})
				if err != nil {
					t.Fatal(err)
				}
				refused = plant == "commit-refusal"
				if plant == "input-change" {
					if err := os.WriteFile(input, []byte("changed"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if restart {
					engine = mustEngine(t, home, WithCommitCheck(check))
					err = engine.Recover(testLock{})
				} else {
					err = engine.Commit(testLock{}, journal.TransactionID)
				}
				want := "old"
				switch plant {
				case "valid":
					want = "new"
					if err != nil {
						t.Fatal(err)
					}
					if err := privatedir.ValidateFile(target.LivePath); err != nil {
						t.Fatal(err)
					}
				case "commit-refusal":
					if !errors.Is(err, injected) {
						t.Fatalf("commit gate lost: %v", err)
					}
				case "input-change":
					if err == nil || !strings.Contains(err.Error(), "transaction_input_changed") {
						t.Fatalf("input gate lost: %v", err)
					}
				}
				if got := mustRead(t, target.LivePath); got != want {
					t.Fatalf("combined-gate live bytes = %q, want %q", got, want)
				}
			})
		}
	}
}
