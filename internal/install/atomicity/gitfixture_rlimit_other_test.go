//go:build !linux && !darwin

package atomicity

import (
	"fmt"
	"runtime"
)

func gitFixtureRlimits() string {
	return fmt.Sprintf("unavailable on %s", runtime.GOOS)
}
