//go:build !windows

package buildrepo

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
)

// EnvHTTPSBrokerTransport names the secret-free Unix descriptor carried to
// the fetch child.
const EnvHTTPSBrokerTransport = "CURATOR_BUILD_HTTPS_ASKPASS_HANDLE"

type unixHTTPSBrokerSecretTransport struct {
	reader      *os.File
	writer      *os.File
	secret      string
	handle      string
	closeReader sync.Once
	closeWriter sync.Once
	readerErr   error
	writerErr   error
}

// NewHTTPSBrokerSecretTransport creates a fetch-scoped pipe for the askpass
// secret and configures command to inherit its read end.
func NewHTTPSBrokerSecretTransport(command *exec.Cmd, secret string) (HTTPSBrokerSecretTransport, error) {
	if command == nil {
		return nil, errors.New("HTTPS broker command is nil")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	fd := 3 + len(command.ExtraFiles)
	command.ExtraFiles = append(command.ExtraFiles, reader)
	return &unixHTTPSBrokerSecretTransport{
		reader: reader,
		writer: writer,
		secret: secret,
		handle: strconv.Itoa(fd),
	}, nil
}

func (transport *unixHTTPSBrokerSecretTransport) EnvironmentValue() string {
	return transport.handle
}

func (transport *unixHTTPSBrokerSecretTransport) ChildStarted() {
	transport.readerErr = transport.closeReaderOnce()
}

func (transport *unixHTTPSBrokerSecretTransport) Serve(ctx context.Context) error {
	if transport.readerErr != nil {
		return transport.readerErr
	}
	writeDone := make(chan error, 1)
	go func() {
		_, err := io.WriteString(transport.writer, transport.secret)
		writeDone <- errors.Join(err, transport.closeWriterOnce())
	}()
	select {
	case err := <-writeDone:
		return err
	case <-ctx.Done():
		closeErr := transport.closeWriterOnce()
		writeErr := <-writeDone
		if ctx.Err() != nil {
			return nil
		}
		return errors.Join(closeErr, writeErr)
	}
}

func (transport *unixHTTPSBrokerSecretTransport) Close() error {
	return errors.Join(transport.closeReaderOnce(), transport.closeWriterOnce())
}

func (transport *unixHTTPSBrokerSecretTransport) closeReaderOnce() error {
	transport.closeReader.Do(func() {
		if transport.reader != nil {
			transport.readerErr = transport.reader.Close()
			if errors.Is(transport.readerErr, os.ErrClosed) || errors.Is(transport.readerErr, syscall.EBADF) {
				transport.readerErr = nil
			}
		}
	})
	return transport.readerErr
}

func (transport *unixHTTPSBrokerSecretTransport) closeWriterOnce() error {
	transport.closeWriter.Do(func() {
		if transport.writer != nil {
			transport.writerErr = transport.writer.Close()
		}
	})
	return transport.writerErr
}

func httpsBrokerSecretTransportPresent(getenv func(string) string) bool {
	_, ok := parseHTTPSBrokerHandle(getenv(EnvHTTPSBrokerTransport))
	return ok
}

func readHTTPSBrokerSecret(getenv func(string) string) ([]byte, bool) {
	handle, ok := parseHTTPSBrokerHandle(getenv(EnvHTTPSBrokerTransport))
	if !ok {
		return nil, false
	}
	pipe := os.NewFile(handle, "https-askpass-secret")
	if pipe == nil {
		return nil, false
	}
	secret, readErr := io.ReadAll(pipe)
	closeErr := pipe.Close()
	if readErr != nil || closeErr != nil {
		clear(secret)
		return nil, false
	}
	return secret, true
}

func parseHTTPSBrokerHandle(value string) (uintptr, bool) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed <= 2 || uint64(uintptr(parsed)) != parsed {
		return 0, false
	}
	return uintptr(parsed), true
}
