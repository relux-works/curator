//go:build !windows

package scriptworker

import (
	"errors"
	"os"
)

func nativeWindowsExecHardlinks(*os.File, string) (windowsExecHardlinkOrigin, error) {
	return windowsExecHardlinkOrigin{}, errors.New("native Windows hard-link origin APIs are unavailable")
}
