//go:build !windows

package closureexec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

func TestPortableRunnerRejectsMutationOfImmutableAdmittedReplay(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	executionRoot := t.TempDir()
	outputRoot := filepath.Join(executionRoot, "output")
	if err := os.MkdirAll(filepath.Join(executionRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(executionRoot, "work"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	toolBytes, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	toolPath := filepath.Join(executionRoot, "bin", "tool")
	if err = os.WriteFile(toolPath, toolBytes, 0o500); err != nil {
		t.Fatal(err)
	}
	toolID := digestBytes(toolBytes)
	runner, err := NewManagerProcessRunner(executionRoot, outputRoot)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := NewAssuredExecutor(AssuranceConfig{}, runner, nil, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permit := permitFixture(receiptID)
	permit.ResourceLimits.WallTimeMillis = 10_000
	permit.ResourceLimitID, err = permit.ResourceLimits.ID()
	if err != nil {
		t.Fatal(err)
	}
	permit.ToolchainFingerprint = toolID
	permit.ExecutableSHA256 = toolID
	permit.AllowedProcesses = []string{}
	permit.Argv = []string{"-test.run=^TestPortableMutationHelper$"}
	permit.Environment = map[string]string{"CURATOR_OUTPUT_ROOT": outputRoot, "HOME": "home"}
	permitID, err := executor.Commit(permit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) {
		return ToolchainIdentity{Fingerprint: toolID, ExecutableSHA256: toolID}, nil
	}, map[closuregraph.ID]AdmittedInput{receiptID: input})
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "closure_input_mutated" {
		t.Fatalf("error = %v", err)
	}
}

func TestPortableMutationHelper(t *testing.T) {
	outputRoot := os.Getenv("CURATOR_OUTPUT_ROOT")
	if outputRoot == "" {
		return
	}
	if err := os.MkdirAll(filepath.Join(outputRoot, "evidence"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputRoot, "evidence", "manifest.json"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join("..", "capture", "source")
	if err := os.Chmod(input, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}
}
