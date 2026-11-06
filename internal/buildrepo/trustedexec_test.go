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

// Complete Git/OpenSSH authentication and host-verification lines retain typed
// reasons. Bare or tailed HTTP text remains unknown under the closed grammar.
// Only SSH authentication gets the provider remedy; recognized HTTPS 401/403
// controls must not. Every refusal stops before audit/build or artifact output.
func TestCredentialFailureTypedReasons(t *testing.T) {
	for _, c := range []struct {
		name      string
		stderr    string
		transport string
		reason    AcquisitionReason
		code      string
	}{
		{"ssh-auth", "Permission denied (publickey).", "ssh", ReasonAuthenticationFailed, CodeCredentialPolicyInvalid},
		{"https-401", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 401", "https", ReasonAuthenticationFailed, CodeCredentialPolicyInvalid},
		{"https-403", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 403", "https", ReasonAuthenticationFailed, CodeCredentialPolicyInvalid},
		{"https-bare-401", "The requested URL returned error: 401", "https", ReasonTransportFailed, CodeSourceUnavailable},
		{"https-tailed-401", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 401 unparsed", "https", ReasonTransportFailed, CodeSourceUnavailable},
		{"ssh-host-key", "Host key verification failed.", "ssh", ReasonHostKeyFailed, CodeCredentialPolicyInvalid},
		{"ssh-connection", "Connection refused", "ssh", ReasonTransportFailed, CodeSourceUnavailable},
	} {
		t.Run(c.name, func(t *testing.T) {
			var phases []string
			result, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Trace: func(phase string) {
				phases = append(phases, phase)
			}, Acquire: func(context.Context) (*Snapshot, error) {
				return nil, laneDiagnostic(&fetchError{stderr: c.stderr}, c.transport)
			}})
			var typed *AdmissionError
			if !errors.As(err, &typed) || typed.Reason != c.reason || typed.Code != c.code || strings.Contains(err.Error(), c.stderr) {
				t.Fatalf("typed refusal: %v", err)
			}
			if c.reason == ReasonAuthenticationFailed && c.transport == "ssh" && (strings.Count(err.Error(), ";") != 1 || !strings.Contains(err.Error(), "security-key identity needs a client with a provider, selected with CURATOR_BUILD_SSH_EXECUTABLE")) {
				t.Fatalf("static authentication remedy: %v", err)
			}
			if c.transport == "https" && (strings.Contains(err.Error(), "CURATOR_BUILD_SSH_EXECUTABLE") || strings.Contains(err.Error(), "security-key identity")) {
				t.Fatalf("HTTPS got SSH remedy: %v", err)
			}
			if len(phases) != 1 || phases[0] != "exact-source-acquisition" || len(result.Artifact) != 0 || len(result.Receipt) != 0 || result.SnapshotKey != "" || result.CacheKey != "" {
				t.Fatalf("refusal advanced the pipeline or produced output: phases=%v result=%+v", phases, result)
			}
		})
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
	// A zero tool and production discovery errors both reach their policy
	// diagnostic even when failed discovery also left the broker absent.
	for _, tool := range []GitTool{
		{},
		{SelectionError: TrustedExecutableSelectionError(errors.New("no git executable found"))},
		{SelectionError: TrustedExecutableSelectionError(errors.New("git executable must be absolute"))},
	} {
		want := "no git executable found"
		if tool.SelectionError != nil {
			want = tool.SelectionError.Error()
		}
		for i := 0; i < 100; i++ {
			snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Tool: tool, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}})
			if snapshot != nil {
				t.Fatal("failed Git selection produced a snapshot")
			}
			if ErrorCode(err) != CodeCredentialPolicyInvalid || !strings.Contains(err.Error(), want) {
				t.Fatalf("Git selection: %v, want %q", err, want)
			}
		}
	}
	// With a selected executable, the neighboring missing-broker refusal
	// remains reachable without executing the ordinary-file fixture.
	snapshot, brokerErr := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Tool: GitTool{Executable: p}, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}})
	if snapshot != nil || ErrorCode(brokerErr) != CodeIdentityInvalid || !strings.Contains(brokerErr.Error(), "HTTPS requires a manager credential broker") {
		t.Fatalf("missing broker: snapshot=%v error=%v", snapshot, brokerErr)
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

// SSH admission preserves credential, Git selection, then wrapper order. Missing,
// relative and managed-root selections produce no snapshot or file mutation;
// the selected-file control reaches the neighboring wrapper refusal.
func TestSSHGitSelectionPrecedence(t *testing.T) {
	source, err := ParseSource("git@fixture.test:tools.git")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(p, []byte("unchanged"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, reason := range []string{"", "git executable must be absolute", "git executable is inside a managed command root"} {
		tool := GitTool{}
		want := "no git executable found"
		if reason != "" {
			tool.SelectionError = TrustedExecutableSelectionError(errors.New(reason))
			want = reason
		}
		request := NetworkRequest{Source: source, Tool: tool, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}}
		snapshot, err := AcquireNetwork(context.Background(), request)
		if snapshot != nil || ErrorCode(err) != CodeSSHCredentialMissing {
			t.Fatalf("credential precedence: %v", err)
		}
		request.Tool.SSHCredentials.Identity = p
		snapshot, err = AcquireNetwork(context.Background(), request)
		if snapshot != nil || ErrorCode(err) != CodeCredentialPolicyInvalid || !strings.Contains(err.Error(), want) {
			t.Fatalf("selection %q: %v", want, err)
		}
	}
	snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Tool: GitTool{Executable: p, SSHCredentials: OperatorSSHCredentials{Identity: p}}, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}})
	if snapshot != nil || !strings.Contains(err.Error(), "SSH requires the exact manager wrapper") {
		t.Fatalf("selected control: %v", err)
	}
	if got, err := os.ReadFile(p); err != nil || string(got) != "unchanged" {
		t.Fatalf("fixture changed: %v", err)
	}
}
