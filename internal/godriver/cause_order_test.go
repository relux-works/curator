package godriver

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The response fault is selected by a file beside the test binary, because a
// Build worker receives only the fixed bootstrap environment. Only the test
// worker's TestMain consults it; the installed worker has no output seam.
func init() { workerOutputTransform = responseFaultOutput }

func responseFaultPath() string {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(executable), "curator-test-worker-response-fault")
}

func responseFaultOutput(output *os.File) io.Writer {
	mode, _ := os.ReadFile(responseFaultPath()) // #nosec G304 -- test-owned path beside the test binary
	if string(mode) != "eof" && string(mode) != "malformed" {
		return output
	}
	return &responseFaultWriter{output: output, malformed: string(mode) == "malformed"}
}

// responseFaultWriter forwards the real ready frame (header and payload are the
// first two writes) and faults the next frame, the list result: it closes the
// response stream, after four zero header bytes when malformed.
type responseFaultWriter struct {
	output    *os.File
	malformed bool
	writes    int
}

func (writer *responseFaultWriter) Write(payload []byte) (int, error) {
	writer.writes++
	if writer.writes < 3 {
		return writer.output.Write(payload)
	}
	if writer.writes == 3 && writer.malformed {
		_, _ = writer.output.Write([]byte{0, 0, 0, 0})
	}
	_ = writer.output.Close()
	return 0, io.ErrClosedPipe
}

// causeOrderBudget is the whole worker-session budget of every case. The
// pause at the post-list seam exceeds it by a margin that also absorbs runtime
// timer latency, so "budget elapsed" is established before a later cancel.
const (
	causeOrderBudget = 3 * time.Second
	causeOrderPause  = causeOrderBudget + 500*time.Millisecond
)

// Public Build with a real worker keeps the first terminal cause through
// production receive, protocolFailure, teardown and the production post-list
// bracket in build.go: a budget that elapses after teardown stopped the run
// timer stays process_timeout although the caller is canceled later, and a
// caller canceled before the budget stays process_canceled. Live EOF and a
// live malformed frame stay protocol-invalid with no artificial cause, and the
// untouched-stream neighbour builds. Only the test worker's response stream is
// faulted; no failed case returns an artifact. This exercises one platform's
// scheduling per run, not every possible interleaving of the two events.
func TestBuildCauseOrderAfterTeardown(t *testing.T) {
	cases := []struct {
		name         string
		fault        string
		cancelBefore bool // cancel the caller at the seam, before the budget
		pause        bool // pause at the seam until the budget has elapsed
		cancelAfter  bool // cancel the caller after that pause
		want         string
	}{
		{name: "valid"},
		{name: "eof-live", fault: "eof", want: CodeWorkerProtocolInvalid},
		{name: "malformed-live", fault: "malformed", want: CodeWorkerProtocolInvalid},
		{name: "eof-budget-during-cleanup", fault: "eof", pause: true, want: "process_timeout"},
		{name: "eof-cancel-before-budget", fault: "eof", cancelBefore: true, pause: true, want: "process_canceled"},
		{name: "eof-budget-before-cancel", fault: "eof", pause: true, cancelAfter: true, want: "process_timeout"},
		{name: "malformed-budget-before-cancel", fault: "malformed", pause: true, cancelAfter: true, want: "process_timeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fault != "" {
				if err := os.WriteFile(responseFaultPath(), []byte(tc.fault), 0o600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Remove(responseFaultPath()) })
			}
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			caller, cancelCaller := context.WithCancel(t.Context())
			defer cancelCaller()
			seamReached := false
			var premise error
			started := time.Now()
			fixture.session.observePhase = func(phase string, _ time.Duration) {
				if phase != "list" {
					return
				}
				seamReached = true
				// The worker budget starts after started, so reaching the seam
				// before started+budget proves the budget has not elapsed yet.
				if time.Since(started) >= causeOrderBudget {
					premise = errors.New("fixture reached the post-list seam after the budget could have elapsed")
				}
				if tc.cancelBefore {
					cancelCaller()
				}
				if tc.pause {
					time.Sleep(causeOrderPause)
				}
				if tc.cancelAfter {
					cancelCaller()
				}
			}
			result, err := Build(caller, fixture.request(ResourceLimits{Timeout: causeOrderBudget}))
			if !seamReached {
				t.Fatalf("fixture never reached the post-list bracket: %v", err)
			}
			if premise != nil {
				t.Fatalf("%v (%v); the ordering claim was not exercised", premise, err)
			}
			if tc.want == "" {
				if err != nil || result.Artifact.StagedPath == "" {
					t.Fatalf("untouched response stream did not build: %v", err)
				}
				return
			}
			if code := DiagnosticCode(err); code != tc.want {
				t.Errorf("Build classified %s as %q; want %q: %v", tc.name, code, tc.want, err)
			}
			switch tc.want {
			case "process_timeout":
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Build lost the elapsed budget cause: %v", err)
				}
				// A later caller cancellation may still appear as a secondary
				// cause (the bracket's toolchain recheck observes the caller);
				// only the primary code above orders the two. Without any
				// caller cancellation, cleanup must not invent one.
				if !tc.cancelAfter && errors.Is(err, context.Canceled) {
					t.Errorf("Build let cleanup cancellation join the timeout: %v", err)
				}
			case "process_canceled":
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Build lost the caller cancellation cause: %v", err)
				}
				if errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Build let the later budget replace caller cancellation: %v", err)
				}
			default:
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Build added a context cause to a live protocol failure: %v", err)
				}
			}
			if result.Artifact.StagedPath != "" {
				t.Errorf("failed Build returned artifact %q", result.Artifact.StagedPath)
			}
		})
	}
}
