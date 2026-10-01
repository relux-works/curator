//go:build !windows

package buildrepo

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// Drive the real askpass dispatch, but wait for its exit before starting the
// secret server. No scheduler choice or race instrumentation is needed.
func TestHTTPSBrokerSecretTransportChildExitsBeforeServe(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapper, state, err := materializeHTTPSCredentialBroker(t.TempDir(), executable,
		NewHTTPSCredentials("fixture.test", "oauth2", "refusal-secret"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		args   []string
		code   int
		output string
	}{
		{"foreign_user", []string{"Password for 'https://root@fixture.test': "}, 1, ""},
		{"bare_prompt", []string{"Password: "}, 1, ""},
		{"no_arguments", nil, 1, ""},
		{"extra_argument", []string{"Username for 'https://fixture.test': ", "extra"}, 1, ""},
		{"username", []string{"Username for 'https://fixture.test': "}, 0, "oauth2\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(wrapper, tc.args...)
			transport, err := NewHTTPSBrokerSecretTransport(cmd, "refusal-secret")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = transport.Close() })
			cmd.Env = []string{EnvHTTPSBrokerState + "=" + state,
				EnvHTTPSBrokerTransport + "=" + transport.EnvironmentValue()}
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			transport.ChildStarted()
			waitErr := cmd.Wait()
			var exitErr *exec.ExitError
			if waitErr != nil && !errors.As(waitErr, &exitErr) {
				t.Fatal(waitErr)
			}
			if cmd.ProcessState.ExitCode() != tc.code || stdout.String() != tc.output || stderr.Len() != 0 {
				t.Fatalf("unexpected child result: code=%d; output redacted", cmd.ProcessState.ExitCode())
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := transport.Serve(ctx); err != nil {
				t.Fatalf("HTTPS broker secret transport after non-reading child exit: %v", err)
			}
			if ctx.Err() != nil {
				t.Fatal("transport did not observe child exit")
			}
		})
	}
}

func TestHTTPSBrokerSecretTransportReportsWriteFailureAfterRequest(t *testing.T) {
	transport, err := NewHTTPSBrokerSecretTransport(exec.Command(os.Args[0]), "transport-secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transport.Close() })
	unix := transport.(*unixHTTPSBrokerSecretTransport)
	// The requesting endpoint stays open for reading. Break only the server's
	// write direction, so a real transport failure cannot look like no request.
	if _, err := unix.reader.WriteString("P"); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Shutdown(int(unix.writer.Fd()), syscall.SHUT_WR); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := transport.Serve(ctx); !errors.Is(err, syscall.EPIPE) {
		t.Fatalf("requested secret write error = %v, want EPIPE", err)
	}
}

func TestHTTPSBrokerSecretTransportRejectsInvalidRequest(t *testing.T) {
	transport, err := NewHTTPSBrokerSecretTransport(exec.Command(os.Args[0]), "transport-secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transport.Close() })
	unix := transport.(*unixHTTPSBrokerSecretTransport)
	if _, err := unix.reader.WriteString("X"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := transport.Serve(ctx); err == nil || err.Error() != "invalid HTTPS broker secret request" {
		t.Fatalf("invalid request error = %v", err)
	}
	var answer [1]byte
	if n, _ := unix.reader.Read(answer[:]); n != 0 {
		t.Fatal("invalid request received secret material")
	}
}

func TestHTTPSCredentialBrokerRejectsClosedSecretSocket(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapper, state, err := materializeHTTPSCredentialBroker(t.TempDir(), executable,
		NewHTTPSCredentials("fixture.test", "oauth2", "transport-secret"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, wrapper, "Password for 'https://oauth2@fixture.test': ")
	transport, err := NewHTTPSBrokerSecretTransport(cmd, "transport-secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transport.Close() })
	unix := transport.(*unixHTTPSBrokerSecretTransport)
	if err := syscall.Shutdown(int(unix.writer.Fd()), syscall.SHUT_WR); err != nil {
		t.Fatal(err)
	}
	cmd.Env = []string{EnvHTTPSBrokerState + "=" + state,
		EnvHTTPSBrokerTransport + "=" + transport.EnvironmentValue()}
	var output, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	transport.ChildStarted()
	waitErr := cmd.Wait()
	var exitErr *exec.ExitError
	if !errors.As(waitErr, &exitErr) || exitErr.ExitCode() != 1 || output.Len() != 0 || stderr.Len() != 0 || ctx.Err() != nil {
		t.Fatalf("closed socket returned code=%d; output redacted, want silent refusal", cmd.ProcessState.ExitCode())
	}
}

func TestHTTPSBrokerSecretTransportCancellationWithoutRequest(t *testing.T) {
	transport, err := NewHTTPSBrokerSecretTransport(exec.Command(os.Args[0]), "transport-secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transport.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := transport.Serve(ctx); err != nil {
		t.Fatalf("cancel unrequested secret: %v", err)
	}
}
