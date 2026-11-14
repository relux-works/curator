package envprofile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Dispatch before m.Run and fixture setup: the transport protocol contains only
// the selected payload, and starting the child cannot touch the observed temp root.
func signaturePipeChildEntry() bool {
	if len(os.Args) < 3 || os.Args[1] != "curator-signature-pipe-child" {
		return false
	}
	switch os.Args[2] {
	case "negative-signature":
		_, _ = fmt.Fprint(os.Stderr, "Could not verify signature.\n")
		os.Exit(-1)
	case "large-output":
		_, _ = fmt.Fprint(os.Stderr, strings.Repeat("x", 128*1024))
		os.Exit(0)
	case "exit-before-open":
		os.Exit(2)
	case "verifier-failure":
		_, _ = fmt.Fprint(os.Stderr, "verifier unavailable\x1b[31m\n")
		os.Exit(1)
	case "roundtrip":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			os.Exit(3)
		}
		parts := []string{string(input)}
		for _, name := range os.Args[3:] {
			file, err := os.Open(name)
			if err != nil {
				os.Exit(4)
			}
			data, readErr := io.ReadAll(file)
			closeErr := file.Close()
			if readErr != nil || closeErr != nil {
				os.Exit(5)
			}
			parts = append(parts, string(data))
		}
		if _, err := fmt.Fprint(os.Stdout, strings.Join(parts, "|")); err != nil {
			os.Exit(6)
		}
		os.Exit(0)
	}
	os.Exit(7)
	return true
}

// A real exited process is unavailable rather than a bad-signature verdict;
// recognized cryptographic/policy negatives and successful execution are controls.
func TestSignatureVerifierProcessFailures(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "curator-signature-pipe-child", "verifier-failure")
	output, runErr := runSignaturePipes(command, nil)
	var exit *exec.ExitError
	if !errors.As(runErr, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("process prerequisite: %v", runErr)
	}
	negativeCommand := exec.Command(executable, "curator-signature-pipe-child", "negative-signature")
	negativeOutput, negativeErr := runSignaturePipes(negativeCommand, nil)
	var negativeExit *exec.ExitError
	if !errors.As(negativeErr, &negativeExit) || !negativeExit.Exited() {
		t.Fatalf("negative process prerequisite: %v", negativeErr)
	}
	if err := signatureVerifierResult("SSH", negativeOutput, negativeErr); err != nil {
		t.Fatalf("normal negative status %d was not a negative verdict: %v", negativeExit.ExitCode(), err)
	}
	t.Logf("%s normal C-style negative process status=%d control PASS", runtime.GOOS, negativeExit.ExitCode())
	failedInput := exec.Command(executable, "curator-signature-pipe-child", "verifier-failure", "/dev/fd/3")
	inputOutput, inputErr := runSignaturePipes(failedInput, nil, make([]byte, 1024*1024))

	for _, kind := range []string{"SSH", "OpenPGP"} {
		t.Run(kind, func(t *testing.T) {
			err := signatureVerifierResult(kind, output, runErr)
			if !errors.Is(err, errSignerVerificationUnavailable) || !strings.Contains(err.Error(), `\x1b`) || strings.ContainsAny(err.Error(), "\n\r\x1b") {
				t.Fatalf("process failure was not safely unavailable: %v", err)
			}
			if err := signatureVerifierResult(kind, nil, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, tc := range []struct {
		name, kind, output string
		negative           bool
	}{
		{"ssh-policy", "SSH", "Could not verify signature.\n", true},
		{"ssh-cryptographic", "SSH", "Signature verification failed: incorrect signature\nCould not verify signature.\n", true},
		{"ssh-input", "SSH", "Couldn't read signature file: No such file\nCould not verify signature.\n", false},
		{"gpg-cryptographic", "OpenPGP", "[GNUPG:] BADSIG ABC fixture\n", true},
		{"gpg-missing-key", "OpenPGP", "[GNUPG:] ERRSIG ABC 1 2\n[GNUPG:] NO_PUBKEY ABC\n", false},
		{"gpg-failure-after-badsig", "OpenPGP", "[GNUPG:] BADSIG ABC fixture\n[GNUPG:] FAILURE verify 1\n", false},
		{"gpg-input", "OpenPGP", "cannot open signature\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := signatureVerifierResult(tc.kind, []byte(tc.output), runErr)
			if (err == nil) != tc.negative {
				t.Fatalf("verifier verdict negative=%v: %v", tc.negative, err)
			}
		})
	}
	if !errors.Is(inputErr, errSignerVerificationUnavailable) || !strings.Contains(string(inputOutput), "verifier unavailable") {
		t.Fatalf("input transport did not dominate process exit: %v %q", inputErr, inputOutput)
	}
	if err := signatureVerifierResult("SSH", []byte("Could not verify signature.\n"), errors.Join(errSignerVerificationUnavailable, runErr)); !errors.Is(err, errSignerVerificationUnavailable) {
		t.Fatalf("transport failure was a verdict: %v", err)
	}
	command = exec.Command(executable, "curator-signature-pipe-child", "large-output")
	captured, captureErr := runSignaturePipes(command, nil)
	if len(captured) != 64*1024 || !errors.Is(captureErr, errSignerVerificationUnavailable) {
		t.Fatalf("verifier output capture was not bounded: %d %v", len(captured), captureErr)
	}
	long := signatureDiagnostic([]byte(strings.Repeat("\x1b\n", 4096)))
	if len(long) > 1024 || !strings.Contains(long, "...") || strings.ContainsAny(long, "\x1b\n") {
		t.Fatalf("unbounded/unsafe diagnostic %q", long)
	}
	t.Logf("%s process-failure/unavailable and genuine negative controls PASS", runtime.GOOS)
}

// The no-fixture child is payload-only through the native descriptor transport;
// the Windows named-pipe counterpart exercises the same child entry.
func TestSignaturePipeChildProtocol(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("signature-child-protocol: native descriptor test; Windows named-pipe counterpart runs separately")
	}
	root := t.TempDir()
	t.Setenv("TMPDIR", root)
	t.Setenv("TMP", root)
	t.Setenv("TEMP", root)
	before := readWitnesses(t, root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "curator-signature-pipe-child", "roundtrip", "/dev/fd/3", "/dev/fd/4")
	out, err := runSignaturePipes(command, []byte("message"), []byte("policy"), []byte("signature"))
	if err != nil || string(out) != "message|policy|signature" {
		t.Fatalf("child protocol: %q %v", out, err)
	}
	assertReadWitnesses(t, root, before)
}

// A real SSH verifier that cannot open its signature is operationally
// unavailable; adjacent real signed/tampered vectors prove the positive and
// cryptographic-negative paths. This also exercises early pipe cancellation.
func TestSignatureVerifierRealInputFailure(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("read-only signature input transport unavailable")
	}
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen is unavailable")
	}
	missing := filepath.Join(t.TempDir(), "missing-signature")
	command := exec.Command("ssh-keygen", "-Y", "verify", "-f", "/dev/fd/3", "-I", "fixture", "-n", "git", "-s", missing)
	output, runErr := runSignaturePipes(command, []byte("message"), []byte("fixture ssh-ed25519 unused\n"))
	err := signatureVerifierResult("SSH", output, runErr)
	if runErr == nil || !errors.Is(err, errSignerVerificationUnavailable) || !strings.Contains(string(output), "signature") {
		t.Fatalf("real SSH input failure was not unavailable with stderr: %v %q", err, output)
	}
	t.Logf("%s real SSH input-failure/unavailable control PASS: %s", runtime.GOOS, signatureDiagnostic(output))
}
