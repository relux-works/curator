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
// File identity covers native path aliases. This witness never starts any
// selected executable and runs on Windows too.
func TestTrustedExecutableSelection(t *testing.T) {
	for _, name := range []string{"ssh", "git"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			tool := filepath.Join(root, name)
			if err := os.WriteFile(tool, []byte("fixture"), 0700); err != nil {
				t.Fatal(err)
			}
			got, err := SelectTrustedExecutable(name, tool)
			if err != nil {
				t.Fatal(err)
			}
			selected, selectedErr := os.Stat(got)
			original, originalErr := os.Stat(tool)
			if selectedErr != nil || originalErr != nil || !os.SameFile(selected, original) {
				t.Fatalf("positive selected a different file: %q %v %v", got, selectedErr, originalErr)
			}
			if _, err := SelectTrustedExecutable(name, tool, root); err == nil || !strings.Contains(err.Error(), "managed command root") {
				t.Fatalf("manager bin: %v", err)
			}
			for _, segment := range []string{".agents/bin", ".curator/bin", "global/bin", "environments/example/bin", "runtime/example/bin"} {
				project := filepath.Join(t.TempDir(), filepath.FromSlash(segment), name)
				if err := os.MkdirAll(filepath.Dir(project), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(project, []byte("foreign"), 0700); err != nil {
					t.Fatal(err)
				}
				if _, err := SelectTrustedExecutable(name, project); err == nil || !strings.Contains(err.Error(), "managed command root") {
					t.Fatalf("project bin: %v", err)
				}
				if got, _ := os.ReadFile(project); string(got) != "foreign" {
					t.Fatal("rejection changed foreign bytes")
				}
			}
			if _, err := SelectTrustedExecutable(name, name); err == nil || !strings.Contains(err.Error(), "absolute") {
				t.Fatalf("relative: %v", err)
			}
			if _, err := SelectTrustedExecutable(name, tool+"-absent"); err == nil || err.Error() != "no "+name+" executable found" {
				t.Fatalf("missing: %v", err)
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
		if c.reason == ReasonAuthenticationFailed && (strings.Count(err.Error(), ";") != 1 || !strings.Contains(err.Error(), "security-key identity needs a client with a provider, selected with CURATOR_BUILD_SSH_EXECUTABLE")) {
			t.Fatalf("static authentication remedy: %v", err)
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

// Absent Git has one stable policy diagnostic before secondary paths on HTTPS
// too; a fixed ordinary-file fixture then reaches the exec-path check without
// starting the fixture. Both paths preserve its bytes.
func TestTrustedGitSelectionDiagnostics(t *testing.T) {
	p := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(p, []byte("not executable code"), 0700); err != nil {
		t.Fatal(err)
	}
	source, parseErr := ParseSource("https://fixture.test/tools.git")
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	for i := 0; i < 100; i++ {
		snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}})
		if snapshot != nil {
			t.Fatal("absent Git produced a snapshot")
		}
		if ErrorCode(err) != CodeCredentialPolicyInvalid || !strings.Contains(err.Error(), "no git executable found") {
			t.Fatalf("missing Git: %v", err)
		}
	}
	err := ValidateGitTool(context.Background(), GitTool{Executable: p})
	if ErrorCode(err) != CodeIdentityInvalid || !strings.Contains(err.Error(), "git exec path path must be absolute") {
		t.Fatalf("ordered positive: %v", err)
	}
	if got, _ := os.ReadFile(p); string(got) != "not executable code" {
		t.Fatal("fixture changed")
	}
}

// The real OpenSSH changed-key banner retains the typed host-key reason.
// Unknown adjacent text and truncated evidence never become host-key proof;
// neither verdict authorizes fallback or exposes remote details.
func TestOpenSSH89ChangedHostDiagnostic(t *testing.T) {
	const changed = `@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now (man-in-the-middle attack)!
It is also possible that a host key has just been changed.
The fingerprint for the ED25519 key sent by the remote host is
SHA256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.
Please contact your system administrator.
Add correct host key in /fixture/known-hosts to get rid of this message.
Offending ED25519 key in /fixture/known-hosts:1
Host key for 127.0.0.1 has changed and you have requested strict checking.
Host key verification failed.
`
	// The image's Ubuntu 8.9 package patches in this advisory pair. A
	// neighboring command spelling is unknown and cannot inherit its label.
	ubuntu := strings.Replace(changed, "Offending ED25519 key in /fixture/known-hosts:1\n", "Offending ED25519 key in /fixture/known-hosts:1\n  remove with:\n  ssh-keygen -f \"/fixture/known-hosts\" -R \"127.0.0.1\"\n", 1)
	for _, c := range []struct {
		output    string
		truncated bool
		reason    AcquisitionReason
	}{{changed, false, ReasonHostKeyFailed}, {ubuntu, false, ReasonHostKeyFailed}, {strings.Replace(ubuntu, "ssh-keygen -f", "foreign-command -f", 1), false, ReasonTransportFailed}, {changed + "unknown detail", false, ReasonTransportFailed}, {changed, true, ReasonTransportFailed}} {
		err := laneDiagnostic(&fetchError{stderr: c.output, truncated: c.truncated})
		var typed *AdmissionError
		if !errors.As(err, &typed) || typed.Reason != c.reason || strings.Contains(err.Error(), "/fixture") {
			t.Fatalf("reason: %v", err)
		}
		if AllowSecondAttempt(TransportFallbackAvailabilityAuth, ClassifyFetchOutput(c.output)) {
			t.Fatal("host key permitted fallback")
		}
	}
}
