//go:build windows

package envprofile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// This proves the named-pipe transport preserves bytes, leaves its selected
// temp root unchanged, and returns when a child exits before opening a pipe.
// Signer vectors cover verifier compatibility.
func TestSignaturePipeTransportReadOnlyAndCancels(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TMPDIR", root)
	t.Setenv("TMP", root)
	t.Setenv("TEMP", root)
	before := readWitnesses(t, root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	output, err := runSignaturePipes(
		signaturePipeTestCommand(executable, "-test.run=^TestSignaturePipeChild$", "curator-signature-pipe-child", "roundtrip", "/dev/fd/3", "/dev/fd/4"),
		[]byte("message"), []byte("policy"), []byte("signature"),
	)
	if err != nil || string(output) != "message|policy|signature" {
		t.Fatalf("pipe roundtrip output=%q err=%v", output, err)
	}
	assertReadWitnesses(t, root, before)

	_, err = runSignaturePipes(
		signaturePipeTestCommand(executable, "-test.run=^TestSignaturePipeChild$", "curator-signature-pipe-child", "exit-before-open", "/dev/fd/3"),
		[]byte("message"), []byte("policy"),
	)
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("early child exit error=%v, want verifier exit error", err)
	}
	assertReadWitnesses(t, root, before)
}

// This proves only canonical placeholders that have matching supplied inputs are accepted, with valid placeholders as the control.
func TestSignaturePipeArgsRejectUnavailableInputs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		inputCount int
		wantError  bool
	}{
		{name: "available inputs", args: []string{"-f", "/dev/fd/3", "-s", "/dev/fd/4"}, inputCount: 2},
		{name: "missing input argument", args: []string{"-f", "/dev/fd/3"}, inputCount: 2, wantError: true},
		{name: "noncanonical input", args: []string{"-f", "/dev/fd/3", "-s", "/dev/fd/03"}, inputCount: 1, wantError: true},
		{name: "unavailable input", args: []string{"-f", "/dev/fd/3", "-s", "/dev/fd/4", "-t", "/dev/fd/5"}, inputCount: 2, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := signaturePipeArgs(test.args, test.inputCount)
			if (err != nil) != test.wantError {
				t.Fatalf("signaturePipeArgs error = %v, wantError=%v", err, test.wantError)
			}
		})
	}
}

func signaturePipeTestCommand(executable string, args ...string) *exec.Cmd {
	// #nosec G204 -- the executable is the current test binary; arguments are fixed test controls and internal pipe paths.
	return exec.Command(executable, args...)
}

// This helper test checks exact bytes from stdin and named-pipe arguments when launched by the transport test.
func TestSignaturePipeChild(t *testing.T) {
	marker := -1
	for i, arg := range os.Args {
		if arg == "curator-signature-pipe-child" {
			marker = i
			break
		}
	}
	if marker < 0 || marker+1 >= len(os.Args) {
		return
	}
	if os.Args[marker+1] == "exit-before-open" {
		t.Fatal("intentional verifier child exit")
	}
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatal(err)
	}
	parts := []string{string(stdin)}
	for _, name := range os.Args[marker+2:] {
		file, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(errors.Join(readErr, closeErr))
		}
		parts = append(parts, string(data))
	}
	if _, err := fmt.Fprint(os.Stdout, strings.Join(parts, "|")); err != nil {
		t.Fatal(err)
	}
}
