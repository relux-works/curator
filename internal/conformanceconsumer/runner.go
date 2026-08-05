package conformanceconsumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// RunRequest is the versioned JSON request delivered to a black-box adapter.
type RunRequest struct {
	Boundary    string `json:"boundary"`
	CorpusRoot  string `json:"corpus_root"`
	FixtureRoot string `json:"fixture_root"`
	Case        string `json:"case"`
}

// RunResult contains raw process observations without interpreting manager
// behavior or making a qualification decision.
type RunResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   []byte `json:"stdout"`
	Stderr   []byte `json:"stderr"`
}

// Runner is the implementation-neutral adapter seam used by corpus consumers.
type Runner interface {
	Run(context.Context, RunRequest) (RunResult, error)
}

// ProcessRunner invokes an injected executable. The request is sent as JSON on
// stdin; no manager implementation package is linked into this package.
type ProcessRunner struct {
	Executable string
	Args       []string
	Env        []string
}

// Run executes the configured black-box process. Non-zero process exits are
// returned as observations; launch, transport, and context failures are errors.
func (r ProcessRunner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	if r.Executable == "" {
		return RunResult{}, fmt.Errorf("runner executable is required")
	}
	if request.Boundary != CorpusBoundaryV1 {
		return RunResult{}, fmt.Errorf("run request boundary %q is unsupported", request.Boundary)
	}
	input, err := json.Marshal(request)
	if err != nil {
		return RunResult{}, fmt.Errorf("encode run request: %w", err)
	}
	command := exec.CommandContext(ctx, r.Executable, r.Args...)
	command.Env = append([]string(nil), r.Env...)
	command.Stdin = bytes.NewReader(append(input, '\n'))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	result := RunResult{ExitCode: 0, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if err == nil {
		return result, nil
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return result, fmt.Errorf("run black-box adapter: %w", err)
	}
	result.ExitCode = exitError.ExitCode()
	return result, nil
}
