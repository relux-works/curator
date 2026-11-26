//go:build !linux && !darwin

package homelease

import (
	"errors"

	"github.com/relux-works/curator/internal/homelease/state"
)

// errKernelChanged never fires without a native observer; it exists so the
// shared bind mapping compiles identically on every platform.
var errKernelChanged = errors.New("kernel incarnation changed during observation pass")

func nativeStartTimeKind() string { return "" }

// The independent observer is Linux and macOS only. Elsewhere no native
// variant is supported, and bind refuses as unavailable authority.
var errKernelUnsupported = errors.New("no native process observer on this platform")

func readKernelProcess(int64) (state.Process, error) {
	return state.Process{}, errKernelUnsupported
}
