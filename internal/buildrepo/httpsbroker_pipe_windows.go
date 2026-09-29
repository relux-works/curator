//go:build windows

package buildrepo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// EnvHTTPSBrokerTransport names the secret-free Windows named pipe carried
	// to the fetch child.
	EnvHTTPSBrokerTransport = "CURATOR_BUILD_HTTPS_ASKPASS_PIPE"
	httpsBrokerPipePrefix   = `\\.\pipe\curator-https-askpass-`
	errorPipeConnected      = syscall.Errno(535)
	pipeBufferSize          = 4096
)

var connectNamedPipeProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("ConnectNamedPipe")

type windowsHTTPSBrokerSecretTransport struct {
	name      string
	handle    windows.Handle
	secret    []byte
	closeOnce sync.Once
	closeErr  error
}

// NewHTTPSBrokerSecretTransport creates a fetch-scoped private named pipe for
// the askpass secret and sets command's pipe-name environment value.
func NewHTTPSBrokerSecretTransport(command *exec.Cmd, secret string) (HTTPSBrokerSecretTransport, error) {
	if command == nil {
		return nil, errors.New("HTTPS broker command is nil")
	}
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("create HTTPS broker pipe name: %w", err)
	}
	name := httpsBrokerPipePrefix + hex.EncodeToString(nonce[:])
	descriptor, err := currentUserPipeSecurityDescriptor()
	if err != nil {
		return nil, err
	}
	attributes := windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: descriptor,
		InheritHandle:      0,
	}
	namePointer, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	handle, err := createHTTPSBrokerNamedPipe(namePointer, &attributes)
	runtime.KeepAlive(descriptor)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(namePointer)
	if err != nil {
		return nil, err
	}
	return &windowsHTTPSBrokerSecretTransport{name: name, handle: handle, secret: []byte(secret)}, nil
}

func (transport *windowsHTTPSBrokerSecretTransport) EnvironmentValue() string { return transport.name }

func (*windowsHTTPSBrokerSecretTransport) ChildStarted() {}

func (transport *windowsHTTPSBrokerSecretTransport) Serve(ctx context.Context) error {
	defer transport.Close()
	defer clear(transport.secret)
	connected, err := transport.connect(ctx)
	if err != nil || !connected {
		return err
	}
	return transport.write(ctx)
}

func (transport *windowsHTTPSBrokerSecretTransport) Close() error {
	transport.closeOnce.Do(func() {
		clear(transport.secret)
		if transport.handle != 0 && transport.handle != windows.InvalidHandle {
			transport.closeErr = windows.CloseHandle(transport.handle)
			transport.handle = 0
		}
	})
	return transport.closeErr
}

func (transport *windowsHTTPSBrokerSecretTransport) connect(ctx context.Context) (bool, error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(event)
	overlapped := windows.Overlapped{HEvent: event}
	result, _, callErr := connectNamedPipeProc.Call(uintptr(transport.handle), uintptr(unsafe.Pointer(&overlapped)))
	if result != 0 || errors.Is(callErr, errorPipeConnected) {
		return true, nil
	}
	if !errors.Is(callErr, windows.ERROR_IO_PENDING) {
		return false, callErr
	}
	completed, err := waitForPipeOperation(ctx, transport.handle, event, &overlapped)
	if err != nil || !completed {
		return false, err
	}
	var transferred uint32
	if err := windows.GetOverlappedResult(transport.handle, &overlapped, &transferred, false); err != nil {
		return false, err
	}
	return true, nil
}

func (transport *windowsHTTPSBrokerSecretTransport) write(ctx context.Context) error {
	return transport.writePayload(ctx, transport.secret)
}

