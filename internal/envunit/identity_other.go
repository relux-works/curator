//go:build !unix && !windows

package envunit

import (
	"fmt"
	"os"
)

func metadataIdentity(string, os.FileInfo) (string, error) {
	return "", fmt.Errorf("durable environment metadata transport is unsupported on this target")
}
