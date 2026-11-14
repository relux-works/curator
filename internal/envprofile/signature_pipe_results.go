package envprofile

import (
	"errors"
	"fmt"
)

var errSignatureVerifierExited = errors.New("signature verifier exited before consuming verification input")

type signaturePipeResult struct {
	input     int
	delivered bool
	written   bool
	connected bool
	fatal     bool
	err       error
}

// Metadata probes are not delivery. A drained connection proves one input was
// consumed. After one drained input, a verifier can close a complete later
// input without reading EOF (SSH's one-line policy), or leave it unopened on
// failure (SSH rejects tampering before policy). A complete write alone never
// qualifies an operation. Partial writes and security errors stay unavailable.
func signaturePipeResults(results []signaturePipeResult, inputCount int, runErr error) error {
	delivered := make([]bool, inputCount)
	connected := make([]bool, inputCount)
	written := make([]bool, inputCount)
	inputErrors := make([]error, inputCount)
	var fatal error
	for _, result := range results {
		if result.input < 0 || result.input >= inputCount {
			return fmt.Errorf("%w: invalid signature input result", errSignerVerificationUnavailable)
		}
		delivered[result.input] = delivered[result.input] || result.delivered
		connected[result.input] = connected[result.input] || result.connected
		written[result.input] = written[result.input] || result.written
		inputErrors[result.input] = errors.Join(inputErrors[result.input], result.err)
		if result.fatal {
			fatal = errors.Join(fatal, result.err)
		}
	}
	anyDelivered := false
	for _, value := range delivered {
		anyDelivered = anyDelivered || value
	}
	for index, value := range delivered {
		if !value && (!anyDelivered || (connected[index] && !written[index]) || (!connected[index] && runErr == nil)) {
			fatal = errors.Join(fatal, fmt.Errorf("signature input %d was not consumed: %v", index, inputErrors[index]))
		}
	}
	if fatal != nil {
		return fmt.Errorf("%w: verification input pipe: %v (verifier: %v)", errSignerVerificationUnavailable, fatal, runErr)
	}
	return runErr
}
