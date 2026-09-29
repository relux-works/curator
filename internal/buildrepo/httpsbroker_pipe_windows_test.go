//go:build windows

package buildrepo

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func newWindowsTestBrokerPipe(t *testing.T, secret string) *windowsHTTPSBrokerSecretTransport {
	t.Helper()
	transport, err := NewHTTPSBrokerSecretTransport(exec.Command("unused"), secret)
	if err != nil {
		t.Fatal(err)
	}
	pipe, ok := transport.(*windowsHTTPSBrokerSecretTransport)
	if !ok {
		t.Fatalf("transport type = %T, want Windows named pipe", transport)
	}
	return pipe
}

func TestHTTPSBrokerNamedPipeHasCurrentUserOnlyDACL(t *testing.T) {
	transport := newWindowsTestBrokerPipe(t, "private-secret")
	defer transport.Close()
	if !validHTTPSBrokerPipeName(transport.name) || strings.Contains(transport.name, "private-secret") {
		t.Fatalf("pipe name is not an unguessable transport identifier: %q", transport.name)
	}
	descriptor, err := windows.GetSecurityInfo(
		transport.handle,
		windows.SE_KERNEL_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		t.Fatalf("GetSecurityInfo on named-pipe handle: %v", err)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatalf("read named-pipe DACL: %v", err)
	}
	if dacl == nil || dacl.AceCount != 1 {
		t.Fatalf("named-pipe DACL has %v ACEs, want exactly one current-user ACE", dacl)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := descriptor.Control()
	if err != nil {
		t.Fatalf("read named-pipe security descriptor control: %v", err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("named-pipe DACL is not protected: %q", descriptor.String())
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(dacl, 0, &ace); err != nil {
		t.Fatalf("GetAce(0) on named-pipe DACL: %v", err)
	}
	if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
		t.Fatalf("named-pipe DACL ACE type = %d, want ACCESS_ALLOWED", ace.Header.AceType)
	}
	aceSID := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if !aceSID.Equals(user.User.Sid) {
		t.Fatalf("named-pipe DACL grants %s, want only the current user %s: %q", aceSID, user.User.Sid, descriptor.String())
	}
	for _, name := range []string{"S-1-1-0", "S-1-5-32-545", "S-1-5-11"} {
		forbidden, err := windows.StringToSid(name)
		if err != nil {
			t.Fatal(err)
		}
		if aceSID.Equals(forbidden) {
			t.Fatalf("named-pipe DACL contains forbidden trustee %s: %q", name, descriptor.String())
		}
	}
}

func TestHTTPSBrokerNamedPipeServesOneClientAndThenRejectsAnother(t *testing.T) {
	const secret = "one-client-only-secret"
	transport := newWindowsTestBrokerPipe(t, secret)
	serveCtx, cancelServe := context.WithCancel(context.Background())
	defer cancelServe()
	serveDone := make(chan error, 1)
	go func() { serveDone <- transport.Serve(serveCtx) }()

	first, err := readWindowsBrokerPipe(transport.name)
	if err != nil || string(first) != secret {
		t.Fatalf("first named-pipe client got %q, err=%v; want secret", first, err)
	}
	secondAttemptDeadline := time.Now().Add(time.Second)
	for {
		second, err := windows.CreateFile(
			windows.StringToUTF16Ptr(transport.name), windows.GENERIC_READ, 0, nil,
			windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0,
		)
		if err == nil {
			_ = windows.CloseHandle(second)
			t.Fatal("second named-pipe client connected after the secret was served")
		}
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			break
		}
		if !errors.Is(err, windows.ERROR_PIPE_BUSY) || time.Now().After(secondAttemptDeadline) {
			t.Fatalf("second named-pipe connection failed unexpectedly: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancelServe()
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("serve first client: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("named-pipe server did not stop after the first client or fetch cancellation")
	}
}

func TestHTTPSBrokerNamedPipeServerStopsWhenFetchExitsWithoutAClient(t *testing.T) {
	transport := newWindowsTestBrokerPipe(t, "unused-secret")
	serveCtx, cancelServe := context.WithCancel(context.Background())
	serveDone := make(chan error, 1)
	go func() { serveDone <- transport.Serve(serveCtx) }()
	cancelServe()
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("cancelled named-pipe server: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("named-pipe server did not stop after fetch cancellation")
	}
}

func readWindowsBrokerPipe(name string) ([]byte, error) {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(name), windows.GENERIC_READ, 0, nil,
		windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0,
	)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle)
	buffer := make([]byte, pipeBufferSize)
	var read uint32
	err = windows.ReadFile(handle, buffer, &read, nil)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buffer[:read]...), nil
}
