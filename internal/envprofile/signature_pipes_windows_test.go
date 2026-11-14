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

// Private regular-file transport preserves bytes and cleans staging even when
// the child exits without reading inputs. Temp-parent mtime changes are allowed.
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
	assertSignatureScratch(t, root, before)

	_, err = runSignaturePipes(
		signaturePipeTestCommand(executable, "curator-signature-pipe-child", "exit-before-open", "/dev/fd/3"),
		[]byte("message"), []byte("policy"),
	)
	if err == nil || errors.Is(err, errSignerVerificationUnavailable) {
		t.Fatalf("early child exit error=%v, want completed verifier exit", err)
	}
	t.Log("Windows private-file roundtrip, cleanup and completed verifier exit PASS")
	assertSignatureScratch(t, root, before)
}

// Metadata opens and multi-buffer reads work on the regular staging files,
// including a policy reader that closes after one line. No staging remains.
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
			t.Log("Windows regular-file metadata open and full input/policy close control PASS")
		})
	}
	assertSignatureScratch(t, root, before)
	t.Log("Windows regular-file metadata/multi-buffer delivery and cleanup PASS")
}

func signaturePipeTestCommand(executable string, args ...string) *exec.Cmd {
	// #nosec G204 -- the executable is the current test binary; arguments are fixed test controls and internal pipe paths.
	return exec.Command(executable, args...)
}
