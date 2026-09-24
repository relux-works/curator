package scriptworker

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

var errOutputLimit = errors.New("interpreter output limit exceeded")

type boundedBuffer struct {
	buffer bytes.Buffer
	budget *outputBudget
	err    error
}

type outputBudget struct {
	mu        sync.Mutex
	remaining int64
	err       error
}

func (buffer *boundedBuffer) Write(payload []byte) (int, error) {
	buffer.budget.mu.Lock()
	defer buffer.budget.mu.Unlock()
	if buffer.budget.remaining <= 0 {
		buffer.err = errOutputLimit
		buffer.budget.err = errOutputLimit
		return 0, buffer.err
	}
	written := len(payload)
	if int64(written) > buffer.budget.remaining {
		written = int(buffer.budget.remaining)
	}
	_, _ = buffer.buffer.Write(payload[:written])
	buffer.budget.remaining -= int64(written)
	if written != len(payload) {
		buffer.err = errOutputLimit
		buffer.budget.err = errOutputLimit
		return written, buffer.err
	}
	return written, nil
}

func (buffer *boundedBuffer) Bytes() []byte {
	return append([]byte(nil), buffer.buffer.Bytes()...)
}

var _ io.Writer = (*boundedBuffer)(nil)
