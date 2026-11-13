//go:build !unix && !windows

package pathboundary

import (
	"fmt"
	"os"
)

func singleLink(_ string, _ os.FileInfo) error {
	return fmt.Errorf("regular-file link metadata is unsupported on this platform")
}
