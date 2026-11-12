//go:build !unix && !windows

package skillintegrity

import (
	"errors"
	"os"
)

func singleLink(string, os.FileInfo) bool { return false }
func openCaptureFile(string) (*os.File, error) {
	return nil, errors.New("skill_integrity_unavailable: regular-file descriptor admission is unavailable")
}
