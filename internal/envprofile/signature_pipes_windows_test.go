//go:build windows

package envprofile

import (
	"crypto/sha256"
	"errors"
	"fmt"
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
		signaturePipeTestCommand(executable, "curator-signature-pipe-child", "roundtrip", "/dev/fd/3", "/dev/fd/4"),
		[]byte("message"), []byte("policy"), []byte("signature"),
	)
	if err != nil || string(output) != "message|policy|signature" {
		t.Fatalf("pipe roundtrip output=%q err=%v", output, err)
	}
	assertReadWitnesses(t, root, before)

	_, err = runSignaturePipes(
		signaturePipeTestCommand(executable, "curator-signature-pipe-child", "exit-before-open", "/dev/fd/3"),
		[]byte("message"), []byte("policy"),
	)
	if !errors.Is(err, errSignerVerificationUnavailable) {
		t.Fatalf("early child exit error=%v, want transport unavailable", err)
	}
	t.Log("Windows named-pipe roundtrip and early-exit cancellation PASS")
	assertReadWitnesses(t, root, before)
}

// An attributes-only opener must not steal the single input stream. Each
// signature reader then consumes more than one kernel pipe buffer. Both
// input digests and the unchanged temp-root witness prove delivery and no disk
// fallback, including a policy reader that closes after one line. The adjacent
// real signer test qualifies actual verifier behavior.
func TestSignaturePipeMetadataProbeAndDrain(t *testing.T) {
	root := t.TempDir()
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(key, root)
	}
	before := readWitnesses(t, root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	input := []byte(strings.Repeat("signature", 128*1024))
	policy := []byte("policy\n")
	for _, mode := range []string{"probe-digest", "probe-policy-line"} {
		t.Run(mode, func(t *testing.T) {
			output, err := runSignaturePipes(signaturePipeTestCommand(executable, "curator-signature-pipe-child", mode, "/dev/fd/3", "/dev/fd/4"), []byte("message"), input, policy)
			want := fmt.Sprintf("message|%d:%x|%d:%x", len(input), sha256.Sum256(input), len(policy), sha256.Sum256(policy))
			if err != nil || string(output) != want {
				t.Fatalf("metadata probe stole signature input or drain lost bytes: %v %q", err, output)
			}
			t.Log("Windows metadata probe with full input and policy close control PASS")
		})
	}
	assertReadWitnesses(t, root, before)
	t.Log("Windows attributes-only probe, full multi-buffer delivery and read-only witness PASS")
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
