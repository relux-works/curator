//go:build unix

package closureexec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
)

func realPortableFixture(t *testing.T, script string) (*Executor, DerivationPermit, map[closuregraph.ID]AdmittedInput, *ManagerProcessRunner) {
	t.Helper()
	_, input, receiptID := admittedFixture(t)
	root := t.TempDir()
	for _, directory := range []string{"bin", "work"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(root, "bin", "tool")
	payload := []byte("#!/bin/sh\nset -eu\n" + script + "\n")
	if err := os.WriteFile(executable, payload, 0o700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "output")
	runner, err := NewManagerProcessRunner(root, output)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := NewAssuredExecutor(DefaultAssuranceConfig(), runner, nil, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permit := portablePermitFixture(receiptID)
	permit.Argv = []string{}
	permit.ExecutableSHA256 = digestBytes(payload)
	permit.Environment = map[string]string{
		"CURATOR_EXECUTION_ROOT": root,
		"CURATOR_OUTPUT_ROOT":    output,
	}
	permit.ResourceLimits.WallTimeMillis = 30_000
	permit.ResourceLimits.OutputBytes = 1_024
	permit.ResourceLimitID, _ = permit.ResourceLimits.ID()
	return executor, permit, map[closuregraph.ID]AdmittedInput{receiptID: input}, runner
}

func executeRealPortable(t *testing.T, executor *Executor, permit DerivationPermit, inputs map[closuregraph.ID]AdmittedInput) (DerivationReceipt, error) {
	return executeRealPortableContext(context.Background(), t, executor, permit, inputs)
}

func executeRealPortableContext(ctx context.Context, t *testing.T, executor *Executor, permit DerivationPermit, inputs map[closuregraph.ID]AdmittedInput) (DerivationReceipt, error) {
	t.Helper()
	permitID, err := executor.Commit(permit)
	if err != nil {
		return DerivationReceipt{}, err
	}
	identity := ToolchainIdentity{Fingerprint: permit.ToolchainFingerprint, ExecutableSHA256: permit.ExecutableSHA256}
	return executor.Execute(ctx, permitID, func(context.Context) (ToolchainIdentity, error) { return identity, nil }, inputs)
}

func TestManagerProcessRunnerExecutesDeclaredOutputWithHonestPortableEvidence(t *testing.T) {
	executor, permit, inputs, _ := realPortableFixture(t, `mkdir -p "$CURATOR_OUTPUT_ROOT/evidence"; printf src >"$CURATOR_OUTPUT_ROOT/evidence/manifest.json"`)
	receipt, err := executeRealPortable(t, executor, permit, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.AssuranceMode != AssurancePortable || receipt.Audit.Network != "not-observed" || len(receipt.Audit.Processes) != 0 || len(receipt.Audit.Reads) != 0 || len(receipt.Audit.Writes) != 0 {
		t.Fatalf("portable evidence inflated host observations: %+v", receipt)
	}
}

func TestManagerProcessRunnerRetainsOnlyDeclaredSuccessfulWorkCopy(t *testing.T) {
	executor, permit, _, runner := realPortableFixture(t, `printf changed >"$CURATOR_EXECUTION_ROOT/work/retained/main.txt"; mkdir -p "$CURATOR_OUTPUT_ROOT/evidence"; printf src >"$CURATOR_OUTPUT_ROOT/evidence/manifest.json"`)
	input, receiptID, _ := admittedTreeFixture(t)
	permit.AdmittedInputReceiptIDs = []closuregraph.ID{receiptID}
	permit.InputMounts = []InputMount{{ReceiptID: receiptID, Path: "capture/tree"}}
	permit.ReadRoots = []string{"capture/tree"}
	permit.WorkCopies = []WorkCopy{{ReceiptID: receiptID, Path: "work/retained", Retain: true}}
	permit.WriteRoots = []string{"evidence/manifest.json", "work/retained"}
	if _, err := executeRealPortable(t, executor, permit, map[closuregraph.ID]AdmittedInput{receiptID: input}); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(runner.ExecutionRoot, "work", "retained", "main.txt"))
	if err != nil || string(payload) != "changed" {
		t.Fatalf("retained typed work copy is unavailable: payload=%q err=%v", payload, err)
	}
	if _, err = os.Lstat(filepath.Join(runner.ExecutionRoot, "capture", "tree")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("immutable replay was retained with derivative: %v", err)
	}
}

func TestManagerProcessRunnerRejectsMutatedReplayAndUndeclaredOutput(t *testing.T) {
	tests := []struct {
		name, script, code string
	}{
		{name: "mutated replay", code: "closure_input_mutated", script: `chmod 600 "$CURATOR_EXECUTION_ROOT/capture/source"; printf bad >"$CURATOR_EXECUTION_ROOT/capture/source"; mkdir -p "$CURATOR_OUTPUT_ROOT/evidence"; printf src >"$CURATOR_OUTPUT_ROOT/evidence/manifest.json"`},
		{name: "undeclared output", code: "closure_write_undeclared", script: `mkdir -p "$CURATOR_OUTPUT_ROOT/evidence"; printf src >"$CURATOR_OUTPUT_ROOT/evidence/manifest.json"; printf extra >"$CURATOR_OUTPUT_ROOT/extra"`},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			executor, permit, inputs, _ := realPortableFixture(t, testCase.script)
			_, err := executeRealPortable(t, executor, permit, inputs)
			var diagnostic *DiagnosticError
			if !errors.As(err, &diagnostic) || diagnostic.Code != testCase.code {
				t.Fatalf("err=%v want code=%s", err, testCase.code)
			}
		})
	}
}

