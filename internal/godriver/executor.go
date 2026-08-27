package godriver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

var (
	errProcessTimeout = errors.New("process deadline exceeded")
	errOutputLimit    = errors.New("process output limit exceeded")
)

// Process is the complete direct-execution contract for one Go probe. Stdin
// is intentionally absent: implementations must keep it closed.
type Process struct {
	Executable  string
	Arguments   []string
	Directory   string
	Environment []string
	Timeout     time.Duration
	OutputLimit int64
}

// Output contains bounded child output.
type Output struct {
	Stdout []byte
	Stderr []byte
}

// Executor is the only process seam used by the package-independent driver.
type Executor interface {
	Run(context.Context, Process) (Output, error)
}

// OSExecutor executes an absolute program directly, without a shell. A nil
// cmd.Stdin causes os/exec to connect the child to the null device.
type OSExecutor struct{}

// Run starts the absolute program directly with a bounded deadline and a shared
// combined-output budget. Standard input stays closed.
func (OSExecutor) Run(ctx context.Context, process Process) (Output, error) {
	runCtx, cancel := context.WithTimeout(ctx, process.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, process.Executable, process.Arguments...) // #nosec G204 -- executable is independently trusted before this seam
	cmd.Dir = process.Directory
	cmd.Env = append([]string(nil), process.Environment...)
	cmd.Stdin = nil
	budget := &outputBudget{remaining: process.OutputLimit}
	stdout := &boundedBuffer{budget: budget}
	stderr := &boundedBuffer{budget: budget}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	result := Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if runCtx.Err() != nil {
		return result, errProcessTimeout
	}
	if errors.Is(stdout.err, errOutputLimit) || errors.Is(stderr.err, errOutputLimit) || errors.Is(budget.err, errOutputLimit) {
		return result, errOutputLimit
	}
	return result, err
}

type boundedBuffer struct {
	buffer bytes.Buffer
	budget *outputBudget
	err    error
}

type outputBudget struct {
	mu        sync.Mutex
	remaining int64
	err       error
}

func (buffer *boundedBuffer) Write(payload []byte) (int, error) {
	buffer.budget.mu.Lock()
	defer buffer.budget.mu.Unlock()
	if buffer.budget.remaining <= 0 {
		buffer.err = errOutputLimit
		buffer.budget.err = errOutputLimit
		return 0, buffer.err
	}
	written := len(payload)
	if int64(written) > buffer.budget.remaining {
		written = int(buffer.budget.remaining)
	}
	_, _ = buffer.buffer.Write(payload[:written])
	buffer.budget.remaining -= int64(written)
	if written != len(payload) {
		buffer.err = errOutputLimit
		buffer.budget.err = errOutputLimit
		return written, buffer.err
	}
	return written, nil
}

// Bytes returns a copy under the shared budget lock, so a reader never races
// the writer goroutines os/exec runs for stdout and stderr.
func (buffer *boundedBuffer) Bytes() []byte {
	buffer.budget.mu.Lock()
	defer buffer.budget.mu.Unlock()
	return append([]byte(nil), buffer.buffer.Bytes()...)
}

var _ io.Writer = (*boundedBuffer)(nil)
