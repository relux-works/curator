package envprofile

import (
	"errors"
	"testing"
)

// Actual delivery is required even for exit zero. A metadata connection or
// partial read is not consumption; a security failure dominates another valid
// delivery. An unused later input is allowed only on unsuccessful execution
// after another input was consumed. These are result-policy controls, not
// Windows pipe execution; the native transport and real signers cover that.
func TestSignaturePipeResultControls(t *testing.T) {
	exited := errors.New("verifier exit")
	for _, tc := range []struct {
		name        string
		results     []signaturePipeResult
		count       int
		runErr      error
		unavailable bool
	}{
		{"consumed", []signaturePipeResult{{input: 0, delivered: true, connected: true}}, 1, nil, false},
		{"consumed-with-probe", []signaturePipeResult{{input: 0, connected: true, err: errSignatureVerifierExited}, {input: 0, delivered: true, connected: true}}, 1, nil, false},
		{"metadata-only", []signaturePipeResult{{input: 0, connected: true}}, 1, exited, true},
		{"metadata-only-written", []signaturePipeResult{{input: 0, connected: true, written: true}}, 1, nil, true},
		{"partial", []signaturePipeResult{{input: 0, connected: true, err: exited}}, 1, exited, true},
		{"unused-after-negative", []signaturePipeResult{{input: 0, delivered: true}, {input: 1, err: errSignatureVerifierExited}}, 2, exited, false},
		{"unused-after-success", []signaturePipeResult{{input: 0, delivered: true}, {input: 1, err: errSignatureVerifierExited}}, 2, nil, true},
		{"policy-close-before-eof", []signaturePipeResult{{input: 0, delivered: true}, {input: 1, connected: true, written: true}}, 2, nil, false},
		{"policy-close-negative", []signaturePipeResult{{input: 0, delivered: true}, {input: 1, connected: true, written: true}}, 2, exited, false},
		{"partial-after-consumed", []signaturePipeResult{{input: 0, delivered: true}, {input: 1, connected: true}}, 2, exited, true},
		{"no-input-consumed", []signaturePipeResult{{input: 0, err: errSignatureVerifierExited}}, 1, exited, true},
		{"security-failure", []signaturePipeResult{{input: 0, delivered: true}, {input: 0, fatal: true, err: errors.New("unexpected process")}}, 1, nil, true},
		{"out-of-range", []signaturePipeResult{{input: 1, delivered: true}}, 1, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := signaturePipeResults(tc.results, tc.count, tc.runErr)
			if tc.unavailable {
				if !errors.Is(err, errSignerVerificationUnavailable) {
					t.Fatalf("undelivered input became a verdict: %v", err)
				}
			} else if err != tc.runErr {
				t.Fatalf("consumed input verdict lost: %v", err)
			}
		})
	}
}
