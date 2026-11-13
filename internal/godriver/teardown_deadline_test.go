package godriver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// The production receive/teardown path and Build's final error join must not
// turn cleanup cancellation into caller intent. A delayed cleanup crosses a
// real budget deadline; genuine caller cancellation retains its own cause.
// This is a channel/cleanup fixture, not a native worker-domain death test.
func TestWorkerCleanupContextClassification(t *testing.T) {
	for _, mode := range []string{"protocol-live", "deadline-during-cleanup", "caller-cancel", "caller-cancel-delayed-cleanup", "deadline-before-caller-cancel"} {
		t.Run(mode, func(t *testing.T) {
			caller, cancelCaller := context.WithCancel(t.Context())
			defer cancelCaller()
			budget, cancelBudget := context.WithTimeout(caller, 500*time.Millisecond)
			defer cancelBudget()
			deadline, _ := budget.Deadline()
			cleanupReached := false
			client := &workerClient{
				callerContext: caller,
				runContext:    budget,
				stdout:        io.NopCloser(bytes.NewReader(nil)),
				stderr:        &boundedBuffer{budget: &outputBudget{remaining: 1024}},
				cancel: func() {
					cleanupReached = true
					if budget.Err() != nil {
						t.Fatal("fixture reached cleanup after its initial budget expired")
					}
					if mode == "caller-cancel-delayed-cleanup" {
						cancelCaller()
					}
					cancelBudget()
					if strings.Contains(mode, "cleanup") {
						// Model a scheduling pause after cleanup cancels the timer,
						// before Build performs its post-child classification.
						time.Sleep(time.Until(deadline) + 10*time.Millisecond)
					}
				},
			}
			if mode == "caller-cancel" {
				cancelCaller()
			}
			if mode == "deadline-before-caller-cancel" {
				<-budget.Done()
				cancelCaller()
			}
			message, receiveErr := client.receive(kindBuildResult)
			// This is the same ordering as Build's post-child bracket.
			err := errors.Join(client.deadlineFailure(), receiveErr)
			want := CodeWorkerProtocolInvalid
			var cause error
			switch mode {
			case "deadline-during-cleanup", "deadline-before-caller-cancel":
				want, cause = "process_timeout", context.DeadlineExceeded
			case "caller-cancel", "caller-cancel-delayed-cleanup":
				want, cause = "process_canceled", context.Canceled
			}
			if DiagnosticCode(err) != want {
				t.Errorf("cleanup classification=%v; want primary %s", err, want)
			}
			if cause != nil && !errors.Is(err, cause) {
				t.Errorf("cleanup classification lost cause %v: %v", cause, err)
			}
			if cause == nil && errors.Is(err, context.Canceled) {
				t.Errorf("cleanup introduced false caller cancellation: %v", err)
			}
			if cause == context.DeadlineExceeded && errors.Is(err, context.Canceled) {
				t.Errorf("cleanup replaced the deadline cause with cancellation: %v", err)
			}
			if mode != "caller-cancel" && mode != "deadline-before-caller-cancel" && !cleanupReached {
				t.Error("fixture did not exercise the production cleanup path")
			}
			if message.Result != nil {
				t.Error("failed response returned a build result")
			}
		})
	}
}
