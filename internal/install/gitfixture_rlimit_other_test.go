//go:build !linux && !darwin

package install

import (
	"fmt"
	"runtime"
)

func gitFixtureRlimits() string {
	return fmt.Sprintf("unavailable on %s", runtime.GOOS)
}
