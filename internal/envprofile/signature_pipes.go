//go:build !windows

package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/gitops"
	"io"
	"os"
	"os/exec"
)

// runSignaturePipes passes public verification inputs through inherited file
// descriptors. Neither signer policy nor signatures require a disk pathname.
func runSignaturePipes(command *exec.Cmd, payload []byte, inputs ...[]byte) ([]byte, error) {
	command.Env = gitops.ReadOnlyEnv()
	command.Stdin = bytes.NewReader(payload)
	readers := make([]*os.File, 0, len(inputs))
	writers := make([]*os.File, 0, len(inputs))
	defer func() {
		for _, f := range readers {
			_ = f.Close()
		}
		for _, f := range writers {
			_ = f.Close()
		}
	}()
	for range inputs {
		reader, writer, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		readers = append(readers, reader)
		writers = append(writers, writer)
	}
	command.ExtraFiles = readers
	done := make(chan error, len(inputs))
	for i, input := range inputs {
		go func(writer *os.File, data []byte) {
			_, err := io.Copy(writer, bytes.NewReader(data))
			_ = writer.Close()
			done <- err
		}(writers[i], input)
	}
	output, runErr := command.CombinedOutput()
	// A verifier may reject without consuming every pipe. Closing the parent's
	// readers before joining writers also bounds that failure path.
	for _, reader := range readers {
		_ = reader.Close()
	}
	var writeErr error
	for range inputs {
		writeErr = errors.Join(writeErr, <-done)
	}
	if runErr != nil {
		return output, runErr
	}
	if writeErr != nil {
		return output, fmt.Errorf("verification input pipe: %w", writeErr)
	}
	return output, nil
}
