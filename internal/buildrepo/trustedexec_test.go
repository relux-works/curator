package buildrepo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Operator-selected absolute tools are admitted outside managed roots; a
// project/manager command, relative path and missing executable are refused.
// This witness never starts any selected executable and runs on Windows too.
func TestTrustedExecutableSelection(t *testing.T) {
	for _, name := range []string{"ssh", "git"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			tool := filepath.Join(root, name)
			if err := os.WriteFile(tool, []byte("fixture"), 0700); err != nil {
				t.Fatal(err)
			}
			if got, err := SelectTrustedExecutable(name, tool); err != nil || got != tool {
				t.Fatalf("positive: %q %v", got, err)
			}
			if _, err := SelectTrustedExecutable(name, tool, root); err == nil || !strings.Contains(err.Error(), "managed command root") {
				t.Fatalf("manager bin: %v", err)
			}
			project := filepath.Join(t.TempDir(), ".agents", "bin", name)
			if err := os.MkdirAll(filepath.Dir(project), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(project, []byte("foreign"), 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := SelectTrustedExecutable(name, project); err == nil || !strings.Contains(err.Error(), "managed command root") {
				t.Fatalf("project bin: %v", err)
			}
			if _, err := SelectTrustedExecutable(name, name); err == nil || !strings.Contains(err.Error(), "absolute") {
				t.Fatalf("relative: %v", err)
			}
			if _, err := SelectTrustedExecutable(name, tool+"-absent"); err == nil || err.Error() != "no "+name+" executable found" {
				t.Fatalf("missing: %v", err)
			}
			if got, _ := os.ReadFile(project); string(got) != "foreign" {
				t.Fatal("rejection changed foreign bytes")
			}
		})
	}
}

// A byte change after selection cannot pass the pinned Git admission check.
// The rejected file is never executed; its changed bytes are preserved.
func TestTrustedGitFingerprintChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(path, []byte("before"), 0700); err != nil {
		t.Fatal(err)
	}
	before, err := ExecutableFingerprint(path)
	if err != nil {
		t.Fatal(err)
	}
	if again, err := ExecutableFingerprint(path); err != nil || again != before {
		t.Fatalf("positive: %s %v", again, err)
	}
	if err := os.WriteFile(path, []byte("after"), 0700); err != nil {
		t.Fatal(err)
	}
	err = ValidateGitTool(context.Background(), GitTool{Executable: path, Fingerprint: before})
	if ErrorCode(err) != CodeCredentialPolicyInvalid || !strings.Contains(err.Error(), "fingerprint changed") {
		t.Fatalf("changed executable: %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != "after" {
		t.Fatal("changed bytes lost")
	}
}

// An absent SSH executable has one stable primary diagnostic, independently of
// missing secondary inputs. A complete fixture policy is the adjacent control.
func TestNoSSHExecutableDiagnostic(t *testing.T) {
	source, _ := ParseSource("git@example.test:skills/tool.git")
	for i := 0; i < 100; i++ {
		_, err := bindSSHWrapperPolicy(GitTool{}, source)
		if ErrorCode(err) != CodeCredentialPolicyInvalid || err.Error() != CodeCredentialPolicyInvalid+": no ssh executable found" {
			t.Fatalf("unstable diagnostic: %v", err)
		}
	}
}

// Static authentication and host verification labels survive the pipeline as
// typed reasons. An ordinary connection refusal remains a transport failure;
// unrelated message text cannot promote an untyped error to either category.
func TestCredentialFailureTypedReasons(t *testing.T) {
	for _, c := range []struct {
		stderr string
		reason AcquisitionReason
		code   string
	}{
		{"Permission denied (publickey).", ReasonAuthenticationFailed, CodeCredentialPolicyInvalid},
		{"Host key verification failed.", ReasonHostKeyFailed, CodeCredentialPolicyInvalid},
		{"Connection refused", ReasonTransportFailed, CodeSourceUnavailable},
	} {
		_, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Acquire: func(context.Context) (*Snapshot, error) { return nil, laneDiagnostic(&fetchError{stderr: c.stderr}) }})
		var typed *AdmissionError
		if !errors.As(err, &typed) || typed.Reason != c.reason || typed.Code != c.code || strings.Contains(err.Error(), c.stderr) {
			t.Fatalf("typed refusal: %v", err)
		}
	}
	untyped := admissionError(CodeIdentityInvalid, "text (authentication_failed)")
	_, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Acquire: func(context.Context) (*Snapshot, error) { return nil, untyped }})
	if ErrorCode(err) == CodeCredentialPolicyInvalid {
		t.Fatalf("message text routed the error: %v", err)
	}
}

// OpenSSH 8.9's complete unknown-host diagnostic is a static host-key refusal,
// never a fallback grant. Unrecognized neighboring text still fails closed as
// unknown; a recognized authentication sentence retains its separate class.
func TestOpenSSH89UnknownHostDiagnostic(t *testing.T) {
	const unknownHost = "No ED25519 host key is known for 127.0.0.1 and you have requested strict checking.\nHost key verification failed.\n"
	for _, c := range []struct {
		text  string
		class FailureClass
	}{
		{unknownHost, FailureHostKey},
		{unknownHost + "unrecognized detail\n", FailureUnknown},
		{"No invented host key is known for 127.0.0.1 and you have requested strict checking.\nHost key verification failed.\n", FailureUnknown},
		{"Permission denied (publickey).\n", FailureAuth},
	} {
		if got := ClassifyFetchOutput(c.text); got != c.class {
			t.Fatalf("class = %v, want %v", got, c.class)
		}
		if c.class != FailureAuth && AllowSecondAttempt(TransportFallbackAvailabilityAuth, c.class) {
			t.Fatalf("host/unknown failure permitted fallback: %v", c.class)
		}
	}
	err := laneDiagnostic(&fetchError{stderr: unknownHost})
	var typed *AdmissionError
	if !errors.As(err, &typed) || typed.Code != CodeCredentialPolicyInvalid || typed.Reason != ReasonHostKeyFailed || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("sanitized host-key reason: %v", err)
	}
}
