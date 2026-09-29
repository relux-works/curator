//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package pathboundary

import (
	"fmt"
	"os"
)

func openReadNoFollow(path string) (*os.File, error) {
	return nil, fmt.Errorf("cannot prove no-follow reads for %s on this platform", path)
}
