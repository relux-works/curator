//go:build windows

package envprofile

import "os/exec"

// Regular files work with native, MSYS and Cygwin verifiers; transient private
// staging replaces named pipes, which real Git-for-Windows SSH could not open.
func runSignaturePipes(command *exec.Cmd, payload []byte, inputs ...[]byte) ([]byte, error) {
	return runSignatureFiles(command, payload, inputs...)
}