func (transport *windowsHTTPSBrokerSecretTransport) writePayload(ctx context.Context, payload []byte) error {
	for len(payload) > 0 {
		if err := ctx.Err(); err != nil {
			return nil
		}
		event, err := windows.CreateEvent(nil, 1, 0, nil)
		if err != nil {
			return err
		}
		overlapped := windows.Overlapped{HEvent: event}
		var written uint32
		err = windows.WriteFile(transport.handle, payload, &written, &overlapped)
		if errors.Is(err, windows.ERROR_IO_PENDING) {
			completed, waitErr := waitForPipeOperation(ctx, transport.handle, event, &overlapped)
			if waitErr != nil || !completed {
				_ = windows.CloseHandle(event)
				return waitErr
			}
			err = windows.GetOverlappedResult(transport.handle, &overlapped, &written, false)
		}
		closeErr := windows.CloseHandle(event)
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}

func waitForPipeOperation(ctx context.Context, pipe windows.Handle, event windows.Handle, overlapped *windows.Overlapped) (bool, error) {
	for {
		if ctx.Err() != nil {
			_ = windows.CancelIoEx(pipe, overlapped)
			_, err := windows.WaitForSingleObject(event, windows.INFINITE)
			if err != nil {
				return false, err
			}
			return false, nil
		}
		status, err := windows.WaitForSingleObject(event, 50)
		if err != nil {
			return false, err
		}
		if status == windows.WAIT_OBJECT_0 {
			return true, nil
		}
		if status != uint32(windows.WAIT_TIMEOUT) {
			return false, fmt.Errorf("unexpected named-pipe wait status %d", status)
		}
	}
}

func createHTTPSBrokerNamedPipe(name *uint16, attributes *windows.SecurityAttributes) (windows.Handle, error) {
	const (
		pipeAccessOutbound      = uint32(0x00000002)
		fileFlagFirstPipe       = uint32(0x00080000)
		fileFlagOverlapped      = uint32(0x40000000)
		pipeTypeByte            = uint32(0x00000000)
		pipeReadModeByte        = uint32(0x00000000)
		pipeWait                = uint32(0x00000000)
		pipeRejectRemoteClients = uint32(0x00000008)
		maxPipeInstances        = uint32(1)
	)
	createNamedPipe := windows.NewLazySystemDLL("kernel32.dll").NewProc("CreateNamedPipeW")
	mode := pipeTypeByte | pipeReadModeByte | pipeWait | pipeRejectRemoteClients
	// PIPE_ACCESS_OUTBOUND maps to GENERIC_WRITE, which includes READ_CONTROL
	// for the created handle so the DACL can be inspected with GetSecurityInfo.
	result, _, callErr := createNamedPipe.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(pipeAccessOutbound|fileFlagFirstPipe|fileFlagOverlapped),
		uintptr(mode), uintptr(maxPipeInstances), uintptr(pipeBufferSize), 0, 0,
		uintptr(unsafe.Pointer(attributes)),
	)
	runtime.KeepAlive(name)
	runtime.KeepAlive(attributes)
	if windows.Handle(result) == windows.InvalidHandle {
		return 0, fmt.Errorf("create private HTTPS broker named pipe: %w", callErr)
	}
	return windows.Handle(result), nil
}

func currentUserPipeSecurityDescriptor() (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("read current user SID for HTTPS broker pipe: %w", err)
	}
	if user == nil || user.User.Sid == nil {
		return nil, errors.New("current user SID is unavailable for HTTPS broker pipe")
	}
	sddl := "D:P(A;;GA;;;" + user.User.Sid.String() + ")"
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return nil, fmt.Errorf("create current-user-only HTTPS broker pipe DACL: %w", err)
	}
	return descriptor, nil
}

func httpsBrokerSecretTransportPresent(getenv func(string) string) bool {
	return validHTTPSBrokerPipeName(getenv(EnvHTTPSBrokerTransport))
}

func validHTTPSBrokerPipeName(name string) bool {
	if !strings.HasPrefix(name, httpsBrokerPipePrefix) {
		return false
	}
	suffix := strings.TrimPrefix(name, httpsBrokerPipePrefix)
	if len(suffix) != 64 {
		return false
	}
	_, err := hex.DecodeString(suffix)
	return err == nil && strings.ToLower(suffix) == suffix
}

func readHTTPSBrokerSecret(getenv func(string) string) ([]byte, bool) {
	name := getenv(EnvHTTPSBrokerTransport)
	if !validHTTPSBrokerPipeName(name) {
		return nil, false
	}
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(name), windows.GENERIC_READ, 0, nil,
		windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0,
	)
	if err != nil {
		return nil, false
	}
	var secret []byte
	buffer := make([]byte, pipeBufferSize)
	defer clear(buffer)
	for {
		var read uint32
		err := windows.ReadFile(handle, buffer, &read, nil)
		if errors.Is(err, windows.ERROR_BROKEN_PIPE) {
			break
		}
		if err != nil {
			clear(secret)
			_ = windows.CloseHandle(handle)
			return nil, false
		}
		if read == 0 {
			break
		}
		secret = append(secret, buffer[:read]...)
	}
	if err := windows.CloseHandle(handle); err != nil {
		clear(secret)
		return nil, false
	}
	return secret, true
}
