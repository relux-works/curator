//go:build windows

package envprofile

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/relux-works/curator/internal/gitops"
	"golang.org/x/sys/windows"
)

var errSignatureVerifierExited = errors.New("signature verifier exited before consuming verification input")

type signatureInputPipe struct {
	handle windows.Handle
}

// runSignaturePipes supplies verifier inputs through restricted, local named
// pipes. Windows os/exec cannot inherit Unix-style ExtraFiles, and no verifier
// input is materialized as a filesystem path.
func runSignaturePipes(command *exec.Cmd, payload []byte, inputs ...[]byte) ([]byte, error) {
	if len(command.Args) == 0 {
		return nil, errors.New("signature verifier command is empty")
	}
	commandArgs, err := signaturePipeArgs(command.Args[1:], len(inputs))
	if err != nil {
		return nil, err
	}
	command.Args = append(command.Args[:1], commandArgs...)
	command.Env = gitops.ReadOnlyEnv()
	command.Stdin = bytes.NewReader(payload)
	if len(inputs) == 0 {
		return command.CombinedOutput()
	}

	pipes, names, err := createSignatureInputPipes(len(inputs))
	if err != nil {
		return nil, fmt.Errorf("%w: Windows signature input pipes are unavailable: %v", errSignerVerificationUnavailable, err)
	}
	for i := range command.Args {
		for inputIndex, name := range names {
			if command.Args[i] == fmt.Sprintf("/dev/fd/%d", inputIndex+3) {
				command.Args[i] = name
			}
		}
	}

	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		closeSignatureInputPipes(pipes)
		return nil, err
	}
	processID := command.Process.Pid
	if processID <= 0 || uint64(processID) > uint64(^uint32(0)) {
		_ = command.Process.Kill()
		_ = command.Wait()
		closeSignatureInputPipes(pipes)
		return output.Bytes(), errors.New("signature verifier process ID is outside the Windows range")
	}
	windowsProcessID := uint32(processID)
	processHandle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, windowsProcessID)
	if err != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		closeSignatureInputPipes(pipes)
		return output.Bytes(), fmt.Errorf("could not monitor signature verifier process: %w", err)
	}

	writeResults := make(chan error, len(inputs))
	for i, pipe := range pipes {
		go func(pipe *signatureInputPipe, input []byte) {
			writeResults <- writeSignatureInputPipe(pipe, processHandle, windowsProcessID, input)
		}(pipe, inputs[i]) // #nosec G602 -- createSignatureInputPipes(len(inputs)) returns exactly one pipe per input.
	}
	runErr := command.Wait()
	var writeErr error
	for range inputs {
		writeErr = errors.Join(writeErr, <-writeResults)
	}
	processCloseErr := windows.CloseHandle(processHandle)
	if runErr != nil {
		return output.Bytes(), runErr
	}
	if writeErr != nil || processCloseErr != nil {
		return output.Bytes(), fmt.Errorf("verification input pipe: %w", errors.Join(writeErr, processCloseErr))
	}
	return output.Bytes(), nil
}

func signaturePipeArgs(args []string, inputCount int) ([]string, error) {
	result := append([]string(nil), args...)
	for inputIndex := 0; inputIndex < inputCount; inputIndex++ {
		placeholder := fmt.Sprintf("/dev/fd/%d", inputIndex+3)
		found := false
		for _, arg := range result {
			if arg == placeholder {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("signature input %d has no verifier argument", inputIndex)
		}
	}
	for _, arg := range result {
		if strings.HasPrefix(arg, "/dev/fd/") {
			index, err := parseSignaturePipeIndex(arg)
			if err != nil || index < 3 || index-3 >= inputCount {
				return nil, fmt.Errorf("verifier requested an unavailable signature input")
			}
		}
	}
	return result, nil
}

func parseSignaturePipeIndex(value string) (int, error) {
	index, err := strconv.Atoi(strings.TrimPrefix(value, "/dev/fd/"))
	if err != nil || fmt.Sprintf("/dev/fd/%d", index) != value {
		if err == nil {
			err = errors.New("non-canonical verifier input path")
		}
		return 0, err
	}
	return index, nil
}

func createSignatureInputPipes(count int) ([]*signatureInputPipe, []string, error) {
	if count == 0 {
		return nil, nil, nil
	}
	groups, err := windows.GetCurrentProcessToken().GetTokenGroups()
	if err != nil {
		return nil, nil, err
	}
	logonSID := ""
	for _, group := range groups.AllGroups() {
		if group.Attributes&windows.SE_GROUP_LOGON_ID == windows.SE_GROUP_LOGON_ID && group.Sid != nil {
			logonSID = group.Sid.String()
			break
		}
	}
	if logonSID == "" {
		return nil, nil, errors.New("current logon SID is unavailable")
	}
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;;GA;;;SY)(A;;GA;;;" + logonSID + ")")
	if err != nil {
		return nil, nil, err
	}
	attributes := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: descriptor,
		InheritHandle:      0,
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, nil, err
	}
	prefix := hex.EncodeToString(random[:])
	pipes := make([]*signatureInputPipe, 0, count)
	names := make([]string, 0, count)
	for index := 0; index < count; index++ {
		name := fmt.Sprintf(`\\.\pipe\curator-readonly-signature-%s-%d`, prefix, index)
		name16, err := windows.UTF16PtrFromString(name)
		if err != nil {
			closeSignatureInputPipes(pipes)
			return nil, nil, err
		}
		handle, err := windows.CreateNamedPipe(
			name16,
			windows.PIPE_ACCESS_OUTBOUND|windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE,
			windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT|windows.PIPE_REJECT_REMOTE_CLIENTS,
			1,
			64*1024,
			0,
			0,
			attributes,
		)
		runtime.KeepAlive(descriptor)
		runtime.KeepAlive(attributes)
		if err != nil {
			closeSignatureInputPipes(pipes)
			return nil, nil, err
		}
		pipes = append(pipes, &signatureInputPipe{handle: handle})
		names = append(names, name)
	}
	return pipes, names, nil
}

