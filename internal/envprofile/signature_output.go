package envprofile

import (
	"fmt"
	"os/exec"
)

// Drain stdout and stderr without retaining unbounded verifier output. os/exec
// serializes writes when both streams share the same comparable writer.
type signatureOutput struct {
	data      []byte
	truncated bool
}

func (output *signatureOutput) Write(data []byte) (int, error) {
	const limit = 64 * 1024
	size := len(data)
	if size > limit-len(output.data) {
		output.truncated = true
	}
	if remaining := limit - len(output.data); remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
		}
		output.data = append(output.data, data...)
	}
	return size, nil
}
func signatureCommandOutput(command *exec.Cmd) ([]byte, error) {
	var output signatureOutput
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	if output.truncated {
		return output.data, fmt.Errorf("%w: verifier output exceeded capture limit", errSignerVerificationUnavailable)
	}
	return output.data, err
}
