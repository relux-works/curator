//go:build !unix && !windows

package envorigin

import (
	"fmt"
	"os"
)

func openDirectoryMetadata(string) (*os.File, error) {
	return nil, fmt.Errorf("native directory metadata transport is unsupported on this target")
}