func closeSignatureInputPipes(pipes []*signatureInputPipe) {
	for _, pipe := range pipes {
		_ = pipe.close()
	}
}

func (pipe *signatureInputPipe) close() error {
	if pipe.handle == 0 || pipe.handle == windows.InvalidHandle {
		return nil
	}
	handle := pipe.handle
	pipe.handle = 0
	return windows.CloseHandle(handle)
}

func writeSignatureInputPipe(pipe *signatureInputPipe, process windows.Handle, processID uint32, input []byte) (result error) {
	defer func() { result = errors.Join(result, pipe.close()) }()
	connectEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fmt.Errorf("create signature pipe event: %w", err)
	}
	connectOverlapped := &windows.Overlapped{HEvent: connectEvent}
	connectErr := windows.ConnectNamedPipe(pipe.handle, connectOverlapped)
	switch {
	case connectErr == nil, errors.Is(connectErr, windows.ERROR_PIPE_CONNECTED):
	case errors.Is(connectErr, windows.ERROR_IO_PENDING):
		if _, err := waitSignaturePipeIO(pipe, process, connectOverlapped, connectEvent); err != nil {
			_ = windows.CloseHandle(connectEvent)
			return fmt.Errorf("connect verifier to signature input pipe: %w", err)
		}
	default:
		_ = windows.CloseHandle(connectEvent)
		return fmt.Errorf("connect verifier to signature input pipe: %w", connectErr)
	}
	if err := windows.CloseHandle(connectEvent); err != nil {
		return fmt.Errorf("close signature pipe event: %w", err)
	}
	var clientProcessID uint32
	if err := windows.GetNamedPipeClientProcessId(pipe.handle, &clientProcessID); err != nil {
		return fmt.Errorf("identify signature pipe client: %w", err)
	}
	if clientProcessID != processID {
		return errors.New("unexpected process connected to signature input pipe")
	}

	for len(input) > 0 {
		chunk := input
		if len(chunk) > 64*1024 {
			chunk = chunk[:64*1024]
		}
		event, err := windows.CreateEvent(nil, 1, 0, nil)
		if err != nil {
			return fmt.Errorf("create signature pipe event: %w", err)
		}
		overlapped := &windows.Overlapped{HEvent: event}
		var written uint32
		writeErr := windows.WriteFile(pipe.handle, chunk, &written, overlapped)
		if errors.Is(writeErr, windows.ERROR_IO_PENDING) {
			written, writeErr = waitSignaturePipeIO(pipe, process, overlapped, event)
		}
		_ = windows.CloseHandle(event)
		if writeErr != nil {
			return fmt.Errorf("write verifier input pipe: %w", writeErr)
		}
		if written == 0 {
			return io.ErrNoProgress
		}
		input = input[written:]
	}
	return nil
}

func waitSignaturePipeIO(pipe *signatureInputPipe, process windows.Handle, overlapped *windows.Overlapped, event windows.Handle) (uint32, error) {
	wait, err := windows.WaitForMultipleObjects([]windows.Handle{event, process}, false, windows.INFINITE)
	if err != nil {
		return 0, cancelSignaturePipeIO(pipe, overlapped, event, fmt.Errorf("wait for verifier input: %w", err))
	}
	if wait == windows.WAIT_OBJECT_0 {
		var completed uint32
		if err := windows.GetOverlappedResult(pipe.handle, overlapped, &completed, false); err != nil {
			return 0, err
		}
		return completed, nil
	}
	if wait == windows.WAIT_OBJECT_0+1 {
		return 0, cancelSignaturePipeIO(pipe, overlapped, event, errSignatureVerifierExited)
	}
	return 0, cancelSignaturePipeIO(pipe, overlapped, event, fmt.Errorf("unexpected verifier pipe wait result 0x%x", wait))
}

func cancelSignaturePipeIO(pipe *signatureInputPipe, overlapped *windows.Overlapped, event windows.Handle, cause error) error {
	cancelErr := windows.CancelIoEx(pipe.handle, overlapped)
	if errors.Is(cancelErr, windows.ERROR_NOT_FOUND) {
		cancelErr = nil
	} else if cancelErr != nil {
		_ = pipe.close()
	}
	wait, waitErr := windows.WaitForSingleObject(event, windows.INFINITE)
	if waitErr == nil && wait != windows.WAIT_OBJECT_0 {
		waitErr = fmt.Errorf("unexpected canceled verifier pipe wait result 0x%x", wait)
	}
	return errors.Join(cause, cancelErr, waitErr)
}
