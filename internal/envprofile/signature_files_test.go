package envprofile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Real child reads prove ordinary pathname compatibility, private directory/file
// boundaries and cleanup after success, rejection and failed launch setup. The
// temp parent's mtime is allowed to change; managed/keyring witnesses are not.
func TestSignatureFileTransport(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, root)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, mode string
		inputs     [][]byte
		valid      bool
	}{
		{"no-input", "negative-signature", nil, false},
		{"private-roundtrip", "private-roundtrip", [][]byte{[]byte("policy"), []byte("signature")}, true},
		{"early-exit", "verifier-failure", [][]byte{[]byte("policy")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"curator-signature-pipe-child", tc.mode}
			for i := range tc.inputs {
				if i == 0 {
					args = append(args, "/dev/fd/3")
				} else {
					args = append(args, "/dev/fd/4")
				}
			}
			cmd := exec.Command(executable, args...)
			out, err := runSignatureFiles(cmd, []byte("message"), tc.inputs...)
			if tc.valid {
				if err != nil || string(out) != "message|policy|signature" {
					t.Fatalf("private regular file roundtrip failed: %q %v", out, err)
				}
			} else {
				var exit *exec.ExitError
				if !errors.As(err, &exit) || !exit.Exited() || errors.Is(err, errSignerVerificationUnavailable) {
					t.Fatalf("expected completed verifier exit: %v", err)
				}
				if tc.name == "no-input" {
					if !signatureSSHNegativeExit(exit.ExitCode()) || signatureVerifierResult("SSH", out, err) != nil {
						t.Fatalf("no-input negative control did not complete: %v %s", err, out)
					}
				} else if exit.ExitCode() != 1 || !strings.Contains(string(out), "verifier unavailable") {
					t.Fatalf("early exit control did not complete: %v %s", err, out)
				}
			}
			entries, e := os.ReadDir(root)
			if e != nil || len(entries) != 0 {
				t.Fatalf("signature staging leaked: %v %v", entries, e)
			}
		})
	}
	command := exec.Command(filepath.Join(root, "missing-verifier"), "/dev/fd/3")
	if _, err := runSignatureFiles(command, nil, []byte("input")); err == nil {
		t.Fatal("missing verifier admitted")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed-start staging leaked: %v %v", entries, err)
	}
	missing := filepath.Join(root, "missing-parent")
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, missing)
	}
	command = exec.Command(executable, "curator-signature-pipe-child", "roundtrip", "/dev/fd/3")
	if _, err := runSignatureFiles(command, nil, []byte("input")); !errors.Is(err, errSignerVerificationUnavailable) || command.Process != nil {
		t.Fatalf("staging failure ran verifier or became a verdict: %v", err)
	}
	t.Logf("%s regular private signature input, cleanup and staging-refusal controls PASS", runtime.GOOS)
}

// Real OpenSSH accepts ordinary private files and rejects altered bytes, an
// unauthorized principal and a mismatched namespace. Each rejection is a
// failed signature verdict rather than verifier unavailability; no input stays.
func TestSignatureFilesRealSSH(t *testing.T) {
	key := filepath.Join(t.TempDir(), "signer")
	if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key).CombinedOutput(); err != nil {
		t.Fatalf("fixture key: %v %s", err, out)
	}
	public, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(public))
	allowed := []byte("fixture " + fields[0] + " " + fields[1] + "\n")
	signer := exec.Command("ssh-keygen", "-Y", "sign", "-n", "git", "-f", key)
	signer.Stdin = strings.NewReader("signed bytes\n")
	signature, err := signer.Output()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, root)
	}
	for _, tc := range []struct {
		name, payload, principal, namespace string
		valid                               bool
	}{
		{"valid", "signed bytes\n", "fixture", "git", true},
		{"tampered", "changed bytes\n", "fixture", "git", false},
		{"wrong-principal", "signed bytes\n", "other", "git", false},
		{"wrong-namespace", "signed bytes\n", "fixture", "not-git", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := exec.Command("ssh-keygen", "-Y", "verify", "-f", "/dev/fd/3", "-I", tc.principal, "-n", tc.namespace, "-s", "/dev/fd/4")
			output, runErr := runSignatureFiles(command, []byte(tc.payload), allowed, signature)
			err := signatureVerifierResult("SSH", output, runErr)
			if err != nil || (runErr == nil) != tc.valid {
				t.Fatalf("regular-file SSH verdict valid=%v: %v (%s)", tc.valid, err, output)
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("signature staging leaked: %v %v", entries, err)
			}
		})
	}
	t.Logf("%s real OpenSSH regular-file accept/reject controls PASS", runtime.GOOS)
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