func TestManagerProcessRunnerRejectsNonzeroExitAndCombinedOutputOverflow(t *testing.T) {
	tests := []struct {
		name, script, code string
	}{
		{name: "nonzero", script: `mkdir -p "$CURATOR_OUTPUT_ROOT/evidence"; printf src >"$CURATOR_OUTPUT_ROOT/evidence/manifest.json"; exit 7`, code: "closure_derivation_drift"},
		{name: "output overflow", script: `while :; do printf 1234567890; done`, code: "portable_output_limit"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			executor, permit, inputs, _ := realPortableFixture(t, testCase.script)
			permit.ResourceLimits.OutputBytes = 32
			permit.ResourceLimitID, _ = permit.ResourceLimits.ID()
			_, err := executeRealPortable(t, executor, permit, inputs)
			var diagnostic *DiagnosticError
			if !errors.As(err, &diagnostic) || diagnostic.Code != testCase.code {
				t.Fatalf("err=%v want code=%s", err, testCase.code)
			}
		})
	}
}

func TestManagerProcessRunnerTimeoutCleansDescendantProcessGroup(t *testing.T) {
	executor, permit, inputs, runner := realPortableFixture(t, `sleep 30 & child=$!; printf '%s' "$child" >"$CURATOR_OUTPUT_ROOT/child.pid"; wait "$child"`)
	permit.ResourceLimits.WallTimeMillis = 30_000
	permit.ResourceLimitID, _ = permit.ResourceLimits.ID()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	result := make(chan error, 1)
	go func() {
		_, err := executeRealPortableContext(ctx, t, executor, permit, inputs)
		result <- err
	}()
	pidPath := filepath.Join(runner.OutputRoot, "child.pid")
	deadline := time.Now().Add(10 * time.Second)
	pid := 0
	for {
		payload, readErr := os.ReadFile(pidPath)
		if readErr == nil {
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(string(payload)))
			if parseErr == nil && parsed > 0 {
				pid = parsed
				break
			}
		} else if !errors.Is(readErr, os.ErrNotExist) {
			t.Fatal(readErr)
		}
		if time.Now().After(deadline) {
			t.Fatal("portable child did not publish its startup handshake")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	err := <-result
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "portable_process_timeout" {
		t.Fatalf("timeout err=%v", err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if killErr := syscall.Kill(pid, 0); killErr == syscall.ESRCH {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("descendant pid %d survived portable timeout", pid)
}

func TestManagerProcessRunnerRejectsOutputAndReplayEscapeBindings(t *testing.T) {
	executor, permit, inputs, runner := realPortableFixture(t, `exit 0`)
	permit.Environment["CURATOR_OUTPUT_ROOT"] = filepath.Join(runner.ExecutionRoot, "other")
	_, err := executeRealPortable(t, executor, permit, inputs)
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "closure_derivation_unauthorized" {
		t.Fatalf("output escape err=%v", err)
	}
}
